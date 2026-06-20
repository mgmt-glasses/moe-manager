package statistics_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/mgmt-glasses/moe-manager/internal/statistics"
	"github.com/mgmt-glasses/moe-manager/internal/statistics/testutil"
)

type envelope struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data"`
	Error   *struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func newTestRouter(h *statistics.Handler) http.Handler {
	r := chi.NewRouter()
	r.Route("/api/v1/users/{userId}", func(r chi.Router) {
		r.Get("/stats/today", h.GetToday)
		r.Get("/stats/daily/{date}", h.GetDaily)
		r.Get("/stats/weekly", h.GetWeekly)
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

func doRequest(t *testing.T, h http.Handler, path string) (*httptest.ResponseRecorder, envelope) {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec, decodeEnvelope(t, rec.Body.Bytes())
}

func TestHandler_GetToday_OK(t *testing.T) {
	router := newTestRouter(statistics.NewHandler(statistics.NewStatisticsService(&testutil.FakeQuery{})))

	rec, env := doRequest(t, router, "/api/v1/users/u_001/stats/today")

	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d, want %d", rec.Code, http.StatusOK)
	}
	var data map[string]any
	json.Unmarshal(env.Data, &data)

	tasks, ok := data["tasks"].(map[string]any)
	if !ok {
		t.Fatalf("expected nested tasks object, got %+v", data["tasks"])
	}
	for _, k := range []string{"completedCount", "todoCount", "totalCount", "completionRate"} {
		if _, ok := tasks[k]; !ok {
			t.Errorf("tasks missing key %q", k)
		}
	}
	entertainment, ok := data["entertainment"].(map[string]any)
	if !ok {
		t.Fatalf("expected nested entertainment object, got %+v", data["entertainment"])
	}
	for _, k := range []string{"minutes", "targetMinutes", "diffMinutes"} {
		if _, ok := entertainment[k]; !ok {
			t.Errorf("entertainment missing key %q", k)
		}
	}
	if data["summaryText"] != "本日はまだタスクも娯楽時間も記録されていません。" {
		t.Errorf("summaryText: got %v", data["summaryText"])
	}
}

func TestHandler_GetToday_ServiceError(t *testing.T) {
	router := newTestRouter(statistics.NewHandler(statistics.NewStatisticsService(&testutil.FakeQuery{Err: errors.New("db down")})))

	rec, env := doRequest(t, router, "/api/v1/users/u_001/stats/today")

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status: got %d, want %d", rec.Code, http.StatusInternalServerError)
	}
	if env.Error == nil || env.Error.Code != "INTERNAL_ERROR" {
		t.Errorf("expected INTERNAL_ERROR code, got %+v", env.Error)
	}
}

func TestHandler_GetDaily_OK(t *testing.T) {
	query := &testutil.FakeQuery{StatsByDate: map[string]statistics.DailyStats{
		"2026-06-01": {
			Date:                       time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
			CompletedTaskCount:         3,
			TodoTaskCount:              2,
			TotalTaskCount:             5,
			TaskCompletionRate:         60,
			EntertainmentMinutes:       150,
			TargetEntertainmentMinutes: 120,
			EntertainmentDiffMinutes:   30,
		},
	}}
	router := newTestRouter(statistics.NewHandler(statistics.NewStatisticsService(query)))

	rec, env := doRequest(t, router, "/api/v1/users/u_001/stats/daily/2026-06-01")

	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d, want %d", rec.Code, http.StatusOK)
	}
	var data map[string]any
	json.Unmarshal(env.Data, &data)

	want := map[string]any{
		"date":                       "2026-06-01",
		"completedTaskCount":         float64(3),
		"todoTaskCount":              float64(2),
		"totalTaskCount":             float64(5),
		"taskCompletionRate":         float64(60),
		"entertainmentMinutes":       float64(150),
		"targetEntertainmentMinutes": float64(120),
		"entertainmentDiffMinutes":   float64(30),
		"summaryText":                "今日は5件中3件のタスクを完了しています。娯楽時間は目標を30分超過しています。",
	}
	for k, v := range want {
		if got := data[k]; got != v {
			t.Errorf("field %q: got %v, want %v", k, got, v)
		}
	}
}

