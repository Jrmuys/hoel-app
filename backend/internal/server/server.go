package server

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"

	"hoel-app/backend/internal/db"
	"hoel-app/backend/internal/integration"
)

type Server struct {
	httpServer              *http.Server
	monitoring              *db.MonitoringRepository
	pghRepository           *db.PGHRepository
	tickTickRepository      *db.TickTickRepository
	tickTickService         *integration.TickTickService
	integrationClient       *integration.Client
	tickTickOAuth           *integration.TickTickOAuthService
	tickTickAPIRoot         string
	tickTickShoppingProject string
	tickTickDailyTag        string
	tickTickMaintenanceTag  string
	tickTickStateStore      map[string]time.Time
	tickTickStateLock       sync.Mutex
	eventSubscribers        map[chan string]struct{}
	eventSubscribersLock    sync.RWMutex
}

type healthResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
}

func New(address string, readTimeout, writeTimeout time.Duration, allowedOrigins []string, monitoring *db.MonitoringRepository, pghRepository *db.PGHRepository, tickTickRepository *db.TickTickRepository, tickTickService *integration.TickTickService, integrationClient *integration.Client, tickTickOAuth *integration.TickTickOAuthService, tickTickAPIRoot, tickTickShoppingProject, tickTickDailyTag, tickTickMaintenanceTag string) *Server {
	mux := http.NewServeMux()
	server := &Server{
		monitoring:              monitoring,
		pghRepository:           pghRepository,
		tickTickRepository:      tickTickRepository,
		tickTickService:         tickTickService,
		integrationClient:       integrationClient,
		tickTickOAuth:           tickTickOAuth,
		tickTickAPIRoot:         tickTickAPIRoot,
		tickTickShoppingProject: strings.TrimSpace(tickTickShoppingProject),
		tickTickDailyTag:        normalizeTickTickTag(tickTickDailyTag),
		tickTickMaintenanceTag:  normalizeTickTickTag(tickTickMaintenanceTag),
		tickTickStateStore:      map[string]time.Time{},
		eventSubscribers:        map[chan string]struct{}{},
	}
	mux.HandleFunc("/api/health", healthHandler)
	mux.HandleFunc("/api/status-bar", server.statusBarHandler)
	mux.HandleFunc("/api/status-bar/alerts/clear", server.statusAlertsClearHandler)
	mux.HandleFunc("/api/events", server.eventsHandler)
	mux.HandleFunc("/api/daily-operations", server.dailyOperationsHandler)
	mux.HandleFunc("/api/logistics-planning", server.logisticsPlanningHandler)
	mux.HandleFunc("/api/ticktick/oauth/start", server.tickTickOAuthStartHandler)
	mux.HandleFunc("/api/ticktick/oauth/callback", server.tickTickOAuthCallbackHandler)
	mux.HandleFunc("/api/ticktick/tasks/create", server.tickTickCreateTaskHandler)
	mux.HandleFunc("/api/ticktick/tasks/update", server.tickTickUpdateTaskHandler)
	mux.HandleFunc("/api/ticktick/tasks/complete", server.tickTickCompleteTaskHandler)
	mux.HandleFunc("/api/debug/ticktick-projects", server.tickTickProjectsHandler)
	mux.HandleFunc("/api/debug/ticktick-sync-now", server.tickTickSyncNowHandler)
	handler := newCORSSettings(allowedOrigins).wrap(mux)

	server.httpServer = &http.Server{
		Addr:         address,
		Handler:      handler,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
	}

	return server
}

func normalizeTickTickTag(tag string) string {
	trimmed := strings.TrimSpace(strings.ToLower(tag))
	return strings.TrimPrefix(trimmed, "#")
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
