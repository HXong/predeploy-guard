package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateDefaultsRuntimeToDockerCompose(t *testing.T) {
	cfg := validTestConfig()

	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate: %v", err)
	}

	if cfg.Runtime.Type != "docker-compose" {
		t.Fatalf("runtime type = %q, want docker-compose", cfg.Runtime.Type)
	}
}

func TestValidateAcceptsKubernetesRuntime(t *testing.T) {
	cfg := validTestConfig()
	cfg.Runtime = RuntimeConfig{
		Type:    "kubernetes",
		Context: "kind-local",
	}

	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate: %v", err)
	}

	if cfg.Runtime.Context != "kind-local" {
		t.Fatalf("runtime context = %q, want kind-local", cfg.Runtime.Context)
	}
}

func TestLoadParsesKubernetesContext(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "predeploy.yaml")
	content := []byte(`
runtime:
  type: kubernetes
  context: kind-local
service:
  name: test-service
  image: example/test-service:latest
  port: 8080
`)
	if err := os.WriteFile(configPath, content, 0600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if cfg.Runtime.Type != "kubernetes" {
		t.Fatalf("runtime type = %q, want kubernetes", cfg.Runtime.Type)
	}
	if cfg.Runtime.Context != "kind-local" {
		t.Fatalf("runtime context = %q, want kind-local", cfg.Runtime.Context)
	}
}

func TestValidateRejectsUnsupportedRuntime(t *testing.T) {
	cfg := validTestConfig()
	cfg.Runtime.Type = "production"

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate error = nil, want unsupported runtime error")
	}

	for _, want := range []string{
		`unsupported runtime.type "production"`,
		`supported runtimes: docker-compose, kubernetes`,
	} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("Validate error = %q, want it to contain %q", err, want)
		}
	}
}

func TestValidateAcceptsHTTPWorkload(t *testing.T) {
	enabled := false
	cfg := validTestConfig()
	cfg.Workloads = []WorkloadConfig{{
		Name:           "warmup-traffic",
		Type:           "http",
		Enabled:        &enabled,
		Method:         "POST",
		Path:           "/warmup",
		Duration:       "2s",
		RatePerSecond:  5,
		ExpectedStatus: 202,
		FailurePolicy:  "warn",
	}}

	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate: %v", err)
	}

	workload := cfg.Workloads[0]
	if workload.Enabled == nil || *workload.Enabled {
		t.Fatalf("enabled = %v, want false", workload.Enabled)
	}
	if workload.Method != "POST" {
		t.Fatalf("method = %q, want POST", workload.Method)
	}
}

func TestLoadParsesHTTPWorkload(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "predeploy.yaml")
	content := []byte(`
service:
  name: test-service
  image: example/test-service:latest
  port: 8080
workloads:
  - name: warmup-traffic
    type: http
    path: /ready
    duration: 1s
    ratePerSecond: 3
    expectedStatus: 204
    failurePolicy: warn
`)
	if err := os.WriteFile(configPath, content, 0600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(cfg.Workloads) != 1 {
		t.Fatalf("workload count = %d, want 1", len(cfg.Workloads))
	}

	workload := cfg.Workloads[0]
	if workload.Name != "warmup-traffic" ||
		workload.Path != "/ready" ||
		workload.RatePerSecond != 3 ||
		workload.ExpectedStatus != 204 ||
		workload.FailurePolicy != "warn" {
		t.Fatalf("workload = %#v, want parsed HTTP workload", workload)
	}
}

func TestValidateDefaultsHTTPWorkload(t *testing.T) {
	cfg := validTestConfig()
	cfg.Workloads = []WorkloadConfig{{
		Name: "warmup-traffic",
		Type: "http",
		Path: "/",
	}}

	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate: %v", err)
	}

	workload := cfg.Workloads[0]
	if workload.Enabled == nil || !*workload.Enabled {
		t.Fatalf("enabled = %v, want true", workload.Enabled)
	}
	if workload.Method != "GET" {
		t.Fatalf("method = %q, want GET", workload.Method)
	}
	if workload.Duration != "10s" {
		t.Fatalf("duration = %q, want 10s", workload.Duration)
	}
	if workload.RatePerSecond != 1 {
		t.Fatalf("rate per second = %d, want 1", workload.RatePerSecond)
	}
	if workload.ExpectedStatus != 200 {
		t.Fatalf("expected status = %d, want 200", workload.ExpectedStatus)
	}
	if workload.FailurePolicy != "fail" {
		t.Fatalf("failure policy = %q, want fail", workload.FailurePolicy)
	}
}

