package report

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/HXong/predeploy-guard/internal/checker"
	"github.com/HXong/predeploy-guard/internal/config"
	"github.com/HXong/predeploy-guard/internal/gateway"
	"github.com/HXong/predeploy-guard/internal/workload"
)

type ReportData struct {
	ServiceName        string
	ActiveProfile      string `json:"activeProfile,omitempty"`
	Image              string
	Runtime            string
	BaseURL            string
	RuntimeEnvironment RuntimeEnvironment `json:"runtimeEnvironment"`
	RunPhases          []RunPhase         `json:"runPhases"`
	StartedAt          time.Time
	FinishedAt         time.Time
	BuildResult        BuildResult
	ReadinessResults   []ReadinessResult
	GatewayResults     []gateway.RouteResult
	Results            []checker.SmokeResult
	WorkloadResults    []workload.Result
	PerformanceResult  PerformanceResult
	Passed             bool
	Logs               string
}

type RuntimeEnvironment struct {
	Runtime         string `json:"runtime"`
	Name            string `json:"name,omitempty"`
	BaseURL         string `json:"baseUrl,omitempty"`
	WorkloadBaseURL string `json:"workloadBaseUrl,omitempty"`
}

type RunPhase struct {
	Name       string        `json:"name"`
	Status     string        `json:"status"`
	StartedAt  time.Time     `json:"startedAt"`
	FinishedAt time.Time     `json:"finishedAt"`
	Duration   time.Duration `json:"duration"`
	Error      string        `json:"error,omitempty"`
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
	path, err := reportPath(cfg, data, "md")
	if err != nil {
		return "", err
	}

	content := buildMarkdown(cfg, data)

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("write markdown report: %w", err)
	}

	return path, nil
}

func WriteJSON(cfg *config.Config, data ReportData) (string, error) {
	path, err := reportPath(cfg, data, "json")
	if err != nil {
		return "", err
	}

	content, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal json report: %w", err)
	}

	if err := os.WriteFile(path, content, 0644); err != nil {
		return "", fmt.Errorf("write json report: %w", err)
	}

	return path, nil
}
