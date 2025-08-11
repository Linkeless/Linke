package dto

import (
	"time"

	"linke/internal/domains/invoice/entities"
	"linke/internal/shared/dto"
)

// ==================== Invoice Response DTO ====================

// InvoiceResponse represents the invoice data structure for API responses
type InvoiceResponse struct {
	ID                  uint       `json:"id" example:"1"`
	UserID              uint       `json:"user_id" example:"1"`
	SubscriptionOrderID uint       `json:"subscription_order_id" example:"1"`
	InvoiceNumber       string     `json:"invoice_number" example:"INV-2024-001"`
	InvoiceType         string     `json:"invoice_type" example:"standard"`
	Status              string     `json:"status" example:"sent"`
	Amount              float64    `json:"amount" example:"29.99"`
	Currency            string     `json:"currency" example:"USD"`
	TaxAmount           float64    `json:"tax_amount" example:"5.99"`
	TotalAmount         float64    `json:"total_amount" example:"35.98"`
	TaxRate             float64    `json:"tax_rate" example:"0.2"`
	TaxType             string     `json:"tax_type,omitempty" example:"VAT"`
	TaxNumber           string     `json:"tax_number,omitempty" example:"GB123456789"`
	BillingName         string     `json:"billing_name" example:"John Doe"`
	BillingEmail        string     `json:"billing_email" example:"john@example.com"`
	BillingAddress      string     `json:"billing_address,omitempty" example:"123 Main St"`
	BillingCity         string     `json:"billing_city,omitempty" example:"New York"`
	BillingState        string     `json:"billing_state,omitempty" example:"NY"`
	BillingCountry      string     `json:"billing_country,omitempty" example:"US"`
	BillingZip          string     `json:"billing_zip,omitempty" example:"10001"`
	CompanyName         string     `json:"company_name,omitempty" example:"Acme Corp"`
	CompanyTaxID        string     `json:"company_tax_id,omitempty" example:"12-3456789"`
	CompanyAddress      string     `json:"company_address,omitempty" example:"456 Business Ave"`
	IssuedAt            time.Time  `json:"issued_at" swaggertype:"string" format:"date-time" example:"2024-01-01T00:00:00Z"`
	DueAt               *time.Time `json:"due_at,omitempty" swaggertype:"string" format:"date-time" example:"2024-01-31T23:59:59Z"`
	PaidAt              *time.Time `json:"paid_at,omitempty" swaggertype:"string" format:"date-time" example:"2024-01-15T10:30:00Z"`
	SentAt              *time.Time `json:"sent_at,omitempty" swaggertype:"string" format:"date-time" example:"2024-01-01T12:00:00Z"`
	VoidedAt            *time.Time `json:"voided_at,omitempty" swaggertype:"string" format:"date-time"`
	PaymentMethod       string     `json:"payment_method,omitempty" example:"credit_card"`
	PaymentReference    string     `json:"payment_reference,omitempty" example:"txn_123456"`
	Template            string     `json:"template,omitempty" example:"default"`
	Language            string     `json:"language,omitempty" example:"en"`
	PDFPath             string     `json:"pdf_path,omitempty" example:"/invoices/INV-2024-001.pdf"`
	PDFSize             int64      `json:"pdf_size,omitempty" example:"12345"`
	Description         string     `json:"description,omitempty" example:"Monthly subscription"`
	Notes               string     `json:"notes,omitempty" example:"Thank you for your business"`
	CreatedAt           time.Time  `json:"created_at" swaggertype:"string" format:"date-time" example:"2024-01-01T00:00:00Z"`
	UpdatedAt           time.Time  `json:"updated_at" swaggertype:"string" format:"date-time" example:"2024-01-01T00:00:00Z"`

	// Related data (to be populated at application layer)
	User              *dto.UserBasicDTO              `json:"user,omitempty"`
	SubscriptionOrder *dto.SubscriptionOrderBasicDTO `json:"subscription_order,omitempty"`

	// Computed fields
	IsOverdue   bool   `json:"is_overdue"`
	DaysOverdue int    `json:"days_overdue,omitempty"`
	DisplayName string `json:"display_name"`
	FullAddress string `json:"full_address"`
}

// ToResponse converts Invoice to InvoiceResponse
func ToResponse(i *entities.Invoice) *InvoiceResponse {
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

	// Note: Related data should be populated at the application layer
	// to avoid cross-domain dependencies

	return resp
}

// ==================== Request DTOs ====================

// CreateInvoiceRequest represents the request to create an invoice
type CreateInvoiceRequest struct {
	UserID              uint   `json:"user_id" binding:"required"`
	SubscriptionOrderID uint   `json:"subscription_order_id" binding:"required"`
	InvoiceType         string `json:"invoice_type,omitempty" example:"standard"`

	// Financial Details
	Amount    float64 `json:"amount" binding:"required,min=0"`
	Currency  string  `json:"currency,omitempty" example:"USD"`
	TaxRate   float64 `json:"tax_rate,omitempty" example:"0.2"`
	TaxType   string  `json:"tax_type,omitempty" example:"VAT"`
	TaxNumber string  `json:"tax_number,omitempty" example:"GB123456789"`

	// Billing Information
	BillingName    string `json:"billing_name" binding:"required"`
	BillingEmail   string `json:"billing_email" binding:"required,email"`
	BillingAddress string `json:"billing_address,omitempty"`
	BillingCity    string `json:"billing_city,omitempty"`
	BillingState   string `json:"billing_state,omitempty"`
	BillingCountry string `json:"billing_country,omitempty"`
	BillingZip     string `json:"billing_zip,omitempty"`

	// Company Information
	CompanyName    string `json:"company_name,omitempty"`
	CompanyTaxID   string `json:"company_tax_id,omitempty"`
	CompanyAddress string `json:"company_address,omitempty"`

	// Additional Information
	Description string `json:"description,omitempty"`
	Notes       string `json:"notes,omitempty"`
	DueDate     string `json:"due_date,omitempty" example:"2024-01-31"`
	Template    string `json:"template,omitempty" example:"default"`
	Language    string `json:"language,omitempty" example:"en"`
	AutoSend    bool   `json:"auto_send,omitempty" example:"false"`
}

