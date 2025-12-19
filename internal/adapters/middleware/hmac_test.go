package middleware

import (
	"context"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"

	"pdf_generator/internal/core/domain"
	"pdf_generator/internal/core/ports"
	"pdf_generator/internal/core/services"
	"pdf_generator/pkg/utils"
)

// Mock Repo
type mockSettingsRepo struct {
	settings map[string]string
}

func (m *mockSettingsRepo) Get(ctx context.Context, key string) (*domain.Settings, error) {
	if val, ok := m.settings[key]; ok {
		return &domain.Settings{Key: key, Value: val}, nil
	}
	return nil, fmt.Errorf("not found")
}
func (m *mockSettingsRepo) Set(ctx context.Context, setting *domain.Settings) error {
	m.settings[setting.Key] = setting.Value
	return nil
}
func (m *mockSettingsRepo) GetAll(ctx context.Context) ([]domain.Settings, error) {
	return nil, nil
}

// Ensure mock satisfies interface
var _ ports.SettingsRepository = &mockSettingsRepo{}

func TestHMACMiddleware(t *testing.T) {
	// Setup Mock Service
	mockRepo := &mockSettingsRepo{settings: make(map[string]string)}
	settingsService := services.NewSettingsService(mockRepo)
	// Default enabled
	mockRepo.settings["enable_hmac"] = "true"

	app := fiber.New()

	// Apply middleware
	app.Post("/api/protected", HMACMiddleware(settingsService), func(c fiber.Ctx) error {
		return c.SendString("Protected")
	})

	// Login exception route
	app.Post("/api/auth/login", HMACMiddleware(settingsService), func(c fiber.Ctx) error {
		return c.SendString("Login Public")
	})

	// 1. Missing Signature -> Fail (Enabled by default)
	reqMissing := httptest.NewRequest("POST", "/api/protected", strings.NewReader(`{"data":"test"}`))
	reqMissing.Header.Set("Content-Type", "application/json")
	respMissing, err := app.Test(reqMissing)
	assert.NoError(t, err)
	assert.Equal(t, 401, respMissing.StatusCode)

	// 2. Fallback Secret
	defaultSecret := "default-hmac-secret"
	body := `{"data":"fallback"}`
	sigFallback := utils.GenerateHMAC(body, defaultSecret)
	
	reqFallback := httptest.NewRequest("POST", "/api/protected", strings.NewReader(body))
	reqFallback.Header.Set("Content-Type", "application/json")
	reqFallback.Header.Set("X-Signature", sigFallback)
	respFallback, err := app.Test(reqFallback)
	assert.NoError(t, err)
	assert.Equal(t, 200, respFallback.StatusCode)

	// 3. API Key as Secret
	apiKey := "my-api-key"
	bodyApi := `{"data":"apikey"}`
	sigApi := utils.GenerateHMAC(bodyApi, apiKey)

	reqApi := httptest.NewRequest("POST", "/api/protected", strings.NewReader(bodyApi))
	reqApi.Header.Set("Content-Type", "application/json")
	reqApi.Header.Set("X-API-Key", apiKey)
	reqApi.Header.Set("X-Signature", sigApi)
	respApi, err := app.Test(reqApi)
	assert.NoError(t, err)
	assert.Equal(t, 200, respApi.StatusCode)

	// 4. Bearer Token as Secret
	token := "my-jwt-token"
	bodyToken := `{"data":"token"}`
	sigToken := utils.GenerateHMAC(bodyToken, token)

	reqToken := httptest.NewRequest("POST", "/api/protected", strings.NewReader(bodyToken))
	reqToken.Header.Set("Content-Type", "application/json")
	reqToken.Header.Set("Authorization", "Bearer "+token)
	reqToken.Header.Set("X-Signature", sigToken)
	respToken, err := app.Test(reqToken)
	assert.NoError(t, err)
	assert.Equal(t, 200, respToken.StatusCode)

	// 5. Empty Body -> Pass (Skip Validation)
	reqEmpty := httptest.NewRequest("POST", "/api/protected", nil) // nil body
	reqEmpty.Header.Set("Content-Type", "application/json")
	respEmpty, err := app.Test(reqEmpty)
	assert.NoError(t, err)
	assert.Equal(t, 200, respEmpty.StatusCode)

	// 6. Login Route (Exempt)
	reqLogin := httptest.NewRequest("POST", "/api/auth/login", strings.NewReader(`{"user":"admin"}`))
	reqLogin.Header.Set("Content-Type", "application/json")
	respLogin, err := app.Test(reqLogin)
	assert.NoError(t, err)
	assert.Equal(t, 200, respLogin.StatusCode)

	// 7. Disabled via Settings
	_ = settingsService.Set(context.Background(), "enable_hmac", "false")
	
	reqDisabled := httptest.NewRequest("POST", "/api/protected", strings.NewReader(`{"data":"should_pass_without_sig"}`))
	reqDisabled.Header.Set("Content-Type", "application/json")
	// No X-Signature
	respDisabled, err := app.Test(reqDisabled)
	assert.NoError(t, err)
	assert.Equal(t, 200, respDisabled.StatusCode)
}
