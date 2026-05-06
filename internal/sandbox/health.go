package sandbox

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/HXong/predeploy-guard/internal/config"
)

type DependencyReadinessResult struct {
	Name    string
	Passed  bool
	Skipped bool
	Error   string
}

func (s *ComposeSandbox) WaitForDependencies(cfg *config.Config) ([]DependencyReadinessResult, error) {
	results := make([]DependencyReadinessResult, 0, len(cfg.Dependencies))

	for _, name := range sortedDependencyNames(cfg.Dependencies) {
		dependency := cfg.Dependencies[name]
		serviceName := sanitizeServiceName(name)

		result := DependencyReadinessResult{
			Name: serviceName,
		}

		if !dependencyHasReadiness(dependency.Readiness) {
			fmt.Printf("Skipping dependency readiness: %s has no readiness command\n", name)
			result.Passed = true
			result.Skipped = true
			results = append(results, result)
			continue
		}

		containerName := fmt.Sprintf("predeploy-%s", serviceName)

		fmt.Printf("Waiting for dependency readiness: %s\n", serviceName)

		if err := waitForContainerHealthy(
			containerName,
			dependency.Readiness.TimeoutSeconds,
		); err != nil {
			result.Passed = false
			result.Error = err.Error()
			results = append(results, result)
			return results, err
		}

		result.Passed = true
		results = append(results, result)

		fmt.Printf("Dependency ready: %s\n", serviceName)
	}

	return results, nil
}

func waitForContainerHealthy(containerName string, timeoutSeconds int) error {
	deadline := time.Now().Add(time.Duration(timeoutSeconds) * time.Second)

	for time.Now().Before(deadline) {
		status, err := getContainerHealthStatus(containerName)
		if err != nil {
			return err
		}

		switch status {
		case "healthy":
			return nil
		case "unhealthy":
			return fmt.Errorf("container %s became unhealthy", containerName)
		case "starting", "none", "":
			time.Sleep(1 * time.Second)
		default:
			time.Sleep(1 * time.Second)
		}
	}

	return fmt.Errorf("container %s did not become healthy within %d seconds", containerName, timeoutSeconds)
}

func getContainerHealthStatus(containerName string) (string, error) {
	cmd := exec.Command(
		"docker",
		"inspect",
		"--format",
		"{{if .State.Health}}{{.State.Health.Status}}{{else}}none{{end}}",
		containerName,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("docker inspect failed for %s: %w\nOutput:\n%s", containerName, err, string(output))
	}

	return strings.TrimSpace(string(output)), nil
}
