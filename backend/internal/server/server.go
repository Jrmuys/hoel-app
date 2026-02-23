package server

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"hoel-app/backend/internal/db"
)

type Server struct {
	httpServer  *http.Server
	monitoring *db.MonitoringRepository
}

type healthResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
}

func New(address string, readTimeout, writeTimeout time.Duration, monitoring *db.MonitoringRepository) *Server {
	mux := http.NewServeMux()
	server := &Server{monitoring: monitoring}
	mux.HandleFunc("/api/health", healthHandler)
	mux.HandleFunc("/api/status-bar", server.statusBarHandler)

	server.httpServer = &http.Server{
		Addr:         address,
		Handler:      mux,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
	}

	return server
}

func (s *Server) ListenAndServe() error {
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	payload := healthResponse{
		Status:    "ok",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	writeJSON(w, http.StatusOK, payload)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	body, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, "unable to marshal response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(body)
}
