package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Workflow is a reusable automation chain scoped to a project.
type Workflow struct {
	ID            uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	ProjectID     uuid.UUID      `gorm:"type:uuid;not null;index" json:"project_id"`
	Name          string         `gorm:"size:255;not null" json:"name"`
	Description   string         `gorm:"type:text;default:''" json:"description"`
	TriggerType   string         `gorm:"size:50;not null;default:manual" json:"trigger_type"`
	TriggerConfig datatypes.JSON `gorm:"type:jsonb;default:'{}'" json:"trigger_config"`
	Enabled       bool           `gorm:"default:true" json:"enabled"`
	LastRunAt     *time.Time     `json:"last_run_at,omitempty"`
	LastRunStatus *string        `gorm:"size:20" json:"last_run_status,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`

	Steps []WorkflowStep `gorm:"foreignKey:WorkflowID" json:"steps,omitempty"`
}

func (w *Workflow) BeforeCreate(tx *gorm.DB) error {
	if w.ID == uuid.Nil {
		w.ID = uuid.New()
	}
	return nil
}

// WorkflowStep is an ordered action within a workflow.
type WorkflowStep struct {
	ID         uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	WorkflowID uuid.UUID      `gorm:"type:uuid;not null;index" json:"workflow_id"`
	Name       string         `gorm:"size:255;not null" json:"name"`
	StepOrder  int            `gorm:"default:0" json:"step_order"`
	StepType   string         `gorm:"size:50;not null" json:"step_type"`
	Config     datatypes.JSON `gorm:"type:jsonb;default:'{}'" json:"config"`
	OnFailure  string         `gorm:"size:20;default:abort" json:"on_failure"`
	TimeoutSec int            `gorm:"default:300" json:"timeout_sec"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
}

func (s *WorkflowStep) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

// WorkflowRun tracks a single execution of a workflow.
type WorkflowRun struct {
	ID           uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	WorkflowID   uuid.UUID  `gorm:"type:uuid;not null;index" json:"workflow_id"`
	ProjectID    uuid.UUID  `gorm:"type:uuid;not null;index" json:"project_id"`
	TriggeredBy  *uuid.UUID `gorm:"type:uuid" json:"triggered_by,omitempty"`
	TriggerType  string     `gorm:"size:50;not null" json:"trigger_type"`
	Status       string     `gorm:"size:20;default:running" json:"status"`
	StartedAt    time.Time  `gorm:"default:NOW()" json:"started_at"`
	FinishedAt   *time.Time `json:"finished_at,omitempty"`
	ErrorMessage *string    `gorm:"type:text" json:"error_message,omitempty"`

	StepRuns []WorkflowStepRun `gorm:"foreignKey:WorkflowRunID" json:"step_runs,omitempty"`
}

func (r *WorkflowRun) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}

// WorkflowStepRun tracks execution of a single step within a workflow run.
type WorkflowStepRun struct {
	ID            uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	WorkflowRunID uuid.UUID  `gorm:"type:uuid;not null;index" json:"workflow_run_id"`
	StepID        uuid.UUID  `gorm:"type:uuid;not null" json:"step_id"`
	Status        string     `gorm:"size:20;default:pending" json:"status"`
	Output        string     `gorm:"type:text;default:''" json:"output"`
	StartedAt     *time.Time `json:"started_at,omitempty"`
	FinishedAt    *time.Time `json:"finished_at,omitempty"`
	ErrorMessage  *string    `gorm:"type:text" json:"error_message,omitempty"`
}

func (sr *WorkflowStepRun) BeforeCreate(tx *gorm.DB) error {
	if sr.ID == uuid.Nil {
		sr.ID = uuid.New()
	}
	return nil
}

// ScheduledJob is a cron-like recurring task.
type ScheduledJob struct {
	ID            uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	ProjectID     uuid.UUID      `gorm:"type:uuid;not null;index" json:"project_id"`
	Name          string         `gorm:"size:255;not null" json:"name"`
	Description   string         `gorm:"type:text;default:''" json:"description"`
	CronExpr      string         `gorm:"size:100;not null" json:"cron_expr"`
	JobType       string         `gorm:"size:50;not null" json:"job_type"`
	Config        datatypes.JSON `gorm:"type:jsonb;default:'{}'" json:"config"`
	Enabled       bool           `gorm:"default:true" json:"enabled"`
	LastRunAt     *time.Time     `json:"last_run_at,omitempty"`
	LastRunStatus *string        `gorm:"size:20" json:"last_run_status,omitempty"`
	NextRunAt     *time.Time     `json:"next_run_at,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}

func (j *ScheduledJob) BeforeCreate(tx *gorm.DB) error {
	if j.ID == uuid.Nil {
		j.ID = uuid.New()
	}
	return nil
}

// ScheduledJobRun tracks a single execution of a scheduled job.
type ScheduledJobRun struct {
	ID           uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	JobID        uuid.UUID  `gorm:"type:uuid;not null;index" json:"job_id"`
	Status       string     `gorm:"size:20;default:running" json:"status"`
	Output       string     `gorm:"type:text;default:''" json:"output"`
	StartedAt    time.Time  `gorm:"default:NOW()" json:"started_at"`
	FinishedAt   *time.Time `json:"finished_at,omitempty"`
	ErrorMessage *string    `gorm:"type:text" json:"error_message,omitempty"`
}

func (r *ScheduledJobRun) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}

// ActivityLog is an audit trail entry.
type ActivityLog struct {
	ID           uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	ProjectID    uuid.UUID      `gorm:"type:uuid;not null;index" json:"project_id"`
	UserID       *uuid.UUID     `gorm:"type:uuid" json:"user_id,omitempty"`
	Action       string         `gorm:"size:100;not null" json:"action"`
	ResourceType string         `gorm:"size:50;not null" json:"resource_type"`
	ResourceID   *uuid.UUID     `gorm:"type:uuid" json:"resource_id,omitempty"`
	Details      datatypes.JSON `gorm:"type:jsonb;default:'{}'" json:"details"`
	CreatedAt    time.Time      `json:"created_at"`
}

func (a *ActivityLog) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}
