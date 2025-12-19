package services

import (
	"context"
	"sync/atomic"

	"github.com/kardianos/service"

	"pdf_generator/internal/core/domain"
)

type ProcessService struct {
	settingsService *SettingsService
    serviceConfig   *service.Config
	cachedStatus    atomic.Value // Stores domain.ServiceStatus
}

func NewProcessService(settingsService *SettingsService) *ProcessService {
    // Config must match the one in cmd/pdf-generator/main.go
    svcConfig := &service.Config{
        Name:        "DataLanePDFWorker",
        DisplayName: "DataLane PDF Worker Service",
        Description: "Background service for generating PDFs for DataLane application.",
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


