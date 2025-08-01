package valueobject

import (
	"errors"
	"strconv"
)

var (
	ErrInvalidServerPort = errors.New("port must be between 1 and 65535")
)

// ServerPort represents a server port number
type ServerPort struct {
	value int
}

// NewServerPort creates a new ServerPort
func NewServerPort(value int) (ServerPort, error) {
	if value < 1 || value > 65535 {
		return ServerPort{}, ErrInvalidServerPort
	}
	
	return ServerPort{value: value}, nil
}

// Value returns the underlying value
func (port ServerPort) Value() int {
	return port.value
}

// String returns string representation
func (port ServerPort) String() string {
	return strconv.Itoa(port.value)
}

// IsWellKnown checks if the port is in well-known range (1-1023)
func (port ServerPort) IsWellKnown() bool {
	return port.value >= 1 && port.value <= 1023
}

// IsRegistered checks if the port is in registered range (1024-49151)
func (port ServerPort) IsRegistered() bool {
	return port.value >= 1024 && port.value <= 49151
}

// IsDynamic checks if the port is in dynamic/private range (49152-65535)
func (port ServerPort) IsDynamic() bool {
	return port.value >= 49152 && port.value <= 65535
}

// Equals checks equality with another ServerPort
func (port ServerPort) Equals(other ServerPort) bool {
	return port.value == other.value
}