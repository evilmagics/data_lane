package database

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"pdf_generator/internal/core/domain"
)

var DB *gorm.DB

// DefaultDBPath is the default database path
const DefaultDBPath = "data/app.db"

// InitDB initializes the SQLite database connection and runs migrations
func InitDB(dbPath string) error {
	if dbPath == "" {
		dbPath = DefaultDBPath
	}

	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Warn,
			IgnoreRecordNotFoundError: true,
			ParameterizedQueries:      true,
			Colorful:                  true,
		},
	)

	var err error
	DB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		return err
	}

	DB.Exec("PRAGMA journal_mode=WAL;")
	DB.Exec("PRAGMA busy_timeout=5000;")
	DB.Exec("PRAGMA synchronous=NORMAL;")

	// Connection pool settings
	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}
	sqlDB.SetMaxIdleConns(1)
	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Run migrations
	if err := runMigrations(); err != nil {
		return err
	}

	// Seed default settings
	if err := seedDefaults(); err != nil {
		return err
	}

	go startBackgroundWALSync(dbPath)

	return nil
}

func startBackgroundWALSync(dbPath string) {
	walPath := dbPath + "-wal"
	lastCheckpoint := time.Now()

	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		// Read settings
		var intervalSetting, sizeSetting domain.Settings
		var intervalMinutes int = 30
		var maxSizeMB float64 = 20

		if err := DB.Where("key = ?", domain.SettingWALCheckpointInterval).First(&intervalSetting).Error; err == nil {
			if v, err := strconv.Atoi(intervalSetting.Value); err == nil {
				intervalMinutes = v
			}
		}
		if err := DB.Where("key = ?", domain.SettingWALMaxSizeMB).First(&sizeSetting).Error; err == nil {
			if v, err := strconv.ParseFloat(sizeSetting.Value, 64); err == nil {
				maxSizeMB = v
			}
		}

		shouldCheckpoint := false

		// Time check
		if time.Since(lastCheckpoint) >= time.Duration(intervalMinutes)*time.Minute {
			shouldCheckpoint = true
		}

		// Size check
		if !shouldCheckpoint {
			info, err := os.Stat(walPath)
			if err == nil {
				sizeMB := float64(info.Size()) / 1024 / 1024
				if sizeMB >= maxSizeMB {
					shouldCheckpoint = true
				}
			}
		}

		if shouldCheckpoint {
			// log.Println("Running WAL checkpoint...")
			DB.Exec("PRAGMA wal_checkpoint(TRUNCATE);")
			lastCheckpoint = time.Now()
		}
	}
}

func runMigrations() error {
	return DB.AutoMigrate(
		&domain.Task{},
		&domain.Schedule{},
		&domain.Settings{},
		&domain.Session{},
		&domain.APIKey{},
		&domain.Log{},
		&domain.Gate{},
	)
}

func seedDefaults() error {
	defaults := domain.DefaultSettings()
	for _, setting := range defaults {
		var existing domain.Settings
		result := DB.Where("key = ?", setting.Key).First(&existing)
		if result.Error == gorm.ErrRecordNotFound {
			if err := DB.Create(&setting).Error; err != nil {
				return err
			}
		} else {
			// Update metadata fields, preserve Value
			existing.Group = setting.Group
			existing.DataType = setting.DataType
			existing.Content = setting.Content
			existing.Name = setting.Name
			existing.Icon = setting.Icon

			// Special migration: time_overlap was changed from minutes (number) to HH:MM (time)
			// Convert old numeric value to time format
			if setting.Key == domain.SettingTimeOverlap && setting.DataType == "time" {
				// Check if existing value is in old format (numeric minutes)
				if minutes, err := strconv.Atoi(existing.Value); err == nil {
					// Convert minutes to HH:MM format
					hours := minutes / 60
					mins := minutes % 60
					existing.Value = fmt.Sprintf("%02d:%02d", hours, mins)
				} else if len(existing.Value) != 5 || existing.Value[2] != ':' {
					// Invalid format, reset to default
					existing.Value = setting.Value
				}
			}

			if err := DB.Save(&existing).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

// GetDB returns the database instance
func GetDB() *gorm.DB {
	return DB
}
