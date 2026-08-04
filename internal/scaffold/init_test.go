package scaffold

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/HXong/predeploy-guard/internal/config"
)

func TestAppAwareInitWritesBuildContextAndLoads(t *testing.T) {
	root := t.TempDir()
	appPath := filepath.Join(root, "My Orders API")
	mustMkdir(t, appPath)
	mustWriteFile(t, filepath.Join(appPath, "Dockerfile"), "FROM scratch\n")
	mustWriteFile(t, filepath.Join(appPath, "package.json"), "{}\n")
	outputPath := filepath.Join(root, "predeploy.yaml")

	result, err := WriteConfig(InitOptions{
		OutputPath: outputPath,
		AppPath:    appPath,
	})
	if err != nil {
		t.Fatalf("WriteConfig: %v", err)
	}
	if !result.BuildConfigured {
		t.Fatal("BuildConfigured = false, want true")
	}
	if result.BuildContext != "./My Orders API" {
		t.Fatalf("BuildContext = %q, want %q", result.BuildContext, "./My Orders API")
	}
	if result.ServiceName != "my-orders-api" {
		t.Fatalf("ServiceName = %q, want my-orders-api", result.ServiceName)
	}
	if result.Image != "predeploy-my-orders-api:local" {
		t.Fatalf("Image = %q", result.Image)
	}

	cfg, err := config.Load(outputPath)
	if err != nil {
		t.Fatalf("config.Load: %v", err)
	}
	if cfg.Service.Build.Context != appPath {
		t.Fatalf("resolved build context = %q, want %q", cfg.Service.Build.Context, appPath)
	}
	if cfg.Performance.Enabled {
		t.Fatal("Performance.Enabled = true, want false")
	}
	if cfg.Checks.Smoke[0].Path != "/health" {
		t.Fatalf("smoke path = %q, want /health", cfg.Checks.Smoke[0].Path)
	}
}

func TestAppAwareInitUsesProvidedValues(t *testing.T) {
	root := t.TempDir()
	appPath := filepath.Join(root, "app")
	mustMkdir(t, appPath)
	mustWriteFile(t, filepath.Join(appPath, "Dockerfile"), "FROM scratch\n")
	outputPath := filepath.Join(root, "predeploy.yaml")

	result, err := WriteConfig(InitOptions{
		OutputPath:  outputPath,
		AppPath:     appPath,
		Runtime:     "kubernetes",
		ServiceName: "Orders API!",
		Image:       "orders-api:local",
		Port:        9090,
		HealthPath:  "/ready",
	})
	if err != nil {
		t.Fatalf("WriteConfig: %v", err)
	}
	if result.Runtime != "kubernetes" || result.ServiceName != "orders-api" ||
		result.Image != "orders-api:local" || result.Port != 9090 || result.HealthPath != "/ready" {
		t.Fatalf("unexpected result: %#v", result)
	}
	if result.DefaultedPort {
		t.Fatal("DefaultedPort = true, want false")
	}

	cfg, err := config.Load(outputPath)
	if err != nil {
		t.Fatalf("config.Load: %v", err)
	}
	if cfg.Runtime.Type != "kubernetes" || cfg.Checks.Smoke[0].Path != "/ready" {
		t.Fatalf("unexpected config: runtime=%q smoke=%q", cfg.Runtime.Type, cfg.Checks.Smoke[0].Path)
	}
}

func TestAppAwareInitNoBuildOmitsBuildContext(t *testing.T) {
	root := t.TempDir()
	appPath := filepath.Join(root, "app")
	mustMkdir(t, appPath)
	mustWriteFile(t, filepath.Join(appPath, "Dockerfile"), "FROM scratch\n")
	outputPath := filepath.Join(root, "predeploy.yaml")

	result, err := WriteConfig(InitOptions{
		OutputPath: outputPath,
		AppPath:    appPath,
		NoBuild:    true,
	})
	if err != nil {
		t.Fatalf("WriteConfig: %v", err)
	}
	if result.BuildConfigured || result.BuildContext != "" {
		t.Fatalf("unexpected build result: %#v", result)
	}

	cfg, err := config.Load(outputPath)
	if err != nil {
		t.Fatalf("config.Load: %v", err)
	}
	if cfg.Service.Build.Context != "" || cfg.Service.Build.Dockerfile != "" {
		t.Fatalf("build config = %#v, want empty", cfg.Service.Build)
	}
}

func TestAppAwareInitMissingDockerfileOmitsBuildAndWarns(t *testing.T) {
	root := t.TempDir()
	appPath := filepath.Join(root, "app")
	mustMkdir(t, appPath)
	mustWriteFile(t, filepath.Join(appPath, "go.mod"), "module example.test/app\n")

	result, err := WriteConfig(InitOptions{
		OutputPath: filepath.Join(root, "predeploy.yaml"),
		AppPath:    appPath,
	})
	if err != nil {
		t.Fatalf("WriteConfig: %v", err)
	}
	if result.BuildConfigured {
		t.Fatal("BuildConfigured = true, want false")
	}
	if !hasWarning(result, "Dockerfile not found") {
		t.Fatalf("Warnings = %#v, want Dockerfile warning", result.Warnings)
	}
}

func TestInitWithoutAppPreservesDefaultConfig(t *testing.T) {
	outputPath := filepath.Join(t.TempDir(), "predeploy.yaml")
	result, err := WriteConfig(InitOptions{OutputPath: outputPath})
	if err != nil {
		t.Fatalf("WriteConfig: %v", err)
	}
	if result.AppPath != "" {
		t.Fatalf("AppPath = %q, want empty", result.AppPath)
	}

	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(content) != buildDefaultConfig(nil) {
		t.Fatal("default config changed when AppPath was empty")
	}
}

