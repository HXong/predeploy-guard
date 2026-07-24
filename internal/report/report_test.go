package report

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/HXong/predeploy-guard/internal/config"
	"github.com/HXong/predeploy-guard/internal/workload"
)

func TestReportIncludesWorkloadResults(t *testing.T) {
	startedAt := time.Date(2026, time.July, 24, 10, 0, 0, 0, time.UTC)
	cfg := &config.Config{
		ConfigDir: t.TempDir(),
		Service: config.ServiceConfig{
			Name: "test-service",
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
		StartedAt:   startedAt,
		FinishedAt:  startedAt.Add(time.Second),
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
	}

	markdown := buildMarkdown(cfg, data)
	for _, want := range []string{
		"## Experiment Workloads",
		"| warmup-traffic | http | fail | GET / | 5 | 5 | 0 | PASS |",
	} {
		if !strings.Contains(markdown, want) {
			t.Fatalf("markdown does not contain %q:\n%s", want, markdown)
		}
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
	var workloads []workload.Result
	if err := json.Unmarshal(decoded["WorkloadResults"], &workloads); err != nil {
		t.Fatalf("unmarshal workloads: %v", err)
	}
	if len(workloads) != 1 || workloads[0].Name != "warmup-traffic" {
		t.Fatalf("workloads = %#v, want warmup-traffic result", workloads)
	}
}
