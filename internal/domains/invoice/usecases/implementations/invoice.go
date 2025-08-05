package implementations

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"time"

	"linke/internal/domains/invoice/entities"
	"linke/internal/domains/invoice/usecases/interfaces"
	userInterfaces "linke/internal/domains/user/usecases/interfaces"
	"linke/internal/shared/logger"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type InvoiceService struct {
	db           *gorm.DB
	userService  userInterfaces.UserService
	pdfGenerator *PDFGeneratorService
	logger       logger.Logger
}

func NewInvoiceService(db *gorm.DB, userService userInterfaces.UserService, pdfGenerator *PDFGeneratorService, logger logger.Logger) *InvoiceService {
	return &InvoiceService{
		db:           db,
		userService:  userService,
		pdfGenerator: pdfGenerator,
		logger:       logger,
	}
}

// CreateInvoice creates a new invoice
func (is *InvoiceService) CreateInvoice(ctx context.Context, req *interfaces.CreateInvoiceRequest) (*entities.Invoice, error) {
	// TODO: Replace with subscription service interface call
	// For now, we'll validate the order ID exists in the subscription context
	if req.SubscriptionOrderID == 0 {
		return nil, fmt.Errorf("subscription order ID is required")
	}

	// TODO: Validate user exists through user service
	// This should be handled through proper subscription service interface

	// Set defaults
	invoiceType := req.InvoiceType
	if invoiceType == "" {
		invoiceType = entities.InvoiceTypeStandard
	}

	currency := req.Currency
	if currency == "" {
		currency = "USD" // Default currency
	}

	template := req.Template
	if template == "" {
		template = "default"
	}

	language := req.Language
	if language == "" {
		language = "en"
	}

	// Calculate tax amount
	taxAmount := req.Amount * req.TaxRate
	totalAmount := req.Amount + taxAmount

	// Parse due date
	var dueAt *time.Time
	if req.DueDate != "" {
		if parsed, err := time.Parse("2006-01-02", req.DueDate); err == nil {
			// Set due time to end of day
			due := time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 23, 59, 59, 0, parsed.Location())
			dueAt = &due
		}
	}

	// Generate invoice number
	invoiceNumber := is.generateInvoiceNumber()

	// Create invoice
	invoice := &entities.Invoice{
		UserID:              req.UserID,
		SubscriptionOrderID: req.SubscriptionOrderID,
		InvoiceNumber:       invoiceNumber,
		InvoiceType:         invoiceType,
		Status:              entities.InvoiceStatusDraft,
		Amount:              req.Amount,
		Currency:            currency,
		TaxAmount:           taxAmount,
		TotalAmount:         totalAmount,
		TaxRate:             req.TaxRate,
		TaxType:             req.TaxType,
		TaxNumber:           req.TaxNumber,
		BillingName:         req.BillingName,
		BillingEmail:        req.BillingEmail,
		BillingAddress:      req.BillingAddress,
		BillingCity:         req.BillingCity,
		BillingState:        req.BillingState,
		BillingCountry:      req.BillingCountry,
		BillingZip:          req.BillingZip,
		CompanyName:         req.CompanyName,
		CompanyTaxID:        req.CompanyTaxID,
		CompanyAddress:      req.CompanyAddress,
		IssuedAt:            time.Now(),
		DueAt:               dueAt,
		Template:            template,
		Language:            language,
		Description:         req.Description,
		Notes:               req.Notes,
	}

	// Save invoice
	if err := is.db.WithContext(ctx).Create(invoice).Error; err != nil {
		logger.Error("Failed to create invoice", logger.Error2("error", err))
		return nil, fmt.Errorf("failed to create invoice: %w", err)
	}

	// Auto-send if requested
	if req.AutoSend {
		emailRequest := &interfaces.SendInvoiceRequest{
			ToEmail: req.BillingEmail,
			Subject: fmt.Sprintf("Invoice %s", invoiceNumber),
			Message: "Please find your invoice attached.",
		}
		if err := is.SendInvoice(ctx, invoice.ID, emailRequest); err != nil {
			logger.Error("Failed to auto-send invoice",
				logger.Error2("error", err),
				logger.Uint("invoice_id", invoice.ID))
			// Don't fail the creation, just log the error
		}
	}

	logger.Info("Invoice created successfully",
		logger.Uint("invoice_id", invoice.ID),
		logger.String("invoice_number", invoiceNumber),
		logger.Uint("user_id", req.UserID),
		logger.Uint("order_id", req.SubscriptionOrderID))

	return invoice, nil
}

