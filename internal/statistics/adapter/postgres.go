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
	nextDate := date.AddDate(0, 0, 1).Format("2006-01-02")

	var completed, todo int
	row := q.db.QueryRowContext(ctx, `
		SELECT
			COUNT(*) FILTER (WHERE status = 'done') AS completed,
			COUNT(*) FILTER (WHERE status = 'todo') AS todo
		FROM tasks
		WHERE user_id = $1 AND created_at >= $2 AND created_at < $3 AND deleted_at IS NULL
	`, userID, targetDate, nextDate)
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

func (q *PostgresStatisticsQuery) GetRangeStats(ctx context.Context, userID string, fromDate, toDate time.Time) ([]statistics.DailyStats, error) {
	tasksRows, err := q.db.QueryContext(ctx, `
		SELECT
			created_at::date as d,
			COUNT(*) FILTER (WHERE status = 'done') AS completed,
			COUNT(*) FILTER (WHERE status = 'todo') AS todo
		FROM tasks
		WHERE user_id = $1 AND created_at >= $2 AND created_at < $3 AND deleted_at IS NULL
		GROUP BY d
	`, userID, fromDate.Format("2006-01-02"), toDate.AddDate(0, 0, 1).Format("2006-01-02"))
	if err != nil {
		return nil, err
	}
	defer tasksRows.Close()

	type taskStat struct{ completed, todo int }
	taskStats := make(map[string]taskStat)
	for tasksRows.Next() {
		var d time.Time
		var completed, todo int
		if err := tasksRows.Scan(&d, &completed, &todo); err != nil {
			return nil, err
		}
		taskStats[d.Format("2006-01-02")] = taskStat{completed, todo}
	}

	stRows, err := q.db.QueryContext(ctx, `
		SELECT date, minutes, target_minutes_snapshot
		FROM screentime_records
		WHERE user_id = $1 AND date >= $2 AND date <= $3
	`, userID, fromDate.Format("2006-01-02"), toDate.Format("2006-01-02"))
	if err != nil {
		return nil, err
	}
	defer stRows.Close()

	type stStat struct{ minutes, target int }
	stStats := make(map[string]stStat)
	for stRows.Next() {
		var d time.Time
		var m, t int
		if err := stRows.Scan(&d, &m, &t); err != nil {
			return nil, err
		}
		stStats[d.Format("2006-01-02")] = stStat{m, t}
	}

	daysCount := int(toDate.Sub(fromDate).Hours()/24) + 1
	if daysCount <= 0 {
		daysCount = 1
	}
	result := make([]statistics.DailyStats, daysCount)
	for i := 0; i < daysCount; i++ {
		d := fromDate.AddDate(0, 0, i)
		dStr := d.Format("2006-01-02")
		
		ts := taskStats[dStr]
		ss := stStats[dStr]
		
		total := ts.completed + ts.todo
		rate := 0
		if total > 0 {
			rate = ts.completed * 100 / total
		}
		
		result[i] = statistics.DailyStats{
			Date:                       d,
			CompletedTaskCount:         ts.completed,
			TodoTaskCount:              ts.todo,
			TotalTaskCount:             total,
			TaskCompletionRate:         rate,
			EntertainmentMinutes:       ss.minutes,
			TargetEntertainmentMinutes: ss.target,
			EntertainmentDiffMinutes:   ss.minutes - ss.target,
		}
	}

	return result, nil
}
