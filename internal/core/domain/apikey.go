package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// APIKey represents an external access key
type APIKey struct {
	ID           string    `gorm:"primaryKey;type:text" json:"id"`
	Name         string    `gorm:"type:text;not null" json:"name"`
	KeyHash      string    `gorm:"type:text;uniqueIndex" json:"-"`         // Bcrypt hash for verification
	EncryptedKey string    `gorm:"type:text" json:"-"`                     // AES encrypted for reveal
	Active       bool      `gorm:"default:true" json:"active"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (k *APIKey) BeforeCreate(tx *gorm.DB) error {
	if k.ID == "" {
		k.ID = uuid.New().String()
	}
	return nil
}
