package ports

import (
	"context"

	"pdf_generator/internal/core/domain"
)

// QueueService defines the interface for task queue operations
type QueueService interface {
	Enqueue(ctx context.Context, taskID string, metadata domain.TaskMetadata) ([]string, error)
	Start(ctx context.Context)
}
