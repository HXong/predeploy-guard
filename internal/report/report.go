package report

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/HXong/predeploy-guard/internal/checker"
	"github.com/HXong/predeploy-guard/internal/config"
)

type ReportData struct {
	ServiceName       string
	Image             string
	Runtime           string
	BaseURL           string
	StartedAt         time.Time
	FinishedAt        time.Time
	BuildResult       BuildResult
	ReadinessResults  []ReadinessResult
	Results           []checker.SmokeResult
	PerformanceResult PerformanceResult
	Passed            bool
	Logs              string
}

type BuildResult struct {
	Enabled    bool
	Image      string
	Context    string
	Dockerfile string
	Passed     bool
	Error      string
	Output     string
}

type ReadinessResult struct {
	Name   string
	Target string
	Passed bool
	Error  string
}

type PerformanceResult struct {
	Enabled  bool   `json:"enabled"`
	Passed   bool   `json:"passed"`
	VUs      int    `json:"vus"`
	Duration string `json:"duration"`

	AvgLatencyMs    float64 `json:"avgLatencyMs"`
	MinLatencyMs    float64 `json:"minLatencyMs"`
	MedianLatencyMs float64 `json:"medianLatencyMs"`
	MaxLatencyMs    float64 `json:"maxLatencyMs"`
	P90LatencyMs    float64 `json:"p90LatencyMs"`
	P95LatencyMs    float64 `json:"p95LatencyMs"`
	MaxP95LatencyMs float64 `json:"maxP95LatencyMs"`

	ErrorRate    float64 `json:"errorRate"`
	MaxErrorRate float64 `json:"maxErrorRate"`

	RequestCount  int64   `json:"requestCount"`
	Iterations    int64   `json:"iterations"`
	ChecksTotal   int64   `json:"checksTotal"`
	CheckPassRate float64 `json:"checkPassRate"`

	Error  string `json:"error,omitempty"`
	Output string `json:"output,omitempty"`
}

func WriteMarkdown(cfg *config.Config, data ReportData) (string, error) {
	if err := os.MkdirAll("reports", 0755); err != nil {
		return "", fmt.Errorf("create reports directory: %w", err)
	}

	filename := buildReportFilename(cfg.Service.Name, data.StartedAt, "md")
	reportPath := filepath.Join("reports", filename)

	content := buildMarkdown(cfg, data)

	if err := os.WriteFile(reportPath, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("write markdown report: %w", err)
	}

	return reportPath, nil
}

func WriteJSON(cfg *config.Config, data ReportData) (string, error) {
	if err := os.MkdirAll("reports", 0755); err != nil {
		return "", fmt.Errorf("create reports directory: %w", err)
	}

	filename := buildReportFilename(cfg.Service.Name, data.StartedAt, "json")
	reportPath := filepath.Join("reports", filename)

	content, err := json.MarshalIndent(data, "", " ")
	if err != nil {
		return "", fmt.Errorf("marshal json report: %w", err)
	}

	if err := os.WriteFile(reportPath, content, 0644); err != nil {
		return "", fmt.Errorf("write json report: %w", err)
	}

	return reportPath, nil
}
