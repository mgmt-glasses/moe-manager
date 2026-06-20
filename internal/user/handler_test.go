package user_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/mgmt-glasses/moe-manager/internal/user"
	"github.com/mgmt-glasses/moe-manager/internal/user/testutil"
)

type envelope struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data"`
	Error   *struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func newTestRouter(h *user.Handler) http.Handler {
	r := chi.NewRouter()
	r.Post("/api/v1/users", h.Create)
	r.Route("/api/v1/users/{userId}", func(r chi.Router) {
		r.Get("/", h.GetByID)
		r.Patch("/", h.Update)
		r.Patch("/selected-character", h.UpdateSelectedCharacter)
	})
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

func doRequest(t *testing.T, h http.Handler, method, path, body string) (*httptest.ResponseRecorder, envelope) {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec, decodeEnvelope(t, rec.Body.Bytes())
}

func TestHandler_Create_OK(t *testing.T) {
	router := newTestRouter(user.NewHandler(user.NewService(testutil.NewFakeRepository(), testutil.NewFakeCharacterValidator("char_001"))))

	rec, env := doRequest(t, router, http.MethodPost, "/api/v1/users", `{"name":"山田太郎","presidentName":"山田社長","targetEntertainmentMinutes":90,"selectedCharacterId":"char_001"}`)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status: got %d, want %d", rec.Code, http.StatusCreated)
	}
	if !env.Success || env.Error != nil {
		t.Fatalf("expected success envelope, got %+v", env)
	}
	var data map[string]any
	if err := json.Unmarshal(env.Data, &data); err != nil {
		t.Fatalf("decode data: %v", err)
	}
	if data["name"] != "山田太郎" || data["presidentName"] != "山田社長" {
		t.Errorf("unexpected data: %+v", data)
	}
	if data["selectedCharacterId"] != "char_001" {
		t.Errorf("selectedCharacterId: got %v, want char_001", data["selectedCharacterId"])
	}
	if data["userId"] == "" || data["userId"] == nil {
		t.Error("expected userId to be set")
	}
	if data["createdAt"] == "" || data["updatedAt"] == "" {
		t.Error("expected createdAt/updatedAt to be set")
	}
}

func TestHandler_Create_InvalidBody(t *testing.T) {
	router := newTestRouter(user.NewHandler(user.NewService(testutil.NewFakeRepository(), testutil.NewFakeCharacterValidator())))

	rec, env := doRequest(t, router, http.MethodPost, "/api/v1/users", `not json`)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status: got %d, want %d", rec.Code, http.StatusBadRequest)
	}
	if env.Error == nil || env.Error.Code != "INVALID_BODY" {
		t.Errorf("expected INVALID_BODY code, got %+v", env.Error)
	}
}

func TestHandler_Create_ValidationError(t *testing.T) {
	router := newTestRouter(user.NewHandler(user.NewService(testutil.NewFakeRepository(), testutil.NewFakeCharacterValidator())))

	rec, env := doRequest(t, router, http.MethodPost, "/api/v1/users", `{"presidentName":"山田社長"}`)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status: got %d, want %d", rec.Code, http.StatusBadRequest)
	}
	if env.Error == nil || env.Error.Code != "VALIDATION_ERROR" {
		t.Errorf("expected VALIDATION_ERROR code, got %+v", env.Error)
	}
}

func TestHandler_Create_CharacterNotFound(t *testing.T) {
	router := newTestRouter(user.NewHandler(user.NewService(testutil.NewFakeRepository(), testutil.NewFakeCharacterValidator())))

	rec, env := doRequest(t, router, http.MethodPost, "/api/v1/users", `{"name":"山田太郎","presidentName":"山田社長","selectedCharacterId":"char_missing"}`)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status: got %d, want %d", rec.Code, http.StatusBadRequest)
	}
	if env.Error == nil || env.Error.Code != "CHARACTER_NOT_FOUND" {
		t.Errorf("expected CHARACTER_NOT_FOUND code, got %+v", env.Error)
	}
}

func TestHandler_GetByID_OK(t *testing.T) {
	existing := user.User{ID: "u_001", Name: "山田太郎", PresidentName: "山田社長", TargetEntertainmentMinutes: 120}
	router := newTestRouter(user.NewHandler(user.NewService(testutil.NewFakeRepository(existing), testutil.NewFakeCharacterValidator())))

	rec, env := doRequest(t, router, http.MethodGet, "/api/v1/users/u_001", "")

	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d, want %d", rec.Code, http.StatusOK)
	}
	var data map[string]any
	json.Unmarshal(env.Data, &data)
	if data["userId"] != "u_001" || data["name"] != "山田太郎" {
		t.Errorf("unexpected data: %+v", data)
	}
}

func TestHandler_GetByID_NotFound(t *testing.T) {
	router := newTestRouter(user.NewHandler(user.NewService(testutil.NewFakeRepository(), testutil.NewFakeCharacterValidator())))

	rec, env := doRequest(t, router, http.MethodGet, "/api/v1/users/u_missing", "")

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status: got %d, want %d", rec.Code, http.StatusNotFound)
	}
	if env.Error == nil || env.Error.Code != "NOT_FOUND" {
		t.Errorf("expected NOT_FOUND code, got %+v", env.Error)
	}
}

