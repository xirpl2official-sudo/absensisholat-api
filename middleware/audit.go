package middleware

import (
	"encoding/json"
	"fmt" // Added for error formatting if needed
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AuditAction represents the type of action being audited
type AuditAction string

const (
	AuditLogin          AuditAction = "LOGIN"
	AuditLogout         AuditAction = "LOGOUT"
	AuditRegister       AuditAction = "REGISTER"
	AuditPasswordChange AuditAction = "PASSWORD_CHANGE"
	AuditPasswordReset  AuditAction = "PASSWORD_RESET"
	AuditEmailChange    AuditAction = "EMAIL_CHANGE"
	AuditCreate         AuditAction = "CREATE"
	AuditUpdate         AuditAction = "UPDATE"
	AuditDelete         AuditAction = "DELETE"
	AuditExport         AuditAction = "EXPORT"
	AuditQRCodeGenerate AuditAction = "QRCODE_GENERATE"
	AuditQRCodeVerify   AuditAction = "QRCODE_VERIFY"
	AuditAccessDenied   AuditAction = "ACCESS_DENIED"
	AuditAuthFailure    AuditAction = "AUTH_FAILURE"
)

// AuditLog represents an audit log entry
type AuditLog struct {
	Timestamp  time.Time              `json:"timestamp"`
	RequestID  string                 `json:"request_id,omitempty"`
	Action     AuditAction            `json:"action"`
	UserID     string                 `json:"user_id,omitempty"`
	Username   string                 `json:"username,omitempty"`
	Role       string                 `json:"role,omitempty"`
	Resource   string                 `json:"resource"`
	ResourceID string                 `json:"resource_id,omitempty"`
	Method     string                 `json:"method"`
	Path       string                 `json:"path"`
	StatusCode int                    `json:"status_code"`
	IPAddress  string                 `json:"ip_address"`
	UserAgent  string                 `json:"user_agent"`
	Duration   int64                  `json:"duration_ms"`
	Success    bool                   `json:"success"`
	ErrorMsg   string                 `json:"error,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// AuditLogger handles audit logging
type AuditLogger struct {
	logger *zap.SugaredLogger
}

// NewAuditLogger creates a new audit logger
func NewAuditLogger(logger *zap.SugaredLogger) *AuditLogger {
	return &AuditLogger{logger: logger}
}

// Log writes an audit log entry
func (a *AuditLogger) Log(entry AuditLog) {
	// FIXED: Handle error from json.Marshal
	data, err := json.Marshal(entry)
	if err != nil {
		// Log the error and the problematic entry details
		a.logger.Errorw("Failed to marshal audit log entry",
			"error", err,
			"timestamp", entry.Timestamp,
			"action", entry.Action,
			"user_id", entry.UserID,
		)
		return // Or handle differently, e.g., send to a fallback logging mechanism
	}
	a.logger.Infow("AUDIT",
		"audit_data", string(data),
	)
}

// AuditMiddleware creates middleware for automatic audit logging
func AuditMiddleware(logger *zap.SugaredLogger) gin.HandlerFunc {
	auditLogger := NewAuditLogger(logger)

	return func(c *gin.Context) {
		// Skip non-auditable paths
		if !shouldAudit(c.Request.URL.Path) {
			c.Next()
			return
		}

		start := time.Now()

		// Process request
		c.Next()

		// Build audit log entry
		entry := AuditLog{
			Timestamp:  start,
			RequestID:  c.GetString("request_id"),
			Action:     determineAction(c),
			Resource:   determineResource(c.Request.URL.Path),
			Method:     c.Request.Method,
			Path:       c.Request.URL.Path,
			StatusCode: c.Writer.Status(),
			IPAddress:  c.ClientIP(),
			UserAgent:  c.Request.UserAgent(),
			Duration:   time.Since(start).Milliseconds(),
			Success:    c.Writer.Status() < 400,
		}

		// Add user info if authenticated
		// FIXED: Safe type assertion for nis
		if nis, exists := c.Get("nis"); exists {
			if s, ok := nis.(string); ok {
				entry.UserID = s
			} else {
				// Optional: Log a warning if type assertion fails unexpectedly
				// logger.Warnw("Failed to assert 'nis' as string for audit log", "type", fmt.Sprintf("%T", nis))
				// Or set a default value
				entry.UserID = fmt.Sprintf("%v", nis) // Convert to string representation if unsure
			}
			entry.Role = "siswa"
		} else if username, exists := c.Get("username"); exists {
			// FIXED: Safe type assertion for username
			if s, ok := username.(string); ok {
				entry.Username = s
			} else {
				// Optional: Log a warning or handle
				entry.Username = fmt.Sprintf("%v", username)
			}
			// FIXED: Safe type assertion for role
			if role, exists := c.Get("role"); exists {
				if r, ok := role.(string); ok {
					entry.Role = r
				} else {
					// Optional: Log a warning or handle
					entry.Role = fmt.Sprintf("%v", role)
				}
			}
		}

		// Add error message if present
		if len(c.Errors) > 0 {
			entry.ErrorMsg = c.Errors.Last().Error()
		}

		auditLogger.Log(entry)
	}
}

// shouldAudit determines if a path should be audited
func shouldAudit(path string) bool {
	auditablePaths := []string{
		"/auth/",
		"/siswa",
		"/export/",
		"/qrcode/",
	}

	for _, p := range auditablePaths {
		if len(path) >= len(p) && path[:len(p)] == p {
			return true
		}
	}

	// Also audit specific patterns
	if path == "/api/v1/auth/login" || path == "/api/auth/login" {
		return true
	}

	return false
}

// determineAction determines the audit action based on the request
func determineAction(c *gin.Context) AuditAction {
	path := c.Request.URL.Path
	method := c.Request.Method

	// Auth actions
	if contains(path, "/auth/login") {
		if c.Writer.Status() == 200 {
			return AuditLogin
		}
		return AuditAuthFailure
	}
	if contains(path, "/auth/register") {
		return AuditRegister
	}
	if contains(path, "/auth/reset-password") {
		return AuditPasswordReset
	}
	if contains(path, "/auth/change-email") {
		return AuditEmailChange
	}

	// Export actions
	if contains(path, "/export/") {
		return AuditExport
	}

	// QR Code actions
	if contains(path, "/qrcode/generate") || contains(path, "/qrcode/image") {
		return AuditQRCodeGenerate
	}
	if contains(path, "/qrcode/verify") {
		return AuditQRCodeVerify
	}

	// CRUD actions based on method
	switch method {
	case "POST":
		return AuditCreate
	case "PUT", "PATCH":
		return AuditUpdate
	case "DELETE":
		return AuditDelete
	}

	return AuditAction(method)
}

// determineResource extracts the resource name from the path
func determineResource(path string) string {
	// Remove /api prefix and version
	path = trimPrefix(path, "/api/v1")
	path = trimPrefix(path, "/api")

	// Get first path segment as resource
	parts := splitPath(path)
	if len(parts) > 0 {
		return parts[0]
	}
	return "unknown"
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func trimPrefix(s, prefix string) string {
	if len(s) >= len(prefix) && s[:len(prefix)] == prefix {
		return s[len(prefix):]
	}
	return s
}

func splitPath(path string) []string {
	var parts []string
	var current string
	for _, c := range path {
		if c == '/' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(c)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts
}

// LogAction manually logs an audit action (for use in handlers)
func LogAction(c *gin.Context, logger *zap.SugaredLogger, action AuditAction, resource, resourceID string, metadata map[string]interface{}) {
	auditLogger := NewAuditLogger(logger)

	entry := AuditLog{
		Timestamp:  time.Now(),
		RequestID:  c.GetString("request_id"),
		Action:     action,
		Resource:   resource,
		ResourceID: resourceID,
		Method:     c.Request.Method,
		Path:       c.Request.URL.Path,
		StatusCode: c.Writer.Status(),
		IPAddress:  c.ClientIP(),
		UserAgent:  c.Request.UserAgent(),
		Success:    true,
		Metadata:   metadata,
	}

	// Add user info
	// FIXED: Safe type assertion for nis (consistent with AuditMiddleware)
	if nis, exists := c.Get("nis"); exists {
		if s, ok := nis.(string); ok {
			entry.UserID = s
		} else {
			// Optional: Handle unexpected type
			entry.UserID = fmt.Sprintf("%v", nis)
		}
		entry.Role = "siswa"
	} else if username, exists := c.Get("username"); exists {
		// FIXED: Safe type assertion for username (consistent with AuditMiddleware)
		if s, ok := username.(string); ok {
			entry.Username = s
		} else {
			// Optional: Handle unexpected type
			entry.Username = fmt.Sprintf("%v", username)
		}
		// FIXED: Safe type assertion for role (consistent with AuditMiddleware)
		if role, exists := c.Get("role"); exists {
			if r, ok := role.(string); ok {
				entry.Role = r
			} else {
				// Optional: Handle unexpected type
				entry.Role = fmt.Sprintf("%v", role)
			}
		}
	}

	auditLogger.Log(entry)
}
