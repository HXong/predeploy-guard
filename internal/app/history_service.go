package app

import (
	"github.com/HXong/predeploy-guard/internal/config"
	"github.com/HXong/predeploy-guard/internal/history"
)

type HistoryService struct {
	cfg *config.Config
}

func NewHistoryService(cfg *config.Config) *HistoryService {
	return &HistoryService{
		cfg: cfg,
	}
}

func (s *HistoryService) ListRuns() ([]history.RunSummary, error) {
	return history.ReadHistory(s.cfg.ConfigDir)
}

func (s *HistoryService) FindRun(runID string) (history.RunSummary, error) {
	return history.FindRun(s.cfg.ConfigDir, runID)
}

func (s *HistoryService) CompareRuns(baseRunID string, targetRunID string) (history.RunSummary, history.RunSummary, error) {
	return history.FindTwoRuns(s.cfg.ConfigDir, baseRunID, targetRunID)
}
