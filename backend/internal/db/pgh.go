package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type PGHSchedule struct {
	NextPickupDate    time.Time
	NextRecyclingDate *time.Time
	IsRecyclingWeek   bool
	ShowIndicator     bool
	SourceUpdatedAt   *time.Time
	UpdatedAt         time.Time
}

type PGHRepository struct {
	database *sql.DB
}

func NewPGHRepository(database *sql.DB) *PGHRepository {
	return &PGHRepository{database: database}
}

func (r *PGHRepository) UpsertSchedule(ctx context.Context, schedule PGHSchedule) error {
	const statement = `
	INSERT INTO pgh_schedule_cache (
		singleton_id,
		next_pickup_date,
		next_recycling_date,
		is_recycling_week,
		show_indicator,
		source_updated_at,
		updated_at
	)
	VALUES (1, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	ON CONFLICT(singleton_id) DO UPDATE SET
		next_pickup_date = excluded.next_pickup_date,
		next_recycling_date = excluded.next_recycling_date,
		is_recycling_week = excluded.is_recycling_week,
		show_indicator = excluded.show_indicator,
		source_updated_at = excluded.source_updated_at,
		updated_at = CURRENT_TIMESTAMP;`

	var recyclingValue any
	if schedule.NextRecyclingDate != nil {
		recyclingValue = schedule.NextRecyclingDate.UTC().Format(time.RFC3339)
	}

	var sourceUpdatedValue any
	if schedule.SourceUpdatedAt != nil {
		sourceUpdatedValue = schedule.SourceUpdatedAt.UTC().Format(time.RFC3339)
	}

	if _, err := r.database.ExecContext(
		ctx,
		statement,
		schedule.NextPickupDate.UTC().Format(time.RFC3339),
		recyclingValue,
		schedule.IsRecyclingWeek,
		schedule.ShowIndicator,
		sourceUpdatedValue,
	); err != nil {
		return fmt.Errorf("upsert pgh schedule cache: %w", err)
	}

	return nil
}

func (r *PGHRepository) GetLatestSchedule(ctx context.Context) (*PGHSchedule, error) {
	const query = `
	SELECT next_pickup_date, next_recycling_date, is_recycling_week, show_indicator, source_updated_at, updated_at
	FROM pgh_schedule_cache
	WHERE singleton_id = 1;`

	var (
		nextPickupRaw    string
		nextRecyclingRaw sql.NullString
		isRecyclingWeek  bool
		showIndicator    bool
		sourceUpdatedRaw sql.NullString
		updatedAtRaw     string
	)

	err := r.database.QueryRowContext(ctx, query).Scan(
		&nextPickupRaw,
		&nextRecyclingRaw,
		&isRecyclingWeek,
		&showIndicator,
		&sourceUpdatedRaw,
		&updatedAtRaw,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("query latest pgh schedule cache: %w", err)
	}

	nextPickupDate, err := parseTimestamp(nextPickupRaw)
	if err != nil {
		return nil, fmt.Errorf("parse next pickup date: %w", err)
	}

	nextRecyclingDate, err := parseNullableTimestamp(nextRecyclingRaw)
	if err != nil {
		return nil, fmt.Errorf("parse next recycling date: %w", err)
	}

	sourceUpdatedAt, err := parseNullableTimestamp(sourceUpdatedRaw)
	if err != nil {
		return nil, fmt.Errorf("parse source updated time: %w", err)
	}

	updatedAt, err := parseTimestamp(updatedAtRaw)
	if err != nil {
		return nil, fmt.Errorf("parse cache updated time: %w", err)
	}

	return &PGHSchedule{
		NextPickupDate:    nextPickupDate,
		NextRecyclingDate: nextRecyclingDate,
		IsRecyclingWeek:   isRecyclingWeek,
		ShowIndicator:     showIndicator,
		SourceUpdatedAt:   sourceUpdatedAt,
		UpdatedAt:         updatedAt,
	}, nil
}
