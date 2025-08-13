package implementations

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"linke/internal/domains/invoice/constants"
	"linke/internal/domains/invoice/entities"
	"linke/internal/shared/logger"

	"github.com/jung-kurt/gofpdf/v2"
	"go.uber.org/zap"
)

// PDFGeneratorService handles PDF generation for invoices
type PDFGeneratorService struct {
	outputDir string
	logger    logger.Logger
}

// NewPDFGeneratorService creates a new PDF generator service
func NewPDFGeneratorService(outputDir string, logger logger.Logger) *PDFGeneratorService {
	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		logger.Error("Failed to create PDF output directory", zap.Error(err))
	}

	return &PDFGeneratorService{
		outputDir: outputDir,
		logger:    logger,
	}
}

// PDFGenerationOptions contains options for PDF generation
type PDFGenerationOptions struct {
	Template     string
	Language     string
	Watermark    string
	SaveToDisk   bool
	IncludeQR    bool
	CompanyInfo  *CompanyInfo
	CustomFields map[string]string
}

// CompanyInfo contains company information for invoice PDF
type CompanyInfo struct {
	Name          string
	Address       string
	City          string
	State         string
	ZIP           string
	Country       string
	Phone         string
	Email         string
	Website       string
	TaxID         string
	BankAccount   string
	RoutingNumber string
	Logo          string // Path to logo file
}

// InvoiceLineItem represents a line item in the invoice
type InvoiceLineItem struct {
	Description string
	Quantity    float64
	UnitPrice   float64
	Amount      float64
}

// GeneratePDF generates a PDF for the given invoice
func (pgs *PDFGeneratorService) GeneratePDF(ctx context.Context, invoice *entities.Invoice, options *PDFGenerationOptions) ([]byte, string, error) {
	if options == nil {
		options = &PDFGenerationOptions{
			Template:   "default",
			Language:   "en",
			SaveToDisk: false,
		}
	}

	// Create PDF
	pdf := gofpdf.New("P", "mm", "A4", "")

	// Add page
	pdf.AddPage()

	// Generate content based on template
	switch options.Template {
	case "professional":
		if err := pgs.generateProfessionalTemplate(pdf, invoice, options); err != nil {
			return nil, "", fmt.Errorf("failed to generate professional template: %w", err)
		}
	case "minimal":
		if err := pgs.generateMinimalTemplate(pdf, invoice, options); err != nil {
			return nil, "", fmt.Errorf("failed to generate minimal template: %w", err)
		}
	default:
		if err := pgs.generateDefaultTemplate(pdf, invoice, options); err != nil {
			return nil, "", fmt.Errorf("failed to generate default template: %w", err)
		}
	}

	// Add watermark if needed
	if options.Watermark != "" {
		pgs.addWatermark(pdf, options.Watermark)
	}

	// Generate output
	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, "", fmt.Errorf("failed to generate PDF output: %w", err)
	}

	pdfBytes := buf.Bytes()

	// Save to disk if requested
	var filePath string
	if options.SaveToDisk {
		fileName := fmt.Sprintf("invoice_%s_%d.pdf",
			invoice.InvoiceNumber,
			time.Now().Unix())
		filePath = filepath.Join(pgs.outputDir, fileName)

		if err := os.WriteFile(filePath, pdfBytes, 0644); err != nil {
			pgs.logger.Error("Failed to save PDF to disk",
				zap.Error(err),
				zap.String("file_path", filePath))
		}
	}

	pgs.logger.Info("PDF generated successfully",
		zap.String("invoice_number", invoice.InvoiceNumber),
		zap.String("template", options.Template),
		zap.Int("size_bytes", len(pdfBytes)))

	return pdfBytes, filePath, nil
}

