package scheduler

import (
	"context"
	"encoding/json"
	"os"
	"strconv"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/rs/zerolog/log"

	"pdf_generator/internal/core/domain"
	"pdf_generator/internal/core/ports"
)

// Scheduler manages cron jobs
type Scheduler struct {
	cron         gocron.Scheduler
	scheduleRepo ports.ScheduleRepository
	taskRepo     ports.TaskRepository
	settingsRepo ports.SettingsRepository
}

// NewScheduler creates a new scheduler
func NewScheduler(
	scheduleRepo ports.ScheduleRepository,
	taskRepo ports.TaskRepository,
	settingsRepo ports.SettingsRepository,
) (*Scheduler, error) {
	cron, err := gocron.NewScheduler()
	if err != nil {
		return nil, err
	}

	return &Scheduler{
		cron:         cron,
		scheduleRepo: scheduleRepo,
		taskRepo:     taskRepo,
		settingsRepo: settingsRepo,
	}, nil
}

// Start loads schedules from DB and starts the scheduler
func (s *Scheduler) Start(ctx context.Context) error {
	// Load active schedules
	schedules, err := s.scheduleRepo.ListActive(ctx)
	if err != nil {
		return err
	}

	for _, schedule := range schedules {
		if err := s.AddJob(ctx, schedule); err != nil {
			log.Error().Err(err).Str("schedule_id", schedule.ID).Msg("Failed to add schedule job")
		}
	}

	// Add cleanup job (runs at midnight)
	s.addCleanupJob(ctx)

	s.cron.Start()
	log.Info().Int("count", len(schedules)).Msg("Scheduler started")
	return nil
}

// AddJob adds a cron job for a schedule
func (s *Scheduler) AddJob(ctx context.Context, schedule domain.Schedule) error {
	_, err := s.cron.NewJob(
		gocron.CronJob(schedule.Cron, false),
		gocron.NewTask(func() {
			s.executeSchedule(context.Background(), schedule)
		}),
		gocron.WithTags(schedule.ID),
	)
	return err
}

// RemoveJob removes a cron job
func (s *Scheduler) RemoveJob(scheduleID string) error {
	s.cron.RemoveByTags(scheduleID)
	return nil
}

// Stop stops the scheduler
func (s *Scheduler) Stop() error {
	return s.cron.Shutdown()
}

func (s *Scheduler) executeSchedule(ctx context.Context, schedule domain.Schedule) {
	log.Info().Str("schedule_id", schedule.ID).Msg("Executing scheduled task")

	// Parse task metadata
	var metadata domain.TaskMetadata
	if err := json.Unmarshal([]byte(schedule.TaskPayload), &metadata); err != nil {
		log.Error().Err(err).Str("schedule_id", schedule.ID).Msg("Failed to parse task payload")
		return
	}

	// Create new task with extracted fields
	filterJSON, _ := json.Marshal(metadata.Filter)
	settingsJSON, _ := json.Marshal(metadata.Settings)
	task := &domain.Task{
		ScheduleID:   &schedule.ID,
		Status:       domain.TaskStatusQueued,
		RootFolder:   metadata.RootFolder,
		GateID:       metadata.GateID,
		StationID:    metadata.StationID,
		FilterJSON:   string(filterJSON),
		SettingsJSON: string(settingsJSON),
	}

	if err := s.taskRepo.Create(ctx, task); err != nil {
		log.Error().Err(err).Str("schedule_id", schedule.ID).Msg("Failed to create task from schedule")
		return
	}

	// Update schedule last run
	now := time.Now()
	schedule.LastRun = &now
	s.scheduleRepo.Update(ctx, &schedule)

	log.Info().Str("schedule_id", schedule.ID).Str("task_id", task.ID).Msg("Task created from schedule")
}

func (s *Scheduler) addCleanupJob(ctx context.Context) {
	_, err := s.cron.NewJob(
		gocron.CronJob("0 0 * * *", false), // Midnight daily
		gocron.NewTask(func() {
			s.cleanupExpiredFiles(context.Background())
		}),
		gocron.WithTags("cleanup"),
	)
	if err != nil {
		log.Error().Err(err).Msg("Failed to add cleanup job")
	}
}

func (s *Scheduler) cleanupExpiredFiles(ctx context.Context) {
	log.Info().Msg("Running file cleanup job")

	// Get max age setting
	maxAgeDays := 7
	setting, err := s.settingsRepo.Get(ctx, domain.SettingMaxOutputAgeDays)
	if err == nil && setting != nil {
		if val, err := strconv.Atoi(setting.Value); err == nil && val > 0 {
			maxAgeDays = val
		}
	}

	// Find expired tasks
	tasks, err := s.taskRepo.FindExpiredCompleted(ctx, maxAgeDays)
	if err != nil {
		log.Error().Err(err).Msg("Failed to find expired tasks")
		return
	}

	for _, task := range tasks {
		// Delete file
		if task.OutputFilePath != "" {
			if err := os.Remove(task.OutputFilePath); err != nil {
				log.Warn().Err(err).Str("task_id", task.ID).Msg("Failed to delete output file")
			}
		}

		// Update status
		task.Status = domain.TaskStatusRemoved
		task.OutputFilePath = ""
		task.OutputFileSize = 0
		s.taskRepo.Update(ctx, &task)

		log.Info().Str("task_id", task.ID).Msg("Task output removed due to age")
	}

	log.Info().Int("count", len(tasks)).Msg("Cleanup completed")
}
