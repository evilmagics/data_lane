package queue_test

import (
	"context"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"

	"pdf_generator/internal/core/domain"
	"pdf_generator/internal/core/ports"
	"pdf_generator/pkg/queue"
)

// Mocks
type MockTaskRepo struct {
	mock.Mock
}
// Stub methods for MockTaskRepo
func (m *MockTaskRepo) Create(ctx context.Context, task *domain.Task) error { return nil }
func (m *MockTaskRepo) GetByID(ctx context.Context, id string) (*domain.Task, error) { return nil, nil }
func (m *MockTaskRepo) Update(ctx context.Context, task *domain.Task) error { return nil }
func (m *MockTaskRepo) Delete(ctx context.Context, id string) error { return nil }
func (m *MockTaskRepo) List(ctx context.Context, filter ports.TaskFilter) ([]domain.Task, int64, error) { 
	return nil, 0, nil
}
func (m *MockTaskRepo) CountByStatus(ctx context.Context, status domain.TaskStatus) (int64, error) { return 0, nil }
func (m *MockTaskRepo) GetQueuePosition(ctx context.Context, id string) (int, error) { return 0, nil }
func (m *MockTaskRepo) FindExpiredCompleted(ctx context.Context, days int) ([]domain.Task, error) { return nil, nil }

// We need to match the signature of List EXACTLY with ports definition, which I can't check easily without looking at ports. 
// Assuming ports.TaskFilter.

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
func (m *MockSettingsRepo) Set(ctx context.Context, setting *domain.Settings) error { return nil }
func (m *MockSettingsRepo) GetAll(ctx context.Context) ([]domain.Settings, error) { return nil, nil }


func TestNewQueue(t *testing.T) {
	// Setup in-memory SQLite DB for backlite
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	assert.NoError(t, err)

	taskRepo := new(MockTaskRepo)
	settingsRepo := new(MockSettingsRepo)

	// Mock settings call
	settingsRepo.On("Get", mock.Anything, domain.SettingQueueConcurrency).Return(&domain.Settings{Value: "1"}, nil).Once()

	q, err := queue.NewQueue(db, taskRepo, settingsRepo)
	assert.NoError(t, err)
	assert.NotNil(t, q)

	settingsRepo.AssertExpectations(t)
}

func TestPDFTask_Config(t *testing.T) {
	task := queue.PDFTask{TaskID: "123"}
	cfg := task.Config()
	assert.Equal(t, "generate_pdf", cfg.Name)
	assert.Equal(t, 3, cfg.MaxAttempts)
}
