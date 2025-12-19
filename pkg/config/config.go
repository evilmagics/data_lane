package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Server   ServerConfig
	Session  SessionConfig
	Database DatabaseConfig
}

type ServerConfig struct {
	Port         int
	Host         string
	AllowOrigins string
}

type SessionConfig struct {
	Expiry        time.Duration
	MaxConcurrent int
}

type DatabaseConfig struct {
	Path string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	// Default config
	cfg := &Config{
		Server: ServerConfig{
			Port: 3001,
			Host: "0.0.0.0",
		},
		Session: SessionConfig{
			Expiry:        12 * time.Hour,
			MaxConcurrent: 5,
		},
		Database: DatabaseConfig{
			Path: "data/app.db",
		},
	}

	// Server Config
	if port := os.Getenv("SERVER_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			cfg.Server.Port = p
		}
	}
	if host := os.Getenv("SERVER_HOST"); host != "" {
		cfg.Server.Host = host
	}
	if origins := os.Getenv("ALLOW_ORIGINS"); origins != "" {
		cfg.Server.AllowOrigins = origins
	}

	// Database Config
	if dbPath := os.Getenv("DB_PATH"); dbPath != "" {
		cfg.Database.Path = dbPath
	}

	// Session Config
	if expiry := os.Getenv("SESSION_EXPIRY"); expiry != "" {
		if duration, err := time.ParseDuration(expiry); err == nil {
			cfg.Session.Expiry = duration
		} else {
             // Try parsing as integer hours if duration parse fails, or just default to hours
             // For now assume standard Go duration string like "24h"
             // Fallback for simple integer check?
             if p, err := strconv.Atoi(expiry); err == nil {
                 cfg.Session.Expiry = time.Duration(p) * time.Hour
             }
        }
	}
    if maxConcurrent := os.Getenv("SESSION_MAX_CONCURRENT"); maxConcurrent != "" {
        if m, err := strconv.Atoi(maxConcurrent); err == nil {
            cfg.Session.MaxConcurrent = m
        }
    }

	return cfg, nil
}
