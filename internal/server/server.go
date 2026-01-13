package server

import (
	"context"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/fiber/v3/middleware/static"
	"github.com/rs/zerolog/log"

	"pdf_generator/internal/adapters/handlers"
	"pdf_generator/internal/adapters/middleware"
	"pdf_generator/internal/assets"
	"pdf_generator/internal/core/ports"
	"pdf_generator/internal/core/services"
)

// Server holds dependencies for the HTTP server
type Server struct {
	app             *fiber.App
	authService     *services.AuthService
	apiKeyService   *services.APIKeyService
	settingsService *services.SettingsService
	gateService     *services.GateService
	processService  *services.ProcessService
	taskRepo        ports.TaskRepository
	scheduleRepo    ports.ScheduleRepository
	queue           ports.QueueService
}

// NewServer creates a new HTTP server
func NewServer(
	authService *services.AuthService,
	apiKeyService *services.APIKeyService,
	settingsService *services.SettingsService,
	gateService *services.GateService,
	processService *services.ProcessService,
	taskRepo ports.TaskRepository,
	scheduleRepo ports.ScheduleRepository,
	queue ports.QueueService,
) *Server {
	app := fiber.New(fiber.Config{
		AppName: "PDF Generator",
	})

	// Global middleware
	app.Use(recover.New())
	app.Use(cors.New())
	app.Use(middleware.LoggerMiddleware())

	return &Server{
		app:             app,
		authService:     authService,
		apiKeyService:   apiKeyService,
		settingsService: settingsService,
		gateService:     gateService,
		processService:  processService,
		taskRepo:        taskRepo,
		scheduleRepo:    scheduleRepo,
		queue:           queue,
	}
}

