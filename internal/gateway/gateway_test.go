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
