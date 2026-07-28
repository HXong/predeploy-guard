package runner

import (
	"time"

	"github.com/HXong/predeploy-guard/internal/builder"
	"github.com/HXong/predeploy-guard/internal/config"
	"github.com/HXong/predeploy-guard/internal/loadtest"
	"github.com/HXong/predeploy-guard/internal/report"
	predeployruntime "github.com/HXong/predeploy-guard/internal/runtime"
)

const maxGeneratedIngressGatewayCheckTimeout = 30 * time.Second

func appendSkippedPhases(phases *[]report.RunPhase, names []string, reason string) {
	for _, name := range names {
		*phases = append(*phases, report.SkippedPhase(name, reason))
	}
}

func generatedIngressGatewayCheckTimeout(cfg *config.Config) time.Duration {
	configuredTimeout := time.Duration(cfg.Settings.TimeoutSeconds) * time.Second
	if configuredTimeout <= 0 || configuredTimeout > maxGeneratedIngressGatewayCheckTimeout {
		return maxGeneratedIngressGatewayCheckTimeout
	}
	return configuredTimeout
}

func runtimeEnvironmentSummary(
	runtimeName string,
	env *predeployruntime.Environment,
) report.RuntimeEnvironment {
	summary := report.RuntimeEnvironment{
		Runtime: runtimeName,
	}
	if env == nil {
		return summary
	}

	summary.Name = env.Name
	summary.BaseURL = env.BaseURL
	summary.WorkloadBaseURL = env.WorkloadBaseURL
	return summary
}

func reportBuildResult(result builder.BuildResult) report.BuildResult {
	return report.BuildResult{
		Enabled:    result.Enabled,
		Image:      result.Image,
		Context:    result.Context,
		Dockerfile: result.Dockerfile,
		Passed:     result.Passed,
		Error:      result.Error,
		Output:     result.Output,
	}
}

func reportPerformanceResult(result loadtest.K6Result) report.PerformanceResult {
	return report.PerformanceResult{
		Enabled:  result.Enabled,
		Passed:   result.Passed,
		VUs:      result.VUs,
		Duration: result.Duration,

		AvgLatencyMs:    result.AvgLatencyMs,
		MinLatencyMs:    result.MinLatencyMs,
		MedianLatencyMs: result.MedianLatencyMs,
		MaxLatencyMs:    result.MaxLatencyMs,
		P90LatencyMs:    result.P90LatencyMs,
		P95LatencyMs:    result.P95LatencyMs,
		MaxP95LatencyMs: result.MaxP95LatencyMs,

		ErrorRate:    result.ErrorRate,
		MaxErrorRate: result.MaxErrorRate,

		RequestCount:  result.RequestCount,
		Iterations:    result.Iterations,
		ChecksTotal:   result.ChecksTotal,
		CheckPassRate: result.CheckPassRate,

		Error:  result.Error,
		Output: result.Output,
	}
}
