package services

import (
	"context"

	"pdf_generator/internal/core/domain"
	"pdf_generator/internal/core/ports"
)

type StationService struct {
	repo ports.StationRepository
}

func NewStationService(repo ports.StationRepository) *StationService {
	return &StationService{repo: repo}
}

func (s *StationService) Create(ctx context.Context, station *domain.Station) error {
	return s.repo.Create(ctx, station)
}

func (s *StationService) CreateBatch(ctx context.Context, stations []domain.Station) error {
	return s.repo.BatchCreate(ctx, stations)
}

func (s *StationService) Get(ctx context.Context, id int) (*domain.Station, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *StationService) List(ctx context.Context) ([]domain.Station, error) {
	return s.repo.List(ctx)
}

func (s *StationService) Update(ctx context.Context, id int, name string) error {
	station, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	station.Name = name
	return s.repo.Update(ctx, station)
}

func (s *StationService) Delete(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}

func (s *StationService) BatchDelete(ctx context.Context, ids []int) error {
	return s.repo.BatchDelete(ctx, ids)
}

func (s *StationService) BatchUpdate(ctx context.Context, stations []domain.Station) error {
	return s.repo.BatchUpdate(ctx, stations)
}

