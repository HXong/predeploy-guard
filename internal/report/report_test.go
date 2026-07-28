package report

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/HXong/predeploy-guard/internal/config"
	"github.com/HXong/predeploy-guard/internal/gateway"
	"github.com/HXong/predeploy-guard/internal/workload"
)

func TestReportIncludesWorkloadResults(t *testing.T) {
	startedAt := time.Date(2026, time.July, 24, 10, 0, 0, 0, time.UTC)
	cfg := &config.Config{
		ConfigDir: t.TempDir(),
		Runtime: config.RuntimeConfig{
			Type: "kubernetes",
		},
		Service: config.ServiceConfig{
			Name: "test-service",
		},
		Gateway: config.GatewayConfig{
			Enabled: true,
			BaseURL: "http://predeploy.local",
			Ingress: config.GatewayIngressConfig{
				Enabled:   true,
				Host:      "predeploy.local",
				ClassName: "local-controller",
			},
			Routes: []config.GatewayRoute{{
				Name: "homepage-via-gateway",
				Path: "/",
			}},
		},
		Workloads: []config.WorkloadConfig{{
			Name: "warmup-traffic",
			Type: "http",
			Path: "/",
		}},
	}
	data := ReportData{
		ServiceName: "test-service",
		Runtime:     "docker-compose",
		BaseURL:     "http://127.0.0.1:8080",
		RuntimeEnvironment: RuntimeEnvironment{
			Runtime:         "docker-compose",
			Name:            "test-service",
			BaseURL:         "http://127.0.0.1:8080",
			WorkloadBaseURL: "http://host.docker.internal:8080",
		},
		RunPhases: []RunPhase{
			{
				Name:       PhaseImageBuild,
				Status:     PhaseStatusSkipped,
				StartedAt:  startedAt,
				FinishedAt: startedAt,
			},
			{
				Name:       PhaseExperimentWorkloads,
				Status:     PhaseStatusPassed,
				StartedAt:  startedAt,
				FinishedAt: startedAt.Add(10 * time.Second),
				Duration:   10 * time.Second,
			},
		},
		StartedAt:  startedAt,
		FinishedAt: startedAt.Add(time.Second),
		GatewayResults: []gateway.RouteResult{{
			Name:           "homepage-via-gateway",
			Method:         "GET",
			Path:           "/",
			ExpectedStatus: 200,
			GatewayStatus:  200,
			GatewayPassed:  true,
			CompareDirect:  true,
			DirectStatus:   200,
			DirectPassed:   true,
			StatusMatched:  true,
			Passed:         true,
		}},
		WorkloadResults: []workload.Result{{
			Name:          "warmup-traffic",
			Type:          "http",
			Enabled:       true,
			Passed:        true,
			FailurePolicy: "fail",
			Method:        "GET",
			Path:          "/",
			RequestCount:  5,
			SuccessCount:  5,
		}},
		Passed: true,
		Logs:   "runtime diagnostic output",
	}

	markdown := buildMarkdown(cfg, data)
	for _, want := range []string{
		"## Runtime Environment",
		"| Runtime | docker-compose |",
		"| Environment | test-service |",
		"| Workload Base URL | http://host.docker.internal:8080 |",
		"## Run Timeline",
		"| image-build | skipped | 0ms | - |",
		"| experiment-workloads | passed | 10s | - |",
		"## Experiment Workloads",
		"| warmup-traffic | http | fail | GET / | 5 | 5 | 0 | PASS |",
		"## Gateway Checks",
		"- Ingress generation: enabled",
		"- Ingress host: predeploy.local",
		"- Ingress class: local-controller",
		"| homepage-via-gateway | GET | / | 200 | 200 | yes | PASS |",
		"- Gateway checks: 1 total, 1 passed, 0 failed",
		"## Runtime Diagnostics",
		"runtime diagnostic output",
	} {
		if !strings.Contains(markdown, want) {
			t.Fatalf("markdown does not contain %q:\n%s", want, markdown)
		}
	}
	if strings.Contains(markdown, "## Container Logs") {
		t.Fatalf("markdown contains legacy Container Logs heading:\n%s", markdown)
	}

	sectionOrder := []string{
		"## Summary",
		"## Runtime Environment",
		"## Run Timeline",
		"## Build Check",
		"## Readiness Checks",
		"## Gateway Checks",
		"## Smoke Checks",
		"## Experiment Workloads",
		"## Performance Check",
		"## Check Summary",
		"## Configuration",
		"## Runtime Diagnostics",
	}
	previousIndex := -1
	for _, section := range sectionOrder {
		index := strings.Index(markdown, section)
		if index <= previousIndex {
			t.Fatalf("section %q is out of order in markdown:\n%s", section, markdown)
		}
		previousIndex = index
	}

	jsonPath, err := WriteJSON(cfg, data)
	if err != nil {
		t.Fatalf("WriteJSON: %v", err)
	}
	content, err := os.ReadFile(jsonPath)
	if err != nil {
		t.Fatalf("read JSON report: %v", err)
	}

	var decoded map[string]json.RawMessage
	if err := json.Unmarshal(content, &decoded); err != nil {
		t.Fatalf("unmarshal JSON report: %v", err)
	}
	for _, existingField := range []string{
		"ServiceName",
		"Runtime",
		"BaseURL",
		"GatewayResults",
		"WorkloadResults",
		"Logs",
	} {
		if _, exists := decoded[existingField]; !exists {
			t.Fatalf("JSON report is missing existing field %q", existingField)
		}
	}
	var workloads []workload.Result
	if err := json.Unmarshal(decoded["WorkloadResults"], &workloads); err != nil {
		t.Fatalf("unmarshal workloads: %v", err)
	}
	if len(workloads) != 1 || workloads[0].Name != "warmup-traffic" {
		t.Fatalf("workloads = %#v, want warmup-traffic result", workloads)
	}

	var gatewayResults []gateway.RouteResult
	if err := json.Unmarshal(decoded["GatewayResults"], &gatewayResults); err != nil {
		t.Fatalf("unmarshal gateway results: %v", err)
	}
	if len(gatewayResults) != 1 || gatewayResults[0].Name != "homepage-via-gateway" {
		t.Fatalf("gateway results = %#v, want homepage-via-gateway result", gatewayResults)
	}

	var environment RuntimeEnvironment
	if err := json.Unmarshal(decoded["runtimeEnvironment"], &environment); err != nil {
		t.Fatalf("unmarshal runtimeEnvironment: %v", err)
	}
	if environment.Runtime != "docker-compose" || environment.Name != "test-service" {
		t.Fatalf("runtimeEnvironment = %#v, want Docker Compose test-service", environment)
	}

	var phases []RunPhase
	if err := json.Unmarshal(decoded["runPhases"], &phases); err != nil {
		t.Fatalf("unmarshal runPhases: %v", err)
	}
	if len(phases) != 2 || phases[1].Name != PhaseExperimentWorkloads {
		t.Fatalf("runPhases = %#v, want workload phase", phases)
	}
}
