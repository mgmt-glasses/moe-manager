package voice

import "time"

type VoiceFile struct {
	ID          string
	UserID      string
	CharacterID string
	SourceText  string
	CreatedAt   time.Time
}

type SynthesisInput struct {
	Text         string
	VoicePresetID string
}
