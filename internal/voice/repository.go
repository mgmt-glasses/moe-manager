package voice

import (
	"context"
	"io"
)

type Synthesizer interface {
	Synthesize(ctx context.Context, input SynthesisInput) ([]byte, error)
}

type AudioStorage interface {
	Save(ctx context.Context, voiceFileID string, audio []byte) error
	Open(ctx context.Context, voiceFileID string) (io.ReadCloser, error)
}

type VoiceFileRepository interface {
	Save(ctx context.Context, vf VoiceFile) error
	FindByID(ctx context.Context, id string) (VoiceFile, error)
}
