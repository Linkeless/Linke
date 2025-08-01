package valueobject

import (
	"fmt"
	"strconv"
)

// InvoiceID represents a unique identifier for an invoice
type InvoiceID struct {
	value uint
}

// NewInvoiceID creates a new InvoiceID with validation
func NewInvoiceID(value uint) (InvoiceID, error) {
	if value == 0 {
		return InvoiceID{}, fmt.Errorf("invoice ID cannot be zero")
	}
	return InvoiceID{value: value}, nil
}

// GenerateInvoiceID generates a new placeholder invoice ID
// In production, this would be set by the repository after database insertion
func GenerateInvoiceID() InvoiceID {
	// The ID will be assigned when the aggregate is persisted
	return InvoiceID{value: 0}
}

// ParseInvoiceID parses an invoice ID from string with validation
func ParseInvoiceID(s string) (InvoiceID, error) {
	if s == "" {
		return InvoiceID{}, fmt.Errorf("invoice ID string cannot be empty")
	}

	value, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		return InvoiceID{}, fmt.Errorf("invalid invoice ID format: %w", err)
	}

	return NewInvoiceID(uint(value))
}

// MustNewInvoiceID creates a new InvoiceID and panics on error
// Use this only when you are certain the value is valid
func MustNewInvoiceID(value uint) InvoiceID {
	id, err := NewInvoiceID(value)
	if err != nil {
		panic(fmt.Sprintf("failed to create invoice ID: %v", err))
	}
	return id
}

// Value returns the underlying value
func (id InvoiceID) Value() uint {
	return id.value
}

// String returns string representation of the invoice ID
func (id InvoiceID) String() string {
	return strconv.FormatUint(uint64(id.value), 10)
}

// IsZero checks if the ID is zero value (unassigned)
func (id InvoiceID) IsZero() bool {
	return id.value == 0
}

// IsValid checks if the ID is valid (non-zero)
func (id InvoiceID) IsValid() bool {
	return id.value != 0
}

// Equals checks if two invoice IDs are equal
func (id InvoiceID) Equals(other InvoiceID) bool {
	return id.value == other.value
}

// MarshalJSON implements json.Marshaler
func (id InvoiceID) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`%d`, id.value)), nil
}

// UnmarshalJSON implements json.Unmarshaler
func (id *InvoiceID) UnmarshalJSON(data []byte) error {
	str := string(data)
	if str == "null" {
		id.value = 0
		return nil
	}
	
	value, err := strconv.ParseUint(str, 10, 32)
	if err != nil {
		return fmt.Errorf("invalid invoice ID JSON: %w", err)
	}
	
	// Allow zero value in JSON unmarshaling (for placeholder IDs)
	id.value = uint(value)
	return nil
}