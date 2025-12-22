package repository

import (
	"context"

	"gorm.io/gorm"

	"pdf_generator/internal/core/domain"
	"pdf_generator/internal/core/ports"
)

type gateRepository struct {
	db *gorm.DB
}

// NewGateRepository creates a new gate repository
func NewGateRepository(db *gorm.DB) ports.GateRepository {
	return &gateRepository{db: db}
}

func (r *gateRepository) Create(ctx context.Context, gate *domain.Gate) error {
	return r.db.WithContext(ctx).Create(gate).Error
}

func (r *gateRepository) GetByID(ctx context.Context, id int) (*domain.Gate, error) {
	var gate domain.Gate
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&gate).Error
	if err != nil {
		return nil, err
	}
	return &gate, nil
}

func (r *gateRepository) Update(ctx context.Context, gate *domain.Gate) error {
	return r.db.WithContext(ctx).Save(gate).Error
}

func (r *gateRepository) Delete(ctx context.Context, id int) error {
	return r.db.WithContext(ctx).Delete(&domain.Gate{}, "id = ?", id).Error
}

func (r *gateRepository) List(ctx context.Context) ([]domain.Gate, error) {
	var gates []domain.Gate
	err := r.db.WithContext(ctx).Find(&gates).Error
	return gates, err
}

func (r *gateRepository) BatchCreate(ctx context.Context, gates []domain.Gate) error {
	return r.db.WithContext(ctx).Create(&gates).Error
}

func (r *gateRepository) BatchUpdate(ctx context.Context, gates []domain.Gate) error {
	return r.db.WithContext(ctx).Save(&gates).Error
}

func (r *gateRepository) BatchDelete(ctx context.Context, ids []int) error {
	return r.db.WithContext(ctx).Delete(&domain.Gate{}, "id IN ?", ids).Error
}
