CREATE TABLE IF NOT EXISTS ticktick_oauth_token (
    singleton_id INTEGER PRIMARY KEY CHECK (singleton_id = 1),
    access_token TEXT NOT NULL,
    refresh_token TEXT,
    token_type TEXT,
    scope TEXT,
    expires_at DATETIME,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