// UpdateInvoiceRequest represents the request to update an invoice
type UpdateInvoiceRequest struct {
	InvoiceType *string  `json:"invoice_type,omitempty" example:"credit"`
	Amount      *float64 `json:"amount,omitempty" example:"100.00"`
	TaxRate     *float64 `json:"tax_rate,omitempty" example:"0.1"`
	TaxType     *string  `json:"tax_type,omitempty" example:"GST"`

	// Billing Information
	BillingName    *string `json:"billing_name,omitempty"`
	BillingEmail   *string `json:"billing_email,omitempty"`
	BillingAddress *string `json:"billing_address,omitempty"`
	BillingCity    *string `json:"billing_city,omitempty"`
	BillingState   *string `json:"billing_state,omitempty"`
	BillingCountry *string `json:"billing_country,omitempty"`
	BillingZip     *string `json:"billing_zip,omitempty"`

	// Company Information
	CompanyName    *string `json:"company_name,omitempty"`
	CompanyTaxID   *string `json:"company_tax_id,omitempty"`
	CompanyAddress *string `json:"company_address,omitempty"`

	// Additional Information
	Description *string `json:"description,omitempty"`
	Notes       *string `json:"notes,omitempty"`
	DueDate     *string `json:"due_date,omitempty"`
	Template    *string `json:"template,omitempty"`
	Language    *string `json:"language,omitempty"`
}

// GetInvoicesRequest represents the request to get invoices
type GetInvoicesRequest struct {
	UserID      uint   `form:"user_id,omitempty" example:"1"`
	Status      string `form:"status,omitempty" example:"pending"`
	InvoiceType string `form:"invoice_type,omitempty" example:"standard"`
	DateFrom    string `form:"date_from,omitempty" example:"2024-01-01"`
	DateTo      string `form:"date_to,omitempty" example:"2024-12-31"`
	Limit       int    `form:"limit,omitempty" example:"10"`
	Offset      int    `form:"offset,omitempty" example:"0"`
}

// SendInvoiceRequest represents the request to send an invoice via email
type SendInvoiceRequest struct {
	ToEmail  string `json:"to_email" binding:"required,email" example:"customer@example.com"`
	CcEmails string `json:"cc_emails,omitempty" example:"manager@example.com"`
	Subject  string `json:"subject,omitempty" example:"Your Invoice"`
	Message  string `json:"message,omitempty" example:"Please find your invoice attached"`
	SendCopy bool   `json:"send_copy,omitempty" example:"true"`
}

// PDFGenerationRequest represents options for PDF generation
type PDFGenerationRequest struct {
	Template     string            `json:"template,omitempty" example:"professional"`
	Language     string            `json:"language,omitempty" example:"en"`
	Watermark    string            `json:"watermark,omitempty" example:"DRAFT"`
	SaveToDisk   bool              `json:"save_to_disk,omitempty" example:"false"`
	IncludeQR    bool              `json:"include_qr,omitempty" example:"true"`
	CompanyInfo  *CompanyInfo      `json:"company_info,omitempty"`
	CustomFields map[string]string `json:"custom_fields,omitempty"`
}

// CompanyInfo contains company information for invoice PDF
type CompanyInfo struct {
	Name          string `json:"name,omitempty" example:"Acme Corp"`
	Address       string `json:"address,omitempty" example:"123 Business Ave"`
	City          string `json:"city,omitempty" example:"New York"`
	State         string `json:"state,omitempty" example:"NY"`
	ZIP           string `json:"zip,omitempty" example:"10001"`
	Country       string `json:"country,omitempty" example:"US"`
	Phone         string `json:"phone,omitempty" example:"+1-555-123-4567"`
	Email         string `json:"email,omitempty" example:"contact@acme.com"`
	Website       string `json:"website,omitempty" example:"https://acme.com"`
	TaxID         string `json:"tax_id,omitempty" example:"12-3456789"`
	BankAccount   string `json:"bank_account,omitempty" example:"1234567890"`
	RoutingNumber string `json:"routing_number,omitempty" example:"123456789"`
	Logo          string `json:"logo,omitempty" example:"./assets/logo.png"`
}

// InvoiceDownloadRecord represents a download history record
type InvoiceDownloadRecord struct {
	ID           uint      `json:"id" example:"1"`
	UserID       uint      `json:"user_id" example:"1"`
	InvoiceID    uint      `json:"invoice_id" example:"1"`
	Template     string    `json:"template" example:"professional"`
	Language     string    `json:"language" example:"en"`
	IPAddress    string    `json:"ip_address" example:"192.168.1.1"`
	UserAgent    string    `json:"user_agent" example:"Mozilla/5.0..."`
	DownloadedAt time.Time `json:"downloaded_at" example:"2024-01-01T00:00:00Z"`
}

// BulkDownloadRequest represents a request for bulk download
type BulkDownloadRequest struct {
	InvoiceIDs []uint                `json:"invoice_ids" binding:"required"`
	PDFOptions *PDFGenerationRequest `json:"pdf_options,omitempty"`
	Format     string                `json:"format,omitempty" example:"zip"` // zip, individual
	IncludeCSV bool                  `json:"include_csv,omitempty" example:"true"`
}