func TestHandler_Update_OK(t *testing.T) {
	existing := user.User{ID: "u_001", Name: "旧名前", PresidentName: "旧社長", TargetEntertainmentMinutes: 60}
	router := newTestRouter(user.NewHandler(user.NewService(testutil.NewFakeRepository(existing), testutil.NewFakeCharacterValidator())))

	rec, env := doRequest(t, router, http.MethodPatch, "/api/v1/users/u_001", `{"name":"新名前"}`)

	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d, want %d", rec.Code, http.StatusOK)
	}
	var data map[string]any
	json.Unmarshal(env.Data, &data)
	if data["name"] != "新名前" {
		t.Errorf("name: got %v, want 新名前", data["name"])
	}
	if data["presidentName"] != "旧社長" {
		t.Errorf("presidentName should be unchanged: got %v", data["presidentName"])
	}
}

func TestHandler_Update_NotFound(t *testing.T) {
	router := newTestRouter(user.NewHandler(user.NewService(testutil.NewFakeRepository(), testutil.NewFakeCharacterValidator())))

	rec, env := doRequest(t, router, http.MethodPatch, "/api/v1/users/u_missing", `{"name":"新名前"}`)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status: got %d, want %d", rec.Code, http.StatusNotFound)
	}
	if env.Error == nil || env.Error.Code != "NOT_FOUND" {
		t.Errorf("expected NOT_FOUND code, got %+v", env.Error)
	}
}

func TestHandler_Update_InvalidBody(t *testing.T) {
	existing := user.User{ID: "u_001", Name: "旧名前", PresidentName: "旧社長"}
	router := newTestRouter(user.NewHandler(user.NewService(testutil.NewFakeRepository(existing), testutil.NewFakeCharacterValidator())))

	rec, env := doRequest(t, router, http.MethodPatch, "/api/v1/users/u_001", `not json`)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status: got %d, want %d", rec.Code, http.StatusBadRequest)
	}
	if env.Error == nil || env.Error.Code != "INVALID_BODY" {
		t.Errorf("expected INVALID_BODY code, got %+v", env.Error)
	}
}

func TestHandler_UpdateSelectedCharacter_OK(t *testing.T) {
	existing := user.User{ID: "u_001", Name: "山田太郎"}
	router := newTestRouter(user.NewHandler(user.NewService(testutil.NewFakeRepository(existing), testutil.NewFakeCharacterValidator("char_001"))))

	rec, env := doRequest(t, router, http.MethodPatch, "/api/v1/users/u_001/selected-character", `{"characterId":"char_001"}`)

	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d, want %d", rec.Code, http.StatusOK)
	}
	var data map[string]any
	json.Unmarshal(env.Data, &data)
	if data["userId"] != "u_001" || data["selectedCharacterId"] != "char_001" {
		t.Errorf("unexpected data: %+v", data)
	}
}

func TestHandler_UpdateSelectedCharacter_MissingCharacterID(t *testing.T) {
	existing := user.User{ID: "u_001", Name: "山田太郎"}
	router := newTestRouter(user.NewHandler(user.NewService(testutil.NewFakeRepository(existing), testutil.NewFakeCharacterValidator())))

	rec, env := doRequest(t, router, http.MethodPatch, "/api/v1/users/u_001/selected-character", `{"characterId":""}`)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status: got %d, want %d", rec.Code, http.StatusBadRequest)
	}
	if env.Error == nil || env.Error.Code != "VALIDATION_ERROR" {
		t.Errorf("expected VALIDATION_ERROR code, got %+v", env.Error)
	}
}

func TestHandler_UpdateSelectedCharacter_UserNotFound(t *testing.T) {
	router := newTestRouter(user.NewHandler(user.NewService(testutil.NewFakeRepository(), testutil.NewFakeCharacterValidator("char_001"))))

	rec, env := doRequest(t, router, http.MethodPatch, "/api/v1/users/u_missing/selected-character", `{"characterId":"char_001"}`)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status: got %d, want %d", rec.Code, http.StatusNotFound)
	}
	if env.Error == nil || env.Error.Code != "NOT_FOUND" {
		t.Errorf("expected NOT_FOUND code, got %+v", env.Error)
	}
}

func TestHandler_UpdateSelectedCharacter_CharacterNotFound(t *testing.T) {
	existing := user.User{ID: "u_001", Name: "山田太郎"}
	router := newTestRouter(user.NewHandler(user.NewService(testutil.NewFakeRepository(existing), testutil.NewFakeCharacterValidator())))

	rec, env := doRequest(t, router, http.MethodPatch, "/api/v1/users/u_001/selected-character", `{"characterId":"char_missing"}`)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status: got %d, want %d", rec.Code, http.StatusBadRequest)
	}
	if env.Error == nil || env.Error.Code != "CHARACTER_NOT_FOUND" {
		t.Errorf("expected CHARACTER_NOT_FOUND code, got %+v", env.Error)
	}
}
