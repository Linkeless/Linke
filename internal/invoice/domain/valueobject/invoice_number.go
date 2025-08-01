package valueobject

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// InvoiceNumber represents a unique invoice number
type InvoiceNumber struct {
	value string
}

// NewInvoiceNumber creates a new InvoiceNumber
func NewInvoiceNumber(value string) (InvoiceNumber, error) {
	if err := validateInvoiceNumber(value); err != nil {
		return InvoiceNumber{}, err
	}

	return InvoiceNumber{value: strings.ToUpper(strings.TrimSpace(value))}, nil
}

// GenerateInvoiceNumber generates a new unique invoice number
func GenerateInvoiceNumber() InvoiceNumber {
	now := time.Now()
	dateStr := now.Format("20060102")
	
	// This is a simplified generation, in production you'd want to ensure uniqueness
	// by checking the database for the sequence number
	sequence := 1
	number := fmt.Sprintf("INV%s%04d", dateStr, sequence)
	
	// This should not fail with our generated format
	invoiceNumber, _ := NewInvoiceNumber(number)
	return invoiceNumber
}

// GenerateInvoiceNumberWithSequence generates an invoice number with specific sequence
func GenerateInvoiceNumberWithSequence(sequence uint) InvoiceNumber {
	now := time.Now()
	dateStr := now.Format("20060102")
	number := fmt.Sprintf("INV%s%06d", dateStr, sequence)
	
	invoiceNumber, _ := NewInvoiceNumber(number)
	return invoiceNumber
}

// ParseInvoiceNumber parses an invoice number from string
func ParseInvoiceNumber(s string) (InvoiceNumber, error) {
	return NewInvoiceNumber(s)
}

// Value returns the underlying value
func (in InvoiceNumber) Value() string {
	return in.value
}

// String returns string representation of the invoice number
func (in InvoiceNumber) String() string {
	return in.value
}

// IsEmpty checks if the invoice number is empty
func (in InvoiceNumber) IsEmpty() bool {
	return in.value == ""
}

// Equals checks if two invoice numbers are equal
func (in InvoiceNumber) Equals(other InvoiceNumber) bool {
	return in.value == other.value
}

// validateInvoiceNumber validates the invoice number format
func validateInvoiceNumber(value string) error {
	value = strings.TrimSpace(value)
	
	if value == "" {
		return fmt.Errorf("invoice number cannot be empty")
	}

	if len(value) > 32 {
		return fmt.Errorf("invoice number cannot exceed 32 characters")
	}

	// Allow alphanumeric characters, hyphens, and underscores
	pattern := `^[A-Za-z0-9\-_]+$`
	matched, err := regexp.MatchString(pattern, value)
	if err != nil {
		return fmt.Errorf("failed to validate invoice number: %w", err)
	}

	if !matched {
		return fmt.Errorf("invoice number can only contain letters, numbers, hyphens, and underscores")
	}

	return nil
}