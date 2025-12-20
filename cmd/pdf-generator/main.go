package main

import (
	"context"
	"flag"
	"os"
	"path/filepath"
	"time"

	"github.com/joho/godotenv"
	"github.com/kardianos/service"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"pdf_generator/internal/adapters/repository"
	"pdf_generator/pkg/database"
	"pdf_generator/pkg/queue"
)

type program struct {
	exit    chan struct{}
	service service.Service
	db      *gorm.DB
	queue   *queue.Queue
}

func (p *program) Start(s service.Service) error {
	p.exit = make(chan struct{})
	
	// Start should not block. Do the actual work async.
	go p.run()
	return nil
}

func (p *program) Stop(s service.Service) error {
	// Stop the queue or any other cleanup
	log.Info().Msg("Stopping service...")
	close(p.exit)
	return nil
}

func (p *program) run() {
	log.Info().Msg("Service started")
	
	// Load environment variables
	envPath := ".env"
	// adjust path if running as service in different dir?
	// Usually service runs with CWD as executable dir or system dir.
	// Best to try loading from executable directory
	ex, err := os.Executable()
	if err == nil {
		exPath := filepath.Dir(ex)
		// CRITICAL: Change CWD to the executable directory
		// when running as a service, CWD is often C:\Windows\system32
		if err := os.Chdir(exPath); err != nil {
			log.Error().Err(err).Msg("Failed to change working directory to executable path")
		} else {
            log.Info().Str("path", exPath).Msg("Changed working directory")
        }

		if _, err := os.Stat(filepath.Join(exPath, ".env")); err == nil {
			envPath = filepath.Join(exPath, ".env")
		}
	}
	
	// Try loading env, but don't fail if not found (might use system envs)
	_ = godotenv.Load(envPath)

	// Initialize Logger
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339})
	
	// Initialize Database
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "data/app.db"
	}
	// Also ensure data dir exists if path contains directory separator?
	// InitDB handles opening, but maybe ensure directory?
	// InitDB likely handles it or sqlite driver errors.

	if err := database.InitDB(dbPath); err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
		return
	}
	p.db = database.GetDB()

	// Initialize Repositories
	settingsRepo := repository.NewSettingsRepository(p.db)
	taskRepo := repository.NewTaskRepository(p.db)

	// Initialize Queue
	q, err := queue.NewQueue(p.db, taskRepo, settingsRepo)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize queue")
		return
	}
	p.queue = q

	// Register Consumer (This is the key difference for the worker service)
	q.RegisterConsumers()

	// Start Queue
	q.Start(context.Background())
	log.Info().Msg("Queue consumers started")

	// Helper ticker to keep the service loop alive and log status if needed
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-p.exit:
			return
		case <-ticker.C:
			// Just a heartbeat or check specific things
			// Could be used to update a "last seen" timestamp in DB if desired
		}
	}
}

func main() {
	svcFlag := flag.String("service", "", "Control the system service.")
	flag.Parse()

	options := make(service.KeyValue)
	options["Restart"] = "on-success"
	options["SuccessExitStatus"] = "1 2 8 42"

	svcConfig := &service.Config{
		Name:        "DataLanePDFWorker",
		DisplayName: "DataLane PDF Worker Service",
		Description: "Background service for generating PDFs for DataLane application.",
		Dependencies: []string{
			"DataLane", // Dependent on main service? No, probably independent or depends on DB
		},
		Option: options,
	}

	prg := &program{}
	s, err := service.New(prg, svcConfig)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create service interface")
	}
	prg.service = s

	if len(*svcFlag) != 0 {
		err := service.Control(s, *svcFlag)
		if err != nil {
			log.Printf("Valid actions: %q\n", service.ControlAction)
			log.Fatal().Err(err).Msg("Service control failed")
		}
		return
	}

	err = s.Run()
	if err != nil {
		log.Error().Err(err).Msg("Service failed to run")
	}
}
