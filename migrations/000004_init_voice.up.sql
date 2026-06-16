CREATE TABLE IF NOT EXISTS voice_files (
    id           TEXT PRIMARY KEY,
    character_id TEXT NOT NULL,
    source_text  TEXT NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_voice_files_character ON voice_files (character_id, created_at);
