package notification

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/muah1987/Aihub/internal/models"
	"gorm.io/gorm"
)

var ErrNotFound = errors.New("notification not found")

// Hub maintains live WebSocket connections per user for real-time pushes.
type Hub struct {
	mu      sync.RWMutex
	clients map[uuid.UUID]map[*websocket.Conn]struct{}
}

func NewHub() *Hub {
	return &Hub{clients: make(map[uuid.UUID]map[*websocket.Conn]struct{})}
}

func (h *Hub) Register(userID uuid.UUID, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.clients[userID]; !ok {
		h.clients[userID] = make(map[*websocket.Conn]struct{})
	}
	h.clients[userID][conn] = struct{}{}
}

func (h *Hub) Unregister(userID uuid.UUID, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if conns, ok := h.clients[userID]; ok {
		delete(conns, conn)
	}
}

func (h *Hub) Push(userID uuid.UUID, n *models.Notification) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	conns := h.clients[userID]
	payload, _ := json.Marshal(map[string]interface{}{
		"event":        "notification",
		"notification": n,
	})
	for conn := range conns {
		conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
		conn.WriteMessage(websocket.TextMessage, payload)
	}
}

// Service manages notification persistence and delivery.
type Service struct {
	db  *gorm.DB
	hub *Hub
}

func NewService(db *gorm.DB, hub *Hub) *Service {
	return &Service{db: db, hub: hub}
}

func (s *Service) Create(userID uuid.UUID, nType, title, message string, data interface{}) (*models.Notification, error) {
	raw, _ := json.Marshal(data)
	n := &models.Notification{
		UserID:  userID,
		Type:    nType,
		Title:   title,
		Message: message,
		Data:    raw,
	}
	if err := s.db.Create(n).Error; err != nil {
		return nil, fmt.Errorf("create notification: %w", err)
	}
	// Push live if user is connected
	s.hub.Push(userID, n)
	return n, nil
}

// NotifyProjectOwner finds the project's owner and sends them a notification.
func (s *Service) NotifyProjectOwner(projectID uuid.UUID, nType, title, message string) {
	var project models.Project
	if err := s.db.First(&project, "id = ?", projectID).Error; err != nil {
		return
	}
	s.Create(project.UserID, nType, title, message, nil) //nolint
}

func (s *Service) List(userID uuid.UUID, unreadOnly bool, limit int) ([]models.Notification, error) {
	q := s.db.Where("user_id = ?", userID)
	if unreadOnly {
		q = q.Where("read = false")
	}
	var ns []models.Notification
	if err := q.Order("created_at DESC").Limit(limit).Find(&ns).Error; err != nil {
		return nil, err
	}
	return ns, nil
}

func (s *Service) MarkRead(userID, id uuid.UUID) error {
	res := s.db.Model(&models.Notification{}).
		Where("id = ? AND user_id = ?", id, userID).
		Update("read", true)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Service) MarkAllRead(userID uuid.UUID) error {
	return s.db.Model(&models.Notification{}).
		Where("user_id = ? AND read = false", userID).
		Update("read", true).Error
}

func (s *Service) CountUnread(userID uuid.UUID) (int64, error) {
	var count int64
	err := s.db.Model(&models.Notification{}).
		Where("user_id = ? AND read = false", userID).
		Count(&count).Error
	return count, err
}

func (s *Service) Delete(userID, id uuid.UUID) error {
	res := s.db.Where("id = ? AND user_id = ?", id, userID).Delete(&models.Notification{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
