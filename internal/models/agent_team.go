package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type AgentTeam struct {
	ID            uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	ProjectID     uuid.UUID      `gorm:"type:uuid;not null;index" json:"project_id"`
	Name          string         `gorm:"size:255;not null" json:"name"`
	Description   string         `gorm:"type:text" json:"description"`
	LeaderAgentID *uuid.UUID     `gorm:"type:uuid" json:"leader_agent_id,omitempty"`
	Strategy      string         `gorm:"size:50;default:sequential" json:"strategy"`
	Settings      datatypes.JSON `gorm:"type:jsonb;default:'{}'" json:"settings"`
	Status        string         `gorm:"size:20;default:idle" json:"status"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`

	Project     Project `gorm:"foreignKey:ProjectID" json:"-"`
	LeaderAgent *Agent  `gorm:"foreignKey:LeaderAgentID" json:"leader_agent,omitempty"`
}

func (t *AgentTeam) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}

type AgentTeamMember struct {
	ID         uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	TeamID     uuid.UUID `gorm:"type:uuid;not null;index" json:"team_id"`
	AgentID    uuid.UUID `gorm:"type:uuid;not null;index" json:"agent_id"`
	RoleInTeam string    `gorm:"size:100;default:member" json:"role_in_team"`
	Priority   int       `gorm:"default:0" json:"priority"`
	CreatedAt  time.Time `json:"created_at"`

	Team  AgentTeam `gorm:"foreignKey:TeamID" json:"-"`
	Agent Agent     `gorm:"foreignKey:AgentID" json:"agent,omitempty"`
}

func (m *AgentTeamMember) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}

type AgentTask struct {
	ID              uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	TeamID          uuid.UUID      `gorm:"type:uuid;not null;index" json:"team_id"`
	ParentTaskID    *uuid.UUID     `gorm:"type:uuid" json:"parent_task_id,omitempty"`
	AssignedAgentID *uuid.UUID     `gorm:"type:uuid" json:"assigned_agent_id,omitempty"`
	Title           string         `gorm:"size:500;not null" json:"title"`
	Description     string         `gorm:"type:text" json:"description"`
	Prompt          string         `gorm:"type:text" json:"prompt"`
	Result          string         `gorm:"type:text" json:"result"`
	Status          string         `gorm:"size:20;default:pending" json:"status"`
	Priority        int            `gorm:"default:0" json:"priority"`
	Metadata        datatypes.JSON `gorm:"type:jsonb;default:'{}'" json:"metadata"`
	CreatedAt       time.Time      `json:"created_at"`
	StartedAt       *time.Time     `json:"started_at,omitempty"`
	CompletedAt     *time.Time     `json:"completed_at,omitempty"`

	Team          AgentTeam `gorm:"foreignKey:TeamID" json:"-"`
	AssignedAgent *Agent    `gorm:"foreignKey:AssignedAgentID" json:"assigned_agent,omitempty"`
}

func (t *AgentTask) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}
