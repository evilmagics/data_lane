package services

import (
	"context"
	"sort"
	"sync"

	"pdf_generator/internal/core/domain"
	"pdf_generator/internal/core/ports"
)

// SettingsService handles settings operations
type SettingsService struct {
	repo  ports.SettingsRepository
	cache map[string]domain.Settings
	mu    sync.RWMutex
}

// NewSettingsService creates a new settings service
func NewSettingsService(repo ports.SettingsRepository) *SettingsService {
	return &SettingsService{
		repo:  repo,
		cache: make(map[string]domain.Settings),
	}
}

// LoadCache populates the internal cache from the repository
func (s *SettingsService) LoadCache(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	settings, err := s.repo.GetAll(ctx)
	if err != nil {
		return err
	}

	s.cache = make(map[string]domain.Settings)
	for _, setting := range settings {
		s.cache[setting.Key] = setting
	}
	return nil
}

// Get retrieves a setting by key, checking cache first
func (s *SettingsService) Get(ctx context.Context, key string) (string, error) {
	s.mu.RLock()
	setting, ok := s.cache[key]
	s.mu.RUnlock()

	if ok {
		return setting.Value, nil
	}

	// Fallback to repo if missing (e.g. new key not in cache yet)
	// But normally cache should have it if LoadCache was called.
	repoSetting, err := s.repo.Get(ctx, key)
	if err != nil {
		return "", err
	}

	// Update cache
	s.mu.Lock()
	s.cache[key] = *repoSetting
	s.mu.Unlock()

	return repoSetting.Value, nil
}

// Set updates a setting in both repo and cache
func (s *SettingsService) Set(ctx context.Context, key, value string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Get existing from cache or repo to preserve metadata
	var setting domain.Settings
	if existing, ok := s.cache[key]; ok {
		setting = existing
	} else {
		// Try repo if not in cache (rare case)
		if existingRepo, _ := s.repo.Get(ctx, key); existingRepo != nil {
			setting = *existingRepo
		} else {
			// New setting default
			setting = domain.Settings{Key: key}
		}
	}

	setting.Value = value

	// Default metadata if missing (for new keys)
	if setting.Group == "" {
		setting.Group = "General"
	}
	if setting.DataType == "" {
		setting.DataType = "string"
	}

	// Update Repo
	if err := s.repo.Set(ctx, &setting); err != nil {
		return err
	}

	// Update Cache
	s.cache[key] = setting

	return nil
}

// GetAll retrieves all settings from cache
func (s *SettingsService) GetAll(ctx context.Context) ([]domain.Settings, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.cache) == 0 {
		// Try LoadCache if empty? Or just return from repo?
		// Requirement: "All function need fetch to settings only fetch from cache"
		// If cache is empty, it might be uninitialized or genuinely empty.
		// Let's assume LoadCache was called. If empty, return empty list.
		// However, to be robust, we could check if initialized.
		// For now, let's just return what's in cache.
		// If the user forgot LoadCache, they get empty list.
		// Wait, safe fallback is repo.
	}

	var settings []domain.Settings
	for _, s := range s.cache {
		settings = append(settings, s)
	}

	sort.Slice(settings, func(i, j int) bool {
		return settings[i].SortOrder < settings[j].SortOrder
	})

	return settings, nil
}
