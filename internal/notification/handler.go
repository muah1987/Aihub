package notification

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/muah1987/Aihub/internal/auth"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  512,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

type Handler struct {
	service *Service
	hub     *Hub
}

func NewHandler(svc *Service, hub *Hub) *Handler {
	return &Handler{service: svc, hub: hub}
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	unreadOnly := r.URL.Query().Get("unread") == "true"
	limit := 30
	if l := r.URL.Query().Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 {
			limit = v
		}
	}
	ns, err := h.service.List(userID, unreadOnly, limit)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	count, _ := h.service.CountUnread(userID)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"notifications": ns,
		"unread_count":  count,
	})
}

func (h *Handler) MarkRead(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	if err := h.service.MarkRead(userID, id); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "marked read"})
}

func (h *Handler) MarkAllRead(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	h.service.MarkAllRead(userID)
	writeJSON(w, http.StatusOK, map[string]string{"message": "all marked read"})
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	h.service.Delete(userID, id)
	writeJSON(w, http.StatusOK, map[string]string{"message": "deleted"})
}

// WebSocket upgrades the connection and registers the user for live pushes.
func (h *Handler) WebSocket(w http.ResponseWriter, r *http.Request) {
	jwtSvc := auth.GetJWTServiceFromContext(r.Context())
	if jwtSvc == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	token := r.URL.Query().Get("token")
	claims, err := jwtSvc.ValidateToken(token)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	userID := claims.UserID

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	h.hub.Register(userID, conn)
	defer func() {
		h.hub.Unregister(userID, conn)
		conn.Close()
	}()

	// Send unread count immediately on connect
	count, _ := h.service.CountUnread(userID)
	conn.WriteJSON(map[string]interface{}{"event": "connected", "unread_count": count})

	// Keep connection alive; client sends pings
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}
	}
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
