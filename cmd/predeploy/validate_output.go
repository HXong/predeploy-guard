package main

import (
	"fmt"
	"sort"

	"github.com/HXong/predeploy-guard/internal/config"
)

func printValidationSummary(cfg *config.Config) {
	fmt.Println("Config validation passed")
	fmt.Println()

	fmt.Println("Config")
	fmt.Printf("  File: %s\n", cfg.ConfigPath)
	fmt.Printf("  Directory: %s\n", cfg.ConfigDir)
	if cfg.ActiveProfile != "" {
		fmt.Printf("  Active profile: %s\n", cfg.ActiveProfile)
	}

	if len(cfg.Profiles) > 0 {
		fmt.Printf("  Available profiles: %v\n", cfg.ProfileNames())
	}
	fmt.Println()

	fmt.Println("Runtime")
	fmt.Printf("  Type: %s\n", cfg.Runtime.Type)
	fmt.Println()

	fmt.Println("Service")
	fmt.Printf("  Name: %s\n", cfg.Service.Name)
	fmt.Printf("  Image: %s\n", cfg.Service.Image)
	fmt.Printf("  Port: %d\n", cfg.Service.Port)
	fmt.Printf("  Health path: %s\n", cfg.Service.HealthPath)

	if cfg.Service.Build.Context != "" {
		fmt.Println()
		fmt.Println("Build")
		fmt.Printf("  Context: %s\n", cfg.Service.Build.Context)
		fmt.Printf("  Dockerfile: %s\n", cfg.Service.Build.Dockerfile)
	} else {
		fmt.Println()
		fmt.Println("Build")
		fmt.Println("  No build context configured")
	}

	fmt.Println()
	fmt.Println("Dependencies")
	if len(cfg.Dependencies) == 0 {
		fmt.Println("  None")
	} else {
		names := make([]string, 0, len(cfg.Dependencies))
		for name := range cfg.Dependencies {
			names = append(names, name)
		}
		sort.Strings(names)

		for _, name := range names {
			dependency := cfg.Dependencies[name]
			fmt.Printf("  - %s: image=%s port=%d readiness=%t\n",
				name,
				dependency.Image,
				dependency.Port,
				dependencyHasReadiness(dependency.Readiness),
			)
		}
	}

	fmt.Println()
	fmt.Println("Smoke Checks")
	if len(cfg.Checks.Smoke) == 0 {
		fmt.Println("  None")
	} else {
		for _, check := range cfg.Checks.Smoke {
			fmt.Printf("  - %s: %s %s expect=%d\n",
				check.Name,
				check.Method,
				check.Path,
				check.ExpectedStatus,
			)
		}
	}

	fmt.Println()
	fmt.Println("Performance")
	fmt.Printf("  Enabled: %t\n", cfg.Performance.Enabled)

	if cfg.Performance.Enabled {
		fmt.Printf("  VUs: %d\n", cfg.Performance.VUs)
		fmt.Printf("  Duration: %s\n", cfg.Performance.Duration)
		fmt.Printf("  Max p95 latency: %.2fms\n", cfg.Performance.Thresholds.MaxP95LatencyMs)
		fmt.Printf("  Max error rate: %.4f\n", cfg.Performance.Thresholds.MaxErrorRate)
		fmt.Printf("  Endpoints: %d\n", len(cfg.Performance.Endpoints))

		for _, endpoint := range cfg.Performance.Endpoints {
			fmt.Printf("    - %s: %s %s\n",
				endpoint.Name,
				endpoint.Method,
				endpoint.Path,
			)
		}
	}

	fmt.Println()
	fmt.Println("Settings")
	fmt.Printf("  Cleanup: %t\n", cfg.Settings.Cleanup)
	fmt.Printf("  Timeout seconds: %d\n", cfg.Settings.TimeoutSeconds)
}

func dependencyHasReadiness(readiness config.ReadinessConfig) bool {
	return len(readiness.Command) > 0 || readiness.Shell != ""
}
