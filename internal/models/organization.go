package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Organization struct {
	ID          uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	Name        string         `gorm:"size:255;not null" json:"name"`
	Slug        string         `gorm:"size:255;uniqueIndex;not null" json:"slug"`
	Description string         `gorm:"type:text" json:"description"`
	OwnerID     uuid.UUID      `gorm:"type:uuid;not null;index" json:"owner_id"`
	Settings    datatypes.JSON `gorm:"type:jsonb;default:'{}'" json:"settings"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`

	Owner User `gorm:"foreignKey:OwnerID" json:"-"`
}

func (o *Organization) BeforeCreate(tx *gorm.DB) error {
	if o.ID == uuid.Nil {
		o.ID = uuid.New()
	}
	return nil
}

type OrganizationMember struct {
	ID             uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	OrganizationID uuid.UUID  `gorm:"type:uuid;not null;index" json:"organization_id"`
	UserID         uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	Role           string     `gorm:"size:20;not null;default:member" json:"role"`
	InvitedBy      *uuid.UUID `gorm:"type:uuid" json:"invited_by,omitempty"`
	JoinedAt       time.Time  `gorm:"default:now()" json:"joined_at"`

	Organization Organization `gorm:"foreignKey:OrganizationID" json:"-"`
	User         User         `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (m *OrganizationMember) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}

type OrganizationInvitation struct {
	ID             uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	OrganizationID uuid.UUID `gorm:"type:uuid;not null;index" json:"organization_id"`
	Email          string    `gorm:"size:255;not null" json:"email"`
	Role           string    `gorm:"size:20;not null;default:member" json:"role"`
	Token          string    `gorm:"size:255;uniqueIndex;not null" json:"-"`
	InvitedBy      uuid.UUID `gorm:"type:uuid;not null" json:"invited_by"`
	ExpiresAt      time.Time `gorm:"not null" json:"expires_at"`
	Accepted       bool      `gorm:"default:false" json:"accepted"`
	CreatedAt      time.Time `json:"created_at"`

	Organization Organization `gorm:"foreignKey:OrganizationID" json:"-"`
}

func (i *OrganizationInvitation) BeforeCreate(tx *gorm.DB) error {
	if i.ID == uuid.Nil {
		i.ID = uuid.New()
	}
	return nil
}
