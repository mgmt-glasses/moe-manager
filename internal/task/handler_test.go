package task_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/mgmt-glasses/moe-manager/internal/task"
	"github.com/mgmt-glasses/moe-manager/internal/task/testutil"
)

type envelope struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data"`
	Error   *struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func newTestRouter(h *task.Handler) http.Handler {
	r := chi.NewRouter()
	r.Route("/api/v1/users/{userId}", func(r chi.Router) {
		r.Post("/tasks", h.Create)
		r.Get("/tasks", h.List)
		r.Patch("/tasks/{taskId}/complete", h.Complete)
		r.Patch("/tasks/{taskId}/reopen", h.Reopen)
		r.Delete("/tasks/{taskId}", h.Delete)
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
	router := newTestRouter(task.NewHandler(task.NewService(testutil.NewFakeRepository())))

	rec, env := doRequest(t, router, http.MethodPost, "/api/v1/users/u_001/tasks", `{"title":"資料作成"}`)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status: got %d, want %d", rec.Code, http.StatusCreated)
	}
	if !env.Success || env.Error != nil {
		t.Fatalf("expected success envelope, got %+v", env)
	}
	var data map[string]any
	json.Unmarshal(env.Data, &data)
	if data["title"] != "資料作成" {
		t.Errorf("title: got %v, want 資料作成", data["title"])
	}
	if data["status"] != "todo" {
		t.Errorf("status: got %v, want todo", data["status"])
	}
	if data["completedAt"] != nil {
		t.Errorf("completedAt should be nil, got %v", data["completedAt"])
	}
}

func TestHandler_Create_ValidationError(t *testing.T) {
	router := newTestRouter(task.NewHandler(task.NewService(testutil.NewFakeRepository())))

	rec, env := doRequest(t, router, http.MethodPost, "/api/v1/users/u_001/tasks", `{"title":""}`)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status: got %d, want %d", rec.Code, http.StatusBadRequest)
	}
	if env.Error == nil || env.Error.Code != "VALIDATION_ERROR" {
		t.Errorf("expected VALIDATION_ERROR code, got %+v", env.Error)
	}
}

func TestHandler_Create_InvalidBody(t *testing.T) {
	router := newTestRouter(task.NewHandler(task.NewService(testutil.NewFakeRepository())))

	rec, env := doRequest(t, router, http.MethodPost, "/api/v1/users/u_001/tasks", `not json`)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status: got %d, want %d", rec.Code, http.StatusBadRequest)
	}
	if env.Error == nil || env.Error.Code != "INVALID_BODY" {
		t.Errorf("expected INVALID_BODY code, got %+v", env.Error)
	}
}

func TestHandler_List_OnlyReturnsOwnTasks(t *testing.T) {
	repo := testutil.NewFakeRepository(
		task.Task{ID: "t_001", UserID: "u_001", Title: "A", Status: task.StatusTodo},
		task.Task{ID: "t_002", UserID: "u_002", Title: "B", Status: task.StatusTodo},
	)
	router := newTestRouter(task.NewHandler(task.NewService(repo)))

	rec, env := doRequest(t, router, http.MethodGet, "/api/v1/users/u_001/tasks", "")

	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d, want %d", rec.Code, http.StatusOK)
	}
	var items []map[string]any
	json.Unmarshal(env.Data, &items)
	if len(items) != 1 || items[0]["id"] != "t_001" {
		t.Errorf("expected only u_001's task, got %+v", items)
	}
}

func TestHandler_List_EmptyResultIsEmptyArray(t *testing.T) {
	router := newTestRouter(task.NewHandler(task.NewService(testutil.NewFakeRepository())))

	rec, env := doRequest(t, router, http.MethodGet, "/api/v1/users/u_001/tasks", "")

	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d, want %d", rec.Code, http.StatusOK)
	}
	var items []map[string]any
	json.Unmarshal(env.Data, &items)
	if len(items) != 0 {
		t.Errorf("expected empty array, got %v", items)
	}
}

func TestHandler_Complete_OK(t *testing.T) {
	repo := testutil.NewFakeRepository(task.Task{ID: "t_001", UserID: "u_001", Title: "A", Status: task.StatusTodo})
	router := newTestRouter(task.NewHandler(task.NewService(repo)))

	rec, env := doRequest(t, router, http.MethodPatch, "/api/v1/users/u_001/tasks/t_001/complete", "")

	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d, want %d", rec.Code, http.StatusOK)
	}
	var data map[string]any
	json.Unmarshal(env.Data, &data)
	if data["status"] != "done" {
		t.Errorf("status: got %v, want done", data["status"])
	}
	if data["completedAt"] == nil {
		t.Error("expected completedAt to be set")
	}
}

