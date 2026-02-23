ALTER TABLE ticktick_task_cache
ADD COLUMN tags_json TEXT NOT NULL DEFAULT '[]';
