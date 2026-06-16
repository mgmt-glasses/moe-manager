package statistics

import "time"

type DailyStats struct {
	Date                       time.Time
	CompletedTaskCount         int
	TodoTaskCount              int
	TotalTaskCount             int
	TaskCompletionRate         int // 0-100（%）
	EntertainmentMinutes       int
	TargetEntertainmentMinutes int
	EntertainmentDiffMinutes   int
}

type WeeklyStats struct {
	From time.Time
	To   time.Time
	Days []DailyStats
}
