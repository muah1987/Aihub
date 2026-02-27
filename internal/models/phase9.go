package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// AuditLog records a user action for compliance tracking.
type AuditLog struct {
	ID             uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	OrganizationID *uuid.UUID     `gorm:"type:uuid" json:"organization_id,omitempty"`
	ProjectID      *uuid.UUID     `gorm:"type:uuid" json:"project_id,omitempty"`
	UserID         *uuid.UUID     `gorm:"type:uuid" json:"user_id,omitempty"`
	Action         string         `gorm:"size:100;not null" json:"action"`
	ResourceType   string         `gorm:"size:50;not null" json:"resource_type"`
	ResourceID     *uuid.UUID     `gorm:"type:uuid" json:"resource_id,omitempty"`
	Details        datatypes.JSON `gorm:"type:jsonb;default:'{}'" json:"details"`
	IPAddress      string         `gorm:"size:45" json:"ip_address,omitempty"`
	UserAgent      string         `gorm:"type:text" json:"user_agent,omitempty"`
	Severity       string         `gorm:"size:20;default:info" json:"severity"`
	CreatedAt      time.Time      `json:"created_at"`
}

func (a *AuditLog) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}

// DataExport represents a project backup job.
type DataExport struct {
	ID               uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	ProjectID        uuid.UUID  `gorm:"type:uuid;not null;index" json:"project_id"`
	RequestedBy      *uuid.UUID `gorm:"type:uuid" json:"requested_by,omitempty"`
	ExportType       string     `gorm:"size:30;not null;default:full" json:"export_type"`
	Format           string     `gorm:"size:20;not null;default:json" json:"format"`
	Status           string     `gorm:"size:20;not null;default:pending" json:"status"`
	FilePath         string     `gorm:"type:text" json:"file_path,omitempty"`
	FileSize         int64      `gorm:"default:0" json:"file_size"`
	IncludeChat      bool       `gorm:"default:true" json:"include_chat"`
	IncludeMemories  bool       `gorm:"default:true" json:"include_memories"`
	IncludeAgents    bool       `gorm:"default:true" json:"include_agents"`
	IncludeWorkflows bool       `gorm:"default:true" json:"include_workflows"`
	IncludeSettings  bool       `gorm:"default:true" json:"include_settings"`
	ErrorMessage     string     `gorm:"type:text" json:"error_message,omitempty"`
	ExpiresAt        *time.Time `json:"expires_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	CompletedAt      *time.Time `json:"completed_at,omitempty"`
}

func (d *DataExport) BeforeCreate(tx *gorm.DB) error {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	return nil
}

// CustomRole is a user-defined role within an organization.
type CustomRole struct {
	ID             uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	OrganizationID uuid.UUID      `gorm:"type:uuid;not null;index" json:"organization_id"`
	Name           string         `gorm:"size:100;not null" json:"name"`
	Description    string         `gorm:"type:text;default:''" json:"description"`
	Permissions    datatypes.JSON `gorm:"type:jsonb;not null;default:'[]'" json:"permissions"`
	IsSystem       bool           `gorm:"default:false" json:"is_system"`
	CreatedBy      *uuid.UUID     `gorm:"type:uuid" json:"created_by,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
}

func (c *CustomRole) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

// ResourcePermission grants a specific permission on a resource to a user.
type ResourcePermission struct {
	ID             uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	OrganizationID uuid.UUID  `gorm:"type:uuid;not null" json:"organization_id"`
	UserID         uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	ResourceType   string     `gorm:"size:50;not null" json:"resource_type"`
	ResourceID     uuid.UUID  `gorm:"type:uuid;not null" json:"resource_id"`
	Permission     string     `gorm:"size:50;not null" json:"permission"`
	GrantedBy      *uuid.UUID `gorm:"type:uuid" json:"granted_by,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
}

func (r *ResourcePermission) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}

// RetentionPolicy defines automatic data cleanup rules.
type RetentionPolicy struct {
	ID             uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	OrganizationID *uuid.UUID `gorm:"type:uuid" json:"organization_id,omitempty"`
	ProjectID      *uuid.UUID `gorm:"type:uuid" json:"project_id,omitempty"`
	ResourceType   string     `gorm:"size:50;not null" json:"resource_type"`
	RetentionDays  int        `gorm:"not null;default:90" json:"retention_days"`
	Enabled        bool       `gorm:"default:true" json:"enabled"`
	LastRunAt      *time.Time `json:"last_run_at,omitempty"`
	RecordsDeleted int        `gorm:"default:0" json:"records_deleted"`
	CreatedBy      *uuid.UUID `gorm:"type:uuid" json:"created_by,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

func (r *RetentionPolicy) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}
