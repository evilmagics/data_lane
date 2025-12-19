package repository

import (
	"context"

	"gorm.io/gorm"

	"pdf_generator/internal/core/domain"
	"pdf_generator/internal/core/ports"
)

type apiKeyRepository struct {
	db *gorm.DB
}

// NewAPIKeyRepository creates a new API key repository
func NewAPIKeyRepository(db *gorm.DB) ports.APIKeyRepository {
	return &apiKeyRepository{db: db}
}

func (r *apiKeyRepository) Create(ctx context.Context, key *domain.APIKey) error {
	return r.db.WithContext(ctx).Create(key).Error
}

func (r *apiKeyRepository) GetByID(ctx context.Context, id string) (*domain.APIKey, error) {
	var key domain.APIKey
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&key).Error
	if err != nil {
		return nil, err
	}
	return &key, nil
}

func (r *apiKeyRepository) GetByKeyHash(ctx context.Context, hash string) (*domain.APIKey, error) {
	var key domain.APIKey
	err := r.db.WithContext(ctx).Where("key_hash = ?", hash).First(&key).Error
	if err != nil {
		return nil, err
	}
	return &key, nil
}

func (r *apiKeyRepository) Update(ctx context.Context, key *domain.APIKey) error {
	return r.db.WithContext(ctx).Save(key).Error
}

func (r *apiKeyRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&domain.APIKey{}, "id = ?", id).Error
}

func (r *apiKeyRepository) List(ctx context.Context) ([]domain.APIKey, error) {
	var keys []domain.APIKey
	err := r.db.WithContext(ctx).Order("created_at DESC").Find(&keys).Error
	return keys, err
}
