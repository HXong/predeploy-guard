package scaffold

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/HXong/predeploy-guard/internal/appdetect"
)

const DefaultConfigFilename = "predeploy.yaml"

type InitOptions struct {
	OutputPath   string
	Overwrite    bool
	Dependencies string

	AppPath     string
	Runtime     string
	ServiceName string
	Image       string
	Port        int
	HealthPath  string
	NoBuild     bool
	Guided      bool
}

type InitWarning struct {
	Message string
	Details string
}

type InitResult struct {
	OutputPath      string
	AppPath         string
	Detection       appdetect.DetectionResult
	Runtime         string
	ServiceName     string
	Image           string
	Port            int
	HealthPath      string
	BuildContext    string
	BuildConfigured bool
	DefaultedPort   bool
	Warnings        []InitWarning
}

func WriteDefaultConfig(options InitOptions) error {
	_, err := WriteConfig(options)
	return err
}

func WriteConfig(options InitOptions) (InitResult, error) {
	outputPath := options.OutputPath
	if outputPath == "" {
		outputPath = DefaultConfigFilename
	}

	absPath, err := filepath.Abs(outputPath)
	if err != nil {
		return InitResult{}, fmt.Errorf("resolve output path: %w", err)
	}

	var result InitResult
	var content string
	appAware := strings.TrimSpace(options.AppPath) != ""
	if !appAware {
		if err := checkOutputAvailable(absPath, options.Overwrite); err != nil {
			return InitResult{}, err
		}
		presets, presetErr := GetDependencyPresets(options.Dependencies)
		if presetErr != nil {
			return InitResult{}, presetErr
		}
		if options.Guided {
			result, content, err = buildGuidedGenericConfig(options, absPath, presets)
			if err != nil {
				return InitResult{}, err
			}
		} else {
			content = buildDefaultConfig(presets)
			result.OutputPath = absPath
		}
	} else {
		result, content, err = buildAppAwareConfig(options, absPath)
		if err != nil {
			return InitResult{}, err
		}
	}

	if appAware {
		if err := checkOutputAvailable(absPath, options.Overwrite); err != nil {
			return InitResult{}, err
		}
	}

	if err := writeConfigFile(absPath, content, options.Overwrite); err != nil {
		return InitResult{}, err
	}

	return result, nil
}

func buildGuidedGenericConfig(
	options InitOptions,
	outputPath string,
	presets []DependencyPreset,
) (InitResult, string, error) {
	runtimeType := strings.ToLower(strings.TrimSpace(options.Runtime))
	if runtimeType == "" {
		runtimeType = "docker-compose"
	}
	if runtimeType != "docker-compose" && runtimeType != "kubernetes" {
		return InitResult{}, "", fmt.Errorf(
			"unsupported runtime %q; supported runtimes: docker-compose, kubernetes",
			options.Runtime,
		)
	}

	serviceName := "my-service"
	if strings.TrimSpace(options.ServiceName) != "" {
		serviceName = appdetect.SanitizeServiceName(options.ServiceName)
	}
	image := strings.TrimSpace(options.Image)
	if image == "" {
		image = "my-service:predeploy"
	}
	port := options.Port
	if port == 0 {
		port = 8080
	}
	if port < 1 || port > 65535 {
		return InitResult{}, "", fmt.Errorf("port must be between 1 and 65535")
	}
	healthPath := strings.TrimSpace(options.HealthPath)
	if healthPath == "" {
		healthPath = "/health"
	}
	if !strings.HasPrefix(healthPath, "/") {
		return InitResult{}, "", fmt.Errorf("health path must start with /")
	}

	result := InitResult{
		OutputPath:      outputPath,
		Runtime:         runtimeType,
		ServiceName:     serviceName,
		Image:           image,
		Port:            port,
		HealthPath:      healthPath,
		BuildConfigured: !options.NoBuild,
	}
	if result.BuildConfigured {
		result.BuildContext = "."
	}

	return result, buildGuidedGenericConfigContent(result, presets), nil
}

