package adapter

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/mgmt-glasses/moe-manager/internal/task"
)

type PostgresTaskRepository struct {
	db *sql.DB
}

func NewPostgresTaskRepository(db *sql.DB) *PostgresTaskRepository {
	return &PostgresTaskRepository{db: db}
}

func (r *PostgresTaskRepository) Create(ctx context.Context, t task.Task) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO tasks (id, user_id, title, status, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`, t.ID, t.UserID, t.Title, string(t.Status), t.CreatedAt)
	return err
}

func (r *PostgresTaskRepository) ListByUserID(ctx context.Context, userID string) ([]task.Task, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, user_id, title, status, created_at, completed_at
		FROM tasks
		WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []task.Task
	for rows.Next() {
		var t task.Task
		var status string
		var completedAt sql.NullTime
		if err := rows.Scan(&t.ID, &t.UserID, &t.Title, &status, &t.CreatedAt, &completedAt); err != nil {
			return nil, err
		}
		t.Status = task.Status(status)
		if completedAt.Valid {
			t.CompletedAt = &completedAt.Time
		}
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}

func (r *PostgresTaskRepository) FindByID(ctx context.Context, taskID string) (task.Task, error) {
	var t task.Task
	var status string
	var completedAt sql.NullTime
	err := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, title, status, created_at, completed_at
		FROM tasks
		WHERE id = $1 AND deleted_at IS NULL
	`, taskID).Scan(&t.ID, &t.UserID, &t.Title, &status, &t.CreatedAt, &completedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return task.Task{}, task.ErrNotFound
	}
	if err != nil {
		return task.Task{}, err
	}
	t.Status = task.Status(status)
	if completedAt.Valid {
		t.CompletedAt = &completedAt.Time
	}
	return t, nil
}

func (r *PostgresTaskRepository) UpdateStatus(ctx context.Context, taskID string, status task.Status, completedAt *time.Time) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE tasks SET status = $1, completed_at = $2 WHERE id = $3
	`, string(status), completedAt, taskID)
	return err
}

func (r *PostgresTaskRepository) SoftDelete(ctx context.Context, taskID string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE tasks SET deleted_at = $1 WHERE id = $2
	`, time.Now(), taskID)
	return err
}
