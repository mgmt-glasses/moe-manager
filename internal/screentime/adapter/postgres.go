package adapter

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/mgmt-glasses/moe-manager/internal/screentime"
)

type PostgresScreenTimeRepository struct {
	db *sql.DB
}

func NewPostgresScreenTimeRepository(db *sql.DB) *PostgresScreenTimeRepository {
	return &PostgresScreenTimeRepository{db: db}
}

func (r *PostgresScreenTimeRepository) Save(ctx context.Context, record screentime.ScreenTimeRecord) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO screentime_records (record_id, user_id, date, minutes, target_minutes_snapshot, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (user_id, date) DO UPDATE SET
			minutes = EXCLUDED.minutes,
			target_minutes_snapshot = EXCLUDED.target_minutes_snapshot,
			updated_at = EXCLUDED.updated_at
	`, record.RecordID, record.UserID, record.Date.Format("2006-01-02"), record.Minutes, record.TargetMinutes, record.CreatedAt, record.UpdatedAt)
	return err
}

func (r *PostgresScreenTimeRepository) GetByDate(ctx context.Context, userID string, date time.Time) (*screentime.ScreenTimeRecord, error) {
	var rec screentime.ScreenTimeRecord
	err := r.db.QueryRowContext(ctx, `
		SELECT record_id, user_id, date, minutes, target_minutes_snapshot, created_at, updated_at
		FROM screentime_records
		WHERE user_id = $1 AND date = $2
	`, userID, date.Format("2006-01-02")).Scan(
		&rec.RecordID, &rec.UserID, &rec.Date, &rec.Minutes, &rec.TargetMinutes, &rec.CreatedAt, &rec.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &rec, nil
}

func (r *PostgresScreenTimeRepository) ListByRange(ctx context.Context, userID string, fromDate, toDate time.Time) ([]screentime.ScreenTimeRecord, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT record_id, user_id, date, minutes, target_minutes_snapshot, created_at, updated_at
		FROM screentime_records
		WHERE user_id = $1 AND date >= $2 AND date <= $3
		ORDER BY date ASC
	`, userID, fromDate.Format("2006-01-02"), toDate.Format("2006-01-02"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []screentime.ScreenTimeRecord
	for rows.Next() {
		var rec screentime.ScreenTimeRecord
		if err := rows.Scan(&rec.RecordID, &rec.UserID, &rec.Date, &rec.Minutes, &rec.TargetMinutes, &rec.CreatedAt, &rec.UpdatedAt); err != nil {
			return nil, err
		}
		records = append(records, rec)
	}
	return records, rows.Err()
}
