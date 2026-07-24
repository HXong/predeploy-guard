package workload

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/HXong/predeploy-guard/internal/config"
)

type Result struct {
	Name           string
	Type           string
	Enabled        bool
	Passed         bool
	FailurePolicy  string
	Method         string
	Path           string
	TargetURL      string
	Duration       string
	RatePerSecond  int
	ExpectedStatus int
	StartedAt      time.Time
	FinishedAt     time.Time
	RequestCount   int
	SuccessCount   int
	FailureCount   int
	StatusCounts   map[int]int
	Error          string
}

func RunAll(ctx context.Context, cfg *config.Config, baseURL string) []Result {
	results := make([]Result, 0, len(cfg.Workloads))

	for _, workloadConfig := range cfg.Workloads {
		result := newResult(workloadConfig, baseURL)
		if !result.Enabled {
			result.Passed = true
			result.FinishedAt = result.StartedAt
			results = append(results, result)
			continue
		}

		switch workloadConfig.Type {
		case "http":
			result = runHTTP(ctx, workloadConfig, result)
		default:
			result.Error = fmt.Sprintf("unsupported workload type %q", workloadConfig.Type)
			result.FailureCount = 1
			result.FinishedAt = time.Now()
		}

		result.Passed = result.FailureCount == 0
		if !result.Passed && result.Error == "" {
			result.Error = fmt.Sprintf("%d request(s) failed", result.FailureCount)
		}

		results = append(results, result)
	}

	return results
}

func HasEnabled(cfg *config.Config) bool {
	for _, workloadConfig := range cfg.Workloads {
		if isEnabled(workloadConfig) {
			return true
		}
	}

	return false
}

func ShouldFailRun(results []Result) bool {
	for _, result := range results {
		if result.Enabled && !result.Passed && result.FailurePolicy == "fail" {
			return true
		}
	}

	return false
}

func newResult(workloadConfig config.WorkloadConfig, baseURL string) Result {
	return Result{
		Name:           workloadConfig.Name,
		Type:           workloadConfig.Type,
		Enabled:        isEnabled(workloadConfig),
		FailurePolicy:  workloadConfig.FailurePolicy,
		Method:         workloadConfig.Method,
		Path:           workloadConfig.Path,
		TargetURL:      joinURL(baseURL, workloadConfig.Path),
		Duration:       workloadConfig.Duration,
		RatePerSecond:  workloadConfig.RatePerSecond,
		ExpectedStatus: workloadConfig.ExpectedStatus,
		StartedAt:      time.Now(),
		StatusCounts:   make(map[int]int),
	}
}

func runHTTP(ctx context.Context, workloadConfig config.WorkloadConfig, result Result) Result {
	duration, err := time.ParseDuration(workloadConfig.Duration)
	if err != nil || duration <= 0 {
		result.FailureCount = 1
		result.Error = fmt.Sprintf("invalid workload duration %q", workloadConfig.Duration)
		result.FinishedAt = time.Now()
		return result
	}
	if workloadConfig.RatePerSecond <= 0 {
		result.FailureCount = 1
		result.Error = "workload ratePerSecond must be greater than 0"
		result.FinishedAt = time.Now()
		return result
	}

	interval := time.Second / time.Duration(workloadConfig.RatePerSecond)
	if interval <= 0 {
		interval = time.Nanosecond
	}

	requestTimeout := duration
	if requestTimeout > 10*time.Second {
		requestTimeout = 10 * time.Second
	}
	client := &http.Client{Timeout: requestTimeout}
	finished := time.NewTimer(duration)
	ticker := time.NewTicker(interval)
	defer finished.Stop()
	defer ticker.Stop()

	sendHTTPRequest(ctx, client, workloadConfig, &result)

	for {
		select {
		case <-ctx.Done():
			if result.Error == "" {
				result.Error = ctx.Err().Error()
				result.FailureCount++
			}
			result.FinishedAt = time.Now()
			return result
		case <-finished.C:
			result.FinishedAt = time.Now()
			return result
		case <-ticker.C:
			sendHTTPRequest(ctx, client, workloadConfig, &result)
		}
	}
}

func sendHTTPRequest(
	ctx context.Context,
	client *http.Client,
	workloadConfig config.WorkloadConfig,
	result *Result,
) {
	request, err := http.NewRequestWithContext(ctx, workloadConfig.Method, result.TargetURL, nil)
	result.RequestCount++
	if err != nil {
		result.FailureCount++
		if result.Error == "" {
			result.Error = err.Error()
		}
		return
	}

	response, err := client.Do(request)
	if err != nil {
		result.FailureCount++
		if result.Error == "" {
			result.Error = err.Error()
		}
		return
	}
	defer response.Body.Close()
	_, _ = io.Copy(io.Discard, response.Body)

	result.StatusCounts[response.StatusCode]++
	if response.StatusCode == workloadConfig.ExpectedStatus {
		result.SuccessCount++
		return
	}

	result.FailureCount++
	if result.Error == "" {
		result.Error = fmt.Sprintf(
			"received HTTP %d; expected HTTP %d",
			response.StatusCode,
			workloadConfig.ExpectedStatus,
		)
	}
}

func isEnabled(workloadConfig config.WorkloadConfig) bool {
	return workloadConfig.Enabled == nil || *workloadConfig.Enabled
}

func joinURL(baseURL string, path string) string {
	return strings.TrimRight(baseURL, "/") + "/" + strings.TrimLeft(path, "/")
}
