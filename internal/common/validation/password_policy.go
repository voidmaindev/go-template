// Package validation provides input validation utilities.
// It wraps github.com/go-playground/validator with custom rules
// and provides a Result type for collecting validation errors.
package validation

import (
	"fmt"
	"strings"
)

// Password validation constants.
const (
	// MinPasswordLength is the minimum required password length.
	MinPasswordLength = 8
	// MaxPasswordLength is the maximum allowed password length.
	MaxPasswordLength = 128
	// DefaultSpecialChars defines the allowed special characters in passwords.
	DefaultSpecialChars = "!@#$%^&*()_+-=[]{}|;':\",./<>?"
)

// PasswordPolicy defines the password validation rules.
// Use DefaultPasswordPolicy() to get the standard policy.
type PasswordPolicy struct {
	MinLength      int
	MaxLength      int
	RequireUpper   bool
	RequireLower   bool
	RequireNumber  bool
	RequireSpecial bool
	SpecialChars   string
}

// DefaultPasswordPolicy returns the default password policy.
func DefaultPasswordPolicy() *PasswordPolicy {
	return &PasswordPolicy{
		MinLength:      MinPasswordLength,
		MaxLength:      MaxPasswordLength,
		RequireUpper:   true,
		RequireLower:   true,
		RequireNumber:  true,
		RequireSpecial: true,
		SpecialChars:   DefaultSpecialChars,
	}
}

// Validate checks if password meets the policy requirements.
func (p *PasswordPolicy) Validate(password string) bool {
	if len(password) < p.MinLength || len(password) > p.MaxLength {
		return false
	}

	var hasUpper, hasLower, hasNumber, hasSpecial bool
	for _, char := range password {
		switch {
		case 'A' <= char && char <= 'Z':
			hasUpper = true
		case 'a' <= char && char <= 'z':
			hasLower = true
		case '0' <= char && char <= '9':
			hasNumber = true
		case strings.ContainsRune(p.SpecialChars, char):
			hasSpecial = true
		}
	}

	if p.RequireUpper && !hasUpper {
		return false
	}
	if p.RequireLower && !hasLower {
		return false
	}
	if p.RequireNumber && !hasNumber {
		return false
	}
	if p.RequireSpecial && !hasSpecial {
		return false
	}

	return true
}

// ErrorMessage returns a human-readable description of the requirements.
func (p *PasswordPolicy) ErrorMessage() string {
	var requirements []string

	requirements = append(requirements, fmt.Sprintf("at least %d characters", p.MinLength))

	if p.RequireUpper {
		requirements = append(requirements, "uppercase letter")
	}
	if p.RequireLower {
		requirements = append(requirements, "lowercase letter")
	}
	if p.RequireNumber {
		requirements = append(requirements, "number")
	}
	if p.RequireSpecial {
		requirements = append(requirements, "special character")
	}

	return "password must contain " + strings.Join(requirements, ", ")
}
