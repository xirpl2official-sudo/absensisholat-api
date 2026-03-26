package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// APIError represents a structured API error response
type APIError struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// Common error codes
const (
	// Authentication errors
	ErrCodeUnauthorized       = "UNAUTHORIZED"
	ErrCodeInvalidToken       = "INVALID_TOKEN"
	ErrCodeTokenExpired       = "TOKEN_EXPIRED"
	ErrCodeInvalidCredentials = "INVALID_CREDENTIALS"

	// Authorization errors
	ErrCodeForbidden        = "FORBIDDEN"
	ErrCodeInsufficientRole = "INSUFFICIENT_ROLE"

	// Validation errors
	ErrCodeValidationFailed = "VALIDATION_FAILED"
	ErrCodeInvalidInput     = "INVALID_INPUT"
	ErrCodeMissingField     = "MISSING_FIELD"
	ErrCodeInvalidEmail     = "INVALID_EMAIL"
	ErrCodeWeakPassword     = "WEAK_PASSWORD"

	// Resource errors
	ErrCodeNotFound      = "NOT_FOUND"
	ErrCodeAlreadyExists = "ALREADY_EXISTS"
	ErrCodeConflict      = "CONFLICT"

	// Server errors
	ErrCodeInternalError   = "INTERNAL_ERROR"
	ErrCodeDatabaseError   = "DATABASE_ERROR"
	ErrCodeExternalService = "EXTERNAL_SERVICE_ERROR"

	// Rate limiting
	ErrCodeRateLimited = "RATE_LIMITED"
)

// NewAPIError creates a new API error
func NewAPIError(code, message string, details interface{}) *APIError {
	return &APIError{
		Code:    code,
		Message: message,
		Details: details,
	}
}

// RespondError sends an error response with the appropriate status code
func RespondError(c *gin.Context, status int, code, message string, details interface{}) {
	c.JSON(status, APIError{
		Code:    code,
		Message: message,
		Details: details,
	})
}

// Common error response helpers

// RespondUnauthorized sends a 401 Unauthorized response
func RespondUnauthorized(c *gin.Context, message string) {
	RespondError(c, http.StatusUnauthorized, ErrCodeUnauthorized, message, nil)
}

// RespondForbidden sends a 403 Forbidden response
func RespondForbidden(c *gin.Context, message string) {
	RespondError(c, http.StatusForbidden, ErrCodeForbidden, message, nil)
}

// RespondNotFound sends a 404 Not Found response
func RespondNotFound(c *gin.Context, resource string) {
	RespondError(c, http.StatusNotFound, ErrCodeNotFound, resource+" tidak ditemukan", nil)
}

// RespondBadRequest sends a 400 Bad Request response
func RespondBadRequest(c *gin.Context, message string, details interface{}) {
	RespondError(c, http.StatusBadRequest, ErrCodeInvalidInput, message, details)
}

// RespondValidationError sends a 400 response for validation errors
func RespondValidationError(c *gin.Context, errors interface{}) {
	RespondError(c, http.StatusBadRequest, ErrCodeValidationFailed, "Validation failed", errors)
}

// RespondConflict sends a 409 Conflict response
func RespondConflict(c *gin.Context, message string) {
	RespondError(c, http.StatusConflict, ErrCodeConflict, message, nil)
}

// RespondInternalError sends a 500 Internal Server Error response
func RespondInternalError(c *gin.Context, message string) {
	RespondError(c, http.StatusInternalServerError, ErrCodeInternalError, message, nil)
}

// RespondDatabaseError sends a 500 response for database errors
func RespondDatabaseError(c *gin.Context) {
	RespondError(c, http.StatusInternalServerError, ErrCodeDatabaseError, "Database error occurred", nil)
}

// RespondSuccess sends a success response with data
func RespondSuccess(c *gin.Context, status int, message string, data interface{}) {
	response := gin.H{
		"message": message,
	}
	if data != nil {
		response["data"] = data
	}
	c.JSON(status, response)
}
