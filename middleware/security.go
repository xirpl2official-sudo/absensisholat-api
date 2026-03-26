package middleware

import (
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// HTTPSEnforcement middleware redirects HTTP requests to HTTPS in production
func HTTPSEnforcement() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip in development
		if os.Getenv("ENVIRONMENT") != "production" {
			c.Next()
			return
		}

		// Check various headers that indicate HTTPS
		// X-Forwarded-Proto is set by most reverse proxies/load balancers
		proto := c.GetHeader("X-Forwarded-Proto")
		if proto == "" {
			proto = c.GetHeader("X-Forwarded-Protocol")
		}
		if proto == "" && c.GetHeader("X-Forwarded-Ssl") == "on" {
			proto = "https"
		}

		// If behind a proxy and not HTTPS, redirect
		if proto != "" && proto != "https" {
			// Build HTTPS URL
			host := c.Request.Host
			path := c.Request.URL.RequestURI()
			httpsURL := "https://" + host + path

			c.Redirect(http.StatusMovedPermanently, httpsURL)
			c.Abort()
			return
		}

		// Set HSTS header for HTTPS connections
		if proto == "https" || c.Request.TLS != nil {
			// Strict-Transport-Security: enforce HTTPS for 1 year, include subdomains
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		}

		c.Next()
	}
}

// SecurityHeaders adds security headers to all responses
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Prevent MIME type sniffing
		c.Header("X-Content-Type-Options", "nosniff")

		// Prevent clickjacking
		c.Header("X-Frame-Options", "DENY")

		// Enable XSS filter
		c.Header("X-XSS-Protection", "1; mode=block")

		// Control referrer information
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// Content Security Policy for API responses
		c.Header("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")

		// Prevent caching of sensitive data
		if isSensitivePath(c.Request.URL.Path) {
			c.Header("Cache-Control", "no-store, no-cache, must-revalidate, private")
			c.Header("Pragma", "no-cache")
			c.Header("Expires", "0")
		}

		c.Next()
	}
}

// isSensitivePath checks if the path contains sensitive data
func isSensitivePath(path string) bool {
	sensitivePaths := []string{
		"/auth/",
		"/me",
		"/export/",
	}

	for _, p := range sensitivePaths {
		if strings.Contains(path, p) {
			return true
		}
	}
	return false
}

// RequestID adds a unique request ID to each request
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if request ID already exists (from load balancer/proxy)
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			// Generate a simple request ID using timestamp and random component
			requestID = generateRequestID()
		}

		// Set request ID in context and response header
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)

		c.Next()
	}
}

// generateRequestID generates a unique request ID
func generateRequestID() string {
	// Use timestamp + random for uniqueness
	// In production, consider using github.com/google/uuid
	return strings.ReplaceAll(
		strings.ReplaceAll(
			time.Now().Format("20060102150405.000000"),
			".", ""),
		"-", "")
}
