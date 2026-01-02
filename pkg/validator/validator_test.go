package validator

import (
	"testing"
)

func TestIsValidPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		expected bool
	}{
		// Valid passwords (uppercase, lowercase, number, special char, 8+ chars)
		{"valid password", "Password1!", true},
		{"valid complex password", "MyP@ssw0rd", true},
		{"valid long password", "VerySecure123!@#", true},
		{"valid with underscore", "Pass_word1!", true},

		// Invalid - too short
		{"too short", "Pa1!", false},
		{"7 chars", "Pass1!a", false},

		// Invalid - missing uppercase
		{"no uppercase", "password1!", false},

		// Invalid - missing lowercase
		{"no lowercase", "PASSWORD1!", false},

		// Invalid - missing number
		{"no number", "Password!@", false},

		// Invalid - missing special character
		{"no special char", "Password123", false},

		// Edge cases
		{"empty password", "", false},
		{"only spaces", "        ", false},
		{"unicode password", "Pässword1!", true}, // Should work with unicode
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidPassword(tt.password)
			if got != tt.expected {
				t.Errorf("IsValidPassword(%q) = %v, want %v", tt.password, got, tt.expected)
			}
		})
	}
}

func TestIsValidEmail(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		expected bool
	}{
		// Valid emails
		{"simple email", "test@example.com", true},
		{"email with subdomain", "user@mail.example.com", true},
		{"email with plus", "user+tag@example.com", true},
		{"email with dots", "first.last@example.com", true},

		// Invalid emails
		{"empty email", "", false},
		{"no at symbol", "testexample.com", false},
		{"no domain", "test@", false},
		{"no local part", "@example.com", false},
		{"spaces", "test @example.com", false},
		{"double at", "test@@example.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidEmail(tt.email)
			if got != tt.expected {
				t.Errorf("IsValidEmail(%q) = %v, want %v", tt.email, got, tt.expected)
			}
		})
	}
}

func TestValidate_RegisterRequest(t *testing.T) {
	type RegisterRequest struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required,password"`
		Name     string `json:"name" validate:"required,min=2,max=100"`
	}

	tests := []struct {
		name        string
		req         RegisterRequest
		expectError bool
	}{
		{
			name: "valid request",
			req: RegisterRequest{
				Email:    "test@example.com",
				Password: "Password1!",
				Name:     "John Doe",
			},
			expectError: false,
		},
		{
			name: "weak password",
			req: RegisterRequest{
				Email:    "test@example.com",
				Password: "password",
				Name:     "John Doe",
			},
			expectError: true,
		},
		{
			name: "invalid email",
			req: RegisterRequest{
				Email:    "notanemail",
				Password: "Password1!",
				Name:     "John Doe",
			},
			expectError: true,
		},
		{
			name: "name too short",
			req: RegisterRequest{
				Email:    "test@example.com",
				Password: "Password1!",
				Name:     "J",
			},
			expectError: true,
		},
		{
			name: "empty fields",
			req: RegisterRequest{
				Email:    "",
				Password: "",
				Name:     "",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := Validate(&tt.req)
			hasError := errs != nil && len(errs) > 0
			if hasError != tt.expectError {
				t.Errorf("Validate() returned errors=%v, expectError=%v", errs, tt.expectError)
			}
		})
	}
}
