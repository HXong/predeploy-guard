package gateway

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/HXong/predeploy-guard/internal/config"
)

type RouteResult struct {
	Name           string `json:"name"`
	Method         string `json:"method"`
	Path           string `json:"path"`
	ExpectedStatus int    `json:"expectedStatus"`

	GatewayURL      string        `json:"gatewayUrl"`
	GatewayStatus   int           `json:"gatewayStatus"`
	GatewayDuration time.Duration `json:"gatewayDuration"`
	GatewayPassed   bool          `json:"gatewayPassed"`
	GatewayError    string        `json:"gatewayError,omitempty"`

	CompareDirect bool `json:"compareDirect"`

	DirectURL      string        `json:"directUrl,omitempty"`
	DirectStatus   int           `json:"directStatus,omitempty"`
	DirectDuration time.Duration `json:"directDuration,omitempty"`
	DirectPassed   bool          `json:"directPassed"`
	DirectError    string        `json:"directError,omitempty"`

	StatusMatched bool `json:"statusMatched"`

	GatewayLatencyMs float64 `json:"gatewayLatencyMs"`
	DirectLatencyMs  float64 `json:"directLatencyMs,omitempty"`

	LatencyCompared bool    `json:"latencyCompared"`
	OverheadMs      float64 `json:"overheadMs,omitempty"`
	OverheadRatio   float64 `json:"overheadRatio,omitempty"`

	LatencyPassed   bool     `json:"latencyPassed"`
	LatencyWarnings []string `json:"latencyWarnings,omitempty"`
	LatencyErrors   []string `json:"latencyErrors,omitempty"`

	Passed bool `json:"passed"`
}

func RunChecks(ctx context.Context, cfg *config.Config, directBaseURL string) []RouteResult {
	if !cfg.Gateway.Enabled {
		return nil
	}

	client := &http.Client{Timeout: 10 * time.Second}
	results := make([]RouteResult, 0, len(cfg.Gateway.Routes))
	for _, route := range cfg.Gateway.Routes {
		results = append(results, runRoute(
			ctx,
			client,
			cfg.Gateway.BaseURL,
			directBaseURL,
			route,
			cfg.Gateway.Latency,
		))
	}
	return results
}

func RunChecksUntilReady(
	ctx context.Context,
	cfg *config.Config,
	directBaseURL string,
	timeout time.Duration,
) []RouteResult {
	return runChecksUntilReady(ctx, cfg, directBaseURL, timeout, time.Second)
}

func runChecksUntilReady(
	ctx context.Context,
	cfg *config.Config,
	directBaseURL string,
	timeout time.Duration,
	retryInterval time.Duration,
) []RouteResult {
	if !cfg.Gateway.Enabled {
		return nil
	}
	if timeout <= 0 {
		return RunChecks(ctx, cfg, directBaseURL)
	}
	if retryInterval <= 0 {
		retryInterval = time.Second
	}

	retryContext, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var lastCompletedResults []RouteResult
	for {
		attemptResults := RunChecks(retryContext, cfg, directBaseURL)
		if AllCorrectnessPassed(attemptResults) {
			return attemptResults
		}
		if retryContext.Err() != nil {
			if len(lastCompletedResults) > 0 {
				return lastCompletedResults
			}
			return attemptResults
		}
		lastCompletedResults = attemptResults

		timer := time.NewTimer(retryInterval)
		select {
		case <-retryContext.Done():
			timer.Stop()
			return lastCompletedResults
		case <-timer.C:
		}
	}
}

func AllPassed(results []RouteResult) bool {
	if len(results) == 0 {
		return false
	}
	for _, result := range results {
		if !result.Passed {
			return false
		}
	}
	return true
}

func AllCorrectnessPassed(results []RouteResult) bool {
	if len(results) == 0 {
		return false
	}
	for _, result := range results {
		if !correctnessPassed(result) {
			return false
		}
	}
	return true
}

