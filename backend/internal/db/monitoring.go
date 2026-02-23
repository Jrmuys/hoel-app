package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type IntegrationStatus struct {
	Service             string
	LastSuccessAt       *time.Time
	LastErrorAt         *time.Time
	ConsecutiveFailures int
	UpdatedAt           time.Time
}

type APIError struct {
	ID          int64
	ServiceName string
	Endpoint    string
	HTTPStatus  *int
	Message     string
	CreatedAt   time.Time
	Resolved    bool
}

type MonitoringRepository struct {
	database *sql.DB
}

func NewMonitoringRepository(database *sql.DB) *MonitoringRepository {
	return &MonitoringRepository{database: database}
}

func (r *MonitoringRepository) RecordIntegrationSuccess(ctx context.Context, service string, occurredAt time.Time) error {
	if service == "" {
		return fmt.Errorf("service is required")
	}

	const statement = `
	INSERT INTO integration_status (service_name, last_success_at, consecutive_failures, updated_at)
	VALUES (?, ?, 0, CURRENT_TIMESTAMP)
	ON CONFLICT(service_name) DO UPDATE SET
		last_success_at = excluded.last_success_at,
		consecutive_failures = 0,
		updated_at = CURRENT_TIMESTAMP;`

	if _, err := r.database.ExecContext(ctx, statement, service, occurredAt.UTC().Format(time.RFC3339)); err != nil {
		return fmt.Errorf("record integration success for %s: %w", service, err)
	}

	return nil
}

func (r *MonitoringRepository) RecordIntegrationFailure(ctx context.Context, service, endpoint string, httpStatus *int, message string, occurredAt time.Time) error {
	if service == "" {
		return fmt.Errorf("service is required")
	}
	if endpoint == "" {
		return fmt.Errorf("endpoint is required")
	}

	transaction, err := r.database.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin failure transaction: %w", err)
	}

	if err := r.insertAPIError(ctx, transaction, service, endpoint, httpStatus, message, occurredAt); err != nil {
		_ = transaction.Rollback()
		return err
	}

	const statusStatement = `
	INSERT INTO integration_status (service_name, last_error_at, consecutive_failures, updated_at)
	VALUES (?, ?, 1, CURRENT_TIMESTAMP)
	ON CONFLICT(service_name) DO UPDATE SET
		last_error_at = excluded.last_error_at,
		consecutive_failures = integration_status.consecutive_failures + 1,
		updated_at = CURRENT_TIMESTAMP;`

	if _, err := transaction.ExecContext(ctx, statusStatement, service, occurredAt.UTC().Format(time.RFC3339)); err != nil {
		_ = transaction.Rollback()
		return fmt.Errorf("update integration failure status for %s: %w", service, err)
	}

	if err := transaction.Commit(); err != nil {
		return fmt.Errorf("commit failure transaction: %w", err)
	}

	return nil
}

func (r *MonitoringRepository) insertAPIError(ctx context.Context, transaction *sql.Tx, service, endpoint string, httpStatus *int, message string, occurredAt time.Time) error {
	const statement = `
	INSERT INTO api_errors (service_name, endpoint, http_status, error_message, created_at, resolved)
	VALUES (?, ?, ?, ?, ?, 0);`

	var statusValue any
	if httpStatus != nil {
		statusValue = *httpStatus
	}

	if _, err := transaction.ExecContext(ctx, statement, service, endpoint, statusValue, message, occurredAt.UTC().Format(time.RFC3339)); err != nil {
		return fmt.Errorf("insert api error for %s: %w", service, err)
	}

	return nil
}

func (r *MonitoringRepository) ListIntegrationStatus(ctx context.Context) ([]IntegrationStatus, error) {
	const query = `
	SELECT service_name, last_success_at, last_error_at, consecutive_failures, updated_at
	FROM integration_status
	ORDER BY service_name ASC;`

	rows, err := r.database.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list integration status: %w", err)
	}
	defer rows.Close()

	statuses := make([]IntegrationStatus, 0)
	for rows.Next() {
		status, err := scanIntegrationStatus(rows)
		if err != nil {
			return nil, err
		}
		statuses = append(statuses, status)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate integration status rows: %w", err)
	}

	return statuses, nil
}

