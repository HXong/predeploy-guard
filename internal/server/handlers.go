package server

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

type runRequest struct {
	Profile string `json:"profile"`
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
	})
}

func (s *Server) handleConfigSummary(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.configService.Summary())
}

func (s *Server) handleConfigExplain(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.configService.Explain())
}

func (s *Server) handleRun(w http.ResponseWriter, r *http.Request) {
	var request runRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil && err != io.EOF {
		writeError(w, http.StatusBadRequest, "invalid JSON request body")
		return
	}

	result, err := s.runService.Run(request.Profile)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleHistory(w http.ResponseWriter, r *http.Request) {
	runs, err := s.historyService.ListRuns()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, runs)
}

func (s *Server) handleHistoryRun(w http.ResponseWriter, r *http.Request) {
	runID := strings.TrimPrefix(r.URL.Path, "/api/history/")
	if runID == "" {
		writeError(w, http.StatusBadRequest, "run id is required")
		return
	}

	run, err := s.historyService.FindRun(runID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, run)
}

func (s *Server) handleReport(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/reports/"), "/")
	if len(parts) != 2 {
		writeError(w, http.StatusBadRequest, "expected /api/reports/{runId}/json or /api/reports/{runId}/markdown")
		return
	}

	runID := parts[0]
	reportType := parts[1]

	switch reportType {
	case "json":
		data, err := s.reportService.ReadJSONReport(runID)
		if err != nil {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(data)

	case "markdown":
		data, err := s.reportService.ReadMarkdownReport(runID)
		if err != nil {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}

		w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(data)

	default:
		writeError(w, http.StatusBadRequest, "unsupported report type")
	}
}
