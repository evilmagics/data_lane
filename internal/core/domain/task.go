package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TaskStatus represents the state of a task
type TaskStatus string

const (
	TaskStatusQueued    TaskStatus = "queued"
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusFailed    TaskStatus = "failed"
	TaskStatusCancelled TaskStatus = "cancelled"
	TaskStatusRemoved   TaskStatus = "removed"
)

// Task represents a PDF generation job
type Task struct {
	ID             string     `gorm:"primaryKey;type:text" json:"id"`
	ScheduleID     *string    `gorm:"type:text;index" json:"schedule_id,omitempty"`
	Status         TaskStatus `gorm:"type:text;index;not null;default:'queued'" json:"status"`
	Metadata       string     `gorm:"type:text" json:"metadata"` // JSON string
	OutputFilePath string     `gorm:"type:text" json:"output_file_path,omitempty"`
	OutputFileSize int64      `gorm:"type:integer;default:0" json:"output_file_size"`
	CreatedAt      time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}

func (t *Task) BeforeCreate(tx *gorm.DB) error {
	if t.ID == "" {
		t.ID = uuid.New().String()
	}
	return nil
}

// TaskMetadata contains the parameters for PDF generation
type TaskMetadata struct {
	RootFolder   string                  `json:"root_folder"`
	BranchID     int                     `json:"branch_id"`
	GateID       int                     `json:"gate_id"`
	StationID    int                     `json:"station_id"`
	Filter       TaskFilter              `json:"filter"`
	Settings     map[string]string       `json:"settings,omitempty"`
}

// TaskFilter contains date and transaction filtering options
type TaskFilter struct {
	DateMode          string `json:"date_mode"`          // daily, range, yesterday
	Date              string `json:"date,omitempty"`
	RangeStart        string `json:"range_start,omitempty"`
	RangeEnd          string `json:"range_end,omitempty"`
	TransactionStatus string `json:"transaction_status,omitempty"`
}
