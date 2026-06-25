package user

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNotFound          = errors.New("user not found")
	ErrBadInput          = errors.New("invalid input")
	ErrCharacterNotFound = errors.New("character not found")
)

type Service struct {
	repo      Repository
	charValid CharacterValidator
}

func NewService(repo Repository, charValid CharacterValidator) *Service {
	return &Service{repo: repo, charValid: charValid}
}

type CreateInput struct {
	UserID                     string
	Name                       string
	PresidentName              string
	TargetEntertainmentMinutes int
	SelectedCharacterID        *string
}

func (s *Service) Create(ctx context.Context, in CreateInput) (User, error) {
	if in.Name == "" {
		return User{}, fmt.Errorf("%w: name is required", ErrBadInput)
	}
	if in.PresidentName == "" {
		return User{}, fmt.Errorf("%w: presidentName is required", ErrBadInput)
	}

	if in.SelectedCharacterID != nil {
		ok, err := s.charValid.Exists(ctx, *in.SelectedCharacterID)
		if err != nil {
			return User{}, err
		}
		if !ok {
			return User{}, ErrCharacterNotFound
		}
	}

	target := in.TargetEntertainmentMinutes
	if target <= 0 {
		target = 120
	}

	now := time.Now()
	userID := in.UserID
	if userID == "" {
		userID = uuid.NewString()
	}
	u := User{
		ID:                         userID,
		Name:                       in.Name,
		PresidentName:              in.PresidentName,
		SelectedCharacterID:        in.SelectedCharacterID,
		TargetEntertainmentMinutes: target,
		CreatedAt:                  now,
		UpdatedAt:                  now,
	}
	if err := s.repo.Create(ctx, u); err != nil {
		return User{}, err
	}
	return u, nil
}

func (s *Service) FindByID(ctx context.Context, id string) (User, error) {
	return s.repo.FindByID(ctx, id)
}

type UpdateInput struct {
	Name                       *string
	PresidentName              *string
	TargetEntertainmentMinutes *int
	SelectedCharacterID        *string
}

func (s *Service) Update(ctx context.Context, id string, in UpdateInput) (User, error) {
	u, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return User{}, err
	}

	if in.Name != nil {
		if *in.Name == "" {
			return User{}, fmt.Errorf("%w: name cannot be empty", ErrBadInput)
		}
		u.Name = *in.Name
	}
	if in.PresidentName != nil {
		if *in.PresidentName == "" {
			return User{}, fmt.Errorf("%w: presidentName cannot be empty", ErrBadInput)
		}
		u.PresidentName = *in.PresidentName
	}
	if in.TargetEntertainmentMinutes != nil {
		if *in.TargetEntertainmentMinutes <= 0 {
			return User{}, fmt.Errorf("%w: targetEntertainmentMinutes must be positive", ErrBadInput)
		}
		u.TargetEntertainmentMinutes = *in.TargetEntertainmentMinutes
	}
	if in.SelectedCharacterID != nil {
		ok, err := s.charValid.Exists(ctx, *in.SelectedCharacterID)
		if err != nil {
			return User{}, err
		}
		if !ok {
			return User{}, ErrCharacterNotFound
		}
		u.SelectedCharacterID = in.SelectedCharacterID
	}

	u.UpdatedAt = time.Now()
	if err := s.repo.Update(ctx, u); err != nil {
		return User{}, err
	}
	return u, nil
}

func (s *Service) UpdateSelectedCharacter(ctx context.Context, userID, characterID string) (User, error) {
	u, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return User{}, err
	}

	ok, err := s.charValid.Exists(ctx, characterID)
	if err != nil {
		return User{}, err
	}
	if !ok {
		return User{}, ErrCharacterNotFound
	}

	u.SelectedCharacterID = &characterID
	u.UpdatedAt = time.Now()
	if err := s.repo.Update(ctx, u); err != nil {
		return User{}, err
	}
	return u, nil
}
