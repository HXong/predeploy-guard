package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/HXong/predeploy-guard/internal/appdetect"
	"github.com/HXong/predeploy-guard/internal/scaffold"
)

func TestPrintInitResultAppAwareOutput(t *testing.T) {
	options := scaffold.InitOptions{
		OutputPath: "predeploy.yaml",
		AppPath:    "./my-app",
	}
	result := scaffold.InitResult{
		Detection: appdetect.DetectionResult{
			HasDockerfile:     true,
			ProjectIndicators: []string{"Dockerfile", "package.json"},
		},
		Runtime:         "docker-compose",
		ServiceName:     "my-app",
		Image:           "predeploy-my-app:local",
		Port:            8080,
		HealthPath:      "/health",
		BuildContext:    "./my-app",
		BuildConfigured: true,
	}

	var output bytes.Buffer
	printInitResult(&output, options, result)

	for _, expected := range []string{
		"Created predeploy.yaml",
		"Detected application:",
		"Project indicators: Dockerfile, package.json",
		"Build context: ./my-app",
		"Next steps:\n  1. Run: predeploy doctor --config predeploy.yaml --app ./my-app\n" +
			"  2. Run: predeploy validate predeploy.yaml\n" +
			"  3. Run: predeploy explain predeploy.yaml\n" +
			"  4. Run: predeploy run predeploy.yaml",
	} {
		if !strings.Contains(output.String(), expected) {
			t.Fatalf("output = %q, want %q", output.String(), expected)
		}
	}
}

func TestPrintInitResultGenericFirstRunFlow(t *testing.T) {
	var output bytes.Buffer
	printInitResult(&output, scaffold.InitOptions{OutputPath: "predeploy.yaml"}, scaffold.InitResult{})

	expected := "Next steps:\n" +
		"  1. Review predeploy.yaml\n" +
		"  2. Run: predeploy doctor --config predeploy.yaml\n" +
		"  3. Run: predeploy validate predeploy.yaml\n" +
		"  4. Run: predeploy explain predeploy.yaml\n" +
		"  5. Run: predeploy run predeploy.yaml"
	if !strings.Contains(output.String(), expected) {
		t.Fatalf("output = %q, want first-run flow %q", output.String(), expected)
	}
	if strings.Index(output.String(), "Optional profile commands:") < strings.Index(output.String(), expected) {
		t.Fatalf("profile commands appeared before first-run flow: %q", output.String())
	}
}

func TestPrintInitResultWarnings(t *testing.T) {
	options := scaffold.InitOptions{AppPath: "./my-app"}
	result := scaffold.InitResult{
		Runtime:     "docker-compose",
		ServiceName: "my-app",
		Image:       "predeploy-my-app:local",
		Port:        8080,
		HealthPath:  "/health",
		Warnings: []scaffold.InitWarning{{
			Message: "Dockerfile not found.",
			Details: "Use an existing image.",
		}},
	}

	var output bytes.Buffer
	printInitResult(&output, options, result)
	if !strings.Contains(output.String(), "[WARN] Dockerfile not found.\n       Use an existing image.") {
		t.Fatalf("output = %q, want warning", output.String())
	}
}
