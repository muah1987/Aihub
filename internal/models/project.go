package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Project struct {
	ID                   uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	UserID               uuid.UUID      `gorm:"type:uuid;not null;index" json:"user_id"`
	Name                 string         `gorm:"size:255;not null" json:"name"`
	Description          string         `gorm:"type:text" json:"description"`
	RepoProvider         string         `gorm:"size:50;not null" json:"repo_provider"`
	RepoOwner            string         `gorm:"size:255;not null" json:"repo_owner"`
	RepoName             string         `gorm:"size:255;not null" json:"repo_name"`
	RepoURL              string         `gorm:"size:500;not null" json:"repo_url"`
	RepoDefaultBranch    string         `gorm:"size:100;default:main" json:"repo_default_branch"`
	ProviderConnectionID *uuid.UUID     `gorm:"type:uuid" json:"provider_connection_id,omitempty"`
	Status               string         `gorm:"size:20;default:active" json:"status"`
	Settings             datatypes.JSON `gorm:"type:jsonb;default:'{}'" json:"settings"`
	CreatedAt            time.Time      `json:"created_at"`
	UpdatedAt            time.Time      `json:"updated_at"`

	User               User                `gorm:"foreignKey:UserID" json:"-"`
	ProviderConnection *ProviderConnection `gorm:"foreignKey:ProviderConnectionID" json:"-"`
}

func (p *Project) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}
