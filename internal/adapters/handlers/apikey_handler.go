package handlers

import (
	"github.com/gofiber/fiber/v3"

	"pdf_generator/internal/core/services"
	"pdf_generator/pkg/api"
)

// APIKeyHandler handles API key endpoints
type APIKeyHandler struct {
	apiKeyService *services.APIKeyService
}

// NewAPIKeyHandler creates a new API key handler
func NewAPIKeyHandler(apiKeyService *services.APIKeyService) *APIKeyHandler {
	return &APIKeyHandler{apiKeyService: apiKeyService}
}

// CreateAPIKeyRequest represents API key creation
type CreateAPIKeyRequest struct {
	Name string `json:"name"`
}

// List handles GET /api-keys
func (h *APIKeyHandler) List(c fiber.Ctx) error {
	keys, err := h.apiKeyService.List(c.Context())
	if err != nil {
		return api.Error(c, api.CodeInternalError, "Failed to list API keys")
	}
	// Don't expose key hash or encrypted key
	items := make([]fiber.Map, len(keys))
	for i, k := range keys {
		items[i] = fiber.Map{
			"id":         k.ID,
			"name":       k.Name,
			"active":     k.Active,
			"created_at": k.CreatedAt,
		}
	}
	return api.Success(c, fiber.Map{"items": items})
}

// Create handles POST /api-keys
func (h *APIKeyHandler) Create(c fiber.Ctx) error {
	var req CreateAPIKeyRequest
	if err := c.Bind().JSON(&req); err != nil {
		return api.Error(c, api.CodeInvalidRequest, "Invalid request body")
	}

	if req.Name == "" {
		return api.Error(c, api.CodeValidationError, "Name is required")
	}

	key, rawKey, err := h.apiKeyService.Create(c.Context(), req.Name)
	if err != nil {
		return api.Error(c, api.CodeInternalError, "Failed to create API key")
	}

	return api.Success(c, fiber.Map{
		"id":      key.ID,
		"name":    key.Name,
		"api_key": rawKey,
		"active":  key.Active,
	})
}

// Show handles GET /api-keys/:id/show
func (h *APIKeyHandler) Show(c fiber.Ctx) error {
	id := c.Params("id")
	rawKey, err := h.apiKeyService.Reveal(c.Context(), id)
	if err != nil {
		return api.Error(c, api.CodeNotFound, "API key not found")
	}
	return api.Success(c, fiber.Map{"id": id, "api_key": rawKey})
}

// Toggle handles PUT /api-keys/:id/toggle
func (h *APIKeyHandler) Toggle(c fiber.Ctx) error {
	id := c.Params("id")
	key, err := h.apiKeyService.Toggle(c.Context(), id)
	if err != nil {
		return api.Error(c, api.CodeNotFound, "API key not found")
	}
	return api.Success(c, fiber.Map{"id": key.ID, "active": key.Active})
}

// Delete handles DELETE /api-keys/:id
func (h *APIKeyHandler) Delete(c fiber.Ctx) error {
	id := c.Params("id")
	if err := h.apiKeyService.Delete(c.Context(), id); err != nil {
		return api.Error(c, api.CodeNotFound, "API key not found")
	}
	return api.Success(c, nil)
}
