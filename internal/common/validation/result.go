package validation

// FieldError represents a validation error for a specific field
type FieldError struct {
	Field   string `json:"field"`
	Code    string `json:"code"`
	Message string `json:"message"`
	Value   any    `json:"value,omitempty"`
}

// Result holds validation results
type Result struct {
	errors []FieldError
}

// NewResult creates an empty validation result
func NewResult() *Result {
	return &Result{errors: make([]FieldError, 0)}
}

// AddError adds a field error
func (r *Result) AddError(field, code, message string) *Result {
	r.errors = append(r.errors, FieldError{
		Field:   field,
		Code:    code,
		Message: message,
	})
	return r
}

// AddErrorWithValue adds a field error with the invalid value
func (r *Result) AddErrorWithValue(field, code, message string, value any) *Result {
	r.errors = append(r.errors, FieldError{
		Field:   field,
		Code:    code,
		Message: message,
		Value:   value,
	})
	return r
}

// AddFieldError adds a FieldError directly
func (r *Result) AddFieldError(fe FieldError) *Result {
	r.errors = append(r.errors, fe)
	return r
}

// Merge combines results from another validation
func (r *Result) Merge(other *Result) *Result {
	if other != nil {
		r.errors = append(r.errors, other.errors...)
	}
	return r
}

// Valid returns true if no errors
func (r *Result) Valid() bool {
	return len(r.errors) == 0
}

// Invalid returns true if there are errors
func (r *Result) Invalid() bool {
	return len(r.errors) > 0
}

// Errors returns all field errors
func (r *Result) Errors() []FieldError {
	return r.errors
}

// ErrorCount returns the number of errors
func (r *Result) ErrorCount() int {
	return len(r.errors)
}

// First returns the first error or nil
func (r *Result) First() *FieldError {
	if len(r.errors) > 0 {
		return &r.errors[0]
	}
	return nil
}

// FieldErrors returns errors for a specific field
func (r *Result) FieldErrors(field string) []FieldError {
	var fieldErrors []FieldError
	for _, e := range r.errors {
		if e.Field == field {
			fieldErrors = append(fieldErrors, e)
		}
	}
	return fieldErrors
}

// HasFieldError checks if a specific field has errors
func (r *Result) HasFieldError(field string) bool {
	for _, e := range r.errors {
		if e.Field == field {
			return true
		}
	}
	return false
}

// Clear removes all errors
func (r *Result) Clear() *Result {
	r.errors = make([]FieldError, 0)
	return r
}

// ToMap converts errors to a map for JSON serialization
func (r *Result) ToMap() map[string][]FieldError {
	result := make(map[string][]FieldError)
	for _, e := range r.errors {
		result[e.Field] = append(result[e.Field], e)
	}
	return result
}
