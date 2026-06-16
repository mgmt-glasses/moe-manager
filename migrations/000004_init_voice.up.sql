CREATE TABLE IF NOT EXISTS voice_files (
    id           TEXT PRIMARY KEY,
    user_id      TEXT NOT NULL,
    character_id TEXT NOT NULL,
    source_text  TEXT NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_voice_files_user ON voice_files (user_id, created_at);
