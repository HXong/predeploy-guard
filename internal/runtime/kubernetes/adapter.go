package kubernetes

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/HXong/predeploy-guard/internal/config"
	predeployruntime "github.com/HXong/predeploy-guard/internal/runtime"
)

const manifestFileName = "resources.yaml"

type Adapter struct {
	kubectl kubectlRunner

	environment      *predeployruntime.Environment
	workDir          string
	manifestPath     string
	namespace        string
	serviceName      string
	localPort        int
	namespaceCreated bool
	portForward      runningProcess
}

var _ predeployruntime.Adapter = (*Adapter)(nil)

func New() *Adapter {
	return newWithKubectlRunner(execKubectlRunner{})
}

func newWithKubectlRunner(runner kubectlRunner) *Adapter {
	return &Adapter{kubectl: runner}
}

func (a *Adapter) Type() predeployruntime.Type {
	return predeployruntime.TypeKubernetes
}

func (a *Adapter) Prepare(
	_ context.Context,
	cfg *config.Config,
) (*predeployruntime.Environment, error) {
	if a.environment != nil {
		return nil, fmt.Errorf("Kubernetes runtime environment is already prepared")
	}

	suffix, err := uniqueSuffix()
	if err != nil {
		return nil, err
	}

	workDir, err := os.MkdirTemp("", "predeploy-guard-kubernetes-*")
	if err != nil {
		return nil, fmt.Errorf("create Kubernetes runtime work directory: %w", err)
	}

	cleanupWorkDir := true
	defer func() {
		if cleanupWorkDir {
			_ = os.RemoveAll(workDir)
		}
	}()

	localPort, err := findFreeLocalPort()
	if err != nil {
		return nil, err
	}

	namespace := namespaceName(cfg.Settings.NamespacePrefix, cfg.Service.Name, suffix)
	manifest, err := generateManifests(cfg, namespace)
	if err != nil {
		return nil, err
	}

	manifestPath := filepath.Join(workDir, manifestFileName)
	if err := os.WriteFile(manifestPath, manifest, 0600); err != nil {
		return nil, fmt.Errorf("write Kubernetes manifests: %w", err)
	}

	env := &predeployruntime.Environment{
		Name:            namespace,
		WorkDir:         workDir,
		BaseURL:         fmt.Sprintf("http://127.0.0.1:%d", localPort),
		WorkloadBaseURL: fmt.Sprintf("http://host.docker.internal:%d", localPort),
	}

	a.environment = env
	a.workDir = workDir
	a.manifestPath = manifestPath
	a.namespace = namespace
	a.serviceName = sanitizeDNS1123(cfg.Service.Name)
	a.localPort = localPort
	a.namespaceCreated = false
	a.portForward = nil
	cleanupWorkDir = false

	return env, nil
}

func (a *Adapter) Start(
	ctx context.Context,
	env *predeployruntime.Environment,
	cfg *config.Config,
) error {
	if err := a.validateEnvironment(env); err != nil {
		return err
	}

	if err := a.runKubectl(
		ctx,
		cfg.Runtime.Context,
		"create temporary namespace",
		"create",
		"namespace",
		a.namespace,
	); err != nil {
		return err
	}
	a.namespaceCreated = true

	if err := a.runKubectl(
		ctx,
		cfg.Runtime.Context,
		"label temporary namespace",
		"label",
		"namespace",
		a.namespace,
		"app.kubernetes.io/managed-by=predeploy-guard",
		"app.kubernetes.io/part-of=predeploy-guard",
		"predeploy.guard/run="+a.namespace,
		"--overwrite",
	); err != nil {
		return err
	}

	// The Kubernetes runtime deliberately does not load locally built images.
	// The configured images must already be accessible to the selected cluster.
	return a.runKubectl(
		ctx,
		cfg.Runtime.Context,
		"apply Kubernetes manifests",
		"apply",
		"--namespace",
		a.namespace,
		"-f",
		a.manifestPath,
	)
}