// CreateInvoiceFromOrder creates an invoice from a subscription order
func (is *InvoiceService) CreateInvoiceFromOrder(ctx context.Context, orderID uint, options *interfaces.CreateInvoiceRequest) (*entities.Invoice, error) {
	// TODO: Replace with subscription service interface call
	// For now, we'll create the invoice with the provided order ID
	if orderID == 0 {
		return nil, fmt.Errorf("subscription order ID is required")
	}

	// TODO: Validate order is paid through subscription service
	// For now, we'll proceed with invoice creation

	// Check if invoice already exists for this order
	var existingInvoice entities.Invoice
	if err := is.db.WithContext(ctx).Where("subscription_order_id = ?", orderID).First(&existingInvoice).Error; err == nil {
		return nil, fmt.Errorf("invoice already exists for this order")
	}

	// Build request from provided options
	if options == nil {
		return nil, fmt.Errorf("options are required for invoice creation")
	}

	req := &interfaces.CreateInvoiceRequest{
		UserID:              options.UserID,
		SubscriptionOrderID: orderID,
		Amount:              options.Amount,
		Currency:            options.Currency,
		Description:         fmt.Sprintf("Subscription order %d", orderID),
		InvoiceType:         options.InvoiceType,
		BillingName:         options.BillingName,
		BillingEmail:        options.BillingEmail,
		BillingAddress:      options.BillingAddress,
		TaxNumber:           options.TaxNumber,
		Notes:               options.Notes,
	}

	// Override with provided options
	if options != nil {
		if options.InvoiceType != "" {
			req.InvoiceType = options.InvoiceType
		}
		if options.TaxRate > 0 {
			req.TaxRate = options.TaxRate
		}
		if options.TaxType != "" {
			req.TaxType = options.TaxType
		}
		if options.TaxNumber != "" {
			req.TaxNumber = options.TaxNumber
		}
		if options.BillingName != "" {
			req.BillingName = options.BillingName
		}
		if options.BillingEmail != "" {
			req.BillingEmail = options.BillingEmail
		}
		if options.BillingAddress != "" {
			req.BillingAddress = options.BillingAddress
		}
		if options.BillingCity != "" {
			req.BillingCity = options.BillingCity
		}
		if options.BillingState != "" {
			req.BillingState = options.BillingState
		}
		if options.BillingCountry != "" {
			req.BillingCountry = options.BillingCountry
		}
		if options.BillingZip != "" {
			req.BillingZip = options.BillingZip
		}
		if options.CompanyName != "" {
			req.CompanyName = options.CompanyName
		}
		if options.CompanyTaxID != "" {
			req.CompanyTaxID = options.CompanyTaxID
		}
		if options.CompanyAddress != "" {
			req.CompanyAddress = options.CompanyAddress
		}
		if options.Description != "" {
			req.Description = options.Description
		}
		if options.Notes != "" {
			req.Notes = options.Notes
		}
		if options.DueDate != "" {
			req.DueDate = options.DueDate
		}
		if options.Template != "" {
			req.Template = options.Template
		}
		if options.Language != "" {
			req.Language = options.Language
		}
		req.AutoSend = options.AutoSend
	}

	return is.CreateInvoice(ctx, req)
}

// GetInvoice gets an invoice by ID
func (is *InvoiceService) GetInvoice(ctx context.Context, invoiceID uint) (*entities.Invoice, error) {
	var invoice entities.Invoice
	if err := is.db.WithContext(ctx).First(&invoice, invoiceID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("invoice not found")
		}
		return nil, fmt.Errorf("failed to get invoice: %w", err)
	}
	return &invoice, nil
}

// GetInvoiceWithRelations gets an invoice with related data
func (is *InvoiceService) GetInvoiceWithRelations(ctx context.Context, invoiceID uint) (*entities.Invoice, error) {
	var invoice entities.Invoice
	if err := is.db.WithContext(ctx).
		Preload("User").
		Preload("SubscriptionOrder").
		First(&invoice, invoiceID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("invoice not found")
		}
		return nil, fmt.Errorf("failed to get invoice: %w", err)
	}
	return &invoice, nil
}

// GetInvoices gets invoices with filtering
func (is *InvoiceService) GetInvoices(ctx context.Context, req *interfaces.GetInvoicesRequest) ([]*entities.Invoice, int64, error) {
	query := is.db.WithContext(ctx).Model(&entities.Invoice{})

	// Apply filters
	if req.UserID != 0 {
		query = query.Where("user_id = ?", req.UserID)
	}

	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}

	if req.InvoiceType != "" {
		query = query.Where("invoice_type = ?", req.InvoiceType)
	}

	// Date range filtering
	if req.DateFrom != "" {
		if startDate, err := time.Parse("2006-01-02", req.DateFrom); err == nil {
			query = query.Where("issued_at >= ?", startDate)
		}
	}

	if req.DateTo != "" {
		if endDate, err := time.Parse("2006-01-02", req.DateTo); err == nil {
			endDate = endDate.Add(24 * time.Hour)
			query = query.Where("issued_at < ?", endDate)
		}
	}

	// Get total count
	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count invoices: %w", err)
	}

	// Apply sorting
	query = query.Order("created_at desc")

	// Apply pagination
	if req.Limit > 0 {
		query = query.Limit(req.Limit)
	}

	if req.Offset > 0 {
		query = query.Offset(req.Offset)
	}

	var invoices []*entities.Invoice
	if err := query.Find(&invoices).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get invoices: %w", err)
	}

	return invoices, totalCount, nil
}

