package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID               uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	Email            string    `gorm:"uniqueIndex;size:255;not null" json:"email"`
	PasswordHash     string    `gorm:"size:255;not null" json:"-"`
	DisplayName      string    `gorm:"size:100;not null" json:"display_name"`
	EmailVerified    bool      `gorm:"default:false" json:"email_verified"`
	TwoFactorEnabled bool      `gorm:"default:false" json:"two_factor_enabled"`
	TwoFactorSecret  string    `gorm:"size:255" json:"-"`
	Role             string    `gorm:"size:20;default:user" json:"role"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

type UserResponse struct {
	ID            uuid.UUID `json:"id"`
	Email         string    `json:"email"`
	DisplayName   string    `json:"display_name"`
	EmailVerified bool      `json:"email_verified"`
	Role          string    `json:"role"`
	CreatedAt     time.Time `json:"created_at"`
}

func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:            u.ID,
		Email:         u.Email,
		DisplayName:   u.DisplayName,
		EmailVerified: u.EmailVerified,
		Role:          u.Role,
		CreatedAt:     u.CreatedAt,
	}
}
