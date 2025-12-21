package ports

import (
	"context"

	"pdf_generator/internal/core/domain"
)

// TaskProgress holds the current progress state for a task
type TaskProgress struct {
	Stage   string `json:"stage"`
	Current int    `json:"current"`
	Total   int    `json:"total"`
}

// QueueService defines the interface for task queue operations
type QueueService interface {
	Enqueue(ctx context.Context, taskID string, metadata domain.TaskMetadata) ([]string, error)
	Start(ctx context.Context)
	GetProgress(taskID string) *TaskProgress
}

