package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// AgentTool is a reusable tool definition scoped to a project.
type AgentTool struct {
	ID          uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	ProjectID   uuid.UUID      `gorm:"type:uuid;not null;index" json:"project_id"`
	Name        string         `gorm:"size:100;not null" json:"name"`
	Description string         `gorm:"type:text;default:''" json:"description"`
	ToolType    string         `gorm:"size:50;not null;default:function" json:"tool_type"` // function, http, shell
	Definition  datatypes.JSON `gorm:"type:jsonb;default:'{}'" json:"definition"`
	Enabled     bool           `gorm:"default:true" json:"enabled"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

func (t *AgentTool) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}

// AgentToolBinding links an agent to a tool.
type AgentToolBinding struct {
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	AgentID   uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_atb_agent_tool" json:"agent_id"`
	ToolID    uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_atb_agent_tool" json:"tool_id"`
	CreatedAt time.Time `json:"created_at"`

	Tool AgentTool `gorm:"foreignKey:ToolID" json:"tool,omitempty"`
}

func (b *AgentToolBinding) BeforeCreate(tx *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}

// ToolExecution logs a single tool invocation.
type ToolExecution struct {
	ID           uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	AgentID      uuid.UUID      `gorm:"type:uuid;not null;index" json:"agent_id"`
	ToolID       uuid.UUID      `gorm:"type:uuid;not null" json:"tool_id"`
	ProjectID    uuid.UUID      `gorm:"type:uuid;not null;index" json:"project_id"`
	TriggeredBy  *uuid.UUID     `gorm:"type:uuid" json:"triggered_by,omitempty"`
	Input        datatypes.JSON `gorm:"type:jsonb;default:'{}'" json:"input"`
	Output       string         `gorm:"type:text;default:''" json:"output"`
	Status       string         `gorm:"size:20;default:success" json:"status"`
	DurationMs   int            `gorm:"default:0" json:"duration_ms"`
	ErrorMessage *string        `gorm:"type:text" json:"error_message,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
}

func (e *ToolExecution) BeforeCreate(tx *gorm.DB) error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	return nil
}

// UsageRecord tracks per-invocation token + cost data.
type UsageRecord struct {
	ID             uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	UserID         uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	ProjectID      uuid.UUID  `gorm:"type:uuid;not null;index" json:"project_id"`
	AgentID        *uuid.UUID `gorm:"type:uuid" json:"agent_id,omitempty"`
	Provider       string     `gorm:"size:50;not null" json:"provider"`
	Model          string     `gorm:"size:100;not null" json:"model"`
	InputTokens    int        `gorm:"default:0" json:"input_tokens"`
	OutputTokens   int        `gorm:"default:0" json:"output_tokens"`
	TotalTokens    int        `gorm:"default:0" json:"total_tokens"`
	CostMicrocents int64      `gorm:"default:0" json:"cost_microcents"`
	ToolCallsCount int        `gorm:"default:0" json:"tool_calls_count"`
	CreatedAt      time.Time  `json:"created_at"`
}

func (u *UsageRecord) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

// CostBudget sets an optional monthly spend cap per project.
type CostBudget struct {
	ID                     uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	ProjectID              uuid.UUID `gorm:"type:uuid;not null;uniqueIndex" json:"project_id"`
	MonthlyLimitMicrocents int64     `gorm:"default:0" json:"monthly_limit_microcents"`
	AlertThresholdPct      int       `gorm:"default:80" json:"alert_threshold_pct"`
	CreatedAt              time.Time `json:"created_at"`
	UpdatedAt              time.Time `json:"updated_at"`
}

func (c *CostBudget) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}
