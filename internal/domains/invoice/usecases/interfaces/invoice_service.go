package interfaces

import (
	"context"
	"linke/internal/domains/invoice/entities"
	"time"
)

// InvoiceService defines the interface for invoice operations
type InvoiceService interface {
	// Invoice CRUD operations
	CreateInvoice(ctx context.Context, req *CreateInvoiceRequest) (*entities.Invoice, error)
	CreateInvoiceFromOrder(ctx context.Context, orderID uint, options *CreateInvoiceRequest) (*entities.Invoice, error)
	GetInvoice(ctx context.Context, invoiceID uint) (*entities.Invoice, error)
	GetInvoiceByNumber(ctx context.Context, invoiceNumber string) (*entities.Invoice, error)
	UpdateInvoice(ctx context.Context, invoiceID uint, req *UpdateInvoiceRequest) (*entities.Invoice, error)
	DeleteInvoice(ctx context.Context, invoiceID uint) error

	// Invoice listing and filtering
	GetInvoices(ctx context.Context, req *GetInvoicesRequest) ([]*entities.Invoice, int64, error)
	GetUserInvoices(ctx context.Context, userID uint, limit, offset int) ([]*entities.Invoice, int64, error)

	// Invoice generation and sending
	GenerateInvoicePDF(ctx context.Context, invoiceID uint) ([]byte, error)
	GenerateInvoicePDFWithOptions(ctx context.Context, invoiceID uint, options *PDFGenerationRequest) ([]byte, string, error)
	GenerateBulkInvoicePDFs(ctx context.Context, invoiceIDs []uint, options *PDFGenerationRequest) ([]byte, error) // Returns ZIP
	SendInvoice(ctx context.Context, invoiceID uint, emailRequest *SendInvoiceRequest) error
	SendInvoiceWithPDF(ctx context.Context, invoiceID uint, emailRequest *SendInvoiceRequest, pdfOptions *PDFGenerationRequest) error
	ResendInvoice(ctx context.Context, invoiceID uint) error

	// Advanced PDF and download features
	GetInvoicePDFCached(ctx context.Context, invoiceID uint, template string) ([]byte, error)
	DownloadInvoiceAsZip(ctx context.Context, invoiceIDs []uint) ([]byte, string, error)
	GetInvoiceDownloadHistory(ctx context.Context, userID uint) ([]*InvoiceDownloadRecord, error)

	// Template and language support
	GetAvailableTemplates(ctx context.Context) ([]string, error)
	GetAvailableLanguages(ctx context.Context) ([]string, error)
	ValidateTemplate(ctx context.Context, template string) (bool, error)

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
	InvoiceIDs []uint                `json:"invoice_ids" binding:"required" example:"[1,2,3]"`
	PDFOptions *PDFGenerationRequest `json:"pdf_options,omitempty"`
	Format     string                `json:"format,omitempty" example:"zip"` // zip, individual
	IncludeCSV bool                  `json:"include_csv,omitempty" example:"true"`
}
