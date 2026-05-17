package server

import (
	"fmt"
	"net/http"

	"github.com/HXong/predeploy-guard/internal/app"
	"github.com/HXong/predeploy-guard/internal/config"
)

type Server struct {
	cfg            *config.Config
	historyService *app.HistoryService
	reportService  *app.ReportService
	mux            *http.ServeMux
}

func New(cfg *config.Config) *Server {
	historyService := app.NewHistoryService(cfg)

	server := &Server{
		cfg:            cfg,
		historyService: historyService,
		reportService:  app.NewReportService(historyService),
		mux:            http.NewServeMux(),
	}

	server.registerRoutes()
	return server
}

func (s *Server) Start(addr string) error {
	fmt.Printf("PreDeploy Guard server listening on http://%s\n", addr)
	return http.ListenAndServe(addr, s.mux)
}

func (s *Server) registerRoutes() {
	s.mux.HandleFunc("GET /api/health", s.handleHealth)
	s.mux.HandleFunc("GET /api/history", s.handleHistory)
	s.mux.HandleFunc("GET /api/history/", s.handleHistoryRun)
	s.mux.HandleFunc("GET /api/reports/", s.handleReport)
}
