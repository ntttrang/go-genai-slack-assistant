package dto

// ValidationError represents a single field validation error
type ValidationError struct {
	Field   string
	Message string
}

// Validator accumulates validation errors
type Validator struct {
	errors []ValidationError
}

// Add adds a validation error for a field
func (v *Validator) Add(field, message string) {
	v.errors = append(v.errors, ValidationError{
		Field:   field,
		Message: message,
	})
}

// Valid returns true if there are no validation errors
func (v *Validator) Valid() bool {
	return len(v.errors) == 0
}

// Errors returns all accumulated validation errors
func (v *Validator) Errors() []ValidationError {
	return v.errors
}

// NewValidator creates a new validator
func NewValidator() *Validator {
	return &Validator{
		errors: []ValidationError{},
	}
}