func TestValidateRejectsUnsupportedWorkloadType(t *testing.T) {
	cfg := validTestConfig()
	cfg.Workloads = []WorkloadConfig{{
		Name: "broker-traffic",
		Type: "kafka",
		Path: "/",
	}}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate error = nil, want unsupported workload type error")
	}

	want := `unsupported workloads[0].type "kafka"; supported workload types: http`
	if !strings.Contains(err.Error(), want) {
		t.Fatalf("Validate error = %q, want it to contain %q", err, want)
	}
}

func TestValidateRejectsDuplicateWorkloadNames(t *testing.T) {
	cfg := validTestConfig()
	cfg.Workloads = []WorkloadConfig{
		{Name: "warmup-traffic", Type: "http", Path: "/"},
		{Name: "warmup-traffic", Type: "http", Path: "/health"},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate error = nil, want duplicate workload name error")
	}
	if !strings.Contains(err.Error(), `workloads[1].name "warmup-traffic" must be unique`) {
		t.Fatalf("Validate error = %q, want duplicate workload name error", err)
	}
}

func TestValidateIgnoresGatewayFieldsWhenDisabled(t *testing.T) {
	cfg := validTestConfig()
	cfg.Gateway = GatewayConfig{
		BaseURL: "not-a-url",
		Routes: []GatewayRoute{{
			Path: "missing-leading-slash",
		}},
	}

	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate: %v", err)
	}
}

func TestValidateAcceptsEnabledGatewayAndDefaultsRoutes(t *testing.T) {
	cfg := validTestConfig()
	cfg.Gateway = GatewayConfig{
		Enabled: true,
		BaseURL: "http://localhost:8088",
		Routes: []GatewayRoute{{
			Name: "homepage-via-gateway",
			Path: "/",
		}},
	}

	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate: %v", err)
	}

	route := cfg.Gateway.Routes[0]
	if route.Method != "GET" {
		t.Fatalf("method = %q, want GET", route.Method)
	}
	if route.ExpectedStatus != 200 {
		t.Fatalf("expected status = %d, want 200", route.ExpectedStatus)
	}
	if route.CompareDirect == nil || !*route.CompareDirect {
		t.Fatalf("compareDirect = %v, want true", route.CompareDirect)
	}
}

func TestValidateRejectsEnabledGatewayWithoutBaseURL(t *testing.T) {
	cfg := validTestConfig()
	cfg.Gateway = GatewayConfig{
		Enabled: true,
		Routes: []GatewayRoute{{
			Name: "homepage-via-gateway",
			Path: "/",
		}},
	}

	err := cfg.Validate()
	if err == nil || !strings.Contains(err.Error(), "gateway.baseURL is required") {
		t.Fatalf("Validate error = %v, want missing gateway.baseURL error", err)
	}
}

func TestValidateRejectsGatewayRoutePathWithoutLeadingSlash(t *testing.T) {
	cfg := validTestConfig()
	cfg.Gateway = GatewayConfig{
		Enabled: true,
		BaseURL: "https://gateway.example.test",
		Routes: []GatewayRoute{{
			Name: "homepage-via-gateway",
			Path: "homepage",
		}},
	}

	err := cfg.Validate()
	if err == nil || !strings.Contains(err.Error(), "gateway.routes[0].path must start with /") {
		t.Fatalf("Validate error = %v, want invalid gateway route path error", err)
	}
}

func TestValidateAcceptsEnabledIngressForKubernetesRuntime(t *testing.T) {
	cfg := validGatewayIngressTestConfig()
	cfg.Gateway.Ingress = GatewayIngressConfig{
		Enabled:   true,
		Host:      "predeploy.local",
		ClassName: "local-controller",
		PathType:  "Exact",
		Annotations: map[string]string{
			"example.test/setting": "enabled",
		},
	}

	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate: %v", err)
	}
}

func TestValidateRejectsEnabledIngressForDockerComposeRuntime(t *testing.T) {
	cfg := validGatewayIngressTestConfig()
	cfg.Runtime.Type = "docker-compose"
	cfg.Gateway.Ingress.Enabled = true

	err := cfg.Validate()
	if err == nil || !strings.Contains(err.Error(), `gateway.ingress.enabled requires runtime.type "kubernetes"`) {
		t.Fatalf("Validate error = %v, want Kubernetes runtime requirement", err)
	}
}

func TestValidateDefaultsIngressPathTypeToPrefix(t *testing.T) {
	cfg := validGatewayIngressTestConfig()
	cfg.Gateway.Ingress.Enabled = true

	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if cfg.Gateway.Ingress.PathType != "Prefix" {
		t.Fatalf("ingress path type = %q, want Prefix", cfg.Gateway.Ingress.PathType)
	}
}

func TestValidateRejectsUnsupportedIngressPathType(t *testing.T) {
	cfg := validGatewayIngressTestConfig()
	cfg.Gateway.Ingress.Enabled = true
	cfg.Gateway.Ingress.PathType = "Regex"

	err := cfg.Validate()
	if err == nil || !strings.Contains(err.Error(), `unsupported gateway.ingress.pathType "Regex"`) {
		t.Fatalf("Validate error = %v, want unsupported ingress path type", err)
	}
}

