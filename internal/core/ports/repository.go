package ports

import (
	"context"

	"pdf_generator/internal/core/domain"
)

// TaskRepository defines the interface for task data access
type TaskRepository interface {
	Create(ctx context.Context, task *domain.Task) error
	GetByID(ctx context.Context, id string) (*domain.Task, error)
	Update(ctx context.Context, task *domain.Task) error
	UpdateProgress(ctx context.Context, id string, stage string, current, total int) error
	UpdateError(ctx context.Context, id string, errMsg string) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, filter TaskFilter) ([]domain.Task, int64, error)
	CountByStatus(ctx context.Context, status domain.TaskStatus) (int64, error)
	GetQueuePosition(ctx context.Context, id string) (int, error)
	FindExpiredCompleted(ctx context.Context, days int) ([]domain.Task, error)
}

// TaskFilter for listing tasks
type TaskFilter struct {
	Status   string
	FromDate string
	ToDate   string
	Page     int
	Limit    int
}

// ScheduleRepository defines the interface for schedule data access
type ScheduleRepository interface {
	Create(ctx context.Context, schedule *domain.Schedule) error
	GetByID(ctx context.Context, id string) (*domain.Schedule, error)
	Update(ctx context.Context, schedule *domain.Schedule) error
	Delete(ctx context.Context, id string) error
	ListActive(ctx context.Context) ([]domain.Schedule, error)
	List(ctx context.Context) ([]domain.Schedule, error)
}

// SettingsRepository defines the interface for settings data access
type SettingsRepository interface {
	Get(ctx context.Context, key string) (*domain.Settings, error)
	Set(ctx context.Context, setting *domain.Settings) error
	GetAll(ctx context.Context) ([]domain.Settings, error)
}

// SessionRepository defines the interface for session data access
type SessionRepository interface {
	Create(ctx context.Context, session *domain.Session) error
	GetByID(ctx context.Context, id string) (*domain.Session, error)
	GetByTokenHash(ctx context.Context, hash string) (*domain.Session, error)
	Delete(ctx context.Context, id string) error
	DeleteByTokenHash(ctx context.Context, hash string) error
	CountActive(ctx context.Context) (int64, error)
	ListActive(ctx context.Context) ([]domain.Session, error)
	CleanupExpired(ctx context.Context) error
}

// APIKeyRepository defines the interface for API key data access
type APIKeyRepository interface {
	Create(ctx context.Context, key *domain.APIKey) error
	GetByID(ctx context.Context, id string) (*domain.APIKey, error)
	GetByKeyHash(ctx context.Context, hash string) (*domain.APIKey, error)
	Update(ctx context.Context, key *domain.APIKey) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context) ([]domain.APIKey, error)
}

// LogRepository defines the interface for log data access
type LogRepository interface {
	Create(ctx context.Context, log *domain.Log) error
	List(ctx context.Context, limit int, offset int) ([]domain.Log, error)
	CleanupOld(ctx context.Context, days int) error
}

// StationRepository defines the interface for station data access
type StationRepository interface {
	Create(ctx context.Context, station *domain.Station) error
	GetByID(ctx context.Context, id int) (*domain.Station, error)
	Update(ctx context.Context, station *domain.Station) error
	Delete(ctx context.Context, id int) error
	List(ctx context.Context) ([]domain.Station, error)
	BatchCreate(ctx context.Context, stations []domain.Station) error
	BatchUpdate(ctx context.Context, stations []domain.Station) error
	BatchDelete(ctx context.Context, ids []int) error
}