// generateDefaultTemplate generates the default invoice template
func (pgs *PDFGeneratorService) generateDefaultTemplate(pdf *gofpdf.Fpdf, invoice *entities.Invoice, options *PDFGenerationOptions) error {
	// Set font
	pdf.SetFont("Arial", "", 12)

	// Header
	pdf.SetFont("Arial", "B", 20)
	pdf.Cell(0, 15, pgs.getLocalizedText("INVOICE", options.Language))
	pdf.Ln(20)

	// Company info section
	if options.CompanyInfo != nil {
		pgs.addCompanyInfo(pdf, options.CompanyInfo, options.Language)
		pdf.Ln(10)
	}

	// Invoice details
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(0, 8, fmt.Sprintf("%s: %s",
		pgs.getLocalizedText("Invoice Number", options.Language),
		invoice.InvoiceNumber))
	pdf.Ln(8)

	// Order information
	if invoice.SubscriptionOrderID > 0 {
		pdf.SetFont("Arial", "", 10)
		pdf.Cell(0, 6, fmt.Sprintf("%s: #%d",
			pgs.getLocalizedText("Order Reference", options.Language),
			invoice.SubscriptionOrderID))
		pdf.Ln(6)
	}

	pdf.SetFont("Arial", "", 10)
	pdf.Cell(0, 6, fmt.Sprintf("%s: %s",
		pgs.getLocalizedText("Date", options.Language),
		invoice.IssuedAt.Format("2006-01-02")))
	pdf.Ln(6)

	if invoice.DueAt != nil {
		pdf.Cell(0, 6, fmt.Sprintf("%s: %s",
			pgs.getLocalizedText("Due Date", options.Language),
			invoice.DueAt.Format("2006-01-02")))
		pdf.Ln(6)
	}

	pdf.Ln(10)

	// Billing information
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(0, 8, pgs.getLocalizedText("Bill To", options.Language))
	pdf.Ln(8)

	pdf.SetFont("Arial", "", 10)
	pdf.Cell(0, 6, invoice.GetDisplayName())
	pdf.Ln(6)

	if invoice.BillingEmail != "" {
		pdf.Cell(0, 6, invoice.BillingEmail)
		pdf.Ln(6)
	}

	if invoice.GetFullAddress() != "" {
		// Split address into lines
		addressLines := strings.Split(invoice.GetFullAddress(), "\n")
		for _, line := range addressLines {
			if strings.TrimSpace(line) != "" {
				pdf.Cell(0, 6, strings.TrimSpace(line))
				pdf.Ln(6)
			}
		}
	}

	pdf.Ln(10)

	// Invoice items table
	pgs.addInvoiceTable(pdf, invoice, options.Language)

	// Add QR code if requested
	if options.IncludeQR {
		pgs.addQRCode(pdf, invoice)
	}

	// Footer
	pgs.addFooter(pdf, invoice, options)

	return nil
}

