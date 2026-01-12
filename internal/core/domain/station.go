package domain

import (
	"time"

	"gorm.io/gorm"
)

// Station represents a toll station
type Station struct {
	ID        int       `gorm:"primaryKey;type:integer" json:"id"`
	Name      string    `gorm:"type:text;not null" json:"name"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (s *Station) BeforeCreate(tx *gorm.DB) error {
	return nil
}
