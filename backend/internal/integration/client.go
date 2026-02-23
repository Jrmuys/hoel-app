package integration

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"hoel-app/backend/internal/db"
)

type Client struct {
	httpClient    *http.Client
	monitoring    *db.MonitoringRepository
	retryCount    int
	retryBackoff  time.Duration
	requestTimout time.Duration
}

type Request struct {
	Service string
	Method  string
	URL     string
	Body    io.Reader
	Headers map[string]string
}

type Response struct {
	StatusCode int
	Headers    http.Header
	Body       []byte
}

type HTTPStatusError struct {
	Service    string
	URL        string
	StatusCode int
	Message    string
}

func (e HTTPStatusError) Error() string {
	if e.Message == "" {
		return fmt.Sprintf("%s request to %s failed with status %d", e.Service, e.URL, e.StatusCode)
	}

	return fmt.Sprintf("%s request to %s failed with status %d: %s", e.Service, e.URL, e.StatusCode, e.Message)
}

func NewClient(timeout time.Duration, retryCount int, retryBackoff time.Duration, monitoring *db.MonitoringRepository) *Client {
	if timeout <= 0 {
		timeout = 8 * time.Second
	}
	if retryCount < 0 {
		retryCount = 0
	}
	if retryBackoff <= 0 {
		retryBackoff = 300 * time.Millisecond
	}

	return &Client{
		httpClient:    &http.Client{Timeout: timeout},
		monitoring:    monitoring,
		retryCount:    retryCount,
		retryBackoff:  retryBackoff,
		requestTimout: timeout,
	}
}

func (c *Client) Do(ctx context.Context, request Request) (Response, error) {
	if err := validateRequest(request); err != nil {
		return Response{}, err
	}

	maxAttempts := c.retryCount + 1
	attempt := 0
	for attempt < maxAttempts {
		attempt++
		response, statusCode, err := c.executeOnce(ctx, request)
		if err == nil {
			if c.monitoring != nil {
				_ = c.monitoring.RecordIntegrationSuccess(ctx, request.Service, time.Now())
			}
			return response, nil
		}

		if c.shouldRetry(statusCode, err, attempt, maxAttempts) {
			if waitErr := c.waitBeforeRetry(ctx, attempt); waitErr != nil {
				return Response{}, waitErr
			}
			continue
		}

		c.recordFailure(ctx, request, statusCode, err)
		return Response{}, err
	}

	return Response{}, fmt.Errorf("request failed after %d attempts", maxAttempts)
}

func validateRequest(request Request) error {
	if strings.TrimSpace(request.Service) == "" {
		return fmt.Errorf("service is required")
	}
	if strings.TrimSpace(request.Method) == "" {
		return fmt.Errorf("method is required")
	}
	if strings.TrimSpace(request.URL) == "" {
		return fmt.Errorf("url is required")
	}

	return nil
}

func (c *Client) executeOnce(ctx context.Context, request Request) (Response, *int, error) {
	httpRequest, err := http.NewRequestWithContext(ctx, request.Method, request.URL, request.Body)
	if err != nil {
		return Response{}, nil, fmt.Errorf("build request: %w", err)
	}

	for key, value := range request.Headers {
		httpRequest.Header.Set(key, value)
	}

	httpResponse, err := c.httpClient.Do(httpRequest)
	if err != nil {
		return Response{}, nil, fmt.Errorf("send request: %w", err)
	}
	defer httpResponse.Body.Close()

	body, err := io.ReadAll(httpResponse.Body)
	if err != nil {
		statusCode := httpResponse.StatusCode
		return Response{}, &statusCode, fmt.Errorf("read response body: %w", err)
	}

	if httpResponse.StatusCode < http.StatusOK || httpResponse.StatusCode >= http.StatusMultipleChoices {
		statusCode := httpResponse.StatusCode
		message := truncateMessage(string(body), 300)
		return Response{}, &statusCode, HTTPStatusError{
			Service:    request.Service,
			URL:        request.URL,
			StatusCode: httpResponse.StatusCode,
			Message:    message,
		}
	}

	return Response{
		StatusCode: httpResponse.StatusCode,
		Headers:    httpResponse.Header,
		Body:       body,
	}, nil, nil
}

func (c *Client) shouldRetry(statusCode *int, err error, attempt, maxAttempts int) bool {
	if attempt >= maxAttempts {
		return false
	}

	if statusCode != nil {
		return retryableStatusCode(*statusCode)
	}

	return retryableTransportError(err)
}

func (c *Client) waitBeforeRetry(ctx context.Context, attempt int) error {
	waitFor := c.retryBackoff * time.Duration(attempt)
	timer := time.NewTimer(waitFor)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func retryableStatusCode(statusCode int) bool {
	if statusCode == http.StatusRequestTimeout || statusCode == http.StatusTooManyRequests {
		return true
	}

	return statusCode >= http.StatusInternalServerError
}

func retryableTransportError(err error) bool {
	if err == nil {
		return false
	}

	var netErr net.Error
	if errors.As(err, &netErr) {
		return netErr.Timeout() || netErr.Temporary()
	}

	return strings.Contains(strings.ToLower(err.Error()), "connection reset") || strings.Contains(strings.ToLower(err.Error()), "connection refused")
}

func (c *Client) recordFailure(ctx context.Context, request Request, statusCode *int, err error) {
	if c.monitoring == nil || err == nil {
		return
	}

	_ = c.monitoring.RecordIntegrationFailure(ctx, request.Service, request.URL, statusCode, truncateMessage(err.Error(), 500), time.Now())
}

func truncateMessage(message string, maxLength int) string {
	trimmed := strings.TrimSpace(message)
	if len(trimmed) <= maxLength {
		return trimmed
	}

	if maxLength <= 3 {
		return trimmed[:maxLength]
	}

	return trimmed[:maxLength-3] + "..."
}
