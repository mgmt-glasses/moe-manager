package screentime

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/mgmt-glasses/moe-manager/internal/shared"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

type updateRequest struct {
	Minutes       int `json:"minutes"`
	TargetMinutes int `json:"target_minutes"`
}

type recordResponse struct {
	RecordID      string `json:"record_id"`
	UserID        string `json:"user_id"`
	Date          string `json:"date"`
	Minutes       int    `json:"minutes"`
	TargetMinutes int    `json:"target_minutes"`
	DiffMinutes   int    `json:"diff_minutes"`
}

func toResponse(r *ScreenTimeRecord) recordResponse {
	return recordResponse{
		RecordID:      r.RecordID,
		UserID:        r.UserID,
		Date:          r.Date.Format("2006-01-02"),
		Minutes:       r.Minutes,
		TargetMinutes: r.TargetMinutes,
		DiffMinutes:   r.DiffMinutes,
	}
}

func (h *Handler) Upsert(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userId")
	dateStr := chi.URLParam(r, "date")

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		shared.WriteError(w, http.StatusBadRequest, "INVALID_DATE", "日付の形式が正しくありません")
		return
	}

	var req updateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		shared.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "リクエストの解析に失敗しました")
		return
	}

	rec, err := h.svc.UpsertRecord(r.Context(), userID, date, req.Minutes, req.TargetMinutes)
	if err != nil {
		shared.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "保存に失敗しました")
		return
	}

	shared.WriteJSON(w, http.StatusOK, shared.Response{
		Success: true,
		Data:    toResponse(rec),
	})
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userId")
	dateStr := chi.URLParam(r, "date")

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		shared.WriteError(w, http.StatusBadRequest, "INVALID_DATE", "日付の形式が正しくありません")
		return
	}

	rec, err := h.svc.GetRecord(r.Context(), userID, date)
	if err != nil {
		shared.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "取得に失敗しました")
		return
	}

	if rec == nil {
		shared.WriteError(w, http.StatusNotFound, "NOT_FOUND", "記録が見つかりません")
		return
	}

	shared.WriteJSON(w, http.StatusOK, shared.Response{
		Success: true,
		Data:    toResponse(rec),
	})
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userId")

	// Default to last 7 days if from/to are not provided
	toDate := time.Now()
	fromDate := toDate.AddDate(0, 0, -6)

	if s := r.URL.Query().Get("from"); s != "" {
		if d, err := time.Parse("2006-01-02", s); err == nil {
			fromDate = d
		}
	}
	if s := r.URL.Query().Get("to"); s != "" {
		if d, err := time.Parse("2006-01-02", s); err == nil {
			toDate = d
		}
	}

	records, err := h.svc.ListRecords(r.Context(), userID, fromDate, toDate)
	if err != nil {
		shared.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "一覧の取得に失敗しました")
		return
	}

	res := make([]recordResponse, len(records))
	for i, rec := range records {
		res[i] = toResponse(&rec)
	}

	shared.WriteJSON(w, http.StatusOK, shared.Response{
		Success: true,
		Data:    res,
	})
}
