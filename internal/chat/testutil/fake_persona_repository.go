package testutil

import (
	"context"

	"github.com/mgmt-glasses/moe-manager/internal/chat"
)

// FakePersonaRepository is a test double for chat.PersonaRepository.
type FakePersonaRepository struct {
	Persona chat.Persona
	Err     error
}

func (f *FakePersonaRepository) Get(_ context.Context, characterID string) (chat.Persona, error) {
	if f.Err != nil {
		return chat.Persona{}, f.Err
	}
	p := f.Persona
	p.CharacterID = characterID
	return p, nil
}
