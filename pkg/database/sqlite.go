package database

import (
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

// InitDB initializes the SQLite database connection and runs migrations
func InitDB(dbPath string) error {
	if dbPath == "" {
		dbPath = os.Getenv("DB_PATH")
		if dbPath == "" {
			dbPath = "app.db"
		}
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
		&domain.Station{},
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
			// Description was removed, so no need to clean it up manually if column dropped,
			// though GORM AutoMigrate doesn't drop columns by default.
			// We can leave it for now.
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
