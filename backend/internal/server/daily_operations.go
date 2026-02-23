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
	NextPickupDate               string `json:"nextPickupDate"`
	NextTrashPickupDate          string `json:"nextTrashPickupDate"`
	NextRecyclingPickupDate      string `json:"nextRecyclingPickupDate"`
	IsRecyclingWeek              bool   `json:"isRecyclingWeek"`
	ShowIndicator                bool   `json:"showIndicator"`
	ShowTrashTakeOutReminder     bool   `json:"showTrashTakeOutReminder"`
	ShowRecyclingTakeOutReminder bool   `json:"showRecyclingTakeOutReminder"`
}

func (s *Server) dailyOperationsHandler(w http.ResponseWriter, r *http.Request) {
	now := time.Now().UTC()

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

	if s.tickTickRepository != nil {
		tasks, err := s.tickTickRepository.ListTasksDueBetween(r.Context(), now.Add(-24*time.Hour), now.Add(24*time.Hour))
		if err != nil {
			http.Error(w, "unable to load daily operations", http.StatusInternalServerError)
			return
		}

		for _, task := range tasks {
			payload.Tasks = append(payload.Tasks, dailyTaskResponse{
				ID:        task.ID,
				Title:     task.Title,
				DueAt:     task.DueAt.UTC().Format(time.RFC3339),
				Completed: task.Completed,
				Source:    "ticktick",
			})
		}
	}

	if s.pghRepository != nil {
		schedule, err := s.pghRepository.GetLatestSchedule(r.Context())
		if err != nil {
			http.Error(w, "unable to load daily operations", http.StatusInternalServerError)
			return
		}

		if schedule != nil {
			payload.Garbage.NextPickupDate = schedule.NextPickupDate.UTC().Format(time.RFC3339)
			payload.Garbage.NextTrashPickupDate = schedule.NextPickupDate.UTC().Format(time.RFC3339)
			payload.Garbage.IsRecyclingWeek = schedule.IsRecyclingWeek
			payload.Garbage.ShowTrashTakeOutReminder = isWithinNextDay(now, schedule.NextPickupDate)
			if payload.Garbage.ShowTrashTakeOutReminder {
				payload.Tasks = append(payload.Tasks, dailyTaskResponse{
					ID:        "system-trash-takeout",
					Title:     "Take out trash tonight",
					DueAt:     schedule.NextPickupDate.UTC().Format(time.RFC3339),
					Completed: false,
					Source:    "system",
				})
			}

			if schedule.NextRecyclingDate != nil {
				payload.Garbage.NextRecyclingPickupDate = schedule.NextRecyclingDate.UTC().Format(time.RFC3339)
				payload.Garbage.ShowRecyclingTakeOutReminder = isWithinNextDay(now, *schedule.NextRecyclingDate)
				if payload.Garbage.ShowRecyclingTakeOutReminder {
					payload.Tasks = append(payload.Tasks, dailyTaskResponse{
						ID:        "system-recycling-takeout",
						Title:     "Take out recycling tonight",
						DueAt:     schedule.NextRecyclingDate.UTC().Format(time.RFC3339),
						Completed: false,
						Source:    "system",
					})
				}
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
