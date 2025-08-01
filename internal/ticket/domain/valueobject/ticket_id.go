package valueobject

import (
	"fmt"
	"strconv"
)

// TicketID represents the unique identifier for a ticket
type TicketID struct {
	value uint
}

// NewTicketID creates a new TicketID
func NewTicketID(value uint) TicketID {
	return TicketID{value: value}
}

// Value returns the underlying value
func (t TicketID) Value() uint {
	return t.value
}

// String returns string representation
func (t TicketID) String() string {
	return strconv.FormatUint(uint64(t.value), 10)
}

// IsZero checks if the ID is zero value
func (t TicketID) IsZero() bool {
	return t.value == 0
}

// Equals checks equality with another TicketID
func (t TicketID) Equals(other TicketID) bool {
	return t.value == other.value
}

// MarshalJSON implements json.Marshaler
func (t TicketID) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`%d`, t.value)), nil
}

// UnmarshalJSON implements json.Unmarshaler
func (t *TicketID) UnmarshalJSON(data []byte) error {
	str := string(data)
	if str == "null" {
		t.value = 0
		return nil
	}
	
	value, err := strconv.ParseUint(str, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid ticket ID: %s", str)
	}
	
	t.value = uint(value)
	return nil
}