package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Webhook triggers auto-deploy when GitHub pushes to a branch.
type Webhook struct {
	ID              uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	ProjectID       uuid.UUID  `gorm:"type:uuid;not null;index" json:"project_id"`
	VPSTargetID     *uuid.UUID `gorm:"type:uuid" json:"vps_target_id,omitempty"`
	Secret          string     `gorm:"type:text;not null" json:"-"` // HMAC secret, never sent to client
	Branch          string     `gorm:"size:255;default:main" json:"branch"`
	Active          bool       `gorm:"default:true" json:"active"`
	LastTriggeredAt *time.Time `json:"last_triggered_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

func (w *Webhook) BeforeCreate(tx *gorm.DB) error {
	if w.ID == uuid.Nil {
		w.ID = uuid.New()
	}
	return nil
}

// PipelineStage is an ordered shell step run on the VPS during deploy.
type PipelineStage struct {
	ID          uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	VPSTargetID uuid.UUID `gorm:"type:uuid;not null;index" json:"vps_target_id"`
	Name        string    `gorm:"size:255;not null" json:"name"`
	Command     string    `gorm:"type:text;not null" json:"command"`
	StageOrder  int       `gorm:"default:0" json:"stage_order"`
	OnFailure   string    `gorm:"size:20;default:abort" json:"on_failure"` // "abort" | "continue"
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (p *PipelineStage) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

// Notification is an in-app alert for a user.
type Notification struct {
	ID        uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	UserID    uuid.UUID      `gorm:"type:uuid;not null;index" json:"user_id"`
	Type      string         `gorm:"size:50;not null" json:"type"`
	Title     string         `gorm:"size:255;not null" json:"title"`
	Message   string         `gorm:"type:text;not null" json:"message"`
	Read      bool           `gorm:"default:false" json:"read"`
	Data      datatypes.JSON `gorm:"type:jsonb;default:'{}'" json:"data,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
}

func (n *Notification) BeforeCreate(tx *gorm.DB) error {
	if n.ID == uuid.Nil {
		n.ID = uuid.New()
	}
	return nil
}

// ServerMetric is a point-in-time snapshot of VPS health.
type ServerMetric struct {
	ID            uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	VPSTargetID   uuid.UUID `gorm:"type:uuid;not null;index" json:"vps_target_id"`
	CPUPercent    float32   `gorm:"default:0" json:"cpu_percent"`
	MemPercent    float32   `gorm:"default:0" json:"mem_percent"`
	DiskPercent   float32   `gorm:"default:0" json:"disk_percent"`
	LoadAvg       string    `gorm:"size:50;default:''" json:"load_avg"`
	UptimeSeconds int64     `gorm:"default:0" json:"uptime_seconds"`
	RecordedAt    time.Time `gorm:"default:NOW()" json:"recorded_at"`
}

func (s *ServerMetric) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}