func TestHandler_GetDaily_InvalidDateFormat(t *testing.T) {
	router := newTestRouter(statistics.NewHandler(statistics.NewStatisticsService(&testutil.FakeQuery{})))

	rec, env := doRequest(t, router, "/api/v1/users/u_001/stats/daily/2026-13-99")

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status: got %d, want %d", rec.Code, http.StatusBadRequest)
	}
	if env.Error == nil || env.Error.Code != "INVALID_DATE" {
		t.Errorf("expected INVALID_DATE code, got %+v", env.Error)
	}
}

func TestHandler_GetDaily_ServiceError(t *testing.T) {
	router := newTestRouter(statistics.NewHandler(statistics.NewStatisticsService(&testutil.FakeQuery{Err: errors.New("db down")})))

	rec, env := doRequest(t, router, "/api/v1/users/u_001/stats/daily/2026-06-01")

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status: got %d, want %d", rec.Code, http.StatusInternalServerError)
	}
	if env.Error == nil || env.Error.Code != "INTERNAL_ERROR" {
		t.Errorf("expected INTERNAL_ERROR code, got %+v", env.Error)
	}
}

func TestHandler_GetWeekly_DefaultsEndDateToToday(t *testing.T) {
	router := newTestRouter(statistics.NewHandler(statistics.NewStatisticsService(&testutil.FakeQuery{})))

	rec, env := doRequest(t, router, "/api/v1/users/u_001/stats/weekly")

	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d, want %d", rec.Code, http.StatusOK)
	}
	var data map[string]any
	json.Unmarshal(env.Data, &data)

	wantTo := time.Now().Format("2006-01-02")
	if data["to"] != wantTo {
		t.Errorf("to: got %v, want %v", data["to"], wantTo)
	}
	days, ok := data["days"].([]any)
	if !ok || len(days) != 7 {
		t.Fatalf("expected 7 days, got %+v", data["days"])
	}
}

func TestHandler_GetWeekly_WithEndDateParam(t *testing.T) {
	router := newTestRouter(statistics.NewHandler(statistics.NewStatisticsService(&testutil.FakeQuery{})))

	rec, env := doRequest(t, router, "/api/v1/users/u_001/stats/weekly?endDate=2026-06-07")

	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d, want %d", rec.Code, http.StatusOK)
	}
	var data map[string]any
	json.Unmarshal(env.Data, &data)

	if data["from"] != "2026-06-01" {
		t.Errorf("from: got %v, want 2026-06-01", data["from"])
	}
	if data["to"] != "2026-06-07" {
		t.Errorf("to: got %v, want 2026-06-07", data["to"])
	}
}

func TestHandler_GetWeekly_InvalidEndDate(t *testing.T) {
	router := newTestRouter(statistics.NewHandler(statistics.NewStatisticsService(&testutil.FakeQuery{})))

	rec, env := doRequest(t, router, "/api/v1/users/u_001/stats/weekly?endDate=not-a-date")

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status: got %d, want %d", rec.Code, http.StatusBadRequest)
	}
	if env.Error == nil || env.Error.Code != "INVALID_DATE" {
		t.Errorf("expected INVALID_DATE code, got %+v", env.Error)
	}
}

func TestHandler_GetWeekly_ServiceError(t *testing.T) {
	router := newTestRouter(statistics.NewHandler(statistics.NewStatisticsService(&testutil.FakeQuery{Err: errors.New("db down")})))

	rec, env := doRequest(t, router, "/api/v1/users/u_001/stats/weekly")

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status: got %d, want %d", rec.Code, http.StatusInternalServerError)
	}
	if env.Error == nil || env.Error.Code != "INTERNAL_ERROR" {
		t.Errorf("expected INTERNAL_ERROR code, got %+v", env.Error)
	}
}
