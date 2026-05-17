package app

import (
	"os"

	"github.com/HXong/predeploy-guard/internal/history"
)

type ReportService struct {
	historyService *HistoryService
}

func NewReportService(historyService *HistoryService) *ReportService {
	return &ReportService{
		historyService: historyService,
	}
}

func (s *ReportService) ReadJSONReport(runID string) ([]byte, error) {
	run, err := s.historyService.FindRun(runID)
	if err != nil {
		return nil, err
	}

	return os.ReadFile(run.JSONPath)
}

func (s *ReportService) ReadMarkdownReport(runID string) ([]byte, error) {
	run, err := s.historyService.FindRun(runID)
	if err != nil {
		return nil, err
	}

	return os.ReadFile(run.MarkdownPath)
}

func (s *ReportService) FindRun(runID string) (history.RunSummary, error) {
	return s.historyService.FindRun(runID)
}
