package server

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

type tickTickCompleteTaskRequest struct {
	TaskID string `json:"taskId"`
}

type tickTickCompleteTaskResponse struct {
	Status   string `json:"status"`
	TaskID   string `json:"taskId"`
	SyncedAt string `json:"syncedAt"`
}

func (s *Server) tickTickCompleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if s.tickTickService == nil || !s.tickTickService.Enabled() {
		http.Error(w, "ticktick sync is not configured", http.StatusBadRequest)
		return
	}

	var request tickTickCompleteTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid request payload", http.StatusBadRequest)
		return
	}

	taskID := strings.TrimSpace(request.TaskID)
	if taskID == "" {
		http.Error(w, "taskId is required", http.StatusBadRequest)
		return
	}

	if err := s.tickTickService.CompleteTask(r.Context(), taskID); err != nil {
		http.Error(w, "unable to complete ticktick task", http.StatusBadGateway)
		return
	}

	if err := s.tickTickService.SyncOnce(r.Context()); err != nil {
		http.Error(w, "task completed but sync refresh failed", http.StatusBadGateway)
		return
	}

	writeJSON(w, http.StatusOK, tickTickCompleteTaskResponse{
		Status:   "ok",
		TaskID:   taskID,
		SyncedAt: time.Now().UTC().Format(time.RFC3339),
	})
}
