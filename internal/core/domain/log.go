package domain

import (
	"time"
)

// Log represents an application log entry stored in the database
type Log struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Level     string    `gorm:"type:text;index" json:"level"` // info, warn, error
	Message   string    `gorm:"type:text" json:"message"`
	Context   string    `gorm:"type:text" json:"context,omitempty"` // JSON context data
	CreatedAt time.Time `gorm:"autoCreateTime;index" json:"created_at"`
}
