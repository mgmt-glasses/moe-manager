package testutil

import (
	"context"
	"time"

	"github.com/mgmt-glasses/moe-manager/internal/screentime"
)

type FakeRepository struct {
	Records map[string]screentime.ScreenTimeRecord
	Err     error
}

func NewFakeRepository() *FakeRepository {
	return &FakeRepository{
		Records: make(map[string]screentime.ScreenTimeRecord),
	}
}

func (f *FakeRepository) key(userID string, date time.Time) string {
	return userID + "_" + date.Format("2006-01-02")
}

func (f *FakeRepository) Save(ctx context.Context, record screentime.ScreenTimeRecord) error {
	if f.Err != nil {
		return f.Err
	}
	f.Records[f.key(record.UserID, record.Date)] = record
	return nil
}

func (f *FakeRepository) GetByDate(ctx context.Context, userID string, date time.Time) (*screentime.ScreenTimeRecord, error) {
	if f.Err != nil {
		return nil, f.Err
	}
	if rec, ok := f.Records[f.key(userID, date)]; ok {
		return &rec, nil
	}
	return nil, nil
}

func (f *FakeRepository) ListByRange(ctx context.Context, userID string, fromDate, toDate time.Time) ([]screentime.ScreenTimeRecord, error) {
	if f.Err != nil {
		return nil, f.Err
	}
	var results []screentime.ScreenTimeRecord
	for _, rec := range f.Records {
		if rec.UserID == userID && !rec.Date.Before(fromDate) && !rec.Date.After(toDate) {
			results = append(results, rec)
		}
	}
	return results, nil
}
