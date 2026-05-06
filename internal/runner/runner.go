package runner

import (
	"fmt"
	"time"

	"github.com/HXong/predeploy-guard/internal/checker"
	"github.com/HXong/predeploy-guard/internal/config"
	"github.com/HXong/predeploy-guard/internal/report"
	"github.com/HXong/predeploy-guard/internal/sandbox"
)

func Run(cfg *config.Config) error {
	if cfg.Runtime.Type != "docker-compose" {
		return fmt.Errorf("unsupported runtime: %s", cfg.Runtime.Type)
	}

	startedAt := time.Now()

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
		Target: sb.BaseURL() + cfg.Service.HealthPath,
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

			passed = checker.AllPassed(results)
		}
	}

	logs := ""
	if dependencyErr != nil || readinessErr != nil || !passed {
		logs = collectLogsSafely(sb)
	}

	finishedAt := time.Now()

	reportPath, reportErr := report.WriteMarkdown(cfg, report.ReportData{
		ServiceName:      cfg.Service.Name,
		Image:            cfg.Service.Image,
		Runtime:          cfg.Runtime.Type,
		BaseURL:          sb.BaseURL(),
		StartedAt:        startedAt,
		FinishedAt:       finishedAt,
		ReadinessResults: readinessResults,
		Results:          results,
		Passed:           passed,
		Logs:             logs,
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
		return fmt.Errorf("one or more smoke checks failed")
	}

	fmt.Println("Result: PASS")
	fmt.Printf("Report written to: %s\n", reportPath)

	return nil
}
