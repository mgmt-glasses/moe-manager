package statistics

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/mgmt-glasses/moe-manager/internal/shared"
)

type Handler struct {
	svc *StatisticsService
}

func NewHandler(svc *StatisticsService) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) GetToday(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userId")
	stats, err := h.svc.GetToday(r.Context(), userID)
	if err != nil {
		shared.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "統計の取得に失敗しました")
		return
	}

	type tasksData struct {
		CompletedCount int `json:"completedCount"`
		TodoCount      int `json:"todoCount"`
		TotalCount     int `json:"totalCount"`
		CompletionRate int `json:"completionRate"`
	}
	type entertainmentData struct {
		Minutes       int `json:"minutes"`
		TargetMinutes int `json:"targetMinutes"`
		DiffMinutes   int `json:"diffMinutes"`
	}
	type todayData struct {
		Date          string            `json:"date"`
		Tasks         tasksData         `json:"tasks"`
		Entertainment entertainmentData `json:"entertainment"`
		SummaryText   string            `json:"summaryText"`
	}

	shared.WriteJSON(w, http.StatusOK, shared.Response{
		Success: true,
		Data: todayData{
			Date: stats.Date.Format("2006-01-02"),
			Tasks: tasksData{
				CompletedCount: stats.CompletedTaskCount,
				TodoCount:      stats.TodoTaskCount,
				TotalCount:     stats.TotalTaskCount,
				CompletionRate: stats.TaskCompletionRate,
			},
			Entertainment: entertainmentData{
				Minutes:       stats.EntertainmentMinutes,
				TargetMinutes: stats.TargetEntertainmentMinutes,
				DiffMinutes:   stats.EntertainmentDiffMinutes,
			},
			SummaryText: buildSummaryText(stats),
		},
	})
}

func (h *Handler) GetDaily(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userId")
	date, err := time.Parse("2006-01-02", chi.URLParam(r, "date"))
	if err != nil {
		shared.WriteError(w, http.StatusBadRequest, "INVALID_DATE", "日付の形式が正しくありません（YYYY-MM-DD）")
		return
	}

	stats, err := h.svc.GetDaily(r.Context(), userID, date)
	if err != nil {
		shared.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "統計の取得に失敗しました")
		return
	}

	type dailyData struct {
		Date                       string `json:"date"`
		CompletedTaskCount         int    `json:"completedTaskCount"`
		TodoTaskCount              int    `json:"todoTaskCount"`
		TotalTaskCount             int    `json:"totalTaskCount"`
		TaskCompletionRate         int    `json:"taskCompletionRate"`
		EntertainmentMinutes       int    `json:"entertainmentMinutes"`
		TargetEntertainmentMinutes int    `json:"targetEntertainmentMinutes"`
		EntertainmentDiffMinutes   int    `json:"entertainmentDiffMinutes"`
		SummaryText                string `json:"summaryText"`
	}

	shared.WriteJSON(w, http.StatusOK, shared.Response{
		Success: true,
		Data: dailyData{
			Date:                       stats.Date.Format("2006-01-02"),
			CompletedTaskCount:         stats.CompletedTaskCount,
			TodoTaskCount:              stats.TodoTaskCount,
			TotalTaskCount:             stats.TotalTaskCount,
			TaskCompletionRate:         stats.TaskCompletionRate,
			EntertainmentMinutes:       stats.EntertainmentMinutes,
			TargetEntertainmentMinutes: stats.TargetEntertainmentMinutes,
			EntertainmentDiffMinutes:   stats.EntertainmentDiffMinutes,
			SummaryText:                buildSummaryText(stats),
		},
	})
}

func (h *Handler) GetWeekly(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userId")

	endDate := time.Now()
	if s := r.URL.Query().Get("endDate"); s != "" {
		var err error
		endDate, err = time.Parse("2006-01-02", s)
		if err != nil {
			shared.WriteError(w, http.StatusBadRequest, "INVALID_DATE", "endDate の形式が正しくありません（YYYY-MM-DD）")
			return
		}
	}

	weekly, err := h.svc.GetWeekly(r.Context(), userID, endDate)
	if err != nil {
		shared.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "統計の取得に失敗しました")
		return
	}

	type dayData struct {
		Date                       string `json:"date"`
		CompletedTaskCount         int    `json:"completedTaskCount"`
		TodoTaskCount              int    `json:"todoTaskCount"`
		TaskCompletionRate         int    `json:"taskCompletionRate"`
		EntertainmentMinutes       int    `json:"entertainmentMinutes"`
		TargetEntertainmentMinutes int    `json:"targetEntertainmentMinutes"`
		EntertainmentDiffMinutes   int    `json:"entertainmentDiffMinutes"`
	}
	type weeklyData struct {
		From string    `json:"from"`
		To   string    `json:"to"`
		Days []dayData `json:"days"`
	}

	days := make([]dayData, len(weekly.Days))
	for i, d := range weekly.Days {
		days[i] = dayData{
			Date:                       d.Date.Format("2006-01-02"),
			CompletedTaskCount:         d.CompletedTaskCount,
			TodoTaskCount:              d.TodoTaskCount,
			TaskCompletionRate:         d.TaskCompletionRate,
			EntertainmentMinutes:       d.EntertainmentMinutes,
			TargetEntertainmentMinutes: d.TargetEntertainmentMinutes,
			EntertainmentDiffMinutes:   d.EntertainmentDiffMinutes,
		}
	}

	shared.WriteJSON(w, http.StatusOK, shared.Response{
		Success: true,
		Data: weeklyData{
			From: weekly.From.Format("2006-01-02"),
			To:   weekly.To.Format("2006-01-02"),
			Days: days,
		},
	})
}

// buildSummaryText はテンプレートベースのサマリーを生成する。
// TODO: チャット Issue で LLM 生成に差し替える。
func buildSummaryText(stats DailyStats) string {
	if stats.TotalTaskCount == 0 && stats.EntertainmentMinutes == 0 {
		return "本日はまだタスクも娯楽時間も記録されていません。"
	}
	summary := ""
	if stats.TotalTaskCount > 0 {
		summary += fmt.Sprintf("今日は%d件中%d件のタスクを完了しています。", stats.TotalTaskCount, stats.CompletedTaskCount)
	}
	if stats.EntertainmentMinutes > 0 {
		switch {
		case stats.EntertainmentDiffMinutes > 0:
			summary += fmt.Sprintf("娯楽時間は目標を%d分超過しています。", stats.EntertainmentDiffMinutes)
		case stats.EntertainmentDiffMinutes < 0:
			summary += fmt.Sprintf("娯楽時間は目標より%d分少なく、良好です。", -stats.EntertainmentDiffMinutes)
		default:
			summary += "娯楽時間は目標通りです。"
		}
	}
	return summary
}
