package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/HXong/predeploy-guard/internal/config"
)

func printExplainSummary(cfg *config.Config) {
	fmt.Println("PreDeploy Guard Execution Plan")
	fmt.Println()

	printExplainProfile(cfg)
	printExplainRuntime(cfg)
	printExplainService(cfg)
	printExplainDependencies(cfg)
	printExplainReadiness(cfg)
	printExplainSmokeChecks(cfg)
	printExplainWorkloads(cfg)
	printExplainPerformance(cfg)
	printExplainReports(cfg)
	printExplainCleanup(cfg)
}

func printExplainProfile(cfg *config.Config) {
	fmt.Println("0. Profile")

	if cfg.ActiveProfile == "" {
		fmt.Println("   No profile was selected. The base configuration will be used.")
	} else {
		fmt.Printf("   Profile `%s` was selected and applied to the base configuration.\n", cfg.ActiveProfile)
	}

	if len(cfg.Profiles) > 0 {
		fmt.Printf("   Available profiles: %v\n", cfg.ProfileNames())
	}

	fmt.Println()
}

func printExplainRuntime(cfg *config.Config) {
	fmt.Println("1. Runtime")
	if cfg.ActiveProfile != "" {
		fmt.Printf("   Active profile: `%s`.\n", cfg.ActiveProfile)
	}

	switch cfg.Runtime.Type {
	case "docker-compose":
		fmt.Println("   PreDeploy Guard will use Docker Compose to create a temporary local sandbox.")
	case "kubernetes":
		if cfg.Runtime.Context == "" {
			fmt.Println("   PreDeploy Guard will use the current kubeconfig context to create a temporary Kubernetes namespace.")
		} else {
			fmt.Printf(
				"   PreDeploy Guard will use kubeconfig context `%s` to create a temporary Kubernetes namespace.\n",
				cfg.Runtime.Context,
			)
		}
	default:
		fmt.Printf("   PreDeploy Guard will use runtime: %s.\n", cfg.Runtime.Type)
	}

	fmt.Println()
}

func printExplainService(cfg *config.Config) {
	fmt.Println("2. Service")

	fmt.Printf("   Service `%s` will be tested.\n", cfg.Service.Name)
	fmt.Printf("   The service image is `%s`.\n", cfg.Service.Image)

	if cfg.Service.Build.Context != "" {
		fmt.Printf("   PreDeploy Guard will build the image from `%s` using `%s`.\n",
			cfg.Service.Build.Context,
			cfg.Service.Build.Dockerfile,
		)
	} else {
		fmt.Println("   No build context is configured, so PreDeploy Guard will use the configured image.")
	}

	fmt.Printf("   The service listens on container port `%d`.\n", cfg.Service.Port)

	if len(cfg.Service.Env) > 0 {
		fmt.Println("   The following environment variables will be passed to the service:")

		keys := sortedMapKeys(cfg.Service.Env)
		for _, key := range keys {
			fmt.Printf("   - %s=%s\n", key, maskIfSensitive(key, cfg.Service.Env[key]))
		}
	}

	fmt.Println()
}

func printExplainDependencies(cfg *config.Config) {
	fmt.Println("3. Dependencies")

	if len(cfg.Dependencies) == 0 {
		fmt.Println("   No dependency services are configured.")
		fmt.Println()
		return
	}

	names := make([]string, 0, len(cfg.Dependencies))
	for name := range cfg.Dependencies {
		names = append(names, name)
	}
	sort.Strings(names)

	fmt.Println("   The following dependency services will be started before validating the target service:")

	for _, name := range names {
		dependency := cfg.Dependencies[name]

		fmt.Printf("   - `%s` using image `%s`", name, dependency.Image)

		if dependency.Port > 0 {
			fmt.Printf(" on internal port `%d`", dependency.Port)
		}

		fmt.Println()

		if dependencyHasReadiness(dependency.Readiness) {
			fmt.Printf("     Readiness: %s\n", describeReadiness(dependency.Readiness))
		} else {
			fmt.Println("     Readiness: not configured; PreDeploy Guard waits for the runtime to report the dependency ready.")
		}
	}

	fmt.Println()
}

func printExplainReadiness(cfg *config.Config) {
	fmt.Println("4. Service Readiness")

	fmt.Printf("   PreDeploy Guard will wait for the service health endpoint `%s` to become reachable.\n",
		cfg.Service.HealthPath,
	)
	fmt.Printf("   Timeout: `%d` seconds.\n", cfg.Settings.TimeoutSeconds)
	fmt.Println()
}