func (a *Adapter) WaitReady(
	ctx context.Context,
	env *predeployruntime.Environment,
	cfg *config.Config,
) ([]predeployruntime.ReadinessResult, error) {
	if err := a.validateEnvironment(env); err != nil {
		return nil, err
	}

	results := make([]predeployruntime.ReadinessResult, 0, len(cfg.Dependencies)+2)
	for _, dependencyName := range sortedDependencyNames(cfg.Dependencies) {
		resourceName := sanitizeDNS1123(dependencyName)
		timeout := cfg.Dependencies[dependencyName].Readiness.TimeoutSeconds
		if timeout <= 0 {
			timeout = runtimeTimeoutSeconds(cfg)
		}

		err := a.waitForDeployment(ctx, cfg.Runtime.Context, resourceName, timeout)
		result := predeployruntime.ReadinessResult{
			Name:   "dependency readiness",
			Target: resourceName,
			Passed: err == nil,
		}
		if err != nil {
			result.Error = err.Error()
			results = append(results, result)
			return results, err
		}
		results = append(results, result)
	}

	err := a.waitForDeployment(
		ctx,
		cfg.Runtime.Context,
		a.serviceName,
		runtimeTimeoutSeconds(cfg),
	)
	serviceResult := predeployruntime.ReadinessResult{
		Name:   "runtime readiness",
		Target: "deployment/" + a.serviceName,
		Passed: err == nil,
	}
	if err != nil {
		serviceResult.Error = err.Error()
		results = append(results, serviceResult)
		return results, err
	}
	results = append(results, serviceResult)

	portForwardArgs := kubectlArgs(
		cfg.Runtime.Context,
		"port-forward",
		"--namespace",
		a.namespace,
		"service/"+a.serviceName,
		fmt.Sprintf("%d:%d", a.localPort, cfg.Service.Port),
		"--address",
		// Dockerized k6 reaches the forwarded port through host.docker.internal.
		"0.0.0.0",
	)
	process, err := a.kubectl.Start(ctx, portForwardArgs...)
	if err != nil {
		return results, fmt.Errorf("start kubectl port-forward: %w", err)
	}
	a.portForward = process

	err = waitForLocalPort(
		ctx,
		"127.0.0.1",
		a.localPort,
		time.Duration(runtimeTimeoutSeconds(cfg))*time.Second,
		process.Done(),
	)
	accessResult := predeployruntime.ReadinessResult{
		Name:   "runtime access",
		Target: env.BaseURL,
		Passed: err == nil,
	}
	if err != nil {
		accessResult.Error = err.Error()
		results = append(results, accessResult)
		output := process.Output()
		_ = process.Stop()
		a.portForward = nil
		if output != "" {
			return results, fmt.Errorf("%w\nOutput:\n%s", err, output)
		}
		return results, err
	}
	results = append(results, accessResult)

	return results, nil
}

func (a *Adapter) CollectDiagnostics(
	ctx context.Context,
	env *predeployruntime.Environment,
	cfg *config.Config,
) (*predeployruntime.Diagnostics, error) {
	if err := a.validateEnvironment(env); err != nil {
		return nil, err
	}

	commands := []struct {
		title string
		args  []string
	}{
		{
			title: "kubectl get all",
			args:  []string{"get", "all", "--namespace", a.namespace},
		},
		{
			title: "kubectl describe pods",
			args:  []string{"describe", "pods", "--namespace", a.namespace},
		},
		{
			title: "kubectl logs",
			args: []string{
				"logs",
				"--namespace",
				a.namespace,
				"--selector",
				"predeploy.guard/run=" + a.namespace,
				"--all-containers=true",
				"--prefix=true",
				"--tail=200",
			},
		},
	}

	diagnostics := &predeployruntime.Diagnostics{
		Runtime: string(a.Type()),
		Details: make([]string, 0, len(commands)),
	}
	for _, command := range commands {
		args := kubectlArgs(cfg.Runtime.Context, command.args...)
		output, err := a.kubectl.Run(ctx, args...)

		var detail strings.Builder
		fmt.Fprintf(&detail, "## %s\n%s", command.title, output)
		if err != nil {
			if output != "" && !strings.HasSuffix(output, "\n") {
				detail.WriteByte('\n')
			}
			fmt.Fprintf(&detail, "Command failed: %v", err)
		}
		diagnostics.Details = append(diagnostics.Details, detail.String())
	}

	return diagnostics, nil
}

