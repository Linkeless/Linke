package valueobject

import (
	"errors"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

// UserID represents a unique identifier for a user
type UserID struct {
	value string
}

// NewUserID creates a new UserID with a UUID v4
func NewUserID() UserID {
	return UserID{value: uuid.New().String()}
}

// NewUserIDFromString creates a UserID from an existing string
func NewUserIDFromString(id string) (UserID, error) {
	if strings.TrimSpace(id) == "" {
		return UserID{}, errors.New("user ID cannot be empty")
	}

	// Validate UUID format if it looks like a UUID
	if strings.Contains(id, "-") {
		if _, err := uuid.Parse(id); err != nil {
			return UserID{}, errors.New("invalid UUID format")
		}
	}

	return UserID{value: id}, nil
}

// NewUserIDFromUint creates a UserID from a uint (for legacy compatibility)
func NewUserIDFromUint(id uint) UserID {
	return UserID{value: strconv.FormatUint(uint64(id), 10)}
}

// String returns the string representation of the UserID
func (u UserID) String() string {
	return u.value
}

// IsEmpty checks if the UserID is empty
func (u UserID) IsEmpty() bool {
	return strings.TrimSpace(u.value) == ""
}

// Equals checks if two UserIDs are equal
func (u UserID) Equals(other UserID) bool {
	return u.value == other.value
}

// ToUint converts UserID to uint (for legacy compatibility)
// Returns 0 if the UserID is not a valid uint
func (u UserID) ToUint() uint {
	if val, err := strconv.ParseUint(u.value, 10, 32); err == nil {
		return uint(val)
	}
	return 0
}

// IsUUID checks if the UserID is in UUID format
func (u UserID) IsUUID() bool {
	_, err := uuid.Parse(u.value)
	return err == nil
}