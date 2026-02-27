package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// IntegrationConnection stores credentials for an external platform (Slack, Discord, webhook).
type IntegrationConnection struct {
	ID           uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	ProjectID    uuid.UUID      `gorm:"type:uuid;not null;index" json:"project_id"`
	CreatedBy    *uuid.UUID     `gorm:"type:uuid" json:"created_by,omitempty"`
	Platform     string         `gorm:"size:30;not null" json:"platform"`
	Name         string         `gorm:"size:200;not null" json:"name"`
	Config       datatypes.JSON `gorm:"type:jsonb;not null;default:'{}'" json:"config"`
	Credentials  string         `gorm:"type:text" json:"-"`
	ChannelID    string         `gorm:"size:200" json:"channel_id,omitempty"`
	Enabled      bool           `gorm:"default:true" json:"enabled"`
	Status       string         `gorm:"size:20;default:pending" json:"status"`
	ErrorMessage string         `gorm:"type:text" json:"error_message,omitempty"`
	LastUsedAt   *time.Time     `json:"last_used_at,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}

func (i *IntegrationConnection) BeforeCreate(tx *gorm.DB) error {
	if i.ID == uuid.Nil {
		i.ID = uuid.New()
	}
	return nil
}

// NotificationRule defines per-user notification preferences.
type NotificationRule struct {
	ID              uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	UserID          uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	ProjectID       *uuid.UUID `gorm:"type:uuid" json:"project_id,omitempty"`
	EventType       string     `gorm:"size:100;not null" json:"event_type"`
	Channel         string     `gorm:"size:30;not null" json:"channel"`
	Enabled         bool       `gorm:"default:true" json:"enabled"`
	MinSeverity     string     `gorm:"size:20;default:info" json:"min_severity"`
	QuietHoursStart *int       `gorm:"type:smallint" json:"quiet_hours_start,omitempty"`
	QuietHoursEnd   *int       `gorm:"type:smallint" json:"quiet_hours_end,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

func (n *NotificationRule) BeforeCreate(tx *gorm.DB) error {
	if n.ID == uuid.Nil {
		n.ID = uuid.New()
	}
	return nil
}

// EmailDigest configures email summary preferences.
type EmailDigest struct {
	ID                   uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	UserID               uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	ProjectID            *uuid.UUID `gorm:"type:uuid" json:"project_id,omitempty"`
	Frequency            string     `gorm:"size:20;not null;default:daily" json:"frequency"`
	IncludeDeployments   bool       `gorm:"default:true" json:"include_deployments"`
	IncludeChatSummary   bool       `gorm:"default:true" json:"include_chat_summary"`
	IncludeAgentActivity bool       `gorm:"default:true" json:"include_agent_activity"`
	IncludeMonitoring    bool       `gorm:"default:true" json:"include_monitoring"`
	LastSentAt           *time.Time `json:"last_sent_at,omitempty"`
	NextSendAt           *time.Time `json:"next_send_at,omitempty"`
	Enabled              bool       `gorm:"default:true" json:"enabled"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

func (e *EmailDigest) BeforeCreate(tx *gorm.DB) error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	return nil
}

// OutboundEvent tracks webhook dispatches to external systems.
type OutboundEvent struct {
	ID            uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	ProjectID     uuid.UUID      `gorm:"type:uuid;not null;index" json:"project_id"`
	IntegrationID *uuid.UUID     `gorm:"type:uuid" json:"integration_id,omitempty"`
	EventType     string         `gorm:"size:100;not null" json:"event_type"`
	Payload       datatypes.JSON `gorm:"type:jsonb;not null;default:'{}'" json:"payload"`
	Status        string         `gorm:"size:20;not null;default:pending" json:"status"`
	HTTPStatus    int            `json:"http_status,omitempty"`
	ResponseBody  string         `gorm:"type:text" json:"response_body,omitempty"`
	Attempts      int            `gorm:"default:0" json:"attempts"`
	MaxAttempts   int            `gorm:"default:3" json:"max_attempts"`
	NextRetryAt   *time.Time     `json:"next_retry_at,omitempty"`
	SentAt        *time.Time     `json:"sent_at,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
}

func (o *OutboundEvent) BeforeCreate(tx *gorm.DB) error {
	if o.ID == uuid.Nil {
		o.ID = uuid.New()
	}
	return nil
}
