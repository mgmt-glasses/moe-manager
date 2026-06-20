package testutil

import (
	"context"
	"time"

	"github.com/mgmt-glasses/moe-manager/internal/statistics"
)

// FakeQuery is an in-memory test double for statistics.StatisticsQuery.
// StatsByDate keys are formatted as "2006-01-02".
type FakeQuery struct {
	StatsByDate map[string]statistics.DailyStats
	Err         error
	Calls       []time.Time
}

func (f *FakeQuery) GetDailyStats(_ context.Context, _ string, date time.Time) (statistics.DailyStats, error) {
	f.Calls = append(f.Calls, date)
	if f.Err != nil {
		return statistics.DailyStats{}, f.Err
	}
	if stats, ok := f.StatsByDate[date.Format("2006-01-02")]; ok {
		return stats, nil
	}
	return statistics.DailyStats{Date: date}, nil
}
