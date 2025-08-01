package valueobject

import (
	"fmt"
	"strconv"
)

// PaymentID represents a unique payment identifier
type PaymentID struct {
	value uint
}

// NewPaymentID creates a new PaymentID from uint
func NewPaymentID(value uint) (PaymentID, error) {
	if value == 0 {
		return PaymentID{}, fmt.Errorf("payment ID cannot be zero")
	}
	return PaymentID{value: value}, nil
}

// Value returns the underlying uint value
func (id PaymentID) Value() uint {
	return id.value
}

// String returns string representation
func (id PaymentID) String() string {
	return strconv.FormatUint(uint64(id.value), 10)
}

// Equals checks if two PaymentIDs are equal
func (id PaymentID) Equals(other PaymentID) bool {
	return id.value == other.value
}

// IsZero checks if the ID is zero value
func (id PaymentID) IsZero() bool {
	return id.value == 0
}