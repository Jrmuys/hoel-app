package server

import (
	"net/http"
	"strings"
	"time"
)

type statusBarResponse struct {
	SystemHealth string                    `json:"system_health"`
	Alerts       []statusAlertResponse     `json:"alerts"`
	Integrations []integrationHealthStatus `json:"integrations"`
	Timestamp    string                    `json:"timestamp"`
}

type statusAlertResponse struct {
	ID        int64  `json:"id"`
	Source    string `json:"source"`
	Severity  string `json:"severity"`
	Message   string `json:"message"`
	CreatedAt string `json:"created_at"`
}

type integrationHealthStatus struct {
	Service             string  `json:"service"`
	State               string  `json:"state"`
	LastSuccessAt       *string `json:"last_success_at"`
	ConsecutiveFailures int     `json:"consecutive_failures"`
}

func (s *Server) statusBarHandler(w http.ResponseWriter, r *http.Request) {
	if s.monitoring == nil {
		writeJSON(w, http.StatusOK, statusBarResponse{
			SystemHealth: "ok",
			Alerts:       []statusAlertResponse{},
			Integrations: []integrationHealthStatus{},
			Timestamp:    time.Now().UTC().Format(time.RFC3339),
		})
		return
	}

	statuses, err := s.monitoring.ListIntegrationStatus(r.Context())
	if err != nil {
		http.Error(w, "unable to load integration status", http.StatusInternalServerError)
		return
	}

	errorsList, err := s.monitoring.ListRecentUnresolvedErrors(r.Context(), 10)
	if err != nil {
		http.Error(w, "unable to load integration alerts", http.StatusInternalServerError)
		return
	}

	integrations := make([]integrationHealthStatus, 0, len(statuses))
	health := "ok"
	for _, status := range statuses {
		state := integrationState(status.ConsecutiveFailures)
		if state == "down" {
			health = "down"
		} else if state == "degraded" && health == "ok" {
			health = "degraded"
		}

		integrations = append(integrations, integrationHealthStatus{
			Service:             status.Service,
			State:               state,
			LastSuccessAt:       formatNullableTime(status.LastSuccessAt),
			ConsecutiveFailures: status.ConsecutiveFailures,
		})
	}

	alerts := make([]statusAlertResponse, 0, len(errorsList))
	for _, record := range errorsList {
		alerts = append(alerts, statusAlertResponse{
			ID:        record.ID,
			Source:    record.ServiceName,
			Severity:  severityFromStatusCode(record.HTTPStatus),
			Message:   buildAlertMessage(record.Endpoint, record.Message),
			CreatedAt: record.CreatedAt.UTC().Format(time.RFC3339),
		})
	}

	payload := statusBarResponse{
		SystemHealth: health,
		Alerts:       alerts,
		Integrations: integrations,
		Timestamp:    time.Now().UTC().Format(time.RFC3339),
	}

	writeJSON(w, http.StatusOK, payload)
}

func (s *Server) statusAlertsClearHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if s.monitoring == nil {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
		return
	}

	if err := s.monitoring.ClearUnresolvedAlerts(r.Context()); err != nil {
		http.Error(w, "unable to clear alerts", http.StatusInternalServerError)
		return
	}

	s.publishEvent("refresh")

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func integrationState(consecutiveFailures int) string {
	if consecutiveFailures >= 3 {
		return "down"
	}
	if consecutiveFailures > 0 {
		return "degraded"
	}
	return "healthy"
}

func severityFromStatusCode(httpStatus *int) string {
	if httpStatus != nil && *httpStatus >= 500 {
		return "critical"
	}
	return "warning"
}

func buildAlertMessage(endpoint, message string) string {
	trimmed := strings.TrimSpace(message)
	if trimmed == "" {
		return endpoint
	}
	return endpoint + ": " + trimmed
}

func formatNullableTime(value *time.Time) *string {
	if value == nil {
		return nil
	}

	formatted := value.UTC().Format(time.RFC3339)
	return &formatted
}
