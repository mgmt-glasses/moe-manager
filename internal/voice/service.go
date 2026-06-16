package voice

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNotFound    = errors.New("voice file not found")
	ErrUserNotFound = errors.New("user not found")
	ErrForbidden   = errors.New("character does not belong to user")
)

type Service struct {
	synthesizer  Synthesizer
	storage      AudioStorage
	repo         VoiceFileRepository
	userChecker  UserCharacterChecker
}

func NewService(synthesizer Synthesizer, storage AudioStorage, repo VoiceFileRepository, userChecker UserCharacterChecker) *Service {
	return &Service{synthesizer: synthesizer, storage: storage, repo: repo, userChecker: userChecker}
}

func (s *Service) Generate(ctx context.Context, userID, characterID, voicePresetID, text string) (VoiceFile, error) {
	selectedCharID, err := s.userChecker.GetSelectedCharacterID(ctx, userID)
	if err != nil {
		return VoiceFile{}, fmt.Errorf("check user: %w", err)
	}
	if selectedCharID != characterID {
		return VoiceFile{}, ErrForbidden
	}

	audio, err := s.synthesizer.Synthesize(ctx, SynthesisInput{
		Text:          text,
		VoicePresetID: voicePresetID,
	})
	if err != nil {
		return VoiceFile{}, fmt.Errorf("synthesize: %w", err)
	}

	vf := VoiceFile{
		ID:          uuid.NewString(),
		CharacterID: characterID,
		SourceText:  text,
		CreatedAt:   time.Now(),
	}

	if err := s.storage.Save(ctx, vf.ID, audio); err != nil {
		return VoiceFile{}, fmt.Errorf("save audio: %w", err)
	}
	if err := s.repo.Save(ctx, vf); err != nil {
		// 補償削除: DB 保存失敗時にストレージから孤立ファイルを除去
		_ = s.storage.Delete(ctx, vf.ID)
		return VoiceFile{}, fmt.Errorf("save voice file record: %w", err)
	}

	return vf, nil
}

func (s *Service) Open(ctx context.Context, voiceFileID string) (io.ReadCloser, error) {
	if _, err := s.repo.FindByID(ctx, voiceFileID); err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("find voice file: %w", err)
	}
	rc, err := s.storage.Open(ctx, voiceFileID)
	if err != nil {
		return nil, err
	}
	return rc, nil
}
