package queue

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/mikestefanello/backlite"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"pdf_generator/internal/core/domain"
	"pdf_generator/internal/core/ports"
	"pdf_generator/pkg/generator"
)

// PDFTask represents the task data for PDF generation
type PDFTask struct {
	TaskID   string              `json:"task_id"`
	Metadata domain.TaskMetadata `json:"metadata"`
}

// Config returns the backlite task configuration
func (t PDFTask) Config() backlite.QueueConfig {
	return backlite.QueueConfig{
		Name:        "generate_pdf",
		MaxAttempts: 3,
		Backoff:     5 * time.Second,
		Timeout:     10 * time.Minute,
	}
}

// Queue wraps the backlite queue
type Queue struct {
	client       *backlite.Client
	taskRepo     ports.TaskRepository
	settingsRepo ports.SettingsRepository
}

// NewQueue creates a new queue instance
func NewQueue(db *gorm.DB, taskRepo ports.TaskRepository, settingsRepo ports.SettingsRepository) (*Queue, error) {
	// Get concurrency from settings
	concurrency := 1
	setting, err := settingsRepo.Get(context.Background(), domain.SettingQueueConcurrency)
	if err == nil && setting != nil {
		if val, _ := strconv.Atoi(setting.Value); val > 0 {
			concurrency = val
		}
	}

	// Get generic database object for backlite
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	client, err := backlite.NewClient(backlite.ClientConfig{
		DB:              sqlDB,
		Logger:          nil, // Use default or wrap access logger?
		NumWorkers:      concurrency,
		ReleaseAfter:    30 * time.Minute,
		CleanupInterval: 1 * time.Hour,
	})
	if err != nil {
		return nil, err
	}

	// Ensure queue tables exist
	if err := client.Install(); err != nil {
		return nil, err
	}

	q := &Queue{
		client:       client,
		taskRepo:     taskRepo,
		settingsRepo: settingsRepo,
	}

	return q, nil
}

// RegisterConsumers registers the task handlers for the queue consumers
func (q *Queue) RegisterConsumers() {
	// Register task handler
	pdfQueue := backlite.NewQueue(q.handlePDFTask)
	q.client.Register(pdfQueue)
}

// Enqueue adds a task to the queue
func (q *Queue) Enqueue(ctx context.Context, taskID string, metadata domain.TaskMetadata) ([]string, error) {
	task := PDFTask{
		TaskID:   taskID,
		Metadata: metadata,
	}
	return q.client.Add(task).Save()
}

// Start starts the queue workers
func (q *Queue) Start(ctx context.Context) {
	q.client.Start(ctx)
}

// handlePDFTask processes a PDF generation task
func (q *Queue) handlePDFTask(ctx context.Context, task PDFTask) error {
	log.Info().Str("task_id", task.TaskID).Msg("Processing PDF task")

	// Update task status to running
	dbTask, err := q.taskRepo.GetByID(ctx, task.TaskID)
	if err != nil {
		log.Error().Err(err).Str("task_id", task.TaskID).Msg("Task not found")
		return err
	}

	dbTask.Status = domain.TaskStatusRunning
	if err := q.taskRepo.Update(ctx, dbTask); err != nil {
		return err
	}

	// Generate PDF
	output, size, err := generator.GeneratePDF(ctx, task.Metadata, q.settingsRepo)
	if err != nil {
		log.Error().Err(err).Str("task_id", task.TaskID).Msg("PDF generation failed")
		dbTask.Status = domain.TaskStatusFailed
		q.taskRepo.Update(ctx, dbTask)
		return err
	}

	// Update task with output
	dbTask.Status = domain.TaskStatusCompleted
	dbTask.OutputFilePath = output
	dbTask.OutputFileSize = size
	if err := q.taskRepo.Update(ctx, dbTask); err != nil {
		return err
	}

	log.Info().Str("task_id", task.TaskID).Str("output", output).Msg("PDF generated successfully")
	return nil
}

// ParseTaskMetadata parses JSON metadata string
func ParseTaskMetadata(metadataJSON string) (domain.TaskMetadata, error) {
	var metadata domain.TaskMetadata
	err := json.Unmarshal([]byte(metadataJSON), &metadata)
	return metadata, err
}
