package history

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/HXong/predeploy-guard/internal/report"
)

const HistoryFilename = "history.json"

type RunSummary struct {
	RunID        string    `json:"runId"`
	ServiceName  string    `json:"serviceName"`
	Profile      string    `json:"profile,omitempty"`
	Passed       bool      `json:"passed"`
	StartedAt    time.Time `json:"startedAt"`
	FinishedAt   time.Time `json:"finishedAt"`
	MarkdownPath string    `json:"markdownPath"`
	JSONPath     string    `json:"jsonPath"`

	P95LatencyMs float64 `json:"p95LatencyMs,omitempty"`
	ErrorRate    float64 `json:"errorRate,omitempty"`
}

func AppendRun(configDir string, data report.ReportData, markdownPath string, jsonPath string) error {
	historyPath := filepath.Join(configDir, "reports", HistoryFilename)

	existing, err := ReadHistory(configDir)
	if err != nil {
		return err
	}

	summary := BuildRunSummary(data, markdownPath, jsonPath)
	existing = append(existing, summary)

	content, err := json.MarshalIndent(existing, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal history: %w", err)
	}

	if err := os.WriteFile(historyPath, content, 0644); err != nil {
		return fmt.Errorf("write history: %w", err)
	}

	return nil
}

func FindRun(configDir string, runID string) (RunSummary, error) {
	summaries, err := ReadHistory(configDir)
	if err != nil {
		return RunSummary{}, err
	}

	for _, summary := range summaries {
		if summary.RunID == runID {
			return summary, nil
		}
	}

	return RunSummary{}, fmt.Errorf("run %q not found", runID)
}

func ReadHistory(configDir string) ([]RunSummary, error) {
	historyPath := filepath.Join(configDir, "reports", HistoryFilename)

	data, err := os.ReadFile(historyPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []RunSummary{}, nil
		}

		return nil, fmt.Errorf("read history: %w", err)
	}

	var summaries []RunSummary
	if err := json.Unmarshal(data, &summaries); err != nil {
		return nil, fmt.Errorf("parse history: %w", err)
	}

	return summaries, nil
}
