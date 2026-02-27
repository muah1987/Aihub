package platform

import (
	"encoding/json"
	"errors"
	"net/http"

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

// ---- Preferences ----

func (h *Handler) GetPreferences(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	prefs, err := h.service.GetPreferences(userID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to get preferences")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]interface{}{"preferences": prefs})
}

func (h *Handler) UpdatePreferences(w http.ResponseWriter, r *http.Request) {
	httputil.LimitBody(w, r, httputil.MaxBodySize)
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var input PreferencesInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	prefs, err := h.service.UpdatePreferences(userID, &input)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to update preferences")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]interface{}{"preferences": prefs})
}

// ---- Pinned Projects ----

func (h *Handler) ListPinned(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	pins, err := h.service.ListPinned(userID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to list pinned")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]interface{}{"pinned": pins})
}

func (h *Handler) PinProject(w http.ResponseWriter, r *http.Request) {
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
	pin, err := h.service.PinProject(userID, projectID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to pin")
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, map[string]interface{}{"pinned": pin})
}

func (h *Handler) UnpinProject(w http.ResponseWriter, r *http.Request) {
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
	if err := h.service.UnpinProject(userID, projectID); err != nil {
		if errors.Is(err, ErrPinnedNotFound) {
			httputil.WriteError(w, http.StatusNotFound, "not pinned")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "failed to unpin")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]string{"message": "unpinned"})
}

func (h *Handler) ReorderPins(w http.ResponseWriter, r *http.Request) {
	httputil.LimitBody(w, r, httputil.MaxBodySize)
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var body struct {
		ProjectIDs []string `json:"project_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	var ids []uuid.UUID
	for _, s := range body.ProjectIDs {
		id, err := uuid.Parse(s)
		if err != nil {
			httputil.WriteError(w, http.StatusBadRequest, "invalid project id: "+s)
			return
		}
		ids = append(ids, id)
	}
	h.service.ReorderPins(userID, ids)
	httputil.WriteJSON(w, http.StatusOK, map[string]string{"message": "reordered"})
}

// ---- Global Search ----

func (h *Handler) GlobalSearch(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	q := r.URL.Query().Get("q")
	if q == "" {
		httputil.WriteError(w, http.StatusBadRequest, "search query is required")
		return
	}

	results, err := h.service.GlobalSearch(userID, q, 20)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "search failed")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]interface{}{"results": results})
}

// ---- Prompt Versions ----

func (h *Handler) ListPromptVersions(w http.ResponseWriter, r *http.Request) {
	agentID, err := uuid.Parse(chi.URLParam(r, "agentId"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid agent id")
		return
	}
	versions, err := h.service.ListPromptVersions(agentID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to list versions")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]interface{}{"versions": versions})
}

func (h *Handler) CreatePromptVersion(w http.ResponseWriter, r *http.Request) {
	httputil.LimitBody(w, r, httputil.MaxBodySize)
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	agentID, err := uuid.Parse(chi.URLParam(r, "agentId"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid agent id")
		return
	}

	var body struct {
		Label        string `json:"label"`
		SystemPrompt string `json:"system_prompt"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Label == "" || body.SystemPrompt == "" {
		httputil.WriteError(w, http.StatusBadRequest, "label and system_prompt are required")
		return
	}

	v, err := h.service.CreatePromptVersion(agentID, &userID, body.Label, body.SystemPrompt)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to create")
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, map[string]interface{}{"version": v})
}

func (h *Handler) SetActiveVersion(w http.ResponseWriter, r *http.Request) {
	httputil.LimitBody(w, r, httputil.MaxBodySize)
	agentID, err := uuid.Parse(chi.URLParam(r, "agentId"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid agent id")
		return
	}
	versionID, err := uuid.Parse(chi.URLParam(r, "versionId"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid version id")
		return
	}

	if err := h.service.SetActivePromptVersion(agentID, versionID); err != nil {
		if errors.Is(err, ErrPromptVersionNotFound) {
			httputil.WriteError(w, http.StatusNotFound, "version not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "failed to activate")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]string{"message": "version activated"})
}

func (h *Handler) DeletePromptVersion(w http.ResponseWriter, r *http.Request) {
	agentID, err := uuid.Parse(chi.URLParam(r, "agentId"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid agent id")
		return
	}
	versionID, err := uuid.Parse(chi.URLParam(r, "versionId"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid version id")
		return
	}

	if err := h.service.DeletePromptVersion(agentID, versionID); err != nil {
		if errors.Is(err, ErrPromptVersionNotFound) {
			httputil.WriteError(w, http.StatusNotFound, "version not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "failed to delete")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]string{"message": "version deleted"})
}