// UpdateInvoice updates an invoice
func (is *InvoiceService) UpdateInvoice(ctx context.Context, invoiceID uint, req *interfaces.UpdateInvoiceRequest) (*entities.Invoice, error) {
	// Get invoice
	invoice, err := is.GetInvoice(ctx, invoiceID)
	if err != nil {
		return nil, err
	}

	// Check if invoice can be edited
	if !invoice.CanBeEdited() {
		return nil, fmt.Errorf("invoice cannot be edited in status: %s", invoice.Status)
	}

	// Prepare update data
	updateData := make(map[string]any)

	if req.InvoiceType != nil {
		updateData["invoice_type"] = *req.InvoiceType
	}
	if req.TaxType != nil {
		updateData["tax_type"] = *req.TaxType
	}
	if req.BillingName != nil {
		updateData["billing_name"] = *req.BillingName
	}
	if req.BillingEmail != nil {
		updateData["billing_email"] = *req.BillingEmail
	}
	if req.BillingAddress != nil {
		updateData["billing_address"] = *req.BillingAddress
	}
	if req.BillingCity != nil {
		updateData["billing_city"] = *req.BillingCity
	}
	if req.BillingState != nil {
		updateData["billing_state"] = *req.BillingState
	}
	if req.BillingCountry != nil {
		updateData["billing_country"] = *req.BillingCountry
	}
	if req.BillingZip != nil {
		updateData["billing_zip"] = *req.BillingZip
	}
	if req.CompanyName != nil {
		updateData["company_name"] = *req.CompanyName
	}
	if req.CompanyTaxID != nil {
		updateData["company_tax_id"] = *req.CompanyTaxID
	}
	if req.CompanyAddress != nil {
		updateData["company_address"] = *req.CompanyAddress
	}
	if req.Description != nil {
		updateData["description"] = *req.Description
	}
	if req.Notes != nil {
		updateData["notes"] = *req.Notes
	}
	if req.Template != nil {
		updateData["template"] = *req.Template
	}
	if req.Language != nil {
		updateData["language"] = *req.Language
	}

	// Handle tax rate changes (recalculate amounts)
	if req.TaxRate != nil {
		updateData["tax_rate"] = *req.TaxRate
		newTaxAmount := invoice.Amount * (*req.TaxRate)
		updateData["tax_amount"] = newTaxAmount
		updateData["total_amount"] = invoice.Amount + newTaxAmount
	}

	// Handle due date
	if req.DueDate != nil {
		if *req.DueDate == "" {
			updateData["due_at"] = nil
		} else {
			if parsed, err := time.Parse("2006-01-02", *req.DueDate); err == nil {
				due := time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 23, 59, 59, 0, parsed.Location())
				updateData["due_at"] = due
			}
		}
	}

	updateData["updated_at"] = time.Now()

	// Update invoice
	if err := is.db.WithContext(ctx).Model(invoice).Updates(updateData).Error; err != nil {
		return nil, fmt.Errorf("failed to update invoice: %w", err)
	}

	// Reload invoice
	if err := is.db.WithContext(ctx).First(invoice, invoiceID).Error; err != nil {
		return nil, fmt.Errorf("failed to reload invoice: %w", err)
	}

	return invoice, nil
}

// SendInvoice sends an invoice to the customer
func (is *InvoiceService) SendInvoice(ctx context.Context, invoiceID uint, emailRequest *interfaces.SendInvoiceRequest) error {
	// Get invoice
	invoice, err := is.GetInvoice(ctx, invoiceID)
	if err != nil {
		return err
	}

	// Check if invoice can be sent
	if !invoice.CanBeSent() {
		return fmt.Errorf("invoice cannot be sent in status: %s", invoice.Status)
	}

	// Update invoice status
	now := time.Now()
	updateData := map[string]any{
		"status":     entities.InvoiceStatusSent,
		"sent_at":    now,
		"updated_at": now,
	}

	if err := is.db.WithContext(ctx).Model(invoice).Updates(updateData).Error; err != nil {
		return fmt.Errorf("failed to update invoice status: %w", err)
	}

	// TODO: Implement actual email sending
	// This would integrate with an email service to send the invoice PDF

	logger.Info("Invoice sent successfully",
		logger.Uint("invoice_id", invoiceID),
		logger.String("invoice_number", invoice.InvoiceNumber),
		logger.String("billing_email", invoice.BillingEmail))

	return nil
}

