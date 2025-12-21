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
	ID              string     `gorm:"primaryKey;type:text" json:"id"`
	ScheduleID      *string    `gorm:"type:text;index" json:"schedule_id,omitempty"`
	Status          TaskStatus `gorm:"type:text;index;not null;default:'queued'" json:"status"`
	ErrorMessage    string     `gorm:"type:text" json:"error_message,omitempty"`
	
	// Extracted metadata fields
	RootFolder      string     `gorm:"type:text" json:"root_folder"`
	StationID       int        `gorm:"type:integer" json:"station_id"`
	FilterJSON      string     `gorm:"type:text" json:"filter_json"` // JSON string of TaskFilter
	
	// Progress tracking
	ProgressStage   string     `gorm:"type:text" json:"progress_stage,omitempty"`   // e.g., "loading", "generating", "saving"
	ProgressTotal   int        `gorm:"type:integer;default:0" json:"progress_total"`   // Total transactions to process
	ProgressCurrent int        `gorm:"type:integer;default:0" json:"progress_current"` // Current processed count
	
	// Legacy metadata field (for backward compatibility during migration)
	Metadata        string     `gorm:"type:text" json:"metadata,omitempty"` // JSON string (deprecated)
	
	OutputFilePath  string     `gorm:"type:text" json:"output_file_path,omitempty"`
	OutputFileSize  int64      `gorm:"type:integer;default:0" json:"output_file_size"`
	CreatedAt       time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}

func (t *Task) BeforeCreate(tx *gorm.DB) error {
	if t.ID == "" {
		t.ID = uuid.New().String()
	}
	return nil
}

// TaskMetadata contains the parameters for PDF generation (used in queue, not stored directly)
type TaskMetadata struct {
	RootFolder   string                  `json:"root_folder"`
	BranchID     int                     `json:"branch_id"`   // Fetched from settings
	GateID       int                     `json:"gate_id"`     // Fetched from settings
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

