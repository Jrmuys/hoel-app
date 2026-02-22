CREATE TABLE IF NOT EXISTS api_errors (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    service_name TEXT NOT NULL,
    endpoint TEXT NOT NULL,
    http_status INTEGER,
    error_message TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    resolved BOOLEAN DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_api_errors_service_created_at
    ON api_errors (service_name, created_at);

CREATE TABLE IF NOT EXISTS integration_status (
    service_name TEXT PRIMARY KEY,
    last_success_at DATETIME,
    last_error_at DATETIME,
    consecutive_failures INTEGER NOT NULL DEFAULT 0,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
