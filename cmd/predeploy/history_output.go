package main

import (
	"fmt"
	"time"

	"github.com/HXong/predeploy-guard/internal/history"
)

func printHistory(summaries []history.RunSummary) {
	fmt.Println("Run History")
	fmt.Println()

	if len(summaries) == 0 {
		fmt.Println("No previous runs found.")
		return
	}

	for i := len(summaries) - 1; i >= 0; i-- {
		summary := summaries[i]

		status := "FAIL"
		if summary.Passed {
			status = "PASS"
		}

		fmt.Println(summary.RunID)
		fmt.Printf("  Service: %s\n", summary.ServiceName)

		if summary.Profile != "" {
			fmt.Printf("  Profile: %s\n", summary.Profile)
		} else {
			fmt.Println("  Profile: base")
		}

		fmt.Printf("  Result: %s\n", status)
		fmt.Printf("  Started: %s\n", summary.StartedAt.Format(time.RFC3339))
		fmt.Printf("  Finished: %s\n", summary.FinishedAt.Format(time.RFC3339))

		if summary.P95LatencyMs > 0 {
			fmt.Printf("  p95 latency: %.2fms\n", summary.P95LatencyMs)
		}

		fmt.Printf("  error rate: %.4f\n", summary.ErrorRate)
		fmt.Printf("  Markdown: %s\n", summary.MarkdownPath)
		fmt.Printf("  JSON: %s\n", summary.JSONPath)
		fmt.Println()
	}
}