func TestAppAwareInitRefusesOverwriteWithoutForce(t *testing.T) {
	root := t.TempDir()
	appPath := filepath.Join(root, "app")
	mustMkdir(t, appPath)
	outputPath := filepath.Join(root, "predeploy.yaml")
	mustWriteFile(t, outputPath, "keep me\n")

	_, err := WriteConfig(InitOptions{OutputPath: outputPath, AppPath: appPath})
	if err == nil || !strings.Contains(err.Error(), "use --force") {
		t.Fatalf("WriteConfig error = %v, want overwrite error", err)
	}
	content, readErr := os.ReadFile(outputPath)
	if readErr != nil {
		t.Fatalf("ReadFile: %v", readErr)
	}
	if string(content) != "keep me\n" {
		t.Fatalf("existing content = %q, want unchanged", content)
	}
	if _, forceErr := WriteConfig(InitOptions{
		OutputPath: outputPath,
		AppPath:    appPath,
		Overwrite:  true,
	}); forceErr != nil {
		t.Fatalf("WriteConfig with force: %v", forceErr)
	}
}

func TestAppAwareInitRelativeContextFromAnotherDirectory(t *testing.T) {
	root := t.TempDir()
	appPath := filepath.Join(root, "apps", "orders")
	outputDir := filepath.Join(root, "configs", "local")
	mustMkdir(t, appPath)
	mustMkdir(t, outputDir)
	mustWriteFile(t, filepath.Join(appPath, "Dockerfile"), "FROM scratch\n")

	result, err := WriteConfig(InitOptions{
		OutputPath: filepath.Join(outputDir, "predeploy.yaml"),
		AppPath:    appPath,
	})
	if err != nil {
		t.Fatalf("WriteConfig: %v", err)
	}
	expected := filepath.ToSlash(filepath.Join("..", "..", "apps", "orders"))
	if result.BuildContext != expected {
		t.Fatalf("BuildContext = %q, want %q", result.BuildContext, expected)
	}
}

func TestAppAwareInitFailsBeforeWritingForMissingAppAndInvalidRuntime(t *testing.T) {
	root := t.TempDir()

	tests := []InitOptions{
		{OutputPath: filepath.Join(root, "missing-app.yaml"), AppPath: filepath.Join(root, "missing")},
		{OutputPath: filepath.Join(root, "invalid-runtime.yaml"), AppPath: root, Runtime: "production"},
		{OutputPath: filepath.Join(root, "invalid-port.yaml"), AppPath: root, Port: 65536},
	}
	for _, options := range tests {
		if _, err := WriteConfig(options); err == nil {
			t.Fatalf("WriteConfig(%#v) error = nil", options)
		}
		if _, err := os.Stat(options.OutputPath); !os.IsNotExist(err) {
			t.Fatalf("output %q exists after failed init", options.OutputPath)
		}
	}
}

func TestAppAwareInitDoesNotModifyAppDirectory(t *testing.T) {
	root := t.TempDir()
	appPath := filepath.Join(root, "app")
	mustMkdir(t, appPath)
	mustWriteFile(t, filepath.Join(appPath, "Dockerfile"), "FROM scratch\n")
	mustWriteFile(t, filepath.Join(appPath, ".env"), "SECRET=do-not-read\n")
	mustWriteFile(t, filepath.Join(appPath, "package.json"), "{\"private\":true}\n")
	before := snapshotDirectory(t, appPath)

	if _, err := WriteConfig(InitOptions{
		OutputPath: filepath.Join(root, "predeploy.yaml"),
		AppPath:    appPath,
	}); err != nil {
		t.Fatalf("WriteConfig: %v", err)
	}

	after := snapshotDirectory(t, appPath)
	if !reflect.DeepEqual(after, before) {
		t.Fatalf("app directory changed:\nbefore=%#v\nafter=%#v", before, after)
	}
}

func TestAppAwareInitRejectsOutputInsideAppDirectory(t *testing.T) {
	appPath := t.TempDir()
	outputPath := filepath.Join(appPath, "predeploy.yaml")

	_, err := WriteConfig(InitOptions{OutputPath: outputPath, AppPath: appPath})
	if err == nil || !strings.Contains(err.Error(), "outside the app directory") {
		t.Fatalf("WriteConfig error = %v, want app-directory safety error", err)
	}
	expectedSuggestion := "Try:\n  predeploy init --app " + commandLineArgument(appPath) +
		" --output " + commandLineArgument(safeOutputPath(appPath))
	if !strings.Contains(err.Error(), expectedSuggestion) {
		t.Fatalf("WriteConfig error = %q, want suggestion %q", err, expectedSuggestion)
	}
	if _, statErr := os.Stat(outputPath); !os.IsNotExist(statErr) {
		t.Fatalf("output exists after safety failure: %v", statErr)
	}
}

func hasWarning(result InitResult, messagePart string) bool {
	for _, warning := range result.Warnings {
		if strings.Contains(warning.Message, messagePart) {
			return true
		}
	}
	return false
}

func mustMkdir(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0755); err != nil {
		t.Fatalf("MkdirAll(%s): %v", path, err)
	}
}

func mustWriteFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatalf("WriteFile(%s): %v", path, err)
	}
}

func snapshotDirectory(t *testing.T, path string) map[string]string {
	t.Helper()
	entries, err := os.ReadDir(path)
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	snapshot := make(map[string]string, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			snapshot[entry.Name()] = "<directory>"
			continue
		}
		content, readErr := os.ReadFile(filepath.Join(path, entry.Name()))
		if readErr != nil {
			t.Fatalf("ReadFile(%s): %v", entry.Name(), readErr)
		}
		snapshot[entry.Name()] = string(content)
	}
	return snapshot
}
