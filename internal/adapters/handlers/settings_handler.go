package handlers

import (
	"github.com/gofiber/fiber/v3"

	"pdf_generator/internal/core/services"
	"pdf_generator/pkg/api"
)

// SettingsHandler handles settings endpoints
type SettingsHandler struct {
	settingsService *services.SettingsService
}

// NewSettingsHandler creates a new settings handler
func NewSettingsHandler(settingsService *services.SettingsService) *SettingsHandler {
	return &SettingsHandler{settingsService: settingsService}
}

// UpdateSettingRequest represents setting update
type UpdateSettingRequest struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// GetAll handles GET /settings
func (h *SettingsHandler) GetAll(c fiber.Ctx) error {
	settings, err := h.settingsService.GetAll(c.Context())
	if err != nil {
		return api.Error(c, api.CodeInternalError, "Failed to get settings")
	}
	return api.Success(c, fiber.Map{"settings": settings})
}

// Update handles PUT /settings
func (h *SettingsHandler) Update(c fiber.Ctx) error {
	var req UpdateSettingRequest
	if err := c.Bind().JSON(&req); err != nil {
		return api.Error(c, api.CodeInvalidRequest, "Invalid request body")
	}

	if req.Key == "" {
		return api.Error(c, api.CodeValidationError, "Key is required")
	}

	if err := h.settingsService.Set(c.Context(), req.Key, req.Value); err != nil {
		return api.Error(c, api.CodeInternalError, "Failed to update setting")
	}

	return api.Success(c, fiber.Map{"key": req.Key, "value": req.Value})
}
