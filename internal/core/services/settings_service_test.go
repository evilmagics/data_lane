package services_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"pdf_generator/internal/core/domain"
	"pdf_generator/internal/core/services"
)

type MockSettingsRepo struct {
	mock.Mock
}

func (m *MockSettingsRepo) Get(ctx context.Context, key string) (*domain.Settings, error) {
	args := m.Called(ctx, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Settings), args.Error(1)
}

func (m *MockSettingsRepo) Set(ctx context.Context, setting *domain.Settings) error {
	args := m.Called(ctx, setting)
	return args.Error(0)
}

func (m *MockSettingsRepo) GetAll(ctx context.Context) ([]domain.Settings, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Settings), args.Error(1)
}

func TestSettingsService_Cache(t *testing.T) {
	repo := new(MockSettingsRepo)
	svc := services.NewSettingsService(repo)

	ctx := context.Background()
	testKey := "test_key"
	testVal := "test_value"

	// Mock data for LoadCache
	settings := []domain.Settings{
		{Key: testKey, Value: testVal},
	}
	
	// Expect GetAll to be called once during LoadCache
	repo.On("GetAll", ctx).Return(settings, nil).Once()

	// 1. Load Cache
	err := svc.LoadCache(ctx)
	assert.NoError(t, err)

	// 2. Get from Cache (should NOT call repo.Get)
	val, err := svc.Get(ctx, testKey)
	assert.NoError(t, err)
	assert.Equal(t, testVal, val)

	// 3. Update Setting -> Should update Repo AND Cache (no repo.Get since it's cached)
	newVal := "new_value"
	
	// Expect repo.Set
	repo.On("Set", ctx, mock.MatchedBy(func(s *domain.Settings) bool {
		return s.Key == testKey && s.Value == newVal
	})).Return(nil).Once()

	err = svc.Set(ctx, testKey, newVal)
	assert.NoError(t, err)

	// 4. Get from Cache again (should be updated value, NO repo call)
	val, err = svc.Get(ctx, testKey)
	assert.NoError(t, err)
	assert.Equal(t, newVal, val)

	// 5. GetAll from Cache (should NOT call repo.GetAll)
	all, err := svc.GetAll(ctx)
	assert.NoError(t, err)
	assert.Len(t, all, 1)
	assert.Equal(t, newVal, all[0].Value)

	repo.AssertExpectations(t)
}
