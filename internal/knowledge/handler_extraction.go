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

func (h *Handler) ListExtractions(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid project id"})
		return
	}

	status := r.URL.Query().Get("status")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	extractions, err := h.service.ListExtractions(projectID, status, limit)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list extractions"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"extractions": extractions})
}

func (h *Handler) AcceptExtraction(w http.ResponseWriter, r *http.Request) {
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
	extractionID, err := uuid.Parse(chi.URLParam(r, "extractionId"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid extraction id"})
		return
	}

	var body struct {
		SaveToMemory bool `json:"save_to_memory"`
	}
	json.NewDecoder(r.Body).Decode(&body)

	if err := h.service.AcceptExtraction(projectID, extractionID, body.SaveToMemory, &userID); err != nil {
		if errors.Is(err, ErrExtractionNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "extraction not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to accept extraction"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "extraction accepted"})
}

func (h *Handler) RejectExtraction(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid project id"})
		return
	}
	extractionID, err := uuid.Parse(chi.URLParam(r, "extractionId"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid extraction id"})
		return
	}

	if err := h.service.RejectExtraction(projectID, extractionID); err != nil {
		if errors.Is(err, ErrExtractionNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "extraction not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to reject extraction"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "extraction rejected"})
}
