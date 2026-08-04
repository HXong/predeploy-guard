package doctor

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type fakeCommandResult struct {
	output string
	err    error
}

type fakeCommandRunner struct {
	paths    map[string]string
	commands map[string]fakeCommandResult
}

func (f fakeCommandRunner) LookPath(file string) (string, error) {
	if path, ok := f.paths[file]; ok {
		return path, nil
	}
	return "", errors.New("executable not found")
}

func (f fakeCommandRunner) Run(_ context.Context, name string, args ...string) (string, error) {
	key := strings.Join(append([]string{name}, args...), " ")
	if result, ok := f.commands[key]; ok {
		return result.output, result.err
	}
	return "", errors.New("unexpected command: " + key)
}

func TestRunReportsPassWhenDockerCommandsSucceed(t *testing.T) {
	report := RunWithRunner(context.Background(), Options{WorkingDir: t.TempDir()}, fakeCommandRunner{
		paths: map[string]string{"docker": "/usr/bin/docker"},
		commands: map[string]fakeCommandResult{
			"docker info":            {},
			"docker compose version": {output: "Docker Compose version v2.30.0"},
		},
	})

	assertResult(t, report, "docker-cli", StatusPass)
	assertResult(t, report, "docker-daemon", StatusPass)
	assertResult(t, report, "docker-compose", StatusPass)
}

func TestRunReportsWarnWhenDockerIsMissing(t *testing.T) {
	report := RunWithRunner(context.Background(), Options{WorkingDir: t.TempDir()}, fakeCommandRunner{})

	assertResult(t, report, "docker-cli", StatusWarn)
	assertResult(t, report, "docker-daemon", StatusWarn)
	assertResult(t, report, "docker-compose", StatusWarn)
}

func TestRunReportsWarnWhenDockerDaemonIsUnreachable(t *testing.T) {
	report := RunWithRunner(context.Background(), Options{WorkingDir: t.TempDir()}, fakeCommandRunner{
		paths: map[string]string{"docker": "/usr/bin/docker"},
		commands: map[string]fakeCommandResult{
			"docker info":            {err: errors.New("daemon unavailable")},
			"docker compose version": {output: "Docker Compose version v2.30.0"},
		},
	})

	assertResult(t, report, "docker-daemon", StatusWarn)
}

func TestRunReportsPassForKubectlContext(t *testing.T) {
	report := RunWithRunner(context.Background(), Options{WorkingDir: t.TempDir()}, fakeCommandRunner{
		paths: map[string]string{"kubectl": "/usr/bin/kubectl"},
		commands: map[string]fakeCommandResult{
			"kubectl config current-context": {output: "kind-local\n"},
			"kubectl cluster-info":           {output: "Kubernetes control plane is running"},
		},
	})

	result := assertResult(t, report, "kubernetes-context", StatusPass)
	if !strings.Contains(result.Message, "kind-local") {
		t.Fatalf("context message = %q, want context name", result.Message)
	}
	assertResult(t, report, "kubernetes-cluster", StatusPass)
}

func TestRunReportsWarnWhenKubectlIsMissing(t *testing.T) {
	report := RunWithRunner(context.Background(), Options{WorkingDir: t.TempDir()}, fakeCommandRunner{})

	assertResult(t, report, "kubectl", StatusWarn)
	assertResult(t, report, "kubernetes-context", StatusWarn)
	assertResult(t, report, "kubernetes-cluster", StatusWarn)
}

func TestRunReportsPassForAppPathWithDockerfile(t *testing.T) {
	workingDir := t.TempDir()
	appDir := filepath.Join(workingDir, "my-app")
	if err := os.Mkdir(appDir, 0755); err != nil {
		t.Fatalf("Mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(appDir, "Dockerfile"), []byte("FROM scratch\n"), 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	if err := os.WriteFile(filepath.Join(appDir, "package.json"), []byte("{}\n"), 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	report := RunWithRunner(context.Background(), Options{
		WorkingDir: workingDir,
		AppPath:    "./my-app",
	}, fakeCommandRunner{})

	assertResult(t, report, "app-path", StatusPass)
	indicatorResult := assertResult(t, report, "project-indicators", StatusPass)
	if indicatorResult.Message != "Project indicators found: Dockerfile, package.json" {
		t.Fatalf("indicator message = %q", indicatorResult.Message)
	}
	assertResult(t, report, "dockerfile", StatusPass)
}

func TestRunReportsWarnForAppPathWithoutProjectIndicators(t *testing.T) {
	workingDir := t.TempDir()
	appDir := filepath.Join(workingDir, "my-app")
	if err := os.Mkdir(appDir, 0755); err != nil {
		t.Fatalf("Mkdir: %v", err)
	}

	report := RunWithRunner(context.Background(), Options{
		WorkingDir: workingDir,
		AppPath:    appDir,
	}, fakeCommandRunner{})

	assertResult(t, report, "app-path", StatusPass)
	indicatorResult := assertResult(t, report, "project-indicators", StatusWarn)
	if indicatorResult.Message != "No common project indicators found" {
		t.Fatalf("indicator message = %q", indicatorResult.Message)
	}
	assertResult(t, report, "dockerfile", StatusWarn)
}

func TestRunReportsFailForMissingAppPath(t *testing.T) {
	report := RunWithRunner(context.Background(), Options{
		WorkingDir: t.TempDir(),
		AppPath:    "./missing-app",
	}, fakeCommandRunner{})

	assertResult(t, report, "app-path", StatusFail)
}

