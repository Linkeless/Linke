package valueobject

import (
	"errors"
	"strings"
)

var (
	ErrEmptyServerName     = errors.New("server name cannot be empty")
	ErrServerNameTooLong   = errors.New("server name cannot exceed 255 characters")
	ErrInvalidServerName   = errors.New("server name contains invalid characters")
)

// ServerName represents a server name
type ServerName struct {
	value string
}

// NewServerName creates a new ServerName
func NewServerName(value string) (ServerName, error) {
	value = strings.TrimSpace(value)
	
	if value == "" {
		return ServerName{}, ErrEmptyServerName
	}
	
	if len(value) > 255 {
		return ServerName{}, ErrServerNameTooLong
	}
	
	// Basic validation - no control characters
	for _, r := range value {
		if r < 32 || r == 127 {
			return ServerName{}, ErrInvalidServerName
		}
	}
	
	return ServerName{value: value}, nil
}

// Value returns the underlying value
func (name ServerName) Value() string {
	return name.value
}

// String returns string representation
func (name ServerName) String() string {
	return name.value
}

// IsEmpty checks if the name is empty
func (name ServerName) IsEmpty() bool {
	return name.value == ""
}

// Equals checks equality with another ServerName
func (name ServerName) Equals(other ServerName) bool {
	return name.value == other.value
}