// SetupRoutes configures all routes
func (s *Server) SetupRoutes() {
	// Handlers
	authHandler := handlers.NewAuthHandler(s.authService)
	settingsHandler := handlers.NewSettingsHandler(s.settingsService)
	apiKeyHandler := handlers.NewAPIKeyHandler(s.apiKeyService)
	gateHandler := handlers.NewGateHandler(s.gateService)
	taskHandler := handlers.NewTaskHandler(s.taskRepo, s.settingsService.GetRepo(), s.queue)
	scheduleHandler := handlers.NewScheduleHandler(s.scheduleRepo)
	sseHandler := handlers.NewSSEHandler(s.taskRepo, s.queue)

	// API group
	api := s.app.Group("/api")

	// Public routes
	api.Post("/auth/login", authHandler.Login)

	// Protected routes (Admin + API Key)
	protected := api.Group("", middleware.AuthMiddleware(s.authService, s.apiKeyService, s.settingsService))

	// HMAC protected routes (POST/PUT with body)
	hmacProtected := protected.Group("", middleware.HMACMiddleware(s.settingsService))

	// Queue/Tasks (Shared)
	hmacProtected.Post("/queue", taskHandler.Enqueue)
	protected.Get("/tasks", taskHandler.List)
	protected.Get("/tasks/:id", taskHandler.Get)
	protected.Delete("/tasks/:id", taskHandler.Cancel)
	protected.Get("/tasks/:id/download", taskHandler.Download)

	// Transaction Statuses (Public - for dropdown options)
	protected.Get("/transaction-statuses", func(c fiber.Ctx) error {
		statuses := []string{"PERIODIK", "BUKA ALB"}
		return c.JSON(fiber.Map{
			"success": true,
			"data":    fiber.Map{"statuses": statuses},
		})
	})

	// Schedules (Shared)
	hmacProtected.Post("/schedules", scheduleHandler.Create)
	protected.Get("/schedules", scheduleHandler.List)
	protected.Delete("/schedules/:id", scheduleHandler.Delete)

	// Gates (Protected)
	protected.Get("/gates", gateHandler.List)
	hmacProtected.Post("/gates", gateHandler.Create)
	hmacProtected.Put("/gates", gateHandler.UpdateBatch)
	hmacProtected.Put("/gates/:id", gateHandler.UpdateSingle)
	hmacProtected.Delete("/gates", gateHandler.Delete) // Delete batch
	protected.Delete("/gates/:id", gateHandler.Delete) // Delete single
	
	// SSE (Task-specific is Shared)
	protected.Get("/sse/tasks/:id", sseHandler.TaskEvents)

	// Admin-only routes
	admin := protected.Group("", middleware.AdminOnly(s.settingsService))
	hmacAdmin := admin.Group("", middleware.HMACMiddleware(s.settingsService))

	// Auth (Admin)
	admin.Post("/auth/logout", authHandler.Logout)
	admin.Get("/sessions", authHandler.ListSessions)
	admin.Delete("/sessions/:id", authHandler.RevokeSession)

	// Settings (Admin)
	admin.Get("/settings", settingsHandler.GetAll)
	hmacAdmin.Put("/settings", settingsHandler.Update)

	// API Keys (Admin)
	admin.Get("/api-keys", apiKeyHandler.List)
	hmacAdmin.Post("/api-keys", apiKeyHandler.Create)
	admin.Get("/api-keys/:id/show", apiKeyHandler.Show)
	admin.Put("/api-keys/:id/toggle", apiKeyHandler.Toggle)
	admin.Delete("/api-keys/:id", apiKeyHandler.Delete)

	// SSE Global (Admin)
	admin.Get("/sse/events", sseHandler.GlobalEvents)

	// Service Control (Admin)
	serviceHandler := handlers.NewServiceHandler(s.processService)
	admin.Get("/system/services/pdf-generator/status", serviceHandler.GetStatus)
	admin.Post("/system/services/pdf-generator/control", serviceHandler.Control)

	// Static Assets (Embedded UI)
	assetFS := assets.GetEmbeddedAssets()

	// Serve config.js dynamically with environment variables
	s.app.Get("/config.js", func(c fiber.Ctx) error {
		// Create a simple JS file that sets window.APP_CONFIG
		// Note: HMAC_SECRET is no longer exposed. UI uses JWT/API Key as secret.
		enabledStr, _ := s.settingsService.Get(c.Context(), "enable_hmac")
		enableHmac := enabledStr != "false"
		// We can inject boolean directly: ENABLE_HMAC: true/false
		js := "window.APP_CONFIG = { API_BASE_URL: '/api', ENABLE_HMAC: " + map[bool]string{true: "true", false: "false"}[enableHmac] + " };"
		c.Set("Content-Type", "application/javascript")
		return c.SendString(js)
	})

	// Serve the static files (for _next, assets, etc.)
	s.app.Use("/", static.New("", static.Config{
		FS:     assetFS,
		Browse: false,
	}))

	// HTML route handler for SPA with static export
	s.app.Get("*", func(c fiber.Ctx) error {
		path := c.Path()

		// Skip API and SSE routes
		if strings.HasPrefix(path, "/api") || strings.HasPrefix(path, "/sse") {
			return c.Next()
		}

		// Try to serve the correct HTML file for the route
		// 1. Try {path}.html (e.g., /dashboard -> dashboard.html)
		// 2. Try {path}/index.html
		// 3. Fall back to index.html

		var htmlPath string
		if path == "/" {
			htmlPath = "index.html"
		} else {
			// Remove leading slash and try as .html file
			cleanPath := strings.TrimPrefix(path, "/")
			htmlPath = cleanPath + ".html"
		}

		// Try to open the specific HTML file
		f, err := assetFS.Open(htmlPath)
		if err != nil {
			// Try index.html in subdirectory
			if path != "/" {
				cleanPath := strings.TrimPrefix(path, "/")
				f, err = assetFS.Open(cleanPath + "/index.html")
			}
		}

		// Final fallback to root index.html
		if err != nil {
			f, err = assetFS.Open("index.html")
			if err != nil {
				return c.SendStatus(404)
			}
		}

		c.Set("Content-Type", "text/html")
		return c.SendStream(f)
	})
}

// Start runs the HTTP server
func (s *Server) Start(addr string) error {
	// Initialize settings cache
	if err := s.settingsService.LoadCache(context.Background()); err != nil {
		return err
	}

	// Start service monitoring
	if s.processService != nil {
		// Auto-start worker service
		if err := s.processService.EnsureRunning(context.Background()); err != nil {
			log.Error().Err(err).Msg("Failed to auto-start worker service")
		}
		s.processService.StartMonitoring(context.Background())
	}

	return s.app.Listen(addr, fiber.ListenConfig{
		DisableStartupMessage: true,
	})
}

// GetApp returns the Fiber app instance
func (s *Server) GetApp() *fiber.App {
	return s.app
}
