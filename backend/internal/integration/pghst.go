package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"hoel-app/backend/internal/db"
)

type PGHService struct {
	client       *Client
	repository   *db.PGHRepository
	endpoint     string
	pollInterval time.Duration
}

type pghPayload struct {
	NextPickupDate    string `json:"next_pickup_date"`
	NextRecyclingDate string `json:"next_recycling_date"`
	PickupDate        string `json:"pickup_date"`
	RecyclingDate     string `json:"recycling_date"`
	NextCollection    string `json:"next_collection_date"`
	CollectionType    string `json:"collection_type"`
	UpdatedAt         string `json:"updated_at"`
}

func NewPGHService(client *Client, repository *db.PGHRepository, endpoint string, pollInterval time.Duration) *PGHService {
	if pollInterval <= 0 {
		pollInterval = 12 * time.Hour
	}

	return &PGHService{
		client:       client,
		repository:   repository,
		endpoint:     strings.TrimSpace(endpoint),
		pollInterval: pollInterval,
	}
}

func (s *PGHService) Enabled() bool {
	return s.client != nil && s.repository != nil && s.endpoint != ""
}

func (s *PGHService) Start(ctx context.Context) {
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

func (s *PGHService) SyncOnce(ctx context.Context) error {
	if !s.Enabled() {
		return nil
	}

	response, err := s.client.Do(ctx, Request{
		Service: "pghst",
		Method:  http.MethodGet,
		URL:     s.endpoint,
		Headers: map[string]string{"Accept": "application/json"},
	})
	if err != nil {
		return err
	}

	payload, err := parsePGHPayload(response.Body)
	if err != nil {
		if s.client.monitoring != nil {
			statusCode := response.StatusCode
			_ = s.client.monitoring.RecordIntegrationFailure(ctx, "pghst", s.endpoint, &statusCode, err.Error(), time.Now())
		}
		return err
	}

	schedule, err := s.toSchedule(payload)
	if err != nil {
		return err
	}

	if err := s.repository.UpsertSchedule(ctx, schedule); err != nil {
		return err
	}

	return nil
}

func parsePGHPayload(body []byte) (pghPayload, error) {
	var payload pghPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return pghPayload{}, fmt.Errorf("decode pgh payload: %w", err)
	}

	return payload, nil
}

func (s *PGHService) toSchedule(payload pghPayload) (db.PGHSchedule, error) {
	nextPickupDate, err := parseBestDate(
		payload.NextPickupDate,
		payload.PickupDate,
		payload.NextCollection,
	)
	if err != nil {
		return db.PGHSchedule{}, fmt.Errorf("parse next pickup date: %w", err)
	}

	nextRecyclingDate, err := parseOptionalBestDate(payload.NextRecyclingDate, payload.RecyclingDate)
	if err != nil {
		return db.PGHSchedule{}, fmt.Errorf("parse next recycling date: %w", err)
	}

	sourceUpdatedAt, err := parseOptionalBestDate(payload.UpdatedAt)
	if err != nil {
		return db.PGHSchedule{}, fmt.Errorf("parse pgh updated date: %w", err)
	}

	isRecyclingWeek := false
	if strings.EqualFold(strings.TrimSpace(payload.CollectionType), "recycling") {
		isRecyclingWeek = true
	} else if nextRecyclingDate != nil && sameCalendarDay(nextPickupDate, *nextRecyclingDate) {
		isRecyclingWeek = true
	}

	now := time.Now().UTC()
	diff := nextPickupDate.Sub(now)
	showIndicator := diff >= 0 && diff <= 24*time.Hour

	return db.PGHSchedule{
		NextPickupDate:    nextPickupDate,
		NextRecyclingDate: nextRecyclingDate,
		IsRecyclingWeek:   isRecyclingWeek,
		ShowIndicator:     showIndicator,
		SourceUpdatedAt:   sourceUpdatedAt,
	}, nil
}

func parseBestDate(values ...string) (time.Time, error) {
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}

		parsed, err := parseFlexibleDate(trimmed)
		if err != nil {
			return time.Time{}, err
		}
		return parsed, nil
	}

	return time.Time{}, fmt.Errorf("date is missing")
}

func parseOptionalBestDate(values ...string) (*time.Time, error) {
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}

		parsed, err := parseFlexibleDate(trimmed)
		if err != nil {
			return nil, err
		}

		return &parsed, nil
	}

	return nil, nil
}

func parseFlexibleDate(value string) (time.Time, error) {
	layouts := []string{
		time.RFC3339,
		"2006-01-02",
		"01/02/2006",
		"2006-01-02 15:04:05",
	}

	for _, layout := range layouts {
		if parsed, err := time.Parse(layout, value); err == nil {
			return parsed.UTC(), nil
		}
	}

	return time.Time{}, fmt.Errorf("unsupported date format %q", value)
}

func sameCalendarDay(a, b time.Time) bool {
	aUTC := a.UTC()
	bUTC := b.UTC()

	return aUTC.Year() == bUTC.Year() && aUTC.Month() == bUTC.Month() && aUTC.Day() == bUTC.Day()
}
