package valueobject

import (
	"fmt"
	"strings"
)

// InvoiceType represents the type of invoice
type InvoiceType string

const (
	TypeStandard   InvoiceType = "standard"
	TypeProforma   InvoiceType = "proforma" 
	TypeCreditNote InvoiceType = "credit_note"
	TypeReceipt    InvoiceType = "receipt"
)

// NewInvoiceType creates a new InvoiceType
func NewInvoiceType(t string) (InvoiceType, error) {
	// Normalize input: trim spaces and convert to lowercase
	normalized := strings.ToLower(strings.TrimSpace(t))
	invoiceType := InvoiceType(normalized)
	if !invoiceType.IsValid() {
		return "", fmt.Errorf("invalid invoice type: %s", t)
	}
	return invoiceType, nil
}

// IsValid checks if the invoice type is valid
func (t InvoiceType) IsValid() bool {
	switch t {
	case TypeStandard, TypeProforma, TypeCreditNote, TypeReceipt:
		return true
	default:
		return false
	}
}

// String returns string representation
func (t InvoiceType) String() string {
	return string(t)
}

// IsStandard checks if this is a standard invoice
func (t InvoiceType) IsStandard() bool {
	return t == TypeStandard
}

// IsProforma checks if this is a proforma invoice
func (t InvoiceType) IsProforma() bool {
	return t == TypeProforma
}

// IsCreditNote checks if this is a credit note
func (t InvoiceType) IsCreditNote() bool {
	return t == TypeCreditNote
}

// IsReceipt checks if this is a receipt
func (t InvoiceType) IsReceipt() bool {
	return t == TypeReceipt
}

// RequiresPayment checks if this invoice type requires payment
func (t InvoiceType) RequiresPayment() bool {
	switch t {
	case TypeStandard, TypeProforma:
		return true
	case TypeCreditNote, TypeReceipt:
		return false
	default:
		return false
	}
}

// DisplayName returns a human-readable name for the invoice type
func (t InvoiceType) DisplayName() string {
	switch t {
	case TypeStandard:
		return "Standard Invoice"
	case TypeProforma:
		return "Proforma Invoice"
	case TypeCreditNote:
		return "Credit Note"
	case TypeReceipt:
		return "Receipt"
	default:
		return string(t)
	}
}

// Value returns the raw string value of the invoice type
func (t InvoiceType) Value() string {
	return string(t)
}

// Equals checks if two invoice types are equal
func (t InvoiceType) Equals(other InvoiceType) bool {
	return t == other
}