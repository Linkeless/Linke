package valueobject

import (
	"fmt"
	"strconv"
)

// PaymentConfigID represents a unique payment config identifier
type PaymentConfigID struct {
	value uint
}

// NewPaymentConfigID creates a new PaymentConfigID from uint
func NewPaymentConfigID(value uint) (PaymentConfigID, error) {
	if value == 0 {
		return PaymentConfigID{}, fmt.Errorf("payment config ID cannot be zero")
	}
	return PaymentConfigID{value: value}, nil
}

// Value returns the underlying uint value
func (id PaymentConfigID) Value() uint {
	return id.value
}

// String returns string representation
func (id PaymentConfigID) String() string {
	return strconv.FormatUint(uint64(id.value), 10)
}

// Equals checks if two PaymentConfigIDs are equal
func (id PaymentConfigID) Equals(other PaymentConfigID) bool {
	return id.value == other.value
}

// IsZero checks if the ID is zero value
func (id PaymentConfigID) IsZero() bool {
	return id.value == 0
}