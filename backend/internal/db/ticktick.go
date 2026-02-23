package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type TickTickTask struct {
	ID            string
	Title         string
	DueAt         time.Time
	HasTime       bool
	Tags          []string
	Completed     bool
	SourceProject string
	UpdatedAt     time.Time
}

type TickTickRepository struct {
	database *sql.DB
}

func NewTickTickRepository(database *sql.DB) *TickTickRepository {
	return &TickTickRepository{database: database}
}

func (r *TickTickRepository) ReplaceTasks(ctx context.Context, tasks []TickTickTask) error {
	transaction, err := r.database.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin ticktick replace transaction: %w", err)
	}

	if _, err := transaction.ExecContext(ctx, `DELETE FROM ticktick_task_cache;`); err != nil {
		_ = transaction.Rollback()
		return fmt.Errorf("clear ticktick task cache: %w", err)
	}

	if len(tasks) > 0 {
		statement, err := transaction.PrepareContext(ctx, `
			INSERT INTO ticktick_task_cache (task_id, title, due_at, has_time, tags_json, completed, source_project, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP);`)
		if err != nil {
			_ = transaction.Rollback()
			return fmt.Errorf("prepare ticktick insert statement: %w", err)
		}

		for _, task := range tasks {
			tagsJSON, err := encodeTaskTags(task.Tags)
			if err != nil {
				_ = statement.Close()
				_ = transaction.Rollback()
				return fmt.Errorf("encode tags for ticktick task %s: %w", task.ID, err)
			}

			if _, err := statement.ExecContext(
				ctx,
				task.ID,
				task.Title,
				task.DueAt.UTC().Format(time.RFC3339),
				task.HasTime,
				tagsJSON,
				task.Completed,
				task.SourceProject,
			); err != nil {
				_ = statement.Close()
				_ = transaction.Rollback()
				return fmt.Errorf("insert ticktick task %s: %w", task.ID, err)
			}
		}

		if err := statement.Close(); err != nil {
			_ = transaction.Rollback()
			return fmt.Errorf("close ticktick insert statement: %w", err)
		}
	}

	if err := transaction.Commit(); err != nil {
		return fmt.Errorf("commit ticktick replace transaction: %w", err)
	}

	return nil
}

func (r *TickTickRepository) ListTasksDueBetween(ctx context.Context, startInclusive, endInclusive time.Time) ([]TickTickTask, error) {
	const query = `
	SELECT task_id, title, due_at, has_time, tags_json, completed, source_project, updated_at
	FROM ticktick_task_cache
	WHERE completed = 0
	  AND due_at >= ?
	  AND due_at <= ?
	ORDER BY due_at ASC;`

	rows, err := r.database.QueryContext(
		ctx,
		query,
		startInclusive.UTC().Format(time.RFC3339),
		endInclusive.UTC().Format(time.RFC3339),
	)
	if err != nil {
		return nil, fmt.Errorf("query ticktick tasks due between window: %w", err)
	}
	defer rows.Close()

	return scanTickTickTasks(rows)
}

func (r *TickTickRepository) ListIncompleteTasks(ctx context.Context) ([]TickTickTask, error) {
	const query = `
	SELECT task_id, title, due_at, has_time, tags_json, completed, source_project, updated_at
	FROM ticktick_task_cache
	WHERE completed = 0
	ORDER BY due_at ASC;`

	rows, err := r.database.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query incomplete ticktick tasks: %w", err)
	}
	defer rows.Close()

	return scanTickTickTasks(rows)
}

func (r *TickTickRepository) MarkTaskCompleted(ctx context.Context, taskID string) error {
	if taskID == "" {
		return fmt.Errorf("task id is required")
	}

	const statement = `
	UPDATE ticktick_task_cache
	SET completed = 1,
	    updated_at = CURRENT_TIMESTAMP
	WHERE task_id = ?;`

	if _, err := r.database.ExecContext(ctx, statement, taskID); err != nil {
		return fmt.Errorf("mark ticktick task completed: %w", err)
	}

	return nil
}

func scanTickTickTasks(rows *sql.Rows) ([]TickTickTask, error) {
	tasks := make([]TickTickTask, 0)
	for rows.Next() {
		var (
			id            string
			title         string
			dueAtRaw      string
			hasTime       bool
			tagsRaw       string
			completed     bool
			sourceProject string
			updatedAtRaw  string
		)

		if err := rows.Scan(&id, &title, &dueAtRaw, &hasTime, &tagsRaw, &completed, &sourceProject, &updatedAtRaw); err != nil {
			return nil, fmt.Errorf("scan ticktick task: %w", err)
		}

		dueAt, err := parseTimestamp(dueAtRaw)
		if err != nil {
			return nil, fmt.Errorf("parse ticktick due_at: %w", err)
		}

		updatedAt, err := parseTimestamp(updatedAtRaw)
		if err != nil {
			return nil, fmt.Errorf("parse ticktick updated_at: %w", err)
		}

		tags, err := decodeTaskTags(tagsRaw)
		if err != nil {
			return nil, fmt.Errorf("parse ticktick task tags: %w", err)
		}

		tasks = append(tasks, TickTickTask{
			ID:            id,
			Title:         title,
			DueAt:         dueAt,
			HasTime:       hasTime,
			Tags:          tags,
			Completed:     completed,
			SourceProject: sourceProject,
			UpdatedAt:     updatedAt,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate ticktick task rows: %w", err)
	}

	return tasks, nil
}

func encodeTaskTags(tags []string) (string, error) {
	normalized := normalizeTaskTags(tags)
	encoded, err := json.Marshal(normalized)
	if err != nil {
		return "", err
	}

	return string(encoded), nil
}

func decodeTaskTags(raw string) ([]string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return []string{}, nil
	}

	var tags []string
	if err := json.Unmarshal([]byte(trimmed), &tags); err != nil {
		return nil, err
	}

	return normalizeTaskTags(tags), nil
}

func normalizeTaskTags(tags []string) []string {
	if len(tags) == 0 {
		return []string{}
	}

	seen := map[string]struct{}{}
	normalized := make([]string, 0, len(tags))
	for _, tag := range tags {
		trimmed := strings.ToLower(strings.TrimSpace(tag))
		trimmed = strings.TrimPrefix(trimmed, "#")
		if trimmed == "" {
			continue
		}

		if _, exists := seen[trimmed]; exists {
			continue
		}

		seen[trimmed] = struct{}{}
		normalized = append(normalized, trimmed)
	}

	return normalized
}
