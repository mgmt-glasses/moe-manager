package character_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/mgmt-glasses/moe-manager/internal/character"
	"github.com/mgmt-glasses/moe-manager/internal/character/testutil"
)

type envelope struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data"`
	Error   *struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func newTestRouter(h *character.Handler) http.Handler {
	r := chi.NewRouter()
	r.Get("/api/v1/characters", h.List)
	r.Get("/api/v1/characters/{characterId}", h.GetByID)
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

func TestHandler_List_OK(t *testing.T) {
	repo := &testutil.FakeRepository{Characters: []character.Character{
		{
			ID:                "char_001",
			Name:              "さくら",
			MBTIType:          "ISTJ",
			PersonalityDesc:   "真面目で責任感が強い",
			SpeechStyle:       "丁寧で落ち着いた口調",
			VoicePresetID:     "voice_001",
			IconPath:          "characters/char_001/icon.png",
			StandingImagePath: "",
		},
	}}
	router := newTestRouter(character.NewHandler(character.NewService(repo)))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/characters", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d, want %d", rec.Code, http.StatusOK)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("content-type: got %q, want application/json", ct)
	}

	env := decodeEnvelope(t, rec.Body.Bytes())
	if !env.Success || env.Error != nil {
		t.Fatalf("expected success envelope, got %+v", env)
	}

	var items []map[string]any
	if err := json.Unmarshal(env.Data, &items); err != nil {
		t.Fatalf("decode data: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}

	want := map[string]any{
		"characterId":      "char_001",
		"name":             "さくら",
		"mbti":             "ISTJ",
		"description":      "真面目で責任感が強い",
		"tone":             "丁寧で落ち着いた口調",
		"voiceId":          "voice_001",
		"iconUrl":          "/assets/characters/char_001/icon.png",
		"standingImageUrl": "",
	}
	for k, v := range want {
		if got := items[0][k]; got != v {
			t.Errorf("field %q: got %v, want %v", k, got, v)
		}
	}
	if _, ok := items[0]["sampleVoiceUrl"]; ok {
		t.Errorf("list item must not expose sampleVoiceUrl (detail-only field)")
	}
}

func TestHandler_List_EmptyResultIsEmptyArray(t *testing.T) {
	router := newTestRouter(character.NewHandler(character.NewService(&testutil.FakeRepository{})))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/characters", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	env := decodeEnvelope(t, rec.Body.Bytes())
	var items []map[string]any
	if err := json.Unmarshal(env.Data, &items); err != nil {
		t.Fatalf("decode data: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("expected empty array, got %v", items)
	}
}

func TestHandler_List_ServiceError(t *testing.T) {
	repo := &testutil.FakeRepository{ListErr: errors.New("db down")}
	router := newTestRouter(character.NewHandler(character.NewService(repo)))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/characters", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status: got %d, want %d", rec.Code, http.StatusInternalServerError)
	}
	env := decodeEnvelope(t, rec.Body.Bytes())
	if env.Success {
		t.Error("expected success=false")
	}
	if env.Error == nil || env.Error.Code != "INTERNAL_ERROR" {
		t.Errorf("expected INTERNAL_ERROR code, got %+v", env.Error)
	}
}

func TestHandler_GetByID_OK(t *testing.T) {
	repo := &testutil.FakeRepository{Characters: []character.Character{
		{
			ID:                "char_001",
			Name:              "さくら",
			MBTIType:          "ISTJ",
			PersonalityDesc:   "真面目で責任感が強い",
			SpeechStyle:       "丁寧で落ち着いた口調",
			VoicePresetID:     "voice_001",
			IconPath:          "characters/char_001/icon.png",
			StandingImagePath: "characters/char_001/standing.png",
			SampleVoicePath:   "characters/char_001/sample.mp3",
		},
	}}
	router := newTestRouter(character.NewHandler(character.NewService(repo)))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/characters/char_001", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d, want %d", rec.Code, http.StatusOK)
	}
	env := decodeEnvelope(t, rec.Body.Bytes())
	var detail map[string]any
	if err := json.Unmarshal(env.Data, &detail); err != nil {
		t.Fatalf("decode data: %v", err)
	}

	want := map[string]any{
		"characterId":      "char_001",
		"mbti":             "ISTJ",
		"voiceId":          "voice_001",
		"iconUrl":          "/assets/characters/char_001/icon.png",
		"standingImageUrl": "/assets/characters/char_001/standing.png",
		"sampleVoiceUrl":   "/assets/characters/char_001/sample.mp3",
	}
	for k, v := range want {
		if got := detail[k]; got != v {
			t.Errorf("field %q: got %v, want %v", k, got, v)
		}
	}
}

func TestHandler_GetByID_NotFound(t *testing.T) {
	router := newTestRouter(character.NewHandler(character.NewService(&testutil.FakeRepository{})))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/characters/char_missing", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status: got %d, want %d", rec.Code, http.StatusNotFound)
	}
	env := decodeEnvelope(t, rec.Body.Bytes())
	if env.Success {
		t.Error("expected success=false")
	}
	if env.Error == nil || env.Error.Code != "NOT_FOUND" {
		t.Errorf("expected NOT_FOUND code, got %+v", env.Error)
	}
}

func TestHandler_GetByID_ServiceError(t *testing.T) {
	repo := &testutil.FakeRepository{FindErr: errors.New("db down")}
	router := newTestRouter(character.NewHandler(character.NewService(repo)))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/characters/char_001", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status: got %d, want %d", rec.Code, http.StatusInternalServerError)
	}
	env := decodeEnvelope(t, rec.Body.Bytes())
	if env.Error == nil || env.Error.Code != "INTERNAL_ERROR" {
		t.Errorf("expected INTERNAL_ERROR code, got %+v", env.Error)
	}
}
