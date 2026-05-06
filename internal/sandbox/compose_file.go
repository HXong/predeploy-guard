package sandbox

import (
	"bytes"
	"fmt"
	"sort"

	"github.com/HXong/predeploy-guard/internal/config"
)

func (s *ComposeSandbox) writeComposeFile(cfg *config.Config) error {
	var builder bytes.Buffer

	fmt.Fprintf(&builder, "services:\n")

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
	fmt.Fprintf(builder, "  %s:\n", s.ServiceName)
	fmt.Fprintf(builder, "    image: %s\n", cfg.Service.Image)
	fmt.Fprintf(builder, "    pull_policy: never\n")
	fmt.Fprintf(builder, "    container_name: predeploy-%s\n", s.ServiceName)
	fmt.Fprintf(builder, "    ports:\n")
	fmt.Fprintf(builder, "      - \"%d:%d\"\n", s.HostPort, cfg.Service.Port)

	if len(cfg.Dependencies) > 0 {
		fmt.Fprintf(builder, "    depends_on:\n")

		for _, name := range sortedDependencyNames(cfg.Dependencies) {
			serviceName := sanitizeServiceName(name)

			fmt.Fprintf(builder, "      %s:\n", serviceName)
			fmt.Fprintf(builder, "        condition: service_started\n")
		}
	}

	writeEnvBlock(builder, cfg.Service.Env)
}

func writeDependencyServices(builder *bytes.Buffer, cfg *config.Config) {
	for _, name := range sortedDependencyNames(cfg.Dependencies) {
		dependency := cfg.Dependencies[name]
		serviceName := sanitizeServiceName(name)

		fmt.Fprintf(builder, "  %s:\n", serviceName)
		fmt.Fprintf(builder, "    image: %s\n", dependency.Image)
		fmt.Fprintf(builder, "    container_name: predeploy-%s\n", serviceName)

		if dependency.Port > 0 {
			fmt.Fprintf(builder, "    expose:\n")
			fmt.Fprintf(builder, "      - \"%d\"\n", dependency.Port)
		}

		writeEnvBlock(builder, dependency.Env)
		writeHealthcheckBlock(builder, dependency.Readiness)
	}
}

func writeEnvBlock(builder *bytes.Buffer, env map[string]string) {
	if len(env) == 0 {
		return
	}

	fmt.Fprintf(builder, "    environment:\n")

	keys := make([]string, 0, len(env))
	for key := range env {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		fmt.Fprintf(builder, "      %s: %q\n", key, env[key])
	}
}

func writeHealthcheckBlock(builder *bytes.Buffer, readiness config.ReadinessConfig) {
	hasCommand := len(readiness.Command) > 0
	hasShell := readiness.Shell != ""

	if !hasCommand && !hasShell {
		return
	}

	fmt.Fprintf(builder, "    healthcheck:\n")

	if hasShell {
		fmt.Fprintf(builder, "      test: [\"CMD-SHELL\", %q]\n", readiness.Shell)
	} else {
		fmt.Fprintf(builder, "      test: [\"CMD\"")

		for _, part := range readiness.Command {
			fmt.Fprintf(builder, ", %q", part)
		}

		fmt.Fprintf(builder, "]\n")
	}

	fmt.Fprintf(builder, "      interval: %ds\n", readiness.IntervalSeconds)
	fmt.Fprintf(builder, "      timeout: 3s\n")

	retries := readiness.TimeoutSeconds / readiness.IntervalSeconds
	if retries <= 0 {
		retries = 1
	}

	fmt.Fprintf(builder, "      retries: %d\n", retries)
	fmt.Fprintf(builder, "      start_period: 3s\n")
}

func sortedDependencyNames(dependencies map[string]config.DependencyConfig) []string {
	names := make([]string, 0, len(dependencies))

	for name := range dependencies {
		names = append(names, name)
	}

	sort.Strings(names)
	return names
}
