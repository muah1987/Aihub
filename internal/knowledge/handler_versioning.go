package knowledge

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/muah1987/Aihub/internal/auth"
)

func (h *Handler) ListVersions(w http.ResponseWriter, r *http.Request) {
	memoryID, err := uuid.Parse(chi.URLParam(r, "memoryId"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid memory id"})
		return
	}

	versions, err := h.service.ListVersions(memoryID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list versions"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"versions": versions})
}

func (h *Handler) GetVersion(w http.ResponseWriter, r *http.Request) {
	memoryID, err := uuid.Parse(chi.URLParam(r, "memoryId"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid memory id"})
		return
	}
	versionNum, err := strconv.Atoi(chi.URLParam(r, "versionNumber"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid version number"})
		return
	}

	version, err := h.service.GetVersion(memoryID, versionNum)
	if err != nil {
		if errors.Is(err, ErrVersionNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "version not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get version"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"version": version})
}

func (h *Handler) RollbackMemory(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	memoryID, err := uuid.Parse(chi.URLParam(r, "memoryId"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid memory id"})
		return
	}

	var body struct {
		VersionNumber int `json:"version_number"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.VersionNumber <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "version_number is required"})
		return
	}

	if err := h.service.RollbackMemory(memoryID, body.VersionNumber, &userID); err != nil {
		if errors.Is(err, ErrVersionNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "version not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to rollback"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "memory rolled back"})
}
