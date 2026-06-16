package voice

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/mgmt-glasses/moe-manager/internal/shared"
)

type voiceGenerator interface {
	Generate(ctx context.Context, characterID, voicePresetID, text string) (VoiceFile, error)
}

type voiceOpener interface {
	Open(ctx context.Context, voiceFileID string) (io.ReadCloser, error)
}

type Handler struct {
	svc interface {
		voiceGenerator
		voiceOpener
	}
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

type generateRequest struct {
	CharacterID   string `json:"characterId"`
	VoicePresetID string `json:"voicePresetId"`
	Text          string `json:"text"`
}

type voiceFileResponse struct {
	ID          string `json:"id"`
	CharacterID string `json:"characterId"`
	SourceText  string `json:"sourceText"`
	CreatedAt   string `json:"createdAt"`
}

func (h *Handler) Generate(w http.ResponseWriter, r *http.Request) {
	var body generateRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		shared.WriteError(w, http.StatusBadRequest, "INVALID_BODY", "リクエストボディが正しくありません")
		return
	}
	if body.Text == "" || body.CharacterID == "" || body.VoicePresetID == "" {
		shared.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "characterId, voicePresetId, text は必須です")
		return
	}

	vf, err := h.svc.Generate(r.Context(), body.CharacterID, body.VoicePresetID, body.Text)
	if err != nil {
		shared.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "音声生成に失敗しました")
		return
	}

	shared.WriteJSON(w, http.StatusCreated, shared.Response{
		Success: true,
		Data: voiceFileResponse{
			ID:          vf.ID,
			CharacterID: vf.CharacterID,
			SourceText:  vf.SourceText,
			CreatedAt:   vf.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		},
	})
}

func (h *Handler) GetAudio(w http.ResponseWriter, r *http.Request) {
	voiceFileID := chi.URLParam(r, "voiceFileId")

	rc, err := h.svc.Open(r.Context(), voiceFileID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			shared.WriteError(w, http.StatusNotFound, "NOT_FOUND", "音声ファイルが見つかりません")
			return
		}
		shared.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "音声ファイルの取得に失敗しました")
		return
	}
	defer rc.Close()

	w.Header().Set("Content-Type", "audio/wav")
	w.WriteHeader(http.StatusOK)
	io.Copy(w, rc)
}
