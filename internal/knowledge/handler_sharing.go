package knowledge

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/muah1987/Aihub/internal/auth"
)

func (h *Handler) ShareMemory(w http.ResponseWriter, r *http.Request) {
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

	var body struct {
		OrganizationID string `json:"organization_id"`
		MemoryID       string `json:"memory_id"`
		AccessLevel    string `json:"access_level"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	orgID, err := uuid.Parse(body.OrganizationID)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid organization_id"})
		return
	}
	memoryID, err := uuid.Parse(body.MemoryID)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid memory_id"})
		return
	}

	shared, err := h.service.ShareMemory(orgID, projectID, memoryID, &userID, body.AccessLevel)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to share memory"})
		return
	}
	writeJSON(w, http.StatusCreated, map[string]interface{}{"shared_memory": shared})
}

func (h *Handler) UnshareMemory(w http.ResponseWriter, r *http.Request) {
	orgID, err := uuid.Parse(chi.URLParam(r, "orgId"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid org id"})
		return
	}
	sharedID, err := uuid.Parse(chi.URLParam(r, "sharedId"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid shared memory id"})
		return
	}

	if err := h.service.UnshareMemory(orgID, sharedID); err != nil {
		if errors.Is(err, ErrSharedMemoryNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "shared memory not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to unshare"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "memory unshared"})
}

func (h *Handler) ListSharedMemories(w http.ResponseWriter, r *http.Request) {
	orgID, err := uuid.Parse(chi.URLParam(r, "orgId"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid org id"})
		return
	}

	shared, err := h.service.ListSharedMemories(orgID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list shared memories"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"shared_memories": shared})
}

func (h *Handler) GetSharedForProject(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid project id"})
		return
	}

	shared, err := h.service.GetSharedMemoriesForProject(projectID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get shared memories"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"shared_memories": shared})
}
