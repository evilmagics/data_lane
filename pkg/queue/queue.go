package queue

import (
	"context"
	"encoding/json"
	"strconv"
	"sync"
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
	
	// Progress tracking for SSE
	progressMu   sync.RWMutex
	progress     map[string]*ports.TaskProgress
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
		Logger:          &BackliteLogger{},
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
		progress:     make(map[string]*ports.TaskProgress),
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

	// Periodically notify to ensure frequent polling (every 1s)
	// This forces the workers to check for new tasks even if idle
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				q.client.Notify()
			}
		}
	}()
}

// GetProgress returns the current progress for a task
func (q *Queue) GetProgress(taskID string) *ports.TaskProgress {
	q.progressMu.RLock()
	defer q.progressMu.RUnlock()
	return q.progress[taskID]
}

// setProgress updates the progress for a task
func (q *Queue) setProgress(taskID, stage string, current, total int) {
	q.progressMu.Lock()
	q.progress[taskID] = &ports.TaskProgress{
		Stage:   stage,
		Current: current,
		Total:   total,
	}
	q.progressMu.Unlock()
}

// clearProgress removes progress tracking for a task
func (q *Queue) clearProgress(taskID string) {
	q.progressMu.Lock()
	delete(q.progress, taskID)
	q.progressMu.Unlock()
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

	// Initialize progress tracking
	q.setProgress(task.TaskID, "initializing", 0, 0)

	// Create a throttled progress callback that updates DB every 1 second
	var lastUpdate time.Time
	progressCallback := func(stage string, current, total int) {
		// Always update in-memory progress
		q.setProgress(task.TaskID, stage, current, total)
		
		// Throttle DB updates to once per second
		now := time.Now()
		if now.Sub(lastUpdate) >= time.Second {
			lastUpdate = now
			if err := q.taskRepo.UpdateProgress(ctx, task.TaskID, stage, current, total); err != nil {
				log.Warn().Err(err).Str("task_id", task.TaskID).Msg("Failed to update progress in DB")
			}
		}
	}

	// Generate PDF with progress tracking
	output, size, err := generator.GeneratePDFWithProgress(ctx, task.Metadata, q.settingsRepo, progressCallback)
	if err != nil {
		log.Error().Err(err).Str("task_id", task.TaskID).Msg("PDF generation failed")
		
		// Update error message in database
		if updateErr := q.taskRepo.UpdateError(ctx, task.TaskID, err.Error()); updateErr != nil {
			log.Error().Err(updateErr).Str("task_id", task.TaskID).Msg("Failed to update error in DB")
		}
		
		q.clearProgress(task.TaskID)
		return err
	}

	// Update task with output
	dbTask.Status = domain.TaskStatusCompleted
	dbTask.OutputFilePath = output
	dbTask.OutputFileSize = size
	dbTask.ProgressStage = "completed"
	if err := q.taskRepo.Update(ctx, dbTask); err != nil {
		q.clearProgress(task.TaskID)
		return err
	}

	q.clearProgress(task.TaskID)
	log.Info().Str("task_id", task.TaskID).Str("output", output).Msg("PDF generated successfully")
	return nil
}

// ParseTaskMetadata parses JSON metadata string
func ParseTaskMetadata(metadataJSON string) (domain.TaskMetadata, error) {
	var metadata domain.TaskMetadata
	err := json.Unmarshal([]byte(metadataJSON), &metadata)
	return metadata, err
}
