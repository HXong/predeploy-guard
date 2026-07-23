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

func validTestConfig() Config {
	return Config{
		Service: ServiceConfig{
			Name:  "test-service",
			Image: "example/test-service:latest",
			Port:  8080,
		},
	}
}
