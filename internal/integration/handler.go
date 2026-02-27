package integration

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/muah1987/Aihub/internal/auth"
	"github.com/muah1987/Aihub/internal/httputil"
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
		httputil.WriteError(w, http.StatusBadRequest, "invalid project id")
		return
	}
	platform := r.URL.Query().Get("platform")
	conns, err := h.service.ListConnections(projectID, platform)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to list connections")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]interface{}{"connections": conns})
}

func (h *Handler) CreateConnection(w http.ResponseWriter, r *http.Request) {
	httputil.LimitBody(w, r, httputil.MaxBodySize)
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid project id")
		return
	}

	var input CreateConnectionInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if input.Platform == "" || input.Name == "" {
		httputil.WriteError(w, http.StatusBadRequest, "platform and name are required")
		return
	}

	conn, err := h.service.CreateConnection(projectID, &userID, &input)
	if err != nil {
		if errors.Is(err, ErrUnsafeURL) {
			httputil.WriteError(w, http.StatusBadRequest, err.Error())
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "failed to create connection")
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, map[string]interface{}{"connection": conn})
}

func (h *Handler) UpdateConnection(w http.ResponseWriter, r *http.Request) {
	httputil.LimitBody(w, r, httputil.MaxBodySize)
	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid project id")
		return
	}
	connID, err := uuid.Parse(chi.URLParam(r, "connId"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid connection id")
		return
	}

	// Whitelist allowed update fields (H5 fix)
	var body struct {
		Name      *string `json:"name"`
		ChannelID *string `json:"channel_id"`
		Enabled   *bool   `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	updates := make(map[string]interface{})
	if body.Name != nil {
		updates["name"] = *body.Name
	}
	if body.ChannelID != nil {
		updates["channel_id"] = *body.ChannelID
	}
	if body.Enabled != nil {
		updates["enabled"] = *body.Enabled
	}

	if len(updates) == 0 {
		httputil.WriteError(w, http.StatusBadRequest, "no valid fields to update")
		return
	}

	if err := h.service.UpdateConnection(projectID, connID, updates); err != nil {
		if errors.Is(err, ErrConnectionNotFound) {
			httputil.WriteError(w, http.StatusNotFound, "connection not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "failed to update")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]string{"message": "connection updated"})
}

func (h *Handler) DeleteConnection(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid project id")
		return
	}
	connID, err := uuid.Parse(chi.URLParam(r, "connId"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid connection id")
		return
	}

	if err := h.service.DeleteConnection(projectID, connID); err != nil {
		if errors.Is(err, ErrConnectionNotFound) {
			httputil.WriteError(w, http.StatusNotFound, "connection not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "failed to delete")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]string{"message": "connection deleted"})
}

func (h *Handler) TestConnection(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid project id")
		return
	}
	connID, err := uuid.Parse(chi.URLParam(r, "connId"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid connection id")
		return
	}

	result, err := h.service.TestConnection(projectID, connID)
	if err != nil {
		httputil.WriteJSON(w, http.StatusOK, map[string]interface{}{"result": "error", "error": err.Error()})
		return
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]interface{}{"result": result})
}

// ---- Outbound Events ----

func (h *Handler) ListOutboundEvents(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid project id")
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	events, err := h.service.ListOutboundEvents(projectID, limit)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to list events")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]interface{}{"events": events})
}

// ---- Notification Rules ----

func (h *Handler) ListRules(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var projectID *uuid.UUID
	if pid := r.URL.Query().Get("project_id"); pid != "" {
		parsed, err := uuid.Parse(pid)
		if err != nil {
			httputil.WriteError(w, http.StatusBadRequest, "invalid project_id")
			return
		}
		projectID = &parsed
	}

	rules, err := h.service.ListRules(userID, projectID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to list rules")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]interface{}{"rules": rules})
}

func (h *Handler) CreateRule(w http.ResponseWriter, r *http.Request) {
	httputil.LimitBody(w, r, httputil.MaxBodySize)
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var body struct {
		ProjectID   string `json:"project_id"`
		EventType   string `json:"event_type"`
		Channel     string `json:"channel"`
		MinSeverity string `json:"min_severity"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if body.EventType == "" || body.Channel == "" {
		httputil.WriteError(w, http.StatusBadRequest, "event_type and channel are required")
		return
	}

	var projectID *uuid.UUID
	if body.ProjectID != "" {
		parsed, err := uuid.Parse(body.ProjectID)
		if err != nil {
			httputil.WriteError(w, http.StatusBadRequest, "invalid project_id")
			return
		}
		projectID = &parsed
	}

	rule, err := h.service.CreateRule(userID, projectID, body.EventType, body.Channel, body.MinSeverity)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to create rule")
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, map[string]interface{}{"rule": rule})
}

