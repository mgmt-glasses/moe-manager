package voiceadapter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/mgmt-glasses/moe-manager/internal/voice"
)

const defaultTTSTimeout = 60 * time.Second

// TTSHTTPClient calls the Python voice-library service (POST /generate → MP3 bytes).
type TTSHTTPClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewTTSHTTPClient creates a client. timeout <= 0 falls back to defaultTTSTimeout.
// Cloud Run の scale-to-zero では TTS のコールドスタート（GPU 割当 + モデルロード）に
// 数十秒かかるため、呼び出し側で十分な timeout を渡せるようにする。
func NewTTSHTTPClient(baseURL string, timeout time.Duration) *TTSHTTPClient {
	if timeout <= 0 {
		timeout = defaultTTSTimeout
	}
	return &TTSHTTPClient{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: timeout},
	}
}

type generateRequest struct {
	Text     string `json:"text"`
	PresetID string `json:"preset_id,omitempty"`
}

func (c *TTSHTTPClient) Synthesize(ctx context.Context, input voice.SynthesisInput) ([]byte, error) {
	body, err := json.Marshal(generateRequest{
		Text:     input.Text,
		PresetID: input.VoicePresetID,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/generate", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call tts service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("tts service returned %d: %s", resp.StatusCode, string(b))
	}

	const maxAudioBytes = 50 * 1024 * 1024 // 50 MB
	audio, err := io.ReadAll(io.LimitReader(resp.Body, maxAudioBytes))
	if err != nil {
		return nil, fmt.Errorf("read audio: %w", err)
	}
	return audio, nil
}
