package valueobject

import (
	"errors"
	"net/mail"
	"strings"
)

// Email represents a validated email address
type Email struct {
	value string
}

// NewEmail creates a new Email with validation
func NewEmail(email string) (Email, error) {
	email = strings.TrimSpace(email)
	if email == "" {
		return Email{}, errors.New("email cannot be empty")
	}

	// Validate email format using Go's standard library
	if _, err := mail.ParseAddress(email); err != nil {
		return Email{}, errors.New("invalid email format")
	}

	// Additional checks
	if len(email) > 255 {
		return Email{}, errors.New("email address too long (max 255 characters)")
	}

	// Convert to lowercase for consistency
	email = strings.ToLower(email)

	return Email{value: email}, nil
}

// String returns the string representation of the email
func (e Email) String() string {
	return e.value
}

// Value returns the email value
func (e Email) Value() string {
	return e.value
}

// IsEmpty checks if the email is empty
func (e Email) IsEmpty() bool {
	return strings.TrimSpace(e.value) == ""
}

// Equals checks if two emails are equal
func (e Email) Equals(other Email) bool {
	return e.value == other.value
}

// Domain returns the domain part of the email
func (e Email) Domain() string {
	parts := strings.Split(e.value, "@")
	if len(parts) != 2 {
		return ""
	}
	return parts[1]
}

// LocalPart returns the local part of the email (before @)
func (e Email) LocalPart() string {
	parts := strings.Split(e.value, "@")
	if len(parts) != 2 {
		return ""
	}
	return parts[0]
}