package chatadapter

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/mgmt-glasses/moe-manager/internal/chat"
)

type PostgresChatContextLoader struct {
	db *sql.DB
}

func NewPostgresChatContextLoader(db *sql.DB) *PostgresChatContextLoader {
	return &PostgresChatContextLoader{db: db}
}

func (l *PostgresChatContextLoader) Load(ctx context.Context, userID string) (chat.ChatContext, error) {
	var selectedCharID sql.NullString
	var targetMinutes int

	err := l.db.QueryRowContext(ctx,
		`SELECT selected_character_id, target_entertainment_minutes FROM users WHERE id = $1`,
		userID,
	).Scan(&selectedCharID, &targetMinutes)
	if err != nil {
		return chat.ChatContext{}, fmt.Errorf("query user: %w", err)
	}

	if !selectedCharID.Valid {
		return chat.ChatContext{ScreenTime: chat.ScreenTimeSummary{TargetMinutes: targetMinutes}}, nil
	}

	var char chat.CharacterProfile
	err = l.db.QueryRowContext(ctx,
		`SELECT id, name, mbti_type, personality_desc, speech_style FROM characters WHERE id = $1`,
		selectedCharID.String,
	).Scan(&char.ID, &char.Name, &char.MBTI, &char.Personality, &char.SpeechStyle)
	if err != nil {
		return chat.ChatContext{}, fmt.Errorf("query character: %w", err)
	}

	tasks, err := l.loadTaskSummary(ctx, userID)
	if err != nil {
		return chat.ChatContext{}, err
	}

	screenTime, err := l.loadScreenTimeSummary(ctx, userID, targetMinutes)
	if err != nil {
		return chat.ChatContext{}, err
	}

	return chat.ChatContext{
		Character:  char,
		Tasks:      tasks,
		ScreenTime: screenTime,
	}, nil
}

func (l *PostgresChatContextLoader) loadTaskSummary(ctx context.Context, userID string) (chat.TaskSummary, error) {
	var completed, total int
	err := l.db.QueryRowContext(ctx,
		`SELECT
			COUNT(*) FILTER (WHERE status = 'done') AS completed,
			COUNT(*) AS total
		FROM tasks
		WHERE user_id = $1 AND deleted_at IS NULL`,
		userID,
	).Scan(&completed, &total)
	if err != nil {
		return chat.TaskSummary{}, fmt.Errorf("query tasks: %w", err)
	}
	return chat.TaskSummary{CompletedCount: completed, TotalCount: total}, nil
}

func (l *PostgresChatContextLoader) loadScreenTimeSummary(ctx context.Context, userID string, targetMinutes int) (chat.ScreenTimeSummary, error) {
	var minutes int
	err := l.db.QueryRowContext(ctx,
		`SELECT minutes FROM screentime_records WHERE user_id = $1 AND date = CURRENT_DATE`,
		userID,
	).Scan(&minutes)
	if err == sql.ErrNoRows {
		return chat.ScreenTimeSummary{TodayMinutes: 0, TargetMinutes: targetMinutes}, nil
	}
	if err != nil {
		return chat.ScreenTimeSummary{}, fmt.Errorf("query screentime: %w", err)
	}
	return chat.ScreenTimeSummary{TodayMinutes: minutes, TargetMinutes: targetMinutes}, nil
}
