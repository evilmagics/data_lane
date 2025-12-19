package handlers

import (
	"pdf_generator/internal/core/services"

	"github.com/gofiber/fiber/v3"
)

type ServiceHandler struct {
	processService *services.ProcessService
}

func NewServiceHandler(processService *services.ProcessService) *ServiceHandler {
	return &ServiceHandler{
		processService: processService,
	}
}

func (h *ServiceHandler) GetStatus(c fiber.Ctx) error {
	status, err := h.processService.GetStatus(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.JSON(fiber.Map{
		"service": "pdf-generator",
		"status":  status,
	})
}

func (h *ServiceHandler) Control(c fiber.Ctx) error {
	var req struct {
		Action string `json:"action"`
	}
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	var err error
	switch req.Action {
	case "start":
		err = h.processService.Start(c.Context())
	case "stop":
		err = h.processService.Stop(c.Context())
	case "restart":
		err = h.processService.Restart(c.Context())
	case "install":
		err = h.processService.Install(c.Context())
	case "uninstall":
		err = h.processService.Uninstall(c.Context())
	default:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid action. Allowed: start, stop, restart, install, uninstall",
		})
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Action executed successfully",
		"action":  req.Action,
	})
}
