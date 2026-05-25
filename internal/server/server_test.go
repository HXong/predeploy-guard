package server

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
)

func TestServerRoutesExposeSafeEndpoints(t *testing.T) {
	configPath := filepath.Clean("../../examples/predeploy.yaml")

	srv, err := New(configPath)
	if err != nil {
		t.Fatalf("new server: %v", err)
	}

	tests := []struct {
		method     string
		path       string
		body       string
		wantStatus int
		wantJSON   bool
	}{
		{method: http.MethodGet, path: "/api/health", wantStatus: http.StatusOK, wantJSON: true},
		{method: http.MethodGet, path: "/api/config/summary", wantStatus: http.StatusOK, wantJSON: true},
		{method: http.MethodGet, path: "/api/config/explain", wantStatus: http.StatusOK, wantJSON: true},
		{method: http.MethodGet, path: "/api/history", wantStatus: http.StatusOK, wantJSON: true},
		{method: http.MethodGet, path: "/api/runs", wantStatus: http.StatusOK, wantJSON: true},
		{method: http.MethodGet, path: "/api/runs/missing-task", wantStatus: http.StatusNotFound, wantJSON: true},
		{method: http.MethodGet, path: "/api/runs/missing-task/logs", wantStatus: http.StatusNotFound, wantJSON: true},
		{method: http.MethodPost, path: "/api/runs", body: "{", wantStatus: http.StatusBadRequest, wantJSON: true},
		{method: http.MethodGet, path: "/api/config", wantStatus: http.StatusNotFound},
	}

	for _, test := range tests {
		t.Run(test.method+" "+test.path, func(t *testing.T) {
			req := httptest.NewRequest(test.method, test.path, strings.NewReader(test.body))
			rec := httptest.NewRecorder()

			srv.mux.ServeHTTP(rec, req)

			if rec.Code != test.wantStatus {
				t.Fatalf("status = %d, want %d; body: %s", rec.Code, test.wantStatus, rec.Body.String())
			}

			if test.wantJSON {
				contentType := rec.Header().Get("Content-Type")
				if !strings.HasPrefix(contentType, "application/json") {
					t.Fatalf("Content-Type = %q, want application/json", contentType)
				}
			}
		})
	}
}
