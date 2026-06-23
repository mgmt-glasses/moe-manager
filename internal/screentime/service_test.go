package screentime_test

import (
	"context"
	"testing"
	"time"

	"github.com/mgmt-glasses/moe-manager/internal/screentime"
	"github.com/mgmt-glasses/moe-manager/internal/screentime/testutil"
)

func TestUpsertRecord(t *testing.T) {
	repo := testutil.NewFakeRepository()
	svc := screentime.NewService(repo, nil)

	ctx := context.Background()
	userID := "user-123"
	date := time.Now().Truncate(24 * time.Hour)

	// 1. 初回のUpsert (Insert)
	rec1, err := svc.UpsertRecord(ctx, userID, date, 100, 60)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec1.RecordID == "" {
		t.Errorf("expected RecordID to be generated")
	}
	if rec1.Minutes != 100 {
		t.Errorf("expected 100 minutes, got %d", rec1.Minutes)
	}
	if rec1.DiffMinutes != 40 {
		t.Errorf("expected diff to be 40, got %d", rec1.DiffMinutes)
	}

	// 2. 2回目のUpsert (Update) - 同じ日のデータ
	time.Sleep(10 * time.Millisecond) // Ensure UpdatedAt is different if needed
	rec2, err := svc.UpsertRecord(ctx, userID, date, 150, 60)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec2.RecordID != rec1.RecordID {
		t.Errorf("expected RecordID to be maintained during update. old=%s, new=%s", rec1.RecordID, rec2.RecordID)
	}
	if !rec2.CreatedAt.Equal(rec1.CreatedAt) {
		t.Errorf("expected CreatedAt to be maintained. old=%v, new=%v", rec1.CreatedAt, rec2.CreatedAt)
	}
	if rec2.Minutes != 150 {
		t.Errorf("expected 150 minutes, got %d", rec2.Minutes)
	}
}

func TestAnalyze_AnalyzerNotConfigured(t *testing.T) {
	repo := testutil.NewFakeRepository()
	// analyzer = nil
	svc := screentime.NewService(repo, nil)

	_, err := svc.Analyze(context.Background(), "dummy_base64", "image/jpeg")
	if err != screentime.ErrAnalyzerUnavailable {
		t.Errorf("expected ErrAnalyzerUnavailable, got %v", err)
	}
}
