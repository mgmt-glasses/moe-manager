package chat

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// ErrNoCharacterSelected is returned when the user has no selected character.
var ErrNoCharacterSelected = errors.New("no character selected")

// Service handles chat message generation.
type Service struct {
	llm    LLMClient
	loader ContextLoader
	logger ChatLogger // nil means no persistence
}

// NewService creates a Service. logger may be nil to skip persistence.
func NewService(llm LLMClient, loader ContextLoader, logger ChatLogger) *Service {
	return &Service{llm: llm, loader: loader, logger: logger}
}

// SendMessage loads user context, generates an AI reply, and optionally logs the exchange.
func (s *Service) SendMessage(ctx context.Context, input SendMessageInput) (SendMessageOutput, error) {
	chatCtx, err := s.loader.Load(ctx, input.UserID)
	if err != nil {
		return SendMessageOutput{}, fmt.Errorf("load context: %w", err)
	}
	if chatCtx.Character.ID == "" {
		return SendMessageOutput{}, ErrNoCharacterSelected
	}

	prompt := buildPrompt(chatCtx, input.Message)

	reply, err := s.llm.GenerateReply(ctx, prompt)
	if err != nil {
		return SendMessageOutput{}, fmt.Errorf("generate reply: %w", err)
	}

	now := time.Now()
	userMsg := Message{Role: "user", Content: input.Message, CreatedAt: now}
	assistantMsg := Message{Role: "assistant", Content: reply, CreatedAt: now}

	if s.logger != nil {
		_ = s.logger.Save(ctx, input.UserID, userMsg)
		_ = s.logger.Save(ctx, input.UserID, assistantMsg)
	}

	return SendMessageOutput{
		UserMessage:      userMsg,
		AssistantMessage: assistantMsg,
		Context:          chatCtx,
	}, nil
}

func buildPrompt(chatCtx ChatContext, userMessage string) Prompt {
	c := chatCtx.Character
	t := chatCtx.Tasks
	st := chatCtx.ScreenTime

	system := fmt.Sprintf(
		"あなたは秘書キャラ「%s」として振る舞ってください。\n\n"+
			"【キャラクター設定】\n"+
			"MBTI: %s\n"+
			"性格: %s\n"+
			"口調: %s\n\n"+
			"【ルール】\n"+
			"- ユーザを社長として扱う\n"+
			"- キャラクターの性格と口調を守る\n"+
			"- 返答は短く自然な会話にする\n"+
			"- 医療・法律・金融などの専門助言は避ける\n"+
			"- タスク管理アプリの秘書として振る舞う\n"+
			"- 必要以上に責めない\n\n"+
			"【今日の状況】\n"+
			"タスク: %d/%d 件完了\n"+
			"娯楽時間: %d分 (目標: %d分)",
		c.Name, c.MBTI, c.Personality, c.SpeechStyle,
		t.CompletedCount, t.TotalCount,
		st.TodayMinutes, st.TargetMinutes,
	)

	return Prompt{System: system, User: userMessage}
}
