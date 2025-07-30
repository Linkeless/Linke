package model

import (
	"time"

	"gorm.io/gorm"
)

// Invoice represents an invoice for a subscription order
type Invoice struct {
	// Primary Key
	ID uint `json:"id" gorm:"primaryKey"`

	// Foreign Keys
	UserID               uint  `json:"user_id" gorm:"not null;index"`
	SubscriptionOrderID  uint  `json:"subscription_order_id" gorm:"not null;index"`
	
	// Invoice Information
	InvoiceNumber string `json:"invoice_number" gorm:"uniqueIndex;size:50;not null"`
	InvoiceType   string `json:"invoice_type" gorm:"size:20;not null;default:'standard'"` // standard, proforma, credit_note
	Status        string `json:"status" gorm:"size:20;not null;default:'draft'"`          // draft, sent, paid, overdue, cancelled, voided
	
	// Financial Details
	Amount        float64 `json:"amount" gorm:"type:decimal(10,2);not null"`
	Currency      string  `json:"currency" gorm:"size:3;not null;default:'USD'"`
	TaxAmount     float64 `json:"tax_amount" gorm:"type:decimal(10,2);default:0"`
	TotalAmount   float64 `json:"total_amount" gorm:"type:decimal(10,2);not null"`
	
	// Tax Information
	TaxRate       float64 `json:"tax_rate" gorm:"type:decimal(5,4);default:0"`           // Tax rate as percentage (e.g., 0.2 for 20%)
	TaxType       string  `json:"tax_type,omitempty" gorm:"size:20"`                     // VAT, GST, etc.
	TaxNumber     string  `json:"tax_number,omitempty" gorm:"size:50"`                   // Business tax number
	
	// Billing Information
	BillingName    string `json:"billing_name" gorm:"size:200;not null"`
	BillingEmail   string `json:"billing_email" gorm:"size:191;not null"`
	BillingAddress string `json:"billing_address,omitempty" gorm:"type:text"`
	BillingCity    string `json:"billing_city,omitempty" gorm:"size:100"`
	BillingState   string `json:"billing_state,omitempty" gorm:"size:100"`
	BillingCountry string `json:"billing_country,omitempty" gorm:"size:2"`               // ISO country code
	BillingZip     string `json:"billing_zip,omitempty" gorm:"size:20"`
	
	// Company Information (for business invoices)
	CompanyName    string `json:"company_name,omitempty" gorm:"size:200"`
	CompanyTaxID   string `json:"company_tax_id,omitempty" gorm:"size:50"`
	CompanyAddress string `json:"company_address,omitempty" gorm:"type:text"`
	
	// Important Dates
	IssuedAt  time.Time  `json:"issued_at" gorm:"not null;index"`
	DueAt     *time.Time `json:"due_at,omitempty" gorm:"index"`
	PaidAt    *time.Time `json:"paid_at,omitempty" gorm:"index"`
	SentAt    *time.Time `json:"sent_at,omitempty" gorm:"index"`
	VoidedAt  *time.Time `json:"voided_at,omitempty" gorm:"index"`
	
	// Payment Information
	PaymentMethod    string `json:"payment_method,omitempty" gorm:"size:50"`
	PaymentReference string `json:"payment_reference,omitempty" gorm:"size:100"`
	
	// Invoice Template and Language
	Template string `json:"template,omitempty" gorm:"size:50;default:'default'"`
	Language string `json:"language,omitempty" gorm:"size:5;default:'en'"`
	
	// File Storage
	PDFPath    string `json:"pdf_path,omitempty" gorm:"size:500"`                      // Path to generated PDF
	PDFSize    int64  `json:"pdf_size,omitempty"`                                      // PDF file size in bytes
	
	// Additional Information
	Description string `json:"description,omitempty" gorm:"type:text"`
	Notes       string `json:"notes,omitempty" gorm:"type:text"`
	Metadata    string `json:"metadata,omitempty" gorm:"type:text"`                    // JSON metadata
	
	// Relationships (no foreign key constraints for performance)
	User             *User             `json:"user,omitempty" gorm:"-"`
	SubscriptionOrder *SubscriptionOrder `json:"subscription_order,omitempty" gorm:"-"`

	// Timestamp Fields
	CreatedAt time.Time      `json:"created_at" gorm:"not null;index"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"not null"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

// TableName returns the table name for Invoice model
func (Invoice) TableName() string {
	return "invoices"
}

// Invoice type constants
const (
	InvoiceTypeStandard  = "standard"
	InvoiceTypeProforma  = "proforma"
	InvoiceTypeCreditNote = "credit_note"
)

// Invoice status constants
const (
	InvoiceStatusDraft     = "draft"
	InvoiceStatusCancelled = "cancelled"
)

// IsDraft checks if the invoice is in draft status
func (i *Invoice) IsDraft() bool {
	return i.Status == InvoiceStatusDraft
}

// IsSent checks if the invoice has been sent
func (i *Invoice) IsSent() bool {
	return i.Status == InvoiceStatusSent
}

// IsPaid checks if the invoice is paid
func (i *Invoice) IsPaid() bool {
	return i.Status == InvoiceStatusPaid
}

// IsOverdue checks if the invoice is overdue
func (i *Invoice) IsOverdue() bool {
	return i.Status == InvoiceStatusOverdue || (i.DueAt != nil && time.Now().After(*i.DueAt) && !i.IsPaid())
}

// IsCancelled checks if the invoice is cancelled
func (i *Invoice) IsCancelled() bool {
	return i.Status == InvoiceStatusCancelled
}

// IsVoided checks if the invoice is voided
func (i *Invoice) IsVoided() bool {
	return i.Status == InvoiceStatusVoided
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

// InvoiceResponse represents the invoice data structure for API responses
type InvoiceResponse struct {
	ID                   uint       `json:"id" example:"1"`
	UserID               uint       `json:"user_id" example:"1"`
	SubscriptionOrderID  uint       `json:"subscription_order_id" example:"1"`
	InvoiceNumber        string     `json:"invoice_number" example:"INV-2024-001"`
	InvoiceType          string     `json:"invoice_type" example:"standard"`
	Status               string     `json:"status" example:"sent"`
	Amount               float64    `json:"amount" example:"29.99"`
	Currency             string     `json:"currency" example:"USD"`
	TaxAmount            float64    `json:"tax_amount" example:"5.99"`
	TotalAmount          float64    `json:"total_amount" example:"35.98"`
	TaxRate              float64    `json:"tax_rate" example:"0.2"`
	TaxType              string     `json:"tax_type,omitempty" example:"VAT"`
	TaxNumber            string     `json:"tax_number,omitempty" example:"GB123456789"`
	BillingName          string     `json:"billing_name" example:"John Doe"`
	BillingEmail         string     `json:"billing_email" example:"john@example.com"`
	BillingAddress       string     `json:"billing_address,omitempty" example:"123 Main St"`
	BillingCity          string     `json:"billing_city,omitempty" example:"New York"`
	BillingState         string     `json:"billing_state,omitempty" example:"NY"`
	BillingCountry       string     `json:"billing_country,omitempty" example:"US"`
	BillingZip           string     `json:"billing_zip,omitempty" example:"10001"`
	CompanyName          string     `json:"company_name,omitempty" example:"Acme Corp"`
	CompanyTaxID         string     `json:"company_tax_id,omitempty" example:"12-3456789"`
	CompanyAddress       string     `json:"company_address,omitempty" example:"456 Business Ave"`
	IssuedAt             time.Time  `json:"issued_at" example:"2024-01-01T00:00:00Z"`
	DueAt                *time.Time `json:"due_at,omitempty" example:"2024-01-31T23:59:59Z"`
	PaidAt               *time.Time `json:"paid_at,omitempty" example:"2024-01-15T10:30:00Z"`
	SentAt               *time.Time `json:"sent_at,omitempty" example:"2024-01-01T12:00:00Z"`
	VoidedAt             *time.Time `json:"voided_at,omitempty"`
	PaymentMethod        string     `json:"payment_method,omitempty" example:"credit_card"`
	PaymentReference     string     `json:"payment_reference,omitempty" example:"txn_123456"`
	Template             string     `json:"template,omitempty" example:"default"`
	Language             string     `json:"language,omitempty" example:"en"`
	PDFPath              string     `json:"pdf_path,omitempty" example:"/invoices/INV-2024-001.pdf"`
	PDFSize              int64      `json:"pdf_size,omitempty" example:"12345"`
	Description          string     `json:"description,omitempty" example:"Monthly subscription"`
	Notes                string     `json:"notes,omitempty" example:"Thank you for your business"`
	CreatedAt            time.Time  `json:"created_at" example:"2024-01-01T00:00:00Z"`
	UpdatedAt            time.Time  `json:"updated_at" example:"2024-01-01T00:00:00Z"`
	
	// Related data
	User             *UserResponse             `json:"user,omitempty"`
	SubscriptionOrder *SubscriptionOrderResponse `json:"subscription_order,omitempty"`
	
	// Computed fields
	IsOverdue    bool   `json:"is_overdue"`
	DaysOverdue  int    `json:"days_overdue,omitempty"`
	DisplayName  string `json:"display_name"`
	FullAddress  string `json:"full_address"`
}

// ToResponse converts Invoice to InvoiceResponse
func (i *Invoice) ToResponse() *InvoiceResponse {
	resp := &InvoiceResponse{
		ID:                  i.ID,
		UserID:              i.UserID,
		SubscriptionOrderID: i.SubscriptionOrderID,
		InvoiceNumber:       i.InvoiceNumber,
		InvoiceType:         i.InvoiceType,
		Status:              i.Status,
		Amount:              i.Amount,
		Currency:            i.Currency,
		TaxAmount:           i.TaxAmount,
		TotalAmount:         i.TotalAmount,
		TaxRate:             i.TaxRate,
		TaxType:             i.TaxType,
		TaxNumber:           i.TaxNumber,
		BillingName:         i.BillingName,
		BillingEmail:        i.BillingEmail,
		BillingAddress:      i.BillingAddress,
		BillingCity:         i.BillingCity,
		BillingState:        i.BillingState,
		BillingCountry:      i.BillingCountry,
		BillingZip:          i.BillingZip,
		CompanyName:         i.CompanyName,
		CompanyTaxID:        i.CompanyTaxID,
		CompanyAddress:      i.CompanyAddress,
		IssuedAt:            i.IssuedAt,
		DueAt:               i.DueAt,
		PaidAt:              i.PaidAt,
		SentAt:              i.SentAt,
		VoidedAt:            i.VoidedAt,
		PaymentMethod:       i.PaymentMethod,
		PaymentReference:    i.PaymentReference,
		Template:            i.Template,
		Language:            i.Language,
		PDFPath:             i.PDFPath,
		PDFSize:             i.PDFSize,
		Description:         i.Description,
		Notes:               i.Notes,
		CreatedAt:           i.CreatedAt,
		UpdatedAt:           i.UpdatedAt,
		IsOverdue:           i.IsOverdue(),
		DisplayName:         i.GetDisplayName(),
		FullAddress:         i.GetFullAddress(),
	}
	
	// Calculate days overdue
	if i.IsOverdue() && i.DueAt != nil {
		days := int(time.Since(*i.DueAt).Hours() / 24)
		if days > 0 {
			resp.DaysOverdue = days
		}
	}
	
	// Include related data if loaded
	if i.User != nil {
		resp.User = i.User.ToResponse()
	}
	if i.SubscriptionOrder != nil {
		resp.SubscriptionOrder = i.SubscriptionOrder.ToResponse()
	}
	
	return resp
}