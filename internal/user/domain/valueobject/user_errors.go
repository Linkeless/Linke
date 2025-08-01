package valueobject

import (
	"errors"
	"fmt"
)

// UserDomainError represents domain-specific errors for the User aggregate
type UserDomainError struct {
	Code    string
	Message string
	Field   string
}

func (e UserDomainError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("user domain error [%s] on field '%s': %s", e.Code, e.Field, e.Message)
	}
	return fmt.Sprintf("user domain error [%s]: %s", e.Code, e.Message)
}

// Common domain errors
var (
	// Authentication errors
	ErrInvalidCredentials        = UserDomainError{Code: "INVALID_CREDENTIALS", Message: "invalid credentials"}
	ErrAccountSuspended         = UserDomainError{Code: "ACCOUNT_SUSPENDED", Message: "user account is suspended"}
	ErrAccountInactive          = UserDomainError{Code: "ACCOUNT_INACTIVE", Message: "user account is not active"}
	ErrOAuthPasswordAuth        = UserDomainError{Code: "OAUTH_PASSWORD_AUTH", Message: "password authentication not supported for OAuth accounts"}
	ErrPasswordChangeNotAllowed = UserDomainError{Code: "PASSWORD_CHANGE_NOT_ALLOWED", Message: "password change not supported for OAuth accounts"}
	ErrCurrentPasswordIncorrect = UserDomainError{Code: "CURRENT_PASSWORD_INCORRECT", Message: "current password is incorrect"}
	
	// Profile errors
	ErrUsernameAlreadyExists    = UserDomainError{Code: "USERNAME_EXISTS", Message: "username already exists", Field: "username"}
	ErrEmailAlreadyExists       = UserDomainError{Code: "EMAIL_EXISTS", Message: "email already exists", Field: "email"}
	
	// OAuth errors
	ErrOAuthAccountExists       = UserDomainError{Code: "OAUTH_ACCOUNT_EXISTS", Message: "OAuth account already exists for this provider"}
	ErrOAuthAccountNotFound     = UserDomainError{Code: "OAUTH_ACCOUNT_NOT_FOUND", Message: "OAuth account not found for this provider"}
	
	// Business rule errors
	ErrCannotDeleteLastAdmin    = UserDomainError{Code: "CANNOT_DELETE_LAST_ADMIN", Message: "cannot delete the last admin user"}
	ErrCannotSuspendSelf        = UserDomainError{Code: "CANNOT_SUSPEND_SELF", Message: "cannot suspend your own account"}
	
	// General errors
	ErrUserNotFound            = UserDomainError{Code: "USER_NOT_FOUND", Message: "user not found"}
	ErrUserDeleted             = UserDomainError{Code: "USER_DELETED", Message: "user is deleted"}
)

// NewFieldValidationError creates a validation error for a specific field
func NewFieldValidationError(field, message string) UserDomainError {
	return UserDomainError{
		Code:    "FIELD_VALIDATION",
		Message: message,
		Field:   field,
	}
}

// NewBusinessRuleViolationError creates a business rule violation error
func NewBusinessRuleViolationError(rule, message string) UserDomainError {
	return UserDomainError{
		Code:    "BUSINESS_RULE_VIOLATION",
		Message: fmt.Sprintf("rule '%s': %s", rule, message),
	}
}

// IsUserDomainError checks if an error is a UserDomainError
func IsUserDomainError(err error) bool {
	var domainErr UserDomainError
	return errors.As(err, &domainErr)
}

// GetUserDomainError extracts the UserDomainError from an error
func GetUserDomainError(err error) (UserDomainError, bool) {
	var domainErr UserDomainError
	if errors.As(err, &domainErr) {
		return domainErr, true
	}
	return UserDomainError{}, false
}