// MarkInvoiceAsPaid marks an invoice as paid
func (is *InvoiceService) MarkInvoiceAsPaid(ctx context.Context, invoiceID uint, paymentDate string) error {
	// Get invoice
	invoice, err := is.GetInvoice(ctx, invoiceID)
	if err != nil {
		return err
	}

	// Check if invoice can be paid
	if !invoice.CanBePaid() {
		return fmt.Errorf("invoice cannot be marked as paid in status: %s", invoice.Status)
	}

	// Parse payment date
	var paidAt time.Time
	if paymentDate != "" {
		if parsed, parseErr := time.Parse("2006-01-02", paymentDate); parseErr == nil {
			paidAt = parsed
		} else {
			return fmt.Errorf("invalid payment date format: %w", parseErr)
		}
	} else {
		paidAt = time.Now()
	}

	// Update invoice status
	updateData := map[string]any{
		"status":     entities.InvoiceStatusPaid,
		"paid_at":    paidAt,
		"updated_at": time.Now(),
	}

	if err := is.db.WithContext(ctx).Model(invoice).Updates(updateData).Error; err != nil {
		return fmt.Errorf("failed to update invoice status: %w", err)
	}

	logger.Info("Invoice marked as paid",
		logger.Uint("invoice_id", invoiceID),
		logger.String("invoice_number", invoice.InvoiceNumber),
		logger.String("payment_date", paymentDate))

	return nil
}

// MarkInvoiceAsVoid voids an invoice
func (is *InvoiceService) MarkInvoiceAsVoid(ctx context.Context, invoiceID uint, reason string) error {
	// Get invoice
	invoice, err := is.GetInvoice(ctx, invoiceID)
	if err != nil {
		return err
	}

	// Check if invoice can be voided
	if !invoice.CanBeVoided() {
		return fmt.Errorf("invoice cannot be voided in status: %s", invoice.Status)
	}

	// Update invoice status
	now := time.Now()
	updateData := map[string]any{
		"status":     entities.InvoiceStatusVoided,
		"voided_at":  now,
		"updated_at": now,
	}

	// Add void reason to notes
	if reason != "" {
		voidNote := fmt.Sprintf("[%s] Invoice voided: %s", now.Format("2006-01-02 15:04:05"), reason)
		if invoice.Notes != "" {
			updateData["notes"] = invoice.Notes + "\n" + voidNote
		} else {
			updateData["notes"] = voidNote
		}
	}

	if err := is.db.WithContext(ctx).Model(invoice).Updates(updateData).Error; err != nil {
		return fmt.Errorf("failed to void invoice: %w", err)
	}

	logger.Info("Invoice voided",
		logger.Uint("invoice_id", invoiceID),
		logger.String("invoice_number", invoice.InvoiceNumber),
		logger.String("reason", reason))

	return nil
}

// DeleteInvoice soft deletes an invoice
func (is *InvoiceService) DeleteInvoice(ctx context.Context, invoiceID uint) error {
	// Get invoice
	invoice, err := is.GetInvoice(ctx, invoiceID)
	if err != nil {
		return err
	}

	// Only allow deletion of draft invoices
	if !invoice.IsDraft() {
		return fmt.Errorf("only draft invoices can be deleted")
	}

	// Soft delete invoice
	if err := is.db.WithContext(ctx).Delete(invoice).Error; err != nil {
		return fmt.Errorf("failed to delete invoice: %w", err)
	}

	logger.Info("Invoice deleted",
		logger.Uint("invoice_id", invoiceID),
		logger.String("invoice_number", invoice.InvoiceNumber))

	return nil
}

// GetInvoiceByNumber gets an invoice by invoice number
func (is *InvoiceService) GetInvoiceByNumber(ctx context.Context, invoiceNumber string) (*entities.Invoice, error) {
	var invoice entities.Invoice
	if err := is.db.WithContext(ctx).Where("invoice_number = ?", invoiceNumber).First(&invoice).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("invoice not found")
		}
		return nil, fmt.Errorf("failed to get invoice: %w", err)
	}
	return &invoice, nil
}

// GetUserInvoices gets invoices for a specific user
func (is *InvoiceService) GetUserInvoices(ctx context.Context, userID uint, limit, offset int) ([]*entities.Invoice, int64, error) {
	query := is.db.WithContext(ctx).Model(&entities.Invoice{}).Where("user_id = ?", userID)

	// Get total count
	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count user invoices: %w", err)
	}

	// Apply pagination
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	query = query.Order("created_at desc")

	var invoices []*entities.Invoice
	if err := query.Find(&invoices).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get user invoices: %w", err)
	}

	return invoices, totalCount, nil
}

