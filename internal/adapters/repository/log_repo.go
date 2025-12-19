package repository

import (
	"context"
	"time"

	"gorm.io/gorm"

	"pdf_generator/internal/core/domain"
	"pdf_generator/internal/core/ports"
)

type logRepository struct {
	db *gorm.DB
}

// NewLogRepository creates a new log repository
func NewLogRepository(db *gorm.DB) ports.LogRepository {
	return &logRepository{db: db}
}

func (r *logRepository) Create(ctx context.Context, log *domain.Log) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *logRepository) List(ctx context.Context, limit int, offset int) ([]domain.Log, error) {
	var logs []domain.Log
	err := r.db.WithContext(ctx).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&logs).Error
	return logs, err
}

func (r *logRepository) CleanupOld(ctx context.Context, days int) error {
	cutoff := time.Now().AddDate(0, 0, -days)
	return r.db.WithContext(ctx).Delete(&domain.Log{}, "created_at < ?", cutoff).Error
}
