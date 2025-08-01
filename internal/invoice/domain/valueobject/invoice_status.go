package valueobject

import (
	"fmt"
	"strings"
)

// InvoiceStatus represents the status of an invoice
type InvoiceStatus struct {
	value string
}

// Invoice status constants
var (
	StatusDraft   = InvoiceStatus{value: "draft"}
	StatusSent    = InvoiceStatus{value: "sent"}
	StatusPaid    = InvoiceStatus{value: "paid"}
	StatusOverdue = InvoiceStatus{value: "overdue"}
	StatusVoided  = InvoiceStatus{value: "voided"}
)

// NewInvoiceStatus creates a new InvoiceStatus
func NewInvoiceStatus(value string) (InvoiceStatus, error) {
	value = strings.ToLower(strings.TrimSpace(value))
	
	validStatuses := map[string]bool{
		"draft":   true,
		"sent":    true,
		"paid":    true,
		"overdue": true,
		"voided":  true,
	}

	if !validStatuses[value] {
		return InvoiceStatus{}, fmt.Errorf("invalid invoice status: %s", value)
	}

	return InvoiceStatus{value: value}, nil
}

// Value returns the underlying value
func (is InvoiceStatus) Value() string {
	return is.value
}

// String returns string representation of the status
func (is InvoiceStatus) String() string {
	return is.value
}

// IsDraft checks if the invoice is in draft status
func (is InvoiceStatus) IsDraft() bool {
	return is.value == "draft"
}

// IsSent checks if the invoice has been sent
func (is InvoiceStatus) IsSent() bool {
	return is.value == "sent"
}

// IsPaid checks if the invoice is paid
func (is InvoiceStatus) IsPaid() bool {
	return is.value == "paid"
}

// IsOverdue checks if the invoice is overdue
func (is InvoiceStatus) IsOverdue() bool {
	return is.value == "overdue"
}

// IsVoided checks if the invoice is voided
func (is InvoiceStatus) IsVoided() bool {
	return is.value == "voided"
}

// CanTransitionTo checks if the status can transition to another status
func (is InvoiceStatus) CanTransitionTo(newStatus InvoiceStatus) bool {
	transitions := map[string][]string{
		"draft":   {"sent", "voided"},
		"sent":    {"paid", "overdue", "voided"},
		"overdue": {"paid", "voided"},
		"paid":    {}, // No transitions from paid
		"voided":  {}, // No transitions from voided
	}

	allowedTransitions, exists := transitions[is.value]
	if !exists {
		return false
	}

	for _, allowed := range allowedTransitions {
		if allowed == newStatus.value {
			return true
		}
	}

	return false
}

// Equals checks if two invoice statuses are equal
func (is InvoiceStatus) Equals(other InvoiceStatus) bool {
	return is.value == other.value
}

// DisplayName returns human-readable status name
func (is InvoiceStatus) DisplayName() string {
	switch is.value {
	case "draft":
		return "Draft"
	case "sent":
		return "Sent"
	case "paid":
		return "Paid"
	case "overdue":
		return "Overdue"
	case "voided":
		return "Voided"
	default:
		return strings.Title(is.value)
	}
}

// IsTerminal checks if the status is terminal (no further transitions possible)
func (is InvoiceStatus) IsTerminal() bool {
	return is.IsPaid() || is.IsVoided()
}

// CanBeModified checks if invoice can be modified in this status
func (is InvoiceStatus) CanBeModified() bool {
	return is.IsDraft()
}

// RequiresPayment checks if invoice requires payment in this status
func (is InvoiceStatus) RequiresPayment() bool {
	return is.IsSent() || is.IsOverdue()
}