package middleware

import (
	"context"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"absensholat-api/utils"

	"github.com/gin-gonic/gin"
)

// RateLimiter implements a simple token bucket rate limiter
type RateLimiter struct {
	visitors map[string]*visitor
	mu       sync.RWMutex
	rate     int           // requests per window
	window   time.Duration // time window
}

type visitor struct {
	tokens    int
	lastReset time.Time
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(requestsPerMinute int) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		rate:     requestsPerMinute,
		window:   time.Minute,
	}

	// Cleanup old entries every 5 minutes
	go rl.cleanup()

	return rl
}

// getVisitor returns the rate limit data for an IP
func (rl *RateLimiter) getVisitor(ip string) *visitor {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, exists := rl.visitors[ip]
	if !exists {
		v = &visitor{
			tokens:    rl.rate,
			lastReset: time.Now(),
		}
		rl.visitors[ip] = v
		return v
	}

	// Reset tokens if window has passed
	if time.Since(v.lastReset) > rl.window {
		v.tokens = rl.rate
		v.lastReset = time.Now()
	}

	return v
}

// Allow checks if the request should be allowed
func (rl *RateLimiter) Allow(ip string) bool {
	v := rl.getVisitor(ip)

	rl.mu.Lock()
	defer rl.mu.Unlock()

	if v.tokens > 0 {
		v.tokens--
		return true
	}

	return false
}

// cleanup removes old visitor entries periodically
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	for range ticker.C {
		rl.mu.Lock()
		for ip, v := range rl.visitors {
			if time.Since(v.lastReset) > 10*time.Minute {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// RateLimitMiddleware returns a Gin middleware for rate limiting
func RateLimitMiddleware() gin.HandlerFunc {
	// Get rate limit from environment or use default
	rpmStr := os.Getenv("RATE_LIMIT_RPM")
	rpm := 60 // default: 60 requests per minute

	if rpmStr != "" {
		if parsed, err := strconv.Atoi(rpmStr); err == nil && parsed > 0 {
			rpm = parsed
		}
	}

	// Use Redis-based rate limiting if available, otherwise fallback to in-memory
	if utils.CacheEnabled() {
		return redisRateLimitMiddleware(rpm)
	}
	return memoryRateLimitMiddleware(rpm)
}

// redisRateLimitMiddleware implements Redis-based rate limiting
func redisRateLimitMiddleware(rpm int) gin.HandlerFunc {
	return redisRateLimitMiddlewareWithMessage(rpm, "Too many requests. Please try again later.", "RATE_LIMIT_EXCEEDED")
}

// redisRateLimitMiddlewareWithMessage implements Redis-based rate limiting with custom message
func redisRateLimitMiddlewareWithMessage(rpm int, message, code string) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.Background()
		ip := c.ClientIP()
		key := "ratelimit:" + ip

		cache := utils.GetCache()
		count, err := cache.Incr(ctx, key)
		if err != nil {
			// Fallback to allowing request if Redis fails
			c.Next()
			return
		}

		// Set expiration on first request
		if count == 1 {
			if errExp := cache.Expire(ctx, key, time.Minute); errExp != nil {
				// Optional: log if needed, though typically ignored in rate limiters
			}
		}

		if count > int64(rpm) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"message": message,
				"code":    code,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// memoryRateLimitMiddleware implements in-memory rate limiting (fallback)
func memoryRateLimitMiddleware(rpm int) gin.HandlerFunc {
	return memoryRateLimitMiddlewareWithMessage(rpm, "Too many requests. Please try again later.", "RATE_LIMIT_EXCEEDED")
}

// memoryRateLimitMiddlewareWithMessage implements in-memory rate limiting with custom message
func memoryRateLimitMiddlewareWithMessage(rpm int, message, code string) gin.HandlerFunc {
	limiter := NewRateLimiter(rpm)

	return func(c *gin.Context) {
		ip := c.ClientIP()

		if !limiter.Allow(ip) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"message": message,
				"code":    code,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// StrictRateLimitMiddleware applies stricter rate limiting for sensitive endpoints
// like login, register, forgot-password (5 requests per minute)
func StrictRateLimitMiddleware() gin.HandlerFunc {
	rpm := 5 // 5 requests per minute for auth endpoints

	// Use Redis-based rate limiting if available, otherwise fallback to in-memory
	if utils.CacheEnabled() {
		return redisRateLimitMiddlewareWithMessage(rpm, "Too many attempts. Please wait before trying again.", "AUTH_RATE_LIMIT_EXCEEDED")
	}
	return memoryRateLimitMiddlewareWithMessage(rpm, "Too many attempts. Please wait before trying again.", "AUTH_RATE_LIMIT_EXCEEDED")
}
