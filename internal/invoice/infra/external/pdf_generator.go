package external

import (
	"context"
	"fmt"

	"linke/internal/invoice/domain/model"
)

// PDFGenerator defines the interface for generating invoice PDFs
type PDFGenerator interface {
	GenerateInvoicePDF(ctx context.Context, invoice *model.Invoice) ([]byte, error)
	SaveInvoicePDF(ctx context.Context, invoice *model.Invoice, pdfData []byte) (string, error)
}

// EmailService defines the interface for sending invoice emails
type EmailService interface {
	SendInvoice(ctx context.Context, invoice *model.Invoice, recipientEmail string, subject, message string, ccEmails []string) error
	SendInvoiceReminder(ctx context.Context, invoice *model.Invoice, recipientEmail string) error
}

// DefaultPDFGenerator provides a default implementation for PDF generation
type DefaultPDFGenerator struct {
	storagePath string
}

// NewDefaultPDFGenerator creates a new default PDF generator
func NewDefaultPDFGenerator(storagePath string) *DefaultPDFGenerator {
	return &DefaultPDFGenerator{
		storagePath: storagePath,
	}
}

// GenerateInvoicePDF generates a PDF for the given invoice
func (g *DefaultPDFGenerator) GenerateInvoicePDF(ctx context.Context, invoice *model.Invoice) ([]byte, error) {
	// TODO: Implement actual PDF generation using a library like gofpdf or wkhtmltopdf
	// For now, return a placeholder
	return []byte("PDF placeholder for invoice " + invoice.InvoiceNumber().String()), nil
}

// SaveInvoicePDF saves the PDF data to storage and returns the file path
func (g *DefaultPDFGenerator) SaveInvoicePDF(ctx context.Context, invoice *model.Invoice, pdfData []byte) (string, error) {
	// TODO: Implement actual file saving logic
	// This should save to the configured storage path and return the full path
	filename := fmt.Sprintf("invoice_%s.pdf", invoice.InvoiceNumber().String())
	filepath := fmt.Sprintf("%s/%s", g.storagePath, filename)
	
	// Placeholder implementation
	// In a real implementation, you would:
	// 1. Write pdfData to the file system or cloud storage
	// 2. Return the accessible file path/URL
	
	return filepath, nil
}

// DefaultEmailService provides a default implementation for email services
type DefaultEmailService struct {
	smtpHost     string
	smtpPort     int
	smtpUsername string
	smtpPassword string
	fromEmail    string
	fromName     string
}

// NewDefaultEmailService creates a new default email service
func NewDefaultEmailService(
	smtpHost string,
	smtpPort int,
	smtpUsername, smtpPassword string,
	fromEmail, fromName string,
) *DefaultEmailService {
	return &DefaultEmailService{
		smtpHost:     smtpHost,
		smtpPort:     smtpPort,
		smtpUsername: smtpUsername,
		smtpPassword: smtpPassword,
		fromEmail:    fromEmail,
		fromName:     fromName,
	}
}

// SendInvoice sends an invoice via email
func (s *DefaultEmailService) SendInvoice(
	ctx context.Context,
	invoice *model.Invoice,
	recipientEmail string,
	subject, message string,
	ccEmails []string,
) error {
	// TODO: Implement actual email sending using SMTP or email service API
	// This should:
	// 1. Create email with invoice PDF attachment
	// 2. Send to recipient with optional CC
	// 3. Handle delivery status and errors
	
	fmt.Printf("Sending invoice %s to %s\n", invoice.InvoiceNumber().String(), recipientEmail)
	fmt.Printf("Subject: %s\n", subject)
	fmt.Printf("Message: %s\n", message)
	if len(ccEmails) > 0 {
		fmt.Printf("CC: %v\n", ccEmails)
	}
	
	return nil
}

// SendInvoiceReminder sends a reminder email for an overdue invoice
func (s *DefaultEmailService) SendInvoiceReminder(
	ctx context.Context,
	invoice *model.Invoice,
	recipientEmail string,
) error {
	// TODO: Implement reminder email logic
	// This should send a templated reminder email
	
	fmt.Printf("Sending reminder for invoice %s to %s\n", 
		invoice.InvoiceNumber().String(), recipientEmail)
	
	return nil
}