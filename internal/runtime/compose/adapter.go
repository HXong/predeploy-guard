package compose

import (
	"context"
	"errors"
	"fmt"

	"github.com/HXong/predeploy-guard/internal/config"
	predeployruntime "github.com/HXong/predeploy-guard/internal/runtime"
	"github.com/HXong/predeploy-guard/internal/sandbox"
)

type Adapter struct {
	sandbox        *sandbox.ComposeSandbox
	environment    *predeployruntime.Environment
	startAttempted bool
}

var _ predeployruntime.Adapter = (*Adapter)(nil)

func New() *Adapter {
	return &Adapter{}
}

func (a *Adapter) Type() predeployruntime.Type {
	return predeployruntime.TypeDockerCompose
}

func (a *Adapter) Prepare(_ context.Context, cfg *config.Config) (*predeployruntime.Environment, error) {
	sb, err := sandbox.NewComposeSandbox(cfg)
	if err != nil {
		return nil, err
	}

	env := &predeployruntime.Environment{
		Name:            sb.ServiceName,
		WorkDir:         sb.WorkDir,
		BaseURL:         sb.BaseURL(),
		WorkloadBaseURL: sb.K6BaseURL(),
	}

	a.sandbox = sb
	a.environment = env
	a.startAttempted = false

	return env, nil
}

func (a *Adapter) Start(_ context.Context, env *predeployruntime.Environment, _ *config.Config) error {
	sb, err := a.sandboxFor(env)
	if err != nil {
		return err
	}

	a.startAttempted = true
	return sb.Start()
}

func (a *Adapter) WaitReady(_ context.Context, env *predeployruntime.Environment, cfg *config.Config) ([]predeployruntime.ReadinessResult, error) {
	sb, err := a.sandboxFor(env)
	if err != nil {
		return nil, err
	}

	dependencyResults, readinessErr := sb.WaitForDependencies(cfg)
	results := make([]predeployruntime.ReadinessResult, 0, len(dependencyResults))

	for _, result := range dependencyResults {
		name := "dependency readiness"
		if result.Skipped {
			name = "dependency readiness (skipped)"
		}

		results = append(results, predeployruntime.ReadinessResult{
			Name:    name,
			Target:  result.Name,
			Passed:  result.Passed,
			Skipped: result.Skipped,
			Error:   result.Error,
		})
	}

	return results, readinessErr
}

func (a *Adapter) CollectDiagnostics(_ context.Context, env *predeployruntime.Environment, _ *config.Config) (*predeployruntime.Diagnostics, error) {
	sb, err := a.sandboxFor(env)
	if err != nil {
		return nil, err
	}

	logs, logsErr := sb.Logs()
	diagnostics := &predeployruntime.Diagnostics{
		Runtime: string(a.Type()),
	}
	if logs != "" {
		diagnostics.Details = []string{logs}
	}

	return diagnostics, logsErr
}

func (a *Adapter) Cleanup(_ context.Context, env *predeployruntime.Environment, _ *config.Config) error {
	sb, err := a.sandboxFor(env)
	if err != nil {
		return err
	}

	var cleanupErrors []error
	if a.startAttempted {
		if err := sb.Stop(); err != nil {
			cleanupErrors = append(cleanupErrors, fmt.Errorf("stop Docker Compose sandbox: %w", err))
		}
	}

	if err := sb.RemoveFiles(); err != nil {
		cleanupErrors = append(cleanupErrors, fmt.Errorf("remove sandbox files: %w", err))
	}

	a.sandbox = nil
	a.environment = nil
	a.startAttempted = false

	return errors.Join(cleanupErrors...)
}

func (a *Adapter) sandboxFor(env *predeployruntime.Environment) (*sandbox.ComposeSandbox, error) {
	if env == nil {
		return nil, fmt.Errorf("Docker Compose runtime environment is required")
	}
	if a.sandbox == nil || a.environment == nil {
		return nil, fmt.Errorf("Docker Compose runtime environment is not prepared")
	}
	if env != a.environment {
		return nil, fmt.Errorf("Docker Compose runtime environment does not belong to this adapter")
	}

	return a.sandbox, nil
}
