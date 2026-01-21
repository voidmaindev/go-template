package validation

import (
	"context"
	"strconv"
	"strings"
)

// Builder allows composing validations
type Builder struct {
	result    *Result
	ctx       context.Context
	validator *Validator
}

// NewBuilder creates a validation builder
func NewBuilder(ctx context.Context) *Builder {
	return &Builder{
		result:    NewResult(),
		ctx:       ctx,
		validator: New(),
	}
}

// NewBuilderWithValidator creates a validation builder with a custom validator
func NewBuilderWithValidator(ctx context.Context, v *Validator) *Builder {
	return &Builder{
		result:    NewResult(),
		ctx:       ctx,
		validator: v,
	}
}

// Struct validates a struct and adds errors to result
func (b *Builder) Struct(obj any) *Builder {
	b.result.Merge(b.validator.Struct(obj))
	return b
}

// StructWith validates a struct with a specific validator
func (b *Builder) StructWith(v *Validator, obj any) *Builder {
	b.result.Merge(v.Struct(obj))
	return b
}

// Field validates a single field
func (b *Builder) Field(fieldName string, value any, tag string) *Builder {
	b.result.Merge(b.validator.FieldWithName(fieldName, value, tag))
	return b
}

// Custom adds a custom validation rule
func (b *Builder) Custom(field string, fn func(ctx context.Context) error) *Builder {
	if err := fn(b.ctx); err != nil {
		b.result.AddError(field, "CUSTOM", err.Error())
	}
	return b
}

// CustomWithCode adds a custom validation with specific code
func (b *Builder) CustomWithCode(field, code string, fn func(ctx context.Context) error) *Builder {
	if err := fn(b.ctx); err != nil {
		b.result.AddError(field, code, err.Error())
	}
	return b
}

// CustomAsync adds an async custom validation (for db lookups, etc.)
func (b *Builder) CustomAsync(field, code string, fn func(ctx context.Context) (bool, string)) *Builder {
	valid, msg := fn(b.ctx)
	if !valid {
		b.result.AddError(field, code, msg)
	}
	return b
}

// Condition adds error if condition is true
func (b *Builder) Condition(condition bool, field, code, message string) *Builder {
	if condition {
		b.result.AddError(field, code, message)
	}
	return b
}

// Required ensures field is not empty
func (b *Builder) Required(field string, value any) *Builder {
	switch v := value.(type) {
	case string:
		if v == "" {
			b.result.AddError(field, "REQUIRED", "field is required")
		}
	case nil:
		b.result.AddError(field, "REQUIRED", "field is required")
	case *string:
		if v == nil || *v == "" {
			b.result.AddError(field, "REQUIRED", "field is required")
		}
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		// Numbers are considered present (0 is valid)
	case float32, float64:
		// Floats are considered present
	default:
		// For other types, check if nil
		if value == nil {
			b.result.AddError(field, "REQUIRED", "field is required")
		}
	}
	return b
}

// MinLength validates minimum string length
func (b *Builder) MinLength(field, value string, min int) *Builder {
	if len(value) < min {
		b.result.AddError(field, "TOO_SHORT", "must be at least "+strconv.Itoa(min)+" characters")
	}
	return b
}

// MaxLength validates maximum string length
func (b *Builder) MaxLength(field, value string, max int) *Builder {
	if len(value) > max {
		b.result.AddError(field, "TOO_LONG", "must be at most "+strconv.Itoa(max)+" characters")
	}
	return b
}

// Range validates numeric range
func (b *Builder) Range(field string, value, min, max int) *Builder {
	if value < min || value > max {
		b.result.AddError(field, "OUT_OF_RANGE", "must be between "+strconv.Itoa(min)+" and "+strconv.Itoa(max))
	}
	return b
}

// OneOf validates value is in allowed list
func (b *Builder) OneOf(field, value string, allowed []string) *Builder {
	for _, a := range allowed {
		if value == a {
			return b
		}
	}
	b.result.AddError(field, "INVALID_CHOICE", "must be one of: "+strings.Join(allowed, ", "))
	return b
}

// WhenNotEmpty runs validation only if value is not empty
func (b *Builder) WhenNotEmpty(field, value string, fn func(*Builder)) *Builder {
	if value != "" {
		fn(b)
	}
	return b
}

// WhenPresent runs validation only if pointer is not nil
func (b *Builder) WhenPresent(field string, value any, fn func(*Builder)) *Builder {
	if value != nil {
		fn(b)
	}
	return b
}

// Result returns the validation result
func (b *Builder) Result() *Result {
	return b.result
}

// Valid returns true if validation passed
func (b *Builder) Valid() bool {
	return b.result.Valid()
}

// Invalid returns true if validation failed
func (b *Builder) Invalid() bool {
	return b.result.Invalid()
}

// Errors returns all validation errors
func (b *Builder) Errors() []FieldError {
	return b.result.Errors()
}

