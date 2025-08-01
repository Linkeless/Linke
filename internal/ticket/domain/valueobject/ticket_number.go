package valueobject

import (
	"fmt"
	"regexp"
	"strings"
)

// TicketNumber represents a unique ticket number
type TicketNumber struct {
	value string
}

var ticketNumberPattern = regexp.MustCompile(`^TK-\d{8}-[A-Z0-9]{6}$`)

// NewTicketNumber creates a new TicketNumber with validation
func NewTicketNumber(value string) (TicketNumber, error) {
	value = strings.TrimSpace(value)
	
	if value == "" {
		return TicketNumber{}, fmt.Errorf("ticket number cannot be empty")
	}
	
	if !ticketNumberPattern.MatchString(value) {
		return TicketNumber{}, fmt.Errorf("invalid ticket number format: %s", value)
	}
	
	return TicketNumber{value: value}, nil
}

// MustTicketNumber creates a TicketNumber and panics if invalid
func MustTicketNumber(value string) TicketNumber {
	tn, err := NewTicketNumber(value)
	if err != nil {
		panic(err)
	}
	return tn
}

// Value returns the underlying value
func (t TicketNumber) Value() string {
	return t.value
}

// String returns string representation
func (t TicketNumber) String() string {
	return t.value
}

// IsEmpty checks if the ticket number is empty
func (t TicketNumber) IsEmpty() bool {
	return t.value == ""
}

// Equals checks equality with another TicketNumber
func (t TicketNumber) Equals(other TicketNumber) bool {
	return t.value == other.value
}

// MarshalJSON implements json.Marshaler
func (t TicketNumber) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, t.value)), nil
}

// UnmarshalJSON implements json.Unmarshaler
func (t *TicketNumber) UnmarshalJSON(data []byte) error {
	str := strings.Trim(string(data), `"`)
	if str == "null" {
		t.value = ""
		return nil
	}
	
	tn, err := NewTicketNumber(str)
	if err != nil {
		return err
	}
	
	*t = tn
	return nil
}