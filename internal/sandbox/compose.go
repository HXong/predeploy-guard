package sandbox

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/HXong/predeploy-guard/internal/config"
)

const DefaultHostPort = 8080

type ComposeSandbox struct {
	WorkDir     string
	ComposeFile string
	HostPort    int
	ServiceName string
}

func NewComposeSandbox(cfg *config.Config) (*ComposeSandbox, error) {
	workDir, err := os.MkdirTemp("", "predeploy-guard-*")
	if err != nil {
		return nil, fmt.Errorf("create temporary sandbox directory: %w", err)
	}

	sandbox := &ComposeSandbox{
		WorkDir:     workDir,
		ComposeFile: filepath.Join(workDir, "docker-compose.yml"),
		HostPort:    DefaultHostPort,
		ServiceName: sanitizeServiceName(cfg.Service.Name),
	}

	if err := sandbox.writeComposeFile(cfg); err != nil {
		_ = os.RemoveAll(workDir)
		return nil, err
	}

	return sandbox, nil
}

func (s *ComposeSandbox) Start() error {
	cmd := exec.Command("docker", "compose", "-f", s.ComposeFile, "up", "-d")
	cmd.Dir = s.WorkDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("docker compose up failed: %w\nOutput:\n%s", err, string(output))
	}

	return nil
}

func (s *ComposeSandbox) Stop() error {
	cmd := exec.Command("docker", "compose", "-f", s.ComposeFile, "down", "--remove-orphans")
	cmd.Dir = s.WorkDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("docker compose down failed: %w\nOutput:\n%s", err, string(output))
	}

	return nil
}

func (s *ComposeSandbox) Logs() (string, error) {
	cmd := exec.Command("docker", "compose", "-f", s.ComposeFile, "logs", "--no-color")
	cmd.Dir = s.WorkDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("docker compose logs failed: %w", err)
	}

	return string(output), nil
}

func (s *ComposeSandbox) RemoveFiles() error {
	if s.WorkDir == "" {
		return nil
	}

	return os.RemoveAll(s.WorkDir)
}

func (s *ComposeSandbox) BaseURL() string {
	return fmt.Sprintf("http://localhost:%d", s.HostPort)
}
