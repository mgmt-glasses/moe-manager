package statistics

import (
	"context"
	"time"
)

type StatisticsQuery interface {
	GetDailyStats(ctx context.Context, userID string, date time.Time) (DailyStats, error)
	GetRangeStats(ctx context.Context, userID string, fromDate, toDate time.Time) ([]DailyStats, error)
}
