package integration

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"hoel-app/backend/internal/db"

	_ "modernc.org/sqlite"
)

func TestCompleteTask_UsesProjectEndpointFirst(t *testing.T) {
	database := newTickTickTestDatabase(t)
	repository := db.NewTickTickRepository(database)
	insertPendingTask(t, database, "task-123")

	requestedPaths := make([]string, 0, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestedPaths = append(requestedPaths, r.URL.Path)
		if r.Method == http.MethodPost && r.URL.Path == "/open/v1/project/project-1/task/task-123/complete" {
			w.WriteHeader(http.StatusOK)
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := NewClient(2*time.Second, 0, 10*time.Millisecond, nil)
	oauth := NewTickTickOAuthService(client, nil, "", "", "", "", "", "static-token")
	service := NewTickTickService(client, repository, oauth, server.URL+"/open/v1", "project-1", time.Minute)

	err := service.CompleteTask(context.Background(), "task-123")
	if err != nil {
		t.Fatalf("complete task returned error: %v", err)
	}

	if len(requestedPaths) != 1 {
		t.Fatalf("expected 1 completion request, got %d (%v)", len(requestedPaths), requestedPaths)
	}

	if requestedPaths[0] != "/open/v1/project/project-1/task/task-123/complete" {
		t.Fatalf("unexpected endpoint called: %s", requestedPaths[0])
	}

	assertTaskCompleted(t, database, "task-123", true)
}

func TestCompleteTask_FallsBackToLegacyEndpointOn404(t *testing.T) {
	database := newTickTickTestDatabase(t)
	repository := db.NewTickTickRepository(database)
	insertPendingTask(t, database, "task-456")

	requestedPaths := make([]string, 0, 2)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestedPaths = append(requestedPaths, r.URL.Path)
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		switch r.URL.Path {
		case "/open/v1/project/project-1/task/task-456/complete":
			w.WriteHeader(http.StatusNotFound)
		case "/open/v1/task/task-456/complete":
			w.WriteHeader(http.StatusOK)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := NewClient(2*time.Second, 0, 10*time.Millisecond, nil)
	oauth := NewTickTickOAuthService(client, nil, "", "", "", "", "", "static-token")
	service := NewTickTickService(client, repository, oauth, server.URL+"/open/v1", "project-1", time.Minute)

	err := service.CompleteTask(context.Background(), "task-456")
	if err != nil {
		t.Fatalf("complete task returned error: %v", err)
	}

	if len(requestedPaths) != 2 {
		t.Fatalf("expected 2 completion requests, got %d (%v)", len(requestedPaths), requestedPaths)
	}

	if requestedPaths[0] != "/open/v1/project/project-1/task/task-456/complete" {
		t.Fatalf("unexpected first endpoint: %s", requestedPaths[0])
	}
	if requestedPaths[1] != "/open/v1/task/task-456/complete" {
		t.Fatalf("unexpected fallback endpoint: %s", requestedPaths[1])
	}

	assertTaskCompleted(t, database, "task-456", true)
}

func TestCompleteTask_ReturnsErrorWhenAllEndpointsFail(t *testing.T) {
	database := newTickTickTestDatabase(t)
	repository := db.NewTickTickRepository(database)
	insertPendingTask(t, database, "task-789")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := NewClient(2*time.Second, 0, 10*time.Millisecond, nil)
	oauth := NewTickTickOAuthService(client, nil, "", "", "", "", "", "static-token")
	service := NewTickTickService(client, repository, oauth, server.URL+"/open/v1", "project-1", time.Minute)

	err := service.CompleteTask(context.Background(), "task-789")
	if err == nil {
		t.Fatal("expected error when all completion endpoints fail")
	}

	if !strings.Contains(err.Error(), "status 404") {
		t.Fatalf("expected 404 in error, got: %v", err)
	}

	assertTaskCompleted(t, database, "task-789", false)
}

func newTickTickTestDatabase(t *testing.T) *sql.DB {
	t.Helper()

	database, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite memory db: %v", err)
	}

	_, err = database.Exec(`
		CREATE TABLE ticktick_task_cache (
			task_id TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			due_at TEXT NOT NULL,
			completed INTEGER NOT NULL DEFAULT 0,
			source_project TEXT NOT NULL,
			updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err != nil {
		t.Fatalf("create ticktick_task_cache table: %v", err)
	}

	t.Cleanup(func() {
		_ = database.Close()
	})

	return database
}

func insertPendingTask(t *testing.T, database *sql.DB, taskID string) {
	t.Helper()

	_, err := database.Exec(`
		INSERT INTO ticktick_task_cache (task_id, title, due_at, completed, source_project, updated_at)
		VALUES (?, ?, ?, 0, ?, CURRENT_TIMESTAMP);
	`, taskID, "Test Task", time.Now().UTC().Format(time.RFC3339), "project-1")
	if err != nil {
		t.Fatalf("insert pending task: %v", err)
	}
}

func assertTaskCompleted(t *testing.T, database *sql.DB, taskID string, expected bool) {
	t.Helper()

	var completed bool
	err := database.QueryRow(`SELECT completed FROM ticktick_task_cache WHERE task_id = ?;`, taskID).Scan(&completed)
	if err != nil {
		t.Fatalf("load task completion state: %v", err)
	}

	if completed != expected {
		t.Fatalf("unexpected completion state for %s: got=%t want=%t", taskID, completed, expected)
	}
}
