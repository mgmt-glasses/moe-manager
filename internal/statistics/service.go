package statistics

import (
	"context"
	"time"
)


type StatisticsService struct {
	query StatisticsQuery
}

func NewStatisticsService(query StatisticsQuery) *StatisticsService {
	return &StatisticsService{query: query}
}

func (s *StatisticsService) GetToday(ctx context.Context, userID string) (DailyStats, error) {
	return s.query.GetDailyStats(ctx, userID, time.Now())
}

func (s *StatisticsService) GetDaily(ctx context.Context, userID string, date time.Time) (DailyStats, error) {
	return s.query.GetDailyStats(ctx, userID, date)
}

func (s *StatisticsService) GetWeekly(ctx context.Context, userID string, endDate time.Time) (WeeklyStats, error) {
	from := endDate.AddDate(0, 0, -6)
	days := make([]DailyStats, 7)
	for i := range 7 {
		d := from.AddDate(0, 0, i)
		stats, err := s.query.GetDailyStats(ctx, userID, d)
		if err != nil {
			return WeeklyStats{}, err
		}
		days[i] = stats
	}
	return WeeklyStats{From: from, To: endDate, Days: days}, nil
}
