package chat

import "context"

// LLMClient generates a reply from a prompt.
type LLMClient interface {
	GenerateReply(ctx context.Context, prompt Prompt) (string, error)
}

// ContextLoader loads the chat context (selected character, tasks, screen time) for a user.
type ContextLoader interface {
	Load(ctx context.Context, userID string) (ChatContext, error)
}

// PersonaRepository fetches GCP-managed AI configuration for a character.
// Firestore adapter implementing this interface will be added after GCP PoC (issue 02).
type PersonaRepository interface {
	Get(ctx context.Context, characterID string) (Persona, error)
}

// ChatLogger persists chat messages.
// PostgreSQL adapter implementing this interface will be added in issue 07-chat-logs.md.
type ChatLogger interface {
	Save(ctx context.Context, userID string, msg Message) error
}
