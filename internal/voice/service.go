package voice

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"
)

var ErrNotFound = errors.New("voice file not found")

type Service struct {
	synthesizer Synthesizer
	storage     AudioStorage
	repo        VoiceFileRepository
}

func NewService(synthesizer Synthesizer, storage AudioStorage, repo VoiceFileRepository) *Service {
	return &Service{synthesizer: synthesizer, storage: storage, repo: repo}
}

func (s *Service) Generate(ctx context.Context, characterID, voicePresetID, text string) (VoiceFile, error) {
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
		return VoiceFile{}, fmt.Errorf("save voice file record: %w", err)
	}

	return vf, nil
}

func (s *Service) Open(ctx context.Context, voiceFileID string) (io.ReadCloser, error) {
	if _, err := s.repo.FindByID(ctx, voiceFileID); err != nil {
		return nil, ErrNotFound
	}
	rc, err := s.storage.Open(ctx, voiceFileID)
	if err != nil {
		return nil, fmt.Errorf("open audio: %w", err)
	}
	return rc, nil
}
