package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestNewRateLimiter(t *testing.T) {
	limiter := NewRateLimiter(10)

	if limiter == nil {
		t.Fatal("NewRateLimiter returned nil")
	}

	if limiter.rate != 10 {
		t.Errorf("Expected rate 10, got %d", limiter.rate)
	}

	if limiter.window != time.Minute {
		t.Errorf("Expected window 1 minute, got %v", limiter.window)
	}
}

func TestRateLimiter_Allow(t *testing.T) {
	limiter := NewRateLimiter(5)
	ip := "192.168.1.1"

	// Should allow first 5 requests
	for i := 0; i < 5; i++ {
		if !limiter.Allow(ip) {
			t.Errorf("Request %d should be allowed", i+1)
		}
	}

	// 6th request should be blocked
	if limiter.Allow(ip) {
		t.Error("6th request should be blocked")
	}
}

func TestRateLimiter_DifferentIPs(t *testing.T) {
	limiter := NewRateLimiter(2)
	ip1 := "192.168.1.1"
	ip2 := "192.168.1.2"

	// Use up ip1's quota
	limiter.Allow(ip1)
	limiter.Allow(ip1)

	// ip1 should be blocked
	if limiter.Allow(ip1) {
		t.Error("ip1 should be blocked after 2 requests")
	}

	// ip2 should still be allowed
	if !limiter.Allow(ip2) {
		t.Error("ip2 should still be allowed")
	}
}

func TestRateLimiter_ConcurrentAccess(t *testing.T) {
	limiter := NewRateLimiter(100)
	ip := "192.168.1.1"

	var wg sync.WaitGroup
	allowed := 0
	var mu sync.Mutex

	// Make 150 concurrent requests
	for i := 0; i < 150; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if limiter.Allow(ip) {
				mu.Lock()
				allowed++
				mu.Unlock()
			}
		}()
	}

	wg.Wait()

	if allowed != 100 {
		t.Errorf("Expected exactly 100 allowed requests, got %d", allowed)
	}
}

func TestRateLimitMiddleware(t *testing.T) {
	// Set a low rate limit for testing
	originalRPM := os.Getenv("RATE_LIMIT_RPM")
	defer os.Setenv("RATE_LIMIT_RPM", originalRPM)
	os.Setenv("RATE_LIMIT_RPM", "5")

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RateLimitMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Make 5 successful requests
	for i := 0; i < 5; i++ {
		req, err := http.NewRequest("GET", "/test", nil)
		require.NoError(t, err)
		req.RemoteAddr = "192.168.1.100:12345"
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Request %d: Expected status %d, got %d", i+1, http.StatusOK, w.Code)
		}
	}

	// 6th request should be rate limited
	req, err := http.NewRequest("GET", "/test", nil)
	require.NoError(t, err)
	req.RemoteAddr = "192.168.1.100:12345"
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("Expected status %d for rate limited request, got %d", http.StatusTooManyRequests, w.Code)
	}
}

func TestRateLimitMiddleware_DefaultRPM(t *testing.T) {
	// Clear the env var to test default
	originalRPM := os.Getenv("RATE_LIMIT_RPM")
	defer os.Setenv("RATE_LIMIT_RPM", originalRPM)
	os.Unsetenv("RATE_LIMIT_RPM")

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RateLimitMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Should allow at least 60 requests with default
	for i := 0; i < 60; i++ {
		req, err := http.NewRequest("GET", "/test", nil)
		require.NoError(t, err)
		req.RemoteAddr = "10.0.0.1:12345"
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Request %d: Expected status %d, got %d (default should be 60 RPM)", i+1, http.StatusOK, w.Code)
			break
		}
	}
}

func TestStrictRateLimitMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(StrictRateLimitMiddleware())
	router.POST("/login", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Make 5 successful requests (strict limit is 5 per minute)
	for i := 0; i < 5; i++ {
		req, err := http.NewRequest("POST", "/login", nil)
		require.NoError(t, err)
		req.RemoteAddr = "192.168.1.200:12345"
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Request %d: Expected status %d, got %d", i+1, http.StatusOK, w.Code)
		}
	}

	// 6th request should be rate limited
	req, err := http.NewRequest("POST", "/login", nil)
	require.NoError(t, err)
	req.RemoteAddr = "192.168.1.200:12345"
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("Expected status %d for strict rate limited request, got %d", http.StatusTooManyRequests, w.Code)
	}

	// Check error response
	body := w.Body.String()
	if !strings.Contains(body, "AUTH_RATE_LIMIT_EXCEEDED") {
		t.Errorf("Expected error code AUTH_RATE_LIMIT_EXCEEDED in response, got: %s", body)
	}
}

func TestRateLimitMiddleware_InvalidRPM(t *testing.T) {
	originalRPM := os.Getenv("RATE_LIMIT_RPM")
	defer os.Setenv("RATE_LIMIT_RPM", originalRPM)
	os.Setenv("RATE_LIMIT_RPM", "invalid")

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RateLimitMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Should fall back to default 60 RPM
	for i := 0; i < 60; i++ {
		req, err := http.NewRequest("GET", "/test", nil)
		require.NoError(t, err)
		req.RemoteAddr = "172.16.0.1:12345"
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Request %d should be allowed with default RPM fallback", i+1)
			break
		}
	}
}

func TestRateLimitMiddleware_NegativeRPM(t *testing.T) {
	originalRPM := os.Getenv("RATE_LIMIT_RPM")
	defer os.Setenv("RATE_LIMIT_RPM", originalRPM)
	os.Setenv("RATE_LIMIT_RPM", "-10")

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RateLimitMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Should fall back to default 60 RPM
	req, err := http.NewRequest("GET", "/test", nil)
	require.NoError(t, err)
	req.RemoteAddr = "172.16.0.2:12345"
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d with negative RPM fallback, got %d", http.StatusOK, w.Code)
	}
}

func TestRateLimiter_GetVisitor(t *testing.T) {
	limiter := NewRateLimiter(10)
	ip := "192.168.1.50"

	// First call should create a new visitor
	v1 := limiter.getVisitor(ip)
	if v1 == nil {
		t.Fatal("getVisitor returned nil")
	}
	if v1.tokens != 10 {
		t.Errorf("Expected 10 tokens, got %d", v1.tokens)
	}

	// Second call should return the same visitor
	v2 := limiter.getVisitor(ip)
	if v1 != v2 {
		t.Error("getVisitor should return the same visitor for the same IP")
	}
}

// Benchmark tests
func BenchmarkRateLimiter_Allow(b *testing.B) {
	limiter := NewRateLimiter(1000000) // High limit to avoid blocking
	ip := "192.168.1.1"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		limiter.Allow(ip)
	}
}

func BenchmarkRateLimiter_Allow_Concurrent(b *testing.B) {
	limiter := NewRateLimiter(1000000) // High limit to avoid blocking

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		ip := "192.168.1.1"
		for pb.Next() {
			limiter.Allow(ip)
		}
	})
}

func BenchmarkRateLimitMiddleware(b *testing.B) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RateLimitMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req, err := http.NewRequest("GET", "/test", nil)
	require.NoError(b, err)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}
