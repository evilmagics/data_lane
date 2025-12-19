package services

import (
	"context"
	"errors"
	"os"
	"strconv"
	"time"

	"gorm.io/gorm"

	"pdf_generator/internal/core/domain"
	"pdf_generator/internal/core/ports"
	"pdf_generator/pkg/utils"
)

// AuthService handles authentication logic
type AuthService struct {
	sessionRepo   ports.SessionRepository
	settingsRepo  ports.SettingsRepository
	sessionExpiry time.Duration
}

// NewAuthService creates a new auth service
func NewAuthService(sessionRepo ports.SessionRepository, settingsRepo ports.SettingsRepository, sessionExpiry time.Duration) *AuthService {
	return &AuthService{
		sessionRepo:   sessionRepo,
		settingsRepo:  settingsRepo,
		sessionExpiry: sessionExpiry,
	}
}

// Login authenticates admin and creates a session
func (s *AuthService) Login(ctx context.Context, username, password, ip, userAgent string) (string, time.Time, string, error) {
	// Validate credentials (single admin user from env)
	adminUser := os.Getenv("ADMIN_USERNAME")
	adminPass := os.Getenv("ADMIN_PASSWORD")
	if adminUser == "" {
		adminUser = "admin"
	}
	if adminPass == "" {
		adminPass = "admin"
	}

	if username != adminUser || password != adminPass {
		return "", time.Time{}, "", errors.New("invalid credentials")
	}

	// Cleanup expired sessions first to free up slots
	if err := s.sessionRepo.CleanupExpired(ctx); err != nil {
		// Log error but continue? Or fail? Best to continue, but maybe log it.
		// For now we continue as it's maintenance.
	}

	// Check concurrent session limit
	maxSessions := s.getMaxConcurrentSessions(ctx)
	activeCount, err := s.sessionRepo.CountActive(ctx)
	if err != nil {
		return "", time.Time{}, "", err
	}
	if activeCount >= int64(maxSessions) {
		return "", time.Time{}, "", errors.New("session limit reached")
	}

	// Generate JWT
	sessionID := ""
	token, expiresAt, err := utils.GenerateJWT(username, sessionID, s.sessionExpiry)
	if err != nil {
		return "", time.Time{}, "", err
	}

	// Create session record
	session := &domain.Session{
		UserID:    username,
		TokenHash: utils.GetTokenSignature(token),
		IPAddress: ip,
		UserAgent: userAgent,
		ExpiresAt: expiresAt,
	}
	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return "", time.Time{}, "", err
	}

	return token, expiresAt, session.ID, nil
}

// Logout invalidates a session
func (s *AuthService) Logout(ctx context.Context, tokenHash string) error {
	err := s.sessionRepo.DeleteByTokenHash(ctx, tokenHash)
	// If session isn't found, it's already logged out or invalid. Return success.
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	return err
}

// ValidateSession checks if a session is valid
func (s *AuthService) ValidateSession(ctx context.Context, tokenHash string) (*domain.Session, error) {
	session, err := s.sessionRepo.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		return nil, errors.New("session not found")
	}
	
	if session.Status != "active" {
		return nil, errors.New("session is " + session.Status)
	}

	if session.IsExpired() {
		// Update status to expired instead of deleting? 
		// Repo.CleanupExpired() does this in bulk, but we can do it here too for immediate feedback.
		// However, VerifySession shouldn't have side effects if possible, but invalidating immediately is safer.
		// Actually, let's just let the scheduled/login cleanup handle status updates for now or just return error.
		return nil, errors.New("session expired")
	}
	return session, nil
}

// RevokeSession forcibly ends a specific session
func (s *AuthService) RevokeSession(ctx context.Context, sessionID string) error {
	return s.sessionRepo.Delete(ctx, sessionID)
}

// ListActiveSessions returns all active sessions
func (s *AuthService) ListActiveSessions(ctx context.Context) ([]domain.Session, error) {
	return s.sessionRepo.ListActive(ctx)
}

func (s *AuthService) getMaxConcurrentSessions(ctx context.Context) int {
	setting, err := s.settingsRepo.Get(ctx, domain.SettingMaxConcurrentSessions)
	if err != nil || setting == nil {
		return 5
	}
	val, _ := strconv.Atoi(setting.Value)
	if val <= 0 {
		return 5
	}
	return val
}
