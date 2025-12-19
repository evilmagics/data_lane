package handlers

import (
	"encoding/json"

	"github.com/gofiber/fiber/v3"

	"pdf_generator/internal/core/domain"
	"pdf_generator/internal/core/ports"
	"pdf_generator/pkg/api"
)

// ScheduleHandler handles schedule endpoints
type ScheduleHandler struct {
	scheduleRepo ports.ScheduleRepository
}

// NewScheduleHandler creates a new schedule handler
func NewScheduleHandler(scheduleRepo ports.ScheduleRepository) *ScheduleHandler {
	return &ScheduleHandler{scheduleRepo: scheduleRepo}
}

// CreateScheduleRequest represents schedule creation
type CreateScheduleRequest struct {
	Cron        string                 `json:"cron"`
	TaskPayload domain.TaskMetadata    `json:"task_payload"`
}

// List handles GET /schedules
func (h *ScheduleHandler) List(c fiber.Ctx) error {
	schedules, err := h.scheduleRepo.List(c.Context())
	if err != nil {
		return api.Error(c, api.CodeInternalError, "Failed to list schedules")
	}
	return api.Success(c, fiber.Map{"items": schedules})
}

// Create handles POST /schedules
func (h *ScheduleHandler) Create(c fiber.Ctx) error {
	var req CreateScheduleRequest
	if err := c.Bind().JSON(&req); err != nil {
		return api.Error(c, api.CodeInvalidRequest, "Invalid request body")
	}

	if req.Cron == "" {
		return api.Error(c, api.CodeValidationError, "Cron expression required")
	}

	payloadJSON, _ := json.Marshal(req.TaskPayload)

	schedule := &domain.Schedule{
		Cron:        req.Cron,
		TaskPayload: string(payloadJSON),
		Active:      true,
	}

	if err := h.scheduleRepo.Create(c.Context(), schedule); err != nil {
		return api.Error(c, api.CodeInternalError, "Failed to create schedule")
	}

	return api.Success(c, fiber.Map{
		"schedule_id": schedule.ID,
		"cron":        schedule.Cron,
		"next_run":    schedule.NextRun,
		"active":      schedule.Active,
	})
}

// Delete handles DELETE /schedules/:id
func (h *ScheduleHandler) Delete(c fiber.Ctx) error {
	id := c.Params("id")
	if err := h.scheduleRepo.Delete(c.Context(), id); err != nil {
		return api.Error(c, api.CodeNotFound, "Schedule not found")
	}
	return api.Success(c, nil)
}
