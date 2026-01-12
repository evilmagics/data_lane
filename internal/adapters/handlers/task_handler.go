package handlers

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog/log"

	"pdf_generator/internal/core/domain"
	"pdf_generator/internal/core/ports"
	"pdf_generator/pkg/api"
	"pdf_generator/pkg/datasource"
)

// TaskHandler handles task endpoints
type TaskHandler struct {
	taskRepo     ports.TaskRepository
	settingsRepo ports.SettingsRepository
	queue        ports.QueueService
}

// NewTaskHandler creates a new task handler
func NewTaskHandler(taskRepo ports.TaskRepository, settingsRepo ports.SettingsRepository, queue ports.QueueService) *TaskHandler {
	return &TaskHandler{
		taskRepo:     taskRepo,
		settingsRepo: settingsRepo,
		queue:        queue,
	}
}

// EnqueueRequest represents a task enqueue request
type EnqueueRequest struct {
	RootFolder string            `json:"root_folder"`
	BranchID   int               `json:"branch_id"`
	GateID     int               `json:"gate_id"`
	StationID  int               `json:"station_id"`
	Filter     domain.TaskFilter `json:"filter"`
	Settings   map[string]string `json:"settings"`
}

// List handles GET /tasks
func (h *TaskHandler) List(c fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	if limit > 100 {
		limit = 100
	}

	var gateID *int
	if sid := c.Query("gate_id"); sid != "" {
		if id, err := strconv.Atoi(sid); err == nil {
			gateID = &id
		}
	}

	var stationID *int
	if sid := c.Query("station_id"); sid != "" {
		if id, err := strconv.Atoi(sid); err == nil {
			stationID = &id
		}
	}

	filter := ports.TaskFilter{
		Status:    c.Query("status"),
		FromDate:  c.Query("from"),
		ToDate:    c.Query("to"),
		GateID:    gateID,
		StationID: stationID,
		Page:      page,
		Limit:     limit,
	}

	tasks, total, err := h.taskRepo.List(c.Context(), filter)
	if err != nil {
		return api.Error(c, api.CodeInternalError, "Failed to list tasks")
	}

	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	return api.Success(c, api.PaginatedResponse{
		Items: tasks,
		Pagination: api.Pagination{
			Total:      total,
			Page:       page,
			Limit:      limit,
			TotalPages: totalPages,
		},
	})
}

// Get handles GET /tasks/:id
func (h *TaskHandler) Get(c fiber.Ctx) error {
	id := c.Params("id")
	task, err := h.taskRepo.GetByID(c.Context(), id)
	if err != nil {
		return api.Error(c, api.CodeNotFound, "Task not found")
	}
	return api.Success(c, task)
}

// Cancel handles DELETE /tasks/:id
func (h *TaskHandler) Cancel(c fiber.Ctx) error {
	id := c.Params("id")
	task, err := h.taskRepo.GetByID(c.Context(), id)
	if err != nil {
		return api.Error(c, api.CodeNotFound, "Task not found")
	}

	if task.Status != domain.TaskStatusQueued && task.Status != domain.TaskStatusPending {
		return api.Error(c, api.CodeValidationError, "Cannot cancel task in current status")
	}

	task.Status = domain.TaskStatusCancelled
	if err := h.taskRepo.Update(c.Context(), task); err != nil {
		return api.Error(c, api.CodeInternalError, "Failed to cancel task")
	}

	return api.Success(c, fiber.Map{"id": task.ID, "status": task.Status})
}

// Download handles GET /tasks/:id/download
func (h *TaskHandler) Download(c fiber.Ctx) error {
	id := c.Params("id")
	task, err := h.taskRepo.GetByID(c.Context(), id)
	if err != nil {
		return api.Error(c, api.CodeNotFound, "Task not found")
	}

	if task.Status != domain.TaskStatusCompleted {
		return api.Error(c, api.CodeTaskNotReady, "Task not ready for download")
	}

	if task.OutputFilePath == "" {
		return api.Error(c, api.CodeNotFound, "Output file not found")
	}

	// Serve file
	filePath := task.OutputFilePath
	if !filepath.IsAbs(filePath) {
		filePath = filepath.Join("output", filePath)
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return api.Error(c, api.CodeNotFound, "Output file not found")
	}

	return c.Download(filePath)
}

