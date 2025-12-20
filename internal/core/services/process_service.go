package services

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"sync/atomic"

	"github.com/kardianos/service"
	"github.com/rs/zerolog/log"

	"pdf_generator/internal/core/domain"
)

type ProcessService struct {
	settingsService *SettingsService
	serviceConfig   *service.Config
	cachedStatus    atomic.Value // Stores domain.ServiceStatus
}

func NewProcessService(settingsService *SettingsService) *ProcessService {
	// Determine binary name based on OS
	binaryName := "pdf-generator"
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}

	// Look for binary in current dir or executable dir
	execPath := ""
	ex, err := os.Executable()
	if err == nil {
		exDir := filepath.Dir(ex)
		// Check typical locations
		if _, err := os.Stat(filepath.Join(exDir, binaryName)); err == nil {
			execPath = filepath.Join(exDir, binaryName)
		} else if _, err := os.Stat(filepath.Join(exDir, "output", binaryName)); err == nil {
			// Dev mode: output folder relative to executable?
			execPath = filepath.Join(exDir, "output", binaryName)
		} else if _, err := os.Stat(binaryName); err == nil {
			// CWD
			execPath, _ = filepath.Abs(binaryName)
		}
	}

	if execPath == "" {
		log.Warn().Msgf("Could not locate %s binary. Auto-install might fail.", binaryName)
	}

	// Config must match the one in cmd/pdf-generator/main.go
	svcConfig := &service.Config{
		Name:        "DataLanePDFWorker",
		DisplayName: "DataLane PDF Worker Service",
		Description: "Background service for generating PDFs for DataLane application.",
		Executable:  execPath,
		Arguments:   []string{"-service", "run"}, // Explicitly tell it to run as service? Or default?
		// Note: kardianos/service usually invokes the binary with no args or specific platform args.
		// If we set Arguments, they are passed.
		// Our pdf-generator main parses flags. But service.Run() handles the rest.
		// To be safe, we usually don't need args if it detects it's a service, 
		// but checking main.go: `svcFlag := flag.String("service", "", ...)`
		// If we don't pass -service, it enters `s.Run()`.
		// If `s.Run()` is called interactively, it tries to run.
		// Ideally, we don't need arguments if we are just running it.
	}

	s := &ProcessService{
		settingsService: settingsService,
		serviceConfig:   svcConfig,
	}
	s.cachedStatus.Store(domain.ServiceStatusUnknown)
	return s
}

func (s *ProcessService) getService() (service.Service, error) {
    // We pass nil as the program because we are only controlling it, not running it here
    return service.New(nil, s.serviceConfig)
}

func (s *ProcessService) Start(ctx context.Context) error {
    svc, err := s.getService()
    if err != nil {
        return err
    }
    return svc.Start()
}

func (s *ProcessService) Stop(ctx context.Context) error {
    svc, err := s.getService()
    if err != nil {
        return err
    }
    return svc.Stop()
}

func (s *ProcessService) Restart(ctx context.Context) error {
    svc, err := s.getService()
    if err != nil {
        return err
    }
    return svc.Restart()
}

func (s *ProcessService) Install(ctx context.Context) error {
    svc, err := s.getService()
    if err != nil {
        return err
    }
    return svc.Install()
}

func (s *ProcessService) Uninstall(ctx context.Context) error {
    svc, err := s.getService()
    if err != nil {
        return err
    }
    return svc.Uninstall()
}

// EnsureRunning checks if the service is installed and running, installing and starting it if necessary.
func (s *ProcessService) EnsureRunning(ctx context.Context) error {
	status, err := s.getStatusDirect(ctx)
	if err != nil {
		if err.Error() == "the service is not installed" || status == domain.ServiceStatusNotInstalled {
			log.Info().Msg("Service not installed. Installing...")
			if err := s.Install(ctx); err != nil {
				return err
			}
			// Re-check status or just proceed to start
		} else {
			return err
		}
	}

	if status == domain.ServiceStatusRunning {
		return nil
	}

	log.Info().Msg("Starting service...")
	return s.Start(ctx)
}


