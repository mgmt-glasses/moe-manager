package statistics_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/mgmt-glasses/moe-manager/internal/statistics"
	"github.com/mgmt-glasses/moe-manager/internal/statistics/testutil"
)

func sameDay(a, b time.Time) bool {
	return a.Year() == b.Year() && a.Month() == b.Month() && a.Day() == b.Day()
}

func TestService_GetToday_QueriesCurrentDate(t *testing.T) {
	query := &testutil.FakeQuery{}
	svc := statistics.NewStatisticsService(query)

	if _, err := svc.GetToday(context.Background(), "u_001"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(query.Calls) != 1 {
		t.Fatalf("expected 1 query call, got %d", len(query.Calls))
	}
	if !sameDay(query.Calls[0], time.Now()) {
		t.Errorf("expected query for today, got %v", query.Calls[0])
	}
}

func TestService_GetToday_QueryError(t *testing.T) {
	svc := statistics.NewStatisticsService(&testutil.FakeQuery{Err: errors.New("db down")})

	_, err := svc.GetToday(context.Background(), "u_001")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestService_GetDaily_PassesDateThrough(t *testing.T) {
	date := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	query := &testutil.FakeQuery{StatsByDate: map[string]statistics.DailyStats{
		"2026-06-01": {Date: date, CompletedTaskCount: 3, TotalTaskCount: 5},
	}}
	svc := statistics.NewStatisticsService(query)

	got, err := svc.GetDaily(context.Background(), "u_001", date)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.CompletedTaskCount != 3 || got.TotalTaskCount != 5 {
		t.Errorf("unexpected stats: %+v", got)
	}
}

func TestService_GetDaily_QueryError(t *testing.T) {
	svc := statistics.NewStatisticsService(&testutil.FakeQuery{Err: errors.New("db down")})

	_, err := svc.GetDaily(context.Background(), "u_001", time.Now())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestService_GetWeekly_AggregatesSevenDaysFromEndDate(t *testing.T) {
	end := time.Date(2026, 6, 7, 0, 0, 0, 0, time.UTC)
	svc := statistics.NewStatisticsService(&testutil.FakeQuery{})

	got, err := svc.GetWeekly(context.Background(), "u_001", end)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got.Days) != 7 {
		t.Fatalf("expected 7 days, got %d", len(got.Days))
	}
	wantFrom := end.AddDate(0, 0, -6)
	if !got.From.Equal(wantFrom) {
		t.Errorf("from: got %v, want %v", got.From, wantFrom)
	}
	if !got.To.Equal(end) {
		t.Errorf("to: got %v, want %v", got.To, end)
	}
	if !got.Days[0].Date.Equal(wantFrom) {
		t.Errorf("first day: got %v, want %v", got.Days[0].Date, wantFrom)
	}
	if !got.Days[6].Date.Equal(end) {
		t.Errorf("last day: got %v, want %v", got.Days[6].Date, end)
	}
}

func TestService_GetWeekly_QueryError(t *testing.T) {
	svc := statistics.NewStatisticsService(&testutil.FakeQuery{Err: errors.New("db down")})

	_, err := svc.GetWeekly(context.Background(), "u_001", time.Now())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
