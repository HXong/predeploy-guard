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

	fmt.Println("Waiting for service readiness...")

	readinessErr := checker.WaitUntilReady(
		sb.BaseURL(),
		cfg.Service.HealthPath,
		cfg.Settings.TimeoutSeconds,
	)

	var results []checker.SmokeResult
	passed := false

	if readinessErr != nil {
		fmt.Printf("Readiness check failed: %v\n", readinessErr)
	} else {
		fmt.Println("Running smoke checks...")

		results = checker.RunSmokeChecks(sb.BaseURL(), cfg.Checks.Smoke)

		for _, result := range results {
			printSmokeResult(result)
		}

		passed = checker.AllPassed(results)
	}

	logs := ""
	if readinessErr != nil || !passed {
		fmt.Println("Collecting container logs...")

		collectedLogs, logErr := sb.Logs()
		if logErr != nil {
			logs = fmt.Sprintf("Failed to collect logs: %v\n\nPartial output:\n%s", logErr, collectedLogs)
		} else {
			logs = collectedLogs
		}
	}

	finishedAt := time.Now()

	reportPath, reportErr := report.WriteMarkdown(cfg, report.ReportData{
		ServiceName: cfg.Service.Name,
		Image:       cfg.Service.Image,
		Runtime:     cfg.Runtime.Type,
		BaseURL:     sb.BaseURL(),
		StartedAt:   startedAt,
		FinishedAt:  finishedAt,
		Results:     results,
		Passed:      passed,
		Logs:        logs,
	})

	if reportErr != nil {
		return reportErr
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
