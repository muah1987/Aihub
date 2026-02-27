package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type ProviderConnection struct {
	ID             uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	UserID         uuid.UUID      `gorm:"type:uuid;not null;index" json:"user_id"`
	ProviderType   string         `gorm:"size:50;not null" json:"provider_type"`
	ProviderName   string         `gorm:"size:100;not null" json:"provider_name"`
	AccessToken    string         `gorm:"type:text;not null" json:"-"`
	RefreshToken   string         `gorm:"type:text" json:"-"`
	TokenExpiresAt *time.Time     `json:"token_expires_at,omitempty"`
	Scopes         string         `gorm:"type:text" json:"scopes"`
	Status         string         `gorm:"size:20;default:active" json:"status"`
	Metadata       datatypes.JSON `gorm:"type:jsonb;default:'{}'" json:"metadata"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`

	User User `gorm:"foreignKey:UserID" json:"-"`
}

func (p *ProviderConnection) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

type ProviderConnectionResponse struct {
	ID           uuid.UUID  `json:"id"`
	ProviderType string     `json:"provider_type"`
	ProviderName string     `json:"provider_name"`
	Scopes       string     `json:"scopes"`
	Status       string     `json:"status"`
	TokenExpires *time.Time `json:"token_expires_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}

func (p *ProviderConnection) ToResponse() ProviderConnectionResponse {
	return ProviderConnectionResponse{
		ID:           p.ID,
		ProviderType: p.ProviderType,
		ProviderName: p.ProviderName,
		Scopes:       p.Scopes,
		Status:       p.Status,
		TokenExpires: p.TokenExpiresAt,
		CreatedAt:    p.CreatedAt,
	}
}