// generateProfessionalTemplate generates a professional invoice template
func (pgs *PDFGeneratorService) generateProfessionalTemplate(pdf *gofpdf.Fpdf, invoice *entities.Invoice, options *PDFGenerationOptions) error {
	// Set colors
	pdf.SetFillColor(240, 240, 240) // Light gray
	pdf.SetTextColor(60, 60, 60)    // Dark gray
	pdf.SetDrawColor(200, 200, 200) // Light gray for borders

	// Header with company info
	if options.CompanyInfo != nil && options.CompanyInfo.Logo != "" {
		// Add logo if available
		// pdf.ImageOptions(options.CompanyInfo.Logo, 15, 15, 30, 0, false, gofpdf.ImageOptions{ImageType: "auto"}, 0, "")
	}

	pdf.SetFont("Arial", "B", 24)
	pdf.SetTextColor(70, 130, 180) // Steel blue
	pdf.Cell(0, 15, "INVOICE")
	pdf.Ln(20)

	// Reset text color
	pdf.SetTextColor(60, 60, 60)

	// Two column layout for header info
	pdf.SetFont("Arial", "B", 11)

	// Left column - Company info
	currentY := pdf.GetY()
	if options.CompanyInfo != nil {
		pdf.SetXY(15, currentY)
		pgs.addCompanyInfoProfessional(pdf, options.CompanyInfo)
	}

	// Right column - Invoice details
	pdf.SetXY(120, currentY)
	pgs.addInvoiceDetailsProfessional(pdf, invoice, options.Language)

	// Move to next section
	pdf.SetY(currentY + 40)
	pdf.Ln(10)

	// Billing information in a box
	pdf.SetFillColor(248, 248, 248)
	pdf.Rect(15, pdf.GetY(), 180, 30, "F")

	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(0, 8, pgs.getLocalizedText("Bill To", options.Language))
	pdf.Ln(8)

	pdf.SetFont("Arial", "", 10)
	pdf.Cell(0, 6, invoice.GetDisplayName())
	pdf.Ln(6)
	pdf.Cell(0, 6, invoice.BillingEmail)
	pdf.Ln(6)

	if invoice.GetFullAddress() != "" {
		addressLines := strings.Split(invoice.GetFullAddress(), "\n")
		for i, line := range addressLines {
			if i < 2 && strings.TrimSpace(line) != "" { // Limit to 2 lines
				pdf.Cell(0, 6, strings.TrimSpace(line))
				pdf.Ln(6)
			}
		}
	}

	pdf.Ln(15)

	// Professional invoice table
	pgs.addProfessionalInvoiceTable(pdf, invoice, options.Language)

	// Professional footer
	pgs.addProfessionalFooter(pdf, invoice, options)

	return nil
}

// generateMinimalTemplate generates a minimal invoice template
func (pgs *PDFGeneratorService) generateMinimalTemplate(pdf *gofpdf.Fpdf, invoice *entities.Invoice, options *PDFGenerationOptions) error {
	pdf.SetFont("Arial", "", 10)

	// Minimal header
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(0, 10, fmt.Sprintf("Invoice %s", invoice.InvoiceNumber))
	pdf.Ln(15)

	// Basic info
	pdf.SetFont("Arial", "", 9)
	pdf.Cell(0, 5, fmt.Sprintf("Date: %s", invoice.IssuedAt.Format("2006-01-02")))
	pdf.Ln(5)

	// Add order reference if available
	if invoice.SubscriptionOrderID > 0 {
		pdf.Cell(0, 5, fmt.Sprintf("Order: #%d", invoice.SubscriptionOrderID))
		pdf.Ln(5)
	}

	if invoice.DueAt != nil {
		pdf.Cell(0, 5, fmt.Sprintf("Due: %s", invoice.DueAt.Format("2006-01-02")))
		pdf.Ln(5)
	}

	pdf.Ln(10)

	// Billing info
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(0, 6, "To:")
	pdf.Ln(6)

	pdf.SetFont("Arial", "", 9)
	pdf.Cell(0, 5, invoice.GetDisplayName())
	pdf.Ln(5)
	pdf.Cell(0, 5, invoice.BillingEmail)
	pdf.Ln(10)

	// Simple table
	pgs.addMinimalInvoiceTable(pdf, invoice, options.Language)

	return nil
}

// addCompanyInfo adds company information to the PDF
func (pgs *PDFGeneratorService) addCompanyInfo(pdf *gofpdf.Fpdf, company *CompanyInfo, language string) {
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(0, 8, company.Name)
	pdf.Ln(8)

	pdf.SetFont("Arial", "", 10)
	if company.Address != "" {
		pdf.Cell(0, 6, company.Address)
		pdf.Ln(6)
	}
	if company.City != "" {
		cityLine := company.City
		if company.State != "" {
			cityLine += ", " + company.State
		}
		if company.ZIP != "" {
			cityLine += " " + company.ZIP
		}
		pdf.Cell(0, 6, cityLine)
		pdf.Ln(6)
	}
	if company.Country != "" {
		pdf.Cell(0, 6, company.Country)
		pdf.Ln(6)
	}
	if company.Phone != "" {
		pdf.Cell(0, 6, fmt.Sprintf("%s: %s", pgs.getLocalizedText("Phone", language), company.Phone))
		pdf.Ln(6)
	}
	if company.Email != "" {
		pdf.Cell(0, 6, fmt.Sprintf("%s: %s", pgs.getLocalizedText("Email", language), company.Email))
		pdf.Ln(6)
	}
}

