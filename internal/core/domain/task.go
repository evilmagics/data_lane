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
	ID           string     `gorm:"primaryKey;type:text" json:"id"`
	ScheduleID   *string    `gorm:"type:text;index" json:"schedule_id,omitempty"`
	Status       TaskStatus `gorm:"type:text;index;not null;default:'queued'" json:"status"`
	ErrorMessage string     `gorm:"type:text" json:"error_message,omitempty"`

	// Extracted metadata fields
	RootFolder   string `gorm:"type:text" json:"root_folder"`
	GateID       int    `gorm:"type:integer" json:"gate_id"`
	FilterJSON   string `gorm:"type:text" json:"filter_json"`   // JSON string of TaskFilter
	SettingsJSON string `gorm:"type:text" json:"settings_json"` // JSON string of custom settings

	// Progress tracking
	ProgressStage   string `gorm:"type:text" json:"progress_stage,omitempty"`      // Detailed stage description
	ProgressTotal   int    `gorm:"type:integer;default:0" json:"progress_total"`   // Total transactions to process
	ProgressCurrent int    `gorm:"type:integer;default:0" json:"progress_current"` // Current processed count

	OutputFilePath string    `gorm:"type:text" json:"output_file_path,omitempty"`
	OutputFileSize int64     `gorm:"type:integer;default:0" json:"output_file_size"`
	CreatedAt      time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (t *Task) BeforeCreate(tx *gorm.DB) error {
	if t.ID == "" {
		t.ID = uuid.New().String()
	}
	return nil
}

// TaskMetadata contains the parameters for PDF generation (used in queue, not stored directly)
type TaskMetadata struct {
	RootFolder string            `json:"root_folder"`
	BranchID   int               `json:"branch_id"` // Fetched from settings
	GateID     int               `json:"gate_id"`
	Filter     TaskFilter        `json:"filter"`
	Settings   map[string]string `json:"settings,omitempty"`
}

// TaskFilter contains date and transaction filtering options
// Date mode is auto-detected: if Date is set, use daily mode; if RangeStart/RangeEnd set, use range mode
type TaskFilter struct {
	Date              string `json:"date,omitempty"`                // Single date (YYYY-MM-DD)
	DayStartTime      string `json:"day_start_time,omitempty"`      // Daily window start time (HH:MM), from settings
	RangeStart        string `json:"range_start,omitempty"`         // Range start datetime
	RangeEnd          string `json:"range_end,omitempty"`           // Range end datetime
	TransactionStatus string `json:"transaction_status,omitempty"`  // Filter by status
	Limit             int    `json:"limit,omitempty"`               // Max transactions to fetch, 0 = unlimited
}