func runRoute(
	ctx context.Context,
	client *http.Client,
	gatewayBaseURL string,
	directBaseURL string,
	route config.GatewayRoute,
	latency config.GatewayLatencyConfig,
) RouteResult {
	compareDirect := route.CompareDirect == nil || *route.CompareDirect
	result := RouteResult{
		Name:           route.Name,
		Method:         strings.ToUpper(route.Method),
		Path:           route.Path,
		ExpectedStatus: route.ExpectedStatus,
		GatewayURL:     joinURL(gatewayBaseURL, route.Path),
		CompareDirect:  compareDirect,
	}

	result.GatewayStatus, result.GatewayDuration, result.GatewayError = doRequest(
		ctx,
		client,
		result.Method,
		result.GatewayURL,
	)
	result.GatewayPassed = result.GatewayError == "" &&
		result.GatewayStatus == result.ExpectedStatus

	if compareDirect {
		result.DirectURL = joinURL(directBaseURL, route.Path)
		result.DirectStatus, result.DirectDuration, result.DirectError = doRequest(
			ctx,
			client,
			result.Method,
			result.DirectURL,
		)
		result.DirectPassed = result.DirectError == "" &&
			result.DirectStatus == result.ExpectedStatus
		result.StatusMatched = result.GatewayError == "" &&
			result.DirectError == "" &&
			result.GatewayStatus == result.DirectStatus
	}

	finalizeRouteResult(&result, latency)
	return result
}

func finalizeRouteResult(result *RouteResult, latency config.GatewayLatencyConfig) {
	result.GatewayLatencyMs = durationMilliseconds(result.GatewayDuration)
	if result.CompareDirect && result.DirectError == "" {
		result.DirectLatencyMs = durationMilliseconds(result.DirectDuration)
	}

	result.LatencyPassed = true
	if latency.Enabled {
		applyLatencyThresholds(result, latency)
	}

	result.Passed = correctnessPassed(*result) &&
		(strings.ToLower(latency.FailurePolicy) != "fail" || result.LatencyPassed)
}

func applyLatencyThresholds(result *RouteResult, latency config.GatewayLatencyConfig) {
	policy := strings.ToLower(latency.FailurePolicy)
	if policy == "" {
		policy = "warn"
	}

	recordViolation := func(message string) {
		result.LatencyPassed = false
		if policy == "fail" {
			result.LatencyErrors = append(result.LatencyErrors, message)
		} else {
			result.LatencyWarnings = append(result.LatencyWarnings, message)
		}
	}

	if latency.MaxGatewayLatencyMs > 0 &&
		result.GatewayLatencyMs > float64(latency.MaxGatewayLatencyMs) {
		recordViolation(fmt.Sprintf(
			"gateway latency %.2fms exceeded maximum %dms",
			result.GatewayLatencyMs,
			latency.MaxGatewayLatencyMs,
		))
	}

	result.LatencyCompared = result.CompareDirect &&
		result.GatewayError == "" &&
		result.DirectError == ""
	overheadConfigured := latency.MaxOverheadMs > 0 || latency.MaxOverheadRatio > 0
	if result.LatencyCompared {
		result.OverheadMs = result.GatewayLatencyMs - result.DirectLatencyMs
		directLatencyForRatio := math.Max(result.DirectLatencyMs, 0.001)
		result.OverheadRatio = result.GatewayLatencyMs / directLatencyForRatio

		if latency.MaxOverheadMs > 0 &&
			result.OverheadMs > float64(latency.MaxOverheadMs) {
			recordViolation(fmt.Sprintf(
				"gateway overhead %.2fms exceeded maximum %dms",
				result.OverheadMs,
				latency.MaxOverheadMs,
			))
		}
		if latency.MaxOverheadRatio > 0 &&
			result.OverheadRatio > latency.MaxOverheadRatio {
			recordViolation(fmt.Sprintf(
				"gateway overhead ratio %.2fx exceeded maximum %.2fx",
				result.OverheadRatio,
				latency.MaxOverheadRatio,
			))
		}
	} else if overheadConfigured {
		recordViolation(
			"gateway overhead comparison unavailable; compareDirect must be enabled and both requests must complete",
		)
	}
}

func correctnessPassed(result RouteResult) bool {
	return result.GatewayPassed &&
		(!result.CompareDirect || (result.DirectPassed && result.StatusMatched))
}

func durationMilliseconds(duration time.Duration) float64 {
	return float64(duration) / float64(time.Millisecond)
}

func doRequest(
	ctx context.Context,
	client *http.Client,
	method string,
	url string,
) (int, time.Duration, string) {
	startedAt := time.Now()
	request, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return 0, time.Since(startedAt), err.Error()
	}

	response, err := client.Do(request)
	if err != nil {
		return 0, time.Since(startedAt), err.Error()
	}
	defer response.Body.Close()

	return response.StatusCode, time.Since(startedAt), ""
}

func joinURL(baseURL string, path string) string {
	return strings.TrimRight(baseURL, "/") + "/" + strings.TrimLeft(path, "/")
}