// GenerateInvoicePDF generates a PDF for the invoice with default options
func (is *InvoiceService) GenerateInvoicePDF(ctx context.Context, invoiceID uint) ([]byte, error) {
	invoice, err := is.GetInvoice(ctx, invoiceID)
	if err != nil {
		return nil, err
	}

	options := &PDFGenerationOptions{
		Template:   invoice.Template,
		Language:   invoice.Language,
		SaveToDisk: false,
	}

	if invoice.Template == "" {
		options.Template = "default"
	}
	if invoice.Language == "" {
		options.Language = "en"
	}

	pdfBytes, _, err := is.pdfGenerator.GeneratePDF(ctx, invoice, options)
	if err != nil {
		return nil, fmt.Errorf("failed to generate PDF for invoice %d: %w", invoiceID, err)
	}

	return pdfBytes, nil
}

// GenerateInvoicePDFWithOptions generates a PDF with custom options
func (is *InvoiceService) GenerateInvoicePDFWithOptions(ctx context.Context, invoiceID uint, options *interfaces.PDFGenerationRequest) ([]byte, string, error) {
	invoice, err := is.GetInvoice(ctx, invoiceID)
	if err != nil {
		return nil, "", err
	}

	// Convert interface options to implementation options
	pdfOptions := &PDFGenerationOptions{
		Template:     options.Template,
		Language:     options.Language,
		Watermark:    options.Watermark,
		SaveToDisk:   options.SaveToDisk,
		IncludeQR:    options.IncludeQR,
		CustomFields: options.CustomFields,
	}

	// Convert company info if provided
	if options.CompanyInfo != nil {
		pdfOptions.CompanyInfo = &CompanyInfo{
			Name:          options.CompanyInfo.Name,
			Address:       options.CompanyInfo.Address,
			City:          options.CompanyInfo.City,
			State:         options.CompanyInfo.State,
			ZIP:           options.CompanyInfo.ZIP,
			Country:       options.CompanyInfo.Country,
			Phone:         options.CompanyInfo.Phone,
			Email:         options.CompanyInfo.Email,
			Website:       options.CompanyInfo.Website,
			TaxID:         options.CompanyInfo.TaxID,
			BankAccount:   options.CompanyInfo.BankAccount,
			RoutingNumber: options.CompanyInfo.RoutingNumber,
			Logo:          options.CompanyInfo.Logo,
		}
	}

	// Set defaults
	if pdfOptions.Template == "" {
		pdfOptions.Template = invoice.Template
		if pdfOptions.Template == "" {
			pdfOptions.Template = "default"
		}
	}
	if pdfOptions.Language == "" {
		pdfOptions.Language = invoice.Language
		if pdfOptions.Language == "" {
			pdfOptions.Language = "en"
		}
	}

	pdfBytes, filePath, err := is.pdfGenerator.GeneratePDF(ctx, invoice, pdfOptions)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate PDF for invoice %d: %w", invoiceID, err)
	}

	// Update invoice with PDF info if saved to disk
	if pdfOptions.SaveToDisk && filePath != "" {
		updateData := map[string]any{
			"pdf_path":   filePath,
			"pdf_size":   len(pdfBytes),
			"updated_at": time.Now(),
		}
		if err := is.db.WithContext(ctx).Model(invoice).Updates(updateData).Error; err != nil {
			is.logger.Error("Failed to update invoice PDF info",
				zap.Error(err),
				zap.Uint("invoice_id", invoiceID))
		}
	}

	return pdfBytes, filePath, nil
}

