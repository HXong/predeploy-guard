package main

import (
	"fmt"
	"math"
	"time"

	"github.com/HXong/predeploy-guard/internal/history"
)

func printRunComparison(base history.RunSummary, target history.RunSummary) {
	fmt.Println("Run Comparison")
	fmt.Println()

	fmt.Printf("Base:   %s\n", base.RunID)
	fmt.Printf("Target: %s\n", target.RunID)
	fmt.Println()

	fmt.Printf("Service: %s → %s\n", base.ServiceName, target.ServiceName)
	fmt.Printf("Profile: %s → %s\n", displayProfile(base.Profile), displayProfile(target.Profile))
	fmt.Printf("Result: %s → %s\n", displayPassFail(base.Passed), displayPassFail(target.Passed))
	fmt.Printf("Started: %s → %s\n",
		base.StartedAt.Format(time.RFC3339),
		target.StartedAt.Format(time.RFC3339),
	)

	baseDuration := base.FinishedAt.Sub(base.StartedAt)
	targetDuration := target.FinishedAt.Sub(target.StartedAt)

	if baseDuration > 0 || targetDuration > 0 {
		fmt.Printf("Duration: %s → %s\n",
			baseDuration.Round(time.Millisecond),
			targetDuration.Round(time.Millisecond),
		)
	}

	fmt.Println()
	printPerformanceComparison(base, target)

	fmt.Println()
	fmt.Println("Reports")
	fmt.Printf("  Base Markdown: %s\n", base.MarkdownPath)
	fmt.Printf("  Base JSON: %s\n", base.JSONPath)
	fmt.Printf("  Target Markdown: %s\n", target.MarkdownPath)
	fmt.Printf("  Target JSON: %s\n", target.JSONPath)
}

func printPerformanceComparison(base history.RunSummary, target history.RunSummary) {
	fmt.Println("Performance")

	if base.P95LatencyMs == 0 && target.P95LatencyMs == 0 {
		fmt.Println("  p95 latency: not available")
	} else {
		delta := target.P95LatencyMs - base.P95LatencyMs
		if base.P95LatencyMs == 0 {
			fmt.Printf("  p95 latency: %.2fms → %.2fms (%s, N/A)\n",
				base.P95LatencyMs,
				target.P95LatencyMs,
				formatSignedFloat(delta, "ms"),
			)
		} else {
			percent := percentageChange(base.P95LatencyMs, target.P95LatencyMs)

			fmt.Printf("  p95 latency: %.2fms → %.2fms (%s, %s)\n",
				base.P95LatencyMs,
				target.P95LatencyMs,
				formatSignedFloat(delta, "ms"),
				formatSignedPercent(percent),
			)
		}
	}

	errorDelta := target.ErrorRate - base.ErrorRate
	errorPercent := percentageChange(base.ErrorRate, target.ErrorRate)

	fmt.Printf("  error rate: %.4f → %.4f (%s, %s)\n",
		base.ErrorRate,
		target.ErrorRate,
		formatSignedFloat(errorDelta, ""),
		formatSignedPercent(errorPercent),
	)
}

func displayProfile(profile string) string {
	if profile == "" {
		return "base"
	}

	return profile
}

func displayPassFail(passed bool) string {
	if passed {
		return "PASS"
	}

	return "FAIL"
}

func percentageChange(base float64, target float64) float64 {
	if base == 0 {
		if target == 0 {
			return 0
		}

		return math.Inf(1)
	}

	return ((target - base) / base) * 100
}

func formatSignedFloat(value float64, suffix string) string {
	sign := "+"
	if value < 0 {
		sign = ""
	}

	return fmt.Sprintf("%s%.2f%s", sign, value, suffix)
}

func formatSignedPercent(value float64) string {
	if math.IsInf(value, 1) {
		return "+Inf%"
	}

	sign := "+"
	if value < 0 {
		sign = ""
	}

	return fmt.Sprintf("%s%.2f%%", sign, value)
}
