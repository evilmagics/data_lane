package handlers

import (
	"encoding/json"
	"strconv"

	"github.com/gofiber/fiber/v3"

	"pdf_generator/internal/core/domain"
	"pdf_generator/internal/core/services"
	"pdf_generator/pkg/api"
)

type GateHandler struct {
	service *services.GateService
}

func NewGateHandler(service *services.GateService) *GateHandler {
	return &GateHandler{service: service}
}

// List handles GET /gates
func (h *GateHandler) List(c fiber.Ctx) error {
	gates, err := h.service.List(c.Context())
	if err != nil {
		return api.Error(c, api.CodeInternalError, "Failed to fetch gates")
	}
	return api.Success(c, fiber.Map{"gates": gates})
}

// Create handles POST /gates (Single or Batch)
func (h *GateHandler) Create(c fiber.Ctx) error {
	body := c.Body()
	if len(body) == 0 {
		return api.Error(c, api.CodeInvalidRequest, "Empty body")
	}

	if body[0] == '[' {
		// Batch
		var gates []domain.Gate
		if err := json.Unmarshal(body, &gates); err != nil {
			return api.Error(c, api.CodeInvalidRequest, "Invalid JSON array")
		}
		// Validate IDs
		for _, g := range gates {
			if g.ID < 0 || g.ID > 100 {
				return api.Error(c, api.CodeValidationError, "Gate ID must be 0-100")
			}
		}

		if err := h.service.CreateBatch(c.Context(), gates); err != nil {
			return api.Error(c, api.CodeInternalError, err.Error())
		}
		return api.Success(c, fiber.Map{"count": len(gates)})
	} else {
		// Single
		var req struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		}
		if err := json.Unmarshal(body, &req); err != nil {
			return api.Error(c, api.CodeInvalidRequest, "Invalid JSON object")
		}
		if req.Name == "" {
			return api.Error(c, api.CodeValidationError, "Name is required")
		}
		if req.ID < 0 || req.ID > 100 {
			return api.Error(c, api.CodeValidationError, "Gate ID must be 0-100")
		}

		gate := &domain.Gate{
			ID:   req.ID,
			Name: req.Name,
		}
		if err := h.service.Create(c.Context(), gate); err != nil {
			return api.Error(c, api.CodeInternalError, err.Error())
		}
		return api.Success(c, gate)
	}
}

// UpdateBatch handles PUT /gates (Batch Update)
func (h *GateHandler) UpdateBatch(c fiber.Ctx) error {
	var gates []domain.Gate
	if err := c.Bind().JSON(&gates); err != nil {
		return api.Error(c, api.CodeInvalidRequest, "Invalid JSON")
	}

	for _, g := range gates {
		if g.ID < 0 || g.ID > 100 {
			return api.Error(c, api.CodeValidationError, "Gate ID must be 0-100")
		}
	}

	if err := h.service.BatchUpdate(c.Context(), gates); err != nil {
		return api.Error(c, api.CodeInternalError, err.Error())
	}

	return api.Success(c, fiber.Map{"count": len(gates)})
}

// UpdateSingle handles PUT /gates/:id
func (h *GateHandler) UpdateSingle(c fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id < 0 || id > 100 {
		return api.Error(c, api.CodeValidationError, "Invalid Gate ID")
	}

	var req struct {
		Name string `json:"name"`
	}
	if err := c.Bind().JSON(&req); err != nil {
		return api.Error(c, api.CodeInvalidRequest, "Invalid JSON")
	}
	if req.Name == "" {
		return api.Error(c, api.CodeValidationError, "Name is required")
	}
	if err := h.service.Update(c.Context(), id, req.Name); err != nil {
		return api.Error(c, api.CodeInternalError, err.Error())
	}
	return api.Success(c, fiber.Map{"id": id, "name": req.Name})
}

// Delete handles DELETE /gates (Batch) or DELETE /gates/:id (Single)
func (h *GateHandler) Delete(c fiber.Ctx) error {
	idStr := c.Params("id")
	if idStr != "" {
		// Single
		id, err := strconv.Atoi(idStr)
		if err != nil {
			return api.Error(c, api.CodeValidationError, "Invalid ID")
		}
		if err := h.service.Delete(c.Context(), id); err != nil {
			return api.Error(c, api.CodeInternalError, err.Error())
		}
		return api.Success(c, fiber.Map{"deleted": id})
	}

	// Batch
	var ids []int
	if err := c.Bind().JSON(&ids); err != nil {
		return api.Error(c, api.CodeInvalidRequest, "Invalid JSON array of IDs")
	}
	if err := h.service.BatchDelete(c.Context(), ids); err != nil {
		return api.Error(c, api.CodeInternalError, err.Error())
	}
	return api.Success(c, fiber.Map{"count": len(ids)})
}
