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
