package voiceadapter

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/mgmt-glasses/moe-manager/internal/voice"
)

type PostgresVoiceFileRepository struct {
	db *sql.DB
}

func NewPostgresVoiceFileRepository(db *sql.DB) *PostgresVoiceFileRepository {
	return &PostgresVoiceFileRepository{db: db}
}

func (r *PostgresVoiceFileRepository) Save(ctx context.Context, vf voice.VoiceFile) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO voice_files (id, character_id, source_text, created_at) VALUES ($1, $2, $3, $4)`,
		vf.ID, vf.CharacterID, vf.SourceText, vf.CreatedAt,
	)
	return err
}

func (r *PostgresVoiceFileRepository) FindByID(ctx context.Context, id string) (voice.VoiceFile, error) {
	var vf voice.VoiceFile
	err := r.db.QueryRowContext(ctx,
		`SELECT id, character_id, source_text, created_at FROM voice_files WHERE id = $1`,
		id,
	).Scan(&vf.ID, &vf.CharacterID, &vf.SourceText, &vf.CreatedAt)
	if err == sql.ErrNoRows {
		return voice.VoiceFile{}, voice.ErrNotFound
	}
	if err != nil {
		return voice.VoiceFile{}, fmt.Errorf("query voice file: %w", err)
	}
	return vf, nil
}
