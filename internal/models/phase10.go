package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// UserPreference stores per-user UI and theme settings.
type UserPreference struct {
	ID                 uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	UserID             uuid.UUID      `gorm:"type:uuid;not null;uniqueIndex" json:"user_id"`
	Theme              string         `gorm:"size:20;not null;default:system" json:"theme"`
	AccentColor        string         `gorm:"size:20;default:blue" json:"accent_color"`
	SidebarCollapsed   bool           `gorm:"default:false" json:"sidebar_collapsed"`
	CompactMode        bool           `gorm:"default:false" json:"compact_mode"`
	EditorFontSize     int            `gorm:"default:14" json:"editor_font_size"`
	NotificationsSound bool           `gorm:"default:true" json:"notifications_sound"`
	Locale             string         `gorm:"size:10;default:en" json:"locale"`
	Timezone           string         `gorm:"size:50;default:UTC" json:"timezone"`
	Settings           datatypes.JSON `gorm:"type:jsonb;default:'{}'" json:"settings"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
}

func (u *UserPreference) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

// PinnedProject marks a project as favorited by a user.
type PinnedProject struct {
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	ProjectID uuid.UUID `gorm:"type:uuid;not null" json:"project_id"`
	PinOrder  int       `gorm:"default:0" json:"pin_order"`
	CreatedAt time.Time `json:"created_at"`

	Project Project `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
}

func (p *PinnedProject) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

// PromptVersion stores a versioned system prompt for A/B testing.
type PromptVersion struct {
	ID               uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	AgentID          uuid.UUID      `gorm:"type:uuid;not null;index" json:"agent_id"`
	VersionLabel     string         `gorm:"size:100;not null" json:"version_label"`
	SystemPrompt     string         `gorm:"type:text;not null" json:"system_prompt"`
	ModelConfig      datatypes.JSON `gorm:"type:jsonb;default:'{}'" json:"model_config"`
	IsActive         bool           `gorm:"default:false" json:"is_active"`
	TotalInvocations int            `gorm:"default:0" json:"total_invocations"`
	AvgRating        float64        `gorm:"type:decimal(3,2);default:0" json:"avg_rating"`
	CreatedBy        *uuid.UUID     `gorm:"type:uuid" json:"created_by,omitempty"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
}

func (p *PromptVersion) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}
