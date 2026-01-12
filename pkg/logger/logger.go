package logger

import (
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	appFile *os.File
	apiFile *os.File
	uiFile  *os.File

	App zerolog.Logger
	API zerolog.Logger
	UI  zerolog.Logger
)

// InitLogger initializes zerolog with console and file output split by App, API, and UI
func InitLogger(logDir string) error {
	if logDir == "" {
		logDir = "logs"
	}

	// Ensure log directory exists
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return err
	}

	// Set global time format to dd-MM-YYYY HH:mm:ss
	zerolog.TimeFieldFormat = "02-01-2006 15:04:05"
	// For console pretty output, we also need to set it in the console writer

	dateStr := time.Now().Format("02-01-2006")
	var err error

	// 1. App Logger (Default)
	appFilename := filepath.Join(logDir, dateStr+"_app.log")
	appFile, err = os.OpenFile(appFilename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0664)
	if err != nil {
		return err
	}

	// 2. API Logger
	apiFilename := filepath.Join(logDir, dateStr+"_api.log")
	apiFile, err = os.OpenFile(apiFilename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0664)
	if err != nil {
		return err
	}

	// 3. UI Logger
	uiFilename := filepath.Join(logDir, dateStr+"_ui.log")
	uiFile, err = os.OpenFile(uiFilename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0664)
	if err != nil {
		return err
	}

	// Console Writer (pretty)
	consoleWriter := zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: "02-01-2006 15:04:05",
	}

	// Create Multi-writers for each category
	appMulti := zerolog.MultiLevelWriter(consoleWriter, appFile)
	apiMulti := zerolog.MultiLevelWriter(consoleWriter, apiFile)
	uiMulti := zerolog.MultiLevelWriter(consoleWriter, uiFile)

	// Initialize Loggers with category tag
	App = zerolog.New(appMulti).With().Timestamp().Str("category", "app").Logger()
	API = zerolog.New(apiMulti).With().Timestamp().Str("category", "api").Logger()
	UI = zerolog.New(uiMulti).With().Timestamp().Str("category", "ui").Logger()

	// Set global log.Logger to App logger for backward compatibility
	log.Logger = App

	// Set global log level
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	return nil
}

// Close closes all log files
func Close() {
	if appFile != nil {
		appFile.Close()
	}
	if apiFile != nil {
		apiFile.Close()
	}
	if uiFile != nil {
		uiFile.Close()
	}
}

// GetFileWriter returns the app file writer (for backward compatibility)
func GetFileWriter() io.Writer {
	return appFile
}
