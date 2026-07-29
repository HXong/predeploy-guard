package app

import (
	"strings"
	"testing"

	"github.com/HXong/predeploy-guard/internal/config"
)

func TestConfigSummaryAndExplainIncludeWorkloads(t *testing.T) {
	cfg := &config.Config{
		ConfigDir: t.TempDir(),
		Runtime: config.RuntimeConfig{
			Type: "kubernetes",
		},
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
		Gateway: config.GatewayConfig{
			Enabled: true,
			BaseURL: "http://localhost:8088",
			Ingress: config.GatewayIngressConfig{
				Enabled:   true,
				Host:      "predeploy.local",
				ClassName: "local-controller",
				Annotations: map[string]string{
					"example.test/setting": "private-value",
				},
			},
			Latency: config.GatewayLatencyConfig{
				Enabled:             true,
				MaxGatewayLatencyMs: 500,
				MaxOverheadMs:       200,
				MaxOverheadRatio:    3,
				FailurePolicy:       "warn",
			},
			Routes: []config.GatewayRoute{{
				Name: "homepage-via-gateway",
				Path: "/",
			}},
		},
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
	if !summary.Gateway.Enabled || summary.Gateway.BaseURL != "http://localhost:8088" {
		t.Fatalf("gateway summary = %#v, want enabled localhost gateway", summary.Gateway)
	}
	if summary.Counts.GatewayRoutes != 1 || len(summary.Gateway.Routes) != 1 {
		t.Fatalf("gateway route summary = %#v, want one route", summary.Gateway.Routes)
	}
	if !summary.Gateway.Routes[0].CompareDirect {
		t.Fatalf("gateway compareDirect = false, want default true")
	}
	if !summary.Gateway.Ingress.Enabled ||
		summary.Gateway.Ingress.Host != "predeploy.local" ||
		summary.Gateway.Ingress.ClassName != "local-controller" ||
		summary.Gateway.Ingress.PathType != "Prefix" ||
		summary.Gateway.Ingress.AnnotationCount != 1 {
		t.Fatalf("gateway ingress summary = %#v, want safe enabled ingress details", summary.Gateway.Ingress)
	}
	if !summary.Gateway.Latency.Enabled ||
		summary.Gateway.Latency.MaxGatewayLatencyMs != 500 ||
		summary.Gateway.Latency.MaxOverheadMs != 200 ||
		summary.Gateway.Latency.MaxOverheadRatio != 3 ||
		summary.Gateway.Latency.FailurePolicy != "warn" {
		t.Fatalf("gateway latency summary = %#v, want configured thresholds", summary.Gateway.Latency)
	}

	explanation := service.Explain()
	if len(explanation.Steps) != 11 {
		t.Fatalf("explanation step count = %d, want 11", len(explanation.Steps))
	}
	gatewayStep := explanation.Steps[5]
	gatewayDetails := strings.Join(gatewayStep.Details, " ")
	for _, want := range []string{
		"Gateway correctness checks run after service readiness",
		"direct service route",
		"homepage-via-gateway",
		"generate an owned Kubernetes Ingress",
		"does not install or manage an ingress controller",
		"must already resolve to the local ingress endpoint",
		"retry every second",
		"capped at 30 seconds",
		"measured from the final gateway and direct check result",
		"not a substitute for k6 load testing",
		"warn reports threshold violations without failing the run",
	} {
		if !strings.Contains(gatewayDetails, want) {
			t.Fatalf("gateway explanation %q does not contain %q", gatewayDetails, want)
		}
	}

	step := explanation.Steps[7]
	if step.Number != 7 || step.Title != "Experiment Workloads" {
		t.Fatalf("workload explanation step = %#v", step)
	}
	details := strings.Join(step.Details, " ")
	for _, want := range []string{"warmup-traffic", "GET /", "2 request(s)/second", "failure policy warn"} {
		if !strings.Contains(details, want) {
			t.Fatalf("workload explanation %q does not contain %q", details, want)
		}
	}
}
