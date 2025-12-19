package handlers

import (
	"encoding/json"
	"strconv"

	"github.com/gofiber/fiber/v3"

	"pdf_generator/internal/core/domain"
	"pdf_generator/internal/core/services"
	"pdf_generator/pkg/api"
)

type StationHandler struct {
	service *services.StationService
}

func NewStationHandler(service *services.StationService) *StationHandler {
	return &StationHandler{service: service}
}

// List handles GET /stations
func (h *StationHandler) List(c fiber.Ctx) error {
	stations, err := h.service.List(c.Context())
	if err != nil {
		return api.Error(c, api.CodeInternalError, "Failed to fetch stations")
	}
	return api.Success(c, fiber.Map{"stations": stations})
}

// Create handles POST /stations (Single or Batch)
func (h *StationHandler) Create(c fiber.Ctx) error {
	body := c.Body()
	if len(body) == 0 {
		return api.Error(c, api.CodeInvalidRequest, "Empty body")
	}

	if body[0] == '[' {
		// Batch
		var stations []domain.Station
		if err := json.Unmarshal(body, &stations); err != nil {
			return api.Error(c, api.CodeInvalidRequest, "Invalid JSON array")
		}
        // Validate IDs
        for _, s := range stations {
            if s.ID < 0 || s.ID > 100 {
                return api.Error(c, api.CodeValidationError, "Station ID must be 0-100")
            }
        }

		if err := h.service.CreateBatch(c.Context(), stations); err != nil {
			return api.Error(c, api.CodeInternalError, err.Error())
		}
		return api.Success(c, fiber.Map{"count": len(stations)})
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
             return api.Error(c, api.CodeValidationError, "Station ID must be 0-100")
        }

		station := &domain.Station{
            ID: req.ID,
            Name: req.Name,
        }
		if err := h.service.Create(c.Context(), station); err != nil {
			return api.Error(c, api.CodeInternalError, err.Error())
		}
		return api.Success(c, station)
	}
}

// UpdateBatch handles PUT /stations (Batch Update)
func (h *StationHandler) UpdateBatch(c fiber.Ctx) error {
	var stations []domain.Station
	if err := c.Bind().JSON(&stations); err != nil {
		return api.Error(c, api.CodeInvalidRequest, "Invalid JSON")
	}
	
    for _, s := range stations {
        if s.ID < 0 || s.ID > 100 {
            return api.Error(c, api.CodeValidationError, "Station ID must be 0-100")
        }
    }

	if err := h.service.BatchUpdate(c.Context(), stations); err != nil {
		return api.Error(c, api.CodeInternalError, err.Error())
	}
	
	return api.Success(c, fiber.Map{"count": len(stations)})
}

// UpdateSingle handles PUT /stations/:id
func (h *StationHandler) UpdateSingle(c fiber.Ctx) error {
	idStr := c.Params("id")
    id, err := strconv.Atoi(idStr)
    if err != nil || id < 0 || id > 100 {
        return api.Error(c, api.CodeValidationError, "Invalid Station ID")
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
    // Update service usually expects ID. Assuming Service Update signature needs change too if it expects string.
    // Checking service usage: h.service.Update(ctx, id, name)
	if err := h.service.Update(c.Context(), id, req.Name); err != nil {
		return api.Error(c, api.CodeInternalError, err.Error())
	}
	return api.Success(c, fiber.Map{"id": id, "name": req.Name})
}

// Delete handles DELETE /stations (Batch) or DELETE /stations/:id (Single)
func (h *StationHandler) Delete(c fiber.Ctx) error {
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
