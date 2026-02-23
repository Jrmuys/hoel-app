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

type TickTickOAuthService struct {
	client       *Client
	repository   *db.TickTickOAuthRepository
	authorizeURL string
	tokenURL     string
	clientID     string
	clientSecret string
	redirectURI  string
	staticToken  string
}

type tickTickTokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	Scope        string `json:"scope"`
}

func NewTickTickOAuthService(
	client *Client,
	repository *db.TickTickOAuthRepository,
	authorizeURL, tokenURL, clientID, clientSecret, redirectURI, staticToken string,
) *TickTickOAuthService {
	return &TickTickOAuthService{
		client:       client,
		repository:   repository,
		authorizeURL: strings.TrimSpace(authorizeURL),
		tokenURL:     strings.TrimSpace(tokenURL),
		clientID:     strings.TrimSpace(clientID),
		clientSecret: strings.TrimSpace(clientSecret),
		redirectURI:  strings.TrimSpace(redirectURI),
		staticToken:  strings.TrimSpace(staticToken),
	}
}

func (s *TickTickOAuthService) OAuthEnabled() bool {
	return s.client != nil && s.repository != nil && s.authorizeURL != "" && s.tokenURL != "" && s.clientID != "" && s.clientSecret != "" && s.redirectURI != ""
}

func (s *TickTickOAuthService) BuildAuthorizeURL(state string) (string, error) {
	if !s.OAuthEnabled() {
		return "", fmt.Errorf("ticktick oauth is not configured")
	}

	parsed, err := url.Parse(s.authorizeURL)
	if err != nil {
		return "", fmt.Errorf("parse ticktick authorize url: %w", err)
	}

	query := parsed.Query()
	query.Set("client_id", s.clientID)
	query.Set("scope", "tasks:read")
	query.Set("state", state)
	query.Set("redirect_uri", s.redirectURI)
	query.Set("response_type", "code")
	parsed.RawQuery = query.Encode()

	return parsed.String(), nil
}

func (s *TickTickOAuthService) ExchangeCode(ctx context.Context, code string) (*db.TickTickOAuthToken, error) {
	if !s.OAuthEnabled() {
		return nil, fmt.Errorf("ticktick oauth is not configured")
	}

	values := url.Values{}
	values.Set("client_id", s.clientID)
	values.Set("client_secret", s.clientSecret)
	values.Set("code", strings.TrimSpace(code))
	values.Set("grant_type", "authorization_code")
	values.Set("redirect_uri", s.redirectURI)

	response, err := s.client.Do(ctx, Request{
		Service: "ticktick",
		Method:  http.MethodPost,
		URL:     s.tokenURL,
		Body:    strings.NewReader(values.Encode()),
		Headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
			"Accept":       "application/json",
		},
	})
	if err != nil {
		return nil, err
	}

	token, err := decodeTickTickToken(response.Body)
	if err != nil {
		return nil, err
	}

	if err := s.repository.SaveToken(ctx, token); err != nil {
		return nil, err
	}

	return &token, nil
}

func (s *TickTickOAuthService) ResolveAccessToken(ctx context.Context) (string, error) {
	if s.staticToken != "" {
		return s.staticToken, nil
	}

	if s.repository == nil {
		return "", fmt.Errorf("ticktick oauth repository is unavailable")
	}

	stored, err := s.repository.GetToken(ctx)
	if err != nil {
		return "", err
	}
	if stored == nil || strings.TrimSpace(stored.AccessToken) == "" {
		return "", fmt.Errorf("ticktick access token is not available")
	}

	if stored.ExpiresAt == nil || stored.ExpiresAt.After(time.Now().UTC().Add(60*time.Second)) {
		return stored.AccessToken, nil
	}

	if strings.TrimSpace(stored.RefreshToken) == "" {
		return "", fmt.Errorf("ticktick access token is expired and refresh token is missing")
	}

	refreshed, err := s.refreshAccessToken(ctx, stored.RefreshToken)
	if err != nil {
		return "", err
	}

	if err := s.repository.SaveToken(ctx, refreshed); err != nil {
		return "", err
	}

	return refreshed.AccessToken, nil
}

func (s *TickTickOAuthService) refreshAccessToken(ctx context.Context, refreshToken string) (db.TickTickOAuthToken, error) {
	if !s.OAuthEnabled() {
		return db.TickTickOAuthToken{}, fmt.Errorf("ticktick oauth is not configured")
	}

	values := url.Values{}
	values.Set("client_id", s.clientID)
	values.Set("client_secret", s.clientSecret)
	values.Set("refresh_token", strings.TrimSpace(refreshToken))
	values.Set("grant_type", "refresh_token")

	response, err := s.client.Do(ctx, Request{
		Service: "ticktick",
		Method:  http.MethodPost,
		URL:     s.tokenURL,
		Body:    strings.NewReader(values.Encode()),
		Headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
			"Accept":       "application/json",
		},
	})
	if err != nil {
		return db.TickTickOAuthToken{}, err
	}

	return decodeTickTickToken(response.Body)
}

func decodeTickTickToken(body []byte) (db.TickTickOAuthToken, error) {
	var payload tickTickTokenResponse
	if err := json.Unmarshal(body, &payload); err != nil {
		return db.TickTickOAuthToken{}, fmt.Errorf("decode ticktick token response: %w", err)
	}

	accessToken := strings.TrimSpace(payload.AccessToken)
	if accessToken == "" {
		return db.TickTickOAuthToken{}, fmt.Errorf("ticktick token response missing access_token")
	}

	var expiresAt *time.Time
	if payload.ExpiresIn > 0 {
		value := time.Now().UTC().Add(time.Duration(payload.ExpiresIn) * time.Second)
		expiresAt = &value
	}

	return db.TickTickOAuthToken{
		AccessToken:  accessToken,
		RefreshToken: strings.TrimSpace(payload.RefreshToken),
		TokenType:    strings.TrimSpace(payload.TokenType),
		Scope:        strings.TrimSpace(payload.Scope),
		ExpiresAt:    expiresAt,
	}, nil
}
