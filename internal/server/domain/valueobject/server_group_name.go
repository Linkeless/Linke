package valueobject

import (
	"errors"
	"strings"
)

var (
	ErrEmptyServerGroupName     = errors.New("server group name cannot be empty")
	ErrServerGroupNameTooLong   = errors.New("server group name cannot exceed 255 characters")
	ErrInvalidServerGroupName   = errors.New("server group name contains invalid characters")
)

// ServerGroupName represents a server group name
type ServerGroupName struct {
	value string
}

// NewServerGroupName creates a new ServerGroupName
func NewServerGroupName(value string) (ServerGroupName, error) {
	value = strings.TrimSpace(value)
	
	if value == "" {
		return ServerGroupName{}, ErrEmptyServerGroupName
	}
	
	if len(value) > 255 {
		return ServerGroupName{}, ErrServerGroupNameTooLong
	}
	
	// Basic validation - no control characters
	for _, r := range value {
		if r < 32 || r == 127 {
			return ServerGroupName{}, ErrInvalidServerGroupName
		}
	}
	
	return ServerGroupName{value: value}, nil
}

// Value returns the underlying value
func (name ServerGroupName) Value() string {
	return name.value
}

// String returns string representation
func (name ServerGroupName) String() string {
	return name.value
}

// IsEmpty checks if the name is empty
func (name ServerGroupName) IsEmpty() bool {
	return name.value == ""
}

// Equals checks equality with another ServerGroupName
func (name ServerGroupName) Equals(other ServerGroupName) bool {
	return name.value == other.value
}