package main

import (
	"context"
	"flag"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/kardianos/service"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"pdf_generator/internal/adapters/repository"
	"pdf_generator/pkg/database"
	"pdf_generator/pkg/queue"
	"pdf_generator/pkg/version"
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
	log.Info().
		Str("version", version.Version).
		Str("build", version.Build).
		Msg("Service started")

	// Try loading env, but don't fail if not found (might use system envs)
	_ = godotenv.Load()

	// Initialize Logger
	zerolog.TimeFieldFormat = "02-01-2006 15:04:05"
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: "02-01-2006 15:04:05"})

	// Initialize Database (uses default path: data/app.db)
	// Ensure directories exist first
	for _, dir := range []string{"data", "output", "logs"} {
		os.MkdirAll(dir, 0755)
	}

	if err := database.InitDB(""); err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
		return
	}
	p.db = database.GetDB()

	// Initialize Repositories
	settingsRepo := repository.NewSettingsRepository(p.db)
	taskRepo := repository.NewTaskRepository(p.db)
	gateRepo := repository.NewGateRepository(p.db)

	// Initialize Queue
	q, err := queue.NewQueue(p.db, taskRepo, settingsRepo, gateRepo)
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
