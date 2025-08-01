package valueobject

import (
	"errors"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// Password represents a hashed password
type Password struct {
	hash string
}

// NewPassword creates a new Password by hashing the plaintext password
func NewPassword(plaintext string) (Password, error) {
	if strings.TrimSpace(plaintext) == "" {
		return Password{}, errors.New("password cannot be empty")
	}

	if len(plaintext) < 6 {
		return Password{}, errors.New("password must be at least 6 characters long")
	}

	if len(plaintext) > 128 {
		return Password{}, errors.New("password too long (max 128 characters)")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(plaintext), bcrypt.DefaultCost)
	if err != nil {
		return Password{}, errors.New("failed to hash password")
	}

	return Password{hash: string(hash)}, nil
}

// NewPasswordFromHash creates a Password from an existing hash
func NewPasswordFromHash(hash string) (Password, error) {
	if strings.TrimSpace(hash) == "" {
		return Password{}, errors.New("password hash cannot be empty")
	}

	return Password{hash: hash}, nil
}

// Hash returns the password hash
func (p Password) Hash() string {
	return p.hash
}

// Verify checks if the given plaintext password matches the stored hash
func (p Password) Verify(plaintext string) bool {
	if strings.TrimSpace(plaintext) == "" {
		return false
	}

	err := bcrypt.CompareHashAndPassword([]byte(p.hash), []byte(plaintext))
	return err == nil
}

// IsEmpty checks if the password is empty
func (p Password) IsEmpty() bool {
	return strings.TrimSpace(p.hash) == ""
}

// Equals checks if two passwords have the same hash
func (p Password) Equals(other Password) bool {
	return p.hash == other.hash
}