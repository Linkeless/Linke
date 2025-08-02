package implementations

import (
	"context"
	"fmt"
	"time"

	"linke/internal/domains/invoice/entities"
	userInterfaces "linke/internal/domains/user/usecases/interfaces"
	"linke/internal/shared/logger"

	"gorm.io/gorm"
)

type InvoiceService struct {
	db          *gorm.DB
	userService userInterfaces.UserService
}

func NewInvoiceService(db *gorm.DB, userService userInterfaces.UserService) *InvoiceService {
	return &InvoiceService{
		db:          db,
		userService: userService,
	}
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

// CreateInvoice creates a new invoice
func (is *InvoiceService) CreateInvoice(ctx context.Context, req *CreateInvoiceRequest) (*entities.Invoice, error) {
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
		if err := is.SendInvoice(ctx, invoice.ID); err != nil {
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
func (is *InvoiceService) CreateInvoiceFromOrder(ctx context.Context, orderID uint, options *CreateInvoiceRequest) (*entities.Invoice, error) {
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

	req := &CreateInvoiceRequest{
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

// GetInvoicesRequest represents the request to get invoices with filtering
type GetInvoicesRequest struct {
	UserID      *uint  `form:"user_id"`
	Status      string `form:"status"`
	InvoiceType string `form:"invoice_type"`
	Currency    string `form:"currency"`
	StartDate   string `form:"start_date"`
	EndDate     string `form:"end_date"`
	Overdue     *bool  `form:"overdue"`
	Search      string `form:"search"`
	SortBy      string `form:"sort_by"`
	SortOrder   string `form:"sort_order"`
	Limit       int    `form:"limit"`
	Offset      int    `form:"offset"`
}

// GetInvoices gets invoices with filtering
func (is *InvoiceService) GetInvoices(ctx context.Context, req *GetInvoicesRequest) ([]*entities.Invoice, int64, error) {
	query := is.db.WithContext(ctx).Model(&entities.Invoice{})

	// Apply filters
	if req.UserID != nil {
		query = query.Where("user_id = ?", *req.UserID)
	}

	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}

	if req.InvoiceType != "" {
		query = query.Where("invoice_type = ?", req.InvoiceType)
	}

	if req.Currency != "" {
		query = query.Where("currency = ?", req.Currency)
	}

	// Date range filtering
	if req.StartDate != "" {
		if startDate, err := time.Parse("2006-01-02", req.StartDate); err == nil {
			query = query.Where("issued_at >= ?", startDate)
		}
	}

	if req.EndDate != "" {
		if endDate, err := time.Parse("2006-01-02", req.EndDate); err == nil {
			endDate = endDate.Add(24 * time.Hour)
			query = query.Where("issued_at < ?", endDate)
		}
	}

	// Overdue filter
	if req.Overdue != nil {
		if *req.Overdue {
			query = query.Where("due_at < ? AND status != ?", time.Now(), entities.InvoiceStatusPaid)
		} else {
			query = query.Where("due_at >= ? OR status = ?", time.Now(), entities.InvoiceStatusPaid)
		}
	}

	// Search functionality
	if req.Search != "" {
		searchPattern := "%" + req.Search + "%"
		query = query.Where(
			"invoice_number LIKE ? OR billing_name LIKE ? OR billing_email LIKE ? OR company_name LIKE ?",
			searchPattern, searchPattern, searchPattern, searchPattern,
		)
	}

	// Get total count
	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count invoices: %w", err)
	}

	// Apply sorting
	sortBy := req.SortBy
	if sortBy == "" {
		sortBy = "created_at"
	}

	sortOrder := req.SortOrder
	if sortOrder == "" {
		sortOrder = "desc"
	}

	validSortFields := map[string]bool{
		"created_at":     true,
		"updated_at":     true,
		"issued_at":      true,
		"due_at":         true,
		"paid_at":        true,
		"total_amount":   true,
		"status":         true,
		"invoice_number": true,
	}

	if !validSortFields[sortBy] {
		sortBy = "created_at"
	}

	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "desc"
	}

	query = query.Order(fmt.Sprintf("%s %s", sortBy, sortOrder))

	// Apply pagination
	if req.Limit > 0 {
		query = query.Limit(req.Limit)
	}

	if req.Offset > 0 {
		query = query.Offset(req.Offset)
	}

	// Load relations
	query = query.Preload("User").Preload("SubscriptionOrder")

	var invoices []*entities.Invoice
	if err := query.Find(&invoices).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get invoices: %w", err)
	}

	return invoices, totalCount, nil
}

// UpdateInvoiceRequest represents the request to update an invoice
type UpdateInvoiceRequest struct {
	InvoiceType    *string  `json:"invoice_type,omitempty"`
	TaxRate        *float64 `json:"tax_rate,omitempty"`
	TaxType        *string  `json:"tax_type,omitempty"`
	TaxNumber      *string  `json:"tax_number,omitempty"`
	BillingName    *string  `json:"billing_name,omitempty"`
	BillingEmail   *string  `json:"billing_email,omitempty"`
	BillingAddress *string  `json:"billing_address,omitempty"`
	BillingCity    *string  `json:"billing_city,omitempty"`
	BillingState   *string  `json:"billing_state,omitempty"`
	BillingCountry *string  `json:"billing_country,omitempty"`
	BillingZip     *string  `json:"billing_zip,omitempty"`
	CompanyName    *string  `json:"company_name,omitempty"`
	CompanyTaxID   *string  `json:"company_tax_id,omitempty"`
	CompanyAddress *string  `json:"company_address,omitempty"`
	Description    *string  `json:"description,omitempty"`
	Notes          *string  `json:"notes,omitempty"`
	DueDate        *string  `json:"due_date,omitempty"`
	Template       *string  `json:"template,omitempty"`
	Language       *string  `json:"language,omitempty"`
}

// UpdateInvoice updates an invoice
func (is *InvoiceService) UpdateInvoice(ctx context.Context, invoiceID uint, req *UpdateInvoiceRequest) (*entities.Invoice, error) {
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
	updateData := make(map[string]interface{})

	if req.InvoiceType != nil {
		updateData["invoice_type"] = *req.InvoiceType
	}
	if req.TaxType != nil {
		updateData["tax_type"] = *req.TaxType
	}
	if req.TaxNumber != nil {
		updateData["tax_number"] = *req.TaxNumber
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
func (is *InvoiceService) SendInvoice(ctx context.Context, invoiceID uint) error {
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
	updateData := map[string]interface{}{
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
func (is *InvoiceService) MarkInvoiceAsPaid(ctx context.Context, invoiceID uint, paymentMethod, paymentReference string) error {
	// Get invoice
	invoice, err := is.GetInvoice(ctx, invoiceID)
	if err != nil {
		return err
	}

	// Check if invoice can be paid
	if !invoice.CanBePaid() {
		return fmt.Errorf("invoice cannot be marked as paid in status: %s", invoice.Status)
	}

	// Update invoice status
	now := time.Now()
	updateData := map[string]interface{}{
		"status":            entities.InvoiceStatusPaid,
		"paid_at":           now,
		"payment_method":    paymentMethod,
		"payment_reference": paymentReference,
		"updated_at":        now,
	}

	if err := is.db.WithContext(ctx).Model(invoice).Updates(updateData).Error; err != nil {
		return fmt.Errorf("failed to update invoice status: %w", err)
	}

	logger.Info("Invoice marked as paid",
		logger.Uint("invoice_id", invoiceID),
		logger.String("invoice_number", invoice.InvoiceNumber),
		logger.String("payment_method", paymentMethod))

	return nil
}

// VoidInvoice voids an invoice
func (is *InvoiceService) VoidInvoice(ctx context.Context, invoiceID uint, reason string) error {
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
	updateData := map[string]interface{}{
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
