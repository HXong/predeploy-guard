package loadtest

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/HXong/predeploy-guard/internal/config"
)

const K6DockerImage = "grafana/k6:latest"

func runK6Docker(workDir string) (string, error) {
	mountPath := toDockerMountPath(workDir)

	cmd := exec.Command(
		"docker",
		"run",
		"--rm",
		"-v",
		fmt.Sprintf("%s:/scripts", mountPath),
		K6DockerImage,
		"run",
		"/scripts/predeploy-k6.js",
	)

	output, err := cmd.CombinedOutput()
	return string(output), err
}

func buildK6Script(cfg *config.Config, baseURL string) string {
	var builder strings.Builder

	fmt.Fprintf(&builder, "import http from 'k6/http';\n")
	fmt.Fprintf(&builder, "import { check, sleep } from 'k6';\n\n")

	fmt.Fprintf(&builder, "export const options = {\n")
	fmt.Fprintf(&builder, "  vus: %d,\n", cfg.Performance.VUs)
	fmt.Fprintf(&builder, "  duration: '%s',\n", cfg.Performance.Duration)
	fmt.Fprintf(&builder, "  summaryTrendStats: ['avg', 'min', 'med', 'max', 'p(90)', 'p(95)'],\n")
	fmt.Fprintf(&builder, "};\n\n")

	fmt.Fprintf(&builder, "export default function () {\n")

	for _, endpoint := range cfg.Performance.Endpoints {
		method := strings.ToUpper(endpoint.Method)
		url := joinURL(baseURL, endpoint.Path)

		fmt.Fprintf(&builder, "  {\n")
		fmt.Fprintf(&builder, "    const res = http.request('%s', '%s');\n", escapeJS(method), escapeJS(url))
		fmt.Fprintf(
			&builder,
			"    check(res, { '%s status is 2xx': (r) => r.status >= 200 && r.status < 300 });\n",
			escapeJS(endpoint.Name),
		)
		fmt.Fprintf(&builder, "  }\n")
	}

	fmt.Fprintf(&builder, "  sleep(1);\n")
	fmt.Fprintf(&builder, "}\n\n")

	fmt.Fprintf(&builder, "export function handleSummary(data) {\n")
	fmt.Fprintf(&builder, "  const duration = data.metrics.http_req_duration;\n")
	fmt.Fprintf(&builder, "  const failed = data.metrics.http_req_failed;\n")
	fmt.Fprintf(&builder, "  const reqs = data.metrics.http_reqs;\n")
	fmt.Fprintf(&builder, "  const iterations = data.metrics.iterations;\n")
	fmt.Fprintf(&builder, "  const checks = data.metrics.checks;\n")
	fmt.Fprintf(&builder, "  const checksTotal = data.metrics.checks_total;\n")
	fmt.Fprintf(&builder, "  const summary = {\n")
	fmt.Fprintf(&builder, "    avgLatencyMs: duration && duration.values ? duration.values.avg : null,\n")
	fmt.Fprintf(&builder, "    minLatencyMs: duration && duration.values ? duration.values.min : null,\n")
	fmt.Fprintf(&builder, "    medianLatencyMs: duration && duration.values ? duration.values.med : null,\n")
	fmt.Fprintf(&builder, "    maxLatencyMs: duration && duration.values ? duration.values.max : null,\n")
	fmt.Fprintf(&builder, "    p90LatencyMs: duration && duration.values ? duration.values['p(90)'] : null,\n")
	fmt.Fprintf(&builder, "    p95LatencyMs: duration && duration.values ? duration.values['p(95)'] : null,\n")
	fmt.Fprintf(&builder, "    errorRate: failed && failed.values ? failed.values.rate : null,\n")
	fmt.Fprintf(&builder, "    requestCount: reqs && reqs.values ? Math.round(reqs.values.count || 0) : null,\n")
	fmt.Fprintf(&builder, "    iterations: iterations && iterations.values ? Math.round(iterations.values.count || 0) : null,\n")
	fmt.Fprintf(&builder, "    checksTotal: checksTotal && checksTotal.values ? Math.round(checksTotal.values.count || 0) : null,\n")
	fmt.Fprintf(&builder, "    checkPassRate: checks && checks.values ? checks.values.rate : null,\n")
	fmt.Fprintf(&builder, "  };\n")
	fmt.Fprintf(&builder, "  return {\n")
	fmt.Fprintf(&builder, "    '/scripts/predeploy-summary.json': JSON.stringify(summary, null, 2),\n")
	fmt.Fprintf(&builder, "  };\n")
	fmt.Fprintf(&builder, "}\n")

	return builder.String()
}

