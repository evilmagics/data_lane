package middleware

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog/log"
)

// LoggerMiddleware logs HTTP requests
func LoggerMiddleware() fiber.Handler {
	return func(c fiber.Ctx) error {
		start := time.Now()
		
		// Process request
		err := c.Next()
		
		// Log after request
		duration := time.Since(start)
		log.Info().
			Str("method", c.Method()).
			Str("path", c.Path()).
			Int("status", c.Response().StatusCode()).
			Dur("duration", duration).
			Str("ip", c.IP()).
			Msg("HTTP Request")
		
		return err
	}
}
