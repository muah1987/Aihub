package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Agent struct {
	ID                   uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	ProjectID            uuid.UUID      `gorm:"type:uuid;not null;index" json:"project_id"`
	Name                 string         `gorm:"size:100;not null" json:"name"`
	Role                 string         `gorm:"size:100;not null" json:"role"`
	Model                string         `gorm:"size:100;not null" json:"model"`
	ProviderConnectionID *uuid.UUID     `gorm:"type:uuid" json:"provider_connection_id,omitempty"`
	SystemPrompt         string         `gorm:"type:text" json:"system_prompt"`
	Temperature          float64        `gorm:"type:decimal(3,2);default:0.7" json:"temperature"`
	MaxTokens            int            `gorm:"default:4096" json:"max_tokens"`
	Status               string         `gorm:"size:20;default:idle" json:"status"`
	TokenUsageTotal      int64          `gorm:"default:0" json:"token_usage_total"`
	LastHeartbeat        *time.Time     `json:"last_heartbeat,omitempty"`
	Settings             datatypes.JSON `gorm:"type:jsonb;default:'{}'" json:"settings"`
	CreatedAt            time.Time      `json:"created_at"`
	UpdatedAt            time.Time      `json:"updated_at"`

	Project            Project             `gorm:"foreignKey:ProjectID" json:"-"`
	ProviderConnection *ProviderConnection `gorm:"foreignKey:ProviderConnectionID" json:"-"`
}

func (a *Agent) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}
