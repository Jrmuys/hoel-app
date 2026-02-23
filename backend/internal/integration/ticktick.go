package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
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

	log.Printf("ticktick project data received: project=%s tasks=%d", s.projectID, len(payload.Tasks))

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
