package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TwoFactorBackupCode struct {
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	CodeHash  string    `gorm:"size:255;not null" json:"-"`
	Used      bool      `gorm:"default:false" json:"used"`
	CreatedAt time.Time `json:"created_at"`

	User User `gorm:"foreignKey:UserID" json:"-"`
}

func (b *TwoFactorBackupCode) BeforeCreate(tx *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}
