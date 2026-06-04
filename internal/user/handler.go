package user

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

type userResponse struct {
	UserID                     string  `json:"userId"`
	Name                       string  `json:"name"`
	PresidentName              string  `json:"presidentName"`
	TargetEntertainmentMinutes int     `json:"targetEntertainmentMinutes"`
	SelectedCharacterID        *string `json:"selectedCharacterId"`
	CreatedAt                  string  `json:"createdAt"`
	UpdatedAt                  string  `json:"updatedAt"`
}

func toResponse(u User) userResponse {
	return userResponse{
		UserID:                     u.ID,
		Name:                       u.Name,
		PresidentName:              u.PresidentName,
		TargetEntertainmentMinutes: u.TargetEntertainmentMinutes,
		SelectedCharacterID:        u.SelectedCharacterID,
		CreatedAt:                  u.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:                  u.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name                       string  `json:"name"`
		PresidentName              string  `json:"presidentName"`
		TargetEntertainmentMinutes int     `json:"targetEntertainmentMinutes"`
		SelectedCharacterID        *string `json:"selectedCharacterId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "リクエストボディが正しくありません")
		return
	}

	u, err := h.svc.Create(r.Context(), CreateInput{
		Name:                       body.Name,
		PresidentName:              body.PresidentName,
		TargetEntertainmentMinutes: body.TargetEntertainmentMinutes,
		SelectedCharacterID:        body.SelectedCharacterID,
	})
	if err != nil {
		switch {
		case errors.Is(err, ErrBadInput):
			writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		case errors.Is(err, ErrCharacterNotFound):
			writeError(w, http.StatusBadRequest, "CHARACTER_NOT_FOUND", "指定したキャラクターが存在しません")
		default:
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "ユーザの作成に失敗しました")
		}
		return
	}

	writeJSON(w, http.StatusCreated, apiResponse{Success: true, Data: toResponse(u)})
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "userId")

	u, err := h.svc.FindByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "ユーザが見つかりません")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "ユーザ情報の取得に失敗しました")
		return
	}

	writeJSON(w, http.StatusOK, apiResponse{Success: true, Data: toResponse(u)})
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "userId")

	var body struct {
		Name                       *string `json:"name"`
		PresidentName              *string `json:"presidentName"`
		TargetEntertainmentMinutes *int    `json:"targetEntertainmentMinutes"`
		SelectedCharacterID        *string `json:"selectedCharacterId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "リクエストボディが正しくありません")
		return
	}

	u, err := h.svc.Update(r.Context(), id, UpdateInput{
		Name:                       body.Name,
		PresidentName:              body.PresidentName,
		TargetEntertainmentMinutes: body.TargetEntertainmentMinutes,
		SelectedCharacterID:        body.SelectedCharacterID,
	})
	if err != nil {
		switch {
		case errors.Is(err, ErrNotFound):
			writeError(w, http.StatusNotFound, "NOT_FOUND", "ユーザが見つかりません")
		case errors.Is(err, ErrBadInput):
			writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		case errors.Is(err, ErrCharacterNotFound):
			writeError(w, http.StatusBadRequest, "CHARACTER_NOT_FOUND", "指定したキャラクターが存在しません")
		default:
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "ユーザ設定の更新に失敗しました")
		}
		return
	}

	writeJSON(w, http.StatusOK, apiResponse{Success: true, Data: toResponse(u)})
}

func (h *Handler) UpdateSelectedCharacter(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userId")

	var body struct {
		CharacterID string `json:"characterId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "リクエストボディが正しくありません")
		return
	}
	if body.CharacterID == "" {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "characterId は必須です")
		return
	}

	u, err := h.svc.UpdateSelectedCharacter(r.Context(), userID, body.CharacterID)
	if err != nil {
		switch {
		case errors.Is(err, ErrNotFound):
			writeError(w, http.StatusNotFound, "NOT_FOUND", "ユーザが見つかりません")
		case errors.Is(err, ErrCharacterNotFound):
			writeError(w, http.StatusBadRequest, "CHARACTER_NOT_FOUND", "指定したキャラクターが存在しません")
		default:
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "選択キャラの更新に失敗しました")
		}
		return
	}

	writeJSON(w, http.StatusOK, apiResponse{
		Success: true,
		Data: map[string]string{
			"userId":              u.ID,
			"selectedCharacterId": *u.SelectedCharacterID,
		},
	})
}
