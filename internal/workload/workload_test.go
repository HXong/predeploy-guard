package workload

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/HXong/predeploy-guard/internal/config"
)

func TestRunAllHTTPWorkloadPasses(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet {
			t.Errorf("method = %q, want GET", request.Method)
		}
		if request.URL.Path != "/ready" {
			t.Errorf("path = %q, want /ready", request.URL.Path)
		}
		writer.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := validatedWorkloadConfig(t, config.WorkloadConfig{
		Name:          "warmup-traffic",
		Type:          "http",
		Path:          "/ready",
		Duration:      "60ms",
		RatePerSecond: 50,
	})

	results := RunAll(context.Background(), cfg, server.URL)
	if len(results) != 1 {
		t.Fatalf("result count = %d, want 1", len(results))
	}

	result := results[0]
	if !result.Passed {
		t.Fatalf("result passed = false, error = %q", result.Error)
	}
	if result.RequestCount == 0 {
		t.Fatal("request count = 0, want at least one request")
	}
	if result.SuccessCount != result.RequestCount {
		t.Fatalf("success count = %d, want request count %d", result.SuccessCount, result.RequestCount)
	}
	if result.FailureCount != 0 {
		t.Fatalf("failure count = %d, want 0", result.FailureCount)
	}
	if result.StatusCounts[http.StatusOK] != result.RequestCount {
		t.Fatalf("HTTP 200 count = %d, want %d", result.StatusCounts[http.StatusOK], result.RequestCount)
	}
}

func TestRunAllHTTPWorkloadRecordsUnexpectedStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusTeapot)
	}))
	defer server.Close()

	cfg := validatedWorkloadConfig(t, config.WorkloadConfig{
		Name:           "warmup-traffic",
		Type:           "http",
		Path:           "/",
		Duration:       "40ms",
		RatePerSecond:  50,
		ExpectedStatus: http.StatusOK,
	})

	result := RunAll(context.Background(), cfg, server.URL)[0]
	if result.Passed {
		t.Fatal("result passed = true, want false")
	}
	if result.FailureCount == 0 {
		t.Fatal("failure count = 0, want at least one failure")
	}
	if result.SuccessCount != 0 {
		t.Fatalf("success count = %d, want 0", result.SuccessCount)
	}
	if result.StatusCounts[http.StatusTeapot] != result.RequestCount {
		t.Fatalf(
			"HTTP 418 count = %d, want request count %d",
			result.StatusCounts[http.StatusTeapot],
			result.RequestCount,
		)
	}
}

func TestShouldFailRunHonorsWarnPolicy(t *testing.T) {
	warnResults := []Result{{
		Name:          "best-effort",
		Enabled:       true,
		Passed:        false,
		FailurePolicy: "warn",
	}}
	if ShouldFailRun(warnResults) {
		t.Fatal("ShouldFailRun = true for warn policy, want false")
	}

	failResults := append(warnResults, Result{
		Name:          "required",
		Enabled:       true,
		Passed:        false,
		FailurePolicy: "fail",
	})
	if !ShouldFailRun(failResults) {
		t.Fatal("ShouldFailRun = false for failed fail policy, want true")
	}
}

func validatedWorkloadConfig(t *testing.T, workloadConfig config.WorkloadConfig) *config.Config {
	t.Helper()

	cfg := &config.Config{
		Service: config.ServiceConfig{
			Name:  "test-service",
			Image: "example/test-service:latest",
			Port:  8080,
		},
		Workloads: []config.WorkloadConfig{workloadConfig},
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate: %v", err)
	}

	return cfg
}
