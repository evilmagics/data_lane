package services

import (
	"context"
	"time"

	"github.com/kardianos/service"
	"github.com/rs/zerolog/log"

	"pdf_generator/internal/core/domain"
)

// StartMonitoring starts a background goroutine to check service status
func (s *ProcessService) StartMonitoring(ctx context.Context) {
	go func() {
		// Initial check
		s.checkAndUpdateStatus()
		
		// Load interval from settings or default 10s
		// For simplicity, we use 10s here, but ideally should watch settings?
		// Or refresh inside the loop.
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				s.checkAndUpdateStatus()
			}
		}
	}()
}

func (s *ProcessService) checkAndUpdateStatus() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	status, err := s.getStatusDirect(ctx)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to check service status during monitoring")
		// Keep previous status or set to unknown?
		// Maybe set to unknown if error persists?
		// For now, let's just log.
		return 
	}
	s.cachedStatus.Store(status)
}

// GetStatus returns the cached status
func (s *ProcessService) GetStatus(ctx context.Context) (domain.ServiceStatus, error) {
	val := s.cachedStatus.Load()
	if val == nil {
		return domain.ServiceStatusUnknown, nil
	}
	return val.(domain.ServiceStatus), nil
}

// getStatusDirect queries the system service manager directly
func (s *ProcessService) getStatusDirect(ctx context.Context) (domain.ServiceStatus, error) {
    svc, err := s.getService()
    if err != nil {
        return domain.ServiceStatusUnknown, err
    }
    
    status, err := svc.Status()
    if err != nil {
        return domain.ServiceStatusUnknown, err
    }
    
    switch status {
    case service.StatusRunning:
        return domain.ServiceStatusRunning, nil
    case service.StatusStopped:
        return domain.ServiceStatusStopped, nil
    default:
        return domain.ServiceStatusUnknown, nil
    }
}