func TestRunReportsFailForInvalidConfigPath(t *testing.T) {
	report := RunWithRunner(context.Background(), Options{
		WorkingDir: t.TempDir(),
		ConfigPath: "missing.yaml",
	}, fakeCommandRunner{})

	assertResult(t, report, "config", StatusFail)
}

func TestRunRecommendsAppAwareInteractiveInitWhenConfigIsMissing(t *testing.T) {
	workingDir := t.TempDir()
	appDir := filepath.Join(workingDir, "my-app")
	if err := os.Mkdir(appDir, 0755); err != nil {
		t.Fatalf("Mkdir: %v", err)
	}

	report := RunWithRunner(context.Background(), Options{
		WorkingDir: workingDir,
		AppPath:    "./my-app",
	}, fakeCommandRunner{})

	assertRecommendation(t, report, "Run guided init for this app:",
		"predeploy init --interactive --app ./my-app")
	assertRecommendation(t, report, "Then validate the generated config:",
		"predeploy validate predeploy.yaml")
}

func TestRunRecommendsGenericInteractiveInitWhenConfigAndAppAreMissing(t *testing.T) {
	report := RunWithRunner(context.Background(), Options{WorkingDir: t.TempDir()}, fakeCommandRunner{})

	assertRecommendation(t, report, "Run guided init:", "predeploy init --interactive")
}

func TestRunRecommendsValidateAndRunForValidConfig(t *testing.T) {
	workingDir := t.TempDir()
	configPath := filepath.Join(workingDir, "predeploy.yaml")
	writeDoctorTestConfig(t, configPath, `
service:
  name: test-service
  image: example/test-service:latest
  port: 8080
`)

	report := RunWithRunner(context.Background(), Options{
		WorkingDir: workingDir,
		ConfigPath: "predeploy.yaml",
	}, fakeCommandRunner{})

	assertRecommendation(t, report, "Validate the configuration:",
		"predeploy validate predeploy.yaml")
	assertRecommendation(t, report, "Run the configured checks:",
		"predeploy run predeploy.yaml")
}

func TestRunRecommendsIngressResponsibility(t *testing.T) {
	workingDir := t.TempDir()
	configPath := filepath.Join(workingDir, "predeploy.yaml")
	writeDoctorTestConfig(t, configPath, `
runtime:
  type: kubernetes
service:
  name: test-service
  image: example/test-service:latest
  port: 8080
gateway:
  enabled: true
  baseURL: http://predeploy.local
  ingress:
    enabled: true
  routes:
    - name: health
      path: /health
`)

	report := RunWithRunner(context.Background(), Options{
		WorkingDir: workingDir,
		ConfigPath: configPath,
	}, fakeCommandRunner{})

	assertRecommendation(t, report,
		"PreDeploy Guard does not install ingress controllers; ensure gateway.baseURL resolves to the ingress endpoint.",
		"")
}

func TestRunRecommendsDockerForEnabledPerformance(t *testing.T) {
	workingDir := t.TempDir()
	configPath := filepath.Join(workingDir, "predeploy.yaml")
	writeDoctorTestConfig(t, configPath, `
runtime:
  type: kubernetes
service:
  name: test-service
  image: example/test-service:latest
  port: 8080
checks:
  smoke:
    - name: health
      path: /health
performance:
  enabled: true
`)

	report := RunWithRunner(context.Background(), Options{
		WorkingDir: workingDir,
		ConfigPath: configPath,
	}, fakeCommandRunner{})

	assertRecommendation(t, report,
		"Dockerized k6 performance checks need an available Docker CLI and daemon.",
		"docker info")
}

func TestRunMakesConfiguredKubernetesChecksRequired(t *testing.T) {
	workingDir := t.TempDir()
	configPath := filepath.Join(workingDir, "predeploy.yaml")
	content := []byte(`
runtime:
  type: kubernetes
service:
  name: test-service
  image: example/test-service:latest
  port: 8080
`)
	if err := os.WriteFile(configPath, content, 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	report := RunWithRunner(context.Background(), Options{
		WorkingDir: workingDir,
		ConfigPath: configPath,
	}, fakeCommandRunner{})

	assertResult(t, report, "configured-runtime", StatusPass)
	assertResult(t, report, "kubectl", StatusFail)
	assertResult(t, report, "kubernetes-context", StatusFail)
	assertResult(t, report, "kubernetes-cluster", StatusFail)
}

func TestReportCounts(t *testing.T) {
	report := Report{Results: []CheckResult{
		{Status: StatusPass},
		{Status: StatusPass},
		{Status: StatusWarn},
		{Status: StatusFail},
	}}

	passed, warned, failed := report.Counts()
	if passed != 2 || warned != 1 || failed != 1 {
		t.Fatalf("Counts() = %d, %d, %d; want 2, 1, 1", passed, warned, failed)
	}
	if !report.HasFailures() {
		t.Fatal("HasFailures() = false, want true")
	}
}

func assertResult(t *testing.T, report Report, name string, status Status) CheckResult {
	t.Helper()
	for _, result := range report.Results {
		if result.Name != name {
			continue
		}
		if result.Status != status {
			t.Fatalf("result %q status = %q, want %q; message: %s", name, result.Status, status, result.Message)
		}
		return result
	}
	t.Fatalf("result %q not found in %#v", name, report.Results)
	return CheckResult{}
}

func assertRecommendation(t *testing.T, report Report, message string, command string) {
	t.Helper()
	for _, recommendation := range report.Recommendations {
		if recommendation.Message == message && recommendation.Command == command {
			return
		}
	}
	t.Fatalf("recommendation (%q, %q) not found in %#v", message, command, report.Recommendations)
}

func writeDoctorTestConfig(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
}
