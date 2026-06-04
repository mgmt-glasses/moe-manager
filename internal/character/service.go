package character

import (
	"context"
	"errors"
)

var ErrNotFound = errors.New("character not found")

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context) ([]Character, error) {
	return s.repo.List(ctx)
}

func (s *Service) FindByID(ctx context.Context, id string) (Character, error) {
	c, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return Character{}, ErrNotFound
	}
	return c, nil
}
