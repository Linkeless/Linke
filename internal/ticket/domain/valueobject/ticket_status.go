package valueobject

import (
	"fmt"
	"strings"
)

// TicketStatus represents the status of a ticket
type TicketStatus struct {
	value string
}

// Valid ticket statuses
const (
	TicketStatusOpen       = "open"
	TicketStatusInProgress = "in_progress"
	TicketStatusPending    = "pending"
	TicketStatusResolved   = "resolved"
	TicketStatusClosed     = "closed"
)

var validStatuses = map[string]bool{
	TicketStatusOpen:       true,
	TicketStatusInProgress: true,
	TicketStatusPending:    true,
	TicketStatusResolved:   true,
	TicketStatusClosed:     true,
}

// NewTicketStatus creates a new TicketStatus with validation
func NewTicketStatus(value string) (TicketStatus, error) {
	value = strings.ToLower(strings.TrimSpace(value))
	
	if value == "" {
		return TicketStatus{}, fmt.Errorf("ticket status cannot be empty")
	}
	
	if !validStatuses[value] {
		return TicketStatus{}, fmt.Errorf("invalid ticket status: %s", value)
	}
	
	return TicketStatus{value: value}, nil
}

// MustTicketStatus creates a TicketStatus and panics if invalid
func MustTicketStatus(value string) TicketStatus {
	ts, err := NewTicketStatus(value)
	if err != nil {
		panic(err)
	}
	return ts
}

// DefaultTicketStatus returns the default status for new tickets
func DefaultTicketStatus() TicketStatus {
	return TicketStatus{value: TicketStatusOpen}
}

// Value returns the underlying value
func (t TicketStatus) Value() string {
	return t.value
}

// String returns string representation
func (t TicketStatus) String() string {
	return t.value
}

// IsOpen checks if status is open
func (t TicketStatus) IsOpen() bool {
	return t.value == TicketStatusOpen
}

// IsInProgress checks if status is in progress
func (t TicketStatus) IsInProgress() bool {
	return t.value == TicketStatusInProgress
}

// IsPending checks if status is pending
func (t TicketStatus) IsPending() bool {
	return t.value == TicketStatusPending
}

// IsResolved checks if status is resolved
func (t TicketStatus) IsResolved() bool {
	return t.value == TicketStatusResolved
}

// IsClosed checks if status is closed
func (t TicketStatus) IsClosed() bool {
	return t.value == TicketStatusClosed
}

// IsActive checks if ticket is in an active state (not resolved or closed)
func (t TicketStatus) IsActive() bool {
	return !t.IsResolved() && !t.IsClosed()
}

// CanTransitionTo checks if status can transition to another status
func (t TicketStatus) CanTransitionTo(target TicketStatus) bool {
	from := t.value
	to := target.value
	
	// Define valid transitions
	validTransitions := map[string][]string{
		TicketStatusOpen:       {TicketStatusInProgress, TicketStatusPending, TicketStatusResolved, TicketStatusClosed},
		TicketStatusInProgress: {TicketStatusPending, TicketStatusResolved, TicketStatusClosed, TicketStatusOpen},
		TicketStatusPending:    {TicketStatusInProgress, TicketStatusResolved, TicketStatusClosed, TicketStatusOpen},
		TicketStatusResolved:   {TicketStatusClosed, TicketStatusOpen, TicketStatusInProgress}, // Can reopen resolved tickets
		TicketStatusClosed:     {TicketStatusOpen}, // Can reopen closed tickets
	}
	
	allowedTargets, exists := validTransitions[from]
	if !exists {
		return false
	}
	
	for _, allowed := range allowedTargets {
		if allowed == to {
			return true
		}
	}
	
	return false
}

// Equals checks equality with another TicketStatus
func (t TicketStatus) Equals(other TicketStatus) bool {
	return t.value == other.value
}

// MarshalJSON implements json.Marshaler
func (t TicketStatus) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, t.value)), nil
}

// UnmarshalJSON implements json.Unmarshaler
func (t *TicketStatus) UnmarshalJSON(data []byte) error {
	str := strings.Trim(string(data), `"`)
	if str == "null" {
		*t = DefaultTicketStatus()
		return nil
	}
	
	ts, err := NewTicketStatus(str)
	if err != nil {
		return err
	}
	
	*t = ts
	return nil
}