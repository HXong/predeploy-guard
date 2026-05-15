package history

import (
	"fmt"
	"path/filepath"

	"github.com/HXong/predeploy-guard/internal/report"
)

func BuildRunSummary(data report.ReportData, markdownPath string, jsonPath string) RunSummary {
	runID := fmt.Sprintf(
		"%s-%s",
		data.StartedAt.Format("2006-01-02-150405"),
		data.ServiceName,
	)

	return RunSummary{
		RunID:        runID,
		ServiceName:  data.ServiceName,
		Profile:      data.ActiveProfile,
		Passed:       data.Passed,
		StartedAt:    data.StartedAt,
		FinishedAt:   data.FinishedAt,
		MarkdownPath: filepath.ToSlash(markdownPath),
		JSONPath:     filepath.ToSlash(jsonPath),

		P95LatencyMs: data.PerformanceResult.P95LatencyMs,
		ErrorRate:    data.PerformanceResult.ErrorRate,
	}
}
