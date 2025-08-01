package valueobject

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

// UserID represents a unique identifier for a user
// It supports both UUID format (for new users) and uint format (for legacy compatibility)
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
			return UserID{}, fmt.Errorf("invalid UUID format: %w", err)
		}
	}

	return UserID{value: id}, nil
}

// NewUserIDFromUint creates a UserID from a uint (for legacy compatibility)
func NewUserIDFromUint(id uint) (UserID, error) {
	if id == 0 {
		return UserID{}, errors.New("user ID cannot be zero")
	}
	return UserID{value: strconv.FormatUint(uint64(id), 10)}, nil
}

// NewUserIDFromUint64 creates a UserID from a uint64 (for legacy compatibility)
func NewUserIDFromUint64(id uint64) (UserID, error) {
	if id == 0 {
		return UserID{}, errors.New("user ID cannot be zero")
	}
	return UserID{value: strconv.FormatUint(id, 10)}, nil
}

// String returns the string representation of the UserID
func (u UserID) String() string {
	return u.value
}

// IsEmpty checks if the UserID is empty
func (u UserID) IsEmpty() bool {
	return strings.TrimSpace(u.value) == ""
}

// IsZero checks if the UserID is zero (alias for IsEmpty for consistency)
func (u UserID) IsZero() bool {
	return u.IsEmpty()
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

// ToUint64 converts UserID to uint64 (for legacy compatibility)
// Returns 0 if the UserID is not a valid uint64
func (u UserID) ToUint64() uint64 {
	if val, err := strconv.ParseUint(u.value, 10, 64); err == nil {
		return val
	}
	return 0
}

// Value returns the underlying value (for database compatibility)
func (u UserID) Value() interface{} {
	return u.value
}

// IsUUID checks if the UserID is in UUID format
func (u UserID) IsUUID() bool {
	_, err := uuid.Parse(u.value)
	return err == nil
}

// IsNumeric checks if the UserID is in numeric format
func (u UserID) IsNumeric() bool {
	_, err := strconv.ParseUint(u.value, 10, 64)
	return err == nil
}

// MarshalJSON implements json.Marshaler
func (u UserID) MarshalJSON() ([]byte, error) {
	// For numeric IDs, return as number; for UUIDs, return as string
	if u.IsNumeric() {
		return []byte(u.value), nil
	}
	return []byte(fmt.Sprintf(`"%s"`, u.value)), nil
}

// UnmarshalJSON implements json.Unmarshaler
func (u *UserID) UnmarshalJSON(data []byte) error {
	str := string(data)
	if str == "null" {
		u.value = ""
		return nil
	}
	
	// Remove quotes if present
	if len(str) >= 2 && str[0] == '"' && str[len(str)-1] == '"' {
		str = str[1 : len(str)-1]
	}
	
	if str == "" {
		return errors.New("user ID cannot be empty")
	}
	
	u.value = str
	return nil
}