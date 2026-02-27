package monitoring

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Handler struct {
	service *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{service: svc}
}

func (h *Handler) CollectMetrics(w http.ResponseWriter, r *http.Request) {
	projectID, ok := parseID(w, r, "id")
	if !ok {
		return
	}
	targetID, ok := parseID(w, r, "targetId")
	if !ok {
		return
	}
	m, err := h.service.CollectMetrics(projectID, targetID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"metric": m})
}

func (h *Handler) LatestMetric(w http.ResponseWriter, r *http.Request) {
	targetID, ok := parseID(w, r, "targetId")
	if !ok {
		return
	}
	m, err := h.service.LatestMetric(targetID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "no metrics yet"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"metric": m})
}

func (h *Handler) MetricsHistory(w http.ResponseWriter, r *http.Request) {
	targetID, ok := parseID(w, r, "targetId")
	if !ok {
		return
	}
	limit := 20
	if l := r.URL.Query().Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 {
			limit = v
		}
	}
	ms, err := h.service.MetricsHistory(targetID, limit)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"metrics": ms})
}

// Pipeline stages

func (h *Handler) ListStages(w http.ResponseWriter, r *http.Request) {
	targetID, ok := parseID(w, r, "targetId")
	if !ok {
		return
	}
	stages, err := h.service.ListStages(targetID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"stages": stages})
}

func (h *Handler) CreateStage(w http.ResponseWriter, r *http.Request) {
	targetID, ok := parseID(w, r, "targetId")
	if !ok {
		return
	}
	var body struct {
		Name      string `json:"name"`
		Command   string `json:"command"`
		Order     int    `json:"stage_order"`
		OnFailure string `json:"on_failure"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Name == "" || body.Command == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "name and command required"})
		return
	}
	st, err := h.service.CreateStage(targetID, body.Name, body.Command, body.OnFailure, body.Order)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, map[string]interface{}{"stage": st})
}

func (h *Handler) DeleteStage(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r, "stageId")
	if !ok {
		return
	}
	if err := h.service.DeleteStage(id); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "deleted"})
}

func parseID(w http.ResponseWriter, r *http.Request, param string) (uuid.UUID, bool) {
	id, err := uuid.Parse(chi.URLParam(r, param))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid " + param})
		return uuid.Nil, false
	}
	return id, true
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
