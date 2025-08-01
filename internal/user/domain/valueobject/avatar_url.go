package valueobject

import (
	"errors"
	"net/url"
	"strings"
)

// AvatarURL represents a user's avatar URL
type AvatarURL struct {
	value string
}

// NewAvatarURL creates a new AvatarURL with validation
func NewAvatarURL(avatarURL string) (AvatarURL, error) {
	avatarURL = strings.TrimSpace(avatarURL)
	
	// Empty avatar URL is allowed (user has no avatar)
	if avatarURL == "" {
		return AvatarURL{value: ""}, nil
	}
	
	// Validate URL format
	parsedURL, err := url.Parse(avatarURL)
	if err != nil {
		return AvatarURL{}, errors.New("invalid avatar URL format")
	}
	
	// Must be HTTP or HTTPS
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return AvatarURL{}, errors.New("avatar URL must use HTTP or HTTPS scheme")
	}
	
	// Must have a host
	if parsedURL.Host == "" {
		return AvatarURL{}, errors.New("avatar URL must have a valid host")
	}
	
	return AvatarURL{value: avatarURL}, nil
}

// NewEmptyAvatarURL creates an empty avatar URL
func NewEmptyAvatarURL() AvatarURL {
	return AvatarURL{value: ""}
}

// String returns the string representation of the avatar URL
func (a AvatarURL) String() string {
	return a.value
}

// Value returns the avatar URL value
func (a AvatarURL) Value() string {
	return a.value
}

// IsEmpty checks if the avatar URL is empty
func (a AvatarURL) IsEmpty() bool {
	return strings.TrimSpace(a.value) == ""
}

// HasAvatar checks if the user has an avatar
func (a AvatarURL) HasAvatar() bool {
	return !a.IsEmpty()
}

// Equals checks if two avatar URLs are equal
func (a AvatarURL) Equals(other AvatarURL) bool {
	return a.value == other.value
}