// addInvoiceTable adds the invoice items table
func (pgs *PDFGeneratorService) addInvoiceTable(pdf *gofpdf.Fpdf, invoice *entities.Invoice, language string) {
	// Table header
	pdf.SetFont("Arial", "B", 10)
	pdf.SetFillColor(220, 220, 220)

	pdf.CellFormat(80, 8, pgs.getLocalizedText("Description", language), "1", 0, "C", true, 0, "")
	pdf.CellFormat(30, 8, pgs.getLocalizedText("Quantity", language), "1", 0, "C", true, 0, "")
	pdf.CellFormat(30, 8, pgs.getLocalizedText("Unit Price", language), "1", 0, "C", true, 0, "")
	pdf.CellFormat(40, 8, pgs.getLocalizedText("Amount", language), "1", 1, "C", true, 0, "")

	// Create line items (in real implementation, these would come from the invoice)
	lineItems := pgs.getInvoiceLineItems(invoice)

	pdf.SetFont("Arial", "", 9)
	pdf.SetFillColor(255, 255, 255)

	for _, item := range lineItems {
		pdf.CellFormat(80, 6, item.Description, "1", 0, "L", false, 0, "")
		pdf.CellFormat(30, 6, fmt.Sprintf("%.2f", item.Quantity), "1", 0, "C", false, 0, "")
		pdf.CellFormat(30, 6, fmt.Sprintf("%.2f %s", item.UnitPrice, invoice.Currency), "1", 0, "R", false, 0, "")
		pdf.CellFormat(40, 6, fmt.Sprintf("%.2f %s", item.Amount, invoice.Currency), "1", 1, "R", false, 0, "")
	}

	// Totals
	pdf.Ln(5)

	if invoice.TaxAmount > 0 {
		pdf.SetFont("Arial", "", 10)
		pdf.Cell(140, 6, "")
		pdf.CellFormat(40, 6, fmt.Sprintf("%s:", pgs.getLocalizedText("Subtotal", language)), "", 0, "R", false, 0, "")
		pdf.CellFormat(40, 6, fmt.Sprintf("%.2f %s", invoice.Amount, invoice.Currency), "", 1, "R", false, 0, "")

		pdf.Cell(140, 6, "")
		pdf.CellFormat(40, 6, fmt.Sprintf("%s (%.1f%%):", pgs.getLocalizedText("Tax", language), invoice.TaxRate*100), "", 0, "R", false, 0, "")
		pdf.CellFormat(40, 6, fmt.Sprintf("%.2f %s", invoice.TaxAmount, invoice.Currency), "", 1, "R", false, 0, "")
	}

	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(140, 8, "")
	pdf.CellFormat(40, 8, fmt.Sprintf("%s:", pgs.getLocalizedText("Total", language)), "", 0, "R", false, 0, "")
	pdf.CellFormat(40, 8, fmt.Sprintf("%.2f %s", invoice.TotalAmount, invoice.Currency), "", 1, "R", false, 0, "")
}

