package server

import (
	"net/http"
	"time"
)

type tickTickSyncNowResponse struct {
	Status        string `json:"status"`
	SyncedAt      string `json:"syncedAt"`
	DueTaskCount  int    `json:"dueTaskCount"`
	ProjectLinked bool   `json:"projectLinked"`
}

func (s *Server) tickTickSyncNowHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if s.tickTickService == nil || !s.tickTickService.Enabled() {
		http.Error(w, "ticktick sync is not configured", http.StatusBadRequest)
		return
	}

	if err := s.tickTickService.SyncOnce(r.Context()); err != nil {
		http.Error(w, "ticktick sync failed", http.StatusBadGateway)
		return
	}

	now := time.Now().UTC()
	dueTaskCount := 0
	if s.tickTickRepository != nil {
		tasks, err := s.tickTickRepository.ListTasksDueBetween(r.Context(), now.Add(-24*time.Hour), now.Add(24*time.Hour))
		if err != nil {
			http.Error(w, "ticktick sync completed but due-task query failed", http.StatusBadGateway)
			return
		}

		dueTaskCount = len(tasks)
	}

	writeJSON(w, http.StatusOK, tickTickSyncNowResponse{
		Status:        "ok",
		SyncedAt:      now.Format(time.RFC3339),
		DueTaskCount:  dueTaskCount,
		ProjectLinked: s.tickTickRepository != nil,
	})
}
