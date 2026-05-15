package main

import (
	"fmt"
	"time"

	"github.com/HXong/predeploy-guard/internal/history"
)

func printRunSummary(summary history.RunSummary) {
	status := "FAIL"
	if summary.Passed {
		status = "PASS"
	}

	profile := summary.Profile
	if profile == "" {
		profile = "base"
	}

	fmt.Println("Run Details")
	fmt.Println()
	fmt.Printf("Run ID: %s\n", summary.RunID)
	fmt.Printf("Service: %s\n", summary.ServiceName)
	fmt.Printf("Profile: %s\n", profile)
	fmt.Printf("Result: %s\n", status)
	fmt.Printf("Started: %s\n", summary.StartedAt.Format(time.RFC3339))
	fmt.Printf("Finished: %s\n", summary.FinishedAt.Format(time.RFC3339))

	duration := summary.FinishedAt.Sub(summary.StartedAt)
	if duration > 0 {
		fmt.Printf("Duration: %s\n", duration.Round(time.Millisecond))
	}

	fmt.Println()

	if summary.P95LatencyMs > 0 || summary.ErrorRate > 0 {
		fmt.Println("Performance")
		if summary.P95LatencyMs > 0 {
			fmt.Printf("  p95 latency: %.2fms\n", summary.P95LatencyMs)
		}
		fmt.Printf("  error rate: %.4f\n", summary.ErrorRate)
		fmt.Println()
	}

	fmt.Println("Reports")
	fmt.Printf("  Markdown: %s\n", summary.MarkdownPath)
	fmt.Printf("  JSON: %s\n", summary.JSONPath)
}
