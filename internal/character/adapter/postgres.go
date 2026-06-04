package adapter

import (
	"context"
	"database/sql"
	"errors"

	"github.com/mgmt-glasses/moe-manager/internal/character"
)

type PostgresCharacterRepository struct {
	db *sql.DB
}

func NewPostgresCharacterRepository(db *sql.DB) *PostgresCharacterRepository {
	return &PostgresCharacterRepository{db: db}
}

func scan(row interface {
	Scan(...any) error
}) (character.Character, error) {
	var c character.Character
	return c, row.Scan(
		&c.ID, &c.Name, &c.MBTIType, &c.PersonalityDesc,
		&c.SpeechStyle, &c.SystemPromptFragment, &c.VoicePresetID,
		&c.IconPath, &c.StandingImagePath, &c.SampleVoicePath,
		&c.CreatedAt, &c.UpdatedAt,
	)
}

func (r *PostgresCharacterRepository) List(ctx context.Context) ([]character.Character, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, name, mbti_type, personality_desc, speech_style,
		       system_prompt_fragment, voice_preset_id, icon_path,
		       standing_image_path, sample_voice_path, created_at, updated_at
		FROM characters
		ORDER BY mbti_type
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var chars []character.Character
	for rows.Next() {
		c, err := scan(rows)
		if err != nil {
			return nil, err
		}
		chars = append(chars, c)
	}
	return chars, rows.Err()
}

func (r *PostgresCharacterRepository) FindByID(ctx context.Context, id string) (character.Character, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, name, mbti_type, personality_desc, speech_style,
		       system_prompt_fragment, voice_preset_id, icon_path,
		       standing_image_path, sample_voice_path, created_at, updated_at
		FROM characters
		WHERE id = $1
	`, id)
	c, err := scan(row)
	if errors.Is(err, sql.ErrNoRows) {
		return character.Character{}, character.ErrNotFound
	}
	return c, err
}

func (r *PostgresCharacterRepository) Exists(ctx context.Context, id string) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM characters WHERE id = $1`, id).Scan(&count)
	return count > 0, err
}
