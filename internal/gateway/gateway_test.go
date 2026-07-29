package gateway

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/HXong/predeploy-guard/internal/config"
)

func TestRunChecksPassesWhenGatewayAndDirectMatch(t *testing.T) {
	gatewayServer := statusServer(http.StatusOK)
	defer gatewayServer.Close()
	directServer := statusServer(http.StatusOK)
	defer directServer.Close()

	results := RunChecks(context.Background(), gatewayConfig(gatewayServer.URL), directServer.URL)

	if len(results) != 1 || !results[0].Passed {
		t.Fatalf("results = %#v, want one passing result", results)
	}
	if !results[0].StatusMatched || results[0].GatewayStatus != http.StatusOK ||
		results[0].DirectStatus != http.StatusOK {
		t.Fatalf("result = %#v, want matching HTTP 200 statuses", results[0])
	}
	if !results[0].LatencyPassed || results[0].GatewayLatencyMs <= 0 ||
		results[0].DirectLatencyMs <= 0 {
		t.Fatalf("result = %#v, want populated non-enforcing latency values", results[0])
	}
}

func TestRunChecksFailsWhenGatewayReturnsUnexpectedStatus(t *testing.T) {
	gatewayServer := statusServer(http.StatusBadGateway)
	defer gatewayServer.Close()
	directServer := statusServer(http.StatusOK)
	defer directServer.Close()

	results := RunChecks(context.Background(), gatewayConfig(gatewayServer.URL), directServer.URL)

	if len(results) != 1 || results[0].Passed {
		t.Fatalf("results = %#v, want one failing result", results)
	}
	if results[0].GatewayPassed {
		t.Fatalf("gateway passed = true, want false")
	}
}

func TestRunChecksFailsWhenDirectStatusDiffers(t *testing.T) {
	gatewayServer := statusServer(http.StatusOK)
	defer gatewayServer.Close()
	directServer := statusServer(http.StatusNoContent)
	defer directServer.Close()

	results := RunChecks(context.Background(), gatewayConfig(gatewayServer.URL), directServer.URL)

	if len(results) != 1 || results[0].Passed {
		t.Fatalf("results = %#v, want one failing result", results)
	}
	if results[0].StatusMatched || results[0].DirectPassed {
		t.Fatalf("result = %#v, want failed direct comparison", results[0])
	}
}

func TestRunChecksSkipsDirectRequestWhenComparisonDisabled(t *testing.T) {
	gatewayServer := statusServer(http.StatusOK)
	defer gatewayServer.Close()
	compareDirect := false
	cfg := gatewayConfig(gatewayServer.URL)
	cfg.Gateway.Routes[0].CompareDirect = &compareDirect

	results := RunChecks(context.Background(), cfg, "://invalid-direct-url")

	if len(results) != 1 || !results[0].Passed {
		t.Fatalf("results = %#v, want gateway-only check to pass", results)
	}
	if results[0].CompareDirect || results[0].DirectURL != "" {
		t.Fatalf("result = %#v, want direct comparison omitted", results[0])
	}
}

func TestRunChecksUntilReadyRetriesUntilGatewayPasses(t *testing.T) {
	var attempts atomic.Int32
	gatewayServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if attempts.Add(1) < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer gatewayServer.Close()
	directServer := statusServer(http.StatusOK)
	defer directServer.Close()

	results := runChecksUntilReady(
		context.Background(),
		gatewayConfig(gatewayServer.URL),
		directServer.URL,
		500*time.Millisecond,
		5*time.Millisecond,
	)

	if len(results) != 1 || !results[0].Passed {
		t.Fatalf("results = %#v, want final retry to pass", results)
	}
	if attempts.Load() < 3 {
		t.Fatalf("attempts = %d, want at least 3", attempts.Load())
	}
}

func TestRunChecksUntilReadyReturnsFinalFailureAtTimeout(t *testing.T) {
	var attempts atomic.Int32
	gatewayServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		attempts.Add(1)
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer gatewayServer.Close()
	directServer := statusServer(http.StatusOK)
	defer directServer.Close()

	results := runChecksUntilReady(
		context.Background(),
		gatewayConfig(gatewayServer.URL),
		directServer.URL,
		150*time.Millisecond,
		5*time.Millisecond,
	)

	if len(results) != 1 || results[0].Passed {
		t.Fatalf("results = %#v, want final retry to fail", results)
	}
	if results[0].GatewayStatus != http.StatusServiceUnavailable {
		t.Fatalf("gateway status = %d, want 503", results[0].GatewayStatus)
	}
	if attempts.Load() < 2 {
		t.Fatalf("attempts = %d, want multiple attempts before timeout", attempts.Load())
	}
}

func TestLatencyAnalysisPassesUnderThresholdsAndComputesOverhead(t *testing.T) {
	result := passingComparedResult(40*time.Millisecond, 10*time.Millisecond)

	finalizeRouteResult(&result, config.GatewayLatencyConfig{
		Enabled:             true,
		MaxGatewayLatencyMs: 100,
		MaxOverheadMs:       50,
		MaxOverheadRatio:    5,
		FailurePolicy:       "fail",
	})

	if !result.Passed || !result.LatencyPassed || !result.LatencyCompared {
		t.Fatalf("result = %#v, want passing latency comparison", result)
	}
	if result.OverheadMs != 30 || result.OverheadRatio != 4 {
		t.Fatalf(
			"overhead = %.2fms ratio %.2fx, want 30ms and 4x",
			result.OverheadMs,
			result.OverheadRatio,
		)
	}
}

