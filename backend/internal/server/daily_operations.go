package server

import (
	"net/http"
	"strings"
	"time"
)

type dailyOperationsResponse struct {
	Tasks            []dailyTaskResponse `json:"tasks"`
	ShoppingTasks    []dailyTaskResponse `json:"shoppingTasks"`
	MaintenanceTasks []dailyTaskResponse `json:"maintenanceTasks"`
	Garbage          garbageResponse     `json:"garbage"`
}

type dailyTaskResponse struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	DueAt     string `json:"dueAt"`
	HasTime   bool   `json:"hasTime"`
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
		Tasks:            []dailyTaskResponse{},
		ShoppingTasks:    []dailyTaskResponse{},
		MaintenanceTasks: []dailyTaskResponse{},
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
		tasks, err := s.tickTickRepository.ListIncompleteTasks(r.Context())
		if err != nil {
			http.Error(w, "unable to load daily operations", http.StatusInternalServerError)
			return
		}

		endOfToday := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, int(time.Second-time.Nanosecond), time.UTC)

		for _, task := range tasks {
			responseTask := dailyTaskResponse{
				ID:        task.ID,
				Title:     task.Title,
				DueAt:     task.DueAt.UTC().Format(time.RFC3339),
				HasTime:   task.HasTime,
				Completed: task.Completed,
				Source:    "ticktick",
			}

			isShoppingTask := s.tickTickShoppingProject != "" && strings.EqualFold(task.SourceProject, s.tickTickShoppingProject)
			if isShoppingTask {
				payload.ShoppingTasks = append(payload.ShoppingTasks, responseTask)
				continue
			}

			isMaintenanceTask := tickTickTaskHasTag(task.Tags, s.tickTickMaintenanceTag)
			if isMaintenanceTask {
				payload.MaintenanceTasks = append(payload.MaintenanceTasks, responseTask)
			}

			isDailyTask := tickTickTaskHasTag(task.Tags, s.tickTickDailyTag)
			isDueTodayOrOverdue := !task.DueAt.After(endOfToday)
			if isDailyTask || isDueTodayOrOverdue {
				payload.Tasks = append(payload.Tasks, responseTask)
			}
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
					HasTime:   true,
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
						HasTime:   true,
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

func tickTickTaskHasTag(tags []string, expected string) bool {
	expected = normalizeTickTickTag(expected)
	if expected == "" || len(tags) == 0 {
		return false
	}

	for _, tag := range tags {
		normalized := normalizeTickTickTag(tag)
		if normalized == expected {
			return true
		}
	}

	return false
}

func isWithinNextDay(now, scheduledAt time.Time) bool {
	delta := scheduledAt.UTC().Sub(now.UTC())
	return delta >= 0 && delta <= 24*time.Hour
}
