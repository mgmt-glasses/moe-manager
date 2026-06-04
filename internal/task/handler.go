package task

import (
	"encoding/json"
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

type taskResponse struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Status      string  `json:"status"`
	CreatedAt   string  `json:"createdAt"`
	CompletedAt *string `json:"completedAt"`
}

func toResponse(t Task) taskResponse {
	r := taskResponse{
		ID:        t.ID,
		Title:     t.Title,
		Status:    string(t.Status),
		CreatedAt: t.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	if t.CompletedAt != nil {
		s := t.CompletedAt.Format("2006-01-02T15:04:05Z07:00")
		r.CompletedAt = &s
	}
	return r
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userId")

	var body struct {
		Title string `json:"title"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		shared.WriteError(w, http.StatusBadRequest, "INVALID_BODY", "リクエストボディが正しくありません")
		return
	}

	t, err := h.svc.Create(r.Context(), CreateInput{UserID: userID, Title: body.Title})
	if err != nil {
		if errors.Is(err, ErrBadInput) {
			shared.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "タイトルは必須です")
			return
		}
		shared.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "タスクの作成に失敗しました")
		return
	}

	shared.WriteJSON(w, http.StatusCreated, shared.Response{Success: true, Data: toResponse(t)})
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userId")

	tasks, err := h.svc.List(r.Context(), userID)
	if err != nil {
		shared.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "タスク一覧の取得に失敗しました")
		return
	}

	items := make([]taskResponse, len(tasks))
	for i, t := range tasks {
		items[i] = toResponse(t)
	}
	shared.WriteJSON(w, http.StatusOK, shared.Response{Success: true, Data: items})
}

func (h *Handler) Complete(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userId")
	taskID := chi.URLParam(r, "taskId")

	t, err := h.svc.Complete(r.Context(), userID, taskID)
	if err != nil {
		switch {
		case errors.Is(err, ErrNotFound):
			shared.WriteError(w, http.StatusNotFound, "NOT_FOUND", "タスクが見つかりません")
		case errors.Is(err, ErrForbidden):
			shared.WriteError(w, http.StatusForbidden, "FORBIDDEN", "このタスクを操作する権限がありません")
		default:
			shared.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "タスクの完了に失敗しました")
		}
		return
	}

	shared.WriteJSON(w, http.StatusOK, shared.Response{Success: true, Data: toResponse(t)})
}

func (h *Handler) Reopen(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userId")
	taskID := chi.URLParam(r, "taskId")

	t, err := h.svc.Reopen(r.Context(), userID, taskID)
	if err != nil {
		switch {
		case errors.Is(err, ErrNotFound):
			shared.WriteError(w, http.StatusNotFound, "NOT_FOUND", "タスクが見つかりません")
		case errors.Is(err, ErrForbidden):
			shared.WriteError(w, http.StatusForbidden, "FORBIDDEN", "このタスクを操作する権限がありません")
		default:
			shared.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "タスクの再オープンに失敗しました")
		}
		return
	}

	shared.WriteJSON(w, http.StatusOK, shared.Response{Success: true, Data: toResponse(t)})
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userId")
	taskID := chi.URLParam(r, "taskId")

	if err := h.svc.Delete(r.Context(), userID, taskID); err != nil {
		switch {
		case errors.Is(err, ErrNotFound):
			shared.WriteError(w, http.StatusNotFound, "NOT_FOUND", "タスクが見つかりません")
		case errors.Is(err, ErrForbidden):
			shared.WriteError(w, http.StatusForbidden, "FORBIDDEN", "このタスクを操作する権限がありません")
		default:
			shared.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "タスクの削除に失敗しました")
		}
		return
	}

	shared.WriteJSON(w, http.StatusOK, shared.Response{Success: true, Data: nil})
}
