package errors

import (
	"errors"
	"fmt"
	"net/http"
)

// ErrorCode represents different types of business errors
type ErrorCode string

const (
	// Resource errors
	ErrCodeNotFound       ErrorCode = "NOT_FOUND"
	ErrCodeAlreadyExists  ErrorCode = "ALREADY_EXISTS"
	ErrCodeConflict       ErrorCode = "CONFLICT"
	
	// Permission errors
	ErrCodeForbidden      ErrorCode = "FORBIDDEN"
	ErrCodeUnauthorized   ErrorCode = "UNAUTHORIZED"
	
	// Input validation errors
	ErrCodeInvalidInput   ErrorCode = "INVALID_INPUT"
	ErrCodeInvalidFormat  ErrorCode = "INVALID_FORMAT"
	ErrCodeInvalidState   ErrorCode = "INVALID_STATE"
	
	// System errors
	ErrCodeInternal       ErrorCode = "INTERNAL_ERROR"
	ErrCodeTimeout        ErrorCode = "TIMEOUT"
	ErrCodeUnavailable    ErrorCode = "UNAVAILABLE"
	
	// Business logic errors
	ErrCodeBusinessRule   ErrorCode = "BUSINESS_RULE_VIOLATION"
	ErrCodeQuotaExceeded  ErrorCode = "QUOTA_EXCEEDED"
)

// BusinessError represents a structured business error
type BusinessError struct {
	Code       ErrorCode `json:"code"`
	Message    string    `json:"message"`
	Detail     string    `json:"detail,omitempty"`
	Field      string    `json:"field,omitempty"`
	Resource   string    `json:"resource,omitempty"`
	ResourceID string    `json:"resource_id,omitempty"`
	Cause      error     `json:"-"`
}

// Error implements the error interface
func (e *BusinessError) Error() string {
	if e.Detail != "" {
		return fmt.Sprintf("%s: %s (%s)", e.Code, e.Message, e.Detail)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying error
func (e *BusinessError) Unwrap() error {
	return e.Cause
}

// Is implements error comparison for errors.Is()
func (e *BusinessError) Is(target error) bool {
	if t, ok := target.(*BusinessError); ok {
		return e.Code == t.Code
	}
	return false
}

// HTTPStatusCode returns the appropriate HTTP status code for the error
func (e *BusinessError) HTTPStatusCode() int {
	switch e.Code {
	case ErrCodeNotFound:
		return http.StatusNotFound
	case ErrCodeAlreadyExists, ErrCodeConflict:
		return http.StatusConflict
	case ErrCodeForbidden:
		return http.StatusForbidden
	case ErrCodeUnauthorized:
		return http.StatusUnauthorized
	case ErrCodeInvalidInput, ErrCodeInvalidFormat, ErrCodeInvalidState:
		return http.StatusBadRequest
	case ErrCodeTimeout:
		return http.StatusRequestTimeout
	case ErrCodeUnavailable:
		return http.StatusServiceUnavailable
	case ErrCodeBusinessRule, ErrCodeQuotaExceeded:
		return http.StatusUnprocessableEntity
	default:
		return http.StatusInternalServerError
	}
}

// New creates a new BusinessError
func New(code ErrorCode, message string) *BusinessError {
	return &BusinessError{
		Code:    code,
		Message: message,
	}
}

// Newf creates a new BusinessError with formatted message
func Newf(code ErrorCode, format string, args ...interface{}) *BusinessError {
	return &BusinessError{
		Code:    code,
		Message: fmt.Sprintf(format, args...),
	}
}

// Wrap wraps an existing error with additional context
func Wrap(err error, code ErrorCode, message string) *BusinessError {
	return &BusinessError{
		Code:    code,
		Message: message,
		Cause:   err,
	}
}

// Wrapf wraps an existing error with formatted message
func Wrapf(err error, code ErrorCode, format string, args ...interface{}) *BusinessError {
	return &BusinessError{
		Code:    code,
		Message: fmt.Sprintf(format, args...),
		Cause:   err,
	}
}

// Predefined errors for common cases
var (
	ErrNotFound       = New(ErrCodeNotFound, "Resource not found")
	ErrAlreadyExists  = New(ErrCodeAlreadyExists, "Resource already exists")
	ErrForbidden      = New(ErrCodeForbidden, "Operation not permitted")
	ErrUnauthorized   = New(ErrCodeUnauthorized, "Authentication required")
	ErrInvalidInput   = New(ErrCodeInvalidInput, "Invalid input provided")
	ErrInternal       = New(ErrCodeInternal, "Internal server error")
	ErrTimeout        = New(ErrCodeTimeout, "Operation timed out")
	ErrUnavailable    = New(ErrCodeUnavailable, "Service unavailable")
)

// IsNotFound checks if an error is a "not found" error
func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}

// IsForbidden checks if an error is a "forbidden" error
func IsForbidden(err error) bool {
	return errors.Is(err, ErrForbidden)
}

// IsUnauthorized checks if an error is an "unauthorized" error
func IsUnauthorized(err error) bool {
	return errors.Is(err, ErrUnauthorized)
}

// IsInvalidInput checks if an error is an "invalid input" error
func IsInvalidInput(err error) bool {
	return errors.Is(err, ErrInvalidInput)
}

// IsBusinessError checks if an error is a BusinessError
func IsBusinessError(err error) bool {
	var be *BusinessError
	return errors.As(err, &be)
}

// AsBusinessError attempts to extract a BusinessError from an error
func AsBusinessError(err error) (*BusinessError, bool) {
	var be *BusinessError
	if errors.As(err, &be) {
		return be, true
	}
	return nil, false
}