func printExplainSmokeChecks(cfg *config.Config) {
	fmt.Println("5. Smoke Checks")

	if len(cfg.Checks.Smoke) == 0 {
		fmt.Println("   No smoke checks are configured.")
		fmt.Println()
		return
	}

	fmt.Printf("   `%d` smoke check(s) will run after service readiness passes:\n", len(cfg.Checks.Smoke))

	for _, check := range cfg.Checks.Smoke {
		fmt.Printf("   - `%s`: %s %s expects HTTP %d\n",
			check.Name,
			strings.ToUpper(check.Method),
			check.Path,
			check.ExpectedStatus,
		)
	}

	fmt.Println()
}

func printExplainPerformance(cfg *config.Config) {
	fmt.Println("7. Performance Checks")

	if !cfg.Performance.Enabled {
		fmt.Println("   Performance checks are disabled.")
		fmt.Println()
		return
	}

	fmt.Println("   Dockerized k6 will run after smoke checks and experiment workloads complete.")
	fmt.Printf("   Virtual users: `%d`\n", cfg.Performance.VUs)
	fmt.Printf("   Duration: `%s`\n", cfg.Performance.Duration)
	fmt.Printf("   Failure threshold: p95 latency must be <= `%.2fms`.\n",
		cfg.Performance.Thresholds.MaxP95LatencyMs,
	)
	fmt.Printf("   Failure threshold: error rate must be <= `%.4f`.\n",
		cfg.Performance.Thresholds.MaxErrorRate,
	)

	if len(cfg.Performance.Endpoints) > 0 {
		fmt.Println("   k6 endpoints:")

		for _, endpoint := range cfg.Performance.Endpoints {
			fmt.Printf("   - `%s`: %s %s\n",
				endpoint.Name,
				strings.ToUpper(endpoint.Method),
				endpoint.Path,
			)
		}
	}

	fmt.Println()
}

func printExplainWorkloads(cfg *config.Config) {
	fmt.Println("6. Experiment Workloads")

	if len(cfg.Workloads) == 0 {
		fmt.Println("   No experiment workloads are configured.")
		fmt.Println()
		return
	}

	fmt.Printf("   `%d` experiment workload(s) are configured to run after smoke checks:\n", len(cfg.Workloads))
	for _, workload := range cfg.Workloads {
		enabled := workload.Enabled == nil || *workload.Enabled
		state := "enabled"
		if !enabled {
			state = "disabled"
		}

		fmt.Printf(
			"   - `%s` (%s): %s %s at %d request(s)/second for %s; expects HTTP %d; failure policy `%s`\n",
			workload.Name,
			state,
			workload.Method,
			workload.Path,
			workload.RatePerSecond,
			workload.Duration,
			workload.ExpectedStatus,
			workload.FailurePolicy,
		)
	}

	fmt.Println()
}

func printExplainReports(cfg *config.Config) {
	fmt.Println("8. Reports")
	fmt.Printf("   Markdown and JSON reports will be written under `%s/reports`.\n", cfg.ConfigDir)
	fmt.Println("   Markdown is intended for humans.")
	fmt.Println("   JSON is intended for automation and CI/CD.")
	fmt.Println()
}

func printExplainCleanup(cfg *config.Config) {
	fmt.Println("9. Cleanup")

	if cfg.Settings.Cleanup {
		fmt.Printf(
			"   Cleanup is enabled. The temporary %s runtime sandbox will be removed after the run.\n",
			cfg.Runtime.Type,
		)
	} else {
		fmt.Printf(
			"   Cleanup is disabled. The temporary %s runtime sandbox will be kept after the run.\n",
			cfg.Runtime.Type,
		)
		fmt.Println("   This is useful for debugging but may leave runtime resources and files behind.")
	}

	fmt.Println()
}

func describeReadiness(readiness config.ReadinessConfig) string {
	if len(readiness.Command) > 0 {
		return fmt.Sprintf("command `%s`", strings.Join(readiness.Command, " "))
	}

	if readiness.Shell != "" {
		return fmt.Sprintf("shell `%s`", readiness.Shell)
	}

	return "not configured"
}

func sortedMapKeys(values map[string]string) []string {
	keys := make([]string, 0, len(values))

	for key := range values {
		keys = append(keys, key)
	}

	sort.Strings(keys)
	return keys
}

func maskIfSensitive(key string, value string) string {
	upperKey := strings.ToUpper(key)

	sensitiveWords := []string{
		"PASSWORD",
		"SECRET",
		"TOKEN",
		"KEY",
	}

	for _, word := range sensitiveWords {
		if strings.Contains(upperKey, word) {
			return "******"
		}
	}

	return value
}
