package services

import (
	"context"

	"pdf_generator/internal/core/domain"
	"pdf_generator/internal/core/ports"
	"pdf_generator/pkg/utils"
)

// APIKeyService handles API key operations
type APIKeyService struct {
	repo ports.APIKeyRepository
}

// NewAPIKeyService creates a new API key service
func NewAPIKeyService(repo ports.APIKeyRepository) *APIKeyService {
	return &APIKeyService{repo: repo}
}

// Create generates a new API key
func (s *APIKeyService) Create(ctx context.Context, name string) (*domain.APIKey, string, error) {
	// Generate random key
	rawKey, err := utils.GenerateRandomKey(16)
	if err != nil {
		return nil, "", err
	}

	// Hash for verification
	keyHash, err := utils.HashPassword(rawKey)
	if err != nil {
		return nil, "", err
	}

	// Encrypt for later reveal
	encryptedKey, err := utils.Encrypt(rawKey)
	if err != nil {
		return nil, "", err
	}

	apiKey := &domain.APIKey{
		Name:         name,
		KeyHash:      keyHash,
		EncryptedKey: encryptedKey,
		Active:       true,
	}

	if err := s.repo.Create(ctx, apiKey); err != nil {
		return nil, "", err
	}

	return apiKey, rawKey, nil
}

// Validate checks if an API key is valid
func (s *APIKeyService) Validate(ctx context.Context, rawKey string) (*domain.APIKey, bool) {
	keys, err := s.repo.List(ctx)
	if err != nil {
		return nil, false
	}

	for _, key := range keys {
		if key.Active && utils.CheckPasswordHash(rawKey, key.KeyHash) {
			return &key, true
		}
	}
	return nil, false
}

// Reveal decrypts and returns the API key
func (s *APIKeyService) Reveal(ctx context.Context, id string) (string, error) {
	key, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return "", err
	}
	return utils.Decrypt(key.EncryptedKey)
}

// Toggle activates or deactivates an API key
func (s *APIKeyService) Toggle(ctx context.Context, id string) (*domain.APIKey, error) {
	key, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	key.Active = !key.Active
	if err := s.repo.Update(ctx, key); err != nil {
		return nil, err
	}
	return key, nil
}

// Delete removes an API key
func (s *APIKeyService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

// List returns all API keys (without revealing the key)
func (s *APIKeyService) List(ctx context.Context) ([]domain.APIKey, error) {
	return s.repo.List(ctx)
}
