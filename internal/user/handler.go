package user

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	appauth "github.com/mgmt-glasses/moe-manager/internal/auth"
	"github.com/mgmt-glasses/moe-manager/internal/shared"
)

type userServicer interface {
	Create(ctx context.Context, in CreateInput) (User, error)
	FindByID(ctx context.Context, id string) (User, error)
	Update(ctx context.Context, id string, in UpdateInput) (User, error)
	UpdateSelectedCharacter(ctx context.Context, userID, characterID string) (User, error)
}

type Handler struct {
	svc userServicer
}

func NewHandler(svc userServicer) *Handler {
	return &Handler{svc: svc}
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
		shared.WriteError(w, http.StatusBadRequest, "INVALID_BODY", "リクエストボディが正しくありません")
		return
	}

	authUser, ok := appauth.UserFromContext(r.Context())
	if !ok || authUser.UID == "" {
		// RequireAuth 配下のため通常は到達しない。万一認証情報が無ければ
		// 匿名作成を許さず弾く（fail-closed）。
		shared.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "認証情報を取得できませんでした")
		return
	}

	u, err := h.svc.Create(r.Context(), CreateInput{
		UserID:                     authUser.UID,
		Name:                       body.Name,
		PresidentName:              body.PresidentName,
		TargetEntertainmentMinutes: body.TargetEntertainmentMinutes,
		SelectedCharacterID:        body.SelectedCharacterID,
	})
	if err != nil {
		switch {
		case errors.Is(err, ErrBadInput):
			shared.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "入力内容が正しくありません")
		case errors.Is(err, ErrCharacterNotFound):
			shared.WriteError(w, http.StatusBadRequest, "CHARACTER_NOT_FOUND", "指定したキャラクターが存在しません")
		case errors.Is(err, ErrAlreadyExists):
			shared.WriteError(w, http.StatusConflict, "ALREADY_EXISTS", "ユーザは既に登録されています")
		default:
			shared.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "ユーザの作成に失敗しました")
		}
		return
	}

	shared.WriteJSON(w, http.StatusCreated, shared.Response{Success: true, Data: toResponse(u)})
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "userId")

	u, err := h.svc.FindByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			shared.WriteError(w, http.StatusNotFound, "NOT_FOUND", "ユーザが見つかりません")
			return
		}
		shared.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "ユーザ情報の取得に失敗しました")
		return
	}

	shared.WriteJSON(w, http.StatusOK, shared.Response{Success: true, Data: toResponse(u)})
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
		shared.WriteError(w, http.StatusBadRequest, "INVALID_BODY", "リクエストボディが正しくありません")
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
			shared.WriteError(w, http.StatusNotFound, "NOT_FOUND", "ユーザが見つかりません")
		case errors.Is(err, ErrBadInput):
			shared.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "入力内容が正しくありません")
		case errors.Is(err, ErrCharacterNotFound):
			shared.WriteError(w, http.StatusBadRequest, "CHARACTER_NOT_FOUND", "指定したキャラクターが存在しません")
		default:
			shared.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "ユーザ設定の更新に失敗しました")
		}
		return
	}

	shared.WriteJSON(w, http.StatusOK, shared.Response{Success: true, Data: toResponse(u)})
}

func (h *Handler) UpdateSelectedCharacter(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userId")

	var body struct {
		CharacterID string `json:"characterId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		shared.WriteError(w, http.StatusBadRequest, "INVALID_BODY", "リクエストボディが正しくありません")
		return
	}
	if body.CharacterID == "" {
		shared.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "characterId は必須です")
		return
	}

	u, err := h.svc.UpdateSelectedCharacter(r.Context(), userID, body.CharacterID)
	if err != nil {
		switch {
		case errors.Is(err, ErrNotFound):
			shared.WriteError(w, http.StatusNotFound, "NOT_FOUND", "ユーザが見つかりません")
		case errors.Is(err, ErrCharacterNotFound):
			shared.WriteError(w, http.StatusBadRequest, "CHARACTER_NOT_FOUND", "指定したキャラクターが存在しません")
		default:
			shared.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "選択キャラの更新に失敗しました")
		}
		return
	}

	selectedCharacterID := ""
	if u.SelectedCharacterID != nil {
		selectedCharacterID = *u.SelectedCharacterID
	}
	shared.WriteJSON(w, http.StatusOK, shared.Response{
		Success: true,
		Data: map[string]string{
			"userId":              u.ID,
			"selectedCharacterId": selectedCharacterID,
		},
	})
}
