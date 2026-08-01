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
		"predeploy doctor --config predeploy.yaml --app ./my-app",
	} {
		if !strings.Contains(output.String(), expected) {
			t.Fatalf("output = %q, want %q", output.String(), expected)
		}
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
