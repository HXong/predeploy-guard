package sandbox

import (
	"bytes"
	"fmt"
	"sort"

	"github.com/HXong/predeploy-guard/internal/config"
)

func (s *ComposeSandbox) writeComposeFile(cfg *config.Config) error {
	var builder bytes.Buffer

	builder.WriteString("services:\n")

	writeTargetService(&builder, s, cfg)

	if len(cfg.Dependencies) > 0 {
		writeDependencyServices(&builder, cfg)
	}

	if err := writeFile(s.ComposeFile, builder.String()); err != nil {
		return fmt.Errorf("write docker-compose.yml: %w", err)
	}

	return nil
}

func writeTargetService(builder *bytes.Buffer, s *ComposeSandbox, cfg *config.Config) {
	builder.WriteString(fmt.Sprintf("  %s:\n", s.ServiceName))
	builder.WriteString(fmt.Sprintf("    image: %s\n", cfg.Service.Image))
	builder.WriteString(fmt.Sprintf("    container_name: predeploy-%s\n", s.ServiceName))
	builder.WriteString("    ports:\n")
	builder.WriteString(fmt.Sprintf("      - \"%d:%d\"\n", s.HostPort, cfg.Service.Port))

	if len(cfg.Dependencies) > 0 {
		builder.WriteString("    depends_on:\n")

		for _, name := range sortedDependencyNames(cfg.Dependencies) {
			builder.WriteString(fmt.Sprintf("      - %s\n", sanitizeServiceName(name)))
		}
	}

	writeEnvBlock(builder, cfg.Service.Env)
}

func writeDependencyServices(builder *bytes.Buffer, cfg *config.Config) {
	for _, name := range sortedDependencyNames(cfg.Dependencies) {
		dependency := cfg.Dependencies[name]
		serviceName := sanitizeServiceName(name)

		builder.WriteString(fmt.Sprintf("  %s:\n", serviceName))
		builder.WriteString(fmt.Sprintf("    image: %s\n", dependency.Image))
		builder.WriteString(fmt.Sprintf("    container_name: predeploy-%s\n", serviceName))

		if dependency.Port > 0 {
			builder.WriteString("    expose:\n")
			builder.WriteString(fmt.Sprintf("      - \"%d\"\n", dependency.Port))
		}

		writeEnvBlock(builder, dependency.Env)
	}
}

func writeEnvBlock(builder *bytes.Buffer, env map[string]string) {
	if len(env) == 0 {
		return
	}

	builder.WriteString("    environment:\n")

	keys := make([]string, 0, len(env))
	for key := range env {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		builder.WriteString(fmt.Sprintf("      %s: %q\n", key, env[key]))
	}
}

func sortedDependencyNames(dependencies map[string]config.DependencyConfig) []string {
	names := make([]string, 0, len(dependencies))

	for name := range dependencies {
		names = append(names, name)
	}

	sort.Strings(names)
	return names
}
