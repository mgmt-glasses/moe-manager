package voiceadapter

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/mgmt-glasses/moe-manager/internal/voice"
)

// LocalAudioStorage stores MP3 files under a configurable base directory.
type LocalAudioStorage struct {
	baseDir string
}

func NewLocalAudioStorage(baseDir string) *LocalAudioStorage {
	return &LocalAudioStorage{baseDir: baseDir}
}

func (s *LocalAudioStorage) Save(_ context.Context, voiceFileID string, audio []byte) error {
	if err := os.MkdirAll(s.baseDir, 0755); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}
	path := filepath.Join(s.baseDir, voiceFileID+".mp3")
	return os.WriteFile(path, audio, 0644)
}

func (s *LocalAudioStorage) Open(_ context.Context, voiceFileID string) (io.ReadCloser, error) {
	path := filepath.Join(s.baseDir, voiceFileID+".mp3")
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, voice.ErrNotFound
		}
		return nil, fmt.Errorf("open audio: %w", err)
	}
	return f, nil
}

func (s *LocalAudioStorage) Delete(_ context.Context, voiceFileID string) error {
	path := filepath.Join(s.baseDir, voiceFileID+".mp3")
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("delete audio: %w", err)
	}
	return nil
}
