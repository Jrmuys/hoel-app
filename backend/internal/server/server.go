package server

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"hoel-app/backend/internal/db"
	"hoel-app/backend/internal/integration"
)

type Server struct {
	httpServer        *http.Server
	monitoring        *db.MonitoringRepository
	pghRepository     *db.PGHRepository
	integrationClient *integration.Client
}

type healthResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
}

func New(address string, readTimeout, writeTimeout time.Duration, allowedOrigins []string, monitoring *db.MonitoringRepository, pghRepository *db.PGHRepository, integrationClient *integration.Client) *Server {
	mux := http.NewServeMux()
	server := &Server{monitoring: monitoring, pghRepository: pghRepository, integrationClient: integrationClient}
	mux.HandleFunc("/api/health", healthHandler)
	mux.HandleFunc("/api/status-bar", server.statusBarHandler)
	mux.HandleFunc("/api/daily-operations", server.dailyOperationsHandler)
	handler := newCORSSettings(allowedOrigins).wrap(mux)

	server.httpServer = &http.Server{
		Addr:         address,
		Handler:      handler,
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
