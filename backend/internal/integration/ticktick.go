package integration

import (
	"context"
	"encoding/json"
	"fmt"
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
		return
	}

	_ = s.SyncOnce(ctx)

	ticker := time.NewTicker(s.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			_ = s.SyncOnce(ctx)
		}
	}
}

func (s *TickTickService) SyncOnce(ctx context.Context) error {
	if !s.Enabled() {
		return nil
	}

	requestURL := fmt.Sprintf("%s/project/%s/data", strings.TrimRight(s.apiRoot, "/"), url.PathEscape(s.projectID))

	accessToken, err := s.oauth.ResolveAccessToken(ctx)
	if err != nil {
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
		return err
	}

	var payload tickTickProjectDataResponse
	if err := json.Unmarshal(response.Body, &payload); err != nil {
		return fmt.Errorf("decode ticktick response: %w", err)
	}

	tasks := make([]db.TickTickTask, 0, len(payload.Tasks))
	for _, task := range payload.Tasks {
		dueAt, ok, err := parseTickTickDueDate(task.DueDate)
		if err != nil {
			return fmt.Errorf("parse ticktick due date for task %s: %w", task.ID, err)
		}
		if !ok {
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
		return err
	}

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
