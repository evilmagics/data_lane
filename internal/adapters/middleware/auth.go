package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v3"

	"pdf_generator/internal/core/domain"
	"pdf_generator/internal/core/services"
	"pdf_generator/pkg/api"
	"pdf_generator/pkg/utils"
)

// AuthMiddleware validates admin JWT or API key
func AuthMiddleware(authService *services.AuthService, apiKeyService *services.APIKeyService, settingsService *services.SettingsService) fiber.Handler {
	return func(c fiber.Ctx) error {
		// Public routes exception
		if strings.HasSuffix(c.Path(), "/auth/login") {
			return c.Next()
		}

		// Check if security is enabled
		securityEnabled, _ := settingsService.Get(c.Context(), domain.SettingSecurityEnabled)
		if securityEnabled != "true" {
			// Security disabled - allow all requests as anonymous
			c.Locals("auth_type", "anonymous")
			return c.Next()
		}

		// Check Bearer token first (Admin)
		authHeader := c.Get("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			token := strings.TrimPrefix(authHeader, "Bearer ")
			claims, err := utils.ParseJWT(token)
			if err != nil {
				return api.Error(c, api.CodeTokenExpired, "Invalid or expired token")
			}

			// Validate session exists
			tokenHash := utils.GetTokenSignature(token)
			session, err := authService.ValidateSession(c.Context(), tokenHash)
			if err != nil {
				return api.Error(c, api.CodeUnauthorized, err.Error())
			}

			// Store auth info in context
			c.Locals("user_id", claims.UserID)
			c.Locals("session_id", session.ID)
			c.Locals("auth_type", "admin")
			return c.Next()
		}

		// Check API Key
		apiKey := c.Get("X-API-Key")
		if apiKey != "" {
			key, valid := apiKeyService.Validate(c.Context(), apiKey)
			if !valid {
				return api.Error(c, api.CodeInvalidAPIKey, "Invalid API key")
			}
			c.Locals("api_key_id", key.ID)
			c.Locals("auth_type", "api_key")
			return c.Next()
		}

		return api.Error(c, api.CodeUnauthorized, "Authentication required")
	}
}

// AdminOnly restricts access to admin users only (when security is enabled)
func AdminOnly(settingsService *services.SettingsService) fiber.Handler {
	return func(c fiber.Ctx) error {
		// Public routes exception
		if strings.HasSuffix(c.Path(), "/auth/login") {
			return c.Next()
		}

		// Check if security is enabled
		securityEnabled, _ := settingsService.Get(c.Context(), domain.SettingSecurityEnabled)
		if securityEnabled != "true" {
			// Security disabled - allow all requests
			return c.Next()
		}

		authType := c.Locals("auth_type")
		if authType != "admin" {
			return api.Error(c, api.CodeUnauthorized, "Admin access required")
		}
		return c.Next()
	}
}
