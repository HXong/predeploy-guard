package report

import (
	"fmt"
	"strings"
	"time"

	"github.com/HXong/predeploy-guard/internal/checker"
	"github.com/HXong/predeploy-guard/internal/config"
)

func buildMarkdown(cfg *config.Config, data ReportData) string {
	var builder strings.Builder

	status := "FAIL"
	if data.Passed {
		status = "PASS"
	}

	totalSmoke := len(data.Results)
	passedSmoke := 0
	failedSmoke := 0

	for _, result := range data.Results {
		if result.Passed {
			passedSmoke++
		} else {
			failedSmoke++
		}
	}

	totalReadiness := len(data.ReadinessResults)
	passedReadiness := 0
	failedReadiness := 0

	for _, result := range data.ReadinessResults {
		if result.Passed {
			passedReadiness++
		} else {
			failedReadiness++
		}
	}

	duration := data.FinishedAt.Sub(data.StartedAt)

	fmt.Fprintf(&builder, "# PreDeploy Guard Report\n\n")

	fmt.Fprintf(&builder, "## Summary\n\n")
	fmt.Fprintf(&builder, "- **Service:** %s\n", data.ServiceName)
	fmt.Fprintf(&builder, "- **Image:** `%s`\n", data.Image)
	fmt.Fprintf(&builder, "- **Runtime:** %s\n", data.Runtime)
	fmt.Fprintf(&builder, "- **Base URL:** %s\n", data.BaseURL)
	fmt.Fprintf(&builder, "- **Result:** **%s**\n", status)
	fmt.Fprintf(&builder, "- **Started At:** %s\n", data.StartedAt.Format(time.RFC3339))
	fmt.Fprintf(&builder, "- **Finished At:** %s\n", data.FinishedAt.Format(time.RFC3339))
	fmt.Fprintf(&builder, "- **Total Duration:** %s\n", duration.Round(time.Millisecond))
	fmt.Fprintf(&builder, "\n")

	writeBuildSection(&builder, data.BuildResult)
	writeReadinessSection(&builder, data.ReadinessResults)
	writeSmokeSection(&builder, data.Results)
	writePerformanceSection(&builder, data.PerformanceResult)

	fmt.Fprintf(&builder, "## Check Summary\n\n")
	fmt.Fprintf(&builder, "- Readiness checks: %d total, %d passed, %d failed\n", totalReadiness, passedReadiness, failedReadiness)
	fmt.Fprintf(&builder, "- Smoke checks: %d total, %d passed, %d failed\n", totalSmoke, passedSmoke, failedSmoke)
	fmt.Fprintf(&builder, "\n")

	fmt.Fprintf(&builder, "## Configuration\n\n")
	fmt.Fprintf(&builder, "- Cleanup enabled: `%t`\n", cfg.Settings.Cleanup)
	fmt.Fprintf(&builder, "- Timeout seconds: `%d`\n", cfg.Settings.TimeoutSeconds)
	fmt.Fprintf(&builder, "- Health path: `%s`\n", cfg.Service.HealthPath)
	fmt.Fprintf(&builder, "\n")

	if data.Logs != "" {
		fmt.Fprintf(&builder, "## Container Logs\n\n")
		fmt.Fprintf(&builder, "```txt\n")
		fmt.Fprintf(&builder, "%s", data.Logs)
		fmt.Fprintf(&builder, "\n```\n\n")
	}

	return builder.String()
}

