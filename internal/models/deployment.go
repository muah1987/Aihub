package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ProjectEnvVar stores encrypted env vars/secrets for a project.
type ProjectEnvVar struct {
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	ProjectID uuid.UUID `gorm:"type:uuid;not null;index" json:"project_id"`
	Key       string    `gorm:"size:255;not null" json:"key"`
	Value     string    `gorm:"type:text;not null" json:"-"` // AES-encrypted, never serialized to clients
	IsSecret  bool      `gorm:"default:false" json:"is_secret"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (e *ProjectEnvVar) BeforeCreate(tx *gorm.DB) error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	return nil
}

// VPSTarget is a configured SSH deployment target for a project.
type VPSTarget struct {
	ID             uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	ProjectID      uuid.UUID  `gorm:"type:uuid;not null;index" json:"project_id"`
	Name           string     `gorm:"size:255;not null" json:"name"`
	Host           string     `gorm:"size:255;not null" json:"host"`
	Port           int        `gorm:"default:22" json:"port"`
	Username       string     `gorm:"size:255;not null" json:"username"`
	AuthType       string     `gorm:"size:20;default:key" json:"auth_type"` // "key" or "password"
	SSHKey         string     `gorm:"type:text" json:"-"`                   // AES-encrypted PEM private key
	SSHPassword    string     `gorm:"type:text" json:"-"`                   // AES-encrypted password
	DeployPath     string     `gorm:"size:500;not null" json:"deploy_path"`
	PreDeployCmd   string     `gorm:"type:text" json:"pre_deploy_cmd"`
	DeployCmd      string     `gorm:"type:text;not null" json:"deploy_cmd"`
	PostDeployCmd  string     `gorm:"type:text" json:"post_deploy_cmd"`
	Status         string     `gorm:"size:20;default:idle" json:"status"` // idle | deploying | success | failed
	LastDeployedAt *time.Time `json:"last_deployed_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

func (v *VPSTarget) BeforeCreate(tx *gorm.DB) error {
	if v.ID == uuid.Nil {
		v.ID = uuid.New()
	}
	return nil
}

// DeploymentRun records a single deployment execution.
type DeploymentRun struct {
	ID          uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	VPSTargetID uuid.UUID  `gorm:"type:uuid;not null;index" json:"vps_target_id"`
	ProjectID   uuid.UUID  `gorm:"type:uuid;not null;index" json:"project_id"`
	Status      string     `gorm:"size:20;default:running" json:"status"` // running | success | failed
	TriggeredBy *uuid.UUID `gorm:"type:uuid" json:"triggered_by,omitempty"`
	LogOutput   string     `gorm:"type:text;default:''" json:"log_output"`
	StartedAt   time.Time  `gorm:"default:NOW()" json:"started_at"`
	FinishedAt  *time.Time `json:"finished_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`

	VPSTarget *VPSTarget `gorm:"foreignKey:VPSTargetID" json:"target,omitempty"`
}

func (d *DeploymentRun) BeforeCreate(tx *gorm.DB) error {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	return nil
}
