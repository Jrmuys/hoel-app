CREATE TABLE IF NOT EXISTS pgh_schedule_cache (
    singleton_id INTEGER PRIMARY KEY CHECK (singleton_id = 1),
    next_pickup_date TEXT NOT NULL,
    next_recycling_date TEXT,
    is_recycling_week BOOLEAN NOT NULL,
    show_indicator BOOLEAN NOT NULL,
    source_updated_at DATETIME,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
