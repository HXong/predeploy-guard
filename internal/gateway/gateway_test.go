package gateway

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

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