func buildGuidedGenericConfigContent(result InitResult, presets []DependencyPreset) string {
	var builder bytes.Buffer

	fmt.Fprintf(&builder, "runtime:\n  type: %s\n\n", result.Runtime)
	builder.WriteString("service:\n")
	fmt.Fprintf(&builder, "  name: %s\n", strconv.Quote(result.ServiceName))
	fmt.Fprintf(&builder, "  image: %s\n", strconv.Quote(result.Image))
	if result.BuildConfigured {
		builder.WriteString("  build:\n")
		builder.WriteString("    context: .\n")
		builder.WriteString("    dockerfile: Dockerfile\n")
	}
	fmt.Fprintf(&builder, "  port: %d\n", result.Port)
	fmt.Fprintf(&builder, "  healthPath: %s\n", strconv.Quote(result.HealthPath))

	writeServiceEnvBlock(&builder, presets)
	writeDependenciesBlock(&builder, presets)

	builder.WriteString("\nchecks:\n  smoke:\n")
	builder.WriteString("    - name: health check\n")
	builder.WriteString("      method: GET\n")
	fmt.Fprintf(&builder, "      path: %s\n", strconv.Quote(result.HealthPath))
	builder.WriteString("      expectedStatus: 200\n")

	builder.WriteString(`
performance:
  enabled: true
  vus: 10
  duration: 15s
  thresholds:
    maxP95LatencyMs: 300
    maxErrorRate: 0.01
  endpoints:
`)
	writeGenericPerformanceEndpoint(&builder, result.HealthPath, 4)
	builder.WriteString(`
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
`)
	writeGenericPerformanceEndpoint(&builder, result.HealthPath, 8)
	builder.WriteString(`
  stress-test:
    performance:
      enabled: true
      vus: 50
      duration: 30s
      thresholds:
        maxP95LatencyMs: 800
        maxErrorRate: 0.05
      endpoints:
`)
	writeGenericPerformanceEndpoint(&builder, result.HealthPath, 8)
	builder.WriteString(`
settings:
  cleanup: true
  timeoutSeconds: 60
`)

	return builder.String()
}

func writeGenericPerformanceEndpoint(builder *bytes.Buffer, healthPath string, indentation int) {
	prefix := strings.Repeat(" ", indentation)
	fmt.Fprintf(builder, "%s- name: health load\n", prefix)
	fmt.Fprintf(builder, "%s  method: GET\n", prefix)
	fmt.Fprintf(builder, "%s  path: %s\n", prefix, strconv.Quote(healthPath))
}

func writeConfigFile(path string, content string, overwrite bool) error {
	flags := os.O_WRONLY | os.O_CREATE
	if overwrite {
		flags |= os.O_TRUNC
	} else {
		flags |= os.O_EXCL
	}

	file, err := os.OpenFile(path, flags, 0644)
	if err != nil {
		if !overwrite && errors.Is(err, os.ErrExist) {
			return fmt.Errorf("config file already exists: %s; use --force to overwrite", path)
		}
		return fmt.Errorf("write default config: %w", err)
	}
	if _, err := file.WriteString(content); err != nil {
		_ = file.Close()
		return fmt.Errorf("write default config: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close default config: %w", err)
	}
	return nil
}

func checkOutputAvailable(path string, overwrite bool) error {
	if overwrite {
		return nil
	}
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("config file already exists: %s; use --force to overwrite", path)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("check existing config file: %w", err)
	}
	return nil
}

func buildAppAwareConfig(options InitOptions, outputPath string) (InitResult, string, error) {
	runtimeType := strings.ToLower(strings.TrimSpace(options.Runtime))
	if runtimeType == "" {
		runtimeType = "docker-compose"
	}
	if runtimeType != "docker-compose" && runtimeType != "kubernetes" {
		return InitResult{}, "", fmt.Errorf(
			"unsupported runtime %q; supported runtimes: docker-compose, kubernetes",
			options.Runtime,
		)
	}
	if options.Port < 0 || options.Port > 65535 {
		return InitResult{}, "", fmt.Errorf("port must be between 1 and 65535")
	}

	healthPath := strings.TrimSpace(options.HealthPath)
	if healthPath == "" {
		healthPath = "/health"
	}
	if !strings.HasPrefix(healthPath, "/") {
		return InitResult{}, "", fmt.Errorf("health path must start with /")
	}

	detection, err := appdetect.Detect(appdetect.Options{AppPath: options.AppPath})
	if err != nil {
		return InitResult{}, "", err
	}
	if pathWithin(outputPath, detection.AppPath) {
		return InitResult{}, "", fmt.Errorf(
			"output config must be outside the app directory to keep the app unchanged: %s\n\nTry:\n  predeploy init --app %s --output %s",
			options.AppPath,
			commandLineArgument(options.AppPath),
			commandLineArgument(safeOutputPath(detection.AppPath)),
		)
	}

	presets, err := GetDependencyPresets(options.Dependencies)
	if err != nil {
		return InitResult{}, "", err
	}

	serviceName := detection.Name
	if strings.TrimSpace(options.ServiceName) != "" {
		serviceName = appdetect.SanitizeServiceName(options.ServiceName)
	}

	image := strings.TrimSpace(options.Image)
	if image == "" {
		image = fmt.Sprintf("predeploy-%s:local", serviceName)
	}

	port := options.Port
	defaultedPort := port == 0
	if defaultedPort {
		port = 8080
	}

	result := InitResult{
		OutputPath:    outputPath,
		AppPath:       options.AppPath,
		Detection:     detection,
		Runtime:       runtimeType,
		ServiceName:   serviceName,
		Image:         image,
		Port:          port,
		HealthPath:    healthPath,
		DefaultedPort: defaultedPort,
	}

	if detection.HasDockerfile && !options.NoBuild {
		result.BuildConfigured = true
		result.BuildContext = relativeBuildContext(filepath.Dir(outputPath), detection.AppPath)
	}
	if !detection.HasDockerfile {
		result.Warnings = append(result.Warnings, InitWarning{
			Message: fmt.Sprintf("Dockerfile not found. Generated config references image %s.", image),
			Details: "Replace service.image with an existing image or add a Dockerfile later.",
		})
	}
	if defaultedPort {
		result.Warnings = append(result.Warnings, InitWarning{
			Message: "No --port provided. Defaulted to 8080; update service.port if your app uses a different port.",
		})
	}

	return result, buildDetectedAppConfig(result, presets), nil
}

