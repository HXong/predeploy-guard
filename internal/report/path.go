package report

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/HXong/predeploy-guard/internal/config"
)

func reportPath(cfg *config.Config, data ReportData, extension string) (string, error) {
	reportsDir := filepath.Join(cfg.ConfigDir, "reports")

	if err := os.MkdirAll(reportsDir, 0755); err != nil {
		return "", fmt.Errorf("create reports directory: %w", err)
	}

	filename := buildReportFilename(cfg.Service.Name, data.StartedAt, extension)
	return filepath.Join(reportsDir, filename), nil
}
