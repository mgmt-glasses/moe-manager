package character

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

type apiResponse struct {
	Success bool      `json:"success"`
	Data    any       `json:"data"`
	Error   *apiError `json:"error"`
}

type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, apiResponse{
		Success: false,
		Data:    nil,
		Error:   &apiError{Code: code, Message: message},
	})
}

type characterListItem struct {
	CharacterID      string `json:"characterId"`
	Name             string `json:"name"`
	MBTI             string `json:"mbti"`
	Description      string `json:"description"`
	Tone             string `json:"tone"`
	VoiceID          string `json:"voiceId"`
	IconURL          string `json:"iconUrl"`
	StandingImageURL string `json:"standingImageUrl"`
}

type characterDetail struct {
	CharacterID      string `json:"characterId"`
	Name             string `json:"name"`
	MBTI             string `json:"mbti"`
	Description      string `json:"description"`
	Tone             string `json:"tone"`
	VoiceID          string `json:"voiceId"`
	IconURL          string `json:"iconUrl"`
	StandingImageURL string `json:"standingImageUrl"`
	SampleVoiceURL   string `json:"sampleVoiceUrl"`
}

func toAssetURL(path string) string {
	if path == "" {
		return ""
	}
	return "/assets/" + path
}

func toListItem(c Character) characterListItem {
	return characterListItem{
		CharacterID:      c.ID,
		Name:             c.Name,
		MBTI:             c.MBTIType,
		Description:      c.PersonalityDesc,
		Tone:             c.SpeechStyle,
		VoiceID:          c.VoicePresetID,
		IconURL:          toAssetURL(c.IconPath),
		StandingImageURL: toAssetURL(c.StandingImagePath),
	}
}

func toDetail(c Character) characterDetail {
	return characterDetail{
		CharacterID:      c.ID,
		Name:             c.Name,
		MBTI:             c.MBTIType,
		Description:      c.PersonalityDesc,
		Tone:             c.SpeechStyle,
		VoiceID:          c.VoicePresetID,
		IconURL:          toAssetURL(c.IconPath),
		StandingImageURL: toAssetURL(c.StandingImagePath),
		SampleVoiceURL:   toAssetURL(c.SampleVoicePath),
	}
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	chars, err := h.svc.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "キャラ一覧の取得に失敗しました")
		return
	}

	items := make([]characterListItem, len(chars))
	for i, c := range chars {
		items[i] = toListItem(c)
	}
	writeJSON(w, http.StatusOK, apiResponse{Success: true, Data: items})
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "characterId")

	c, err := h.svc.FindByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "キャラクターが見つかりません")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "キャラ詳細の取得に失敗しました")
		return
	}

	writeJSON(w, http.StatusOK, apiResponse{Success: true, Data: toDetail(c)})
}
