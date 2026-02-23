package server

import (
	"net/http"
	"time"
)

type dailyOperationsResponse struct {
	Tasks   []dailyTaskResponse `json:"tasks"`
	Garbage garbageResponse     `json:"garbage"`
}

type dailyTaskResponse struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	DueAt     string `json:"dueAt"`
	Completed bool   `json:"completed"`
	Source    string `json:"source"`
}

type garbageResponse struct {
	NextPickupDate  string `json:"nextPickupDate"`
	IsRecyclingWeek bool   `json:"isRecyclingWeek"`
	ShowIndicator   bool   `json:"showIndicator"`
}

func (s *Server) dailyOperationsHandler(w http.ResponseWriter, r *http.Request) {
	payload := dailyOperationsResponse{
		Tasks: []dailyTaskResponse{},
		Garbage: garbageResponse{
			NextPickupDate:  "",
			IsRecyclingWeek: false,
			ShowIndicator:   false,
		},
	}

	if s.pghRepository != nil {
		schedule, err := s.pghRepository.GetLatestSchedule(r.Context())
		if err != nil {
			http.Error(w, "unable to load daily operations", http.StatusInternalServerError)
			return
		}

		if schedule != nil {
			payload.Garbage.NextPickupDate = schedule.NextPickupDate.UTC().Format(time.RFC3339)
			payload.Garbage.IsRecyclingWeek = schedule.IsRecyclingWeek
			payload.Garbage.ShowIndicator = schedule.ShowIndicator
		}
	}

	writeJSON(w, http.StatusOK, payload)
}
