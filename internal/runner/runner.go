package runner

import (
	"context"
	"fmt"
	"time"

	"github.com/HXong/predeploy-guard/internal/builder"
	"github.com/HXong/predeploy-guard/internal/checker"
	"github.com/HXong/predeploy-guard/internal/config"
	"github.com/HXong/predeploy-guard/internal/loadtest"
	"github.com/HXong/predeploy-guard/internal/report"
	runtimefactory "github.com/HXong/predeploy-guard/internal/runtime/factory"
)

func Run(cfg *config.Config) error {
	ctx := context.Background()
	runtimeAdapter, err := runtimefactory.NewAdapter(cfg.Runtime.Type)
	if err != nil {
		return err
	}
	runtimeName := string(runtimeAdapter.Type())

	startedAt := time.Now()

	fmt.Println("Checking whether target image build is required...")

	buildResult := builder.BuildImageIfNeeded(cfg)

	if buildResult.Enabled {
		if buildResult.Passed {
			fmt.Printf("Built target image successfully: %s\n", buildResult.Image)
		} else {
			fmt.Printf("Target image build failed: %s\n", buildResult.Error)

			finishedAt := time.Now()

			reportData := report.ReportData{
				ServiceName: cfg.Service.Name,
				Image:       cfg.Service.Image,
				Runtime:     runtimeName,
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
			}

			paths, reportErr := writeReports(cfg, reportData)
			if reportErr != nil {
				return reportErr
			}

			fmt.Println("Result: FAIL")
			printReportPaths(paths)

			return fmt.Errorf("%s", buildResult.Error)
		}
	} else {
		fmt.Println("No target image build configured")
	}

	fmt.Printf("Preparing %s runtime sandbox...\n", runtimeName)

	env, err := runtimeAdapter.Prepare(ctx, cfg)
	if err != nil {
		return err
	}

	fmt.Printf("Sandbox directory: %s\n", env.WorkDir)

	startAttempted := false
	startFailed := false
	if cfg.Settings.Cleanup {
		defer func() {
			if startAttempted {
				if startFailed {
					fmt.Printf("Stopping %s runtime sandbox after start failure...\n", runtimeName)
				} else {
					fmt.Printf("Stopping %s runtime sandbox...\n", runtimeName)
				}
			}

			cleanupErr := runtimeAdapter.Cleanup(ctx, env, cfg, func(message string) {
				fmt.Println(message)
			})

			if cleanupErr != nil {
				fmt.Printf("Failed to clean up sandbox: %v\n", cleanupErr)
			}
		}()
	}

	fmt.Printf("Starting service with %s runtime...\n", runtimeName)

	startAttempted = true
	if err := runtimeAdapter.Start(ctx, env, cfg); err != nil {
		startFailed = true
		return err
	}

	fmt.Printf("Service base URL: %s\n", env.BaseURL)

	readinessResults := make([]report.ReadinessResult, 0)
	fmt.Printf("Waiting for %s runtime readiness...\n", runtimeName)

	runtimeReadinessResults, runtimeReadinessErr := runtimeAdapter.WaitReady(ctx, env, cfg)
	for _, result := range runtimeReadinessResults {
		readinessResults = append(readinessResults, report.ReadinessResult{
			Name:   result.Name,
			Target: result.Target,
			Passed: result.Passed,
			Error:  result.Error,
		})
	}

	var results []checker.SmokeResult
	passed := false
	readinessErr := error(nil)

	serviceReadiness := report.ReadinessResult{
		Name:   "service readiness",
		Target: joinURL(env.BaseURL, cfg.Service.HealthPath),
	}

	performanceResult := loadtest.K6Result{
		Enabled: cfg.Performance.Enabled,
		Passed:  !cfg.Performance.Enabled,
	}

	if runtimeReadinessErr != nil {
		fmt.Printf("Runtime readiness failed: %v\n", runtimeReadinessErr)
		serviceReadiness.Passed = false
		serviceReadiness.Error = "service readiness skipped because runtime readiness failed"
		readinessResults = append(readinessResults, serviceReadiness)
	} else {
		fmt.Println("Waiting for service readiness...")

		readinessErr = checker.WaitUntilReady(
			env.BaseURL,
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

			results = checker.RunSmokeChecks(env.BaseURL, cfg.Checks.Smoke)

			for _, result := range results {
				printSmokeResult(result)
			}

			smokePassed := len(results) > 0 && checker.AllPassed(results)

			if smokePassed && cfg.Performance.Enabled {
				fmt.Println("Running performance checks with k6...")

				performanceResult = loadtest.RunK6IfEnabled(cfg, env.WorkloadBaseURL, env.WorkDir)

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
	if runtimeReadinessErr != nil || readinessErr != nil || !passed {
		logs = collectDiagnosticsSafely(ctx, runtimeAdapter, env, cfg)
	}

	finishedAt := time.Now()

	reportData := report.ReportData{
		ServiceName:   cfg.Service.Name,
		ActiveProfile: cfg.ActiveProfile,
		Image:         cfg.Service.Image,
		Runtime:       runtimeName,
		BaseURL:       env.BaseURL,
		StartedAt:     startedAt,
		FinishedAt:    finishedAt,
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
			Enabled:  performanceResult.Enabled,
			Passed:   performanceResult.Passed,
			VUs:      performanceResult.VUs,
			Duration: performanceResult.Duration,

			AvgLatencyMs:    performanceResult.AvgLatencyMs,
			MinLatencyMs:    performanceResult.MinLatencyMs,
			MedianLatencyMs: performanceResult.MedianLatencyMs,
			MaxLatencyMs:    performanceResult.MaxLatencyMs,
			P90LatencyMs:    performanceResult.P90LatencyMs,
			P95LatencyMs:    performanceResult.P95LatencyMs,
			MaxP95LatencyMs: performanceResult.MaxP95LatencyMs,

			ErrorRate:    performanceResult.ErrorRate,
			MaxErrorRate: performanceResult.MaxErrorRate,

			RequestCount:  performanceResult.RequestCount,
			Iterations:    performanceResult.Iterations,
			ChecksTotal:   performanceResult.ChecksTotal,
			CheckPassRate: performanceResult.CheckPassRate,

			Error:  performanceResult.Error,
			Output: performanceResult.Output,
		},
		Passed: passed,
		Logs:   logs,
	}

	paths, reportErr := writeReports(cfg, reportData)
	if reportErr != nil {
		return reportErr
	}

	if runtimeReadinessErr != nil {
		fmt.Println("Result: FAIL")
		printReportPaths(paths)
		return runtimeReadinessErr
	}

	if readinessErr != nil {
		fmt.Println("Result: FAIL")
		printReportPaths(paths)
		return readinessErr
	}

	if !passed {
		fmt.Println("Result: FAIL")
		printReportPaths(paths)
		return fmt.Errorf("one or more validation checks failed")
	}

	fmt.Println("Result: PASS")
	printReportPaths(paths)

	return nil
}