func (h *Handler) UpdateRule(w http.ResponseWriter, r *http.Request) {
	httputil.LimitBody(w, r, httputil.MaxBodySize)
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	ruleID, err := uuid.Parse(chi.URLParam(r, "ruleId"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid rule id")
		return
	}

	// Whitelist allowed update fields (H7 fix)
	var body struct {
		EventType       *string `json:"event_type"`
		Channel         *string `json:"channel"`
		Enabled         *bool   `json:"enabled"`
		MinSeverity     *string `json:"min_severity"`
		QuietHoursStart *int    `json:"quiet_hours_start"`
		QuietHoursEnd   *int    `json:"quiet_hours_end"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	updates := make(map[string]interface{})
	if body.EventType != nil {
		updates["event_type"] = *body.EventType
	}
	if body.Channel != nil {
		updates["channel"] = *body.Channel
	}
	if body.Enabled != nil {
		updates["enabled"] = *body.Enabled
	}
	if body.MinSeverity != nil {
		updates["min_severity"] = *body.MinSeverity
	}
	if body.QuietHoursStart != nil {
		updates["quiet_hours_start"] = *body.QuietHoursStart
	}
	if body.QuietHoursEnd != nil {
		updates["quiet_hours_end"] = *body.QuietHoursEnd
	}

	if err := h.service.UpdateRule(userID, ruleID, updates); err != nil {
		if errors.Is(err, ErrRuleNotFound) {
			httputil.WriteError(w, http.StatusNotFound, "rule not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "failed to update rule")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]string{"message": "rule updated"})
}

func (h *Handler) DeleteRule(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	ruleID, err := uuid.Parse(chi.URLParam(r, "ruleId"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid rule id")
		return
	}

	if err := h.service.DeleteRule(userID, ruleID); err != nil {
		if errors.Is(err, ErrRuleNotFound) {
			httputil.WriteError(w, http.StatusNotFound, "rule not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "failed to delete rule")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]string{"message": "rule deleted"})
}

// ---- Email Digests ----

func (h *Handler) GetDigest(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var projectID *uuid.UUID
	if pid := r.URL.Query().Get("project_id"); pid != "" {
		parsed, err := uuid.Parse(pid)
		if err != nil {
			httputil.WriteError(w, http.StatusBadRequest, "invalid project_id")
			return
		}
		projectID = &parsed
	}

	digest, err := h.service.GetDigest(userID, projectID)
	if err != nil {
		if errors.Is(err, ErrDigestNotFound) {
			httputil.WriteError(w, http.StatusNotFound, "digest not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "failed to get digest")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]interface{}{"digest": digest})
}

func (h *Handler) UpsertDigest(w http.ResponseWriter, r *http.Request) {
	httputil.LimitBody(w, r, httputil.MaxBodySize)
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var body struct {
		ProjectID string           `json:"project_id"`
		Input     EmailDigestInput `json:"input"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if body.Input.Frequency == "" {
		body.Input.Frequency = "daily"
	}

	var projectID *uuid.UUID
	if body.ProjectID != "" {
		parsed, err := uuid.Parse(body.ProjectID)
		if err != nil {
			httputil.WriteError(w, http.StatusBadRequest, "invalid project_id")
			return
		}
		projectID = &parsed
	}

	digest, err := h.service.UpsertDigest(userID, projectID, &body.Input)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to save digest")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]interface{}{"digest": digest})
}

func (h *Handler) ListDigests(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	digests, err := h.service.ListDigests(userID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to list digests")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]interface{}{"digests": digests})
}

func (h *Handler) DeleteDigest(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	digestID, err := uuid.Parse(chi.URLParam(r, "digestId"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid digest id")
		return
	}

	if err := h.service.DeleteDigest(userID, digestID); err != nil {
		if errors.Is(err, ErrDigestNotFound) {
			httputil.WriteError(w, http.StatusNotFound, "digest not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "failed to delete digest")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]string{"message": "digest deleted"})
}
