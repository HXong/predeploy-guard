package server

import (
	"net/http"

	"github.com/HXong/predeploy-guard/internal/app"
	"github.com/HXong/predeploy-guard/internal/config"
)

type Server struct {
	configService  *app.ConfigService
	historyService *app.HistoryService
	reportService  *app.ReportService
	runManager     *app.RunManager
	mux            *http.ServeMux
}

func New(configPath string) (*Server, error) {
	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, err
	}

	historyService := app.NewHistoryService(cfg)

	server := &Server{
		configService:  app.NewConfigService(cfg),
		historyService: historyService,
		reportService:  app.NewReportService(historyService),
		runManager:     app.NewRunManager(app.NewRunService(cfg.ConfigPath)),
		mux:            http.NewServeMux(),
	}

	server.registerRoutes()
	return server, nil
}

func (s *Server) Start(addr string) error {
	return http.ListenAndServe(addr, s.mux)
}

func (s *Server) registerRoutes() {
	s.mux.HandleFunc("GET /api/health", s.handleHealth)
	s.mux.HandleFunc("GET /api/config/summary", s.handleConfigSummary)
	s.mux.HandleFunc("GET /api/config/explain", s.handleConfigExplain)
	s.mux.HandleFunc("GET /api/history", s.handleHistory)
	s.mux.HandleFunc("GET /api/history/", s.handleHistoryRun)
	s.mux.HandleFunc("GET /api/reports/", s.handleReport)
	s.mux.HandleFunc("GET /api/runs", s.handleRuns)
	s.mux.HandleFunc("GET /api/runs/{taskID}", s.handleRunTask)
	s.mux.HandleFunc("GET /api/runs/{taskID}/logs", s.handleRunTaskLogs)
	s.mux.HandleFunc("POST /api/runs", s.handleRun)
}