// Enqueue handles POST /queue - creates a new task
func (h *TaskHandler) Enqueue(c fiber.Ctx) error {
	var req EnqueueRequest
	if err := c.Bind().JSON(&req); err != nil {
		return api.Error(c, api.CodeInvalidRequest, "Invalid request body")
	}

	// Validate IDs (0-100)
	if req.BranchID < 0 || req.BranchID > 100 {
		return api.Error(c, api.CodeValidationError, "Branch ID must be between 0 and 100")
	}
	if req.GateID < 0 || req.GateID > 100 {
		return api.Error(c, api.CodeValidationError, "Gate ID must be between 0 and 100")
	}
	if req.StationID < 0 || req.StationID > 100 {
		return api.Error(c, api.CodeValidationError, "Station ID must be between 0 and 100")
	}

	// Normalize root folder path based on OS
	normalizedRoot := req.RootFolder
	if runtime.GOOS == "windows" {
		normalizedRoot = filepath.FromSlash(req.RootFolder)
	} else {
		normalizedRoot = filepath.ToSlash(req.RootFolder)
	}

	// Check datasource from filepath while adding new task
	var targetDate time.Time
	if req.Filter.Date != "" {
		targetDate, _ = time.Parse("2006-01-02", req.Filter.Date)
	} else if req.Filter.RangeStart != "" {
		targetDate, _ = time.Parse("2006-01-02", req.Filter.RangeStart)
	} else {
		targetDate = time.Now()
	}

	// Fetch datasource path format from settings
	datasourceFormat := "{MM}{YY}/{StationID}/{DD}{MM}{YYYY}.mdb"
	if setting, err := h.settingsRepo.Get(c.Context(), domain.SettingDataSourcePathFormat); err == nil && setting != nil {
		datasourceFormat = setting.Value
	}

	dbPath := datasource.GetDataSourcePath(datasourceFormat, normalizedRoot, targetDate, req.BranchID, req.GateID, req.StationID)
	dbPath = filepath.FromSlash(dbPath)
	if _, err := os.Stat(dbPath); err != nil {
		log.Info().Str("path", dbPath).Msg("Datasource file not found while adding new task")
	} else {
		log.Info().Str("path", dbPath).Msg("Datasource file found while adding new task")
	}

	// Create task metadata for queue
	metadata := domain.TaskMetadata{
		RootFolder: normalizedRoot,
		BranchID:   req.BranchID,
		GateID:     req.GateID,
		StationID:  req.StationID,
		Filter:     req.Filter,
		Settings:   req.Settings,
	}
	filterJSON, _ := json.Marshal(req.Filter)
	settingsJSON, _ := json.Marshal(req.Settings)

	task := &domain.Task{
		Status:       domain.TaskStatusQueued,
		RootFolder:   normalizedRoot,
		BranchID:     req.BranchID,
		GateID:       req.GateID,
		StationID:    req.StationID,
		FilterJSON:   string(filterJSON),
		SettingsJSON: string(settingsJSON),
	}

	if err := h.taskRepo.Create(c.Context(), task); err != nil {
		return api.Error(c, api.CodeInternalError, "Failed to create task")
	}

	// Process queue asynchronously if queue is available
	if h.queue != nil {
		if _, err := h.queue.Enqueue(c.Context(), task.ID, metadata); err != nil {
			// If enqueue fails, should we fail the request? Or just log?
			// Usually we want to ensure it's enqueued.
			// Revert task creation or mark as failed?
			// For now, let's log error and return 500.
			task.Status = domain.TaskStatusFailed
			h.taskRepo.Update(c.Context(), task)
			return api.Error(c, api.CodeInternalError, "Failed to enqueue task: "+err.Error())
		}
	}

	position, _ := h.taskRepo.GetQueuePosition(c.Context(), task.ID)
	queueSize, _ := h.taskRepo.CountByStatus(c.Context(), domain.TaskStatusQueued)

	return api.Success(c, fiber.Map{
		"task_id":        task.ID,
		"status":         task.Status,
		"queue_position": position,
		"queue_size":     queueSize,
		"created_at":     task.CreatedAt,
	})
}
