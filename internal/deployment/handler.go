package deployment

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/muah1987/Aihub/internal/auth"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// ─── Env Vars ────────────────────────────────────────────────────────────────

func (h *Handler) ListEnvVars(w http.ResponseWriter, r *http.Request) {
	projectID, ok := parseProjectID(w, r)
	if !ok {
		return
	}
	vars, err := h.service.ListEnvVars(projectID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"env_vars": vars})
}

func (h *Handler) SetEnvVar(w http.ResponseWriter, r *http.Request) {
	projectID, ok := parseProjectID(w, r)
	if !ok {
		return
	}
	var input EnvVarInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid body"})
		return
	}
	if input.Key == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "key is required"})
		return
	}
	result, err := h.service.SetEnvVar(projectID, &input)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"env_var": result})
}

func (h *Handler) DeleteEnvVar(w http.ResponseWriter, r *http.Request) {
	projectID, ok := parseProjectID(w, r)
	if !ok {
		return
	}
	envID, err := uuid.Parse(chi.URLParam(r, "envId"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid env var id"})
		return
	}
	if err := h.service.DeleteEnvVar(projectID, envID); err != nil {
		if errors.Is(err, ErrEnvVarNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "deleted"})
}

// ─── VPS Targets ─────────────────────────────────────────────────────────────

func (h *Handler) ListTargets(w http.ResponseWriter, r *http.Request) {
	projectID, ok := parseProjectID(w, r)
	if !ok {
		return
	}
	targets, err := h.service.ListTargets(projectID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"targets": targets})
}

func (h *Handler) CreateTarget(w http.ResponseWriter, r *http.Request) {
	projectID, ok := parseProjectID(w, r)
	if !ok {
		return
	}
	var input TargetInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid body"})
		return
	}
	if input.Name == "" || input.Host == "" || input.Username == "" || input.DeployCmd == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "name, host, username, deploy_cmd required"})
		return
	}
	target, err := h.service.CreateTarget(projectID, &input)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, map[string]interface{}{"target": target})
}

func (h *Handler) UpdateTarget(w http.ResponseWriter, r *http.Request) {
	projectID, ok := parseProjectID(w, r)
	if !ok {
		return
	}
	targetID, err := uuid.Parse(chi.URLParam(r, "targetId"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid target id"})
		return
	}
	var input TargetInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid body"})
		return
	}
	target, err := h.service.UpdateTarget(projectID, targetID, &input)
	if err != nil {
		if errors.Is(err, ErrTargetNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"target": target})
}

func (h *Handler) DeleteTarget(w http.ResponseWriter, r *http.Request) {
	projectID, ok := parseProjectID(w, r)
	if !ok {
		return
	}
	targetID, err := uuid.Parse(chi.URLParam(r, "targetId"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid target id"})
		return
	}
	if err := h.service.DeleteTarget(projectID, targetID); err != nil {
		if errors.Is(err, ErrTargetNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "deleted"})
}

// ─── Deploy ──────────────────────────────────────────────────────────────────

func (h *Handler) Deploy(w http.ResponseWriter, r *http.Request) {
	userID, ok2 := auth.GetUserIDFromContext(r.Context())
	if !ok2 {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	projectID, ok := parseProjectID(w, r)
	if !ok {
		return
	}
	targetID, err := uuid.Parse(chi.URLParam(r, "targetId"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid target id"})
		return
	}
	run, err := h.service.Deploy(projectID, targetID, &userID)
	if err != nil {
		if errors.Is(err, ErrTargetNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "target not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusAccepted, map[string]interface{}{"run": run})
}

// ─── Deployment Runs ─────────────────────────────────────────────────────────

func (h *Handler) ListRuns(w http.ResponseWriter, r *http.Request) {
	projectID, ok := parseProjectID(w, r)
	if !ok {
		return
	}
	var targetID *uuid.UUID
	if tid := chi.URLParam(r, "targetId"); tid != "" {
		parsed, err := uuid.Parse(tid)
		if err == nil {
			targetID = &parsed
		}
	}
	limit := 20
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	runs, err := h.service.ListRuns(projectID, targetID, limit)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"runs": runs})
}

func (h *Handler) GetRun(w http.ResponseWriter, r *http.Request) {
	runID, err := uuid.Parse(chi.URLParam(r, "runId"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid run id"})
		return
	}
	run, err := h.service.GetRun(runID)
	if err != nil {
		if errors.Is(err, ErrRunNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"run": run})
}

// ─── helpers ─────────────────────────────────────────────────────────────────

func parseProjectID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid project id"})
		return uuid.Nil, false
	}
	return id, true
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
