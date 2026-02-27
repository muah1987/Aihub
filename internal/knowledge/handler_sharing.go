package knowledge

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/muah1987/Aihub/internal/auth"
	"github.com/muah1987/Aihub/internal/httputil"
)

func (h *Handler) ShareMemory(w http.ResponseWriter, r *http.Request) {
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

	var body struct {
		OrganizationID string `json:"organization_id"`
		MemoryID       string `json:"memory_id"`
		AccessLevel    string `json:"access_level"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	orgID, err := uuid.Parse(body.OrganizationID)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid organization_id")
		return
	}
	memoryID, err := uuid.Parse(body.MemoryID)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid memory_id")
		return
	}

	shared, err := h.service.ShareMemory(orgID, projectID, memoryID, &userID, body.AccessLevel)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to share memory")
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, map[string]interface{}{"shared_memory": shared})
}

func (h *Handler) UnshareMemory(w http.ResponseWriter, r *http.Request) {
	orgID, err := uuid.Parse(chi.URLParam(r, "orgId"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid org id")
		return
	}
	sharedID, err := uuid.Parse(chi.URLParam(r, "sharedId"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid shared memory id")
		return
	}

	if err := h.service.UnshareMemory(orgID, sharedID); err != nil {
		if errors.Is(err, ErrSharedMemoryNotFound) {
			httputil.WriteError(w, http.StatusNotFound, "shared memory not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "failed to unshare")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]string{"message": "memory unshared"})
}

func (h *Handler) ListSharedMemories(w http.ResponseWriter, r *http.Request) {
	orgID, err := uuid.Parse(chi.URLParam(r, "orgId"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid org id")
		return
	}

	shared, err := h.service.ListSharedMemories(orgID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to list shared memories")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]interface{}{"shared_memories": shared})
}

func (h *Handler) GetSharedForProject(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid project id")
		return
	}

	shared, err := h.service.GetSharedMemoriesForProject(projectID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to get shared memories")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]interface{}{"shared_memories": shared})
}
