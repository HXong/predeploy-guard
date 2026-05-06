package sandbox

import (
	"bytes"
	"fmt"
	"os"

	"github.com/HXong/predeploy-guard/internal/config"
)

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
