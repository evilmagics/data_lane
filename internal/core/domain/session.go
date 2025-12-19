package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Session represents an active admin login session
type Session struct {
	ID        string    `gorm:"primaryKey;type:text" json:"id"`
	UserID    string    `gorm:"type:text;not null;index" json:"user_id"`
	TokenHash string    `gorm:"type:text;uniqueIndex" json:"-"` // SHA256 of JWT for revocation
	IPAddress string    `gorm:"type:text" json:"ip_address"`
	UserAgent string    `gorm:"type:text" json:"user_agent"`
	Status    string    `gorm:"type:text;default:'active';index" json:"status"` // active, revoked, expired
	ExpiresAt time.Time `gorm:"type:datetime;index" json:"expires_at"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (s *Session) BeforeCreate(tx *gorm.DB) error {
	if s.ID == "" {
		s.ID = uuid.New().String()
	}
	if s.Status == "" {
		s.Status = "active"
	}
	return nil
}

// IsExpired checks if the session has expired
func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}
