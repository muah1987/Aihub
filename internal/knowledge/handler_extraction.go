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

func (h *Handler) ListExtractions(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid project id")
		return
	}

	status := r.URL.Query().Get("status")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	extractions, err := h.service.ListExtractions(projectID, status, limit)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to list extractions")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]interface{}{"extractions": extractions})
}

func (h *Handler) AcceptExtraction(w http.ResponseWriter, r *http.Request) {
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
	extractionID, err := uuid.Parse(chi.URLParam(r, "extractionId"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid extraction id")
		return
	}

	var body struct {
		SaveToMemory bool `json:"save_to_memory"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.service.AcceptExtraction(projectID, extractionID, body.SaveToMemory, &userID); err != nil {
		if errors.Is(err, ErrExtractionNotFound) {
			httputil.WriteError(w, http.StatusNotFound, "extraction not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "failed to accept extraction")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]string{"message": "extraction accepted"})
}

func (h *Handler) RejectExtraction(w http.ResponseWriter, r *http.Request) {
	_, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid project id")
		return
	}
	extractionID, err := uuid.Parse(chi.URLParam(r, "extractionId"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid extraction id")
		return
	}

	if err := h.service.RejectExtraction(projectID, extractionID); err != nil {
		if errors.Is(err, ErrExtractionNotFound) {
			httputil.WriteError(w, http.StatusNotFound, "extraction not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "failed to reject extraction")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]string{"message": "extraction rejected"})
}
