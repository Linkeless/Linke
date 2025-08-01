package valueobject

import (
	"fmt"
	"strconv"
)

// InvoiceID represents a unique identifier for an invoice
type InvoiceID struct {
	value uint
}

// NewInvoiceID creates a new InvoiceID
func NewInvoiceID(value uint) InvoiceID {
	return InvoiceID{value: value}
}

// GenerateInvoiceID generates a new unique invoice ID (placeholder)
func GenerateInvoiceID() InvoiceID {
	// In production, this would be set by the repository after database insertion
	// The ID will be assigned when the aggregate is persisted
	return InvoiceID{value: 0}
}

// ParseInvoiceID parses an invoice ID from string
func ParseInvoiceID(s string) (InvoiceID, error) {
	if s == "" {
		return InvoiceID{}, fmt.Errorf("invoice ID cannot be empty")
	}

	value, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		return InvoiceID{}, fmt.Errorf("invalid invoice ID format: %w", err)
	}

	return NewInvoiceID(uint(value)), nil
}

// Value returns the underlying value
func (id InvoiceID) Value() uint {
	return id.value
}

// String returns string representation of the invoice ID
func (id InvoiceID) String() string {
	return strconv.FormatUint(uint64(id.value), 10)
}

// IsZero checks if the ID is zero value
func (id InvoiceID) IsZero() bool {
	return id.value == 0
}

// Equals checks if two invoice IDs are equal
func (id InvoiceID) Equals(other InvoiceID) bool {
	return id.value == other.value
}