package adapter

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/mgmt-glasses/moe-manager/internal/user"
)

// pgUniqueViolation は PostgreSQL の一意制約違反コード。
const pgUniqueViolation = "23505"

type PostgresUserRepository struct {
	db *sql.DB
}

func NewPostgresUserRepository(db *sql.DB) *PostgresUserRepository {
	return &PostgresUserRepository{db: db}
}

func (r *PostgresUserRepository) Create(ctx context.Context, u user.User) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO users (id, name, president_name, selected_character_id, target_entertainment_minutes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, u.ID, u.Name, u.PresidentName, u.SelectedCharacterID, u.TargetEntertainmentMinutes, u.CreatedAt, u.UpdatedAt)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
		return user.ErrAlreadyExists
	}
	return err
}

func (r *PostgresUserRepository) FindByID(ctx context.Context, id string) (user.User, error) {
	var u user.User
	var selectedCharID sql.NullString
	err := r.db.QueryRowContext(ctx, `
		SELECT id, name, president_name, selected_character_id, target_entertainment_minutes, created_at, updated_at
		FROM users
		WHERE id = $1
	`, id).Scan(
		&u.ID, &u.Name, &u.PresidentName, &selectedCharID,
		&u.TargetEntertainmentMinutes, &u.CreatedAt, &u.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return user.User{}, user.ErrNotFound
	}
	if err != nil {
		return user.User{}, err
	}
	if selectedCharID.Valid {
		u.SelectedCharacterID = &selectedCharID.String
	}
	return u, nil
}

func (r *PostgresUserRepository) Update(ctx context.Context, u user.User) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE users
		SET name = $1, president_name = $2, selected_character_id = $3,
		    target_entertainment_minutes = $4, updated_at = $5
		WHERE id = $6
	`, u.Name, u.PresidentName, u.SelectedCharacterID, u.TargetEntertainmentMinutes, u.UpdatedAt, u.ID)
	return err
}
