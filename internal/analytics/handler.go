package analytics

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Summary(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid project id"})
		return
	}

	days, _ := strconv.Atoi(r.URL.Query().Get("days"))
	if days <= 0 {
		days = 30
	}
	since := time.Now().AddDate(0, 0, -days)

	summary, err := h.service.ProjectSummaryForPeriod(projectID, since)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get summary"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"summary": summary})
}

func (h *Handler) Daily(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid project id"})
		return
	}

	days, _ := strconv.Atoi(r.URL.Query().Get("days"))
	daily, err := h.service.DailySummary(projectID, days)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get daily summary"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"daily": daily})
}

func (h *Handler) Models(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid project id"})
		return
	}

	days, _ := strconv.Atoi(r.URL.Query().Get("days"))
	if days <= 0 {
		days = 30
	}
	since := time.Now().AddDate(0, 0, -days)

	models, err := h.service.ModelBreakdown(projectID, since)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get model breakdown"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"models": models})
}

func (h *Handler) Recent(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid project id"})
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	records, err := h.service.RecentRecords(projectID, limit)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get records"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"records": records})
}

func (h *Handler) GetBudget(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid project id"})
		return
	}

	budget, err := h.service.GetBudget(projectID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get budget"})
		return
	}

	current, limit, over := h.service.CheckBudget(projectID)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"budget":        budget,
		"current_spend": current,
		"monthly_limit": limit,
		"over_budget":   over,
	})
}

func (h *Handler) SetBudget(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid project id"})
		return
	}

	var body struct {
		MonthlyLimitMicrocents int64 `json:"monthly_limit_microcents"`
		AlertThresholdPct      int   `json:"alert_threshold_pct"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if body.AlertThresholdPct <= 0 {
		body.AlertThresholdPct = 80
	}

	budget, err := h.service.SetBudget(projectID, body.MonthlyLimitMicrocents, body.AlertThresholdPct)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"budget": budget})
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
