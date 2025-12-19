package repository

import (
	"context"
	"time"

	"gorm.io/gorm"

	"pdf_generator/internal/core/domain"
	"pdf_generator/internal/core/ports"
)

type sessionRepository struct {
	db *gorm.DB
}

// NewSessionRepository creates a new session repository
func NewSessionRepository(db *gorm.DB) ports.SessionRepository {
	return &sessionRepository{db: db}
}

func (r *sessionRepository) Create(ctx context.Context, session *domain.Session) error {
	return r.db.WithContext(ctx).Create(session).Error
}

func (r *sessionRepository) GetByID(ctx context.Context, id string) (*domain.Session, error) {
	var session domain.Session
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *sessionRepository) GetByTokenHash(ctx context.Context, hash string) (*domain.Session, error) {
	var session domain.Session
	err := r.db.WithContext(ctx).Where("token_hash = ?", hash).First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *sessionRepository) Delete(ctx context.Context, id string) error {
	// Soft delete: mark as revoked
	return r.db.WithContext(ctx).Model(&domain.Session{}).
		Where("id = ?", id).
		Update("status", "revoked").Error
}

func (r *sessionRepository) DeleteByTokenHash(ctx context.Context, hash string) error {
	// Soft delete: mark as revoked
	result := r.db.WithContext(ctx).Model(&domain.Session{}).
		Where("token_hash = ?", hash).
		Update("status", "revoked")
	
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *sessionRepository) CountActive(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.Session{}).
		Where("expires_at > ? AND status = 'active'", time.Now()).
		Count(&count).Error
	return count, err
}

func (r *sessionRepository) ListActive(ctx context.Context) ([]domain.Session, error) {
	// Rename to ListRecent internally? Or just return all non-hard-deleted?
	// The interface calls it ListActive, but user wants to see revoked too.
	// We will return all sessions that aren't hard deleted, sorted by active status and date.
	var sessions []domain.Session
	err := r.db.WithContext(ctx).
		Order("CASE WHEN status = 'active' THEN 0 ELSE 1 END, created_at DESC").
		Find(&sessions).Error
	return sessions, err
}

func (r *sessionRepository) CleanupExpired(ctx context.Context) error {
	// 1. Mark expired active sessions as 'expired'
	if err := r.db.WithContext(ctx).Model(&domain.Session{}).
		Where("expires_at < ? AND status = 'active'", time.Now()).
		Update("status", "expired").Error; err != nil {
		return err
	}

	// 2. Hard delete sessions older than 30 days (from UpdatedAt or ExpiresAt)
	retentionCutoff := time.Now().AddDate(0, 0, -30)
	return r.db.WithContext(ctx).
		Delete(&domain.Session{}, "updated_at < ? AND status IN ('revoked', 'expired')", retentionCutoff).Error
}
