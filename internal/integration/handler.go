package integration

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

// ---- Integration Connections ----

func (h *Handler) ListConnections(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid project id"})
		return
	}
	platform := r.URL.Query().Get("platform")
	conns, err := h.service.ListConnections(projectID, platform)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list connections"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"connections": conns})
}

func (h *Handler) CreateConnection(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid project id"})
		return
	}

	var input CreateConnectionInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if input.Platform == "" || input.Name == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "platform and name are required"})
		return
	}

	conn, err := h.service.CreateConnection(projectID, &userID, &input)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create connection"})
		return
	}
	writeJSON(w, http.StatusCreated, map[string]interface{}{"connection": conn})
}

func (h *Handler) UpdateConnection(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid project id"})
		return
	}
	connID, err := uuid.Parse(chi.URLParam(r, "connId"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid connection id"})
		return
	}

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	// Disallow updating sensitive fields directly
	delete(updates, "id")
	delete(updates, "project_id")
	delete(updates, "created_by")

	if err := h.service.UpdateConnection(projectID, connID, updates); err != nil {
		if errors.Is(err, ErrConnectionNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "connection not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to update"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "connection updated"})
}

func (h *Handler) DeleteConnection(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid project id"})
		return
	}
	connID, err := uuid.Parse(chi.URLParam(r, "connId"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid connection id"})
		return
	}

	if err := h.service.DeleteConnection(projectID, connID); err != nil {
		if errors.Is(err, ErrConnectionNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "connection not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to delete"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "connection deleted"})
}

func (h *Handler) TestConnection(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid project id"})
		return
	}
	connID, err := uuid.Parse(chi.URLParam(r, "connId"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid connection id"})
		return
	}

	result, err := h.service.TestConnection(projectID, connID)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{"result": "error", "error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"result": result})
}

// ---- Outbound Events ----

func (h *Handler) ListOutboundEvents(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid project id"})
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	events, err := h.service.ListOutboundEvents(projectID, limit)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list events"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"events": events})
}

// ---- Notification Rules ----

func (h *Handler) ListRules(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	var projectID *uuid.UUID
	if pid := r.URL.Query().Get("project_id"); pid != "" {
		parsed, err := uuid.Parse(pid)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid project_id"})
			return
		}
		projectID = &parsed
	}

	rules, err := h.service.ListRules(userID, projectID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list rules"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"rules": rules})
}

func (h *Handler) CreateRule(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	var body struct {
		ProjectID   string `json:"project_id"`
		EventType   string `json:"event_type"`
		Channel     string `json:"channel"`
		MinSeverity string `json:"min_severity"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if body.EventType == "" || body.Channel == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "event_type and channel are required"})
		return
	}

	var projectID *uuid.UUID
	if body.ProjectID != "" {
		parsed, err := uuid.Parse(body.ProjectID)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid project_id"})
			return
		}
		projectID = &parsed
	}

	rule, err := h.service.CreateRule(userID, projectID, body.EventType, body.Channel, body.MinSeverity)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create rule"})
		return
	}
	writeJSON(w, http.StatusCreated, map[string]interface{}{"rule": rule})
}

func (h *Handler) UpdateRule(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	ruleID, err := uuid.Parse(chi.URLParam(r, "ruleId"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid rule id"})
		return
	}

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	delete(updates, "id")
	delete(updates, "user_id")

	if err := h.service.UpdateRule(userID, ruleID, updates); err != nil {
		if errors.Is(err, ErrRuleNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "rule not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to update rule"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "rule updated"})
}

func (h *Handler) DeleteRule(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	ruleID, err := uuid.Parse(chi.URLParam(r, "ruleId"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid rule id"})
		return
	}

	if err := h.service.DeleteRule(userID, ruleID); err != nil {
		if errors.Is(err, ErrRuleNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "rule not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to delete rule"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "rule deleted"})
}

// ---- Email Digests ----

func (h *Handler) GetDigest(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	var projectID *uuid.UUID
	if pid := r.URL.Query().Get("project_id"); pid != "" {
		parsed, _ := uuid.Parse(pid)
		projectID = &parsed
	}

	digest, err := h.service.GetDigest(userID, projectID)
	if err != nil {
		if errors.Is(err, ErrDigestNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "digest not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get digest"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"digest": digest})
}

func (h *Handler) UpsertDigest(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	var body struct {
		ProjectID string           `json:"project_id"`
		Input     EmailDigestInput `json:"input"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if body.Input.Frequency == "" {
		body.Input.Frequency = "daily"
	}

	var projectID *uuid.UUID
	if body.ProjectID != "" {
		parsed, _ := uuid.Parse(body.ProjectID)
		projectID = &parsed
	}

	digest, err := h.service.UpsertDigest(userID, projectID, &body.Input)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to save digest"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"digest": digest})
}

func (h *Handler) ListDigests(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	digests, err := h.service.ListDigests(userID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list digests"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"digests": digests})
}

func (h *Handler) DeleteDigest(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	digestID, err := uuid.Parse(chi.URLParam(r, "digestId"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid digest id"})
		return
	}

	if err := h.service.DeleteDigest(userID, digestID); err != nil {
		if errors.Is(err, ErrDigestNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "digest not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to delete digest"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "digest deleted"})
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