func parseK6Summary(summaryPath string, result *K6Result) error {
	data, err := os.ReadFile(summaryPath)
	if err != nil {
		return fmt.Errorf("read k6 summary: %w", err)
	}

	var summary predeployK6Summary
	if err := json.Unmarshal(data, &summary); err != nil {
		return fmt.Errorf("parse k6 summary: %w", err)
	}

	if summary.AvgLatencyMs == nil {
		return fmt.Errorf("custom k6 summary missing avgLatencyMs")
	}
	if summary.MinLatencyMs == nil {
		return fmt.Errorf("custom k6 summary missing minLatencyMs")
	}
	if summary.MedianLatencyMs == nil {
		return fmt.Errorf("custom k6 summary missing medianLatencyMs")
	}
	if summary.MaxLatencyMs == nil {
		return fmt.Errorf("custom k6 summary missing maxLatencyMs")
	}
	if summary.P90LatencyMs == nil {
		return fmt.Errorf("custom k6 summary missing p90LatencyMs")
	}
	if summary.P95LatencyMs == nil {
		return fmt.Errorf("custom k6 summary missing p95LatencyMs")
	}
	if summary.ErrorRate == nil {
		return fmt.Errorf("custom k6 summary missing errorRate")
	}
	if summary.RequestCount == nil {
		return fmt.Errorf("custom k6 summary missing requestCount")
	}
	if summary.Iterations == nil {
		return fmt.Errorf("custom k6 summary missing iterations")
	}
	if summary.ChecksTotal != nil {
		result.ChecksTotal = *summary.ChecksTotal
	}

	if summary.CheckPassRate != nil {
		result.CheckPassRate = *summary.CheckPassRate
	}

	result.AvgLatencyMs = *summary.AvgLatencyMs
	result.MinLatencyMs = *summary.MinLatencyMs
	result.MedianLatencyMs = *summary.MedianLatencyMs
	result.MaxLatencyMs = *summary.MaxLatencyMs
	result.P90LatencyMs = *summary.P90LatencyMs
	result.P95LatencyMs = *summary.P95LatencyMs
	result.ErrorRate = *summary.ErrorRate
	result.RequestCount = *summary.RequestCount
	result.Iterations = *summary.Iterations
	if summary.ChecksTotal != nil {
		result.ChecksTotal = *summary.ChecksTotal
	}
	if summary.CheckPassRate != nil {
		result.CheckPassRate = *summary.CheckPassRate
	}

	return nil
}

func evaluateK6Result(result K6Result) bool {
	if result.P95LatencyMs > result.MaxP95LatencyMs {
		return false
	}

	if result.ErrorRate > result.MaxErrorRate {
		return false
	}

	return true
}

func joinURL(baseURL string, path string) string {
	baseURL = strings.TrimRight(baseURL, "/")
	path = strings.TrimLeft(path, "/")

	return baseURL + "/" + path
}

func escapeJS(value string) string {
	value = strings.ReplaceAll(value, "\\", "\\\\")
	value = strings.ReplaceAll(value, "'", "\\'")
	return value
}

func toDockerMountPath(path string) string {
	return filepath.ToSlash(path)
}
