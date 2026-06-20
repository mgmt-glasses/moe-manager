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

func (f *FakeQuery) GetRangeStats(_ context.Context, _ string, fromDate, toDate time.Time) ([]statistics.DailyStats, error) {
	if f.Err != nil {
		return nil, f.Err
	}

	daysCount := int(toDate.Sub(fromDate).Hours()/24) + 1
	if daysCount <= 0 {
		daysCount = 1
	}

	var results []statistics.DailyStats
	for i := 0; i < daysCount; i++ {
		d := fromDate.AddDate(0, 0, i)
		if stats, ok := f.StatsByDate[d.Format("2006-01-02")]; ok {
			results = append(results, stats)
		} else {
			results = append(results, statistics.DailyStats{Date: d})
		}
	}
	return results, nil
}
