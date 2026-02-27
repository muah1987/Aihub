package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TerminalSession struct {
	ID           uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	ProjectID    uuid.UUID `gorm:"type:uuid;not null;index" json:"project_id"`
	UserID       uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	ContainerID  string    `gorm:"size:255" json:"container_id"`
	Status       string    `gorm:"size:20;default:starting" json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	LastActivity time.Time `json:"last_activity"`

	Project Project `gorm:"foreignKey:ProjectID" json:"-"`
	User    User    `gorm:"foreignKey:UserID" json:"-"`
}

func (t *TerminalSession) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}
