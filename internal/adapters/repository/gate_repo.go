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

func (r *gateRepository) List(ctx context.Context, filter ports.GateFilter) ([]domain.Gate, int64, error) {
	var gates []domain.Gate
	var total int64
	
	db := r.db.WithContext(ctx).Model(&domain.Gate{})

	if filter.Query != "" {
		// Search by ID or Name
		// If query can be parsed as int, search ID too
		db = db.Where("name LIKE ? OR CAST(id AS CHAR) LIKE ?", "%"+filter.Query+"%", "%"+filter.Query+"%")
	}

    if len(filter.IDs) > 0 {
        db = db.Where("id IN ?", filter.IDs)
    }

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

    if filter.Limit > 0 {
	    offset := (filter.Page - 1) * filter.Limit
	    db = db.Limit(filter.Limit).Offset(offset)
    }

	// Added Order("id asc") from Station repo logic
	err := db.Order("id asc").Find(&gates).Error
	return gates, total, err
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
