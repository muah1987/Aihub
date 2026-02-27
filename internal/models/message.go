package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Message struct {
	ID          uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	ProjectID   uuid.UUID      `gorm:"type:uuid;not null;index:idx_messages_project_created" json:"project_id"`
	SenderType  string         `gorm:"size:20;not null" json:"sender_type"`
	SenderID    *uuid.UUID     `gorm:"type:uuid" json:"sender_id,omitempty"`
	SenderName  string         `gorm:"size:100" json:"sender_name"`
	Content     string         `gorm:"type:text;not null" json:"content"`
	MessageType string         `gorm:"size:30;default:text" json:"message_type"`
	Metadata    datatypes.JSON `gorm:"type:jsonb;default:'{}'" json:"metadata"`
	CreatedAt   time.Time      `gorm:"index:idx_messages_project_created" json:"created_at"`

	Project Project `gorm:"foreignKey:ProjectID" json:"-"`
}

func (m *Message) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}
