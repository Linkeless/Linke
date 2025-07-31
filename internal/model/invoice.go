package model

import (
	"time"

	"gorm.io/gorm"
)

// Invoice represents an invoice (payment request and financial compliance)
type Invoice struct {
	// Primary Key
	ID uint `json:"id" gorm:"primaryKey"`

	// Foreign Keys  
	OrderID uint `json:"order_id" gorm:"not null;index"`
	UserID  uint `json:"user_id" gorm:"not null;index"`

	// Invoice Information
	InvoiceNumber string `json:"invoice_number" gorm:"uniqueIndex;size:32;not null"`
	InvoiceType   string `json:"invoice_type" gorm:"size:20;not null;default:'standard'"` // standard, proforma, credit_note
	Status        string `json:"status" gorm:"size:20;not null;default:'draft'"`          // draft, sent, paid, overdue, voided

	// Financial Information
	Subtotal    float64 `json:"subtotal" gorm:"type:decimal(10,2);not null"`
	TaxRate     float64 `json:"tax_rate" gorm:"type:decimal(5,4);default:0"`           // Tax rate as decimal (e.g., 0.2 for 20%)
	TaxAmount   float64 `json:"tax_amount" gorm:"type:decimal(10,2);default:0"`
	TotalAmount float64 `json:"total_amount" gorm:"type:decimal(10,2);not null"`
	Currency    string  `json:"currency" gorm:"size:3;not null;default:'USD'"`

	// Payment Terms Information
	IssuedAt         *time.Time `json:"issued_at,omitempty" gorm:"index"`
	DueAt            *time.Time `json:"due_at,omitempty" gorm:"index"`
	PaymentTermsDays int        `json:"payment_terms_days" gorm:"default:30"`

	// Payment Information
	PaidAmount float64    `json:"paid_amount" gorm:"type:decimal(10,2);default:0"`
	PaidAt     *time.Time `json:"paid_at,omitempty" gorm:"index"`

	// Billing Information
	BillingName    string `json:"billing_name" gorm:"size:255;not null"`
	BillingEmail   string `json:"billing_email" gorm:"size:255;not null"`
	BillingAddress string `json:"billing_address,omitempty" gorm:"type:text"`
	BillingCity    string `json:"billing_city,omitempty" gorm:"size:100"`
	BillingState   string `json:"billing_state,omitempty" gorm:"size:100"`
	BillingCountry string `json:"billing_country,omitempty" gorm:"size:2"`  // ISO country code
	BillingZip     string `json:"billing_zip,omitempty" gorm:"size:20"`

	// Tax Information
	TaxNumber string `json:"tax_number,omitempty" gorm:"size:50"`
	TaxType   string `json:"tax_type,omitempty" gorm:"size:20"` // VAT, GST, etc.

	// Company Information (for business invoices)
	CompanyName    string `json:"company_name,omitempty" gorm:"size:255"`
	CompanyAddress string `json:"company_address,omitempty" gorm:"type:text"`
	CompanyTaxID   string `json:"company_tax_id,omitempty" gorm:"size:50"`

	// Invoice Content
	Description string `json:"description" gorm:"type:text;not null"`
	LineItems   string `json:"line_items,omitempty" gorm:"type:json"` // JSON array of line items

	// Document Management
	PDFPath  string `json:"pdf_path,omitempty" gorm:"size:500"`
	PDFSize  int    `json:"pdf_size,omitempty"`
	Template string `json:"template" gorm:"size:50;default:'default'"`
	Language string `json:"language" gorm:"size:5;default:'en'"`

	// Sending Records
	SentAt         *time.Time `json:"sent_at,omitempty" gorm:"index"`
	SendCount      int        `json:"send_count" gorm:"default:0"`
	LastReminderAt *time.Time `json:"last_reminder_at,omitempty" gorm:"index"`

	// Voiding Information
	VoidedAt   *time.Time `json:"voided_at,omitempty" gorm:"index"`
	VoidReason string     `json:"void_reason,omitempty" gorm:"type:text"`

	// Business Fields
	Notes         string `json:"notes,omitempty" gorm:"type:text"`
	InternalNotes string `json:"internal_notes,omitempty" gorm:"type:text"`
	Metadata      string `json:"metadata,omitempty" gorm:"type:json"`

	// Relationships (no foreign key constraints for performance)
	Order *Order `json:"order,omitempty" gorm:"-"`
	User  *User  `json:"user,omitempty" gorm:"-"`

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
	InvoiceTypeStandard   = "standard"
	InvoiceTypeProforma   = "proforma"
	InvoiceTypeCreditNote = "credit_note"
)

