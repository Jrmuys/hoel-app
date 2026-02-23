package integration

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"hoel-app/backend/internal/db"
)

type TickTickService struct {
	client       *Client
	repository   *db.TickTickRepository
	oauth        *TickTickOAuthService
	apiRoot      string
	projectID    string
	pollInterval time.Duration
}

type tickTickProjectDataResponse struct {
	Tasks []tickTickTaskDTO `json:"tasks"`
}

type tickTickTaskDTO struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	DueDate  string `json:"dueDate"`
	Status   int    `json:"status"`
	Priority int    `json:"priority"`
}

func NewTickTickService(client *Client, repository *db.TickTickRepository, oauth *TickTickOAuthService, apiRoot, projectID string, pollInterval time.Duration) *TickTickService {
	if pollInterval <= 0 {
		pollInterval = 10 * time.Minute
	}

	return &TickTickService{
		client:       client,
		repository:   repository,
		oauth:        oauth,
		apiRoot:      strings.TrimSpace(apiRoot),
		projectID:    strings.TrimSpace(projectID),
		pollInterval: pollInterval,
	}
}

func (s *TickTickService) Enabled() bool {
	return s.client != nil && s.repository != nil && s.oauth != nil && s.apiRoot != "" && s.projectID != ""
}

func (s *TickTickService) Start(ctx context.Context) {
	if !s.Enabled() {
		log.Printf("ticktick sync disabled: missing required config or dependencies")
		return
	}

	log.Printf("ticktick sync started: project=%s interval=%s", s.projectID, s.pollInterval)

	if err := s.SyncOnce(ctx); err != nil {
		log.Printf("ticktick initial sync failed: %v", err)
	}

	ticker := time.NewTicker(s.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Printf("ticktick sync stopped: context canceled")
			return
		case <-ticker.C:
			if err := s.SyncOnce(ctx); err != nil {
				log.Printf("ticktick periodic sync failed: %v", err)
			}
		}
	}
}

func (s *TickTickService) SyncOnce(ctx context.Context) error {
	if !s.Enabled() {
		log.Printf("ticktick sync skipped: service not enabled")
		return nil
	}

	requestURL := fmt.Sprintf("%s/project/%s/data", strings.TrimRight(s.apiRoot, "/"), url.PathEscape(s.projectID))
	log.Printf("ticktick sync request: project=%s endpoint=%s", s.projectID, requestURL)

	accessToken, err := s.oauth.ResolveAccessToken(ctx)
	if err != nil {
		log.Printf("ticktick token resolution failed: %v", err)
		return err
	}

	response, err := s.client.Do(ctx, Request{
		Service: "ticktick",
		Method:  http.MethodGet,
		URL:     requestURL,
		Headers: map[string]string{
			"Accept":        "application/json",
			"Authorization": "Bearer " + accessToken,
		},
	})
	if err != nil {
		log.Printf("ticktick project data request failed: %v", err)
		return err
	}

	var payload tickTickProjectDataResponse
	if err := json.Unmarshal(response.Body, &payload); err != nil {
		log.Printf("ticktick response decode failed: %v", err)
		return fmt.Errorf("decode ticktick response: %w", err)
	}

	if len(payload.Tasks) == 0 {
		var decoded any
		if err := json.Unmarshal(response.Body, &decoded); err == nil {
			payload.Tasks = extractTickTickTasks(decoded)
		}
	}

	log.Printf("ticktick project data received: project=%s tasks=%d", s.projectID, len(payload.Tasks))
	if len(payload.Tasks) == 0 {
		log.Printf("ticktick project data empty tasks payload preview=%q", truncateTickTickBodyPreview(response.Body, 600))
	}

	tasks := make([]db.TickTickTask, 0, len(payload.Tasks))
	skippedWithoutDueDate := 0
	for _, task := range payload.Tasks {
		dueAt, ok, err := parseTickTickDueDate(task.DueDate)
		if err != nil {
			log.Printf("ticktick task due-date parse failed: task=%s value=%q err=%v", task.ID, task.DueDate, err)
			return fmt.Errorf("parse ticktick due date for task %s: %w", task.ID, err)
		}
		if !ok {
			skippedWithoutDueDate++
			continue
		}

		tasks = append(tasks, db.TickTickTask{
			ID:            strings.TrimSpace(task.ID),
			Title:         strings.TrimSpace(task.Title),
			DueAt:         dueAt,
			Completed:     tickTickCompleted(task.Status),
			SourceProject: s.projectID,
		})
	}

	if err := s.repository.ReplaceTasks(ctx, tasks); err != nil {
		log.Printf("ticktick cache update failed: %v", err)
		return err
	}

	log.Printf("ticktick sync completed: cached_tasks=%d skipped_without_due_date=%d", len(tasks), skippedWithoutDueDate)

	return nil
}

