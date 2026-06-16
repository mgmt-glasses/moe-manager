package voice

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/mgmt-glasses/moe-manager/internal/shared"
)

type voiceService interface {
	Generate(ctx context.Context, userID, characterID, voicePresetID, text string) (VoiceFile, error)
	Open(ctx context.Context, userID, voiceFileID string) (io.ReadCloser, error)
}

type Handler struct {
	svc voiceService
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
	userID := chi.URLParam(r, "userId")

	var body generateRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		shared.WriteError(w, http.StatusBadRequest, "INVALID_BODY", "リクエストボディが正しくありません")
		return
	}
	if body.Text == "" || body.CharacterID == "" || body.VoicePresetID == "" {
		shared.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "characterId, voicePresetId, text は必須です")
		return
	}

	vf, err := h.svc.Generate(r.Context(), userID, body.CharacterID, body.VoicePresetID, body.Text)
	if err != nil {
		switch {
		case errors.Is(err, ErrForbidden):
			shared.WriteError(w, http.StatusForbidden, "FORBIDDEN", "指定されたキャラクターはこのユーザーに選択されていません")
		case errors.Is(err, ErrUserNotFound):
			shared.WriteError(w, http.StatusNotFound, "NOT_FOUND", "ユーザーが見つかりません")
		default:
			shared.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "音声生成に失敗しました")
		}
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
	userID := chi.URLParam(r, "userId")
	voiceFileID := chi.URLParam(r, "voiceFileId")

	rc, err := h.svc.Open(r.Context(), userID, voiceFileID)
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
	if _, err := io.Copy(w, rc); err != nil {
		log.Printf("voice GetAudio: stream error for %s: %v", voiceFileID, err)
	}
}
