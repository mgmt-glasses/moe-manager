package chat

import "time"

type CharacterProfile struct {
	ID          string
	Name        string
	MBTI        string
	Personality string
	SpeechStyle string
}

// Persona holds GCP-managed AI configuration for a character.
// Fetched via PersonaRepository and applied to the LLM system instruction.
type Persona struct {
	CharacterID       string
	SystemInstruction string
}

type TaskSummary struct {
	CompletedCount int
	TotalCount     int
}

type ScreenTimeSummary struct {
	TodayMinutes  int
	TargetMinutes int
}

// ChatContext holds all context needed to build an LLM prompt.
type ChatContext struct {
	Character  CharacterProfile
	Tasks      TaskSummary
	ScreenTime ScreenTimeSummary
}

// Prompt is the assembled LLM prompt sent to the LLMClient.
type Prompt struct {
	System string
	User   string
}

// Message is a single chat turn.
type Message struct {
	Role        string // "user" or "assistant"
	CharacterID string
	Content     string
	CreatedAt   time.Time
}

// SendMessageInput is the input to Service.SendMessage.
type SendMessageInput struct {
	UserID  string
	Message string
}

// SendMessageOutput is the result of Service.SendMessage.
type SendMessageOutput struct {
	UserMessage      Message
	AssistantMessage Message
	Context          ChatContext
}