// addProfessionalInvoiceTable adds a professional-styled invoice table
func (pgs *PDFGeneratorService) addProfessionalInvoiceTable(pdf *gofpdf.Fpdf, invoice *entities.Invoice, language string) {
	// Table header with professional styling
	pdf.SetFont("Arial", "B", 10)
	pdf.SetFillColor(70, 130, 180)  // Steel blue
	pdf.SetTextColor(255, 255, 255) // White text

	pdf.CellFormat(80, 10, pgs.getLocalizedText("Description", language), "1", 0, "C", true, 0, "")
	pdf.CellFormat(25, 10, pgs.getLocalizedText("Qty", language), "1", 0, "C", true, 0, "")
	pdf.CellFormat(35, 10, pgs.getLocalizedText("Unit Price", language), "1", 0, "C", true, 0, "")
	pdf.CellFormat(40, 10, pgs.getLocalizedText("Amount", language), "1", 1, "C", true, 0, "")

	// Reset text color
	pdf.SetTextColor(60, 60, 60)

	// Table rows with alternating colors
	lineItems := pgs.getInvoiceLineItems(invoice)
	pdf.SetFont("Arial", "", 9)

	for i, item := range lineItems {
		if i%2 == 0 {
			pdf.SetFillColor(248, 248, 248) // Light gray
		} else {
			pdf.SetFillColor(255, 255, 255) // White
		}

		pdf.CellFormat(80, 8, item.Description, "1", 0, "L", true, 0, "")
		pdf.CellFormat(25, 8, fmt.Sprintf("%.0f", item.Quantity), "1", 0, "C", true, 0, "")
		pdf.CellFormat(35, 8, fmt.Sprintf("%.2f", item.UnitPrice), "1", 0, "R", true, 0, "")
		pdf.CellFormat(40, 8, fmt.Sprintf("%.2f", item.Amount), "1", 1, "R", true, 0, "")
	}

	// Professional totals section
	pdf.Ln(5)
	pgs.addProfessionalTotals(pdf, invoice, language)
}

// addMinimalInvoiceTable adds a minimal invoice table
func (pgs *PDFGeneratorService) addMinimalInvoiceTable(pdf *gofpdf.Fpdf, invoice *entities.Invoice, language string) {
	lineItems := pgs.getInvoiceLineItems(invoice)

	pdf.SetFont("Arial", "", 9)
	for _, item := range lineItems {
		pdf.Cell(0, 5, fmt.Sprintf("%s - %.2f %s", item.Description, item.Amount, invoice.Currency))
		pdf.Ln(5)
	}

	pdf.Ln(5)
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(0, 6, fmt.Sprintf("Total: %.2f %s", invoice.TotalAmount, invoice.Currency))
	pdf.Ln(6)
}

// getInvoiceLineItems creates line items from invoice data
func (pgs *PDFGeneratorService) getInvoiceLineItems(invoice *entities.Invoice) []InvoiceLineItem {
	// In a real implementation, this would come from a related line items table
	// For now, create a default line item based on the invoice description
	description := invoice.Description
	if description == "" {
		description = "Subscription Service"
	}

	return []InvoiceLineItem{
		{
			Description: description,
			Quantity:    1,
			UnitPrice:   invoice.Amount,
			Amount:      invoice.Amount,
		},
	}
}

// addWatermark adds a watermark to the PDF
func (pgs *PDFGeneratorService) addWatermark(pdf *gofpdf.Fpdf, watermark string) {
	pdf.SetFont("Arial", "B", 50)
	pdf.SetTextColor(200, 200, 200) // Light gray

	// Rotate text for diagonal watermark
	pdf.TransformBegin()
	pdf.TransformRotate(45, 105, 150)
	pdf.Text(50, 150, watermark)
	pdf.TransformEnd()
}

// addQRCode adds a QR code to the invoice (placeholder)
func (pgs *PDFGeneratorService) addQRCode(pdf *gofpdf.Fpdf, invoice *entities.Invoice) {
	// This would require a QR code library to implement properly
	// For now, just add a placeholder
	currentY := pdf.GetY()
	pdf.SetXY(160, currentY+10)
	pdf.SetFont("Arial", "", 8)
	pdf.Cell(0, 4, "QR Code would be here")
	pdf.Ln(4)
	pdf.Cell(0, 4, fmt.Sprintf("Invoice: %s", invoice.InvoiceNumber))
}

