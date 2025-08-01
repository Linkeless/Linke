package valueobject

import (
	"errors"
	"strings"
)

// UserRole represents the role of a user
type UserRole struct {
	value string
}

// User role constants
const (
	UserRoleUser  = "user"
	UserRoleAdmin = "admin"
)

// NewUserRole creates a new UserRole with validation
func NewUserRole(role string) (UserRole, error) {
	role = strings.TrimSpace(strings.ToLower(role))
	
	switch role {
	case UserRoleUser, UserRoleAdmin:
		return UserRole{value: role}, nil
	default:
		return UserRole{}, errors.New("invalid user role: must be user or admin")
	}
}

// String returns the string representation of the role
func (r UserRole) String() string {
	return r.value
}

// Value returns the role value
func (r UserRole) Value() string {
	return r.value
}

// IsUser checks if the role is user
func (r UserRole) IsUser() bool {
	return r.value == UserRoleUser
}

// IsAdmin checks if the role is admin
func (r UserRole) IsAdmin() bool {
	return r.value == UserRoleAdmin
}

// Equals checks if two roles are equal
func (r UserRole) Equals(other UserRole) bool {
	return r.value == other.value
}

// UserRole returns a user role
func User() UserRole {
	return UserRole{value: UserRoleUser}
}

// AdminRole returns an admin role
func Admin() UserRole {
	return UserRole{value: UserRoleAdmin}
}