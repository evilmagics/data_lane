package services

import (
	"context"
	"testing"

	"pdf_generator/internal/core/domain"
	"pdf_generator/internal/core/ports"
)

// Mock Repo
type MockGateRepo struct{}

func (m *MockGateRepo) Create(ctx context.Context, gate *domain.Gate) error { return nil }
func (m *MockGateRepo) GetByID(ctx context.Context, id int) (*domain.Gate, error) {
	return &domain.Gate{ID: id, Name: "Test Gate"}, nil
}
func (m *MockGateRepo) Update(ctx context.Context, gate *domain.Gate) error { return nil }
func (m *MockGateRepo) Delete(ctx context.Context, id int) error        { return nil }
func (m *MockGateRepo) List(ctx context.Context, filter ports.GateFilter) ([]domain.Gate, int64, error) {
	return []domain.Gate{{ID: 1, Name: "G1"}}, 1, nil
}
func (m *MockGateRepo) BatchCreate(ctx context.Context, gates []domain.Gate) error { return nil }
func (m *MockGateRepo) BatchUpdate(ctx context.Context, gates []domain.Gate) error { return nil }
func (m *MockGateRepo) BatchDelete(ctx context.Context, ids []int) error        { return nil }

func TestGateService(t *testing.T) {
    s := NewGateService(&MockGateRepo{})
    if s == nil {
        t.Fatal("Service invalid")
    }
    
    // Test Get
    g, err := s.GetByID(context.Background(), 1)
    if err != nil {
        t.Errorf("GetByID failed: %v", err)
    }
    if g.Name != "Test Gate" {
        t.Errorf("Expected Test Gate, got %s", g.Name)
    }
}
