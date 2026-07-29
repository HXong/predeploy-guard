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
	if cfg.Runtime.Context != "" {
		fmt.Printf("  Context: %s\n", cfg.Runtime.Context)
	}
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
	fmt.Println("Gateway")
	fmt.Printf("  Enabled: %t\n", cfg.Gateway.Enabled)
	if cfg.Gateway.Enabled {
		fmt.Printf("  Base URL: %s\n", cfg.Gateway.BaseURL)
		ingressState := "disabled"
		if cfg.Gateway.Ingress.Enabled {
			ingressState = "enabled"
		}
		fmt.Printf("  Ingress: %s\n", ingressState)
		if cfg.Gateway.Ingress.Enabled {
			if cfg.Gateway.Ingress.Host != "" {
				fmt.Printf("  Ingress host: %s\n", cfg.Gateway.Ingress.Host)
			}
			if cfg.Gateway.Ingress.ClassName != "" {
				fmt.Printf("  Ingress class: %s\n", cfg.Gateway.Ingress.ClassName)
			}
			fmt.Printf("  Ingress path type: %s\n", cfg.Gateway.Ingress.PathType)
		}
		latencyState := "disabled"
		if cfg.Gateway.Latency.Enabled {
			latencyState = "enabled"
		}
		fmt.Printf("  Latency comparison: %s\n", latencyState)
		if cfg.Gateway.Latency.Enabled {
			fmt.Printf("  Latency failure policy: %s\n", cfg.Gateway.Latency.FailurePolicy)
			if cfg.Gateway.Latency.MaxGatewayLatencyMs > 0 {
				fmt.Printf(
					"  Max gateway latency: %dms\n",
					cfg.Gateway.Latency.MaxGatewayLatencyMs,
				)
			}
			if cfg.Gateway.Latency.MaxOverheadMs > 0 {
				fmt.Printf("  Max overhead: %dms\n", cfg.Gateway.Latency.MaxOverheadMs)
			}
			if cfg.Gateway.Latency.MaxOverheadRatio > 0 {
				fmt.Printf(
					"  Max overhead ratio: %.2fx\n",
					cfg.Gateway.Latency.MaxOverheadRatio,
				)
			}
		}
		fmt.Printf("  Routes: %d\n", len(cfg.Gateway.Routes))
		for _, route := range cfg.Gateway.Routes {
			compareDirect := route.CompareDirect == nil || *route.CompareDirect
			fmt.Printf(
				"    - %s: %s %s expect=%d compare-direct=%t\n",
				route.Name,
				route.Method,
				route.Path,
				route.ExpectedStatus,
				compareDirect,
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
	fmt.Println("Experiment Workloads")
	if len(cfg.Workloads) == 0 {
		fmt.Println("  None")
	} else {
		for _, workload := range cfg.Workloads {
			enabled := workload.Enabled == nil || *workload.Enabled
			fmt.Printf(
				"  - %s: enabled=%t type=%s target=%s %s rate=%d/s duration=%s expect=%d policy=%s\n",
				workload.Name,
				enabled,
				workload.Type,
				workload.Method,
				workload.Path,
				workload.RatePerSecond,
				workload.Duration,
				workload.ExpectedStatus,
				workload.FailurePolicy,
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
