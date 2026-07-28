package kubernetes

import (
	"strings"
	"testing"

	"github.com/HXong/predeploy-guard/internal/config"
)

func TestGenerateManifestsForSimpleService(t *testing.T) {
	cfg := testManifestConfig()
	cfg.Service.Env = map[string]string{
		"MODE": "test",
	}

	manifest := mustGenerateManifest(t, &cfg)

	for _, want := range []string{
		"apiVersion: apps/v1",
		"kind: Deployment",
		"kind: Service",
		"name: booking-service",
		"namespace: predeploy-booking-a1b2c3d4",
		"app.kubernetes.io/managed-by: predeploy-guard",
		"app.kubernetes.io/part-of: predeploy-guard",
		"predeploy.guard/run: predeploy-booking-a1b2c3d4",
		"image: example/booking-service:latest",
		"containerPort: 8080",
		"path: /health",
		"name: MODE",
		"value: test",
	} {
		if !strings.Contains(manifest, want) {
			t.Fatalf("manifest does not contain %q:\n%s", want, manifest)
		}
	}
}

func TestGenerateManifestsForDependencyCommandReadiness(t *testing.T) {
	cfg := testManifestConfig()
	cfg.Dependencies = map[string]config.DependencyConfig{
		"Postgres_DB": {
			Image: "postgres:16",
			Port:  5432,
			Readiness: config.ReadinessConfig{
				Command:         []string{"pg_isready", "-U", "test"},
				IntervalSeconds: 2,
				TimeoutSeconds:  30,
			},
		},
	}

	manifest := mustGenerateManifest(t, &cfg)

	for _, want := range []string{
		"name: postgres-db",
		"image: postgres:16",
		"containerPort: 5432",
		"command:",
		"- pg_isready",
		"- -U",
		"- test",
		"periodSeconds: 2",
		"timeoutSeconds: 30",
	} {
		if !strings.Contains(manifest, want) {
			t.Fatalf("manifest does not contain %q:\n%s", want, manifest)
		}
	}
}

func TestGenerateManifestsForDependencyShellReadiness(t *testing.T) {
	cfg := testManifestConfig()
	cfg.Dependencies = map[string]config.DependencyConfig{
		"redis": {
			Image: "redis:7",
			Port:  6379,
			Readiness: config.ReadinessConfig{
				Shell:           "redis-cli ping | grep PONG",
				IntervalSeconds: 3,
				TimeoutSeconds:  20,
			},
		},
	}

	manifest := mustGenerateManifest(t, &cfg)

	for _, want := range []string{
		"name: redis",
		"- sh",
		"- -c",
		"- redis-cli ping | grep PONG",
		"periodSeconds: 3",
		"timeoutSeconds: 20",
	} {
		if !strings.Contains(manifest, want) {
			t.Fatalf("manifest does not contain %q:\n%s", want, manifest)
		}
	}
}

func TestGenerateManifestsIncludesIngressWhenEnabled(t *testing.T) {
	cfg := testManifestConfig()
	cfg.Runtime.Type = "kubernetes"
	cfg.Gateway = config.GatewayConfig{
		Enabled: true,
		BaseURL: "http://predeploy.local",
		Ingress: config.GatewayIngressConfig{
			Enabled:   true,
			Host:      "predeploy.local",
			ClassName: "local-controller",
			PathType:  "Exact",
			Annotations: map[string]string{
				"example.test/setting": "enabled",
			},
		},
		Routes: []config.GatewayRoute{
			{Name: "homepage", Path: "/"},
			{Name: "api", Path: "/api"},
		},
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate: %v", err)
	}

	manifest := mustGenerateManifest(t, &cfg)

	for _, want := range []string{
		"apiVersion: networking.k8s.io/v1",
		"kind: Ingress",
		"name: booking-service",
		"namespace: predeploy-booking-a1b2c3d4",
		"app.kubernetes.io/managed-by: predeploy-guard",
		"predeploy.guard/run: predeploy-booking-a1b2c3d4",
		"example.test/setting: enabled",
		"ingressClassName: local-controller",
		"host: predeploy.local",
		"path: /",
		"path: /api",
		"pathType: Exact",
		"number: 8080",
	} {
		if !strings.Contains(manifest, want) {
			t.Fatalf("manifest does not contain %q:\n%s", want, manifest)
		}
	}
}

func TestGenerateManifestsOmitsIngressWhenDisabled(t *testing.T) {
	cfg := testManifestConfig()
	cfg.Gateway = config.GatewayConfig{
		Enabled: true,
		BaseURL: "http://predeploy.local",
		Routes: []config.GatewayRoute{{
			Name: "homepage",
			Path: "/",
		}},
	}

	manifest := mustGenerateManifest(t, &cfg)

	if strings.Contains(manifest, "kind: Ingress") {
		t.Fatalf("manifest unexpectedly contains Ingress:\n%s", manifest)
	}
}

func TestGenerateIngressOmitsOptionalHostAndClassAndDeduplicatesPaths(t *testing.T) {
	cfg := testManifestConfig()
	cfg.Gateway = config.GatewayConfig{
		Enabled: true,
		BaseURL: "http://127.0.0.1",
		Ingress: config.GatewayIngressConfig{
			Enabled: true,
		},
		Routes: []config.GatewayRoute{
			{Name: "homepage-get", Path: "/"},
			{Name: "homepage-post", Method: "POST", Path: "/"},
		},
	}

	manifest := mustGenerateManifest(t, &cfg)
	ingressStart := strings.Index(manifest, "apiVersion: networking.k8s.io/v1")
	if ingressStart < 0 {
		t.Fatalf("manifest does not contain Ingress:\n%s", manifest)
	}
	ingressDocument := manifest[ingressStart:]

	if strings.Contains(ingressDocument, "ingressClassName:") {
		t.Fatalf("ingress unexpectedly contains ingressClassName:\n%s", ingressDocument)
	}
	if strings.Contains(ingressDocument, "host:") {
		t.Fatalf("ingress unexpectedly contains host:\n%s", ingressDocument)
	}
	if strings.Count(ingressDocument, "path: /") != 1 {
		t.Fatalf("ingress path count = %d, want 1:\n%s", strings.Count(ingressDocument, "path: /"), ingressDocument)
	}
	if !strings.Contains(ingressDocument, "pathType: Prefix") {
		t.Fatalf("ingress does not contain default Prefix path type:\n%s", ingressDocument)
	}
}

func mustGenerateManifest(t *testing.T, cfg *config.Config) string {
	t.Helper()

	data, err := generateManifests(cfg, "predeploy-booking-a1b2c3d4")
	if err != nil {
		t.Fatalf("generateManifests: %v", err)
	}

	return string(data)
}

func testManifestConfig() config.Config {
	return config.Config{
		Service: config.ServiceConfig{
			Name:       "Booking_Service",
			Image:      "example/booking-service:latest",
			Port:       8080,
			HealthPath: "/health",
		},
	}
}
