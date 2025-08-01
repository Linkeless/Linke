package valueobject

import (
	"fmt"
	"strings"
)

// TicketPriority represents the priority level of a ticket
type TicketPriority struct {
	value string
}

// Valid ticket priorities
const (
	TicketPriorityLow      = "low"
	TicketPriorityNormal   = "normal"
	TicketPriorityHigh     = "high"
	TicketPriorityUrgent   = "urgent"
	TicketPriorityCritical = "critical"
)

var validPriorities = map[string]int{
	TicketPriorityLow:      1,
	TicketPriorityNormal:   2,
	TicketPriorityHigh:     3,
	TicketPriorityUrgent:   4,
	TicketPriorityCritical: 5,
}

// NewTicketPriority creates a new TicketPriority with validation
func NewTicketPriority(value string) (TicketPriority, error) {
	value = strings.ToLower(strings.TrimSpace(value))
	
	if value == "" {
		return TicketPriority{}, fmt.Errorf("ticket priority cannot be empty")
	}
	
	if _, exists := validPriorities[value]; !exists {
		return TicketPriority{}, fmt.Errorf("invalid ticket priority: %s", value)
	}
	
	return TicketPriority{value: value}, nil
}

// MustTicketPriority creates a TicketPriority and panics if invalid
func MustTicketPriority(value string) TicketPriority {
	tp, err := NewTicketPriority(value)
	if err != nil {
		panic(err)
	}
	return tp
}

// DefaultTicketPriority returns the default priority for new tickets
func DefaultTicketPriority() TicketPriority {
	return TicketPriority{value: TicketPriorityNormal}
}

// Value returns the underlying value
func (t TicketPriority) Value() string {
	return t.value
}

// String returns string representation
func (t TicketPriority) String() string {
	return t.value
}

// Level returns the numeric priority level (1-5, higher is more urgent)
func (t TicketPriority) Level() int {
	return validPriorities[t.value]
}

// IsLow checks if priority is low
func (t TicketPriority) IsLow() bool {
	return t.value == TicketPriorityLow
}

// IsNormal checks if priority is normal
func (t TicketPriority) IsNormal() bool {
	return t.value == TicketPriorityNormal
}

// IsHigh checks if priority is high
func (t TicketPriority) IsHigh() bool {
	return t.value == TicketPriorityHigh
}

// IsUrgent checks if priority is urgent
func (t TicketPriority) IsUrgent() bool {
	return t.value == TicketPriorityUrgent
}

// IsCritical checks if priority is critical
func (t TicketPriority) IsCritical() bool {
	return t.value == TicketPriorityCritical
}

// IsHigherThan checks if this priority is higher than another
func (t TicketPriority) IsHigherThan(other TicketPriority) bool {
	return t.Level() > other.Level()
}

// IsLowerThan checks if this priority is lower than another
func (t TicketPriority) IsLowerThan(other TicketPriority) bool {
	return t.Level() < other.Level()
}

// RequiresImmediateAttention checks if priority requires immediate attention
func (t TicketPriority) RequiresImmediateAttention() bool {
	return t.IsUrgent() || t.IsCritical()
}

// Equals checks equality with another TicketPriority
func (t TicketPriority) Equals(other TicketPriority) bool {
	return t.value == other.value
}

// MarshalJSON implements json.Marshaler
func (t TicketPriority) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, t.value)), nil
}

// UnmarshalJSON implements json.Unmarshaler
func (t *TicketPriority) UnmarshalJSON(data []byte) error {
	str := strings.Trim(string(data), `"`)
	if str == "null" {
		*t = DefaultTicketPriority()
		return nil
	}
	
	tp, err := NewTicketPriority(str)
	if err != nil {
		return err
	}
	
	*t = tp
	return nil
}