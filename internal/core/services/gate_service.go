package services

import (
	"context"

	"pdf_generator/internal/core/domain"
	"pdf_generator/internal/core/ports"
)

type GateService struct {
	repo ports.GateRepository
}

func NewGateService(repo ports.GateRepository) *GateService {
	return &GateService{repo: repo}
}

func (s *GateService) Create(ctx context.Context, gate *domain.Gate) error {
	return s.repo.Create(ctx, gate)
}

func (s *GateService) CreateBatch(ctx context.Context, gates []domain.Gate) error {
	return s.repo.BatchCreate(ctx, gates)
}

func (s *GateService) List(ctx context.Context) ([]domain.Gate, error) {
	return s.repo.List(ctx)
}

func (s *GateService) GetByID(ctx context.Context, id int) (*domain.Gate, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *GateService) Update(ctx context.Context, id int, name string) error {
	gate, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	gate.Name = name
	return s.repo.Update(ctx, gate)
}

func (s *GateService) BatchUpdate(ctx context.Context, gates []domain.Gate) error {
	return s.repo.BatchUpdate(ctx, gates)
}

func (s *GateService) Delete(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}

func (s *GateService) BatchDelete(ctx context.Context, ids []int) error {
	return s.repo.BatchDelete(ctx, ids)
}
