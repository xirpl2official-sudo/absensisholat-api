package middleware

import (
	"github.com/stretchr/testify/require"

	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"absensholat-api/utils"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// setupRouter creates a test router with the auth middleware
func setupRouter(allowedRoles ...string) *gin.Engine {
	router := gin.New()
	router.GET("/test", AuthMiddleware(allowedRoles...), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"nis":      c.GetString("nis"),
			"role":     c.GetString("role"),
			"email":    c.GetString("email"),
			"username": c.GetString("username"),
		})
	})
	return router
}

// ... (fungsi-fungsi test lainnya tidak berubah) ...

func TestAuthMiddleware_ValidToken(t *testing.T) {
	router := setupRouter()

	// Generate a valid token
	token, err := utils.GenerateToken("12345", "test@gmail.com", "siswa", "Test User")
	require.NoError(t, err) // <--- FIXED: Ganti if err != nil { t.Fatalf(...) } menjadi require.NoError

	req, err := http.NewRequest("GET", "/test", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestAuthMiddleware_TokenWithoutBearer(t *testing.T) {
	router := setupRouter()

	// Generate a valid token
	token, err := utils.GenerateToken("12345", "test@gmail.com", "siswa", "Test User")
	require.NoError(t, err) // <--- FIXED: Ganti if err != nil { t.Fatalf(...) } menjadi require.NoError

	req, err := http.NewRequest("GET", "/test", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", token) // No Bearer prefix
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d (token without Bearer should work)", http.StatusOK, w.Code)
	}
}

// ... (fungsi-fungsi test lainnya tidak berubah) ...

func TestAuthMiddleware_StaffToken(t *testing.T) {
	router := setupRouter("admin", "guru")

	// Generate a staff token
	token, err := utils.GenerateTokenWithNIP("admin001", "admin@school.com", "admin", "Admin User", "198501012010011001")
	require.NoError(t, err) // <--- FIXED: Ganti if err != nil { t.Fatalf(...) } menjadi require.NoError

	req, err := http.NewRequest("GET", "/test", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

// ... (fungsi-fungsi test lainnya tidak berubah) ...

func TestAuthMiddleware_ContextValues(t *testing.T) {
	router := gin.New()
	router.GET("/test", AuthMiddleware(), func(c *gin.Context) {
		nis := c.GetString("nis")
		email := c.GetString("email")
		role := c.GetString("role")
		name := c.GetString("name")

		if nis != "12345" {
			t.Errorf("Expected NIS '12345', got '%s'", nis)
		}
		if email != "test@gmail.com" {
			t.Errorf("Expected email 'test@gmail.com', got '%s'", email)
		}
		if role != "siswa" {
			t.Errorf("Expected role 'siswa', got '%s'", role)
		}
		if name != "Test User" {
			t.Errorf("Expected name 'Test User', got '%s'", name)
		}

		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	token, err := utils.GenerateToken("12345", "test@gmail.com", "siswa", "Test User")
	require.NoError(t, err) // <--- FIXED: Hapus if wrapper, langsung require.NoError
	req, err := http.NewRequest("GET", "/test", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

// ... (fungsi-fungsi test lainnya tidak berubah) ...

func TestAuthMiddleware_DebugInfoInDevelopment(t *testing.T) {
	// Ensure we're in development mode
	originalEnv := os.Getenv("ENVIRONMENT")
	defer os.Setenv("ENVIRONMENT", originalEnv)
	os.Setenv("ENVIRONMENT", "development")

	router := setupRouter()

	req, err := http.NewRequest("GET", "/test", nil)
	require.NoError(t, err) // <--- FIXED: Hapus if wrapper, langsung require.NoError
	// No Authorization header
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// In development, response should contain debug info
	body := w.Body.String()
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}

	// Debug field should be present in development
	if !testContains(body, "debug") {
		t.Errorf("Expected 'debug' field in development response, got: %s", body)
	}
}

func TestAuthMiddleware_NoDebugInfoInProduction(t *testing.T) {
	// Set production mode
	originalEnv := os.Getenv("ENVIRONMENT")
	defer os.Setenv("ENVIRONMENT", originalEnv)
	os.Setenv("ENVIRONMENT", "production")

	router := setupRouter()

	req, err := http.NewRequest("GET", "/test", nil)
	require.NoError(t, err) // <--- FIXED: Hapus if wrapper, langsung require.NoError
	// No Authorization header
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// In production, response should NOT contain debug info
	body := w.Body.String()
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}

	// Debug field should NOT be present in production
	if testContains(body, "debug") {
		t.Errorf("Expected no 'debug' field in production response, got: %s", body)
	}
}

// ... (fungsi testContains, testContainsHelper, BenchmarkAuthMiddleware tidak berubah) ...

// Helper function - renamed to avoid conflict with audit.go
func testContains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && testContainsHelper(s, substr))
}

func testContainsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Benchmark tests
func BenchmarkAuthMiddleware(b *testing.B) {
	gin.SetMode(gin.TestMode)
	router := setupRouter("siswa")

	token, err := utils.GenerateToken("12345", "test@gmail.com", "siswa", "Test User")
	require.NoError(b, err)
	req, err := http.NewRequest("GET", "/test", nil)
	require.NoError(b, err)
	req.Header.Set("Authorization", "Bearer "+token)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}
