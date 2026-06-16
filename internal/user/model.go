package user

import "time"

type User struct {
	ID                        string
	Name                      string
	PresidentName             string
	SelectedCharacterID       *string
	TargetEntertainmentMinutes int
	CreatedAt                 time.Time
	UpdatedAt                 time.Time
}
