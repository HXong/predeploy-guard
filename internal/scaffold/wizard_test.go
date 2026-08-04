package scaffold

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInitWizardAcceptsGenericDefaults(t *testing.T) {
	result, output, err := runWizardTest(InitOptions{}, strings.Repeat("\n", 9))
	if err != nil {
		t.Fatalf("RunInitWizard: %v", err)
	}
	if result.Cancelled {
		t.Fatal("Cancelled = true, want false")
	}
	if !result.Options.Guided {
		t.Fatal("Guided = false, want true")
	}
	if result.Options.AppPath != "" || result.Options.ServiceName != "my-service" ||
		result.Options.Runtime != "docker-compose" || result.Options.Image != "my-service:predeploy" ||
		result.Options.Port != 8080 || result.Options.HealthPath != "/health" ||
		result.Options.OutputPath != DefaultConfigFilename {
		t.Fatalf("unexpected options: %#v", result.Options)
	}
	for _, expected := range []string{
		"PreDeploy Guard guided init",
		"Application directory [leave blank to use generic starter config]:",
		"Generated config preview:",
		"Create this config? [Y/n]:",
	} {
		if !strings.Contains(output, expected) {
			t.Fatalf("output = %q, want %q", output, expected)
		}
	}
}

func TestInitWizardUsesProvidedAppDefaults(t *testing.T) {
	appPath := filepath.Join(t.TempDir(), "Orders API")
	if err := os.Mkdir(appPath, 0755); err != nil {
		t.Fatalf("Mkdir: %v", err)
	}
	mustWriteFile(t, filepath.Join(appPath, "Dockerfile"), "FROM scratch\n")
	mustWriteFile(t, filepath.Join(appPath, "package.json"), "{}\n")
	initial := InitOptions{
		AppPath:      appPath,
		Runtime:      "kubernetes",
		ServiceName:  "Selected Orders",
		Image:        "orders:test",
		Port:         3000,
		HealthPath:   "/ready",
		NoBuild:      true,
		Dependencies: "redis",
		OutputPath:   "selected.yaml",
	}

	result, output, err := runWizardTest(initial, strings.Repeat("\n", 10))
	if err != nil {
		t.Fatalf("RunInitWizard: %v", err)
	}
	if result.Options.ServiceName != "selected-orders" || result.Options.Runtime != "kubernetes" ||
		result.Options.Image != "orders:test" || result.Options.Port != 3000 ||
		result.Options.HealthPath != "/ready" || !result.Options.NoBuild ||
		result.Options.Dependencies != "redis" || result.Options.OutputPath != "selected.yaml" {
		t.Fatalf("unexpected options: %#v", result.Options)
	}
	for _, expected := range []string{
		"Detected application:",
		"Project indicators: Dockerfile, package.json",
		"Use Dockerfile build context? [y/N]:",
	} {
		if !strings.Contains(output, expected) {
			t.Fatalf("output = %q, want %q", output, expected)
		}
	}
}

func TestInitWizardGenericImageDefaultFollowsChangedServiceName(t *testing.T) {
	input := strings.Join([]string{
		"", "orders-api", "", "", "", "", "", "", "",
	}, "\n") + "\n"

	result, output, err := runWizardTest(InitOptions{}, input)
	if err != nil {
		t.Fatalf("RunInitWizard: %v", err)
	}
	if result.Options.Image != "orders-api:predeploy" {
		t.Fatalf("Image = %q, want orders-api:predeploy", result.Options.Image)
	}
	if !strings.Contains(output, "Image [orders-api:predeploy]:") {
		t.Fatalf("output = %q, want service-aware image prompt", output)
	}
}

func TestInitWizardAppImageDefaultFollowsChangedServiceName(t *testing.T) {
	appPath := t.TempDir()
	mustWriteFile(t, filepath.Join(appPath, "Dockerfile"), "FROM scratch\n")
	input := strings.Join([]string{
		"", "orders-api", "", "", "", "", "", "", "", "",
	}, "\n") + "\n"

	result, output, err := runWizardTest(InitOptions{AppPath: appPath}, input)
	if err != nil {
		t.Fatalf("RunInitWizard: %v", err)
	}
	if result.Options.Image != "predeploy-orders-api:local" {
		t.Fatalf("Image = %q, want predeploy-orders-api:local", result.Options.Image)
	}
	if !strings.Contains(output, "Image [predeploy-orders-api:local]:") {
		t.Fatalf("output = %q, want app-aware image prompt", output)
	}
}

func TestInitWizardProvidedImageIsPreservedAfterServiceChange(t *testing.T) {
	input := strings.Join([]string{
		"", "orders-api", "", "", "", "", "", "", "",
	}, "\n") + "\n"

	result, output, err := runWizardTest(InitOptions{Image: "custom/image:test"}, input)
	if err != nil {
		t.Fatalf("RunInitWizard: %v", err)
	}
	if result.Options.Image != "custom/image:test" {
		t.Fatalf("Image = %q, want custom/image:test", result.Options.Image)
	}
	if !strings.Contains(output, "Image [custom/image:test]:") {
		t.Fatalf("output = %q, want provided image prompt", output)
	}
}

