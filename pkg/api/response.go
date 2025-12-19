package api

import "github.com/gofiber/fiber/v3"

// Response represents the standard API response
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Error codes
const (
	CodeSuccess               = 0
	CodeInvalidRequest        = 1001
	CodeValidationError       = 1002
	CodeUnauthorized          = 2001
	CodeInvalidCredentials    = 2002
	CodeSessionLimitReached   = 2003
	CodeTokenExpired          = 2004
	CodeInvalidAPIKey         = 2005
	CodeHMACMismatch          = 2006
	CodeNotFound              = 3001
	CodeTaskNotReady          = 3002
	CodeInternalError         = 5001
)

// Success returns a successful response
func Success(c fiber.Ctx, data interface{}) error {
	return c.JSON(Response{
		Code:    CodeSuccess,
		Message: "OK",
		Data:    data,
	})
}

// Error returns an error response
func Error(c fiber.Ctx, code int, message string) error {
	status := fiber.StatusOK
	if code >= 2000 && code < 3000 {
		status = fiber.StatusUnauthorized
	} else if code >= 3000 && code < 4000 {
		status = fiber.StatusNotFound
	} else if code >= 5000 {
		status = fiber.StatusInternalServerError
	} else if code >= 1000 && code < 2000 {
		status = fiber.StatusBadRequest
	}

	return c.Status(status).JSON(Response{
		Code:    code,
		Message: message,
		Data:    nil,
	})
}

// PaginatedResponse wraps paginated data
type PaginatedResponse struct {
	Items      interface{} `json:"items"`
	Pagination Pagination  `json:"pagination"`
}

// Pagination metadata
type Pagination struct {
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	TotalPages int   `json:"total_pages"`
}
