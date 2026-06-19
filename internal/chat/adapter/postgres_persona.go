package chatadapter

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/mgmt-glasses/moe-manager/internal/chat"
)

type PostgresPersonaRepository struct {
	db *sql.DB
}

func NewPostgresPersonaRepository(db *sql.DB) *PostgresPersonaRepository {
	return &PostgresPersonaRepository{db: db}
}

func (r *PostgresPersonaRepository) Get(ctx context.Context, characterID string) (chat.Persona, error) {
	var fragment string
	err := r.db.QueryRowContext(ctx,
		`SELECT system_prompt_fragment FROM characters WHERE id = $1`,
		characterID,
	).Scan(&fragment)
	if errors.Is(err, sql.ErrNoRows) {
		return chat.Persona{}, fmt.Errorf("character %s not found", characterID)
	}
	if err != nil {
		return chat.Persona{}, fmt.Errorf("query persona: %w", err)
	}
	return chat.Persona{
		CharacterID:       characterID,
		SystemInstruction: fragment,
	}, nil
}