func TestValidateRejectsEmptyIngressAnnotationKey(t *testing.T) {
	cfg := validGatewayIngressTestConfig()
	cfg.Gateway.Ingress.Enabled = true
	cfg.Gateway.Ingress.Annotations = map[string]string{"": "value"}

	err := cfg.Validate()
	if err == nil || !strings.Contains(err.Error(), "gateway.ingress.annotations cannot contain an empty key") {
		t.Fatalf("Validate error = %v, want empty annotation key error", err)
	}
}

func TestValidateRejectsIngressWhenGatewayDisabled(t *testing.T) {
	cfg := validTestConfig()
	cfg.Runtime.Type = "kubernetes"
	cfg.Gateway.Ingress.Enabled = true

	err := cfg.Validate()
	if err == nil || !strings.Contains(err.Error(), "gateway.ingress.enabled requires gateway.enabled") {
		t.Fatalf("Validate error = %v, want gateway enabled requirement", err)
	}
}

func TestValidateAcceptsGatewayLatencyThreshold(t *testing.T) {
	cfg := validGatewayIngressTestConfig()
	cfg.Gateway.Latency = GatewayLatencyConfig{
		Enabled:             true,
		MaxGatewayLatencyMs: 500,
	}

	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if cfg.Gateway.Latency.FailurePolicy != "warn" {
		t.Fatalf("latency failure policy = %q, want warn", cfg.Gateway.Latency.FailurePolicy)
	}
}

func TestValidateRejectsGatewayLatencyWithoutThresholds(t *testing.T) {
	cfg := validGatewayIngressTestConfig()
	cfg.Gateway.Latency.Enabled = true

	err := cfg.Validate()
	if err == nil || !strings.Contains(err.Error(), "gateway.latency.enabled requires at least one latency threshold") {
		t.Fatalf("Validate error = %v, want missing latency threshold error", err)
	}
}

func TestValidateRejectsUnsupportedGatewayLatencyFailurePolicy(t *testing.T) {
	cfg := validGatewayIngressTestConfig()
	cfg.Gateway.Latency = GatewayLatencyConfig{
		Enabled:             true,
		MaxGatewayLatencyMs: 500,
		FailurePolicy:       "ignore",
	}

	err := cfg.Validate()
	if err == nil || !strings.Contains(err.Error(), `unsupported gateway.latency.failurePolicy "ignore"`) {
		t.Fatalf("Validate error = %v, want unsupported latency failure policy", err)
	}
}

func TestValidateRejectsNonPositiveConfiguredGatewayLatencyThresholds(t *testing.T) {
	tests := []struct {
		name    string
		latency GatewayLatencyConfig
		field   string
	}{
		{
			name: "gateway latency",
			latency: GatewayLatencyConfig{
				Enabled:             true,
				MaxGatewayLatencyMs: -1,
			},
			field: "maxGatewayLatencyMs",
		},
		{
			name: "overhead milliseconds",
			latency: GatewayLatencyConfig{
				Enabled:       true,
				MaxOverheadMs: -1,
			},
			field: "maxOverheadMs",
		},
		{
			name: "overhead ratio",
			latency: GatewayLatencyConfig{
				Enabled:          true,
				MaxOverheadRatio: -1,
			},
			field: "maxOverheadRatio",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cfg := validGatewayIngressTestConfig()
			cfg.Gateway.Latency = test.latency

			err := cfg.Validate()
			if err == nil || !strings.Contains(err.Error(), "gateway.latency."+test.field) {
				t.Fatalf("Validate error = %v, want %s validation error", err, test.field)
			}
		})
	}
}

func TestValidateRejectsGatewayLatencyWhenGatewayDisabled(t *testing.T) {
	cfg := validTestConfig()
	cfg.Gateway.Latency = GatewayLatencyConfig{
		Enabled:             true,
		MaxGatewayLatencyMs: 500,
	}

	err := cfg.Validate()
	if err == nil || !strings.Contains(err.Error(), "gateway.latency.enabled requires gateway.enabled") {
		t.Fatalf("Validate error = %v, want gateway enabled requirement", err)
	}
}

func validGatewayIngressTestConfig() Config {
	cfg := validTestConfig()
	cfg.Runtime.Type = "kubernetes"
	cfg.Gateway = GatewayConfig{
		Enabled: true,
		BaseURL: "http://predeploy.local",
		Routes: []GatewayRoute{{
			Name: "homepage-via-gateway",
			Path: "/",
		}},
	}
	return cfg
}

func validTestConfig() Config {
	return Config{
		Service: ServiceConfig{
			Name:  "test-service",
			Image: "example/test-service:latest",
			Port:  8080,
		},
	}
}
