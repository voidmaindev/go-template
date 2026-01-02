package utils

import (
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestHashPassword(t *testing.T) {
	password := "testPassword123!"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}

	if hash == "" {
		t.Error("HashPassword() returned empty hash")
	}

	// Hash should be different from password
	if hash == password {
		t.Error("HashPassword() hash should not equal plain password")
	}

	// Should be valid bcrypt hash
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		t.Errorf("Hash is not valid bcrypt hash: %v", err)
	}
}

func TestHashPassword_DifferentHashes(t *testing.T) {
	password := "testPassword123!"

	hash1, _ := HashPassword(password)
	hash2, _ := HashPassword(password)

	// Same password should produce different hashes (due to salt)
	if hash1 == hash2 {
		t.Error("HashPassword() should produce different hashes for same password")
	}
}

func TestCheckPassword(t *testing.T) {
	password := "testPassword123!"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}

	tests := []struct {
		name     string
		password string
		expected bool
	}{
		{"correct password", password, true},
		{"wrong password", "wrongPassword", false},
		{"empty password", "", false},
		{"similar password", "testPassword123", false},
		{"case sensitive", "TESTPASSWORD123!", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CheckPassword(tt.password, hash)
			if result != tt.expected {
				t.Errorf("CheckPassword(%q) = %v, want %v", tt.password, result, tt.expected)
			}
		})
	}
}

func TestCheckPassword_InvalidHash(t *testing.T) {
	// Should return false for invalid hash, not panic
	result := CheckPassword("password", "not-a-valid-hash")
	if result {
		t.Error("CheckPassword() should return false for invalid hash")
	}
}

func TestHashPasswordWithCost(t *testing.T) {
	password := "testPassword123!"

	tests := []struct {
		name         string
		cost         int
		expectedCost int
	}{
		{"minimum cost", bcrypt.MinCost, bcrypt.MinCost},
		{"default cost", DefaultCost, DefaultCost},
		{"below minimum", 1, bcrypt.MinCost},
		// Note: "above maximum" test removed - bcrypt.MaxCost (31) is too slow for unit tests
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := HashPasswordWithCost(password, tt.cost)
			if err != nil {
				t.Fatalf("HashPasswordWithCost() error = %v", err)
			}

			// Verify cost
			cost, err := bcrypt.Cost([]byte(hash))
			if err != nil {
				t.Fatalf("bcrypt.Cost() error = %v", err)
			}

			if cost != tt.expectedCost {
				t.Errorf("Cost = %d, want %d", cost, tt.expectedCost)
			}

			// Verify password still matches
			if !CheckPassword(password, hash) {
				t.Error("CheckPassword() should return true for correct password")
			}
		})
	}
}

func TestHashPassword_EmptyPassword(t *testing.T) {
	// Empty password should still hash successfully
	hash, err := HashPassword("")
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}

	if !CheckPassword("", hash) {
		t.Error("CheckPassword() should return true for empty password")
	}

	if CheckPassword("not-empty", hash) {
		t.Error("CheckPassword() should return false for non-empty password")
	}
}

func TestHashPassword_LongPassword(t *testing.T) {
	// bcrypt has a 72 byte limit and returns error for longer passwords
	longPassword := string(make([]byte, 100))
	for i := range longPassword {
		longPassword = longPassword[:i] + "a" + longPassword[i+1:]
	}

	_, err := HashPassword(longPassword)
	if err == nil {
		t.Error("HashPassword() should return error for passwords > 72 bytes")
	}

	// Test that 72-byte password works
	maxPassword := string(make([]byte, 72))
	for i := range maxPassword {
		maxPassword = maxPassword[:i] + "b" + maxPassword[i+1:]
	}

	hash, err := HashPassword(maxPassword)
	if err != nil {
		t.Fatalf("HashPassword() should handle 72-byte password: %v", err)
	}

	if !CheckPassword(maxPassword, hash) {
		t.Error("CheckPassword() should work for 72-byte password")
	}
}

func TestHashPassword_UnicodePassword(t *testing.T) {
	password := "пароль123!Password"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}

	if !CheckPassword(password, hash) {
		t.Error("CheckPassword() should handle unicode passwords")
	}
}
