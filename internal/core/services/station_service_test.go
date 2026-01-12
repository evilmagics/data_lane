package services_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"pdf_generator/internal/core/domain"
	"pdf_generator/internal/core/services"
)

type MockStationRepo struct {
	mock.Mock
}

func (m *MockStationRepo) Create(ctx context.Context, station *domain.Station) error {
	args := m.Called(ctx, station)
	return args.Error(0)
}

func (m *MockStationRepo) GetByID(ctx context.Context, id int) (*domain.Station, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Station), args.Error(1)
}

func (m *MockStationRepo) Update(ctx context.Context, station *domain.Station) error {
	args := m.Called(ctx, station)
	return args.Error(0)
}

func (m *MockStationRepo) Delete(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockStationRepo) List(ctx context.Context) ([]domain.Station, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Station), args.Error(1)
}

func (m *MockStationRepo) BatchCreate(ctx context.Context, stations []domain.Station) error {
	args := m.Called(ctx, stations)
	return args.Error(0)
}

func (m *MockStationRepo) BatchUpdate(ctx context.Context, stations []domain.Station) error {
	args := m.Called(ctx, stations)
	return args.Error(0)
}

func (m *MockStationRepo) BatchDelete(ctx context.Context, ids []int) error {
	args := m.Called(ctx, ids)
	return args.Error(0)
}

func TestStationService_Create(t *testing.T) {
	repo := new(MockStationRepo)
	svc := services.NewStationService(repo)
	ctx := context.Background()

	station := &domain.Station{ID: 1, Name: "Test"}

	repo.On("Create", ctx, station).Return(nil).Once()

	err := svc.Create(ctx, station)
	assert.NoError(t, err)

	repo.AssertExpectations(t)
}

func TestStationService_List(t *testing.T) {
	repo := new(MockStationRepo)
	svc := services.NewStationService(repo)
	ctx := context.Background()

	stations := []domain.Station{
		{ID: 1, Name: "A"},
		{ID: 2, Name: "B"},
	}

	repo.On("List", ctx).Return(stations, nil).Once()

	result, err := svc.List(ctx)
	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "A", result[0].Name)

	repo.AssertExpectations(t)
}

func TestStationService_BatchCreate(t *testing.T) {
	repo := new(MockStationRepo)
	svc := services.NewStationService(repo)
	ctx := context.Background()

	stations := []domain.Station{
		{ID: 1, Name: "A"},
		{ID: 2, Name: "B"},
	}

	repo.On("BatchCreate", ctx, stations).Return(nil).Once()

	err := svc.CreateBatch(ctx, stations)
	assert.NoError(t, err)

	repo.AssertExpectations(t)
}
