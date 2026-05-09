package runner

import (
	"fmt"

	"github.com/HXong/predeploy-guard/internal/config"
	"github.com/HXong/predeploy-guard/internal/report"
)

type writtenReports struct {
	MarkdownPath string
	JSONPath     string
}

func writeReports(cfg *config.Config, data report.ReportData) (writtenReports, error) {
	markdownPath, err := report.WriteMarkdown(cfg, data)
	if err != nil {
		return writtenReports{}, err
	}

	jsonPath, err := report.WriteJSON(cfg, data)
	if err != nil {
		return writtenReports{}, err
	}

	return writtenReports{
		MarkdownPath: markdownPath,
		JSONPath:     jsonPath,
	}, nil
}

func printReportPaths(paths writtenReports) {
	fmt.Printf("Markdown report written to: %s\n", paths.MarkdownPath)
	fmt.Printf("JSON report written to: %s\n", paths.JSONPath)
}
