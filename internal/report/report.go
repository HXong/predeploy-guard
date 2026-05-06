package report

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/HXong/predeploy-guard/internal/checker"
	"github.com/HXong/predeploy-guard/internal/config"
)

type ReportData struct {
	ServiceName      string
	Image            string
	Runtime          string
	BaseURL          string
	StartedAt        time.Time
	FinishedAt       time.Time
	ReadinessResults []ReadinessResult
	Results          []checker.SmokeResult
	Passed           bool
	Logs             string
}

type ReadinessResult struct {
	Name   string
	Target string
	Passed bool
	Error  string
}

func WriteMarkdown(cfg *config.Config, data ReportData) (string, error) {
	if err := os.MkdirAll("reports", 0755); err != nil {
		return "", fmt.Errorf("create reports directory: %w", err)
	}

	filename := buildReportFilename(cfg.Service.Name, data.StartedAt)
	reportPath := filepath.Join("reports", filename)

	content := buildMarkdown(cfg, data)

	if err := os.WriteFile(reportPath, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("write markdown report: %w", err)
	}

	return reportPath, nil
}
