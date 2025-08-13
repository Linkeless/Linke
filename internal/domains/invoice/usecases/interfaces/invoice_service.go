package interfaces

import (
	"context"
	"linke/internal/domains/invoice/dto"
	"linke/internal/domains/invoice/entities"
)

// InvoiceService defines invoice-specific operations
type InvoiceService interface {
	// Invoice-specific operations that don't fit generic patterns
	CreateInvoiceFromOrder(ctx context.Context, orderID uint, options *dto.CreateInvoiceRequest) (*entities.Invoice, error)
	GetInvoiceByNumber(ctx context.Context, invoiceNumber string) (*entities.Invoice, error)

	// PDF generation and sending (domain-specific)
	GenerateInvoicePDF(ctx context.Context, invoiceID uint) ([]byte, error)
	GenerateInvoicePDFWithOptions(ctx context.Context, invoiceID uint, options *dto.PDFGenerationRequest) ([]byte, string, error)
	GenerateBulkInvoicePDFs(ctx context.Context, invoiceIDs []uint, options *dto.PDFGenerationRequest) ([]byte, error)
	SendInvoice(ctx context.Context, invoiceID uint, emailRequest *dto.SendInvoiceRequest) error
	SendInvoiceWithPDF(ctx context.Context, invoiceID uint, emailRequest *dto.SendInvoiceRequest, pdfOptions *dto.PDFGenerationRequest) error
	ResendInvoice(ctx context.Context, invoiceID uint) error

	// Advanced PDF and download features
	GetInvoicePDFCached(ctx context.Context, invoiceID uint, template string) ([]byte, error)
	DownloadInvoiceAsZip(ctx context.Context, invoiceIDs []uint) ([]byte, string, error)
	GetInvoiceDownloadHistory(ctx context.Context, userID uint) ([]*dto.InvoiceDownloadRecord, error)

	// Template and language support
	GetAvailableTemplates(ctx context.Context) ([]string, error)
	GetAvailableLanguages(ctx context.Context) ([]string, error)
	ValidateTemplate(ctx context.Context, template string) (bool, error)

	// Invoice status management (extends generic status management)
	MarkInvoiceAsPaid(ctx context.Context, invoiceID uint, paymentDate string) error
	MarkInvoiceAsVoid(ctx context.Context, invoiceID uint, reason string) error
	MarkInvoiceAsOverdue(ctx context.Context, invoiceID uint) error

	// Legacy method support for backward compatibility
	CreateInvoice(ctx context.Context, req *dto.CreateInvoiceRequest) (*entities.Invoice, error)
	GetInvoice(ctx context.Context, invoiceID uint) (*entities.Invoice, error)
	UpdateInvoice(ctx context.Context, invoiceID uint, req *dto.UpdateInvoiceRequest) (*entities.Invoice, error)
	DeleteInvoice(ctx context.Context, invoiceID uint) error
	GetInvoices(ctx context.Context, req *dto.GetInvoicesRequest) ([]*entities.Invoice, int64, error)
	GetUserInvoices(ctx context.Context, userID uint, limit, offset int) ([]*entities.Invoice, int64, error)
	GetInvoiceStatistics(ctx context.Context, fromDate, toDate string) (map[string]any, error)
	GetUserInvoiceStatistics(ctx context.Context, userID uint) (map[string]any, error)
}
