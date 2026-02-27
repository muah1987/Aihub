package activity

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

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid project id"})
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	action := r.URL.Query().Get("action")
	resource := r.URL.Query().Get("resource_type")

	if action != "" {
		logs, err := h.service.ListByAction(projectID, action, limit)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list activity"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{"activity": logs})
		return
	}

	if resource != "" {
		logs, err := h.service.ListByResource(projectID, resource, limit)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list activity"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{"activity": logs})
		return
	}

	logs, total, err := h.service.List(projectID, limit, offset)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list activity"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"activity": logs, "total": total})
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
