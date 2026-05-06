package runner

import (
	"fmt"

	"github.com/HXong/predeploy-guard/internal/checker"
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
