package scaffold

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

const DefaultConfigFilename = "predeploy.yaml"

type InitOptions struct {
	OutputPath   string
	Overwrite    bool
	Dependencies string
}

func WriteDefaultConfig(options InitOptions) error {
	outputPath := options.OutputPath
	if outputPath == "" {
		outputPath = DefaultConfigFilename
	}

	absPath, err := filepath.Abs(outputPath)
	if err != nil {
		return fmt.Errorf("resolve output path: %w", err)
	}

	if !options.Overwrite {
		if _, err := os.Stat(absPath); err == nil {
			return fmt.Errorf("config file already exists: %s; use --force to overwrite", absPath)
		} else if !os.IsNotExist(err) {
			return fmt.Errorf("check existing config file: %w", err)
		}
	}

	presets, err := GetDependencyPresets(options.Dependencies)
	if err != nil {
		return err
	}

	content := buildDefaultConfig(presets)

	if err := os.WriteFile(absPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("write default config: %w", err)
	}

	return nil
}

func buildDefaultConfig(presets []DependencyPreset) string {
	var builder bytes.Buffer

	builder.WriteString(`runtime:
  type: docker-compose

service:
  name: my-service
  image: my-service:predeploy
  build:
    context: .
    dockerfile: Dockerfile
  port: 8080
  healthPath: /health
`)

	writeServiceEnvBlock(&builder, presets)
	writeDependenciesBlock(&builder, presets)

	builder.WriteString(`
checks:
  smoke:
    - name: health check
      method: GET
      path: /health
      expectedStatus: 200

performance:
  enabled: true
  vus: 10
  duration: 15s
  thresholds:
    maxP95LatencyMs: 300
    maxErrorRate: 0.01
  endpoints:
    - name: health load
      method: GET
      path: /health

profiles:
  smoke-only:
    performance:
      enabled: false

  light-load:
    performance:
      enabled: true
      vus: 10
      duration: 15s
      thresholds:
        maxP95LatencyMs: 300
        maxErrorRate: 0.01
      endpoints:
        - name: health load
          method: GET
          path: /health

  stress-test:
    performance:
      enabled: true
      vus: 50
      duration: 30s
      thresholds:
        maxP95LatencyMs: 800
        maxErrorRate: 0.05
      endpoints:
        - name: health load
          method: GET
          path: /health

settings:
  cleanup: true
  timeoutSeconds: 60
`)

	return builder.String()
}

func writeServiceEnvBlock(builder *bytes.Buffer, presets []DependencyPreset) {
	env := make(map[string]string)

	for _, preset := range presets {
		for key, value := range preset.Env {
			env[key] = value
		}
	}

	if len(env) == 0 {
		return
	}

	builder.WriteString("  env:\n")

	keys := make([]string, 0, len(env))
	for key := range env {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		fmt.Fprintf(builder, "    %s: %q\n", key, env[key])
	}
}

func writeDependenciesBlock(builder *bytes.Buffer, presets []DependencyPreset) {
	if len(presets) == 0 {
		return
	}

	builder.WriteString("\ndependencies:\n")

	for _, preset := range presets {
		builder.WriteString(preset.YAML)
	}
}
