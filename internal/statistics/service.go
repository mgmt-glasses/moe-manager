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
	days, err := s.query.GetRangeStats(ctx, userID, from, endDate)
	if err != nil {
		return WeeklyStats{}, err
	}
	return WeeklyStats{From: from, To: endDate, Days: days}, nil
}
