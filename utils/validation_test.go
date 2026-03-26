package utils

import (
	"testing"
)

func TestValidator_Required(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		hasError bool
	}{
		{"empty string", "", true},
		{"whitespace only", "   ", true},
		{"valid value", "test", false},
		{"value with spaces", "  test  ", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator()
			v.Required("field", tt.value)
			if v.HasErrors() != tt.hasError {
				t.Errorf("Required(%q) hasError = %v, want %v", tt.value, v.HasErrors(), tt.hasError)
			}
		})
	}
}

func TestValidator_Email(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		hasError bool
	}{
		{"empty string", "", false}, // Empty is valid (use Required for mandatory)
		{"valid email", "test@example.com", false},
		{"valid gmail", "user@gmail.com", false},
		{"missing @", "testexample.com", true},
		{"missing domain", "test@", true},
		{"missing TLD", "test@example", true},
		{"special chars", "test+tag@example.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator()
			v.Email("email", tt.value)
			if v.HasErrors() != tt.hasError {
				t.Errorf("Email(%q) hasError = %v, want %v", tt.value, v.HasErrors(), tt.hasError)
			}
		})
	}
}

func TestValidator_NIS(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		hasError bool
	}{
		{"empty string", "", false},
		{"valid 5 digits", "12345", false},
		{"valid 10 digits", "1234567890", false},
		{"valid 20 digits", "12345678901234567890", false},
		{"too short", "1234", true},
		{"too long", "123456789012345678901", true},
		{"contains letters", "12345abc", true},
		{"contains special", "12345-678", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator()
			v.NIS("nis", tt.value)
			if v.HasErrors() != tt.hasError {
				t.Errorf("NIS(%q) hasError = %v, want %v", tt.value, v.HasErrors(), tt.hasError)
			}
		})
	}
}

func TestValidator_NIP(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		hasError bool
	}{
		{"empty string", "", false},
		{"valid 18 digits", "198501012010011001", false},
		{"too short", "12345678901234567", true},
		{"too long", "1234567890123456789", true},
		{"contains letters", "12345678901234567a", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator()
			v.NIP("nip", tt.value)
			if v.HasErrors() != tt.hasError {
				t.Errorf("NIP(%q) hasError = %v, want %v", tt.value, v.HasErrors(), tt.hasError)
			}
		})
	}
}

func TestValidator_Gender(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		hasError bool
	}{
		{"empty string", "", false},
		{"L", "L", false},
		{"P", "P", false},
		{"Laki-laki", "Laki-laki", false},
		{"Perempuan", "Perempuan", false},
		{"lowercase l", "l", false},
		{"invalid", "M", true},
		{"invalid word", "Male", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator()
			v.Gender("gender", tt.value)
			if v.HasErrors() != tt.hasError {
				t.Errorf("Gender(%q) hasError = %v, want %v", tt.value, v.HasErrors(), tt.hasError)
			}
		})
	}
}

func TestValidator_Date(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		hasError bool
	}{
		{"empty string", "", false},
		{"valid date", "2024-01-15", false},
		{"invalid format", "15-01-2024", true},
		{"invalid separator", "2024/01/15", true},
		{"missing parts", "2024-01", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator()
			v.Date("date", tt.value)
			if v.HasErrors() != tt.hasError {
				t.Errorf("Date(%q) hasError = %v, want %v", tt.value, v.HasErrors(), tt.hasError)
			}
		})
	}
}

func TestValidator_Time(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		hasError bool
	}{
		{"empty string", "", false},
		{"HH:MM format", "12:30", false},
		{"HH:MM:SS format", "12:30:45", false},
		{"midnight", "00:00", false},
		{"end of day", "23:59:59", false},
		{"invalid hour", "25:00", true},
		{"invalid minute", "12:60", true},
		{"wrong format", "12.30", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator()
			v.Time("time", tt.value)
			if v.HasErrors() != tt.hasError {
				t.Errorf("Time(%q) hasError = %v, want %v", tt.value, v.HasErrors(), tt.hasError)
			}
		})
	}
}

