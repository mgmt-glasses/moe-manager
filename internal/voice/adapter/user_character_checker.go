package voiceadapter

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/mgmt-glasses/moe-manager/internal/voice"
)

// PostgresUserCharacterChecker implements voice.UserCharacterChecker using the users table.
type PostgresUserCharacterChecker struct {
	db *sql.DB
}

func NewPostgresUserCharacterChecker(db *sql.DB) *PostgresUserCharacterChecker {
	return &PostgresUserCharacterChecker{db: db}
}

func (c *PostgresUserCharacterChecker) GetSelectedCharacterID(ctx context.Context, userID string) (string, error) {
	var selectedCharID sql.NullString
	err := c.db.QueryRowContext(ctx,
		`SELECT selected_character_id FROM users WHERE id = $1`,
		userID,
	).Scan(&selectedCharID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", voice.ErrUserNotFound
	}
	if err != nil {
		return "", fmt.Errorf("query user: %w", err)
	}
	if !selectedCharID.Valid {
		return "", voice.ErrNoCharacterSelected
	}
	return selectedCharID.String, nil
}
