package valueobject

import (
	"errors"
	"strconv"
)

var (
	ErrInvalidServerID = errors.New("invalid server ID")
)

// ServerID represents a shadowsocks server identifier
type ServerID struct {
	value int
}

// NewServerID creates a new ServerID
func NewServerID(value int) (ServerID, error) {
	if value <= 0 {
		return ServerID{}, ErrInvalidServerID
	}
	
	return ServerID{value: value}, nil
}

// Value returns the underlying value
func (id ServerID) Value() int {
	return id.value
}

// String returns string representation
func (id ServerID) String() string {
	return strconv.Itoa(id.value)
}

// IsZero checks if the ID is zero value
func (id ServerID) IsZero() bool {
	return id.value == 0
}

// Equals checks equality with another ServerID
func (id ServerID) Equals(other ServerID) bool {
	return id.value == other.value
}