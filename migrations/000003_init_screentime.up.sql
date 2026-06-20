CREATE TABLE IF NOT EXISTS screentime_records (
    id                      TEXT PRIMARY KEY,
    user_id                 TEXT NOT NULL,
    date                    DATE NOT NULL,
    minutes                 INTEGER NOT NULL,
    target_minutes_snapshot INTEGER NOT NULL,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, date)
);

CREATE INDEX IF NOT EXISTS idx_screentime_user_date ON screentime_records (user_id, date);
