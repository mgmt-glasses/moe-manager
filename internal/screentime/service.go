package screentime

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrAnalyzerUnavailable = errors.New("analyzer not configured")

type Service struct {
	repo     Repository
	analyzer ImageAnalyzer
}

func NewService(repo Repository, analyzer ImageAnalyzer) *Service {
	return &Service{repo: repo, analyzer: analyzer}
}

func (s *Service) UpsertRecord(ctx context.Context, userID string, date time.Time, minutes, targetMinutes int) (*ScreenTimeRecord, error) {
	now := time.Now()

	existing, err := s.repo.GetByDate(ctx, userID, date)
	if err != nil {
		return nil, err
	}

	recordID := "ent_" + uuid.New().String()
	createdAt := now
	if existing != nil {
		recordID = existing.RecordID
		createdAt = existing.CreatedAt
	}

	rec := ScreenTimeRecord{
		RecordID:      recordID,
		UserID:        userID,
		Date:          date,
		Minutes:       minutes,
		TargetMinutes: targetMinutes,
		CreatedAt:     createdAt,
		UpdatedAt:     now,
	}

	rec.CalculateDiff()

	if err := s.repo.Save(ctx, rec); err != nil {
		return nil, err
	}

	return &rec, nil
}

func (s *Service) GetRecord(ctx context.Context, userID string, date time.Time) (*ScreenTimeRecord, error) {
	rec, err := s.repo.GetByDate(ctx, userID, date)
	if err != nil {
		return nil, err
	}
	if rec != nil {
		rec.CalculateDiff()
	}
	return rec, nil
}

func (s *Service) ListRecords(ctx context.Context, userID string, fromDate, toDate time.Time) ([]ScreenTimeRecord, error) {
	records, err := s.repo.ListByRange(ctx, userID, fromDate, toDate)
	if err != nil {
		return nil, err
	}
	for i := range records {
		records[i].CalculateDiff()
	}
	return records, nil
}

func (s *Service) Analyze(ctx context.Context, base64Data, mimeType string) (*AnalysisResult, error) {
	if s.analyzer == nil {
		return nil, ErrAnalyzerUnavailable
	}
	return s.analyzer.AnalyzeImage(ctx, base64Data, mimeType)
}
