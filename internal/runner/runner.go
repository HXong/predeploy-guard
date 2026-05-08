package runner

import (
	"fmt"
	"time"

	"github.com/HXong/predeploy-guard/internal/builder"
	"github.com/HXong/predeploy-guard/internal/checker"
	"github.com/HXong/predeploy-guard/internal/config"
	"github.com/HXong/predeploy-guard/internal/loadtest"
	"github.com/HXong/predeploy-guard/internal/report"
	"github.com/HXong/predeploy-guard/internal/sandbox"
)

func Run(cfg *config.Config) error {
	if cfg.Runtime.Type != "docker-compose" {
		return fmt.Errorf("unsupported runtime: %s", cfg.Runtime.Type)
	}

	startedAt := time.Now()

	fmt.Println("Checking whether target image build is required...")

	buildResult := builder.BuildImageIfNeeded(cfg)

	if buildResult.Enabled {
		if buildResult.Passed {
			fmt.Printf("Built target image successfully: %s\n", buildResult.Image)
		} else {
			fmt.Printf("Target image build failed: %s\n", buildResult.Error)

			finishedAt := time.Now()

			reportPath, reportErr := report.WriteMarkdown(cfg, report.ReportData{
				ServiceName: cfg.Service.Name,
				Image:       cfg.Service.Image,
				Runtime:     cfg.Runtime.Type,
				BaseURL:     "-",
				StartedAt:   startedAt,
				FinishedAt:  finishedAt,
				BuildResult: report.BuildResult{
					Enabled:    buildResult.Enabled,
					Image:      buildResult.Image,
					Context:    buildResult.Context,
					Dockerfile: buildResult.Dockerfile,
					Passed:     buildResult.Passed,
					Error:      buildResult.Error,
					Output:     buildResult.Output,
				},
				Passed: false,
			})

			if reportErr != nil {
				return reportErr
			}

			fmt.Println("Result: FAIL")
			fmt.Printf("Report written to: %s\n", reportPath)

			return fmt.Errorf("%s", buildResult.Error)
		}
	} else {
		fmt.Println("No target image build configured")
	}

	fmt.Println("Creating Docker Compose sandbox...")

	sb, err := sandbox.NewComposeSandbox(cfg)
	if err != nil {
		return err
	}

	fmt.Printf("Sandbox directory: %s\n", sb.WorkDir)

	if cfg.Settings.Cleanup {
		defer func() {
			fmt.Println("Removing sandbox files...")

			if err := sb.RemoveFiles(); err != nil {
				fmt.Printf("Failed to remove sandbox files: %v\n", err)
			}
		}()
	}

	fmt.Println("Starting service with Docker Compose...")

	if err := sb.Start(); err != nil {
		if cfg.Settings.Cleanup {
			fmt.Println("Stopping Docker Compose sandbox after start failure...")

			if stopErr := sb.Stop(); stopErr != nil {
				fmt.Printf("Failed to stop sandbox after start failure: %v\n", stopErr)
			}
		}

		return err
	}

	if cfg.Settings.Cleanup {
		defer func() {
			fmt.Println("Stopping Docker Compose sandbox...")

			if err := sb.Stop(); err != nil {
				fmt.Printf("Failed to stop sandbox: %v\n", err)
			}
		}()
	}

	fmt.Printf("Service base URL: %s\n", sb.BaseURL())

	readinessResults := make([]report.ReadinessResult, 0)
	dependencyErr := error(nil)

	if len(cfg.Dependencies) > 0 {
		fmt.Println("Waiting for dependency readiness...")

		dependencyResults, err := sb.WaitForDependencies(cfg)
		dependencyErr = err

		for _, result := range dependencyResults {
			statusName := "dependency readiness"
			if result.Skipped {
				statusName = "dependency readiness (skipped)"
			}

			readinessResults = append(readinessResults, report.ReadinessResult{
				Name:   statusName,
				Target: result.Name,
				Passed: result.Passed,
				Error:  result.Error,
			})
		}
	}

	var results []checker.SmokeResult
	passed := false
	readinessErr := error(nil)

	serviceReadiness := report.ReadinessResult{
		Name:   "service readiness",
		Target: joinURL(sb.BaseURL(), cfg.Service.HealthPath),
	}

	performanceResult := loadtest.K6Result{
		Enabled: cfg.Performance.Enabled,
		Passed:  !cfg.Performance.Enabled,
	}

	if dependencyErr != nil {
		fmt.Printf("Dependency readiness failed: %v\n", dependencyErr)
		serviceReadiness.Passed = false
		serviceReadiness.Error = "service readiness skipped because dependency readiness failed"
		readinessResults = append(readinessResults, serviceReadiness)
	} else {
		fmt.Println("Waiting for service readiness...")

		readinessErr = checker.WaitUntilReady(
			sb.BaseURL(),
			cfg.Service.HealthPath,
			cfg.Settings.TimeoutSeconds,
		)

		if readinessErr != nil {
			fmt.Printf("Readiness check failed: %v\n", readinessErr)
			serviceReadiness.Passed = false
			serviceReadiness.Error = readinessErr.Error()
			readinessResults = append(readinessResults, serviceReadiness)
		} else {
			serviceReadiness.Passed = true
			readinessResults = append(readinessResults, serviceReadiness)

			fmt.Println("Running smoke checks...")

			results = checker.RunSmokeChecks(sb.BaseURL(), cfg.Checks.Smoke)

			for _, result := range results {
				printSmokeResult(result)
			}

			smokePassed := len(results) > 0 && checker.AllPassed(results)

			if smokePassed && cfg.Performance.Enabled {
				fmt.Println("Running performance checks with k6...")

				performanceResult = loadtest.RunK6IfEnabled(cfg, sb.K6BaseURL(), sb.WorkDir)

				if performanceResult.Enabled {
					if performanceResult.Passed {
						fmt.Println("Performance check passed")
					} else {
						fmt.Printf("Performance check failed: %s\n", performanceResult.Error)
					}
				}
			}

			passed = smokePassed && performanceResult.Passed
		}
	}

	logs := ""
	if dependencyErr != nil || readinessErr != nil || !passed {
		logs = collectLogsSafely(sb)
	}

	finishedAt := time.Now()

	reportPath, reportErr := report.WriteMarkdown(cfg, report.ReportData{
		ServiceName: cfg.Service.Name,
		Image:       cfg.Service.Image,
		Runtime:     cfg.Runtime.Type,
		BaseURL:     sb.BaseURL(),
		StartedAt:   startedAt,
		FinishedAt:  finishedAt,
		BuildResult: report.BuildResult{
			Enabled:    buildResult.Enabled,
			Image:      buildResult.Image,
			Context:    buildResult.Context,
			Dockerfile: buildResult.Dockerfile,
			Passed:     buildResult.Passed,
			Error:      buildResult.Error,
			Output:     buildResult.Output,
		},
		ReadinessResults: readinessResults,
		Results:          results,
		PerformanceResult: report.PerformanceResult{
			Enabled:         performanceResult.Enabled,
			Passed:          performanceResult.Passed,
			VUs:             performanceResult.VUs,
			Duration:        performanceResult.Duration,
			P95LatencyMs:    performanceResult.P95LatencyMs,
			MaxP95LatencyMs: performanceResult.MaxP95LatencyMs,
			ErrorRate:       performanceResult.ErrorRate,
			MaxErrorRate:    performanceResult.MaxErrorRate,
			Error:           performanceResult.Error,
			Output:          performanceResult.Output,
		},
		Passed: passed,
		Logs:   logs,
	})

	if reportErr != nil {
		return reportErr
	}

	if dependencyErr != nil {
		fmt.Println("Result: FAIL")
		fmt.Printf("Report written to: %s\n", reportPath)
		return dependencyErr
	}

	if readinessErr != nil {
		fmt.Println("Result: FAIL")
		fmt.Printf("Report written to: %s\n", reportPath)
		return readinessErr
	}

	if !passed {
		fmt.Println("Result: FAIL")
		fmt.Printf("Report written to: %s\n", reportPath)
		return fmt.Errorf("one or more validation checks failed")
	}

	fmt.Println("Result: PASS")
	fmt.Printf("Report written to: %s\n", reportPath)

	return nil
}
