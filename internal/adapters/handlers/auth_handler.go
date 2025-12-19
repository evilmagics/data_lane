package handlers

import (
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog/log"

	"pdf_generator/internal/core/services"
	"pdf_generator/pkg/api"
	"pdf_generator/pkg/utils"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	authService *services.AuthService
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// LoginRequest represents login credentials
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse represents login result
type LoginResponse struct {
	Token     string `json:"token"`
	ExpiresAt string `json:"expires_at"`
	SessionID string `json:"session_id"`
}

// Login handles POST /auth/login
func (h *AuthHandler) Login(c fiber.Ctx) error {
	var req LoginRequest
	if err := c.Bind().JSON(&req); err != nil {
		return api.Error(c, api.CodeInvalidRequest, "Invalid request body")
	}

	if req.Username == "" || req.Password == "" {
		return api.Error(c, api.CodeValidationError, "Username and password required")
	}

	token, expiresAt, sessionID, err := h.authService.Login(
		c.Context(),
		req.Username,
		req.Password,
		c.IP(),
		c.Get("User-Agent"),
	)
	if err != nil {
		if err.Error() == "session limit reached" {
			return api.Error(c, api.CodeSessionLimitReached, err.Error())
		}
		return api.Error(c, api.CodeInvalidCredentials, "Invalid credentials")
	}

	return api.Success(c, LoginResponse{
		Token:     token,
		ExpiresAt: expiresAt.Format("2006-01-02T15:04:05Z07:00"),
		SessionID: sessionID,
	})
}

// Logout handles POST /auth/logout
func (h *AuthHandler) Logout(c fiber.Ctx) error {
	token := c.Get("Authorization")
	// Robust token extraction
	if len(token) > 0 {
		token = strings.TrimPrefix(token, "Bearer ")
		token = strings.TrimSpace(token)
	}

	tokenHash := utils.GetTokenSignature(token)
	log.Debug().Msgf("DEBUG LOGOUT: Original Header: %s | Extracted Token: %s | Hash: %s", c.Get("Authorization"), token, tokenHash)

	if err := h.authService.Logout(c.Context(), tokenHash); err != nil {
		log.Error().Err(err).Msg("DEBUG LOGOUT FAILED")
		return api.Error(c, api.CodeInternalError, "Failed to logout")
	}

	return api.Success(c, nil)
}

// ListSessions handles GET /sessions
func (h *AuthHandler) ListSessions(c fiber.Ctx) error {
	sessions, err := h.authService.ListActiveSessions(c.Context())
	if err != nil {
		return api.Error(c, api.CodeInternalError, "Failed to list sessions")
	}
	return api.Success(c, fiber.Map{"items": sessions})
}

// RevokeSession handles DELETE /sessions/:id
func (h *AuthHandler) RevokeSession(c fiber.Ctx) error {
	id := c.Params("id")
	if err := h.authService.RevokeSession(c.Context(), id); err != nil {
		return api.Error(c, api.CodeNotFound, "Session not found")
	}
	return api.Success(c, nil)
}
