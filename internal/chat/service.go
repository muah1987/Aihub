package chat

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/muah1987/Aihub/internal/models"
	"gorm.io/gorm"
)

type SendInput struct {
	Content     string `json:"content" validate:"required"`
	MessageType string `json:"message_type"`
}

type Service struct {
	db  *gorm.DB
	hub *Hub
}

func NewService(db *gorm.DB, hub *Hub) *Service {
	return &Service{db: db, hub: hub}
}

func (s *Service) SendMessage(projectID, senderID uuid.UUID, senderName, senderType string, input *SendInput) (*models.Message, error) {
	msgType := input.MessageType
	if msgType == "" {
		msgType = "text"
	}

	msg := &models.Message{
		ProjectID:   projectID,
		SenderType:  senderType,
		SenderID:    &senderID,
		SenderName:  senderName,
		Content:     input.Content,
		MessageType: msgType,
	}

	if err := s.db.Create(msg).Error; err != nil {
		return nil, fmt.Errorf("failed to save message: %w", err)
	}

	// Broadcast to all WebSocket clients in this project
	s.hub.Broadcast(projectID, msg)

	return msg, nil
}

func (s *Service) SendSystemMessage(projectID uuid.UUID, content, messageType string) (*models.Message, error) {
	msg := &models.Message{
		ProjectID:   projectID,
		SenderType:  "system",
		SenderName:  "System",
		Content:     content,
		MessageType: messageType,
	}

	if err := s.db.Create(msg).Error; err != nil {
		return nil, fmt.Errorf("failed to save system message: %w", err)
	}

	s.hub.Broadcast(projectID, msg)
	return msg, nil
}

func (s *Service) GetHistory(projectID uuid.UUID, before *time.Time, limit int) ([]models.Message, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	query := s.db.Where("project_id = ?", projectID)
	if before != nil {
		query = query.Where("created_at < ?", before)
	}

	var messages []models.Message
	if err := query.Order("created_at DESC").Limit(limit).Find(&messages).Error; err != nil {
		return nil, fmt.Errorf("failed to get message history: %w", err)
	}

	// Reverse to get chronological order
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}
