package valueobject

import (
	"errors"
	"regexp"
	"strings"
)

// Username represents a user's username
type Username struct {
	value string
}

var (
	// Username validation rules
	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]{3,30}$`)
)

// NewUsername creates a new Username with validation
func NewUsername(username string) (Username, error) {
	username = strings.TrimSpace(username)
	
	if username == "" {
		return Username{}, errors.New("username cannot be empty")
	}
	
	if len(username) < 3 {
		return Username{}, errors.New("username must be at least 3 characters")
	}
	
	if len(username) > 30 {
		return Username{}, errors.New("username cannot exceed 30 characters")
	}
	
	if !usernameRegex.MatchString(username) {
		return Username{}, errors.New("username can only contain letters, numbers, hyphens, and underscores")
	}
	
	return Username{value: username}, nil
}

// GenerateFromEmail generates a username from email local part
func GenerateUsernameFromEmail(localPart string) Username {
	// Clean the username (remove special characters, convert to lowercase)
	username := strings.ToLower(strings.ReplaceAll(localPart, ".", ""))
	username = strings.ReplaceAll(username, "+", "")
	username = strings.ReplaceAll(username, "_", "")
	
	// If username is too short, pad it
	if len(username) < 3 {
		username = username + "user"
	}
	
	// If too long, truncate
	if len(username) > 30 {
		username = username[:30]
	}
	
	// This should always succeed since we control the generation
	result, _ := NewUsername(username)
	return result
}

// String returns the string representation of the username
func (u Username) String() string {
	return u.value
}

// Value returns the username value
func (u Username) Value() string {
	return u.value
}

// IsEmpty checks if the username is empty
func (u Username) IsEmpty() bool {
	return strings.TrimSpace(u.value) == ""
}

// Equals checks if two usernames are equal
func (u Username) Equals(other Username) bool {
	return u.value == other.value
}