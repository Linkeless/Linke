package command

import (
	"time"

	"linke/internal/invoice/domain/valueobject"
	"linke/internal/invoice/handler/dto"
)

// CreateInvoiceCommand represents a command to create an invoice
type CreateInvoiceCommand struct {
	OrderID         uint                           `json:"order_id"`
	UserID          uint                           `json:"user_id"`
	Type            string                         `json:"type"`
	Subtotal        valueobject.Money              `json:"subtotal"`
	TaxInfo         valueobject.TaxInfo            `json:"tax_info"`
	BillingInfo     valueobject.BillingAddress     `json:"billing_info"`
	CompanyInfo     *valueobject.CompanyInfo       `json:"company_info"`
	Description     string                         `json:"description"`
	Notes           string                         `json:"notes"`
	DueDate         *time.Time                     `json:"due_date"`
	PaymentTerms    int                            `json:"payment_terms"`
	Template        string                         `json:"template"`
	Language        string                         `json:"language"`
}

// SendInvoiceCommand represents a command to send an invoice
type SendInvoiceCommand struct {
	InvoiceID string   `json:"invoice_id"`
	Email     string   `json:"email"`
	Subject   string   `json:"subject"`
	Message   string   `json:"message"`
	CCEmails  []string `json:"cc_emails"`
}

// UpdateInvoiceCommand represents a command to update an invoice
type UpdateInvoiceCommand struct {
	InvoiceID       string                       `json:"invoice_id"`
	BillingInfo     *valueobject.BillingAddress  `json:"billing_info"`
	CompanyInfo     *valueobject.CompanyInfo     `json:"company_info"`
	TaxInfo         *dto.TaxInfoDTO              `json:"tax_info"`
	Description     *string                      `json:"description"`
	Notes           *string                      `json:"notes"`
	DueDate         *time.Time                   `json:"due_date"`
	PaymentTerms    *int                         `json:"payment_terms"`
	InternalNotes   *string                      `json:"internal_notes"`
}

// PayInvoiceCommand represents a command to mark an invoice as paid
type PayInvoiceCommand struct {
	InvoiceID  string            `json:"invoice_id"`
	Amount     valueobject.Money `json:"amount"`
	PaymentRef string            `json:"payment_ref"`
	Notes      string            `json:"notes"`
}

// VoidInvoiceCommand represents a command to void an invoice
type VoidInvoiceCommand struct {
	InvoiceID string `json:"invoice_id"`
	Reason    string `json:"reason"`
}

// MarkOverdueCommand represents a command to mark an invoice as overdue
type MarkOverdueCommand struct {
	InvoiceID string `json:"invoice_id"`
}

// DeleteInvoiceCommand represents a command to delete an invoice
type DeleteInvoiceCommand struct {
	InvoiceID string `json:"invoice_id"`
}

// MarkInvoiceAsPaidCommand is an alias for PayInvoiceCommand for compatibility
type MarkInvoiceAsPaidCommand = PayInvoiceCommand