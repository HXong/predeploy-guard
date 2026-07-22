package runtime

import (
	"context"

	"github.com/HXong/predeploy-guard/internal/config"
)

type Type string

const (
	TypeDockerCompose Type = "docker-compose"
)

type Environment struct {
	Name            string
	WorkDir         string
	BaseURL         string
	WorkloadBaseURL string
}

type ReadinessResult struct {
	Name    string
	Target  string
	Passed  bool
	Skipped bool
	Error   string
}

type Diagnostics struct {
	Runtime string
	Details []string
}

type Adapter interface {
	Type() Type

	Prepare(ctx context.Context, cfg *config.Config) (*Environment, error)
	Start(ctx context.Context, env *Environment, cfg *config.Config) error
	WaitReady(ctx context.Context, env *Environment, cfg *config.Config) ([]ReadinessResult, error)
	CollectDiagnostics(ctx context.Context, env *Environment, cfg *config.Config) (*Diagnostics, error)
	Cleanup(ctx context.Context, env *Environment, cfg *config.Config) error
}
