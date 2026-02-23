package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type TickTickOAuthToken struct {
	AccessToken  string
	RefreshToken string
	TokenType    string
	Scope        string
	ExpiresAt    *time.Time
	UpdatedAt    time.Time
}

type TickTickOAuthRepository struct {
	database *sql.DB
}

func NewTickTickOAuthRepository(database *sql.DB) *TickTickOAuthRepository {
	return &TickTickOAuthRepository{database: database}
}

func (r *TickTickOAuthRepository) SaveToken(ctx context.Context, token TickTickOAuthToken) error {
	if token.AccessToken == "" {
		return fmt.Errorf("access token is required")
	}

	const statement = `
	INSERT INTO ticktick_oauth_token (singleton_id, access_token, refresh_token, token_type, scope, expires_at, updated_at)
	VALUES (1, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	ON CONFLICT(singleton_id) DO UPDATE SET
		access_token = excluded.access_token,
		refresh_token = excluded.refresh_token,
		token_type = excluded.token_type,
		scope = excluded.scope,
		expires_at = excluded.expires_at,
		updated_at = CURRENT_TIMESTAMP;`

	var expiresAt any
	if token.ExpiresAt != nil {
		expiresAt = token.ExpiresAt.UTC().Format(time.RFC3339)
	}

	if _, err := r.database.ExecContext(
		ctx,
		statement,
		token.AccessToken,
		token.RefreshToken,
		token.TokenType,
		token.Scope,
		expiresAt,
	); err != nil {
		return fmt.Errorf("save ticktick oauth token: %w", err)
	}

	return nil
}

func (r *TickTickOAuthRepository) GetToken(ctx context.Context) (*TickTickOAuthToken, error) {
	const query = `
	SELECT access_token, refresh_token, token_type, scope, expires_at, updated_at
	FROM ticktick_oauth_token
	WHERE singleton_id = 1;`

	var (
		accessToken  string
		refreshToken sql.NullString
		tokenType    sql.NullString
		scope        sql.NullString
		expiresAtRaw sql.NullString
		updatedAtRaw string
	)

	err := r.database.QueryRowContext(ctx, query).Scan(
		&accessToken,
		&refreshToken,
		&tokenType,
		&scope,
		&expiresAtRaw,
		&updatedAtRaw,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("query ticktick oauth token: %w", err)
	}

	updatedAt, err := parseTimestamp(updatedAtRaw)
	if err != nil {
		return nil, fmt.Errorf("parse ticktick oauth updated_at: %w", err)
	}

	expiresAt, err := parseNullableTimestamp(expiresAtRaw)
	if err != nil {
		return nil, fmt.Errorf("parse ticktick oauth expires_at: %w", err)
	}

	return &TickTickOAuthToken{
		AccessToken:  accessToken,
		RefreshToken: refreshToken.String,
		TokenType:    tokenType.String,
		Scope:        scope.String,
		ExpiresAt:    expiresAt,
		UpdatedAt:    updatedAt,
	}, nil
}
