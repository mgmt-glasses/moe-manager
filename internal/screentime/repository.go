package screentime

import (
	"context"
	"time"
)

type Repository interface {
	Save(ctx context.Context, record ScreenTimeRecord) error
	GetByDate(ctx context.Context, userID string, date time.Time) (*ScreenTimeRecord, error)
	ListByRange(ctx context.Context, userID string, fromDate, toDate time.Time) ([]ScreenTimeRecord, error)
}
