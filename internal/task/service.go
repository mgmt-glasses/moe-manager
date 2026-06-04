package task

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNotFound  = errors.New("task not found")
	ErrForbidden = errors.New("operation not permitted")
	ErrBadInput  = errors.New("invalid input")
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

type CreateInput struct {
	UserID string
	Title  string
}

func (s *Service) Create(ctx context.Context, in CreateInput) (Task, error) {
	if in.Title == "" {
		return Task{}, fmt.Errorf("%w: title is required", ErrBadInput)
	}
	t := Task{
		ID:        uuid.NewString(),
		UserID:    in.UserID,
		Title:     in.Title,
		Status:    StatusTodo,
		CreatedAt: time.Now(),
	}
	if err := s.repo.Create(ctx, t); err != nil {
		return Task{}, err
	}
	return t, nil
}

func (s *Service) List(ctx context.Context, userID string) ([]Task, error) {
	return s.repo.ListByUserID(ctx, userID)
}

func (s *Service) Complete(ctx context.Context, userID, taskID string) (Task, error) {
	t, err := s.repo.FindByID(ctx, taskID)
	if err != nil {
		return Task{}, ErrNotFound
	}
	if t.UserID != userID {
		return Task{}, ErrForbidden
	}
	now := time.Now()
	if err := s.repo.UpdateStatus(ctx, taskID, StatusDone, &now); err != nil {
		return Task{}, err
	}
	t.Status = StatusDone
	t.CompletedAt = &now
	return t, nil
}

func (s *Service) Reopen(ctx context.Context, userID, taskID string) (Task, error) {
	t, err := s.repo.FindByID(ctx, taskID)
	if err != nil {
		return Task{}, ErrNotFound
	}
	if t.UserID != userID {
		return Task{}, ErrForbidden
	}
	if err := s.repo.UpdateStatus(ctx, taskID, StatusTodo, nil); err != nil {
		return Task{}, err
	}
	t.Status = StatusTodo
	t.CompletedAt = nil
	return t, nil
}

func (s *Service) Delete(ctx context.Context, userID, taskID string) error {
	t, err := s.repo.FindByID(ctx, taskID)
	if err != nil {
		return ErrNotFound
	}
	if t.UserID != userID {
		return ErrForbidden
	}
	return s.repo.SoftDelete(ctx, taskID)
}
