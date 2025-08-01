package valueobject

import (
	"errors"
	"strings"
)

// UserStatus represents the status of a user
type UserStatus struct {
	value string
}

// User status constants
const (
	UserStatusActive   = "active"
	UserStatusInactive = "inactive"
	UserStatusBanned   = "banned"
)

// NewUserStatus creates a new UserStatus with validation
func NewUserStatus(status string) (UserStatus, error) {
	status = strings.TrimSpace(strings.ToLower(status))
	
	switch status {
	case UserStatusActive, UserStatusInactive, UserStatusBanned:
		return UserStatus{value: status}, nil
	default:
		return UserStatus{}, errors.New("invalid user status: must be active, inactive, or banned")
	}
}

// String returns the string representation of the status
func (s UserStatus) String() string {
	return s.value
}

// Value returns the status value
func (s UserStatus) Value() string {
	return s.value
}

// IsActive checks if the user status is active
func (s UserStatus) IsActive() bool {
	return s.value == UserStatusActive
}

// IsInactive checks if the user status is inactive
func (s UserStatus) IsInactive() bool {
	return s.value == UserStatusInactive
}

// IsBanned checks if the user status is banned
func (s UserStatus) IsBanned() bool {
	return s.value == UserStatusBanned
}

// IsSuspended checks if the user is suspended (banned or inactive)
func (s UserStatus) IsSuspended() bool {
	return s.value == UserStatusBanned || s.value == UserStatusInactive
}

// Equals checks if two statuses are equal
func (s UserStatus) Equals(other UserStatus) bool {
	return s.value == other.value
}

// Active returns an active status
func ActiveStatus() UserStatus {
	return UserStatus{value: UserStatusActive}
}

// InactiveStatus returns an inactive status
func InactiveStatus() UserStatus {
	return UserStatus{value: UserStatusInactive}
}

// BannedStatus returns a banned status
func BannedStatus() UserStatus {
	return UserStatus{value: UserStatusBanned}
}