package service

import (
	"context"
	"fmt"
	"time"

	"linke/internal/logger"
	"linke/internal/model"

	"gorm.io/gorm"
)

type InvoiceService struct {
	db          *gorm.DB
	userService *UserService
}

func NewInvoiceService(db *gorm.DB, userService *UserService) *InvoiceService {
	return &InvoiceService{
		db:          db,
		userService: userService,
	}
}

// CreateInvoiceFromOrderRequest represents the request to create an invoice from an order
type CreateInvoiceFromOrderRequest struct {
	OrderID          uint    `json:"order_id" binding:"required"`
	InvoiceType      string  `json:"invoice_type,omitempty" example:"standard"`
	TaxRate          float64 `json:"tax_rate,omitempty" example:"0.2"`
	TaxType          string  `json:"tax_type,omitempty" example:"VAT"`
	TaxNumber        string  `json:"tax_number,omitempty" example:"GB123456789"`
	BillingName      string  `json:"billing_name,omitempty"`
	BillingEmail     string  `json:"billing_email,omitempty"`
	BillingAddress   string  `json:"billing_address,omitempty"`
	BillingCity      string  `json:"billing_city,omitempty"`
	BillingState     string  `json:"billing_state,omitempty"`
	BillingCountry   string  `json:"billing_country,omitempty"`
	BillingZip       string  `json:"billing_zip,omitempty"`
	CompanyName      string  `json:"company_name,omitempty"`
	CompanyTaxID     string  `json:"company_tax_id,omitempty"`
	CompanyAddress   string  `json:"company_address,omitempty"`
	Description      string  `json:"description,omitempty"`
	Notes            string  `json:"notes,omitempty"`
	DueDate          string  `json:"due_date,omitempty" example:"2024-01-31"`
	PaymentTermsDays *int    `json:"payment_terms_days,omitempty" example:"30"`
	Template         string  `json:"template,omitempty" example:"default"`
	Language         string  `json:"language,omitempty" example:"en"`
	AutoSend         bool    `json:"auto_send,omitempty" example:"false"`
}

// CreateInvoiceFromOrder creates an invoice from a confirmed order
func (is *InvoiceService) CreateInvoiceFromOrder(ctx context.Context, req *CreateInvoiceFromOrderRequest) (*model.Invoice, error) {
	// Get order with user details
	var order model.Order
	if err := is.db.WithContext(ctx).Preload("User").First(&order, req.OrderID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("order not found")
		}
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	// Only create invoice for confirmed orders
	if !order.IsConfirmed() {
		return nil, fmt.Errorf("can only create invoices for confirmed orders, current status: %s", order.Status)
	}

	// Check if invoice already exists for this order
	var existingInvoice model.Invoice
	if err := is.db.WithContext(ctx).Where("order_id = ?", req.OrderID).First(&existingInvoice).Error; err == nil {
		return nil, fmt.Errorf("invoice already exists for this order")
	}

	// Set defaults
	invoiceType := req.InvoiceType
	if invoiceType == "" {
		invoiceType = model.InvoiceTypeStandard
	}

	template := req.Template
	if template == "" {
		template = "default"
	}

	language := req.Language
	if language == "" {
		language = "en"
	}

	paymentTermsDays := 30
	if req.PaymentTermsDays != nil {
		paymentTermsDays = *req.PaymentTermsDays
	}

	// Calculate tax amount
	taxAmount := order.TotalAmount * req.TaxRate
	totalAmount := order.TotalAmount + taxAmount

	// Parse due date
	var dueAt *time.Time
	if req.DueDate != "" {
		if parsed, err := time.Parse("2006-01-02", req.DueDate); err == nil {
			// Set due time to end of day
			due := time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 23, 59, 59, 0, parsed.Location())
			dueAt = &due
		}
	}

	// Use billing information from request or order's user
	billingName := req.BillingName
	if billingName == "" && order.User != nil {
		billingName = order.User.Name
		if billingName == "" {
			billingName = order.User.Username
		}
	}

	billingEmail := req.BillingEmail
	if billingEmail == "" && order.User != nil {
		billingEmail = order.User.Email
	}

	if billingName == "" || billingEmail == "" {
		return nil, fmt.Errorf("billing name and email are required")
	}

	description := req.Description
	if description == "" {
		description = fmt.Sprintf("Order %s - Service subscription", order.OrderNumber)
	}

	// Generate invoice number
	invoiceNumber := is.generateInvoiceNumber()

	// Create invoice
	now := time.Now()
	invoice := &model.Invoice{
		OrderID:          req.OrderID,
		UserID:           order.UserID,
		InvoiceNumber:    invoiceNumber,
		InvoiceType:      invoiceType,
		Status:           model.InvoiceStatusDraft,
		Subtotal:         order.TotalAmount,
		TaxRate:          req.TaxRate,
		TaxAmount:        taxAmount,
		TotalAmount:      totalAmount,
		Currency:         order.Currency,
		IssuedAt:         &now,
		DueAt:            dueAt,
		PaymentTermsDays: paymentTermsDays,
		BillingName:      billingName,
		BillingEmail:     billingEmail,
		BillingAddress:   req.BillingAddress,
		BillingCity:      req.BillingCity,
		BillingState:     req.BillingState,
		BillingCountry:   req.BillingCountry,
		BillingZip:       req.BillingZip,
		TaxNumber:        req.TaxNumber,
		TaxType:          req.TaxType,
		CompanyName:      req.CompanyName,
		CompanyTaxID:     req.CompanyTaxID,
		CompanyAddress:   req.CompanyAddress,
		Description:      description,
		Template:         template,
		Language:         language,
		Notes:            req.Notes,
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
		logger.Uint("order_id", req.OrderID))

	return invoice, nil
}

