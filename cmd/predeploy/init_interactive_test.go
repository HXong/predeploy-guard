package main

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/HXong/predeploy-guard/internal/config"
	"github.com/HXong/predeploy-guard/internal/scaffold"
)

func TestRunInteractiveInitWritesValidAppAwareConfigWithoutChangingApp(t *testing.T) {
	root := t.TempDir()
	appPath := filepath.Join(root, "my-app")
	if err := os.Mkdir(appPath, 0755); err != nil {
		t.Fatalf("Mkdir: %v", err)
	}
	writeInteractiveTestFile(t, filepath.Join(appPath, "Dockerfile"), "FROM scratch\n")
	writeInteractiveTestFile(t, filepath.Join(appPath, "package.json"), "{}\n")
	writeInteractiveTestFile(t, filepath.Join(appPath, ".env"), "SECRET=must-not-be-read\n")
	before := snapshotInteractiveTestDirectory(t, appPath)
	outputPath := filepath.Join(root, "predeploy.yaml")

	input := strings.NewReader(strings.Repeat("\n", 10))
	var output bytes.Buffer
	err := runInteractiveInit(input, &output, scaffold.InitOptions{
		AppPath:    appPath,
		OutputPath: outputPath,
	})
	if err != nil {
		t.Fatalf("runInteractiveInit: %v", err)
	}

	cfg, err := config.Load(outputPath)
	if err != nil {
		t.Fatalf("config.Load: %v", err)
	}
	if cfg.Service.Name != "my-app" || cfg.Service.Build.Context != appPath {
		t.Fatalf("unexpected generated service: %#v", cfg.Service)
	}
	if !strings.Contains(output.String(), "Created "+outputPath) {
		t.Fatalf("output = %q, want created message", output.String())
	}
	after := snapshotInteractiveTestDirectory(t, appPath)
	if !reflect.DeepEqual(after, before) {
		t.Fatalf("app changed:\nbefore=%#v\nafter=%#v", before, after)
	}
}

func TestRunInteractiveInitWritesValidGenericConfig(t *testing.T) {
	outputPath := filepath.Join(t.TempDir(), "predeploy.yaml")
	var output bytes.Buffer

	err := runInteractiveInit(
		strings.NewReader(strings.Repeat("\n", 9)),
		&output,
		scaffold.InitOptions{OutputPath: outputPath},
	)
	if err != nil {
		t.Fatalf("runInteractiveInit: %v", err)
	}
	if _, err := config.Load(outputPath); err != nil {
		t.Fatalf("config.Load: %v", err)
	}
}

func TestRunInteractiveInitCancellationWritesNothing(t *testing.T) {
	outputPath := filepath.Join(t.TempDir(), "predeploy.yaml")
	input := strings.Join([]string{
		"", "", "", "", "", "", "", "", "no",
	}, "\n") + "\n"
	var output bytes.Buffer

	if err := runInteractiveInit(
		strings.NewReader(input),
		&output,
		scaffold.InitOptions{OutputPath: outputPath},
	); err != nil {
		t.Fatalf("runInteractiveInit: %v", err)
	}
	if _, err := os.Stat(outputPath); !os.IsNotExist(err) {
		t.Fatalf("output exists after cancellation: %v", err)
	}
	if !strings.Contains(output.String(), "Init cancelled. No files were written.") {
		t.Fatalf("output = %q, want cancellation message", output.String())
	}
}

func TestRunInteractiveInitDoesNotOverwriteWithoutForce(t *testing.T) {
	outputPath := filepath.Join(t.TempDir(), "predeploy.yaml")
	writeInteractiveTestFile(t, outputPath, "keep me\n")
	var output bytes.Buffer

	err := runInteractiveInit(
		strings.NewReader(strings.Repeat("\n", 9)),
		&output,
		scaffold.InitOptions{OutputPath: outputPath},
	)
	if err == nil || !strings.Contains(err.Error(), "use --force") {
		t.Fatalf("runInteractiveInit error = %v, want overwrite error", err)
	}
	content, readErr := os.ReadFile(outputPath)
	if readErr != nil {
		t.Fatalf("ReadFile: %v", readErr)
	}
	if string(content) != "keep me\n" {
		t.Fatalf("content = %q, want unchanged", content)
	}
}

func writeInteractiveTestFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatalf("WriteFile(%s): %v", path, err)
	}
}

func snapshotInteractiveTestDirectory(t *testing.T, path string) map[string]string {
	t.Helper()
	entries, err := os.ReadDir(path)
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	result := make(map[string]string, len(entries))
	for _, entry := range entries {
		content, readErr := os.ReadFile(filepath.Join(path, entry.Name()))
		if readErr != nil {
			t.Fatalf("ReadFile(%s): %v", entry.Name(), readErr)
		}
		result[entry.Name()] = string(content)
	}
	return result
}
