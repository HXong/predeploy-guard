package sandbox

import "github.com/HXong/predeploy-guard/internal/config"

func dependencyHasReadiness(readiness config.ReadinessConfig) bool {
	return len(readiness.Command) > 0 || readiness.Shell != ""
}
