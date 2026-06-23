package screentime

import "time"

type ScreenTimeRecord struct {
	RecordID            string
	UserID              string
	Date                time.Time
	Minutes             int
	TargetMinutes       int
	DiffMinutes         int
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

func (r *ScreenTimeRecord) CalculateDiff() {
	r.DiffMinutes = r.Minutes - r.TargetMinutes
}