// SendInvoice sends an invoice to the customer and creates payment record
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
		"status":     model.InvoiceStatusSent,
		"sent_at":    now,
		"send_count": invoice.SendCount + 1,
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
func (is *InvoiceService) MarkInvoiceAsPaid(ctx context.Context, invoiceID uint, paidAmount float64, paymentReference string) error {
	// Start transaction
	tx := is.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Get invoice with lock
	var invoice model.Invoice
	if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&invoice, invoiceID).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("invoice not found")
		}
		return fmt.Errorf("failed to get invoice: %w", err)
	}

	// Check if invoice can be paid
	if !invoice.CanBePaid() {
		tx.Rollback()
		return fmt.Errorf("invoice cannot be marked as paid in status: %s", invoice.Status)
	}

	// Update invoice status
	now := time.Now()
	newPaidAmount := invoice.PaidAmount + paidAmount
	updateData := map[string]interface{}{
		"paid_amount": newPaidAmount,
		"paid_at":     now,
		"updated_at":  now,
	}

	// Check if fully paid
	if newPaidAmount >= invoice.TotalAmount {
		updateData["status"] = model.InvoiceStatusPaid
	}

	if err := tx.Model(&invoice).Updates(updateData).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update invoice status: %w", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit invoice payment: %w", err)
	}

	logger.Info("Invoice marked as paid",
		logger.Uint("invoice_id", invoiceID),
		logger.String("invoice_number", invoice.InvoiceNumber),
		logger.Any("paid_amount", paidAmount),
		logger.String("payment_reference", paymentReference))

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
		"status":      model.InvoiceStatusVoided,
		"voided_at":   now,
		"void_reason": reason,
		"updated_at":  now,
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

// GetInvoice gets an invoice by ID
func (is *InvoiceService) GetInvoice(ctx context.Context, invoiceID uint) (*model.Invoice, error) {
	var invoice model.Invoice
	if err := is.db.WithContext(ctx).First(&invoice, invoiceID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("invoice not found")
		}
		return nil, fmt.Errorf("failed to get invoice: %w", err)
	}
	return &invoice, nil
}

// GetInvoiceWithRelations gets an invoice with related data
func (is *InvoiceService) GetInvoiceWithRelations(ctx context.Context, invoiceID uint) (*model.Invoice, error) {
	var invoice model.Invoice
	if err := is.db.WithContext(ctx).
		Preload("Order").
		Preload("User").
		First(&invoice, invoiceID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("invoice not found")
		}
		return nil, fmt.Errorf("failed to get invoice: %w", err)
	}
	return &invoice, nil
}

// InvoiceFilters represents filters for invoice listing
type InvoiceFilters struct {
	UserID      *uint  `form:"user_id"`
	OrderID     *uint  `form:"order_id"`
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

// ListInvoices lists invoices with filtering and pagination
func (is *InvoiceService) ListInvoices(ctx context.Context, filters *InvoiceFilters) ([]*model.Invoice, int64, error) {
	query := is.db.WithContext(ctx).Model(&model.Invoice{})

	// Apply filters
	if filters.UserID != nil {
		query = query.Where("user_id = ?", *filters.UserID)
	}

	if filters.OrderID != nil {
		query = query.Where("order_id = ?", *filters.OrderID)
	}

	if filters.Status != "" {
		query = query.Where("status = ?", filters.Status)
	}

	if filters.InvoiceType != "" {
		query = query.Where("invoice_type = ?", filters.InvoiceType)
	}

	if filters.Currency != "" {
		query = query.Where("currency = ?", filters.Currency)
	}

	// Date range filtering
	if filters.StartDate != "" {
		if startDate, err := time.Parse("2006-01-02", filters.StartDate); err == nil {
			query = query.Where("issued_at >= ?", startDate)
		}
	}

	if filters.EndDate != "" {
		if endDate, err := time.Parse("2006-01-02", filters.EndDate); err == nil {
			endDate = endDate.Add(24 * time.Hour)
			query = query.Where("issued_at < ?", endDate)
		}
	}

	// Overdue filter
	if filters.Overdue != nil {
		if *filters.Overdue {
			query = query.Where("due_at < ? AND status != ?", time.Now(), model.InvoiceStatusPaid)
		} else {
			query = query.Where("due_at >= ? OR status = ?", time.Now(), model.InvoiceStatusPaid)
		}
	}

	// Search functionality
	if filters.Search != "" {
		searchPattern := "%" + filters.Search + "%"
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
	sortBy := filters.SortBy
	if sortBy == "" {
		sortBy = "created_at"
	}

	sortOrder := filters.SortOrder
	if sortOrder == "" {
		sortOrder = "desc"
	}

	validSortFields := map[string]bool{
		"created_at":    true,
		"updated_at":    true,
		"issued_at":     true,
		"due_at":        true,
		"paid_at":       true,
		"total_amount":  true,
		"status":        true,
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
	if filters.Limit > 0 {
		query = query.Limit(filters.Limit)
	}

	if filters.Offset > 0 {
		query = query.Offset(filters.Offset)
	}

	var invoices []*model.Invoice
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
func (is *InvoiceService) UpdateInvoice(ctx context.Context, invoiceID uint, req *UpdateInvoiceRequest) (*model.Invoice, error) {
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
		newTaxAmount := invoice.Subtotal * (*req.TaxRate)
		updateData["tax_amount"] = newTaxAmount
		updateData["total_amount"] = invoice.Subtotal + newTaxAmount
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

	is.db.Model(&model.Invoice{}).
		Where("created_at >= ? AND created_at < ?", startOfDay, endOfDay).
		Count(&count)

	sequence := count + 1
	return fmt.Sprintf("INV%s%04d", dateStr, sequence)
}