// addFooter adds footer information
func (pgs *PDFGeneratorService) addFooter(pdf *gofpdf.Fpdf, invoice *entities.Invoice, options *PDFGenerationOptions) {
	pdf.Ln(20)
	pdf.SetFont("Arial", "", 8)
	pdf.SetTextColor(100, 100, 100)

	if invoice.Notes != "" {
		pdf.Cell(0, 4, pgs.getLocalizedText("Notes", options.Language)+": "+invoice.Notes)
		pdf.Ln(4)
	}

	pdf.Cell(0, 4, pgs.getLocalizedText("Thank you for your business!", options.Language))
	pdf.Ln(4)

	pdf.Cell(0, 4, fmt.Sprintf("%s: %s",
		pgs.getLocalizedText("Generated on", options.Language),
		time.Now().Format("2006-01-02 15:04:05")))
}

// Helper functions for professional template
func (pgs *PDFGeneratorService) addCompanyInfoProfessional(pdf *gofpdf.Fpdf, company *CompanyInfo) {
	pdf.SetFont("Arial", "B", 11)
	pdf.Cell(0, 6, company.Name)
	pdf.Ln(6)

	pdf.SetFont("Arial", "", 9)
	if company.Address != "" {
		pdf.Cell(0, 4, company.Address)
		pdf.Ln(4)
	}
	if company.Phone != "" {
		pdf.Cell(0, 4, company.Phone)
		pdf.Ln(4)
	}
	if company.Email != "" {
		pdf.Cell(0, 4, company.Email)
		pdf.Ln(4)
	}
}

func (pgs *PDFGeneratorService) addInvoiceDetailsProfessional(pdf *gofpdf.Fpdf, invoice *entities.Invoice, language string) {
	pdf.SetFont("Arial", "B", 11)
	pdf.Cell(0, 6, fmt.Sprintf("%s: %s", pgs.getLocalizedText("Invoice", language), invoice.InvoiceNumber))
	pdf.Ln(6)

	// Add order reference if available
	if invoice.SubscriptionOrderID > 0 {
		pdf.SetFont("Arial", "", 9)
		pdf.Cell(0, 4, fmt.Sprintf("%s: #%d", pgs.getLocalizedText("Order Reference", language), invoice.SubscriptionOrderID))
		pdf.Ln(4)
	}

	pdf.SetFont("Arial", "", 9)
	pdf.Cell(0, 4, fmt.Sprintf("%s: %s", pgs.getLocalizedText("Date", language), invoice.IssuedAt.Format("2006-01-02")))
	pdf.Ln(4)

	if invoice.DueAt != nil {
		pdf.Cell(0, 4, fmt.Sprintf("%s: %s", pgs.getLocalizedText("Due Date", language), invoice.DueAt.Format("2006-01-02")))
		pdf.Ln(4)
	}

	statusColor := pgs.getStatusColor(invoice.Status)
	pdf.SetTextColor(statusColor.R, statusColor.G, statusColor.B)
	pdf.Cell(0, 4, fmt.Sprintf("%s: %s", pgs.getLocalizedText("Status", language), strings.ToUpper(invoice.Status)))
	pdf.SetTextColor(60, 60, 60) // Reset color
}

func (pgs *PDFGeneratorService) addProfessionalTotals(pdf *gofpdf.Fpdf, invoice *entities.Invoice, language string) {
	// Totals box
	startY := pdf.GetY()
	pdf.SetFillColor(248, 248, 248)
	pdf.Rect(120, startY, 75, 25, "F")

	pdf.SetXY(125, startY+5)
	pdf.SetFont("Arial", "", 9)

	if invoice.TaxAmount > 0 {
		pdf.Cell(30, 4, pgs.getLocalizedText("Subtotal", language)+":")
		pdf.Cell(35, 4, fmt.Sprintf("%.2f %s", invoice.Amount, invoice.Currency))
		pdf.Ln(4)

		pdf.SetX(125)
		pdf.Cell(30, 4, fmt.Sprintf("%s (%.1f%%):", pgs.getLocalizedText("Tax", language), invoice.TaxRate*100))
		pdf.Cell(35, 4, fmt.Sprintf("%.2f %s", invoice.TaxAmount, invoice.Currency))
		pdf.Ln(4)

		pdf.SetX(125)
	}

	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(30, 6, pgs.getLocalizedText("Total", language)+":")
	pdf.Cell(35, 6, fmt.Sprintf("%.2f %s", invoice.TotalAmount, invoice.Currency))

	pdf.SetY(startY + 30)
}