func writeBuildSection(builder *strings.Builder, result BuildResult) {
	fmt.Fprintf(builder, "## Build Check\n\n")

	if !result.Enabled {
		fmt.Fprintf(builder, "No image build was configured.\n\n")
		return
	}

	status := "FAIL"
	if result.Passed {
		status = "PASS"
	}

	errorText := result.Error
	if errorText == "" {
		errorText = "-"
	}

	fmt.Fprintf(builder, "| Step | Image | Context | Dockerfile | Result | Error |\n")
	fmt.Fprintf(builder, "|---|---|---|---|---|---|\n")

	fmt.Fprintf(
		builder,
		"| docker build | `%s` | `%s` | `%s` | %s | %s |\n\n",
		escapeMarkdownTable(result.Image),
		escapeMarkdownTable(result.Context),
		escapeMarkdownTable(result.Dockerfile),
		status,
		escapeMarkdownTable(errorText),
	)

	if result.Output != "" && !result.Passed {
		fmt.Fprintf(builder, "### Build Output\n\n")
		fmt.Fprintf(builder, "```txt\n")
		builder.WriteString(result.Output)
		fmt.Fprintf(builder, "\n```\n\n")
	}
}

func writeReadinessSection(builder *strings.Builder, results []ReadinessResult) {
	fmt.Fprintf(builder, "## Readiness Checks\n\n")

	if len(results) == 0 {
		fmt.Fprintf(builder, "No readiness checks were recorded.\n\n")
		return
	}

	fmt.Fprintf(builder, "| Check | Target | Result | Error |\n")
	fmt.Fprintf(builder, "|---|---|---|---|\n")

	for _, result := range results {
		status := "FAIL"
		if result.Passed {
			status = "PASS"
		}

		errorText := result.Error
		if errorText == "" {
			errorText = "-"
		}

		fmt.Fprintf(builder, "| %s | %s | %s | %s |\n",
			escapeMarkdownTable(result.Name),
			escapeMarkdownTable(result.Target),
			status,
			escapeMarkdownTable(errorText),
		)
	}

	fmt.Fprintf(builder, "\n")
}

func writeSmokeSection(builder *strings.Builder, results []checker.SmokeResult) {
	fmt.Fprintf(builder, "## Smoke Checks\n\n")

	if len(results) == 0 {
		fmt.Fprintf(builder, "No smoke checks were executed.\n\n")
		return
	}

	fmt.Fprintf(builder, "| Check | Method | URL | Expected | Actual | Duration | Result | Error |\n")
	fmt.Fprintf(builder, "|---|---|---|---:|---:|---:|---|---|\n")

	for _, result := range results {
		checkStatus := "FAIL"
		if result.Passed {
			checkStatus = "PASS"
		}

		errorText := result.Error
		if errorText == "" {
			errorText = "-"
		}

		fmt.Fprintf(builder, "| %s | %s | %s | %d | %d | %s | %s | %s |\n",
			escapeMarkdownTable(result.Name),
			result.Method,
			result.URL,
			result.ExpectedStatus,
			result.ActualStatus,
			result.Duration.Round(time.Millisecond),
			checkStatus,
			escapeMarkdownTable(errorText),
		)
	}

	fmt.Fprintf(builder, "\n")
}

