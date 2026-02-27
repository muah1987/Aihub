package knowledge

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

func (h *Handler) ListVersions(w http.ResponseWriter, r *http.Request) {
	memoryID, err := uuid.Parse(chi.URLParam(r, "memoryId"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid memory id")
		return
	}

	versions, err := h.service.ListVersions(memoryID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to list versions")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]interface{}{"versions": versions})
}

func (h *Handler) GetVersion(w http.ResponseWriter, r *http.Request) {
	memoryID, err := uuid.Parse(chi.URLParam(r, "memoryId"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid memory id")
		return
	}
	versionNum, err := strconv.Atoi(chi.URLParam(r, "versionNumber"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid version number")
		return
	}

	version, err := h.service.GetVersion(memoryID, versionNum)
	if err != nil {
		if errors.Is(err, ErrVersionNotFound) {
			httputil.WriteError(w, http.StatusNotFound, "version not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "failed to get version")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]interface{}{"version": version})
}

func (h *Handler) RollbackMemory(w http.ResponseWriter, r *http.Request) {
	httputil.LimitBody(w, r, httputil.MaxBodySize)
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	memoryID, err := uuid.Parse(chi.URLParam(r, "memoryId"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid memory id")
		return
	}

	var body struct {
		VersionNumber int `json:"version_number"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.VersionNumber <= 0 {
		httputil.WriteError(w, http.StatusBadRequest, "version_number is required")
		return
	}

	if err := h.service.RollbackMemory(memoryID, body.VersionNumber, &userID); err != nil {
		if errors.Is(err, ErrVersionNotFound) {
			httputil.WriteError(w, http.StatusNotFound, "version not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "failed to rollback")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]string{"message": "memory rolled back"})
}
