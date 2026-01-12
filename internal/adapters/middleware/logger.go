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
		
        // Read request body
        reqBody := c.Body()

		// Process request
		err := c.Next()
		
		// Log after request
		duration := time.Since(start)
		path := c.Path()

        // Read response body
        resBody := c.Response().Body()
		
		// Select logger based on path
		l := logger.API
		if !strings.HasPrefix(path, "/api") && !strings.HasPrefix(path, "/sse") {
			l = logger.UI
		}

		// Create log event
        evt := l.Info().
			Str("method", c.Method()).
			Str("path", path).
			Int("status", c.Response().StatusCode()).
			Dur("duration", duration).
			Str("ip", c.IP())
        
        // Add bodies for API requests if not too large
        if strings.HasPrefix(path, "/api") {
            if len(reqBody) > 0 && len(reqBody) < 2048 {
                 evt.Str("req_body", string(reqBody))
            }
             if len(resBody) > 0 && len(resBody) < 2048 {
                 evt.Str("res_body", string(resBody))
            }
        }

        evt.Msg("HTTP Request")
		
		return err
	}
}