func TestValidator_Username(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		hasError bool
	}{
		{"empty string", "", false},
		{"valid username", "john_doe", false},
		{"valid with numbers", "admin123", false},
		{"starts with number", "123admin", true},
		{"too short", "ab", true},
		{"with special chars", "john@doe", true},
		{"with spaces", "john doe", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator()
			v.Username("username", tt.value)
			if v.HasErrors() != tt.hasError {
				t.Errorf("Username(%q) hasError = %v, want %v", tt.value, v.HasErrors(), tt.hasError)
			}
		})
	}
}

func TestValidator_InList(t *testing.T) {
	allowed := []string{"admin", "guru", "wali_kelas", "siswa"}

	tests := []struct {
		name     string
		value    string
		hasError bool
	}{
		{"empty string", "", false},
		{"valid admin", "admin", false},
		{"valid siswa", "siswa", false},
		{"invalid role", "superadmin", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator()
			v.InList("role", tt.value, allowed)
			if v.HasErrors() != tt.hasError {
				t.Errorf("InList(%q) hasError = %v, want %v", tt.value, v.HasErrors(), tt.hasError)
			}
		})
	}
}

func TestValidator_PositiveInt(t *testing.T) {
	tests := []struct {
		name     string
		value    int
		hasError bool
	}{
		{"positive", 5, false},
		{"zero", 0, true},
		{"negative", -1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator()
			v.PositiveInt("num", tt.value)
			if v.HasErrors() != tt.hasError {
				t.Errorf("PositiveInt(%d) hasError = %v, want %v", tt.value, v.HasErrors(), tt.hasError)
			}
		})
	}
}

func TestValidator_Chaining(t *testing.T) {
	v := NewValidator()
	v.Required("email", "test@example.com").
		Email("email", "test@example.com").
		Required("username", "admin").
		Username("username", "admin")

	if v.HasErrors() {
		t.Errorf("Chained validation should not have errors: %v", v.Errors())
	}
}

func TestValidator_MultipleErrors(t *testing.T) {
	v := NewValidator()
	v.Required("email", "").
		Required("username", "").
		Required("password", "")

	if !v.HasErrors() {
		t.Error("Should have errors")
	}

	errors := v.Errors()
	if len(errors) != 3 {
		t.Errorf("Expected 3 errors, got %d", len(errors))
	}
}

func TestValidationErrors_Error(t *testing.T) {
	v := NewValidator()
	v.Required("email", "").Required("username", "")

	errStr := v.Errors().Error()
	if errStr == "" {
		t.Error("Error string should not be empty")
	}
}

func TestValidateNIS(t *testing.T) {
	if err := ValidateNIS("12345"); err != nil {
		t.Errorf("ValidateNIS should pass for valid NIS: %v", err)
	}

	if err := ValidateNIS(""); err == nil {
		t.Error("ValidateNIS should fail for empty NIS")
	}

	if err := ValidateNIS("abc"); err == nil {
		t.Error("ValidateNIS should fail for invalid NIS")
	}
}

func TestValidateEmail(t *testing.T) {
	if err := ValidateEmail("test@example.com"); err != nil {
		t.Errorf("ValidateEmail should pass for valid email: %v", err)
	}

	if err := ValidateEmail(""); err == nil {
		t.Error("ValidateEmail should fail for empty email")
	}

	if err := ValidateEmail("invalid"); err == nil {
		t.Error("ValidateEmail should fail for invalid email")
	}
}

func TestValidateUsername(t *testing.T) {
	if err := ValidateUsername("admin123"); err != nil {
		t.Errorf("ValidateUsername should pass for valid username: %v", err)
	}

	if err := ValidateUsername(""); err == nil {
		t.Error("ValidateUsername should fail for empty username")
	}

	if err := ValidateUsername("ab"); err == nil {
		t.Error("ValidateUsername should fail for short username")
	}
}
