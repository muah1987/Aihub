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

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) ListDocuments(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid project id"})
		return
	}

	docs, err := h.service.ListDocuments(projectID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list documents"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"documents": docs})
}

func (h *Handler) CreateDocument(w http.ResponseWriter, r *http.Request) {
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

	var input CreateDocumentInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if input.Title == "" || input.Content == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "title and content are required"})
		return
	}
	if input.FileType == "" {
		input.FileType = "text"
	}
	if input.FileName == "" {
		input.FileName = input.Title
	}

	doc, err := h.service.CreateDocument(projectID, &userID, &input)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create document"})
		return
	}
	writeJSON(w, http.StatusCreated, map[string]interface{}{"document": doc})
}

func (h *Handler) GetDocument(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid project id"})
		return
	}
	docID, err := uuid.Parse(chi.URLParam(r, "docId"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid document id"})
		return
	}

	doc, err := h.service.GetDocument(projectID, docID)
	if err != nil {
		if errors.Is(err, ErrDocumentNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "document not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get document"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"document": doc})
}

func (h *Handler) DeleteDocument(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid project id"})
		return
	}
	docID, err := uuid.Parse(chi.URLParam(r, "docId"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid document id"})
		return
	}

	if err := h.service.DeleteDocument(projectID, docID); err != nil {
		if errors.Is(err, ErrDocumentNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "document not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to delete document"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "document deleted"})
}

func (h *Handler) GetChunks(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid project id"})
		return
	}
	docID, err := uuid.Parse(chi.URLParam(r, "docId"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid document id"})
		return
	}

	chunks, err := h.service.GetChunks(projectID, docID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get chunks"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"chunks": chunks})
}

func (h *Handler) SearchChunks(w http.ResponseWriter, r *http.Request) {
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
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	chunks, err := h.service.SearchChunks(projectID, q, limit)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "search failed"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"chunks": chunks})
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
