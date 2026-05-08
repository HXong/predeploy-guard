package loadtest

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/HXong/predeploy-guard/internal/config"
)

type K6Result struct {
	Enabled         bool
	Passed          bool
	VUs             int
	Duration        string
	P95LatencyMs    float64
	MaxP95LatencyMs float64
	ErrorRate       float64
	MaxErrorRate    float64
	ScriptPath      string
	SummaryPath     string
	Error           string
	Output          string
}

type predeployK6Summary struct {
	P95LatencyMs *float64 `json:"p95LatencyMs"`
	ErrorRate    *float64 `json:"errorRate"`
}

func RunK6IfEnabled(cfg *config.Config, K6BaseURL string, workDir string) K6Result {
	perf := cfg.Performance

	result := K6Result{
		Enabled:         perf.Enabled,
		VUs:             perf.VUs,
		Duration:        perf.Duration,
		MaxP95LatencyMs: perf.Thresholds.MaxP95LatencyMs,
		MaxErrorRate:    perf.Thresholds.MaxErrorRate,
	}

	if !perf.Enabled {
		result.Passed = true
		return result
	}

	scriptPath := filepath.Join(workDir, "predeploy-k6.js")
	summaryPath := filepath.Join(workDir, "predeploy-summary.json")

	result.ScriptPath = scriptPath
	result.SummaryPath = summaryPath

	script := buildK6Script(cfg, K6BaseURL)

	if err := os.WriteFile(scriptPath, []byte(script), 0644); err != nil {
		result.Passed = false
		result.Error = fmt.Sprintf("write k6 script: %v", err)
		return result
	}

	output, err := runK6Docker(workDir)
	result.Output = output

	if err != nil {
		result.Passed = false
		result.Error = fmt.Sprintf("k6 docker run failed: %v", err)
		return result
	}

	if err := parseK6Summary(summaryPath, &result); err != nil {
		result.Passed = false
		result.Error = err.Error()
		return result
	}

	result.Passed = evaluateK6Result(result)

	return result
}
