package interfaces

import (
	"context"
	"linke/internal/domains/invoice/entities"
)

// InvoiceService defines the interface for invoice operations
type InvoiceService interface {
	// Invoice CRUD operations
	CreateInvoice(ctx context.Context, req *CreateInvoiceRequest) (*entities.Invoice, error)
	GetInvoice(ctx context.Context, invoiceID uint) (*entities.Invoice, error)
	GetInvoiceByNumber(ctx context.Context, invoiceNumber string) (*entities.Invoice, error)
	UpdateInvoice(ctx context.Context, invoiceID uint, req *UpdateInvoiceRequest) (*entities.Invoice, error)
	DeleteInvoice(ctx context.Context, invoiceID uint) error

	// Invoice listing and filtering
	GetInvoices(ctx context.Context, req *GetInvoicesRequest) ([]*entities.Invoice, int64, error)
	GetUserInvoices(ctx context.Context, userID uint, limit, offset int) ([]*entities.Invoice, int64, error)

	// Invoice generation and sending
	GenerateInvoicePDF(ctx context.Context, invoiceID uint) ([]byte, error)
	SendInvoice(ctx context.Context, invoiceID uint, emailRequest *SendInvoiceRequest) error
	ResendInvoice(ctx context.Context, invoiceID uint) error

	// Invoice status management
	MarkInvoiceAsPaid(ctx context.Context, invoiceID uint, paymentDate string) error
	MarkInvoiceAsVoid(ctx context.Context, invoiceID uint, reason string) error
	MarkInvoiceAsOverdue(ctx context.Context, invoiceID uint) error

	// Invoice statistics
	GetInvoiceStatistics(ctx context.Context, fromDate, toDate string) (map[string]interface{}, error)
	GetUserInvoiceStatistics(ctx context.Context, userID uint) (map[string]interface{}, error)
}

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
