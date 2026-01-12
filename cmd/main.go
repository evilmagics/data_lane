package main

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"

	"pdf_generator/internal/adapters/repository"
	"pdf_generator/internal/core/services"
	"pdf_generator/internal/server"
	"pdf_generator/pkg/database"
	"pdf_generator/pkg/logger"
	"pdf_generator/pkg/queue"
	"pdf_generator/pkg/version"
)

func main() {
	// Load .env
	if err := godotenv.Load(); err != nil {
		log.Warn().Msg(".env file not found, using environment variables")
	}

	// Ensure required directories exist
	ensureDirectories()

	// Initialize logger
	if err := logger.InitLogger("logs"); err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize logger")
	}
	defer logger.Close()

	log.Info().
		Str("version", version.Version).
		Str("build", version.Build).
		Msg("Starting PDF Generator Application")

	// Initialize database (uses default path: data/app.db)
	if err := database.InitDB(""); err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize database")
	}
	log.Info().Msg("Database initialized")

	db := database.GetDB()

	// Initialize repositories
	taskRepo := repository.NewTaskRepository(db)
	scheduleRepo := repository.NewScheduleRepository(db)
	settingsRepo := repository.NewSettingsRepository(db)
	sessionRepo := repository.NewSessionRepository(db)
	apiKeyRepo := repository.NewAPIKeyRepository(db)
	gateRepo := repository.NewGateRepository(db)

	// Dependency injection
	sessionExpiry := 12 * time.Hour
	if envExp := os.Getenv("SESSION_EXPIRY"); envExp != "" {
		if dur, err := time.ParseDuration(envExp); err == nil {
			sessionExpiry = dur
		} else if val, err := strconv.Atoi(envExp); err == nil {
			sessionExpiry = time.Duration(val) * time.Hour
		}
	}

	// Initialize services
	authService := services.NewAuthService(sessionRepo, settingsRepo, sessionExpiry)
	apiKeyService := services.NewAPIKeyService(apiKeyRepo)
	settingsService := services.NewSettingsService(settingsRepo)
	gateService := services.NewGateService(gateRepo)
	processService := services.NewProcessService(settingsService)

	// Initialize Queue
	taskQueue, err := queue.NewQueue(db, taskRepo, settingsRepo, gateRepo)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize queue")
	}

	// Initialize HTTP server
	srv := server.NewServer(
		authService,
		apiKeyService,
		settingsService,
		gateService,
		processService,
		taskRepo,
		scheduleRepo,
		taskQueue,
	)
	srv.SetupRoutes()

	// Get server address from environment
	serverAddr := getServerAddress()

	log.Info().Str("address", serverAddr).Msg("Starting HTTP server")
	if err := srv.Start(serverAddr); err != nil {
		log.Fatal().Err(err).Msg("Server failed")
	}
}

// getServerAddress returns the server address from SERVER_URL env variable.
// Supports formats: "host:port", "http://host:port", or ":port"
// Default: "localhost:3110"
func getServerAddress() string {
	serverURL := os.Getenv("SERVER_URL")
	if serverURL == "" {
		return "localhost:3110"
	}

	// If it looks like a URL with scheme, parse it
	if len(serverURL) > 7 && (serverURL[:7] == "http://" || serverURL[:8] == "https://") {
		// Extract host:port from URL
		start := 7
		if serverURL[:8] == "https://" {
			start = 8
		}
		hostPort := serverURL[start:]
		// Remove any path
		if idx := len(hostPort); idx > 0 {
			for i, c := range hostPort {
				if c == '/' {
					hostPort = hostPort[:i]
					break
				}
			}
		}
		return hostPort
	}

	return serverURL
}

// ensureDirectories creates required application directories if they don't exist.
// Called once at startup before any other initialization.
func ensureDirectories() {
	dirs := []string{"data", "output", "logs"}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Fatal().Err(err).Str("dir", dir).Msg("Failed to create directory")
		}
	}
}
