package testutil

import "context"

// FakeCharacterValidator is an in-memory test double for user.CharacterValidator.
type FakeCharacterValidator struct {
	ExistingIDs map[string]bool
	Err         error
}

// NewFakeCharacterValidator marks the given character IDs as existing.
func NewFakeCharacterValidator(ids ...string) *FakeCharacterValidator {
	m := make(map[string]bool, len(ids))
	for _, id := range ids {
		m[id] = true
	}
	return &FakeCharacterValidator{ExistingIDs: m}
}

func (f *FakeCharacterValidator) Exists(_ context.Context, characterID string) (bool, error) {
	if f.Err != nil {
		return false, f.Err
	}
	return f.ExistingIDs[characterID], nil
}