// Invoice status constants
const (
	InvoiceStatusDraft   = "draft"
	InvoiceStatusSent    = "sent"
	InvoiceStatusPaid    = "paid"
	InvoiceStatusOverdue = "overdue"
	InvoiceStatusVoided  = "voided"
)

// Business logic methods

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
	if i.Status == InvoiceStatusOverdue {
		return true
	}
	return i.DueAt != nil && time.Now().After(*i.DueAt) && !i.IsPaid()
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
	return i.IsSent() && !i.IsPaid() && !i.IsVoided()
}

// CanBeVoided checks if the invoice can be voided
func (i *Invoice) CanBeVoided() bool {
	return !i.IsPaid() && !i.IsVoided()
}

// IsFullyPaid checks if the invoice is fully paid
func (i *Invoice) IsFullyPaid() bool {
	return i.PaidAmount >= i.TotalAmount
}

// GetRemainingAmount returns the remaining amount to be paid
func (i *Invoice) GetRemainingAmount() float64 {
	remaining := i.TotalAmount - i.PaidAmount
	if remaining < 0 {
		return 0
	}
	return remaining
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

// IsDeleted checks if the invoice is soft deleted
func (i *Invoice) IsDeleted() bool {
	return i.DeletedAt.Valid
}

// InvoiceResponse represents the invoice data structure for API responses
type InvoiceResponse struct {
	ID            uint       `json:"id" example:"1"`
	OrderID       uint       `json:"order_id" example:"1"`
	UserID        uint       `json:"user_id" example:"1"`
	InvoiceNumber string     `json:"invoice_number" example:"INV20240101001"`
	InvoiceType   string     `json:"invoice_type" example:"standard"`
	Status        string     `json:"status" example:"sent"`
	
	// Financial Information
	Subtotal    float64 `json:"subtotal" example:"29.99"`
	TaxRate     float64 `json:"tax_rate" example:"0.2"`
	TaxAmount   float64 `json:"tax_amount" example:"5.99"`
	TotalAmount float64 `json:"total_amount" example:"35.98"`
	Currency    string  `json:"currency" example:"USD"`
	
	// Payment Terms Information
	IssuedAt         *time.Time `json:"issued_at,omitempty" example:"2024-01-01T00:00:00Z"`
	DueAt            *time.Time `json:"due_at,omitempty" example:"2024-01-31T23:59:59Z"`
	PaymentTermsDays int        `json:"payment_terms_days" example:"30"`
	
	// Payment Information
	PaidAmount float64    `json:"paid_amount" example:"35.98"`
	PaidAt     *time.Time `json:"paid_at,omitempty" example:"2024-01-15T10:30:00Z"`
	
	// Billing Information
	BillingName    string `json:"billing_name" example:"John Doe"`
	BillingEmail   string `json:"billing_email" example:"john@example.com"`
	BillingAddress string `json:"billing_address,omitempty" example:"123 Main St"`
	BillingCity    string `json:"billing_city,omitempty" example:"New York"`
	BillingState   string `json:"billing_state,omitempty" example:"NY"`
	BillingCountry string `json:"billing_country,omitempty" example:"US"`
	BillingZip     string `json:"billing_zip,omitempty" example:"10001"`
	
	// Tax Information
	TaxNumber string `json:"tax_number,omitempty" example:"GB123456789"`
	TaxType   string `json:"tax_type,omitempty" example:"VAT"`
	
	// Company Information
	CompanyName    string `json:"company_name,omitempty" example:"Acme Corp"`
	CompanyAddress string `json:"company_address,omitempty" example:"456 Business Ave"`
	CompanyTaxID   string `json:"company_tax_id,omitempty" example:"12-3456789"`
	
	// Invoice Content
	Description string `json:"description" example:"Monthly subscription service"`
	
	// Document Management
	PDFPath  string `json:"pdf_path,omitempty" example:"/invoices/INV20240101001.pdf"`
	PDFSize  int    `json:"pdf_size,omitempty" example:"12345"`
	Template string `json:"template" example:"default"`
	Language string `json:"language" example:"en"`
	
	// Sending Records
	SentAt         *time.Time `json:"sent_at,omitempty" example:"2024-01-01T12:00:00Z"`
	SendCount      int        `json:"send_count" example:"1"`
	LastReminderAt *time.Time `json:"last_reminder_at,omitempty"`
	
	// Voiding Information
	VoidedAt   *time.Time `json:"voided_at,omitempty"`
	VoidReason string     `json:"void_reason,omitempty"`
	
	// Business Fields
	Notes         string `json:"notes,omitempty" example:"Thank you for your business"`
	InternalNotes string `json:"internal_notes,omitempty" example:"Customer requested PDF copy"`
	
	CreatedAt time.Time `json:"created_at" example:"2024-01-01T00:00:00Z"`
	UpdatedAt time.Time `json:"updated_at" example:"2024-01-01T00:00:00Z"`
	
	// Related data
	Order *OrderResponse `json:"order,omitempty"`
	User  *UserResponse  `json:"user,omitempty"`
	
	// Computed fields
	IsOverdue       bool    `json:"is_overdue" example:"false"`
	DaysOverdue     int     `json:"days_overdue,omitempty" example:"0"`
	RemainingAmount float64 `json:"remaining_amount" example:"0"`
	DisplayName     string  `json:"display_name" example:"John Doe"`
	FullAddress     string  `json:"full_address" example:"123 Main St\nNew York, NY 10001\nUS"`
	IsFullyPaid     bool    `json:"is_fully_paid" example:"true"`
	CanEdit         bool    `json:"can_edit" example:"false"`
	CanSend         bool    `json:"can_send" example:"false"`
	CanPay          bool    `json:"can_pay" example:"false"`
	CanVoid         bool    `json:"can_void" example:"false"`
}

// ToResponse converts Invoice to InvoiceResponse
func (i *Invoice) ToResponse() *InvoiceResponse {
	resp := &InvoiceResponse{
		ID:               i.ID,
		OrderID:          i.OrderID,
		UserID:           i.UserID,
		InvoiceNumber:    i.InvoiceNumber,
		InvoiceType:      i.InvoiceType,
		Status:           i.Status,
		Subtotal:         i.Subtotal,
		TaxRate:          i.TaxRate,
		TaxAmount:        i.TaxAmount,
		TotalAmount:      i.TotalAmount,
		Currency:         i.Currency,
		IssuedAt:         i.IssuedAt,
		DueAt:            i.DueAt,
		PaymentTermsDays: i.PaymentTermsDays,
		PaidAmount:       i.PaidAmount,
		PaidAt:           i.PaidAt,
		BillingName:      i.BillingName,
		BillingEmail:     i.BillingEmail,
		BillingAddress:   i.BillingAddress,
		BillingCity:      i.BillingCity,
		BillingState:     i.BillingState,
		BillingCountry:   i.BillingCountry,
		BillingZip:       i.BillingZip,
		TaxNumber:        i.TaxNumber,
		TaxType:          i.TaxType,
		CompanyName:      i.CompanyName,
		CompanyAddress:   i.CompanyAddress,
		CompanyTaxID:     i.CompanyTaxID,
		Description:      i.Description,
		PDFPath:          i.PDFPath,
		PDFSize:          i.PDFSize,
		Template:         i.Template,
		Language:         i.Language,
		SentAt:           i.SentAt,
		SendCount:        i.SendCount,
		LastReminderAt:   i.LastReminderAt,
		VoidedAt:         i.VoidedAt,
		VoidReason:       i.VoidReason,
		Notes:            i.Notes,
		InternalNotes:    i.InternalNotes,
		CreatedAt:        i.CreatedAt,
		UpdatedAt:        i.UpdatedAt,
		
		// Computed fields
		IsOverdue:       i.IsOverdue(),
		RemainingAmount: i.GetRemainingAmount(),
		DisplayName:     i.GetDisplayName(),
		FullAddress:     i.GetFullAddress(),
		IsFullyPaid:     i.IsFullyPaid(),
		CanEdit:         i.CanBeEdited(),
		CanSend:         i.CanBeSent(),
		CanPay:          i.CanBePaid(),
		CanVoid:         i.CanBeVoided(),
	}
	
	// Calculate days overdue
	if i.IsOverdue() && i.DueAt != nil {
		days := int(time.Since(*i.DueAt).Hours() / 24)
		if days > 0 {
			resp.DaysOverdue = days
		}
	}
	
	// Include related data if loaded
	if i.Order != nil {
		resp.Order = i.Order.ToResponse()
	}
	if i.User != nil {
		resp.User = i.User.ToResponse()
	}
	
	return resp
}