func TestHandler_Complete_NotFound(t *testing.T) {
	router := newTestRouter(task.NewHandler(task.NewService(testutil.NewFakeRepository())))

	rec, env := doRequest(t, router, http.MethodPatch, "/api/v1/users/u_001/tasks/t_missing/complete", "")

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status: got %d, want %d", rec.Code, http.StatusNotFound)
	}
	if env.Error == nil || env.Error.Code != "NOT_FOUND" {
		t.Errorf("expected NOT_FOUND code, got %+v", env.Error)
	}
}

func TestHandler_Complete_ForbiddenForOtherUser(t *testing.T) {
	repo := testutil.NewFakeRepository(task.Task{ID: "t_001", UserID: "u_001", Title: "A", Status: task.StatusTodo})
	router := newTestRouter(task.NewHandler(task.NewService(repo)))

	rec, env := doRequest(t, router, http.MethodPatch, "/api/v1/users/u_002/tasks/t_001/complete", "")

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status: got %d, want %d", rec.Code, http.StatusForbidden)
	}
	if env.Error == nil || env.Error.Code != "FORBIDDEN" {
		t.Errorf("expected FORBIDDEN code, got %+v", env.Error)
	}
}

func TestHandler_Reopen_OK(t *testing.T) {
	repo := testutil.NewFakeRepository(task.Task{ID: "t_001", UserID: "u_001", Title: "A", Status: task.StatusDone})
	router := newTestRouter(task.NewHandler(task.NewService(repo)))

	rec, env := doRequest(t, router, http.MethodPatch, "/api/v1/users/u_001/tasks/t_001/reopen", "")

	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d, want %d", rec.Code, http.StatusOK)
	}
	var data map[string]any
	json.Unmarshal(env.Data, &data)
	if data["status"] != "todo" {
		t.Errorf("status: got %v, want todo", data["status"])
	}
	if data["completedAt"] != nil {
		t.Errorf("completedAt should be cleared, got %v", data["completedAt"])
	}
}

func TestHandler_Reopen_ForbiddenForOtherUser(t *testing.T) {
	repo := testutil.NewFakeRepository(task.Task{ID: "t_001", UserID: "u_001", Title: "A", Status: task.StatusDone})
	router := newTestRouter(task.NewHandler(task.NewService(repo)))

	rec, env := doRequest(t, router, http.MethodPatch, "/api/v1/users/u_002/tasks/t_001/reopen", "")

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status: got %d, want %d", rec.Code, http.StatusForbidden)
	}
	if env.Error == nil || env.Error.Code != "FORBIDDEN" {
		t.Errorf("expected FORBIDDEN code, got %+v", env.Error)
	}
}

func TestHandler_Delete_OK(t *testing.T) {
	repo := testutil.NewFakeRepository(task.Task{ID: "t_001", UserID: "u_001", Title: "A", Status: task.StatusTodo})
	router := newTestRouter(task.NewHandler(task.NewService(repo)))

	rec, env := doRequest(t, router, http.MethodDelete, "/api/v1/users/u_001/tasks/t_001", "")

	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d, want %d", rec.Code, http.StatusOK)
	}
	if !env.Success {
		t.Error("expected success=true")
	}
	if string(env.Data) != "null" {
		t.Errorf("expected data=null, got %s", env.Data)
	}
}

func TestHandler_Delete_NotFound(t *testing.T) {
	router := newTestRouter(task.NewHandler(task.NewService(testutil.NewFakeRepository())))

	rec, env := doRequest(t, router, http.MethodDelete, "/api/v1/users/u_001/tasks/t_missing", "")

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status: got %d, want %d", rec.Code, http.StatusNotFound)
	}
	if env.Error == nil || env.Error.Code != "NOT_FOUND" {
		t.Errorf("expected NOT_FOUND code, got %+v", env.Error)
	}
}

func TestHandler_Delete_ForbiddenForOtherUser(t *testing.T) {
	repo := testutil.NewFakeRepository(task.Task{ID: "t_001", UserID: "u_001", Title: "A", Status: task.StatusTodo})
	router := newTestRouter(task.NewHandler(task.NewService(repo)))

	rec, env := doRequest(t, router, http.MethodDelete, "/api/v1/users/u_002/tasks/t_001", "")

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status: got %d, want %d", rec.Code, http.StatusForbidden)
	}
	if env.Error == nil || env.Error.Code != "FORBIDDEN" {
		t.Errorf("expected FORBIDDEN code, got %+v", env.Error)
	}
}
