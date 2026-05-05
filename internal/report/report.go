package report

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/HXong/predeploy-guard/internal/checker"
	"github.com/HXong/predeploy-guard/internal/config"
)

type ReportData struct {
	ServiceName string
	Image       string
	Runtime     string
	BaseURL     string
	StartedAt   time.Time
	FinishedAt  time.Time
	Results     []checker.SmokeResult
	Passed      bool
}

func WriteMarkdown(cfg *config.Config, data ReportData) (string, error) {
	if err := os.MkdirAll("reports", 0755); err != nil {
		return "", fmt.Errorf("create reports directory: %w", err)
	}

	filename := buildReportFilename(cfg.Service.Name, data.StartedAt)
	reportPath := filepath.Join("reports", filename)

	content := buildMarkdown(cfg, data)

	if err := os.WriteFile(reportPath, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("write markdown report: %w", err)
	}

	return reportPath, nil
}

func buildMarkdown(cfg *config.Config, data ReportData) string {
	var builder strings.Builder

	status := "FAIL"
	if data.Passed {
		status = "PASS"
	}

	total := len(data.Results)
	passed := 0
	failed := 0

	for _, result := range data.Results {
		if result.Passed {
			passed++
		} else {
			failed++
		}
	}

	duration := data.FinishedAt.Sub(data.StartedAt)

	builder.WriteString("# Predeploy Guard Report\n\n")

	builder.WriteString("## Summary\n\n")
	builder.WriteString(fmt.Sprintf("- **Service:** %s\n", data.ServiceName))
	builder.WriteString(fmt.Sprintf("- **Image:** `%s`\n", data.Image))
	builder.WriteString(fmt.Sprintf("- **Runtime:** %s\n", data.Runtime))
	builder.WriteString(fmt.Sprintf("- **Base URL:** %s\n", data.BaseURL))
	builder.WriteString(fmt.Sprintf("- **Result:** **%s**\n", status))
	builder.WriteString(fmt.Sprintf("- **Started At:** %s\n", data.StartedAt.Format(time.RFC3339)))
	builder.WriteString(fmt.Sprintf("- **Finished At:** %s\n", data.FinishedAt.Format(time.RFC3339)))
	builder.WriteString(fmt.Sprintf("- **Total Duration:** %s\n", duration.Round(time.Millisecond)))
	builder.WriteString("\n")

	builder.WriteString("## Smoke Checks\n\n")
	builder.WriteString("| Check | Method | URL | Expected | Actual | Duration | Result | Error |\n")
	builder.WriteString("|---|---|---|---:|---:|---:|---|---|\n")

	for _, result := range data.Results {
		checkStatus := "FAIL"
		if result.Passed {
			checkStatus = "PASS"
		}

		errorText := result.Error
		if errorText == "" {
			errorText = "-"
		}

		builder.WriteString(
			fmt.Sprintf(
				"| %s | %s | %s | %d | %d | %s | %s | %s |\n",
				escapeMarkdownTable(result.Name),
				result.Method,
				result.URL,
				result.ExpectedStatus,
				result.ActualStatus,
				result.Duration.Round(time.Millisecond),
				checkStatus,
				escapeMarkdownTable(errorText),
			))
	}

	builder.WriteString("\n")

	builder.WriteString("## Check Summary\n\n")
	builder.WriteString(fmt.Sprintf("- Total checks: %d\n", total))
	builder.WriteString(fmt.Sprintf("- Passed: %d\n", passed))
	builder.WriteString(fmt.Sprintf("- Failed: %d\n", failed))
	builder.WriteString("\n")

	builder.WriteString("## Configuration\n\n")
	builder.WriteString(fmt.Sprintf("- Cleanup enabled: `%t`\n", cfg.Settings.Cleanup))
	builder.WriteString(fmt.Sprintf("- Timeout seconds: `%d`\n", cfg.Settings.TimeoutSeconds))
	builder.WriteString(fmt.Sprintf("- Health path: `%s`\n", cfg.Service.HealthPath))
	builder.WriteString("\n")

	return builder.String()
}

func buildReportFilename(serviceName string, timestamp time.Time) string {
	cleanName := strings.ToLower(serviceName)
	cleanName = strings.ReplaceAll(cleanName, "_", "-")
	cleanName = strings.ReplaceAll(cleanName, " ", "-")

	return fmt.Sprintf(
		"predeploy-%s-%s.md",
		cleanName,
		timestamp.Format("2006-01-02-150405"),
	)
}

func escapeMarkdownTable(value string) string {
	value = strings.ReplaceAll(value, "|", "\\|")
	value = strings.ReplaceAll(value, "\n", " ")
	value = strings.ReplaceAll(value, "\r", " ")
	return value
}