// GenerateBulkInvoicePDFs generates PDFs for multiple invoices and returns them as a ZIP
func (is *InvoiceService) GenerateBulkInvoicePDFs(ctx context.Context, invoiceIDs []uint, options *interfaces.PDFGenerationRequest) ([]byte, error) {
	if len(invoiceIDs) == 0 {
		return nil, fmt.Errorf("no invoice IDs provided")
	}

	// Get all invoices
	var invoices []*entities.Invoice
	if err := is.db.WithContext(ctx).Where("id IN ?", invoiceIDs).Find(&invoices).Error; err != nil {
		return nil, fmt.Errorf("failed to get invoices: %w", err)
	}

	if len(invoices) == 0 {
		return nil, fmt.Errorf("no invoices found")
	}

	// Convert options
	pdfOptions := &PDFGenerationOptions{
		SaveToDisk: false, // Never save to disk for bulk
	}
	if options != nil {
		pdfOptions.Template = options.Template
		pdfOptions.Language = options.Language
		pdfOptions.Watermark = options.Watermark
		pdfOptions.IncludeQR = options.IncludeQR
		pdfOptions.CustomFields = options.CustomFields

		if options.CompanyInfo != nil {
			pdfOptions.CompanyInfo = &CompanyInfo{
				Name:          options.CompanyInfo.Name,
				Address:       options.CompanyInfo.Address,
				City:          options.CompanyInfo.City,
				State:         options.CompanyInfo.State,
				ZIP:           options.CompanyInfo.ZIP,
				Country:       options.CompanyInfo.Country,
				Phone:         options.CompanyInfo.Phone,
				Email:         options.CompanyInfo.Email,
				Website:       options.CompanyInfo.Website,
				TaxID:         options.CompanyInfo.TaxID,
				BankAccount:   options.CompanyInfo.BankAccount,
				RoutingNumber: options.CompanyInfo.RoutingNumber,
				Logo:          options.CompanyInfo.Logo,
			}
		}
	}

	// Create ZIP file
	var zipBuffer bytes.Buffer
	zipWriter := zip.NewWriter(&zipBuffer)

	// Generate PDFs and add to ZIP
	for _, invoice := range invoices {
		// Set invoice-specific defaults if not provided
		invoiceOptions := *pdfOptions
		if invoiceOptions.Template == "" {
			invoiceOptions.Template = invoice.Template
			if invoiceOptions.Template == "" {
				invoiceOptions.Template = "default"
			}
		}
		if invoiceOptions.Language == "" {
			invoiceOptions.Language = invoice.Language
			if invoiceOptions.Language == "" {
				invoiceOptions.Language = "en"
			}
		}

		pdfBytes, _, err := is.pdfGenerator.GeneratePDF(ctx, invoice, &invoiceOptions)
		if err != nil {
			is.logger.Error("Failed to generate PDF for invoice in bulk operation",
				zap.Error(err),
				zap.Uint("invoice_id", invoice.ID),
				zap.String("invoice_number", invoice.InvoiceNumber))
			continue
		}

		// Add PDF to ZIP
		fileName := fmt.Sprintf("invoice_%s.pdf", invoice.InvoiceNumber)
		fileWriter, err := zipWriter.Create(fileName)
		if err != nil {
			is.logger.Error("Failed to create file in ZIP",
				zap.Error(err),
				zap.String("filename", fileName))
			continue
		}

		if _, err := fileWriter.Write(pdfBytes); err != nil {
			is.logger.Error("Failed to write PDF to ZIP",
				zap.Error(err),
				zap.String("filename", fileName))
			continue
		}
	}

	// Add CSV summary (always include for bulk downloads)
	if err := is.addInvoiceCSVToZip(zipWriter, invoices); err != nil {
		is.logger.Error("Failed to add CSV to ZIP", zap.Error(err))
	}

	if err := zipWriter.Close(); err != nil {
		return nil, fmt.Errorf("failed to close ZIP writer: %w", err)
	}

	is.logger.Info("Bulk PDF generation completed",
		zap.Int("total_invoices", len(invoices)),
		zap.Int("requested_invoices", len(invoiceIDs)))

	return zipBuffer.Bytes(), nil
}

// addInvoiceCSVToZip adds a CSV summary of invoices to the ZIP file
func (is *InvoiceService) addInvoiceCSVToZip(zipWriter *zip.Writer, invoices []*entities.Invoice) error {
	csvWriter, err := zipWriter.Create("invoice_summary.csv")
	if err != nil {
		return err
	}

	writer := csv.NewWriter(csvWriter)
	defer writer.Flush()

	// Write header
	header := []string{
		"Invoice Number", "Date", "Due Date", "Amount", "Currency", "Tax Amount",
		"Total Amount", "Status", "Billing Name", "Billing Email", "Description",
	}
	if err := writer.Write(header); err != nil {
		return err
	}

	// Write data
	for _, invoice := range invoices {
		dueDate := ""
		if invoice.DueAt != nil {
			dueDate = invoice.DueAt.Format("2006-01-02")
		}

		record := []string{
			invoice.InvoiceNumber,
			invoice.IssuedAt.Format("2006-01-02"),
			dueDate,
			fmt.Sprintf("%.2f", invoice.Amount),
			invoice.Currency,
			fmt.Sprintf("%.2f", invoice.TaxAmount),
			fmt.Sprintf("%.2f", invoice.TotalAmount),
			invoice.Status,
			invoice.BillingName,
			invoice.BillingEmail,
			invoice.Description,
		}
		if err := writer.Write(record); err != nil {
			return err
		}
	}

	return nil
}

// ResendInvoice resends an invoice
func (is *InvoiceService) ResendInvoice(ctx context.Context, invoiceID uint) error {
	// Get invoice
	invoice, err := is.GetInvoice(ctx, invoiceID)
	if err != nil {
		return err
	}

	// Check if invoice can be sent
	if !invoice.CanBeSent() && invoice.Status != entities.InvoiceStatusSent {
		return fmt.Errorf("invoice cannot be resent in status: %s", invoice.Status)
	}

	// Create a default send request
	emailRequest := &interfaces.SendInvoiceRequest{
		ToEmail: invoice.BillingEmail,
		Subject: fmt.Sprintf("Invoice %s", invoice.InvoiceNumber),
		Message: "Please find your invoice attached.",
	}

	return is.SendInvoice(ctx, invoiceID, emailRequest)
}