func TestAbsoluteGatewayLatencyWorksWithoutDirectComparison(t *testing.T) {
	result := RouteResult{
		GatewayPassed:   true,
		GatewayDuration: 40 * time.Millisecond,
	}

	finalizeRouteResult(&result, config.GatewayLatencyConfig{
		Enabled:             true,
		MaxGatewayLatencyMs: 50,
		FailurePolicy:       "fail",
	})

	if !result.Passed || !result.LatencyPassed {
		t.Fatalf("result = %#v, want absolute gateway latency pass", result)
	}
	if result.LatencyCompared || result.DirectLatencyMs != 0 {
		t.Fatalf("result = %#v, want no direct latency comparison", result)
	}
}

func TestLatencyWarningDoesNotFailCorrectRoute(t *testing.T) {
	result := passingComparedResult(80*time.Millisecond, 10*time.Millisecond)

	finalizeRouteResult(&result, config.GatewayLatencyConfig{
		Enabled:          true,
		MaxOverheadMs:    20,
		MaxOverheadRatio: 3,
		FailurePolicy:    "warn",
	})

	if !result.Passed || result.LatencyPassed {
		t.Fatalf("result = %#v, want passing route with latency warning", result)
	}
	if len(result.LatencyWarnings) != 2 || len(result.LatencyErrors) != 0 {
		t.Fatalf("warnings/errors = %#v/%#v, want two warnings", result.LatencyWarnings, result.LatencyErrors)
	}
}

func TestLatencyFailureFailsCorrectRoute(t *testing.T) {
	result := passingComparedResult(80*time.Millisecond, 10*time.Millisecond)

	finalizeRouteResult(&result, config.GatewayLatencyConfig{
		Enabled:          true,
		MaxOverheadRatio: 3,
		FailurePolicy:    "fail",
	})

	if result.Passed || result.LatencyPassed {
		t.Fatalf("result = %#v, want failed latency policy", result)
	}
	if len(result.LatencyErrors) != 1 || len(result.LatencyWarnings) != 0 {
		t.Fatalf("warnings/errors = %#v/%#v, want one error", result.LatencyWarnings, result.LatencyErrors)
	}
}

func TestUnavailableOverheadComparisonUsesFailurePolicy(t *testing.T) {
	for _, policy := range []string{"warn", "fail"} {
		t.Run(policy, func(t *testing.T) {
			result := RouteResult{
				GatewayPassed:   true,
				GatewayDuration: 20 * time.Millisecond,
			}

			finalizeRouteResult(&result, config.GatewayLatencyConfig{
				Enabled:       true,
				MaxOverheadMs: 10,
				FailurePolicy: policy,
			})

			if result.LatencyPassed || result.LatencyCompared {
				t.Fatalf("result = %#v, want unavailable overhead comparison", result)
			}
			if policy == "warn" {
				if !result.Passed || len(result.LatencyWarnings) != 1 {
					t.Fatalf("result = %#v, want warning without route failure", result)
				}
			} else if result.Passed || len(result.LatencyErrors) != 1 {
				t.Fatalf("result = %#v, want latency policy failure", result)
			}
		})
	}
}

func TestIngressRetryStopsAfterCorrectnessPassesDespiteLatencyFailure(t *testing.T) {
	var attempts atomic.Int32
	gatewayServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		attempts.Add(1)
		time.Sleep(20 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer gatewayServer.Close()
	directServer := statusServer(http.StatusOK)
	defer directServer.Close()
	cfg := gatewayConfig(gatewayServer.URL)
	cfg.Gateway.Latency = config.GatewayLatencyConfig{
		Enabled:             true,
		MaxGatewayLatencyMs: 1,
		FailurePolicy:       "fail",
	}

	results := runChecksUntilReady(
		context.Background(),
		cfg,
		directServer.URL,
		500*time.Millisecond,
		5*time.Millisecond,
	)

	if len(results) != 1 || results[0].Passed || !AllCorrectnessPassed(results) {
		t.Fatalf("results = %#v, want correctness pass with latency failure", results)
	}
	if attempts.Load() != 1 {
		t.Fatalf("attempts = %d, want no retry for latency-only failure", attempts.Load())
	}
}

func passingComparedResult(gatewayDuration time.Duration, directDuration time.Duration) RouteResult {
	return RouteResult{
		GatewayPassed:   true,
		GatewayDuration: gatewayDuration,
		CompareDirect:   true,
		DirectPassed:    true,
		DirectDuration:  directDuration,
		StatusMatched:   true,
	}
}

func statusServer(status int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(status)
	}))
}

func gatewayConfig(baseURL string) *config.Config {
	compareDirect := true
	return &config.Config{
		Gateway: config.GatewayConfig{
			Enabled: true,
			BaseURL: baseURL,
			Routes: []config.GatewayRoute{{
				Name:           "homepage-via-gateway",
				Method:         http.MethodGet,
				Path:           "/",
				ExpectedStatus: http.StatusOK,
				CompareDirect:  &compareDirect,
			}},
		},
	}
}
