package validation

import (
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

// Validator provides validation capabilities
type Validator struct {
	validate *validator.Validate
}

// New creates a new Validator
func New() *Validator {
	v := validator.New()

	// Use JSON tag names for field names
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return fld.Name
		}
		return name
	})

	// Register custom validations
	registerCustomValidators(v)

	return &Validator{validate: v}
}

// Struct validates a struct using tags
func (v *Validator) Struct(obj any) *Result {
	result := NewResult()

	err := v.validate.Struct(obj)
	if err == nil {
		return result
	}

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			result.AddErrorWithValue(
				e.Field(),
				tagToCode(e.Tag()),
				tagToMessage(e),
				e.Value(),
			)
		}
	}

	return result
}

// Field validates a single field value
func (v *Validator) Field(value any, tag string) *Result {
	result := NewResult()

	err := v.validate.Var(value, tag)
	if err == nil {
		return result
	}

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			result.AddError("value", tagToCode(e.Tag()), tagToMessage(e))
		}
	}

	return result
}

// FieldWithName validates a field with a custom field name
func (v *Validator) FieldWithName(fieldName string, value any, tag string) *Result {
	result := NewResult()

	err := v.validate.Var(value, tag)
	if err == nil {
		return result
	}

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			result.AddError(fieldName, tagToCode(e.Tag()), tagToMessage(e))
		}
	}

	return result
}

// RegisterValidation registers a custom validation
func (v *Validator) RegisterValidation(tag string, fn validator.Func) error {
	return v.validate.RegisterValidation(tag, fn)
}

// registerCustomValidators registers custom validation rules
func registerCustomValidators(v *validator.Validate) {
	// Password validation: 8-128 chars, 1 uppercase, 1 lowercase, 1 number, 1 special char
	_ = v.RegisterValidation("password", func(fl validator.FieldLevel) bool {
		password := fl.Field().String()
		if len(password) < 8 || len(password) > 128 {
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

// tagToCode maps validation tags to error codes
func tagToCode(tag string) string {
	codes := map[string]string{
		"required":  "REQUIRED",
		"email":     "INVALID_EMAIL",
		"min":       "TOO_SHORT",
		"max":       "TOO_LONG",
		"len":       "INVALID_LENGTH",
		"password":  "WEAK_PASSWORD",
		"gt":        "TOO_SMALL",
		"gte":       "TOO_SMALL",
		"lt":        "TOO_LARGE",
		"lte":       "TOO_LARGE",
		"oneof":     "INVALID_CHOICE",
		"url":       "INVALID_URL",
		"uuid":      "INVALID_UUID",
		"alpha":     "INVALID_ALPHA",
		"alphanum":  "INVALID_ALPHANUM",
		"numeric":   "INVALID_NUMERIC",
		"eqfield":   "FIELD_MISMATCH",
		"nefield":   "FIELD_MATCH",
		"notempty":  "EMPTY",
		"unique":    "NOT_UNIQUE",
	}
	if code, ok := codes[tag]; ok {
		return code
	}
	return "INVALID"
}

// tagToMessage generates human-readable error messages
func tagToMessage(e validator.FieldError) string {
	messages := map[string]string{
		"required":  "field is required",
		"email":     "invalid email format",
		"min":       "value is too short",
		"max":       "value is too long",
		"len":       "invalid length",
		"password":  "password must be at least 8 characters with uppercase, lowercase, number, and special character",
		"gt":        "value must be greater than " + e.Param(),
		"gte":       "value must be at least " + e.Param(),
		"lt":        "value must be less than " + e.Param(),
		"lte":       "value must be at most " + e.Param(),
		"oneof":     "must be one of: " + e.Param(),
		"url":       "invalid URL format",
		"uuid":      "invalid UUID format",
		"alpha":     "must contain only alphabetic characters",
		"alphanum":  "must contain only alphanumeric characters",
		"numeric":   "must be a numeric value",
		"eqfield":   "must match " + e.Param(),
		"nefield":   "must not match " + e.Param(),
		"notempty":  "cannot be empty or whitespace",
		"unique":    "must be unique",
	}
	if msg, ok := messages[e.Tag()]; ok {
		return msg
	}
	return "invalid value"
}
