package repository

import (
	"context"

	"gorm.io/gorm"

	"pdf_generator/internal/core/domain"
	"pdf_generator/internal/core/ports"
)

type settingsRepository struct {
	db *gorm.DB
}

// NewSettingsRepository creates a new settings repository
func NewSettingsRepository(db *gorm.DB) ports.SettingsRepository {
	return &settingsRepository{db: db}
}

func (r *settingsRepository) Get(ctx context.Context, key string) (*domain.Settings, error) {
	var setting domain.Settings
	err := r.db.WithContext(ctx).Where("key = ?", key).First(&setting).Error
	if err != nil {
		return nil, err
	}
	return &setting, nil
}

func (r *settingsRepository) Set(ctx context.Context, setting *domain.Settings) error {
	return r.db.WithContext(ctx).Save(setting).Error
}

func (r *settingsRepository) GetAll(ctx context.Context) ([]domain.Settings, error) {
	var settings []domain.Settings
	err := r.db.WithContext(ctx).Order("sort_order asc").Find(&settings).Error
	return settings, err
}
