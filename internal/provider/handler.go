package provider

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

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	var input CreateInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if input.ProviderType == "" || input.ProviderName == "" || input.AccessToken == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "provider_type, provider_name, and access_token are required"})
		return
	}

	conn, err := h.service.Create(userID, &input)
	if err != nil {
		if errors.Is(err, ErrInvalidProvider) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "unsupported provider type"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create connection"})
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"provider": conn.ToResponse(),
	})
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	connections, err := h.service.List(userID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list connections"})
		return
	}

	responses := make([]interface{}, len(connections))
	for i, c := range connections {
		resp := c.ToResponse()
		responses[i] = resp
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"providers": responses,
	})
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	connID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid provider id"})
		return
	}

	if err := h.service.Delete(userID, connID); err != nil {
		if errors.Is(err, ErrProviderNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "provider connection not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to delete connection"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "provider disconnected"})
}

func (h *Handler) Validate(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	connID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid provider id"})
		return
	}

	token, err := h.service.GetDecryptedToken(userID, connID)
	if err != nil {
		if errors.Is(err, ErrProviderNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "provider connection not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to retrieve token"})
		return
	}

	conn, _ := h.service.GetByID(userID, connID)

	switch conn.ProviderType {
	case "github":
		client := NewGitHubClient(token)
		ghUser, err := client.ValidateToken()
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "token validation failed", "details": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"valid":    true,
			"provider": "github",
			"user":     ghUser,
		})
	default:
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"valid":   true,
			"message": "token stored (validation not implemented for this provider)",
		})
	}
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
