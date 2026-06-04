package task

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
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "リクエストボディが正しくありません")
		return
	}

	t, err := h.svc.Create(r.Context(), CreateInput{UserID: userID, Title: body.Title})
	if err != nil {
		if errors.Is(err, ErrBadInput) {
			writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "タイトルは必須です")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "タスクの作成に失敗しました")
		return
	}

	writeJSON(w, http.StatusCreated, apiResponse{Success: true, Data: toResponse(t)})
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userId")

	tasks, err := h.svc.List(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "タスク一覧の取得に失敗しました")
		return
	}

	items := make([]taskResponse, len(tasks))
	for i, t := range tasks {
		items[i] = toResponse(t)
	}
	writeJSON(w, http.StatusOK, apiResponse{Success: true, Data: items})
}

func (h *Handler) Complete(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userId")
	taskID := chi.URLParam(r, "taskId")

	t, err := h.svc.Complete(r.Context(), userID, taskID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "タスクが見つかりません")
			return
		}
		if errors.Is(err, ErrForbidden) {
			writeError(w, http.StatusForbidden, "FORBIDDEN", "このタスクを操作する権限がありません")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "タスクの完了に失敗しました")
		return
	}

	writeJSON(w, http.StatusOK, apiResponse{Success: true, Data: toResponse(t)})
}

func (h *Handler) Reopen(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userId")
	taskID := chi.URLParam(r, "taskId")

	t, err := h.svc.Reopen(r.Context(), userID, taskID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "タスクが見つかりません")
			return
		}
		if errors.Is(err, ErrForbidden) {
			writeError(w, http.StatusForbidden, "FORBIDDEN", "このタスクを操作する権限がありません")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "タスクの再オープンに失敗しました")
		return
	}

	writeJSON(w, http.StatusOK, apiResponse{Success: true, Data: toResponse(t)})
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userId")
	taskID := chi.URLParam(r, "taskId")

	if err := h.svc.Delete(r.Context(), userID, taskID); err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "タスクが見つかりません")
			return
		}
		if errors.Is(err, ErrForbidden) {
			writeError(w, http.StatusForbidden, "FORBIDDEN", "このタスクを操作する権限がありません")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "タスクの削除に失敗しました")
		return
	}

	writeJSON(w, http.StatusOK, apiResponse{Success: true, Data: nil})
}
