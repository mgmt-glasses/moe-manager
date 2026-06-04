CREATE TABLE IF NOT EXISTS tasks (
    id           TEXT PRIMARY KEY,
    user_id      TEXT NOT NULL,
    title        TEXT NOT NULL,
    status       TEXT NOT NULL DEFAULT 'todo',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ,
    deleted_at   TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_tasks_user_status ON tasks (user_id, status, created_at);
