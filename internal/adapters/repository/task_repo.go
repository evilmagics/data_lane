package repository

import (
	"context"
	"time"

	"gorm.io/gorm"

	"pdf_generator/internal/core/domain"
	"pdf_generator/internal/core/ports"
)

type taskRepository struct {
	db *gorm.DB
}

// NewTaskRepository creates a new task repository
func NewTaskRepository(db *gorm.DB) ports.TaskRepository {
	return &taskRepository{db: db}
}

func (r *taskRepository) Create(ctx context.Context, task *domain.Task) error {
	return r.db.WithContext(ctx).Create(task).Error
}

func (r *taskRepository) GetByID(ctx context.Context, id string) (*domain.Task, error) {
	var task domain.Task
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&task).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func (r *taskRepository) Update(ctx context.Context, task *domain.Task) error {
	return r.db.WithContext(ctx).Save(task).Error
}

func (r *taskRepository) UpdateProgress(ctx context.Context, id string, stage string, current, total int) error {
	return r.db.WithContext(ctx).Model(&domain.Task{}).Where("id = ?", id).Updates(map[string]interface{}{
		"progress_stage":   stage,
		"progress_current": current,
		"progress_total":   total,
	}).Error
}

func (r *taskRepository) UpdateError(ctx context.Context, id string, errMsg string) error {
	return r.db.WithContext(ctx).Model(&domain.Task{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":        domain.TaskStatusFailed,
		"error_message": errMsg,
	}).Error
}

func (r *taskRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&domain.Task{}, "id = ?", id).Error
}

func (r *taskRepository) List(ctx context.Context, filter ports.TaskFilter) ([]domain.Task, int64, error) {
	var tasks []domain.Task
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.Task{})

	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.FromDate != "" {
		query = query.Where("created_at >= ?", filter.FromDate)
	}
	if filter.ToDate != "" {
		query = query.Where("created_at <= ?", filter.ToDate)
	}

	query.Count(&total)

	if filter.Limit <= 0 {
		filter.Limit = 10
	}
	if filter.Page <= 0 {
		filter.Page = 1
	}
	offset := (filter.Page - 1) * filter.Limit

	err := query.Order("created_at DESC").Offset(offset).Limit(filter.Limit).Find(&tasks).Error
	return tasks, total, err
}

func (r *taskRepository) CountByStatus(ctx context.Context, status domain.TaskStatus) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.Task{}).Where("status = ?", status).Count(&count).Error
	return count, err
}

func (r *taskRepository) GetQueuePosition(ctx context.Context, id string) (int, error) {
	var task domain.Task
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&task).Error; err != nil {
		return 0, err
	}

	var position int64
	err := r.db.WithContext(ctx).Model(&domain.Task{}).
		Where("status = ? AND created_at < ?", domain.TaskStatusQueued, task.CreatedAt).
		Count(&position).Error
	if err != nil {
		return 0, err
	}

	return int(position) + 1, nil
}

func (r *taskRepository) FindExpiredCompleted(ctx context.Context, days int) ([]domain.Task, error) {
	var tasks []domain.Task
	cutoff := time.Now().AddDate(0, 0, -days)
	err := r.db.WithContext(ctx).
		Where("status = ? AND updated_at < ?", domain.TaskStatusCompleted, cutoff).
		Find(&tasks).Error
	return tasks, err
}
