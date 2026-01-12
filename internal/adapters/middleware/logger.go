package middleware

import (
	"strings"
	"time"

	"pdf_generator/pkg/logger"

	"github.com/gofiber/fiber/v3"
)

// LoggerMiddleware logs HTTP requests categorized by API or UI
func LoggerMiddleware() fiber.Handler {
	return func(c fiber.Ctx) error {
		start := time.Now()
		
		// Process request
		err := c.Next()
		
		// Log after request
		duration := time.Since(start)
		path := c.Path()
		
		// Select logger based on path
		l := logger.API
		if !strings.HasPrefix(path, "/api") && !strings.HasPrefix(path, "/sse") {
			l = logger.UI
		}

		l.Info().
			Str("method", c.Method()).
			Str("path", path).
			Int("status", c.Response().StatusCode()).
			Dur("duration", duration).
			Str("ip", c.IP()).
			Msg("HTTP Request")
		
		return err
	}
}
