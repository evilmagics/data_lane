package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Schedule represents a recurring task definition
type Schedule struct {
	ID          string     `gorm:"primaryKey;type:text" json:"id"`
	Cron        string     `gorm:"type:text;not null" json:"cron"`
	TaskPayload string     `gorm:"type:text;not null" json:"task_payload"` // JSON of TaskMetadata
	Active      bool       `gorm:"default:true" json:"active"`
	LastRun     *time.Time `gorm:"type:datetime" json:"last_run,omitempty"`
	NextRun     *time.Time `gorm:"type:datetime" json:"next_run,omitempty"`
	CreatedAt   time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}

func (s *Schedule) BeforeCreate(tx *gorm.DB) error {
	if s.ID == "" {
		s.ID = uuid.New().String()
	}
	return nil
}
