package app

import (
	"strings"
	"testing"

	"github.com/HXong/predeploy-guard/internal/config"
)

func TestConfigSummaryAndExplainIncludeWorkloads(t *testing.T) {
	cfg := &config.Config{
		ConfigDir: t.TempDir(),
		Service: config.ServiceConfig{
			Name:  "test-service",
			Image: "example/test-service:latest",
			Port:  8080,
		},
		Workloads: []config.WorkloadConfig{{
			Name:          "warmup-traffic",
			Type:          "http",
			Path:          "/",
			Duration:      "1s",
			RatePerSecond: 2,
			FailurePolicy: "warn",
		}},
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate: %v", err)
	}

	service := NewConfigService(cfg)
	summary := service.Summary()
	if summary.Counts.Workloads != 1 {
		t.Fatalf("workload count = %d, want 1", summary.Counts.Workloads)
	}
	if len(summary.Workloads) != 1 {
		t.Fatalf("workload summaries = %d, want 1", len(summary.Workloads))
	}
	if summary.Workloads[0].Name != "warmup-traffic" || !summary.Workloads[0].Enabled {
		t.Fatalf("workload summary = %#v, want enabled warmup-traffic", summary.Workloads[0])
	}

	explanation := service.Explain()
	if len(explanation.Steps) != 10 {
		t.Fatalf("explanation step count = %d, want 10", len(explanation.Steps))
	}
	step := explanation.Steps[6]
	if step.Number != 6 || step.Title != "Experiment Workloads" {
		t.Fatalf("workload explanation step = %#v", step)
	}
	details := strings.Join(step.Details, " ")
	for _, want := range []string{"warmup-traffic", "GET /", "2 request(s)/second", "failure policy warn"} {
		if !strings.Contains(details, want) {
			t.Fatalf("workload explanation %q does not contain %q", details, want)
		}
	}
}
