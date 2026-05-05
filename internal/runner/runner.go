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

	if err := checker.WaitUntilReady(
		sb.BaseURL(),
		cfg.Service.HealthPath,
		cfg.Settings.TimeoutSeconds,
	); err != nil {
		return err
	}

	fmt.Println("Running smoke checks...")

	results := checker.RunSmokeChecks(sb.BaseURL(), cfg.Checks.Smoke)

	for _, result := range results {
		status := "FAIL"
		if result.Passed {
			status = "PASS"
		}

		if result.Error != "" {
			fmt.Printf("[%s] %s %s %s error=%s duration=%s\n",
				status,
				result.Name,
				result.Method,
				result.URL,
				result.Error,
				result.Duration,
			)
			continue
		}

		fmt.Printf("[%s] %s %s %s expected=%d actual=%d duration=%s\n",
			status,
			result.Name,
			result.Method,
			result.URL,
			result.ExpectedStatus,
			result.ActualStatus,
			result.Duration,
		)
	}

	passed := checker.AllPassed(results)
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
	})

	if reportErr != nil {
		return reportErr
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
