package handlers_test

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"pdf_generator/internal/adapters/handlers"
	"pdf_generator/internal/core/domain"
	"pdf_generator/internal/core/ports"
)

// Mocks
type MockTaskRepo struct {
	mock.Mock
}

func (m *MockTaskRepo) Create(ctx context.Context, task *domain.Task) error {
	task.ID = "test-task-id" // Simulate ID generation
	task.CreatedAt = time.Now()
	args := m.Called(ctx, task)
	return args.Error(0)
}

func (m *MockTaskRepo) GetQueuePosition(ctx context.Context, id string) (int, error) {
	args := m.Called(ctx, id)
	return args.Int(0), args.Error(1)
}

func (m *MockTaskRepo) CountByStatus(ctx context.Context, status domain.TaskStatus) (int64, error) {
	args := m.Called(ctx, status)
	return args.Get(0).(int64), args.Error(1)
}

// ... other TaskRepo methods if needed
func (m *MockTaskRepo) GetByID(ctx context.Context, id string) (*domain.Task, error) { return nil, nil }
func (m *MockTaskRepo) Update(ctx context.Context, task *domain.Task) error          { return nil }
func (m *MockTaskRepo) UpdateProgress(ctx context.Context, id string, stage string, current, total int) error {
	return nil
}
func (m *MockTaskRepo) UpdateError(ctx context.Context, id string, errMsg string) error { return nil }
func (m *MockTaskRepo) Delete(ctx context.Context, id string) error                     { return nil }
func (m *MockTaskRepo) List(ctx context.Context, filter ports.TaskFilter) ([]domain.Task, int64, error) {
	return nil, 0, nil
}
func (m *MockTaskRepo) FindExpiredCompleted(ctx context.Context, days int) ([]domain.Task, error) {
	return nil, nil
}

type MockQueue struct {
	mock.Mock
}

func (m *MockQueue) Enqueue(ctx context.Context, taskID string, metadata domain.TaskMetadata) ([]string, error) {
	args := m.Called(ctx, taskID, metadata)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockQueue) Start(ctx context.Context) {
	m.Called(ctx)
}

func (m *MockQueue) GetProgress(taskID string) *ports.TaskProgress {
	return nil
}

func TestTaskHandler_Enqueue_PathNormalization(t *testing.T) {
	taskRepo := new(MockTaskRepo)
	queue := new(MockQueue)
	handler := handlers.NewTaskHandler(taskRepo, queue)

	app := fiber.New()
	app.Post("/queue", handler.Enqueue)

	// Payload with forward slashes (even on Windows)
	inputPath := "D:/test/data"
	reqBody := handlers.EnqueueRequest{
		RootFolder: inputPath,
		BranchID:   1,
		GateID:     2,
	}
	body, _ := json.Marshal(reqBody)

	// Expectations
	taskRepo.On("Create", mock.Anything, mock.MatchedBy(func(task *domain.Task) bool {
		// Verify normalized path in task
		if runtime.GOOS == "windows" {
			return task.RootFolder == "D:\\test\\data"
		}
		return task.RootFolder == "D:/test/data"
	})).Return(nil).Once()

	queue.On("Enqueue", mock.Anything, "test-task-id", mock.MatchedBy(func(m domain.TaskMetadata) bool {
		// Verify normalized path in metadata
		if runtime.GOOS == "windows" {
			return m.RootFolder == "D:\\test\\data"
		}
		return m.RootFolder == "D:/test/data"
	})).Return([]string{"job-id"}, nil).Once()

	taskRepo.On("GetQueuePosition", mock.Anything, "test-task-id").Return(1, nil).Once()
	taskRepo.On("CountByStatus", mock.Anything, domain.TaskStatusQueued).Return(int64(1), nil).Once()

	// Request
	req := httptest.NewRequest("POST", "/queue", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	taskRepo.AssertExpectations(t)
	queue.AssertExpectations(t)
}

func TestTaskHandler_Enqueue(t *testing.T) {
	taskRepo := new(MockTaskRepo)
	queue := new(MockQueue)
	handler := handlers.NewTaskHandler(taskRepo, queue)

	app := fiber.New()
	app.Post("/queue", handler.Enqueue)

	// Payload
	// Payload
	reqBody := handlers.EnqueueRequest{
		BranchID: 1,
		GateID:   2,

		Settings: map[string]string{"foo": "bar"},
	}
	body, _ := json.Marshal(reqBody)

	// Expectations
	taskRepo.On("Create", mock.Anything, mock.MatchedBy(func(t *domain.Task) bool {
		return t.Status == domain.TaskStatusQueued
	})).Return(nil).Once()

	// IMPORTANT: Verify Queue.Enqueue is called
	queue.On("Enqueue", mock.Anything, "test-task-id", mock.MatchedBy(func(m domain.TaskMetadata) bool {
		return m.BranchID == 1
	})).Return([]string{"job-id"}, nil).Once()

	taskRepo.On("GetQueuePosition", mock.Anything, "test-task-id").Return(5, nil).Once()
	taskRepo.On("CountByStatus", mock.Anything, domain.TaskStatusQueued).Return(int64(10), nil).Once()

	// Request
	req := httptest.NewRequest("POST", "/queue", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	taskRepo.AssertExpectations(t)
	queue.AssertExpectations(t)
}
