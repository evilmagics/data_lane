package domain

import (
	"time"

	"gorm.io/gorm"
)

// Gate represents a toll gate entry point
type Gate struct {
	ID        int       `gorm:"primaryKey;type:integer" json:"id"`
	Name      string    `gorm:"type:text;not null" json:"name"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (g *Gate) BeforeCreate(tx *gorm.DB) error {
	return nil
}
