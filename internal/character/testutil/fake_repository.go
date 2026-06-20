package testutil

import (
	"context"

	"github.com/mgmt-glasses/moe-manager/internal/character"
)

// FakeRepository is an in-memory test double for character.Repository.
type FakeRepository struct {
	Characters []character.Character
	ListErr    error
	FindErr    error
	ExistsErr  error
}

func (f *FakeRepository) List(_ context.Context) ([]character.Character, error) {
	if f.ListErr != nil {
		return nil, f.ListErr
	}
	return f.Characters, nil
}

func (f *FakeRepository) FindByID(_ context.Context, id string) (character.Character, error) {
	if f.FindErr != nil {
		return character.Character{}, f.FindErr
	}
	for _, c := range f.Characters {
		if c.ID == id {
			return c, nil
		}
	}
	return character.Character{}, character.ErrNotFound
}

func (f *FakeRepository) Exists(_ context.Context, id string) (bool, error) {
	if f.ExistsErr != nil {
		return false, f.ExistsErr
	}
	for _, c := range f.Characters {
		if c.ID == id {
			return true, nil
		}
	}
	return false, nil
}
