package character

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/mgmt-glasses/moe-manager/internal/shared"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
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
		shared.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "キャラ一覧の取得に失敗しました")
		return
	}

	items := make([]characterListItem, len(chars))
	for i, c := range chars {
		items[i] = toListItem(c)
	}
	shared.WriteJSON(w, http.StatusOK, shared.Response{Success: true, Data: items})
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "characterId")

	c, err := h.svc.FindByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			shared.WriteError(w, http.StatusNotFound, "NOT_FOUND", "キャラクターが見つかりません")
			return
		}
		shared.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "キャラ詳細の取得に失敗しました")
		return
	}

	shared.WriteJSON(w, http.StatusOK, shared.Response{Success: true, Data: toDetail(c)})
}
