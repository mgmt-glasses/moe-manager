package chat_test

import (
	"context"
	"errors"
	"testing"

	"moe-manager/internal/chat"
	"moe-manager/internal/chat/testutil"
)

var baseCharacter = chat.CharacterProfile{
	ID:          "char_001",
	Name:        "さくら",
	MBTI:        "ISTJ",
	Personality: "真面目で責任感が強い",
	SpeechStyle: "丁寧で落ち着いた口調",
}

var baseContext = chat.ChatContext{
	Character:  baseCharacter,
	Tasks:      chat.TaskSummary{CompletedCount: 3, TotalCount: 5},
	ScreenTime: chat.ScreenTimeSummary{TodayMinutes: 180, TargetMinutes: 120},
}

func TestSendMessage_OK(t *testing.T) {
	loader := &testutil.FakeContextLoader{Ctx: baseContext}
	llm := &testutil.FakeLLMClient{Reply: "了解しました、社長。"}
	logger := &testutil.FakeChatLogger{}

	svc := chat.NewService(llm, loader, logger)
	out, err := svc.SendMessage(context.Background(), chat.SendMessageInput{
		UserID:  "user_001",
		Message: "今日はゲームしすぎた",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.AssistantMessage.Content != "了解しました、社長。" {
		t.Errorf("reply: got %q, want %q", out.AssistantMessage.Content, "了解しました、社長。")
	}
	if out.UserMessage.Content != "今日はゲームしすぎた" {
		t.Errorf("user message not preserved")
	}
	if out.Context.Character.MBTI != "ISTJ" {
		t.Errorf("character MBTI not in context")
	}
	if out.Context.Tasks.CompletedCount != 3 {
		t.Errorf("task context missing: got %d", out.Context.Tasks.CompletedCount)
	}
	if out.Context.ScreenTime.TodayMinutes != 180 {
		t.Errorf("screen time context missing: got %d", out.Context.ScreenTime.TodayMinutes)
	}
}

func TestSendMessage_LoggerReceivesBothMessages(t *testing.T) {
	loader := &testutil.FakeContextLoader{Ctx: baseContext}
	llm := &testutil.FakeLLMClient{Reply: "お疲れ様です。"}
	logger := &testutil.FakeChatLogger{}

	svc := chat.NewService(llm, loader, logger)
	_, err := svc.SendMessage(context.Background(), chat.SendMessageInput{
		UserID:  "user_001",
		Message: "hello",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(logger.Saved) != 2 {
		t.Errorf("expected 2 saved messages, got %d", len(logger.Saved))
	}
	if logger.Saved[0].Role != "user" {
		t.Errorf("first saved message role: got %q, want %q", logger.Saved[0].Role, "user")
	}
	if logger.Saved[1].Role != "assistant" {
		t.Errorf("second saved message role: got %q, want %q", logger.Saved[1].Role, "assistant")
	}
}

func TestSendMessage_NilLoggerSkipsPersistence(t *testing.T) {
	loader := &testutil.FakeContextLoader{Ctx: baseContext}
	llm := &testutil.FakeLLMClient{Reply: "ok"}

	svc := chat.NewService(llm, loader, nil)
	_, err := svc.SendMessage(context.Background(), chat.SendMessageInput{
		UserID:  "user_001",
		Message: "hello",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSendMessage_NoCharacter(t *testing.T) {
	loader := &testutil.FakeContextLoader{Ctx: chat.ChatContext{}} // no character
	llm := &testutil.FakeLLMClient{}
	svc := chat.NewService(llm, loader, nil)

	_, err := svc.SendMessage(context.Background(), chat.SendMessageInput{
		UserID:  "user_001",
		Message: "hello",
	})
	if !errors.Is(err, chat.ErrNoCharacterSelected) {
		t.Errorf("expected ErrNoCharacterSelected, got %v", err)
	}
}

func TestSendMessage_LLMError(t *testing.T) {
	loader := &testutil.FakeContextLoader{Ctx: baseContext}
	llm := &testutil.FakeLLMClient{Err: errors.New("LLM timeout")}
	svc := chat.NewService(llm, loader, nil)

	_, err := svc.SendMessage(context.Background(), chat.SendMessageInput{
		UserID:  "user_001",
		Message: "hello",
	})
	if err == nil {
		t.Fatal("expected error from LLM, got nil")
	}
}

func TestSendMessage_ContextLoadError(t *testing.T) {
	loader := &testutil.FakeContextLoader{Err: errors.New("DB unavailable")}
	llm := &testutil.FakeLLMClient{}
	svc := chat.NewService(llm, loader, nil)

	_, err := svc.SendMessage(context.Background(), chat.SendMessageInput{
		UserID:  "user_001",
		Message: "hello",
	})
	if err == nil {
		t.Fatal("expected error from context loader, got nil")
	}
}