func TestInitWizardSetsNoBuildWhenAnswerIsNo(t *testing.T) {
	appPath := t.TempDir()
	mustWriteFile(t, filepath.Join(appPath, "Dockerfile"), "FROM scratch\n")
	input := strings.Join([]string{
		"", "", "", "", "", "", "no", "", "", "",
	}, "\n") + "\n"

	result, _, err := runWizardTest(InitOptions{AppPath: appPath}, input)
	if err != nil {
		t.Fatalf("RunInitWizard: %v", err)
	}
	if !result.Options.NoBuild {
		t.Fatal("NoBuild = false, want true")
	}
}

func TestInitWizardWarnsWhenDockerfileIsMissing(t *testing.T) {
	appPath := t.TempDir()
	input := strings.Repeat("\n", 9)

	result, output, err := runWizardTest(InitOptions{AppPath: appPath}, input)
	if err != nil {
		t.Fatalf("RunInitWizard: %v", err)
	}
	if !result.Options.NoBuild {
		t.Fatal("NoBuild = false, want true")
	}
	if !strings.Contains(output, "Dockerfile not found. Generated config will reference an image only.") {
		t.Fatalf("output = %q, want Dockerfile warning", output)
	}
}

func TestInitWizardRepromptsInvalidRuntime(t *testing.T) {
	input := strings.Join([]string{
		"", "", "invalid", "kubernetes", "", "", "", "", "", "",
	}, "\n") + "\n"

	result, output, err := runWizardTest(InitOptions{}, input)
	if err != nil {
		t.Fatalf("RunInitWizard: %v", err)
	}
	if result.Options.Runtime != "kubernetes" {
		t.Fatalf("Runtime = %q, want kubernetes", result.Options.Runtime)
	}
	if !strings.Contains(output, "Please enter docker-compose or kubernetes.") {
		t.Fatalf("output = %q, want validation message", output)
	}
}

func TestInitWizardRepromptsInvalidPort(t *testing.T) {
	input := strings.Join([]string{
		"", "", "", "", "not-a-port", "70000", "3000", "", "", "", "",
	}, "\n") + "\n"

	result, output, err := runWizardTest(InitOptions{}, input)
	if err != nil {
		t.Fatalf("RunInitWizard: %v", err)
	}
	if result.Options.Port != 3000 {
		t.Fatalf("Port = %d, want 3000", result.Options.Port)
	}
	if strings.Count(output, "Please enter a port from 1 to 65535.") != 2 {
		t.Fatalf("output = %q, want two port validation messages", output)
	}
}

func TestInitWizardRepromptsInvalidHealthPath(t *testing.T) {
	input := strings.Join([]string{
		"", "", "", "", "", "ready", "/ready", "", "", "",
	}, "\n") + "\n"

	result, output, err := runWizardTest(InitOptions{}, input)
	if err != nil {
		t.Fatalf("RunInitWizard: %v", err)
	}
	if result.Options.HealthPath != "/ready" {
		t.Fatalf("HealthPath = %q, want /ready", result.Options.HealthPath)
	}
	if !strings.Contains(output, "Please enter a health path that starts with /.") {
		t.Fatalf("output = %q, want validation message", output)
	}
}

func TestInitWizardAcceptsAndCanonicalizesDependencyPresets(t *testing.T) {
	input := strings.Join([]string{
		"", "", "", "", "", "", "redis,postgres", "", "",
	}, "\n") + "\n"

	result, _, err := runWizardTest(InitOptions{}, input)
	if err != nil {
		t.Fatalf("RunInitWizard: %v", err)
	}
	if result.Options.Dependencies != "postgres,redis" {
		t.Fatalf("Dependencies = %q, want postgres,redis", result.Options.Dependencies)
	}
}

func TestInitWizardRepromptsInvalidDependencyPreset(t *testing.T) {
	input := strings.Join([]string{
		"", "", "", "", "", "", "mysql", "postgres", "", "",
	}, "\n") + "\n"

	result, output, err := runWizardTest(InitOptions{}, input)
	if err != nil {
		t.Fatalf("RunInitWizard: %v", err)
	}
	if result.Options.Dependencies != "postgres" {
		t.Fatalf("Dependencies = %q, want postgres", result.Options.Dependencies)
	}
	if !strings.Contains(output, "Please enter none, postgres, redis, or postgres,redis.") {
		t.Fatalf("output = %q, want dependency validation message", output)
	}
}

func TestInitWizardCanCancel(t *testing.T) {
	input := strings.Join([]string{
		"", "", "", "", "", "", "", "", "no",
	}, "\n") + "\n"

	result, _, err := runWizardTest(InitOptions{}, input)
	if err != nil {
		t.Fatalf("RunInitWizard: %v", err)
	}
	if !result.Cancelled {
		t.Fatal("Cancelled = false, want true")
	}
}

func TestInitWizardFailsForMissingAppBeforeConfirmation(t *testing.T) {
	missingPath := filepath.Join(t.TempDir(), "missing")
	result, _, err := runWizardTest(InitOptions{}, missingPath+"\n")
	if err == nil {
		t.Fatal("RunInitWizard error = nil, want app detection error")
	}
	if result.Options.OutputPath != "" {
		t.Fatalf("result = %#v, want empty result", result)
	}
}

func runWizardTest(initial InitOptions, input string) (WizardResult, string, error) {
	var output bytes.Buffer
	result, err := RunInitWizard(WizardOptions{
		Initial: initial,
		Input:   strings.NewReader(input),
		Output:  &output,
	})
	return result, output.String(), err
}
