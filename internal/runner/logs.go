package runner

import (
	"context"
	"fmt"
	"strings"

	"github.com/HXong/predeploy-guard/internal/config"
	predeployruntime "github.com/HXong/predeploy-guard/internal/runtime"
)

func collectDiagnosticsSafely(
	ctx context.Context,
	adapter predeployruntime.Adapter,
	env *predeployruntime.Environment,
	cfg *config.Config,
) (string, error) {
	fmt.Println("Collecting runtime diagnostics...")

	diagnostics, err := adapter.CollectDiagnostics(ctx, env, cfg)
	details := ""
	if diagnostics != nil {
		details = strings.Join(diagnostics.Details, "\n")
	}

	if err != nil {
		return fmt.Sprintf("Failed to collect runtime diagnostics: %v\n\nPartial output:\n%s", err, details), err
	}

	return details, nil
}
