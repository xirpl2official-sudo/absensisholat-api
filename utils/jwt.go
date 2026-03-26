package utils

import (
	"errors"
	"log"
	"os"
	"time"
	"unicode"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var jwtSecret = []byte(getJWTSecret())

func getJWTSecret() string {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		env := os.Getenv("ENVIRONMENT")
		if env == "production" {
			log.Fatal("CRITICAL: JWT_SECRET environment variable must be set in production")
		}
		// Only allow default secret in development
		log.Println("WARNING: Using default JWT secret. Set JWT_SECRET for production.")
		secret = "dev-secret-key-not-for-production"
	}
	if len(secret) < 32 {
		log.Println("WARNING: JWT_SECRET should be at least 32 characters for security")
	}
	return secret
}

type Claims struct {
	NIS      string `json:"nis,omitempty"`      // For siswa
	Username string `json:"username,omitempty"` // For staff
	Email    string `json:"email"`
	Role     string `json:"role"`          // siswa, guru, wali_kelas, admin
	Name     string `json:"name"`          // User's display name
	NIP      string `json:"nip,omitempty"` // For guru/wali_kelas
	jwt.RegisteredClaims
}

// GenerateToken creates a JWT token for siswa
func GenerateToken(nis, email, role, name string) (string, error) {
	claims := &Claims{
		NIS:   nis,
		Email: email,
		Role:  role,
		Name:  name,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)), // Short-lived access token
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// GenerateTokenWithNIP creates a JWT token for staff with NIP
func GenerateTokenWithNIP(username, email, role, name, nip string) (string, error) {
	claims := &Claims{
		Username: username,
		Email:    email,
		Role:     role,
		Name:     name,
		NIP:      nip,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)), // Short-lived access token
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// GenerateRefreshToken creates a long-lived JWT token for refreshing
func GenerateRefreshToken(userID, role string) (string, error) {
	claims := &Claims{
		// Use Username field for generic UserID storage in refresh token
		Username: userID,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)), // Long-lived refresh token
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// PasswordValidationError represents password validation errors
type PasswordValidationError struct {
	Errors []string `json:"errors"`
}

func (e *PasswordValidationError) Error() string {
	return "password validation failed"
}

// ValidatePassword checks password strength requirements
// Returns nil if password is valid, otherwise returns PasswordValidationError
func ValidatePassword(password string) error {
	var validationErrors []string

	// Minimum length check
	if len(password) < 8 {
		validationErrors = append(validationErrors, "Password harus minimal 8 karakter")
	}

	// Maximum length check (prevent DoS with bcrypt)
	if len(password) > 72 {
		validationErrors = append(validationErrors, "Password maksimal 72 karakter")
	}

	var (
		hasUpper   bool
		hasLower   bool
		hasNumber  bool
		hasSpecial bool
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if !hasUpper {
		validationErrors = append(validationErrors, "Password harus mengandung minimal 1 huruf besar")
	}
	if !hasLower {
		validationErrors = append(validationErrors, "Password harus mengandung minimal 1 huruf kecil")
	}
	if !hasNumber {
		validationErrors = append(validationErrors, "Password harus mengandung minimal 1 angka")
	}
	if !hasSpecial {
		validationErrors = append(validationErrors, "Password harus mengandung minimal 1 karakter khusus")
	}

	if len(validationErrors) > 0 {
		return &PasswordValidationError{Errors: validationErrors}
	}

	return nil
}
