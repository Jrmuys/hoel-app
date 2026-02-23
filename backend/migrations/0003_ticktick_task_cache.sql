CREATE TABLE IF NOT EXISTS ticktick_task_cache (
    task_id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    due_at TEXT NOT NULL,
    completed BOOLEAN NOT NULL DEFAULT 0,
    source_project TEXT NOT NULL,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_ticktick_task_cache_due_at
    ON ticktick_task_cache (due_at);
