package repository

import (
	"context"

	"gorm.io/gorm"

	"pdf_generator/internal/core/domain"
	"pdf_generator/internal/core/ports"
)

type scheduleRepository struct {
	db *gorm.DB
}

// NewScheduleRepository creates a new schedule repository
func NewScheduleRepository(db *gorm.DB) ports.ScheduleRepository {
	return &scheduleRepository{db: db}
}

func (r *scheduleRepository) Create(ctx context.Context, schedule *domain.Schedule) error {
	return r.db.WithContext(ctx).Create(schedule).Error
}

func (r *scheduleRepository) GetByID(ctx context.Context, id string) (*domain.Schedule, error) {
	var schedule domain.Schedule
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&schedule).Error
	if err != nil {
		return nil, err
	}
	return &schedule, nil
}

func (r *scheduleRepository) Update(ctx context.Context, schedule *domain.Schedule) error {
	return r.db.WithContext(ctx).Save(schedule).Error
}

func (r *scheduleRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&domain.Schedule{}, "id = ?", id).Error
}

func (r *scheduleRepository) ListActive(ctx context.Context) ([]domain.Schedule, error) {
	var schedules []domain.Schedule
	err := r.db.WithContext(ctx).Where("active = ?", true).Find(&schedules).Error
	return schedules, err
}

func (r *scheduleRepository) List(ctx context.Context) ([]domain.Schedule, error) {
	var schedules []domain.Schedule
	err := r.db.WithContext(ctx).Order("created_at DESC").Find(&schedules).Error
	return schedules, err
}
