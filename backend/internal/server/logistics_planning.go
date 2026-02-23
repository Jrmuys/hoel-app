package server

import (
	"net/http"
	"sort"
	"strings"
)

type logisticsPlanningResponse struct {
	ShoppingTasks    []dailyTaskResponse `json:"shoppingTasks"`
	MaintenanceTasks []dailyTaskResponse `json:"maintenanceTasks"`
}

func (s *Server) logisticsPlanningHandler(w http.ResponseWriter, r *http.Request) {
	payload := logisticsPlanningResponse{
		ShoppingTasks:    []dailyTaskResponse{},
		MaintenanceTasks: []dailyTaskResponse{},
	}

	if s.tickTickRepository == nil {
		writeJSON(w, http.StatusOK, payload)
		return
	}

	tasks, err := s.tickTickRepository.ListIncompleteTasks(r.Context())
	if err != nil {
		http.Error(w, "unable to load logistics planning", http.StatusInternalServerError)
		return
	}

	for _, task := range tasks {
		dueAt := ""
		if !task.DueAt.IsZero() {
			dueAt = task.DueAt.UTC().Format(timeLayoutRFC3339)
		}

		responseTask := dailyTaskResponse{
			ID:        task.ID,
			Title:     task.Title,
			DueAt:     dueAt,
			HasTime:   task.HasTime,
			Completed: task.Completed,
			Source:    "ticktick",
		}

		if s.tickTickShoppingProject != "" && strings.EqualFold(task.SourceProject, s.tickTickShoppingProject) {
			payload.ShoppingTasks = append(payload.ShoppingTasks, responseTask)
			continue
		}

		if tickTickTaskHasTag(task.Tags, s.tickTickMaintenanceTag) {
			payload.MaintenanceTasks = append(payload.MaintenanceTasks, responseTask)
		}
	}

	sort.Slice(payload.ShoppingTasks, func(i, j int) bool {
		return payload.ShoppingTasks[i].DueAt < payload.ShoppingTasks[j].DueAt
	})
	sort.Slice(payload.MaintenanceTasks, func(i, j int) bool {
		return payload.MaintenanceTasks[i].DueAt < payload.MaintenanceTasks[j].DueAt
	})

	writeJSON(w, http.StatusOK, payload)
}

const timeLayoutRFC3339 = "2006-01-02T15:04:05Z07:00"
