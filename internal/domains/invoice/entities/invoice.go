package entities

import (
	"time"

	"gorm.io/gorm"

	"linke/internal/domains/invoice/constants"
)

// Invoice represents an invoice for a subscription order
type Invoice struct {
	// Primary Key
	ID uint `json:"id" gorm:"primaryKey"`

	// Foreign Keys
	UserID              uint `json:"user_id" gorm:"not null;index"`
	SubscriptionOrderID uint `json:"subscription_order_id" gorm:"not null;index"`

	// Invoice Information
	InvoiceNumber string `json:"invoice_number" gorm:"uniqueIndex;size:50;not null"`
	InvoiceType   string `json:"invoice_type" gorm:"size:20;not null;default:'standard'"` // standard, proforma, credit_note
	Status        string `json:"status" gorm:"size:20;not null;default:'draft'"`          // draft, sent, paid, overdue, cancelled, voided

	// Financial Details
	Amount      float64 `json:"amount" gorm:"type:decimal(10,2);not null"`
	Currency    string  `json:"currency" gorm:"size:3;not null;default:'CNY'"`
	TaxAmount   float64 `json:"tax_amount" gorm:"type:decimal(10,2);default:0"`
	TotalAmount float64 `json:"total_amount" gorm:"type:decimal(10,2);not null"`

	// Tax Information
	TaxRate   float64 `json:"tax_rate" gorm:"type:decimal(5,4);default:0"` // Tax rate as percentage (e.g., 0.2 for 20%)
	TaxType   string  `json:"tax_type,omitempty" gorm:"size:20"`           // VAT, GST, etc.
	TaxNumber string  `json:"tax_number,omitempty" gorm:"size:50"`         // Business tax number

	// Billing Information
	BillingName    string `json:"billing_name" gorm:"size:200;not null"`
	BillingEmail   string `json:"billing_email" gorm:"size:191;not null"`
	BillingAddress string `json:"billing_address,omitempty" gorm:"type:text"`
	BillingCity    string `json:"billing_city,omitempty" gorm:"size:100"`
	BillingState   string `json:"billing_state,omitempty" gorm:"size:100"`
	BillingCountry string `json:"billing_country,omitempty" gorm:"size:2"` // ISO country code
	BillingZip     string `json:"billing_zip,omitempty" gorm:"size:20"`

	// Company Information (for business invoices)
	CompanyName    string `json:"company_name,omitempty" gorm:"size:200"`
	CompanyTaxID   string `json:"company_tax_id,omitempty" gorm:"size:50"`
	CompanyAddress string `json:"company_address,omitempty" gorm:"type:text"`

	// Important Dates
	IssuedAt time.Time  `json:"issued_at" gorm:"not null;index"`
	DueAt    *time.Time `json:"due_at,omitempty" gorm:"index"`
	PaidAt   *time.Time `json:"paid_at,omitempty" gorm:"index"`
	SentAt   *time.Time `json:"sent_at,omitempty" gorm:"index"`
	VoidedAt *time.Time `json:"voided_at,omitempty" gorm:"index"`

	// Payment Information
	PaymentMethod    string `json:"payment_method,omitempty" gorm:"size:50"`
	PaymentReference string `json:"payment_reference,omitempty" gorm:"size:100"`

	// Invoice Template and Language
	Template string `json:"template,omitempty" gorm:"size:50;default:'default'"`
	Language string `json:"language,omitempty" gorm:"size:5;default:'en'"`

	// File Storage
	PDFPath string `json:"pdf_path,omitempty" gorm:"size:500"` // Path to generated PDF
	PDFSize int64  `json:"pdf_size,omitempty"`                 // PDF file size in bytes

	// Additional Information
	Description string `json:"description,omitempty" gorm:"type:text"`
	Notes       string `json:"notes,omitempty" gorm:"type:text"`
	Metadata    string `json:"metadata,omitempty" gorm:"type:text"` // JSON metadata

	// Note: Relationships removed to avoid cross-domain dependencies
	// Related data should be fetched and assembled at the application layer

	// Timestamp Fields
	CreatedAt time.Time      `json:"created_at" gorm:"not null;index"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"not null"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

// TableName returns the table name for Invoice model
func (Invoice) TableName() string {
	return "invoices"
}


// IsDraft checks if the invoice is in draft status
func (i *Invoice) IsDraft() bool {
	return i.Status == constants.InvoiceStatusDraft
}

// IsSent checks if the invoice has been sent
func (i *Invoice) IsSent() bool {
	return i.Status == constants.InvoiceStatusSent
}

// IsPaid checks if the invoice is paid
func (i *Invoice) IsPaid() bool {
	return i.Status == constants.InvoiceStatusPaid
}

// IsOverdue checks if the invoice is overdue
func (i *Invoice) IsOverdue() bool {
	return i.Status == constants.InvoiceStatusOverdue || (i.DueAt != nil && time.Now().After(*i.DueAt) && !i.IsPaid())
}

// IsCancelled checks if the invoice is cancelled
func (i *Invoice) IsCancelled() bool {
	return i.Status == constants.InvoiceStatusCancelled
}

// IsVoided checks if the invoice is voided
func (i *Invoice) IsVoided() bool {
	return i.Status == constants.InvoiceStatusVoided
}

// CanBeEdited checks if the invoice can be edited
func (i *Invoice) CanBeEdited() bool {
	return i.IsDraft()
}

// CanBeSent checks if the invoice can be sent
func (i *Invoice) CanBeSent() bool {
	return i.IsDraft()
}

// CanBePaid checks if the invoice can be marked as paid
func (i *Invoice) CanBePaid() bool {
	return i.IsSent() && !i.IsPaid() && !i.IsVoided() && !i.IsCancelled()
}

// CanBeVoided checks if the invoice can be voided
func (i *Invoice) CanBeVoided() bool {
	return !i.IsPaid() && !i.IsVoided() && !i.IsCancelled()
}

// GetDisplayName returns the display name for billing
func (i *Invoice) GetDisplayName() string {
	if i.CompanyName != "" {
		return i.CompanyName
	}
	return i.BillingName
}

// GetFullAddress returns the complete billing address
func (i *Invoice) GetFullAddress() string {
	address := i.BillingAddress
	if i.BillingCity != "" {
		if address != "" {
			address += "\n"
		}
		address += i.BillingCity
	}
	if i.BillingState != "" {
		if address != "" && i.BillingCity != "" {
			address += ", "
		} else if address != "" {
			address += "\n"
		}
		address += i.BillingState
	}
	if i.BillingZip != "" {
		if address != "" {
			address += " "
		}
		address += i.BillingZip
	}
	if i.BillingCountry != "" {
		if address != "" {
			address += "\n"
		}
		address += i.BillingCountry
	}
	return address
}

