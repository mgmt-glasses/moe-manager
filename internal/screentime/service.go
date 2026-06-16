package screentime

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) UpsertRecord(ctx context.Context, userID string, date time.Time, minutes, targetMinutes int) (*ScreenTimeRecord, error) {
	existing, err := s.repo.GetByDate(ctx, userID, date)
	if err != nil {
		return nil, err
	}

	var rec ScreenTimeRecord
	now := time.Now()

	if existing != nil {
		rec = *existing
		rec.Minutes = minutes
		rec.TargetMinutes = targetMinutes
		rec.UpdatedAt = now
	} else {
		rec = ScreenTimeRecord{
			RecordID:      "ent_" + uuid.New().String()[:8],
			UserID:        userID,
			Date:          date,
			Minutes:       minutes,
			TargetMinutes: targetMinutes,
			CreatedAt:     now,
			UpdatedAt:     now,
		}
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