func (pgs *PDFGeneratorService) addProfessionalFooter(pdf *gofpdf.Fpdf, invoice *entities.Invoice, options *PDFGenerationOptions) {
	pdf.Ln(15)

	// Payment info
	if options.CompanyInfo != nil && options.CompanyInfo.BankAccount != "" {
		pdf.SetFont("Arial", "B", 10)
		pdf.Cell(0, 6, pgs.getLocalizedText("Payment Information", options.Language))
		pdf.Ln(6)

		pdf.SetFont("Arial", "", 9)
		if options.CompanyInfo.BankAccount != "" {
			pdf.Cell(0, 4, pgs.getLocalizedText("Account", options.Language)+": "+options.CompanyInfo.BankAccount)
			pdf.Ln(4)
		}
		if options.CompanyInfo.RoutingNumber != "" {
			pdf.Cell(0, 4, pgs.getLocalizedText("Routing", options.Language)+": "+options.CompanyInfo.RoutingNumber)
			pdf.Ln(4)
		}
		pdf.Ln(5)
	}

	// Notes
	if invoice.Notes != "" {
		pdf.SetFont("Arial", "B", 10)
		pdf.Cell(0, 6, pgs.getLocalizedText("Notes", options.Language))
		pdf.Ln(6)

		pdf.SetFont("Arial", "", 9)
		pdf.Cell(0, 4, invoice.Notes)
		pdf.Ln(10)
	}

	// Footer line
	pdf.SetDrawColor(200, 200, 200)
	pdf.Line(15, pdf.GetY(), 195, pdf.GetY())
	pdf.Ln(5)

	pdf.SetFont("Arial", "", 8)
	pdf.SetTextColor(120, 120, 120)
	pdf.Cell(0, 4, pgs.getLocalizedText("Thank you for your business!", options.Language))
}

// getStatusColor returns appropriate color for invoice status
func (pgs *PDFGeneratorService) getStatusColor(status string) struct{ R, G, B int } {
	switch status {
	case constants.InvoiceStatusPaid:
		return struct{ R, G, B int }{34, 139, 34} // Forest green
	case constants.InvoiceStatusOverdue:
		return struct{ R, G, B int }{220, 20, 60} // Crimson
	case constants.InvoiceStatusVoided:
		return struct{ R, G, B int }{128, 128, 128} // Gray
	case constants.InvoiceStatusSent:
		return struct{ R, G, B int }{70, 130, 180} // Steel blue
	default:
		return struct{ R, G, B int }{255, 140, 0} // Dark orange
	}
}