// MarkInvoiceAsOverdue marks an invoice as overdue
func (is *InvoiceService) MarkInvoiceAsOverdue(ctx context.Context, invoiceID uint) error {
	// Get invoice
	invoice, err := is.GetInvoice(ctx, invoiceID)
	if err != nil {
		return err
	}

	// Check if invoice is actually overdue
	if invoice.DueAt == nil || invoice.DueAt.After(time.Now()) {
		return fmt.Errorf("invoice is not overdue")
	}

	// Check if invoice can be marked as overdue
	if invoice.Status == entities.InvoiceStatusPaid || invoice.Status == entities.InvoiceStatusVoided {
		return fmt.Errorf("invoice cannot be marked as overdue in status: %s", invoice.Status)
	}

	// Update invoice status
	updateData := map[string]any{
		"status":     entities.InvoiceStatusOverdue,
		"updated_at": time.Now(),
	}

	if err := is.db.WithContext(ctx).Model(invoice).Updates(updateData).Error; err != nil {
		return fmt.Errorf("failed to mark invoice as overdue: %w", err)
	}

	logger.Info("Invoice marked as overdue",
		logger.Uint("invoice_id", invoiceID),
		logger.String("invoice_number", invoice.InvoiceNumber))

	return nil
}

// GetInvoiceStatistics gets invoice statistics for a date range
func (is *InvoiceService) GetInvoiceStatistics(ctx context.Context, fromDate, toDate string) (map[string]any, error) {
	stats := make(map[string]any)

	// Parse dates
	var startDate, endDate time.Time
	var err error

	if fromDate != "" {
		startDate, err = time.Parse("2006-01-02", fromDate)
		if err != nil {
			return nil, fmt.Errorf("invalid from_date format: %w", err)
		}
	} else {
		startDate = time.Now().AddDate(0, -1, 0) // Default to last month
	}

	if toDate != "" {
		endDate, err = time.Parse("2006-01-02", toDate)
		if err != nil {
			return nil, fmt.Errorf("invalid to_date format: %w", err)
		}
		endDate = endDate.Add(24 * time.Hour) // Include the end date
	} else {
		endDate = time.Now()
	}

	query := is.db.WithContext(ctx).Model(&entities.Invoice{}).
		Where("created_at >= ? AND created_at < ?", startDate, endDate)

	// Total count
	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count invoices: %w", err)
	}
	stats["total_count"] = totalCount

	// Count by status
	type StatusCount struct {
		Status string
		Count  int64
	}
	var statusCounts []StatusCount
	if err := query.Select("status, COUNT(*) as count").Group("status").Find(&statusCounts).Error; err != nil {
		return nil, fmt.Errorf("failed to get status counts: %w", err)
	}

	statusStats := make(map[string]int64)
	for _, sc := range statusCounts {
		statusStats[sc.Status] = sc.Count
	}
	stats["by_status"] = statusStats

	// Total amount
	type AmountSum struct {
		TotalAmount float64
	}
	var amountSum AmountSum
	if err := query.Select("SUM(total_amount) as total_amount").Find(&amountSum).Error; err != nil {
		return nil, fmt.Errorf("failed to calculate total amount: %w", err)
	}
	stats["total_amount"] = amountSum.TotalAmount

	stats["from_date"] = fromDate
	stats["to_date"] = toDate

	return stats, nil
}

// GetUserInvoiceStatistics gets invoice statistics for a specific user
func (is *InvoiceService) GetUserInvoiceStatistics(ctx context.Context, userID uint) (map[string]any, error) {
	stats := make(map[string]any)

	query := is.db.WithContext(ctx).Model(&entities.Invoice{}).Where("user_id = ?", userID)

	// Total count
	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count user invoices: %w", err)
	}
	stats["total_count"] = totalCount

	// Count by status
	type StatusCount struct {
		Status string
		Count  int64
	}
	var statusCounts []StatusCount
	if err := query.Select("status, COUNT(*) as count").Group("status").Find(&statusCounts).Error; err != nil {
		return nil, fmt.Errorf("failed to get status counts: %w", err)
	}

	statusStats := make(map[string]int64)
	for _, sc := range statusCounts {
		statusStats[sc.Status] = sc.Count
	}
	stats["by_status"] = statusStats

	// Total amount
	type AmountSum struct {
		TotalAmount float64
	}
	var amountSum AmountSum
	if err := query.Select("SUM(total_amount) as total_amount").Find(&amountSum).Error; err != nil {
		return nil, fmt.Errorf("failed to calculate total amount: %w", err)
	}
	stats["total_amount"] = amountSum.TotalAmount

	// Overdue count
	var overdueCount int64
	if err := query.Where("due_at < ? AND status != ?", time.Now(), entities.InvoiceStatusPaid).Count(&overdueCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count overdue invoices: %w", err)
	}
	stats["overdue_count"] = overdueCount

	stats["user_id"] = userID

	return stats, nil
}

