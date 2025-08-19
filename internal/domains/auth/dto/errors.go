package dto

import (
	"errors"
	"fmt"
	"time"
)

// AuthErrorType represents different types of authentication errors
type AuthErrorType string

const (
	ErrorTypeInvalidToken       AuthErrorType = "invalid_token"
	ErrorTypeExpiredToken       AuthErrorType = "expired_token"
	ErrorTypeRevokedToken       AuthErrorType = "revoked_token"
	ErrorTypeInvalidCredentials AuthErrorType = "invalid_credentials"
	ErrorTypeAccountLocked      AuthErrorType = "account_locked"
	ErrorTypeAccountInactive    AuthErrorType = "account_inactive"
	ErrorTypeSecretRotation     AuthErrorType = "secret_rotation"
	ErrorTypeBlacklistFailure   AuthErrorType = "blacklist_failure"
	ErrorTypeCacheFailure       AuthErrorType = "cache_failure"
	ErrorTypeValidationFailure  AuthErrorType = "validation_failure"
)

// AuthError represents a structured authentication error with context
type AuthError struct {
	Type      AuthErrorType          `json:"type"`
	Message   string                 `json:"message"`
	Details   string                 `json:"details,omitempty"`
	Context   map[string]interface{} `json:"context,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Cause     error                  `json:"-"` // Original error, not serialized
}

// Error implements the error interface
func (e *AuthError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("auth error [%s]: %s (%s)", e.Type, e.Message, e.Details)
	}
	return fmt.Sprintf("auth error [%s]: %s", e.Type, e.Message)
}

// Unwrap returns the underlying cause error
func (e *AuthError) Unwrap() error {
	return e.Cause
}

// Is checks if the error matches the target error type
func (e *AuthError) Is(target error) bool {
	if authErr, ok := target.(*AuthError); ok {
		return e.Type == authErr.Type
	}
	return errors.Is(e.Cause, target)
}

// NewAuthError creates a new structured auth error
func NewAuthError(errorType AuthErrorType, message string) *AuthError {
	return &AuthError{
		Type:      errorType,
		Message:   message,
		Timestamp: time.Now(),
		Context:   make(map[string]interface{}),
	}
}

// NewAuthErrorWithCause creates a new auth error with an underlying cause
func NewAuthErrorWithCause(errorType AuthErrorType, message string, cause error) *AuthError {
	return &AuthError{
		Type:      errorType,
		Message:   message,
		Cause:     cause,
		Timestamp: time.Now(),
		Context:   make(map[string]interface{}),
	}
}

// WithDetails adds additional details to the error
func (e *AuthError) WithDetails(details string) *AuthError {
	e.Details = details
	return e
}

// WithContext adds context information to the error
func (e *AuthError) WithContext(key string, value interface{}) *AuthError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// WithUserID adds user ID to the error context
func (e *AuthError) WithUserID(userID uint) *AuthError {
	return e.WithContext("user_id", userID)
}

// WithTokenHash adds token hash to the error context (for logging)
func (e *AuthError) WithTokenHash(tokenHash string) *AuthError {
	return e.WithContext("token_hash", tokenHash)
}

// WithIP adds IP address to the error context
func (e *AuthError) WithIP(ip string) *AuthError {
	return e.WithContext("ip", ip)
}

// Predefined error constructors for common auth errors
func ErrInvalidToken(details string) *AuthError {
	return NewAuthError(ErrorTypeInvalidToken, "Invalid token").WithDetails(details)
}

func ErrExpiredToken() *AuthError {
	return NewAuthError(ErrorTypeExpiredToken, "Token has expired")
}

func ErrRevokedToken() *AuthError {
	return NewAuthError(ErrorTypeRevokedToken, "Token has been revoked")
}

func ErrInvalidCredentials() *AuthError {
	return NewAuthError(ErrorTypeInvalidCredentials, "Invalid email or password")
}

func ErrAccountLocked(duration time.Duration) *AuthError {
	return NewAuthError(ErrorTypeAccountLocked, "Account is temporarily locked").
		WithDetails(fmt.Sprintf("locked for %v", duration))
}

func ErrAccountInactive(status string) *AuthError {
	return NewAuthError(ErrorTypeAccountInactive, "Account is inactive").
		WithDetails(fmt.Sprintf("status: %s", status))
}

// CleanupManager handles resource cleanup for auth operations
type CleanupManager struct {
	cleanupFuncs []func() error
}

// NewCleanupManager creates a new cleanup manager
func NewCleanupManager() *CleanupManager {
	return &CleanupManager{
		cleanupFuncs: make([]func() error, 0),
	}
}

// Add adds a cleanup function to be executed later
func (cm *CleanupManager) Add(cleanupFunc func() error) {
	cm.cleanupFuncs = append(cm.cleanupFuncs, cleanupFunc)
}

// AddSimple adds a simple cleanup function that doesn't return an error
func (cm *CleanupManager) AddSimple(cleanupFunc func()) {
	cm.Add(func() error {
		cleanupFunc()
		return nil
	})
}

// Execute runs all cleanup functions in reverse order (LIFO)
func (cm *CleanupManager) Execute() error {
	var firstError error

	// Execute in reverse order (last added, first executed)
	for i := len(cm.cleanupFuncs) - 1; i >= 0; i-- {
		if err := cm.cleanupFuncs[i](); err != nil && firstError == nil {
			firstError = err
		}
	}

	// Clear the cleanup functions after execution
	cm.cleanupFuncs = cm.cleanupFuncs[:0]

	return firstError
}

// Reset clears all cleanup functions without executing them
func (cm *CleanupManager) Reset() {
	cm.cleanupFuncs = cm.cleanupFuncs[:0]
}

// Count returns the number of registered cleanup functions
func (cm *CleanupManager) Count() int {
	return len(cm.cleanupFuncs)
}
