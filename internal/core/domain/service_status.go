package domain

type ServiceStatus string

const (
	ServiceStatusRunning      ServiceStatus = "running"
	ServiceStatusStopped      ServiceStatus = "stopped"
	ServiceStatusNotInstalled ServiceStatus = "not_installed"
	ServiceStatusUnknown      ServiceStatus = "unknown"
)
