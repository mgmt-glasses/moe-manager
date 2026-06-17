package testutil

import (
	"context"

	"github.com/mgmt-glasses/moe-manager/internal/chat"
)

// FakeLLMClient is a test double for chat.LLMClient.
type FakeLLMClient struct {
	Reply string
	Err   error
}

func (f *FakeLLMClient) GenerateReply(_ context.Context, _ chat.Prompt) (string, error) {
	return f.Reply, f.Err
}