// SendInvoiceWithPDF sends an invoice with custom PDF options
func (is *InvoiceService) SendInvoiceWithPDF(ctx context.Context, invoiceID uint, emailRequest *interfaces.SendInvoiceRequest, pdfOptions *interfaces.PDFGenerationRequest) error {
	// Generate PDF with custom options
	pdfBytes, _, err := is.GenerateInvoicePDFWithOptions(ctx, invoiceID, pdfOptions)
	if err != nil {
		return fmt.Errorf("failed to generate PDF for email: %w", err)
	}

	// TODO: Implement actual email sending with PDF attachment
	// This would integrate with an email service to send the invoice PDF

	// For now, update invoice status as sent
	invoice, err := is.GetInvoice(ctx, invoiceID)
	if err != nil {
		return err
	}

	now := time.Now()
	updateData := map[string]any{
		"status":     entities.InvoiceStatusSent,
		"sent_at":    now,
		"updated_at": now,
	}

	if err := is.db.WithContext(ctx).Model(invoice).Updates(updateData).Error; err != nil {
		return fmt.Errorf("failed to update invoice status: %w", err)
	}

	is.logger.Info("Invoice sent with custom PDF",
		zap.Uint("invoice_id", invoiceID),
		zap.String("invoice_number", invoice.InvoiceNumber),
		zap.String("template", pdfOptions.Template),
		zap.Int("pdf_size", len(pdfBytes)))

	return nil
}

// GetInvoicePDFCached returns a cached PDF or generates one
func (is *InvoiceService) GetInvoicePDFCached(ctx context.Context, invoiceID uint, template string) ([]byte, error) {
	invoice, err := is.GetInvoice(ctx, invoiceID)
	if err != nil {
		return nil, err
	}

	// Check if we have a cached PDF with the correct template
	if invoice.PDFPath != "" && (template == "" || template == invoice.Template) {
		// Try to read from disk cache
		if pdfBytes, err := is.readPDFFromDisk(invoice.PDFPath); err == nil {
			return pdfBytes, nil
		}
	}

	// Generate new PDF
	options := &interfaces.PDFGenerationRequest{
		Template:   template,
		Language:   invoice.Language,
		SaveToDisk: true,
	}

	if template == "" {
		options.Template = invoice.Template
	}

	pdfBytes, _, err := is.GenerateInvoicePDFWithOptions(ctx, invoiceID, options)
	if err != nil {
		return nil, err
	}

	return pdfBytes, nil
}

// DownloadInvoiceAsZip creates a ZIP file with multiple invoices
func (is *InvoiceService) DownloadInvoiceAsZip(ctx context.Context, invoiceIDs []uint) ([]byte, string, error) {
	zipBytes, err := is.GenerateBulkInvoicePDFs(ctx, invoiceIDs, &interfaces.PDFGenerationRequest{
		Template: "professional",
		Language: "en",
	})
	if err != nil {
		return nil, "", err
	}

	filename := fmt.Sprintf("invoices_%d_%d.zip", len(invoiceIDs), time.Now().Unix())
	return zipBytes, filename, nil
}

// GetInvoiceDownloadHistory returns download history for a user
func (is *InvoiceService) GetInvoiceDownloadHistory(ctx context.Context, userID uint) ([]*interfaces.InvoiceDownloadRecord, error) {
	// TODO: Implement download history tracking
	// This would require a separate table to track downloads
	return []*interfaces.InvoiceDownloadRecord{}, nil
}

// GetAvailableTemplates returns available PDF templates
func (is *InvoiceService) GetAvailableTemplates(ctx context.Context) ([]string, error) {
	return is.pdfGenerator.GetAvailableTemplates(), nil
}

// GetAvailableLanguages returns available languages
func (is *InvoiceService) GetAvailableLanguages(ctx context.Context) ([]string, error) {
	return is.pdfGenerator.GetAvailableLanguages(), nil
}

// ValidateTemplate validates if a template exists
func (is *InvoiceService) ValidateTemplate(ctx context.Context, template string) (bool, error) {
	return is.pdfGenerator.ValidateTemplate(template), nil
}

// readPDFFromDisk reads a PDF file from disk
func (is *InvoiceService) readPDFFromDisk(filePath string) ([]byte, error) {
	// This would read the PDF from the file system
	// For now, return an error to force regeneration
	return nil, fmt.Errorf("PDF file not found or not accessible")
}

// generateInvoiceNumber generates a unique invoice number
func (is *InvoiceService) generateInvoiceNumber() string {
	// Generate format: INV + YYYYMMDD + 4-digit sequence
	now := time.Now()
	dateStr := now.Format("20060102")

	// Get the count of invoices created today
	var count int64
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	is.db.Model(&entities.Invoice{}).
		Where("created_at >= ? AND created_at < ?", startOfDay, endOfDay).
		Count(&count)

	sequence := count + 1
	return fmt.Sprintf("INV%s%04d", dateStr, sequence)
}
