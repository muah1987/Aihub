package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type TeamMemory struct {
	ID             uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	ProjectID      uuid.UUID      `gorm:"type:uuid;not null;index" json:"project_id"`
	Category       string         `gorm:"size:100;not null;default:general" json:"category"`
	Key            string         `gorm:"size:255;not null" json:"key"`
	Content        string         `gorm:"type:text;not null" json:"content"`
	ContentType    string         `gorm:"size:50;default:text" json:"content_type"`
	Metadata       datatypes.JSON `gorm:"type:jsonb;default:'{}'" json:"metadata"`
	CreatedBy      *uuid.UUID     `gorm:"type:uuid" json:"created_by,omitempty"`
	CreatedByAgent *uuid.UUID     `gorm:"type:uuid" json:"created_by_agent,omitempty"`
	Pinned         bool           `gorm:"default:false" json:"pinned"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`

	Project Project `gorm:"foreignKey:ProjectID" json:"-"`
}

func (m *TeamMemory) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}
