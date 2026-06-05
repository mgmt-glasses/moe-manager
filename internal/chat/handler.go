package chat

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"
)

// Handler is the HTTP handler for the chat domain.
type Handler struct {
	service *Service
}

// NewHandler creates a Handler.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

type sendMessageRequest struct {
	Message       string `json:"message"`
	CharacterID   string `json:"characterId"`
	GenerateVoice bool   `json:"generateVoice"`
}

type messageDTO struct {
	Role      string `json:"role"`
	Message   string `json:"message"`
	CreatedAt string `json:"createdAt"`
}

type contextDTO struct {
	TodayTaskCompletedCount    int `json:"todayTaskCompletedCount"`
	TodayTaskTotalCount        int `json:"todayTaskTotalCount"`
	TodayEntertainmentMinutes  int `json:"todayEntertainmentMinutes"`
	TargetEntertainmentMinutes int `json:"targetEntertainmentMinutes"`
}

type sendMessageData struct {
	UserMessage      messageDTO `json:"userMessage"`
	AssistantMessage messageDTO `json:"assistantMessage"`
	Context          contextDTO `json:"context"`
}

type apiResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Error   *apiError   `json:"error"`
}

type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// HandleSendMessage handles POST /api/v1/users/{userId}/chat/messages.
func (h *Handler) HandleSendMessage(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("userId")

	var req sendMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "リクエストが不正です")
		return
	}
	if req.Message == "" {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "メッセージを入力してください")
		return
	}

	out, err := h.service.SendMessage(r.Context(), SendMessageInput{
		UserID:  userID,
		Message: req.Message,
	})
	if err != nil {
		if errors.Is(err, ErrNoCharacterSelected) {
			writeError(w, http.StatusUnprocessableEntity, "NO_CHARACTER_SELECTED", "キャラクターが選択されていません")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "処理に失敗しました")
		return
	}

	data := sendMessageData{
		UserMessage: messageDTO{
			Role:      out.UserMessage.Role,
			Message:   out.UserMessage.Content,
			CreatedAt: out.UserMessage.CreatedAt.Format(time.RFC3339),
		},
		AssistantMessage: messageDTO{
			Role:      out.AssistantMessage.Role,
			Message:   out.AssistantMessage.Content,
			CreatedAt: out.AssistantMessage.CreatedAt.Format(time.RFC3339),
		},
		Context: contextDTO{
			TodayTaskCompletedCount:    out.Context.Tasks.CompletedCount,
			TodayTaskTotalCount:        out.Context.Tasks.TotalCount,
			TodayEntertainmentMinutes:  out.Context.ScreenTime.TodayMinutes,
			TargetEntertainmentMinutes: out.Context.ScreenTime.TargetMinutes,
		},
	}

	writeJSON(w, http.StatusOK, apiResponse{Success: true, Data: data, Error: nil})
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v) //nolint:errcheck
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, apiResponse{
		Success: false,
		Data:    nil,
		Error:   &apiError{Code: code, Message: message},
	})
}
