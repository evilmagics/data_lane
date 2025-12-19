package middleware

import (
	"os"
	"strings"

	"github.com/gofiber/fiber/v3"

	"pdf_generator/internal/core/services"
	"pdf_generator/pkg/api"
	"pdf_generator/pkg/utils"
)

// HMACMiddleware validates HMAC signature for POST/PUT requests
func HMACMiddleware(settingsService *services.SettingsService) fiber.Handler {
	return func(c fiber.Ctx) error {
		// Only validate for methods with body
		method := c.Method()
		if method != "POST" && method != "PUT" && method != "PATCH" {
			return c.Next()
		}

		// Check if HMAC is enabled in settings
		// Default to true if setting is missing or error
		enabledStr, _ := settingsService.Get(c.Context(), "enable_hmac")
		if enabledStr == "false" {
			return c.Next()
		}

		// Skip public routes
		if c.Path() == "/api/auth/login" || strings.HasSuffix(c.Path(), "/auth/login") {
			return c.Next()
		}

		// Skip if body is empty
		body := c.Body()
		if len(body) == 0 {
			return c.Next()
		}

		signature := c.Get("X-Signature")
		if signature == "" {
			return api.Error(c, api.CodeHMACMismatch, "Missing X-Signature header")
		}


		// Determine secret from credentials
		var secret string

		// 1. Try API Key
		if apiKey := c.Get("X-API-Key"); apiKey != "" {
			secret = apiKey
		}

		// 2. Try JWT Token (Authorization: Bearer <token>)
		if secret == "" {
			authHeader := c.Get("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				secret = strings.TrimPrefix(authHeader, "Bearer ")
			}
		}

		// 3. Fallback to global secret (for legacy/testing)
		if secret == "" {
			secret = os.Getenv("HMAC_SECRET")
			if secret == "" {
				secret = "default-hmac-secret"
			}
		}


		if !utils.VerifyHMAC(string(body), signature, secret) {
			return api.Error(c, api.CodeHMACMismatch, "Invalid signature")
		}

		return c.Next()
	}
}