func safeOutputPath(appPath string) string {
	candidate := filepath.Join(filepath.Dir(appPath), DefaultConfigFilename)
	workingDir, err := os.Getwd()
	if err != nil {
		return filepath.ToSlash(candidate)
	}
	relative, err := filepath.Rel(workingDir, candidate)
	if err != nil {
		return filepath.ToSlash(candidate)
	}
	relative = filepath.ToSlash(relative)
	if relative != "." && relative != ".." && !strings.HasPrefix(relative, "../") {
		return "./" + relative
	}
	return relative
}

func commandLineArgument(value string) string {
	if strings.ContainsAny(value, " \t\"") {
		return `"` + strings.ReplaceAll(value, `"`, `\"`) + `"`
	}
	return value
}

func buildDetectedAppConfig(result InitResult, presets []DependencyPreset) string {
	var builder bytes.Buffer

	fmt.Fprintf(&builder, "runtime:\n  type: %s\n\n", result.Runtime)
	builder.WriteString("service:\n")
	fmt.Fprintf(&builder, "  name: %s\n", strconv.Quote(result.ServiceName))
	fmt.Fprintf(&builder, "  image: %s\n", strconv.Quote(result.Image))
	if result.BuildConfigured {
		builder.WriteString("  build:\n")
		fmt.Fprintf(&builder, "    context: %s\n", strconv.Quote(result.BuildContext))
		builder.WriteString("    dockerfile: Dockerfile\n")
	}
	fmt.Fprintf(&builder, "  port: %d\n", result.Port)
	fmt.Fprintf(&builder, "  healthPath: %s\n", strconv.Quote(result.HealthPath))

	writeServiceEnvBlock(&builder, presets)
	writeDependenciesBlock(&builder, presets)

	builder.WriteString("\nchecks:\n  smoke:\n")
	builder.WriteString("    - name: health endpoint\n")
	builder.WriteString("      method: GET\n")
	fmt.Fprintf(&builder, "      path: %s\n", strconv.Quote(result.HealthPath))
	builder.WriteString("      expectedStatus: 200\n")

	builder.WriteString(`
gateway:
  enabled: false

performance:
  enabled: false

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
`)
	writePerformanceEndpoint(&builder, result.HealthPath)
	builder.WriteString(`
  stress-test:
    performance:
      enabled: true
      vus: 50
      duration: 30s
      thresholds:
        maxP95LatencyMs: 800
        maxErrorRate: 0.05
      endpoints:
`)
	writePerformanceEndpoint(&builder, result.HealthPath)
	builder.WriteString(`
settings:
  cleanup: true
  timeoutSeconds: 60
`)

	return builder.String()
}

func writePerformanceEndpoint(builder *bytes.Buffer, healthPath string) {
	builder.WriteString("        - name: health load\n")
	builder.WriteString("          method: GET\n")
	fmt.Fprintf(builder, "          path: %s\n", strconv.Quote(healthPath))
}

func relativeBuildContext(outputDir string, appPath string) string {
	relative, err := filepath.Rel(outputDir, appPath)
	if err != nil {
		return filepath.ToSlash(appPath)
	}
	relative = filepath.ToSlash(relative)
	if relative != "." && relative != ".." && !strings.HasPrefix(relative, "../") {
		return "./" + relative
	}
	return relative
}

func pathWithin(path string, directory string) bool {
	relative, err := filepath.Rel(directory, path)
	if err != nil {
		return false
	}
	return relative == "." || (relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)))
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
