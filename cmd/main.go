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
)

func main() {
	// Load .env
	if err := godotenv.Load(); err != nil {
		log.Warn().Msg(".env file not found, using environment variables")
	}

	// Initialize logger
	if err := logger.InitLogger("logs"); err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize logger")
	}
	defer logger.Close()

	log.Info().Msg("Starting PDF Generator Application")

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

	// Get server URL from environment
	serverUrl := os.Getenv("SERVER_URL")
	if serverUrl == "" {
		serverUrl = "http://localhost:3000"
	}

	log.Info().Str("serverUrl", serverUrl).Msg("Starting HTTP server")
	if err := srv.Start(serverUrl); err != nil {
		log.Fatal().Err(err).Msg("Server failed")
	}
}
