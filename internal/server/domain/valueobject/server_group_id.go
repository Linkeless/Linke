package valueobject

import (
	"errors"
	"strconv"
)

var (
	ErrInvalidServerGroupID = errors.New("invalid server group ID")
)

// ServerGroupID represents a server group identifier
type ServerGroupID struct {
	value uint
}

// NewServerGroupID creates a new ServerGroupID
func NewServerGroupID(value uint) (ServerGroupID, error) {
	if value == 0 {
		return ServerGroupID{}, ErrInvalidServerGroupID
	}
	
	return ServerGroupID{value: value}, nil
}

// Value returns the underlying value
func (id ServerGroupID) Value() uint {
	return id.value
}

// String returns string representation
func (id ServerGroupID) String() string {
	return strconv.FormatUint(uint64(id.value), 10)
}

// IsZero checks if the ID is zero value
func (id ServerGroupID) IsZero() bool {
	return id.value == 0
}

// Equals checks equality with another ServerGroupID
func (id ServerGroupID) Equals(other ServerGroupID) bool {
	return id.value == other.value
}