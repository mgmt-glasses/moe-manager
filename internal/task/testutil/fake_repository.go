package testutil

import (
	"context"
	"time"

	"github.com/mgmt-glasses/moe-manager/internal/task"
)

// FakeRepository is an in-memory test double for task.Repository.
type FakeRepository struct {
	Tasks           map[string]task.Task
	CreateErr       error
	ListErr         error
	FindErr         error
	UpdateStatusErr error
	SoftDeleteErr   error
}

// NewFakeRepository seeds the repository with the given tasks.
func NewFakeRepository(tasks ...task.Task) *FakeRepository {
	m := make(map[string]task.Task, len(tasks))
	for _, t := range tasks {
		m[t.ID] = t
	}
	return &FakeRepository{Tasks: m}
}

func (f *FakeRepository) Create(_ context.Context, t task.Task) error {
	if f.CreateErr != nil {
		return f.CreateErr
	}
	if f.Tasks == nil {
		f.Tasks = make(map[string]task.Task)
	}
	f.Tasks[t.ID] = t
	return nil
}

func (f *FakeRepository) ListByUserID(_ context.Context, userID string) ([]task.Task, error) {
	if f.ListErr != nil {
		return nil, f.ListErr
	}
	var result []task.Task
	for _, t := range f.Tasks {
		if t.UserID == userID {
			result = append(result, t)
		}
	}
	return result, nil
}

func (f *FakeRepository) FindByID(_ context.Context, taskID string) (task.Task, error) {
	if f.FindErr != nil {
		return task.Task{}, f.FindErr
	}
	t, ok := f.Tasks[taskID]
	if !ok {
		return task.Task{}, task.ErrNotFound
	}
	return t, nil
}

func (f *FakeRepository) UpdateStatus(_ context.Context, taskID string, status task.Status, completedAt *time.Time) error {
	if f.UpdateStatusErr != nil {
		return f.UpdateStatusErr
	}
	t, ok := f.Tasks[taskID]
	if !ok {
		return task.ErrNotFound
	}
	t.Status = status
	t.CompletedAt = completedAt
	f.Tasks[taskID] = t
	return nil
}

func (f *FakeRepository) SoftDelete(_ context.Context, taskID string) error {
	if f.SoftDeleteErr != nil {
		return f.SoftDeleteErr
	}
	if _, ok := f.Tasks[taskID]; !ok {
		return task.ErrNotFound
	}
	delete(f.Tasks, taskID)
	return nil
}
