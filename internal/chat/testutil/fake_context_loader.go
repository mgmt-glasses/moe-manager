package testutil

import (
	"context"

	"moe-manager/internal/chat"
)

// FakeContextLoader is a test double for chat.ContextLoader.
type FakeContextLoader struct {
	Ctx chat.ChatContext
	Err error
}

func (f *FakeContextLoader) Load(_ context.Context, _ string) (chat.ChatContext, error) {
	return f.Ctx, f.Err
}
