package appdetect

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestDetectDockerfile(t *testing.T) {
	appPath := t.TempDir()
	writeTestFile(t, filepath.Join(appPath, "Dockerfile"), "FROM scratch\n")

	result, err := Detect(Options{AppPath: appPath})
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}
	if !result.HasDockerfile {
		t.Fatal("HasDockerfile = false, want true")
	}
	if result.DockerfilePath != filepath.Join(appPath, "Dockerfile") {
		t.Fatalf("DockerfilePath = %q", result.DockerfilePath)
	}
	if !reflect.DeepEqual(result.ProjectTypes, []ProjectType{ProjectDocker}) {
		t.Fatalf("ProjectTypes = %#v", result.ProjectTypes)
	}
}

func TestDetectPackageJSON(t *testing.T) {
	appPath := t.TempDir()
	writeTestFile(t, filepath.Join(appPath, "package.json"), "{}\n")

	result, err := Detect(Options{AppPath: appPath})
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}
	if !reflect.DeepEqual(result.ProjectTypes, []ProjectType{ProjectNode}) {
		t.Fatalf("ProjectTypes = %#v", result.ProjectTypes)
	}
	if !reflect.DeepEqual(result.ProjectIndicators, []string{"package.json"}) {
		t.Fatalf("ProjectIndicators = %#v", result.ProjectIndicators)
	}
}

func TestDetectGoMod(t *testing.T) {
	appPath := t.TempDir()
	writeTestFile(t, filepath.Join(appPath, "go.mod"), "module example.test/app\n")

	result, err := Detect(Options{AppPath: appPath})
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}
	if !reflect.DeepEqual(result.ProjectTypes, []ProjectType{ProjectGo}) {
		t.Fatalf("ProjectTypes = %#v", result.ProjectTypes)
	}
}

func TestDetectEmptyDirectoryReturnsUnknown(t *testing.T) {
	result, err := Detect(Options{AppPath: t.TempDir()})
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}
	if !reflect.DeepEqual(result.ProjectTypes, []ProjectType{ProjectUnknown}) {
		t.Fatalf("ProjectTypes = %#v", result.ProjectTypes)
	}
	if len(result.ProjectIndicators) != 0 {
		t.Fatalf("ProjectIndicators = %#v, want empty", result.ProjectIndicators)
	}
}

func TestDetectMissingPath(t *testing.T) {
	_, err := Detect(Options{AppPath: filepath.Join(t.TempDir(), "missing")})
	if err == nil {
		t.Fatal("Detect error = nil, want missing path error")
	}
}

func TestDetectRejectsFilePath(t *testing.T) {
	path := filepath.Join(t.TempDir(), "app.txt")
	writeTestFile(t, path, "not a directory\n")

	_, err := Detect(Options{AppPath: path})
	if err == nil {
		t.Fatal("Detect error = nil, want directory error")
	}
}

func TestSanitizeServiceName(t *testing.T) {
	tests := map[string]string{
		"My Orders API":       "my-orders-api",
		"UPPER___and...lower": "upper-and-lower",
		"--- symbols ---":     "symbols",
		"!!!":                 "app",
		"東京":                  "app",
	}

	for input, expected := range tests {
		t.Run(input, func(t *testing.T) {
			if actual := SanitizeServiceName(input); actual != expected {
				t.Fatalf("SanitizeServiceName(%q) = %q, want %q", input, actual, expected)
			}
		})
	}
}

func writeTestFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatalf("WriteFile(%s): %v", path, err)
	}
}