func writePerformanceSection(builder *strings.Builder, result PerformanceResult) {
	builder.WriteString("## Performance Check\n\n")

	if !result.Enabled {
		builder.WriteString("No performance check was configured.\n\n")
		return
	}

	status := "FAIL"
	if result.Passed {
		status = "PASS"
	}

	builder.WriteString("### Summary\n\n")
	fmt.Fprintf(builder, "- VUs: `%d`\n", result.VUs)
	fmt.Fprintf(builder, "- Duration: `%s`\n", result.Duration)
	fmt.Fprintf(builder, "- Result: **%s**\n\n", status)

	builder.WriteString("### Latency Metrics\n\n")
	builder.WriteString("| Metric | Value | Threshold | Result |\n")
	builder.WriteString("|---|---:|---:|---|\n")

	if result.Error != "" {
		fmt.Fprintf(builder, "| avg latency | N/A | - | N/A |\n")
		fmt.Fprintf(builder, "| min latency | N/A | - | N/A |\n")
		fmt.Fprintf(builder, "| median latency | N/A | - | N/A |\n")
		fmt.Fprintf(builder, "| max latency | N/A | - | N/A |\n")
		fmt.Fprintf(builder, "| p90 latency | N/A | - | N/A |\n")
		fmt.Fprintf(builder, "| p95 latency | N/A | %.2fms | N/A |\n\n", result.MaxP95LatencyMs)
	} else {
		p95Status := "PASS"
		if result.P95LatencyMs > result.MaxP95LatencyMs {
			p95Status = "FAIL"
		}

		fmt.Fprintf(builder, "| avg latency | %.2fms | - | - |\n", result.AvgLatencyMs)
		fmt.Fprintf(builder, "| min latency | %.2fms | - | - |\n", result.MinLatencyMs)
		fmt.Fprintf(builder, "| median latency | %.2fms | - | - |\n", result.MedianLatencyMs)
		fmt.Fprintf(builder, "| max latency | %.2fms | - | - |\n", result.MaxLatencyMs)
		fmt.Fprintf(builder, "| p90 latency | %.2fms | - | - |\n", result.P90LatencyMs)
		fmt.Fprintf(
			builder,
			"| p95 latency | %.2fms | %.2fms | %s |\n\n",
			result.P95LatencyMs,
			result.MaxP95LatencyMs,
			p95Status,
		)
	}

	builder.WriteString("### Reliability Metrics\n\n")
	builder.WriteString("| Metric | Value | Threshold | Result |\n")
	builder.WriteString("|---|---:|---:|---|\n")

	if result.Error != "" {
		fmt.Fprintf(builder, "| error rate | N/A | %.4f | N/A |\n", result.MaxErrorRate)
		fmt.Fprintf(builder, "| request count | N/A | - | N/A |\n")
		fmt.Fprintf(builder, "| iterations | N/A | - | N/A |\n")
		fmt.Fprintf(builder, "| checks total | N/A | - | N/A |\n")
		fmt.Fprintf(builder, "| check pass rate | N/A | 1.0000 | N/A |\n\n")
	} else {
		errorRateStatus := "PASS"
		if result.ErrorRate > result.MaxErrorRate {
			errorRateStatus = "FAIL"
		}

		fmt.Fprintf(
			builder,
			"| error rate | %.4f | %.4f | %s |\n",
			result.ErrorRate,
			result.MaxErrorRate,
			errorRateStatus,
		)
		fmt.Fprintf(builder, "| request count | %d | - | - |\n", result.RequestCount)
		fmt.Fprintf(builder, "| iterations | %d | - | - |\n", result.Iterations)

		checkStatus := "PASS"
		if result.CheckPassRate < 1 {
			checkStatus = "FAIL"
		}

		fmt.Fprintf(builder, "| checks total | %d | - | - |\n", result.ChecksTotal)
		fmt.Fprintf(builder, "| check pass rate | %.4f | 1.0000 | %s |\n\n", result.CheckPassRate, checkStatus)
	}

	if result.Error != "" {
		builder.WriteString("### Performance Error\n\n")
		builder.WriteString("```txt\n")
		builder.WriteString(result.Error)
		builder.WriteString("\n```\n\n")
	}

	if result.Output != "" && !result.Passed {
		builder.WriteString("### k6 Output\n\n")
		builder.WriteString("```txt\n")
		builder.WriteString(result.Output)
		builder.WriteString("\n```\n\n")
	}
}

func buildReportFilename(serviceName string, timestamp time.Time, extension string) string {
	cleanName := strings.ToLower(serviceName)
	cleanName = strings.ReplaceAll(cleanName, "_", "-")
	cleanName = strings.ReplaceAll(cleanName, " ", "-")

	return fmt.Sprintf(
		"predeploy-%s-%s.%s",
		cleanName,
		timestamp.Format("2006-01-02-150405"),
		extension,
	)
}

func escapeMarkdownTable(value string) string {
	value = strings.ReplaceAll(value, "|", "\\|")
	value = strings.ReplaceAll(value, "\n", " ")
	value = strings.ReplaceAll(value, "\r", " ")
	return value
}
