package testutil

import (
	"context"

	"github.com/mgmt-glasses/moe-manager/internal/user"
)

// FakeRepository is an in-memory test double for user.Repository.
type FakeRepository struct {
	Users     map[string]user.User
	CreateErr error
	FindErr   error
	UpdateErr error
}

// NewFakeRepository seeds the repository with the given users.
func NewFakeRepository(users ...user.User) *FakeRepository {
	m := make(map[string]user.User, len(users))
	for _, u := range users {
		m[u.ID] = u
	}
	return &FakeRepository{Users: m}
}

func (f *FakeRepository) Create(_ context.Context, u user.User) error {
	if f.CreateErr != nil {
		return f.CreateErr
	}
	if f.Users == nil {
		f.Users = make(map[string]user.User)
	}
	f.Users[u.ID] = u
	return nil
}

func (f *FakeRepository) FindByID(_ context.Context, id string) (user.User, error) {
	if f.FindErr != nil {
		return user.User{}, f.FindErr
	}
	u, ok := f.Users[id]
	if !ok {
		return user.User{}, user.ErrNotFound
	}
	return u, nil
}

func (f *FakeRepository) Update(_ context.Context, u user.User) error {
	if f.UpdateErr != nil {
		return f.UpdateErr
	}
	if _, ok := f.Users[u.ID]; !ok {
		return user.ErrNotFound
	}
	f.Users[u.ID] = u
	return nil
}
