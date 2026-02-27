package memory

import (
	"encoding/json"
	"errors"
	"net/http"

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

func (h *Handler) Set(w http.ResponseWriter, r *http.Request) {
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

	var input SetInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if input.Key == "" || input.Content == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "key and content are required"})
		return
	}

	mem, err := h.service.Set(projectID, &userID, &input)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to set memory"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"memory": mem})
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid project id"})
		return
	}

	category := chi.URLParam(r, "category")
	key := chi.URLParam(r, "key")

	mem, err := h.service.Get(projectID, category, key)
	if err != nil {
		if errors.Is(err, ErrMemoryNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "memory not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get memory"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"memory": mem})
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid project id"})
		return
	}

	category := r.URL.Query().Get("category")

	memories, err := h.service.List(projectID, category)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list memories"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"memories": memories})
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid project id"})
		return
	}

	memoryID, err := uuid.Parse(chi.URLParam(r, "memoryId"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid memory id"})
		return
	}

	if err := h.service.Delete(projectID, memoryID); err != nil {
		if errors.Is(err, ErrMemoryNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "memory not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to delete memory"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "memory deleted"})
}

func (h *Handler) Search(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid project id"})
		return
	}

	q := r.URL.Query().Get("q")
	if q == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "search query is required"})
		return
	}

	memories, err := h.service.Search(projectID, q)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "search failed"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"memories": memories})
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
