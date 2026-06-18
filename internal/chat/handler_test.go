package chat_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/mgmt-glasses/moe-manager/internal/chat"
	"github.com/mgmt-glasses/moe-manager/internal/chat/testutil"
)

type envelope struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data"`
	Error   *struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func newTestRouter(h *chat.Handler) http.Handler {
	r := chi.NewRouter()
	r.Post("/api/v1/users/{userId}/chat/messages", h.HandleSendMessage)
	return r
}

func decodeEnvelope(t *testing.T, body []byte) envelope {
	t.Helper()
	var env envelope
	if err := json.Unmarshal(body, &env); err != nil {
		t.Fatalf("decode envelope: %v (body=%s)", err, body)
	}
	return env
}

func doRequest(t *testing.T, h http.Handler, path, body string) (*httptest.ResponseRecorder, envelope) {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec, decodeEnvelope(t, rec.Body.Bytes())
}

func TestHandler_HandleSendMessage_OK(t *testing.T) {
	loader := &testutil.FakeContextLoader{Ctx: chat.ChatContext{
		Character:  chat.CharacterProfile{ID: "char_001", Name: "さくら", MBTI: "ISTJ", Personality: "真面目", SpeechStyle: "丁寧"},
		Tasks:      chat.TaskSummary{CompletedCount: 3, TotalCount: 5},
		ScreenTime: chat.ScreenTimeSummary{TodayMinutes: 180, TargetMinutes: 120},
	}}
	llm := &testutil.FakeLLMClient{Reply: "了解しました、社長。"}
	logger := &testutil.FakeChatLogger{}
	svc := chat.NewService(llm, loader, nil, logger)
	router := newTestRouter(chat.NewHandler(svc))

	rec, env := doRequest(t, router, "/api/v1/users/u_001/chat/messages", `{"message":"今日はゲームしすぎた"}`)

	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d, want %d", rec.Code, http.StatusOK)
	}
	if !env.Success || env.Error != nil {
		t.Fatalf("expected success envelope, got %+v", env)
	}

	var data map[string]any
	if err := json.Unmarshal(env.Data, &data); err != nil {
		t.Fatalf("decode data: %v", err)
	}

	userMsg, ok := data["userMessage"].(map[string]any)
	if !ok {
		t.Fatalf("expected userMessage object, got %+v", data["userMessage"])
	}
	if userMsg["role"] != "user" || userMsg["message"] != "今日はゲームしすぎた" {
		t.Errorf("unexpected userMessage: %+v", userMsg)
	}
	if userMsg["createdAt"] == "" || userMsg["createdAt"] == nil {
		t.Error("expected userMessage.createdAt to be set")
	}

	assistantMsg, ok := data["assistantMessage"].(map[string]any)
	if !ok {
		t.Fatalf("expected assistantMessage object, got %+v", data["assistantMessage"])
	}
	if assistantMsg["role"] != "assistant" || assistantMsg["message"] != "了解しました、社長。" {
		t.Errorf("unexpected assistantMessage: %+v", assistantMsg)
	}

	ctxData, ok := data["context"].(map[string]any)
	if !ok {
		t.Fatalf("expected context object, got %+v", data["context"])
	}
	want := map[string]any{
		"todayTaskCompletedCount":    float64(3),
		"todayTaskTotalCount":        float64(5),
		"todayEntertainmentMinutes":  float64(180),
		"targetEntertainmentMinutes": float64(120),
	}
	for k, v := range want {
		if got := ctxData[k]; got != v {
			t.Errorf("context field %q: got %v, want %v", k, got, v)
		}
	}

	if len(logger.Saved) != 2 {
		t.Errorf("expected 2 messages persisted, got %d", len(logger.Saved))
	}
}

func TestHandler_HandleSendMessage_EmptyMessageRejected(t *testing.T) {
	loader := &testutil.FakeContextLoader{}
	llm := &testutil.FakeLLMClient{}
	router := newTestRouter(chat.NewHandler(chat.NewService(llm, loader, nil, nil)))

	rec, env := doRequest(t, router, "/api/v1/users/u_001/chat/messages", `{"message":""}`)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status: got %d, want %d", rec.Code, http.StatusBadRequest)
	}
	if env.Error == nil || env.Error.Code != "VALIDATION_ERROR" {
		t.Errorf("expected VALIDATION_ERROR code, got %+v", env.Error)
	}
}

func TestHandler_HandleSendMessage_InvalidBody(t *testing.T) {
	loader := &testutil.FakeContextLoader{}
	llm := &testutil.FakeLLMClient{}
	router := newTestRouter(chat.NewHandler(chat.NewService(llm, loader, nil, nil)))

	rec, env := doRequest(t, router, "/api/v1/users/u_001/chat/messages", `not json`)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status: got %d, want %d", rec.Code, http.StatusBadRequest)
	}
	if env.Error == nil || env.Error.Code != "INVALID_REQUEST" {
		t.Errorf("expected INVALID_REQUEST code, got %+v", env.Error)
	}
}

func TestHandler_HandleSendMessage_NoCharacterSelected(t *testing.T) {
	loader := &testutil.FakeContextLoader{Ctx: chat.ChatContext{}} // no character
	llm := &testutil.FakeLLMClient{}
	router := newTestRouter(chat.NewHandler(chat.NewService(llm, loader, nil, nil)))

	rec, env := doRequest(t, router, "/api/v1/users/u_001/chat/messages", `{"message":"hello"}`)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status: got %d, want %d", rec.Code, http.StatusUnprocessableEntity)
	}
	if env.Error == nil || env.Error.Code != "NO_CHARACTER_SELECTED" {
		t.Errorf("expected NO_CHARACTER_SELECTED code, got %+v", env.Error)
	}
}

func TestHandler_HandleSendMessage_LLMError(t *testing.T) {
	loader := &testutil.FakeContextLoader{Ctx: chat.ChatContext{Character: chat.CharacterProfile{ID: "char_001"}}}
	llm := &testutil.FakeLLMClient{Err: errors.New("LLM timeout")}
	router := newTestRouter(chat.NewHandler(chat.NewService(llm, loader, nil, nil)))

	rec, env := doRequest(t, router, "/api/v1/users/u_001/chat/messages", `{"message":"hello"}`)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status: got %d, want %d", rec.Code, http.StatusInternalServerError)
	}
	if env.Error == nil || env.Error.Code != "INTERNAL_ERROR" {
		t.Errorf("expected INTERNAL_ERROR code, got %+v", env.Error)
	}
}

func TestHandler_HandleSendMessage_ContextLoadError(t *testing.T) {
	loader := &testutil.FakeContextLoader{Err: errors.New("DB unavailable")}
	llm := &testutil.FakeLLMClient{}
	router := newTestRouter(chat.NewHandler(chat.NewService(llm, loader, nil, nil)))

	rec, env := doRequest(t, router, "/api/v1/users/u_001/chat/messages", `{"message":"hello"}`)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status: got %d, want %d", rec.Code, http.StatusInternalServerError)
	}
	if env.Error == nil || env.Error.Code != "INTERNAL_ERROR" {
		t.Errorf("expected INTERNAL_ERROR code, got %+v", env.Error)
	}
}
