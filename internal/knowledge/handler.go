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

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) ListDocuments(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid project id")
		return
	}

	docs, err := h.service.ListDocuments(projectID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to list documents")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]interface{}{"documents": docs})
}

func (h *Handler) CreateDocument(w http.ResponseWriter, r *http.Request) {
	httputil.LimitBody(w, r, httputil.LargeBodySize)

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

	var input CreateDocumentInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if input.Title == "" || input.Content == "" {
		httputil.WriteError(w, http.StatusBadRequest, "title and content are required")
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
		httputil.WriteError(w, http.StatusInternalServerError, "failed to create document")
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, map[string]interface{}{"document": doc})
}

func (h *Handler) GetDocument(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid project id")
		return
	}
	docID, err := uuid.Parse(chi.URLParam(r, "docId"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid document id")
		return
	}

	doc, err := h.service.GetDocument(projectID, docID)
	if err != nil {
		if errors.Is(err, ErrDocumentNotFound) {
			httputil.WriteError(w, http.StatusNotFound, "document not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "failed to get document")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]interface{}{"document": doc})
}

func (h *Handler) DeleteDocument(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid project id")
		return
	}
	docID, err := uuid.Parse(chi.URLParam(r, "docId"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid document id")
		return
	}

	if err := h.service.DeleteDocument(projectID, docID); err != nil {
		if errors.Is(err, ErrDocumentNotFound) {
			httputil.WriteError(w, http.StatusNotFound, "document not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "failed to delete document")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]string{"message": "document deleted"})
}

func (h *Handler) GetChunks(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid project id")
		return
	}
	docID, err := uuid.Parse(chi.URLParam(r, "docId"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid document id")
		return
	}

	chunks, err := h.service.GetChunks(projectID, docID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to get chunks")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]interface{}{"chunks": chunks})
}

func (h *Handler) SearchChunks(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid project id")
		return
	}

	q := r.URL.Query().Get("q")
	if q == "" {
		httputil.WriteError(w, http.StatusBadRequest, "search query is required")
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	chunks, err := h.service.SearchChunks(projectID, q, limit)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "search failed")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]interface{}{"chunks": chunks})
}
