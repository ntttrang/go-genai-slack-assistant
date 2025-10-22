package model

import "fmt"

type ErrorType string

const (
	ErrorTypeValidation    ErrorType = "VALIDATION_ERROR"
	ErrorTypeNotFound      ErrorType = "NOT_FOUND"
	ErrorTypeRateLimit     ErrorType = "RATE_LIMIT"
	ErrorTypeInternalError ErrorType = "INTERNAL_ERROR"
	ErrorTypeUnauthorized  ErrorType = "UNAUTHORIZED"
	ErrorTypeBadRequest    ErrorType = "BAD_REQUEST"
)

type DomainError struct {
	Type    ErrorType
	Message string
	Cause   error
}

func (e *DomainError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (cause: %v)", e.Type, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

func NewValidationError(message string) *DomainError {
	return &DomainError{
		Type:    ErrorTypeValidation,
		Message: message,
	}
}

func NewNotFoundError(message string) *DomainError {
	return &DomainError{
		Type:    ErrorTypeNotFound,
		Message: message,
	}
}

func NewRateLimitError(message string) *DomainError {
	return &DomainError{
		Type:    ErrorTypeRateLimit,
		Message: message,
	}
}

func NewInternalError(message string, cause error) *DomainError {
	return &DomainError{
		Type:    ErrorTypeInternalError,
		Message: message,
		Cause:   cause,
	}
}

func NewUnauthorizedError(message string) *DomainError {
	return &DomainError{
		Type:    ErrorTypeUnauthorized,
		Message: message,
	}
}

func NewBadRequestError(message string) *DomainError {
	return &DomainError{
		Type:    ErrorTypeBadRequest,
		Message: message,
	}
}
