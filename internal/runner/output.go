package runner

import (
	"fmt"
	"strings"

	"github.com/HXong/predeploy-guard/internal/checker"
	"github.com/HXong/predeploy-guard/internal/workload"
)

func printSmokeResult(result checker.SmokeResult) {
	status := "FAIL"
	if result.Passed {
		status = "PASS"
	}

	if result.Error != "" {
		fmt.Printf("[%s] %s %s %s error=%s duration=%s\n",
			status,
			result.Name,
			result.Method,
			result.URL,
			result.Error,
			result.Duration,
		)
		return
	}

	fmt.Printf("[%s] %s %s %s expected=%d actual=%d duration=%s\n",
		status,
		result.Name,
		result.Method,
		result.URL,
		result.ExpectedStatus,
		result.ActualStatus,
		result.Duration,
	)
}

func printWorkloadResult(result workload.Result) {
	if !result.Enabled {
		return
	}

	status := "FAIL"
	if result.Passed {
		status = "PASS"
	} else if result.FailurePolicy == "warn" {
		status = "WARN"
	}

	fmt.Printf(
		"[%s] %s requests=%d success=%d failed=%d duration=%s\n",
		status,
		result.Name,
		result.RequestCount,
		result.SuccessCount,
		result.FailureCount,
		result.Duration,
	)
}

func joinURL(baseURL string, path string) string {
	baseURL = strings.TrimRight(baseURL, "/")
	path = strings.TrimLeft(path, "/")
	return baseURL + "/" + path
}