func (a *Adapter) Cleanup(
	ctx context.Context,
	env *predeployruntime.Environment,
	cfg *config.Config,
) error {
	if err := a.validateEnvironment(env); err != nil {
		return err
	}

	var cleanupErrors []error
	if a.portForward != nil {
		if err := a.portForward.Stop(); err != nil {
			cleanupErrors = append(cleanupErrors, fmt.Errorf("stop kubectl port-forward: %w", err))
		}
		a.portForward = nil
	}

	if a.namespaceCreated {
		output, err := a.kubectl.Run(
			ctx,
			kubectlArgs(
				cfg.Runtime.Context,
				"delete",
				"namespace",
				a.namespace,
				"--ignore-not-found=true",
				"--wait=true",
				"--timeout=60s",
			)...,
		)
		if err != nil {
			cleanupErrors = append(
				cleanupErrors,
				kubectlCommandError("delete temporary namespace", output, err),
			)
		}
	}

	if a.workDir != "" {
		if err := os.RemoveAll(a.workDir); err != nil {
			cleanupErrors = append(cleanupErrors, fmt.Errorf("remove Kubernetes work directory: %w", err))
		}
	}

	a.environment = nil
	a.workDir = ""
	a.manifestPath = ""
	a.namespace = ""
	a.serviceName = ""
	a.localPort = 0
	a.namespaceCreated = false

	return errors.Join(cleanupErrors...)
}

func (a *Adapter) waitForDeployment(
	ctx context.Context,
	contextName string,
	name string,
	timeoutSeconds int,
) error {
	output, err := a.kubectl.Run(
		ctx,
		kubectlArgs(
			contextName,
			"rollout",
			"status",
			"deployment/"+name,
			"--namespace",
			a.namespace,
			fmt.Sprintf("--timeout=%ds", timeoutSeconds),
		)...,
	)
	if err != nil {
		return kubectlCommandError("wait for deployment "+name, output, err)
	}

	return nil
}

func (a *Adapter) runKubectl(
	ctx context.Context,
	contextName string,
	action string,
	args ...string,
) error {
	output, err := a.kubectl.Run(ctx, kubectlArgs(contextName, args...)...)
	if err != nil {
		return kubectlCommandError(action, output, err)
	}

	return nil
}

func (a *Adapter) validateEnvironment(env *predeployruntime.Environment) error {
	if env == nil {
		return fmt.Errorf("Kubernetes runtime environment is required")
	}
	if a.environment == nil {
		return fmt.Errorf("Kubernetes runtime environment is not prepared")
	}
	if env != a.environment {
		return fmt.Errorf("Kubernetes runtime environment does not belong to this adapter")
	}

	return nil
}

func kubectlCommandError(action string, output string, err error) error {
	if output == "" {
		return fmt.Errorf("%s: %w", action, err)
	}

	return fmt.Errorf("%s: %w\nOutput:\n%s", action, err, output)
}

func runtimeTimeoutSeconds(cfg *config.Config) int {
	if cfg.Settings.TimeoutSeconds > 0 {
		return cfg.Settings.TimeoutSeconds
	}

	return 60
}

func waitForLocalPort(
	ctx context.Context,
	host string,
	port int,
	timeout time.Duration,
	processDone <-chan error,
) error {
	deadline := time.NewTimer(timeout)
	defer deadline.Stop()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	address := net.JoinHostPort(host, fmt.Sprintf("%d", port))
	for {
		connection, err := net.DialTimeout("tcp", address, 100*time.Millisecond)
		if err == nil {
			_ = connection.Close()
			return nil
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf("wait for kubectl port-forward: %w", ctx.Err())
		case processErr := <-processDone:
			if processErr == nil {
				return fmt.Errorf("kubectl port-forward exited before local port %s was ready", address)
			}
			return fmt.Errorf(
				"kubectl port-forward exited before local port %s was ready: %w",
				address,
				processErr,
			)
		case <-deadline.C:
			return fmt.Errorf("kubectl port-forward did not open local port %s within %s", address, timeout)
		case <-ticker.C:
		}
	}
}