func scanIntegrationStatus(row scanner) (IntegrationStatus, error) {
	var (
		serviceName         string
		lastSuccessRaw      sql.NullString
		lastErrorRaw        sql.NullString
		consecutiveFailures int
		updatedAtRaw        string
	)

	if err := row.Scan(&serviceName, &lastSuccessRaw, &lastErrorRaw, &consecutiveFailures, &updatedAtRaw); err != nil {
		return IntegrationStatus{}, fmt.Errorf("scan integration status row: %w", err)
	}

	updatedAt, err := parseTimestamp(updatedAtRaw)
	if err != nil {
		return IntegrationStatus{}, fmt.Errorf("parse integration updated_at: %w", err)
	}

	lastSuccessAt, err := parseNullableTimestamp(lastSuccessRaw)
	if err != nil {
		return IntegrationStatus{}, fmt.Errorf("parse integration last_success_at: %w", err)
	}

	lastErrorAt, err := parseNullableTimestamp(lastErrorRaw)
	if err != nil {
		return IntegrationStatus{}, fmt.Errorf("parse integration last_error_at: %w", err)
	}

	return IntegrationStatus{
		Service:             serviceName,
		LastSuccessAt:       lastSuccessAt,
		LastErrorAt:         lastErrorAt,
		ConsecutiveFailures: consecutiveFailures,
		UpdatedAt:           updatedAt,
	}, nil
}

func (r *MonitoringRepository) ListRecentUnresolvedErrors(ctx context.Context, limit int) ([]APIError, error) {
	if limit <= 0 {
		limit = 5
	}

	const query = `
	SELECT id, service_name, endpoint, http_status, error_message, created_at, resolved
	FROM api_errors
	WHERE resolved = 0
	ORDER BY created_at DESC
	LIMIT ?;`

	rows, err := r.database.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("list unresolved api errors: %w", err)
	}
	defer rows.Close()

	errorsList := make([]APIError, 0, limit)
	for rows.Next() {
		record, err := scanAPIError(rows)
		if err != nil {
			return nil, err
		}
		errorsList = append(errorsList, record)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate api error rows: %w", err)
	}

	return errorsList, nil
}

func scanAPIError(row scanner) (APIError, error) {
	var (
		id          int64
		serviceName string
		endpoint    string
		httpStatus  sql.NullInt64
		message     sql.NullString
		createdAtRaw string
		resolved    bool
	)

	if err := row.Scan(&id, &serviceName, &endpoint, &httpStatus, &message, &createdAtRaw, &resolved); err != nil {
		return APIError{}, fmt.Errorf("scan api error row: %w", err)
	}

	createdAt, err := parseTimestamp(createdAtRaw)
	if err != nil {
		return APIError{}, fmt.Errorf("parse api error created_at: %w", err)
	}

	var statusPointer *int
	if httpStatus.Valid {
		statusValue := int(httpStatus.Int64)
		statusPointer = &statusValue
	}

	return APIError{
		ID:          id,
		ServiceName: serviceName,
		Endpoint:    endpoint,
		HTTPStatus:  statusPointer,
		Message:     message.String,
		CreatedAt:   createdAt,
		Resolved:    resolved,
	}, nil
}

type scanner interface {
	Scan(dest ...any) error
}

func parseNullableTimestamp(value sql.NullString) (*time.Time, error) {
	if !value.Valid || value.String == "" {
		return nil, nil
	}

	parsed, err := parseTimestamp(value.String)
	if err != nil {
		return nil, err
	}

	return &parsed, nil
}

func parseTimestamp(value string) (time.Time, error) {
	if parsed, err := time.Parse(time.RFC3339, value); err == nil {
		return parsed, nil
	}

	const sqliteDateTimeFormat = "2006-01-02 15:04:05"
	parsed, err := time.ParseInLocation(sqliteDateTimeFormat, value, time.UTC)
	if err != nil {
		return time.Time{}, fmt.Errorf("unsupported timestamp format %q: %w", value, err)
	}

	return parsed, nil
}
