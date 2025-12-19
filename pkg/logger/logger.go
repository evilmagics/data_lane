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
	logFile *os.File
)

// InitLogger initializes zerolog with console and file output
func InitLogger(logDir string) error {
	if logDir == "" {
		logDir = "logs"
	}

	// Ensure log directory exists
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return err
	}

	// Daily log file
	filename := filepath.Join(logDir, time.Now().Format("2006-01-02")+".log")
	var err error
	logFile, err = os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0664)
	if err != nil {
		return err
	}

	// Multi-writer: Console (pretty) + File (JSON)
	consoleWriter := zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: time.RFC3339,
	}

	multi := zerolog.MultiLevelWriter(consoleWriter, logFile)
	log.Logger = zerolog.New(multi).With().Timestamp().Logger()

	// Set global log level
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	return nil
}

// Close closes the log file
func Close() {
	if logFile != nil {
		logFile.Close()
	}
}

// GetFileWriter returns the file writer for external use
func GetFileWriter() io.Writer {
	return logFile
}
