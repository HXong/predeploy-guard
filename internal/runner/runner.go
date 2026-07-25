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
	"github.com/HXong/predeploy-guard/internal/workload"
)

func Run(cfg *config.Config) error {
	ctx := context.Background()
	runtimeAdapter, err := runtimefactory.NewAdapter(cfg.Runtime.Type)
	if err != nil {
		return err
	}
	runtimeName := string(runtimeAdapter.Type())

	startedAt := time.Now()
	runPhases := make([]report.RunPhase, 0, 9)

	fmt.Println("Checking whether target image build is required...")

	buildPhase := report.StartPhase(report.PhaseImageBuild)
	buildResult := builder.BuildImageIfNeeded(cfg)
	switch {
	case !buildResult.Enabled:
		runPhases = append(runPhases, buildPhase.FinishSkipped(""))
	case buildResult.Passed:
		runPhases = append(runPhases, buildPhase.FinishPassed())
	default:
		runPhases = append(runPhases, buildPhase.FinishFailed(buildResult.Error))
	}

	writeFailedRun := func(
		envName string,
		baseURL string,
		workloadBaseURL string,
		logs string,
		runErr error,
	) error {
		if baseURL == "" {
			baseURL = "-"
		}

		finishedAt := time.Now()
		reportData := report.ReportData{
			ServiceName:   cfg.Service.Name,
			ActiveProfile: cfg.ActiveProfile,
			Image:         cfg.Service.Image,
			Runtime:       runtimeName,
			BaseURL:       baseURL,
			RuntimeEnvironment: report.RuntimeEnvironment{
				Runtime:         runtimeName,
				Name:            envName,
				BaseURL:         emptyIfPlaceholder(baseURL),
				WorkloadBaseURL: workloadBaseURL,
			},
			RunPhases:   runPhases,
			StartedAt:   startedAt,
			FinishedAt:  finishedAt,
			BuildResult: reportBuildResult(buildResult),
			Passed:      false,
			Logs:        logs,
		}

		paths, reportErr := writeReports(cfg, reportData)
		if reportErr != nil {
			return reportErr
		}

		fmt.Println("Result: FAIL")
		printReportPaths(paths)
		return runErr
	}

	if buildResult.Enabled {
		if buildResult.Passed {
			fmt.Printf("Built target image successfully: %s\n", buildResult.Image)
		} else {
			fmt.Printf("Target image build failed: %s\n", buildResult.Error)
			appendSkippedPhases(
				&runPhases,
				[]string{
					report.PhaseRuntimePrepare,
					report.PhaseRuntimeStart,
					report.PhaseRuntimeReadiness,
					report.PhaseServiceReadiness,
					report.PhaseSmokeChecks,
					report.PhaseExperimentWorkloads,
					report.PhasePerformanceChecks,
					report.PhaseDiagnostics,
				},
				"skipped because image build failed",
			)
			return writeFailedRun("", "", "", "", fmt.Errorf("%s", buildResult.Error))
		}
	} else {
		fmt.Println("No target image build configured")
	}

	fmt.Printf("Preparing %s runtime sandbox...\n", runtimeName)

	preparePhase := report.StartPhase(report.PhaseRuntimePrepare)
	env, err := runtimeAdapter.Prepare(ctx, cfg)
	if err != nil {
		runPhases = append(runPhases, preparePhase.FinishFailed(err.Error()))
		appendSkippedPhases(
			&runPhases,
			[]string{
				report.PhaseRuntimeStart,
				report.PhaseRuntimeReadiness,
				report.PhaseServiceReadiness,
				report.PhaseSmokeChecks,
				report.PhaseExperimentWorkloads,
				report.PhasePerformanceChecks,
				report.PhaseDiagnostics,
			},
			"skipped because runtime preparation failed",
		)
		return writeFailedRun("", "", "", "", err)
	}
	runPhases = append(runPhases, preparePhase.FinishPassed())

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
	startPhase := report.StartPhase(report.PhaseRuntimeStart)
	if err := runtimeAdapter.Start(ctx, env, cfg); err != nil {
		startFailed = true
		runPhases = append(runPhases, startPhase.FinishFailed(err.Error()))
		appendSkippedPhases(
			&runPhases,
			[]string{
				report.PhaseRuntimeReadiness,
				report.PhaseServiceReadiness,
				report.PhaseSmokeChecks,
				report.PhaseExperimentWorkloads,
				report.PhasePerformanceChecks,
			},
			"skipped because runtime start failed",
		)

		diagnosticsPhase := report.StartPhase(report.PhaseDiagnostics)
		logs, diagnosticsErr := collectDiagnosticsSafely(ctx, runtimeAdapter, env, cfg)
		if diagnosticsErr != nil {
			runPhases = append(runPhases, diagnosticsPhase.FinishFailed(diagnosticsErr.Error()))
		} else {
			runPhases = append(runPhases, diagnosticsPhase.FinishPassed())
		}

		return writeFailedRun(env.Name, env.BaseURL, env.WorkloadBaseURL, logs, err)
	}
	runPhases = append(runPhases, startPhase.FinishPassed())

	fmt.Printf("Service base URL: %s\n", env.BaseURL)

	readinessResults := make([]report.ReadinessResult, 0)
	fmt.Printf("Waiting for %s runtime readiness...\n", runtimeName)

	runtimeReadinessPhase := report.StartPhase(report.PhaseRuntimeReadiness)
	runtimeReadinessResults, runtimeReadinessErr := runtimeAdapter.WaitReady(ctx, env, cfg)
	if runtimeReadinessErr != nil {
		runPhases = append(runPhases, runtimeReadinessPhase.FinishFailed(runtimeReadinessErr.Error()))
	} else {
		runPhases = append(runPhases, runtimeReadinessPhase.FinishPassed())
	}
	for _, result := range runtimeReadinessResults {
		readinessResults = append(readinessResults, report.ReadinessResult{
			Name:   result.Name,
			Target: result.Target,
			Passed: result.Passed,
			Error:  result.Error,
		})
	}

	var results []checker.SmokeResult
	var workloadResults []workload.Result
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
		appendSkippedPhases(
			&runPhases,
			[]string{
				report.PhaseServiceReadiness,
				report.PhaseSmokeChecks,
				report.PhaseExperimentWorkloads,
				report.PhasePerformanceChecks,
			},
			"skipped because runtime readiness failed",
		)
	} else {
		fmt.Println("Waiting for service readiness...")

		serviceReadinessPhase := report.StartPhase(report.PhaseServiceReadiness)
		readinessErr = checker.WaitUntilReady(
			env.BaseURL,
			cfg.Service.HealthPath,
			cfg.Settings.TimeoutSeconds,
		)

		if readinessErr != nil {
			runPhases = append(runPhases, serviceReadinessPhase.FinishFailed(readinessErr.Error()))
			fmt.Printf("Readiness check failed: %v\n", readinessErr)
			serviceReadiness.Passed = false
			serviceReadiness.Error = readinessErr.Error()
			readinessResults = append(readinessResults, serviceReadiness)
			appendSkippedPhases(
				&runPhases,
				[]string{
					report.PhaseSmokeChecks,
					report.PhaseExperimentWorkloads,
					report.PhasePerformanceChecks,
				},
				"skipped because service readiness failed",
			)
		} else {
			runPhases = append(runPhases, serviceReadinessPhase.FinishPassed())
			serviceReadiness.Passed = true
			readinessResults = append(readinessResults, serviceReadiness)

			fmt.Println("Running smoke checks...")

			smokePhase := report.StartPhase(report.PhaseSmokeChecks)
			results = checker.RunSmokeChecks(env.BaseURL, cfg.Checks.Smoke)

			for _, result := range results {
				printSmokeResult(result)
			}

			smokePassed := len(results) > 0 && checker.AllPassed(results)
			workloadsPassed := true
			if smokePassed {
				runPhases = append(runPhases, smokePhase.FinishPassed())
			} else {
				smokeError := "one or more smoke checks failed"
				if len(results) == 0 {
					smokeError = "no smoke checks were executed"
				}
				runPhases = append(runPhases, smokePhase.FinishFailed(smokeError))
			}

			if smokePassed {
				switch {
				case len(cfg.Workloads) == 0:
					runPhases = append(
						runPhases,
						report.SkippedPhase(report.PhaseExperimentWorkloads, ""),
					)
				case !workload.HasEnabled(cfg):
					workloadResults = workload.RunAll(ctx, cfg, env.BaseURL)
					runPhases = append(
						runPhases,
						report.SkippedPhase(report.PhaseExperimentWorkloads, ""),
					)
				default:
					fmt.Println("Running experiment workloads...")
					workloadPhase := report.StartPhase(report.PhaseExperimentWorkloads)
					workloadResults = workload.RunAll(ctx, cfg, env.BaseURL)
					for _, result := range workloadResults {
						printWorkloadResult(result)
					}
					workloadsPassed = !workload.ShouldFailRun(workloadResults)
					if workloadsPassed {
						runPhases = append(runPhases, workloadPhase.FinishPassed())
					} else {
						runPhases = append(
							runPhases,
							workloadPhase.FinishFailed("one or more required experiment workloads failed"),
						)
					}
				}
			} else {
				runPhases = append(
					runPhases,
					report.SkippedPhase(
						report.PhaseExperimentWorkloads,
						"skipped because smoke checks failed",
					),
				)
			}

			if smokePassed && cfg.Performance.Enabled {
				fmt.Println("Running performance checks with k6...")

				performancePhase := report.StartPhase(report.PhasePerformanceChecks)
				performanceResult = loadtest.RunK6IfEnabled(cfg, env.WorkloadBaseURL, env.WorkDir)

				if performanceResult.Enabled {
					if performanceResult.Passed {
						fmt.Println("Performance check passed")
						runPhases = append(runPhases, performancePhase.FinishPassed())
					} else {
						fmt.Printf("Performance check failed: %s\n", performanceResult.Error)
						errorMessage := performanceResult.Error
						if errorMessage == "" {
							errorMessage = "performance check failed"
						}
						runPhases = append(runPhases, performancePhase.FinishFailed(errorMessage))
					}
				}
			} else {
				reason := ""
				if !smokePassed {
					reason = "skipped because smoke checks failed"
				}
				runPhases = append(
					runPhases,
					report.SkippedPhase(report.PhasePerformanceChecks, reason),
				)
			}

			passed = smokePassed && workloadsPassed && performanceResult.Passed
		}
	}

	logs := ""
	if runtimeReadinessErr != nil || readinessErr != nil || !passed {
		diagnosticsPhase := report.StartPhase(report.PhaseDiagnostics)
		var diagnosticsErr error
		logs, diagnosticsErr = collectDiagnosticsSafely(ctx, runtimeAdapter, env, cfg)
		if diagnosticsErr != nil {
			runPhases = append(runPhases, diagnosticsPhase.FinishFailed(diagnosticsErr.Error()))
		} else {
			runPhases = append(runPhases, diagnosticsPhase.FinishPassed())
		}
	} else {
		runPhases = append(runPhases, report.SkippedPhase(report.PhaseDiagnostics, ""))
	}

	finishedAt := time.Now()

	reportData := report.ReportData{
		ServiceName:        cfg.Service.Name,
		ActiveProfile:      cfg.ActiveProfile,
		Image:              cfg.Service.Image,
		Runtime:            runtimeName,
		BaseURL:            env.BaseURL,
		RuntimeEnvironment: runtimeEnvironmentSummary(runtimeName, env),
		RunPhases:          runPhases,
		StartedAt:          startedAt,
		FinishedAt:         finishedAt,
		BuildResult:        reportBuildResult(buildResult),
		ReadinessResults:   readinessResults,
		Results:            results,
		WorkloadResults:    workloadResults,
		PerformanceResult:  reportPerformanceResult(performanceResult),
		Passed:             passed,
		Logs:               logs,
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

func emptyIfPlaceholder(value string) string {
	if value == "-" {
		return ""
	}
	return value
}
