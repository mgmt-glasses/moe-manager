package character

import "time"

type Character struct {
	ID                   string
	Name                 string
	MBTIType             string
	PersonalityDesc      string
	SpeechStyle          string
	SystemPromptFragment string
	VoicePresetID        string
	IconPath             string
	StandingImagePath    string
	SampleVoicePath      string
	CreatedAt            time.Time
	UpdatedAt            time.Time
}
