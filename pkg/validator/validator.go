package validator

import (
	"reflect"
	"strings"
	"sync"

	"github.com/go-playground/validator/v10"
)

var (
	validate *validator.Validate
	once     sync.Once
)

// Get returns the singleton validator instance
func Get() *validator.Validate {
	once.Do(func() {
		validate = validator.New()

		// Use JSON tag names for field names in error messages
		validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
			name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
			if name == "-" {
				return ""
			}
			return name
		})

		// Register custom validators
		registerCustomValidators(validate)
	})

	return validate
}

// Validate validates a struct and returns formatted errors
func Validate(s any) []ValidationError {
	err := Get().Struct(s)
	if err == nil {
		return nil
	}

	var errors []ValidationError
	for _, err := range err.(validator.ValidationErrors) {
		errors = append(errors, ValidationError{
			Field:   err.Field(),
			Tag:     err.Tag(),
			Value:   err.Value(),
			Message: getErrorMessage(err),
		})
	}

	return errors
}

// ValidationError represents a single validation error
type ValidationError struct {
	Field   string      `json:"field"`
	Tag     string      `json:"tag"`
	Value   any `json:"value,omitempty"`
	Message string      `json:"message"`
}

// getErrorMessage returns a human-readable error message for a validation error
func getErrorMessage(err validator.FieldError) string {
	switch err.Tag() {
	case "required":
		return err.Field() + " is required"
	case "email":
		return err.Field() + " must be a valid email address"
	case "min":
		return err.Field() + " must be at least " + err.Param() + " characters"
	case "max":
		return err.Field() + " must be at most " + err.Param() + " characters"
	case "len":
		return err.Field() + " must be exactly " + err.Param() + " characters"
	case "gte":
		return err.Field() + " must be greater than or equal to " + err.Param()
	case "lte":
		return err.Field() + " must be less than or equal to " + err.Param()
	case "gt":
		return err.Field() + " must be greater than " + err.Param()
	case "lt":
		return err.Field() + " must be less than " + err.Param()
	case "eqfield":
		return err.Field() + " must be equal to " + err.Param()
	case "nefield":
		return err.Field() + " must not be equal to " + err.Param()
	case "oneof":
		return err.Field() + " must be one of: " + err.Param()
	case "url":
		return err.Field() + " must be a valid URL"
	case "uuid":
		return err.Field() + " must be a valid UUID"
	case "alpha":
		return err.Field() + " must contain only alphabetic characters"
	case "alphanum":
		return err.Field() + " must contain only alphanumeric characters"
	case "numeric":
		return err.Field() + " must be a numeric value"
	case "password":
		return err.Field() + " must be at least 8 characters with uppercase, lowercase, number, and special character"
	default:
		return err.Field() + " is invalid"
	}
}

// registerCustomValidators registers custom validation rules
func registerCustomValidators(v *validator.Validate) {
	// Password validation: at least 8 chars, 1 uppercase, 1 lowercase, 1 number, 1 special char
	_ = v.RegisterValidation("password", func(fl validator.FieldLevel) bool {
		password := fl.Field().String()
		if len(password) < 8 {
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
			case strings.ContainsRune("!@#$%^&*()_+-=[]{}|;':\",./<>?", char):
				hasSpecial = true
			}
		}

		return hasUpper && hasLower && hasNumber && hasSpecial
	})

	// Non-empty string validation
	_ = v.RegisterValidation("notempty", func(fl validator.FieldLevel) bool {
		return strings.TrimSpace(fl.Field().String()) != ""
	})
}

// ValidateVar validates a single variable
func ValidateVar(field any, tag string) error {
	return Get().Var(field, tag)
}

// IsValidEmail checks if a string is a valid email
func IsValidEmail(email string) bool {
	return ValidateVar(email, "required,email") == nil
}

// IsValidPassword checks if a password meets requirements
func IsValidPassword(password string) bool {
	return ValidateVar(password, "required,password") == nil
}
