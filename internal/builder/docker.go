package builder

import (
	"fmt"
	"os/exec"

	"github.com/HXong/predeploy-guard/internal/config"
)

type BuildResult struct {
	Enabled    bool
	Image      string
	Context    string
	Dockerfile string
	Passed     bool
	Error      string
	Output     string
}

func BuildImageIfNeeded(cfg *config.Config) BuildResult {
	build := cfg.Service.Build

	result := BuildResult{
		Enabled:    build.Context != "",
		Image:      cfg.Service.Image,
		Context:    build.Context,
		Dockerfile: build.Dockerfile,
	}

	if !result.Enabled {
		result.Passed = true
		return result
	}

	cmd := exec.Command(
		"docker",
		"build",
		"-t",
		cfg.Service.Image,
		"-f",
		build.Dockerfile,
		".",
	)

	cmd.Dir = build.Context
	output, err := cmd.CombinedOutput()
	result.Output = string(output)

	if err != nil {
		result.Passed = false
		result.Error = fmt.Sprintf("docker build failed: %v", err)
		return result
	}
	result.Passed = true
	return result
}
