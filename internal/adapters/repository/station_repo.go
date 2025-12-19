package repository

import (
	"context"

	"gorm.io/gorm"

	"pdf_generator/internal/core/domain"
	"pdf_generator/internal/core/ports"
)

type stationRepository struct {
	db *gorm.DB
}

// NewStationRepository creates a new station repository
func NewStationRepository(db *gorm.DB) ports.StationRepository {
	return &stationRepository{db: db}
}

func (r *stationRepository) Create(ctx context.Context, station *domain.Station) error {
	return r.db.WithContext(ctx).Create(station).Error
}

func (r *stationRepository) GetByID(ctx context.Context, id int) (*domain.Station, error) {
	var station domain.Station
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&station).Error
	if err != nil {
		return nil, err
	}
	return &station, nil
}

func (r *stationRepository) Update(ctx context.Context, station *domain.Station) error {
	return r.db.WithContext(ctx).Save(station).Error
}

func (r *stationRepository) Delete(ctx context.Context, id int) error {
	return r.db.WithContext(ctx).Delete(&domain.Station{}, "id = ?", id).Error
}

func (r *stationRepository) List(ctx context.Context) ([]domain.Station, error) {
	var stations []domain.Station
	err := r.db.WithContext(ctx).Find(&stations).Error
	return stations, err
}

func (r *stationRepository) BatchCreate(ctx context.Context, stations []domain.Station) error {
	return r.db.WithContext(ctx).Create(&stations).Error
}

func (r *stationRepository) BatchUpdate(ctx context.Context, stations []domain.Station) error {
	return r.db.WithContext(ctx).Save(&stations).Error
}

func (r *stationRepository) BatchDelete(ctx context.Context, ids []int) error {
	return r.db.WithContext(ctx).Delete(&domain.Station{}, "id IN ?", ids).Error
}