// getLocalizedText returns localized text for the given key
func (pgs *PDFGeneratorService) getLocalizedText(key, language string) string {
	translations := map[string]map[string]string{
		"en": {
			"INVOICE":                      "INVOICE",
			"Invoice Number":               "Invoice Number",
			"Invoice":                      "Invoice",
			"Order Reference":              "Order Reference",
			"Date":                         "Date",
			"Due Date":                     "Due Date",
			"Status":                       "Status",
			"Bill To":                      "Bill To",
			"Description":                  "Description",
			"Quantity":                     "Quantity",
			"Qty":                          "Qty",
			"Unit Price":                   "Unit Price",
			"Amount":                       "Amount",
			"Subtotal":                     "Subtotal",
			"Tax":                          "Tax",
			"Total":                        "Total",
			"Phone":                        "Phone",
			"Email":                        "Email",
			"Notes":                        "Notes",
			"Thank you for your business!": "Thank you for your business!",
			"Generated on":                 "Generated on",
			"Payment Information":          "Payment Information",
			"Account":                      "Account",
			"Routing":                      "Routing",
		},
		"zh": {
			"INVOICE":                      "发票",
			"Invoice Number":               "发票号码",
			"Invoice":                      "发票",
			"Order Reference":              "订单编号",
			"Date":                         "日期",
			"Due Date":                     "到期日期",
			"Status":                       "状态",
			"Bill To":                      "收款方",
			"Description":                  "描述",
			"Quantity":                     "数量",
			"Qty":                          "数量",
			"Unit Price":                   "单价",
			"Amount":                       "金额",
			"Subtotal":                     "小计",
			"Tax":                          "税费",
			"Total":                        "总计",
			"Phone":                        "电话",
			"Email":                        "邮箱",
			"Notes":                        "备注",
			"Thank you for your business!": "感谢您的惠顾！",
			"Generated on":                 "生成于",
			"Payment Information":          "付款信息",
			"Account":                      "账户",
			"Routing":                      "路由",
		},
		"es": {
			"INVOICE":                      "FACTURA",
			"Invoice Number":               "Número de Factura",
			"Invoice":                      "Factura",
			"Date":                         "Fecha",
			"Due Date":                     "Fecha de Vencimiento",
			"Status":                       "Estado",
			"Bill To":                      "Facturar a",
			"Description":                  "Descripción",
			"Quantity":                     "Cantidad",
			"Qty":                          "Cant",
			"Unit Price":                   "Precio Unitario",
			"Amount":                       "Importe",
			"Subtotal":                     "Subtotal",
			"Tax":                          "Impuesto",
			"Total":                        "Total",
			"Phone":                        "Teléfono",
			"Email":                        "Correo",
			"Notes":                        "Notas",
			"Thank you for your business!": "¡Gracias por su negocio!",
			"Generated on":                 "Generado el",
			"Payment Information":          "Información de Pago",
			"Account":                      "Cuenta",
			"Routing":                      "Enrutamiento",
		},
	}

	if langMap, exists := translations[language]; exists {
		if text, exists := langMap[key]; exists {
			return text
		}
	}

	// Fallback to English
	if engMap, exists := translations["en"]; exists {
		if text, exists := engMap[key]; exists {
			return text
		}
	}

	// Ultimate fallback
	return key
}

// GenerateBulkPDFs generates PDFs for multiple invoices
func (pgs *PDFGeneratorService) GenerateBulkPDFs(ctx context.Context, invoices []*entities.Invoice, options *PDFGenerationOptions) ([][]byte, []string, error) {
	var pdfBytes [][]byte
	var filePaths []string
	var errors []error

	for _, invoice := range invoices {
		pdf, path, err := pgs.GeneratePDF(ctx, invoice, options)
		if err != nil {
			errors = append(errors, fmt.Errorf("failed to generate PDF for invoice %s: %w", invoice.InvoiceNumber, err))
			continue
		}

		pdfBytes = append(pdfBytes, pdf)
		filePaths = append(filePaths, path)
	}

	if len(errors) > 0 {
		pgs.logger.Error("Some PDFs failed to generate",
			zap.Int("total_invoices", len(invoices)),
			zap.Int("failed_count", len(errors)))
	}

	return pdfBytes, filePaths, nil
}

// ValidateTemplate checks if the template is valid
func (pgs *PDFGeneratorService) ValidateTemplate(template string) bool {
	validTemplates := []string{"default", "professional", "minimal"}
	for _, valid := range validTemplates {
		if template == valid {
			return true
		}
	}
	return false
}

// GetAvailableTemplates returns list of available templates
func (pgs *PDFGeneratorService) GetAvailableTemplates() []string {
	return []string{"default", "professional", "minimal"}
}

// GetAvailableLanguages returns list of supported languages
func (pgs *PDFGeneratorService) GetAvailableLanguages() []string {
	return []string{"en", "zh", "es"}
}
