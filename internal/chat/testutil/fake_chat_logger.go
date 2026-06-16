package testutil

import (
	"context"

	"github.com/mgmt-glasses/moe-manager/internal/chat"
)

// FakeChatLogger is an in-memory test double for chat.ChatLogger.
type FakeChatLogger struct {
	Saved []chat.Message
	Err   error
}

func (f *FakeChatLogger) Save(_ context.Context, _ string, msg chat.Message) error {
	if f.Err != nil {
		return f.Err
	}
	f.Saved = append(f.Saved, msg)
	return nil
}
