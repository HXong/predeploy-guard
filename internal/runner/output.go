package runner

import (
	"fmt"
	"strings"

	"github.com/HXong/predeploy-guard/internal/checker"
	"github.com/HXong/predeploy-guard/internal/gateway"
	"github.com/HXong/predeploy-guard/internal/workload"
)

func printGatewayResult(result gateway.RouteResult) {
	status := "FAIL"
	if result.Passed && len(result.LatencyWarnings) > 0 {
		status = "WARN"
	} else if result.Passed {
		status = "PASS"
	}

	directStatus := ""
	if result.CompareDirect {
		directStatus = fmt.Sprintf(" direct=%d", result.DirectStatus)
	}

	latencyText := fmt.Sprintf(" gatewayLatency=%.2fms", result.GatewayLatencyMs)
	if result.CompareDirect && result.DirectError == "" {
		latencyText += fmt.Sprintf(" directLatency=%.2fms", result.DirectLatencyMs)
	}
	if result.LatencyCompared {
		latencyText += fmt.Sprintf(
			" overhead=%.2fms ratio=%.2fx",
			result.OverheadMs,
			result.OverheadRatio,
		)
	}

	errorText := ""
	if result.GatewayError != "" {
		errorText += " gatewayError=" + result.GatewayError
	}
	if result.DirectError != "" {
		errorText += " directError=" + result.DirectError
	}
	if len(result.LatencyWarnings) > 0 {
		errorText += fmt.Sprintf(
			" latencyWarning=%q",
			strings.Join(result.LatencyWarnings, "; "),
		)
	}
	if len(result.LatencyErrors) > 0 {
		errorText += fmt.Sprintf(
			" latencyError=%q",
			strings.Join(result.LatencyErrors, "; "),
		)
	}

	fmt.Printf(
		"[%s] %s gateway=%d%s%s%s\n",
		status,
		result.Name,
		result.GatewayStatus,
		directStatus,
		latencyText,
		errorText,
	)
}

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
