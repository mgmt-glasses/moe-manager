package task

import (
	"context"
	"time"
)

type Repository interface {
	Create(ctx context.Context, t Task) error
	ListByUserID(ctx context.Context, userID string) ([]Task, error)
	FindByID(ctx context.Context, taskID string) (Task, error)
	UpdateStatus(ctx context.Context, taskID string, status Status, completedAt *time.Time) error
	SoftDelete(ctx context.Context, taskID string) error
}
