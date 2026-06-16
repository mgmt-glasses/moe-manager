package adapter

import (
	"context"
	"database/sql"
	"time"

	"github.com/mgmt-glasses/moe-manager/internal/statistics"
)

type PostgresStatisticsQuery struct {
	db *sql.DB
}

func NewPostgresStatisticsQuery(db *sql.DB) *PostgresStatisticsQuery {
	return &PostgresStatisticsQuery{db: db}
}

func (q *PostgresStatisticsQuery) GetDailyStats(ctx context.Context, userID string, date time.Time) (statistics.DailyStats, error) {
	targetDate := date.Format("2006-01-02")

	var completed, todo int
	row := q.db.QueryRowContext(ctx, `
		SELECT
			COUNT(*) FILTER (WHERE status = 'done') AS completed,
			COUNT(*) FILTER (WHERE status = 'todo') AS todo
		FROM tasks
		WHERE user_id = $1 AND created_at::date = $2 AND deleted_at IS NULL
	`, userID, targetDate)
	if err := row.Scan(&completed, &todo); err != nil {
		return statistics.DailyStats{}, err
	}

	total := completed + todo
	completionRate := 0
	if total > 0 {
		completionRate = completed * 100 / total
	}

	var entertainmentMinutes, targetMinutes int
	row = q.db.QueryRowContext(ctx, `
		SELECT minutes, target_minutes_snapshot
		FROM screentime_records
		WHERE user_id = $1 AND date = $2
	`, userID, targetDate)
	if err := row.Scan(&entertainmentMinutes, &targetMinutes); err != nil && err != sql.ErrNoRows {
		return statistics.DailyStats{}, err
	}

	return statistics.DailyStats{
		Date:                       date,
		CompletedTaskCount:         completed,
		TodoTaskCount:              todo,
		TotalTaskCount:             total,
		TaskCompletionRate:         completionRate,
		EntertainmentMinutes:       entertainmentMinutes,
		TargetEntertainmentMinutes: targetMinutes,
		EntertainmentDiffMinutes:   entertainmentMinutes - targetMinutes,
	}, nil
}
