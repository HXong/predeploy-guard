package gateway

import (
	"context"
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
	Passed        bool `json:"passed"`
}

func RunChecks(ctx context.Context, cfg *config.Config, directBaseURL string) []RouteResult {
	if !cfg.Gateway.Enabled {
		return nil
	}

	client := &http.Client{Timeout: 10 * time.Second}
	results := make([]RouteResult, 0, len(cfg.Gateway.Routes))
	for _, route := range cfg.Gateway.Routes {
		results = append(results, runRoute(ctx, client, cfg.Gateway.BaseURL, directBaseURL, route))
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
		if AllPassed(attemptResults) {
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

func runRoute(
	ctx context.Context,
	client *http.Client,
	gatewayBaseURL string,
	directBaseURL string,
	route config.GatewayRoute,
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

	result.Passed = result.GatewayPassed &&
		(!compareDirect || (result.DirectPassed && result.StatusMatched))
	return result
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
