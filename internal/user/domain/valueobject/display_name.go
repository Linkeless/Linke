package valueobject

import (
	"errors"
	"strings"
)

// DisplayName represents a user's display name
type DisplayName struct {
	value string
}

// NewDisplayName creates a new DisplayName with validation
func NewDisplayName(name string) (DisplayName, error) {
	name = strings.TrimSpace(name)
	
	if name == "" {
		return DisplayName{}, errors.New("display name cannot be empty")
	}
	
	if len(name) > 100 {
		return DisplayName{}, errors.New("display name cannot exceed 100 characters")
	}
	
	// Basic validation - no control characters
	for _, r := range name {
		if r < 32 && r != 9 { // Allow tab but no other control characters
			return DisplayName{}, errors.New("display name contains invalid characters")
		}
	}
	
	return DisplayName{value: name}, nil
}

// GenerateFromEmail generates a display name from email local part
func GenerateDisplayNameFromEmail(localPart string) DisplayName {
	if len(localPart) == 0 {
		return DisplayName{value: "User"}
	}
	
	// Capitalize first letter and keep reasonable length
	name := strings.ToUpper(string(localPart[0])) + localPart[1:]
	if len(name) > 100 {
		name = name[:100]
	}
	
	// This should always succeed since we control the generation
	result, _ := NewDisplayName(name)
	return result
}

// String returns the string representation of the display name
func (d DisplayName) String() string {
	return d.value
}

// Value returns the display name value
func (d DisplayName) Value() string {
	return d.value
}

// IsEmpty checks if the display name is empty
func (d DisplayName) IsEmpty() bool {
	return strings.TrimSpace(d.value) == ""
}

// Equals checks if two display names are equal
func (d DisplayName) Equals(other DisplayName) bool {
	return d.value == other.value
}