func (s *TickTickService) CompleteTask(ctx context.Context, taskID string) error {
	if !s.Enabled() {
		log.Printf("ticktick complete-task skipped: service not enabled")
		return fmt.Errorf("ticktick service is not enabled")
	}

	taskID = strings.TrimSpace(taskID)
	if taskID == "" {
		return fmt.Errorf("task id is required")
	}

	accessToken, err := s.oauth.ResolveAccessToken(ctx)
	if err != nil {
		log.Printf("ticktick complete-task token resolution failed: %v", err)
		return err
	}

	baseURL := strings.TrimRight(s.apiRoot, "/")
	requestURLs := []string{
		fmt.Sprintf("%s/project/%s/task/%s/complete", baseURL, url.PathEscape(s.projectID), url.PathEscape(taskID)),
		fmt.Sprintf("%s/task/%s/complete", baseURL, url.PathEscape(taskID)),
	}

	var lastError error
	for _, requestURL := range requestURLs {
		log.Printf("ticktick complete-task request: task=%s endpoint=%s", taskID, requestURL)

		_, err = s.client.Do(ctx, Request{
			Service: "ticktick",
			Method:  http.MethodPost,
			URL:     requestURL,
			Headers: map[string]string{
				"Accept":        "application/json",
				"Authorization": "Bearer " + accessToken,
			},
		})
		if err == nil {
			lastError = nil
			break
		}

		var statusError HTTPStatusError
		if errors.As(err, &statusError) && statusError.StatusCode == http.StatusNotFound {
			lastError = err
			continue
		}

		log.Printf("ticktick complete-task request failed: task=%s err=%v", taskID, err)
		return err
	}

	if lastError != nil {
		log.Printf("ticktick complete-task request failed on all endpoints: task=%s err=%v", taskID, lastError)
		return lastError
	}

	if s.repository != nil {
		if err := s.repository.MarkTaskCompleted(ctx, taskID); err != nil {
			log.Printf("ticktick local cache complete-task update failed: task=%s err=%v", taskID, err)
			return err
		}
	}

	log.Printf("ticktick complete-task completed: task=%s", taskID)
	return nil
}

func parseTickTickDueDate(value string) (time.Time, bool, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return time.Time{}, false, nil
	}

	layouts := []string{
		time.RFC3339,
		"2006-01-02T15:04:05.000-0700",
		"2006-01-02T15:04:05-0700",
		"2006-01-02",
	}

	for _, layout := range layouts {
		if parsed, err := time.Parse(layout, trimmed); err == nil {
			return parsed.UTC(), true, nil
		}
	}

	return time.Time{}, false, fmt.Errorf("unsupported due date format %q", value)
}

func tickTickCompleted(status int) bool {
	return status != 0
}

func extractTickTickTasks(value any) []tickTickTaskDTO {
	switch typed := value.(type) {
	case map[string]any:
		if rawTasks, ok := typed["tasks"]; ok {
			if tasks, ok := extractTickTickTasksFromArray(rawTasks); ok {
				return tasks
			}
		}

		for _, nested := range typed {
			tasks := extractTickTickTasks(nested)
			if len(tasks) > 0 {
				return tasks
			}
		}

		return nil
	case []any:
		for _, nested := range typed {
			tasks := extractTickTickTasks(nested)
			if len(tasks) > 0 {
				return tasks
			}
		}

		return nil
	default:
		return nil
	}
}

func extractTickTickTasksFromArray(value any) ([]tickTickTaskDTO, bool) {
	array, ok := value.([]any)
	if !ok {
		return nil, false
	}

	tasks := make([]tickTickTaskDTO, 0, len(array))
	for _, item := range array {
		entry, ok := item.(map[string]any)
		if !ok {
			continue
		}

		id := mapStringValue(entry, "id")
		title := mapStringValue(entry, "title")
		if strings.TrimSpace(id) == "" && strings.TrimSpace(title) == "" {
			continue
		}

		tasks = append(tasks, tickTickTaskDTO{
			ID:      id,
			Title:   title,
			DueDate: mapStringValue(entry, "dueDate"),
			Status:  mapIntValue(entry, "status"),
		})
	}

	return tasks, true
}

func mapStringValue(entry map[string]any, key string) string {
	value, ok := entry[key]
	if !ok || value == nil {
		return ""
	}

	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed)
	case float64:
		if typed == float64(int64(typed)) {
			return strconv.FormatInt(int64(typed), 10)
		}
		return strconv.FormatFloat(typed, 'f', -1, 64)
	default:
		return ""
	}
}

func mapIntValue(entry map[string]any, key string) int {
	value, ok := entry[key]
	if !ok || value == nil {
		return 0
	}

	switch typed := value.(type) {
	case float64:
		return int(typed)
	case int:
		return typed
	case int64:
		return int(typed)
	case string:
		parsed, err := strconv.Atoi(strings.TrimSpace(typed))
		if err == nil {
			return parsed
		}
		return 0
	default:
		return 0
	}
}

func truncateTickTickBodyPreview(body []byte, maxLength int) string {
	preview := strings.TrimSpace(string(body))
	if len(preview) <= maxLength {
		return preview
	}

	if maxLength <= 3 {
		return preview[:maxLength]
	}

	return preview[:maxLength-3] + "..."
}
