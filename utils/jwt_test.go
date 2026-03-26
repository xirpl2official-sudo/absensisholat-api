package utils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHashPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{"valid password", "MySecurePass123!", false},
		{"empty password", "", false},
		{"long password", "ThisIsAVeryLongPasswordThatIsStillWithinThe72CharLimit!!", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := HashPassword(tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("HashPassword() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && hash == "" {
				t.Error("HashPassword() returned empty hash")
			}
			if !tt.wantErr && hash == tt.password {
				t.Error("HashPassword() returned unhashed password")
			}
		})
	}
}

func TestCheckPasswordHash(t *testing.T) {
	password := "TestPassword123!"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	tests := []struct {
		name     string
		password string
		hash     string
		want     bool
	}{
		{"correct password", password, hash, true},
		{"incorrect password", "WrongPassword123!", hash, false},
		{"empty password against valid hash", "", hash, false},
		{"password against invalid hash", password, "invalid-hash", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CheckPasswordHash(tt.password, tt.hash); got != tt.want {
				t.Errorf("CheckPasswordHash() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name       string
		password   string
		wantErr    bool
		errorCount int
	}{
		{"valid password with all requirements", "SecurePass123!", false, 0},
		{"too short password", "Ab1!", true, 1},
		{"missing uppercase", "securepass123!", true, 1},
		{"missing lowercase", "SECUREPASS123!", true, 1},
		{"missing number", "SecurePassword!", true, 1},
		{"missing special character", "SecurePass123", true, 1},
		{"all lowercase only", "password", true, 3},
		{"numbers only", "12345678", true, 3},
		{"empty password", "", true, 5},
		{"exact minimum length", "Abcd12!@", false, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePassword(tt.password)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePassword() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && err != nil {
				pwdErr, ok := err.(*PasswordValidationError)
				if !ok {
					t.Errorf("Expected PasswordValidationError, got %T", err)
					return
				}
				if len(pwdErr.Errors) != tt.errorCount {
					t.Errorf("Expected %d errors, got %d: %v", tt.errorCount, len(pwdErr.Errors), pwdErr.Errors)
				}
			}
		})
	}
}

func TestGenerateToken(t *testing.T) {
	tests := []struct {
		name    string
		nis     string
		email   string
		role    string
		nameStr string
		wantErr bool
	}{
		{"valid siswa token", "12345", "student@gmail.com", "siswa", "John Doe", false},
		{"empty fields", "", "", "", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := GenerateToken(tt.nis, tt.email, tt.role, tt.nameStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && token == "" {
				t.Error("GenerateToken() returned empty token")
			}
		})
	}
}

func TestGenerateTokenWithNIP(t *testing.T) {
	tests := []struct {
		name     string
		username string
		email    string
		role     string
		nameStr  string
		nip      string
		wantErr  bool
	}{
		{"valid guru token", "guru001", "guru@school.com", "guru", "Budi Santoso", "198501012010011001", false},
		{"valid admin token", "admin001", "admin@school.com", "admin", "Admin User", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := GenerateTokenWithNIP(tt.username, tt.email, tt.role, tt.nameStr, tt.nip)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateTokenWithNIP() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && token == "" {
				t.Error("GenerateTokenWithNIP() returned empty token")
			}
		})
	}
}

func TestValidateToken(t *testing.T) {
	validToken, err := GenerateToken("12345", "test@gmail.com", "siswa", "Test User")
	if err != nil {
		t.Fatalf("Failed to generate token for testing: %v", err)
	}

	tests := []struct {
		name      string
		token     string
		wantErr   bool
		wantNIS   string
		wantRole  string
		wantEmail string
	}{
		{"valid token", validToken, false, "12345", "siswa", "test@gmail.com"},
		{"empty token", "", true, "", "", ""},
		{"invalid token format", "not.a.valid.jwt.token", true, "", "", ""},
		{"malformed token", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.invalid", true, "", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := ValidateToken(tt.token)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && claims != nil {
				if claims.NIS != tt.wantNIS {
					t.Errorf("ValidateToken() NIS = %v, want %v", claims.NIS, tt.wantNIS)
				}
				if claims.Role != tt.wantRole {
					t.Errorf("ValidateToken() Role = %v, want %v", claims.Role, tt.wantRole)
				}
				if claims.Email != tt.wantEmail {
					t.Errorf("ValidateToken() Email = %v, want %v", claims.Email, tt.wantEmail)
				}
			}
		})
	}
}

func TestPasswordValidationError(t *testing.T) {
	err := &PasswordValidationError{
		Errors: []string{"error1", "error2"},
	}

	if err.Error() != "password validation failed" {
		t.Errorf("PasswordValidationError.Error() = %v, want 'password validation failed'", err.Error())
	}

	if len(err.Errors) != 2 {
		t.Errorf("Expected 2 errors, got %d", len(err.Errors))
	}
}

func BenchmarkHashPassword(b *testing.B) {
	password := "SecurePassword123!"
	for i := 0; i < b.N; i++ {
		_, err := HashPassword(password)
		require.NoError(b, err)
	}
}

func BenchmarkCheckPasswordHash(b *testing.B) {
	password := "SecurePassword123!"
	hash, err := HashPassword(password)
	require.NoError(b, err)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CheckPasswordHash(password, hash)
	}
}

func BenchmarkValidatePassword(b *testing.B) {
	password := "SecurePassword123!"
	for i := 0; i < b.N; i++ {
		err := ValidatePassword(password)
		require.NoError(b, err)
	}
}

func BenchmarkGenerateToken(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := GenerateToken("12345", "test@gmail.com", "siswa", "Test User")
		require.NoError(b, err)
	}
}

func BenchmarkValidateToken(b *testing.B) {
	token, err := GenerateToken("12345", "test@gmail.com", "siswa", "Test User")
	require.NoError(b, err)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ValidateToken(token)
		require.NoError(b, err)
	}
}
