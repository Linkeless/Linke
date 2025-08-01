package persistence

import (
	"time"

	"gorm.io/gorm"
)

// InvoicePO represents the invoice persistence object for GORM
type InvoicePO struct {
	// Primary Key
	ID uint `gorm:"primaryKey" json:"id"`

	// Foreign Keys
	OrderID uint `gorm:"not null;index" json:"order_id"`
	UserID  uint `gorm:"not null;index" json:"user_id"`

	// Invoice Information
	InvoiceNumber string `gorm:"uniqueIndex;size:32;not null" json:"invoice_number"`
	InvoiceType   string `gorm:"size:20;not null;default:'standard'" json:"invoice_type"`
	Status        string `gorm:"size:20;not null;default:'draft'" json:"status"`

	// Financial Information
	Subtotal    float64 `gorm:"type:decimal(10,2);not null" json:"subtotal"`
	TaxRate     float64 `gorm:"type:decimal(5,4);default:0" json:"tax_rate"`
	TaxAmount   float64 `gorm:"type:decimal(10,2);default:0" json:"tax_amount"`
	TotalAmount float64 `gorm:"type:decimal(10,2);not null" json:"total_amount"`
	Currency    string  `gorm:"size:3;not null;default:'USD'" json:"currency"`

	// Payment Terms Information
	IssuedAt         *time.Time `gorm:"index" json:"issued_at"`
	DueAt            *time.Time `gorm:"index" json:"due_at"`
	PaymentTermsDays int        `gorm:"default:30" json:"payment_terms_days"`

	// Payment Information
	PaidAmount float64    `gorm:"type:decimal(10,2);default:0" json:"paid_amount"`
	PaidAt     *time.Time `gorm:"index" json:"paid_at"`

	// Billing Information
	BillingName    string `gorm:"size:255;not null" json:"billing_name"`
	BillingEmail   string `gorm:"size:255;not null" json:"billing_email"`
	BillingAddress string `gorm:"type:text" json:"billing_address"`
	BillingCity    string `gorm:"size:100" json:"billing_city"`
	BillingState   string `gorm:"size:100" json:"billing_state"`
	BillingCountry string `gorm:"size:2" json:"billing_country"`
	BillingZip     string `gorm:"size:20" json:"billing_zip"`

	// Tax Information
	TaxNumber string `gorm:"size:50" json:"tax_number"`
	TaxType   string `gorm:"size:20" json:"tax_type"`

	// Company Information
	CompanyName    string `gorm:"size:255" json:"company_name"`
	CompanyAddress string `gorm:"type:text" json:"company_address"`
	CompanyTaxID   string `gorm:"size:50" json:"company_tax_id"`

	// Invoice Content
	Description string `gorm:"type:text;not null" json:"description"`
	LineItems   string `gorm:"type:json" json:"line_items"`

	// Document Management
	PDFPath  string `gorm:"size:500" json:"pdf_path"`
	PDFSize  int    `json:"pdf_size"`
	Template string `gorm:"size:50;default:'default'" json:"template"`
	Language string `gorm:"size:5;default:'en'" json:"language"`

	// Sending Records
	SentAt         *time.Time `gorm:"index" json:"sent_at"`
	SendCount      int        `gorm:"default:0" json:"send_count"`
	LastReminderAt *time.Time `gorm:"index" json:"last_reminder_at"`

	// Voiding Information
	VoidedAt   *time.Time `gorm:"index" json:"voided_at"`
	VoidReason string     `gorm:"type:text" json:"void_reason"`

	// Business Fields
	Notes         string `gorm:"type:text" json:"notes"`
	InternalNotes string `gorm:"type:text" json:"internal_notes"`
	Metadata      string `gorm:"type:json" json:"metadata"`

	// Timestamp Fields
	CreatedAt time.Time      `gorm:"not null;index" json:"created_at"`
	UpdatedAt time.Time      `gorm:"not null" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

// TableName returns the table name for the InvoicePO
func (InvoicePO) TableName() string {
	return "invoices"
}