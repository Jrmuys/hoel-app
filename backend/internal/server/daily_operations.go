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
	NextPickupDate             string `json:"nextPickupDate"`
	NextTrashPickupDate        string `json:"nextTrashPickupDate"`
	NextRecyclingPickupDate    string `json:"nextRecyclingPickupDate"`
	IsRecyclingWeek            bool   `json:"isRecyclingWeek"`
	ShowIndicator              bool   `json:"showIndicator"`
	ShowTrashTakeOutReminder   bool   `json:"showTrashTakeOutReminder"`
	ShowRecyclingTakeOutReminder bool `json:"showRecyclingTakeOutReminder"`
}

func (s *Server) dailyOperationsHandler(w http.ResponseWriter, r *http.Request) {
	payload := dailyOperationsResponse{
		Tasks: []dailyTaskResponse{},
		Garbage: garbageResponse{
			NextPickupDate:               "",
			NextTrashPickupDate:          "",
			NextRecyclingPickupDate:      "",
			IsRecyclingWeek:              false,
			ShowIndicator:                false,
			ShowTrashTakeOutReminder:     false,
			ShowRecyclingTakeOutReminder: false,
		},
	}

	if s.pghRepository != nil {
		schedule, err := s.pghRepository.GetLatestSchedule(r.Context())
		if err != nil {
			http.Error(w, "unable to load daily operations", http.StatusInternalServerError)
			return
		}

		if schedule != nil {
			now := time.Now().UTC()
			payload.Garbage.NextPickupDate = schedule.NextPickupDate.UTC().Format(time.RFC3339)
			payload.Garbage.NextTrashPickupDate = schedule.NextPickupDate.UTC().Format(time.RFC3339)
			payload.Garbage.IsRecyclingWeek = schedule.IsRecyclingWeek
			payload.Garbage.ShowTrashTakeOutReminder = isWithinNextDay(now, schedule.NextPickupDate)

			if schedule.NextRecyclingDate != nil {
				payload.Garbage.NextRecyclingPickupDate = schedule.NextRecyclingDate.UTC().Format(time.RFC3339)
				payload.Garbage.ShowRecyclingTakeOutReminder = isWithinNextDay(now, *schedule.NextRecyclingDate)
			}

			payload.Garbage.ShowIndicator = payload.Garbage.ShowTrashTakeOutReminder || payload.Garbage.ShowRecyclingTakeOutReminder || schedule.ShowIndicator
		}
	}

	writeJSON(w, http.StatusOK, payload)
}

func isWithinNextDay(now, scheduledAt time.Time) bool {
	delta := scheduledAt.UTC().Sub(now.UTC())
	return delta >= 0 && delta <= 24*time.Hour
}
