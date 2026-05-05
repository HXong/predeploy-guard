package sandbox

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

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

func (s *ComposeSandbox) RemoveFiles() error {
	if s.WorkDir == "" {
		return nil
	}

	return os.RemoveAll(s.WorkDir)
}

func (s *ComposeSandbox) BaseURL() string {
	return fmt.Sprintf("http://localhost:%d", s.HostPort)
}

func (s *ComposeSandbox) writeComposeFile(cfg *config.Config) error {
	var envBlock bytes.Buffer

	for key, value := range cfg.Service.Env {
		envBlock.WriteString(fmt.Sprintf("      %s: %q\n", key, value))
	}

	environmentSection := ""
	if len(cfg.Service.Env) > 0 {
		environmentSection = fmt.Sprintf("    environment:\n%s", envBlock.String())
	}

	content := fmt.Sprintf(`services:
  %s:
    image: %s
    container_name: predeploy-%s
    ports:
      - "%d:%d"
%s
`, s.ServiceName, cfg.Service.Image, s.ServiceName, s.HostPort, cfg.Service.Port, environmentSection)

	if err := os.WriteFile(s.ComposeFile, []byte(content), 0644); err != nil {
		return fmt.Errorf("write docker-compose.yml: %w", err)
	}

	return nil
}

func sanitizeServiceName(name string) string {
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, "_", "-")
	name = strings.ReplaceAll(name, " ", "-")

	var builder strings.Builder

	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			builder.WriteRune(r)
		}
	}

	result := builder.String()
	if result == "" {
		return "service"
	}

	return result
}
