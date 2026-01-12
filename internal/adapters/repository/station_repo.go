package repository

import (
	"context"

	"gorm.io/gorm"

	"pdf_generator/internal/core/domain"
	"pdf_generator/internal/core/ports"
)

type StationRepository struct {
	db *gorm.DB
}

func NewStationRepository(db *gorm.DB) ports.StationRepository {
	return &StationRepository{db: db}
}

func (r *StationRepository) Create(ctx context.Context, station *domain.Station) error {
	return r.db.WithContext(ctx).Create(station).Error
}

func (r *StationRepository) GetByID(ctx context.Context, id int) (*domain.Station, error) {
	var station domain.Station
	err := r.db.WithContext(ctx).First(&station, id).Error
	return &station, err
}

func (r *StationRepository) Update(ctx context.Context, station *domain.Station) error {
	return r.db.WithContext(ctx).Save(station).Error
}

func (r *StationRepository) Delete(ctx context.Context, id int) error {
	return r.db.WithContext(ctx).Delete(&domain.Station{}, id).Error
}

func (r *StationRepository) List(ctx context.Context) ([]domain.Station, error) {
	var stations []domain.Station
	err := r.db.WithContext(ctx).Order("id asc").Find(&stations).Error
	return stations, err
}

func (r *StationRepository) BatchCreate(ctx context.Context, stations []domain.Station) error {
	return r.db.WithContext(ctx).Create(&stations).Error
}

func (r *StationRepository) BatchUpdate(ctx context.Context, stations []domain.Station) error {
	return r.db.WithContext(ctx).Save(&stations).Error
}

func (r *StationRepository) BatchDelete(ctx context.Context, ids []int) error {
	return r.db.WithContext(ctx).Delete(&domain.Station{}, ids).Error
}
