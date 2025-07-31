package service

import (
	"context"
	"fmt"
	"time"

	"linke/internal/logger"
	"linke/internal/model"

	"gorm.io/gorm"
)

type PaymentService struct {
	db             *gorm.DB
	invoiceService *InvoiceService
}

func NewPaymentService(db *gorm.DB, invoiceService *InvoiceService) *PaymentService {
	return &PaymentService{
		db:             db,
		invoiceService: invoiceService,
	}
}

// CreatePaymentRequest represents the request to create a payment
type CreatePaymentRequest struct {
	InvoiceID       uint    `json:"invoice_id" binding:"required"`
	UserID          uint    `json:"user_id" binding:"required"`
	Amount          float64 `json:"amount" binding:"required,min=0"`
	Currency        string  `json:"currency" binding:"required"`
	PaymentGateway  string  `json:"payment_gateway" binding:"required"`
	PaymentMethod   string  `json:"payment_method" binding:"required"`
	Description     string  `json:"description,omitempty"`
	Reference       string  `json:"reference,omitempty"`
	ExternalID      string  `json:"external_id,omitempty"`
	GatewayResponse string  `json:"gateway_response,omitempty"`
	ProcessorFee    float64 `json:"processor_fee,omitempty"`
	Notes           string  `json:"notes,omitempty"`
}

// CreatePayment creates a new payment record
func (ps *PaymentService) CreatePayment(ctx context.Context, req *CreatePaymentRequest) (*model.Payment, error) {
	// Verify invoice exists and belongs to user
	invoice, err := ps.invoiceService.GetInvoice(ctx, req.InvoiceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get invoice: %w", err)
	}

	if invoice.UserID != req.UserID {
		return nil, fmt.Errorf("invoice does not belong to the specified user")
	}

	// Validate payment amount doesn't exceed remaining invoice amount
	remainingAmount := invoice.GetRemainingAmount()
	if req.Amount > remainingAmount {
		return nil, fmt.Errorf("payment amount %.2f exceeds remaining invoice amount %.2f", req.Amount, remainingAmount)
	}

	// Start transaction
	tx := ps.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Generate payment reference if not provided
	reference := req.Reference
	if reference == "" {
		reference = ps.generatePaymentReference()
	}

	// Create payment record
	payment := &model.Payment{
		InvoiceID:       req.InvoiceID,
		UserID:          req.UserID,
		PaymentNumber:   reference,
		Amount:          req.Amount,
		Currency:        req.Currency,
		Status:          model.NewPaymentStatusPending,
		PaymentGateway:  req.PaymentGateway,
		PaymentMethod:   req.PaymentMethod,
		GatewayTransactionID: req.ExternalID,
		GatewayFee:      req.ProcessorFee,
		WebhookData:     req.GatewayResponse,
		Notes:           req.Notes,
	}

	if err := tx.Create(payment).Error; err != nil {
		tx.Rollback()
		logger.Error("Failed to create payment", logger.Error2("error", err))
		return nil, fmt.Errorf("failed to create payment: %w", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit payment creation: %w", err)
	}

	logger.Info("Payment created successfully",
		logger.Uint("payment_id", payment.ID),
		logger.String("reference", reference),
		logger.Uint("invoice_id", req.InvoiceID),
		logger.Any("amount", req.Amount))

	return payment, nil
}

// ProcessPayment processes a payment (marks as processing)
func (ps *PaymentService) ProcessPayment(ctx context.Context, paymentID uint) error {
	return ps.updatePaymentStatus(ctx, paymentID, model.NewPaymentStatusProcessing, "Payment is being processed")
}

// CompletePayment marks a payment as completed and updates the invoice
func (ps *PaymentService) CompletePayment(ctx context.Context, paymentID uint, gatewayResponse string) error {
	// Start transaction
	tx := ps.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Get payment with lock
	var payment model.Payment
	if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&payment, paymentID).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("payment not found")
		}
		return fmt.Errorf("failed to get payment: %w", err)
	}

	// Validate payment can be completed
	if !payment.IsPending() && !payment.IsProcessing() {
		tx.Rollback()
		return fmt.Errorf("payment cannot be completed in status: %s", payment.Status)
	}

	// Update payment status
	now := time.Now()
	updateData := map[string]interface{}{
		"status":           model.NewPaymentStatusCompleted,
		"completed_at":     now,
		"webhook_data":     gatewayResponse,
		"updated_at":       now,
	}

	if err := tx.Model(&payment).Updates(updateData).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update payment status: %w", err)
	}

	// Update invoice paid amount
	if err := ps.invoiceService.MarkInvoiceAsPaid(ctx, payment.InvoiceID, payment.Amount, payment.PaymentNumber); err != nil {
		tx.Rollback()
		logger.Error("Failed to update invoice payment",
			logger.Error2("error", err),
			logger.Uint("payment_id", paymentID),
			logger.Uint("invoice_id", payment.InvoiceID))
		return fmt.Errorf("failed to update invoice payment: %w", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit payment completion: %w", err)
	}

	logger.Info("Payment completed successfully",
		logger.Uint("payment_id", paymentID),
		logger.String("payment_number", payment.PaymentNumber),
		logger.Any("amount", payment.Amount))

	return nil
}

// FailPayment marks a payment as failed
func (ps *PaymentService) FailPayment(ctx context.Context, paymentID uint, reason, gatewayResponse string) error {
	// Start transaction
	tx := ps.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Get payment with lock
	var payment model.Payment
	if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&payment, paymentID).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("payment not found")
		}
		return fmt.Errorf("failed to get payment: %w", err)
	}

	// Validate payment can be failed
	if payment.IsCompleted() {
		tx.Rollback()
		return fmt.Errorf("payment cannot be failed in status: %s", payment.Status)
	}

	// Update payment status
	now := time.Now()
	updateData := map[string]interface{}{
		"status":           model.NewPaymentStatusFailed,
		"updated_at":       now,
		"webhook_data":     gatewayResponse,
		"notes":            reason,
	}

	if err := tx.Model(&payment).Updates(updateData).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update payment status: %w", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit payment failure: %w", err)
	}

	logger.Info("Payment marked as failed",
		logger.Uint("payment_id", paymentID),
		logger.String("payment_number", payment.PaymentNumber),
		logger.String("reason", reason))

	return nil
}

// CancelPayment marks a payment as cancelled
func (ps *PaymentService) CancelPayment(ctx context.Context, paymentID uint, reason string) error {
	return ps.updatePaymentStatusWithReason(ctx, paymentID, model.NewPaymentStatusCancelled, reason)
}

// RefundPayment creates a refund for a completed payment
func (ps *PaymentService) RefundPayment(ctx context.Context, paymentID uint, refundAmount float64, reason string) (*model.Payment, error) {
	// Start transaction
	tx := ps.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Get original payment with lock
	var originalPayment model.Payment
	if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&originalPayment, paymentID).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("payment not found")
		}
		return nil, fmt.Errorf("failed to get payment: %w", err)
	}

	// Validate payment can be refunded
	if !originalPayment.CanBeRefunded() {
		tx.Rollback()
		return nil, fmt.Errorf("payment cannot be refunded in status: %s", originalPayment.Status)
	}

	// Validate refund amount
	if refundAmount <= 0 || refundAmount > originalPayment.Amount {
		tx.Rollback()
		return nil, fmt.Errorf("invalid refund amount: %.2f (original: %.2f)", refundAmount, originalPayment.Amount)
	}

	// Create refund payment record (negative amount)
	refund := &model.Payment{
		InvoiceID:      originalPayment.InvoiceID,
		UserID:         originalPayment.UserID,
		Amount:         -refundAmount, // Negative amount for refund
		Currency:       originalPayment.Currency,
		Status:         model.NewPaymentStatusCompleted,
		PaymentGateway: originalPayment.PaymentGateway,
		PaymentMethod:  originalPayment.PaymentMethod,
		PaymentNumber:  ps.generateRefundReference(originalPayment.PaymentNumber),
		RefundReason:   fmt.Sprintf("Refund for payment %s", originalPayment.PaymentNumber),
		Notes:          fmt.Sprintf("Refund reason: %s", reason),
		CompletedAt:    &time.Time{},
	}
	now := time.Now()
	refund.CompletedAt = &now

	if err := tx.Create(refund).Error; err != nil {
		tx.Rollback()
		logger.Error("Failed to create refund", logger.Error2("error", err))
		return nil, fmt.Errorf("failed to create refund: %w", err)
	}

	// Update original payment refund amount
	newRefundedAmount := originalPayment.RefundAmount + refundAmount
	updateData := map[string]interface{}{
		"refund_amount": newRefundedAmount,
		"updated_at":    now,
	}

	// Mark as fully refunded if applicable
	if newRefundedAmount >= originalPayment.Amount {
		updateData["refunded_at"] = now
		updateData["refund_reason"] = reason
	}

	if err := tx.Model(&originalPayment).Updates(updateData).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to update original payment: %w", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit refund: %w", err)
	}

	logger.Info("Refund processed successfully",
		logger.Uint("refund_id", refund.ID),
		logger.Uint("original_payment_id", paymentID),
		logger.Any("refund_amount", refundAmount))

	return refund, nil
}

// GetPayment gets a payment by ID
func (ps *PaymentService) GetPayment(ctx context.Context, paymentID uint) (*model.Payment, error) {
	var payment model.Payment
	if err := ps.db.WithContext(ctx).First(&payment, paymentID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("payment not found")
		}
		return nil, fmt.Errorf("failed to get payment: %w", err)
	}
	return &payment, nil
}

// GetPaymentWithRelations gets a payment with related data
func (ps *PaymentService) GetPaymentWithRelations(ctx context.Context, paymentID uint) (*model.Payment, error) {
	var payment model.Payment
	if err := ps.db.WithContext(ctx).
		Preload("Invoice").
		Preload("User").
		First(&payment, paymentID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("payment not found")
		}
		return nil, fmt.Errorf("failed to get payment: %w", err)
	}
	return &payment, nil
}

// PaymentFilters represents filters for payment listing
type PaymentFilters struct {
	UserID         *uint  `form:"user_id"`
	InvoiceID      *uint  `form:"invoice_id"`
	Status         string `form:"status"`
	PaymentGateway string `form:"payment_gateway"`
	PaymentMethod  string `form:"payment_method"`
	Currency       string `form:"currency"`
	StartDate      string `form:"start_date"`
	EndDate        string `form:"end_date"`
	MinAmount      *float64 `form:"min_amount"`
	MaxAmount      *float64 `form:"max_amount"`
	Search         string `form:"search"`
	SortBy         string `form:"sort_by"`
	SortOrder      string `form:"sort_order"`
	Limit          int    `form:"limit"`
	Offset         int    `form:"offset"`
}

// ListPayments lists payments with filtering and pagination
func (ps *PaymentService) ListPayments(ctx context.Context, filters *PaymentFilters) ([]*model.Payment, int64, error) {
	query := ps.db.WithContext(ctx).Model(&model.Payment{})

	// Apply filters
	if filters.UserID != nil {
		query = query.Where("user_id = ?", *filters.UserID)
	}

	if filters.InvoiceID != nil {
		query = query.Where("invoice_id = ?", *filters.InvoiceID)
	}

	if filters.Status != "" {
		query = query.Where("status = ?", filters.Status)
	}

	if filters.PaymentGateway != "" {
		query = query.Where("payment_gateway = ?", filters.PaymentGateway)
	}

	if filters.PaymentMethod != "" {
		query = query.Where("payment_method = ?", filters.PaymentMethod)
	}

	if filters.Currency != "" {
		query = query.Where("currency = ?", filters.Currency)
	}

	// Amount range filtering
	if filters.MinAmount != nil {
		query = query.Where("amount >= ?", *filters.MinAmount)
	}

	if filters.MaxAmount != nil {
		query = query.Where("amount <= ?", *filters.MaxAmount)
	}

	// Date range filtering
	if filters.StartDate != "" {
		if startDate, err := time.Parse("2006-01-02", filters.StartDate); err == nil {
			query = query.Where("created_at >= ?", startDate)
		}
	}

	if filters.EndDate != "" {
		if endDate, err := time.Parse("2006-01-02", filters.EndDate); err == nil {
			endDate = endDate.Add(24 * time.Hour)
			query = query.Where("created_at < ?", endDate)
		}
	}

	// Search functionality
	if filters.Search != "" {
		searchPattern := "%" + filters.Search + "%"
		query = query.Where(
			"payment_number LIKE ? OR gateway_transaction_id LIKE ? OR notes LIKE ?",
			searchPattern, searchPattern, searchPattern,
		)
	}

	// Get total count
	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count payments: %w", err)
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
		"created_at":   true,
		"updated_at":   true,
		"completed_at": true,
		"failed_at":    true,
		"amount":       true,
		"status":       true,
		"payment_number": true,
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

	var payments []*model.Payment
	if err := query.Find(&payments).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get payments: %w", err)
	}

	return payments, totalCount, nil
}

// Helper methods

func (ps *PaymentService) updatePaymentStatus(ctx context.Context, paymentID uint, status, reason string) error {
	updateData := map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}

	if reason != "" {
		updateData["notes"] = reason
	}

	if err := ps.db.WithContext(ctx).Model(&model.Payment{}).Where("id = ?", paymentID).Updates(updateData).Error; err != nil {
		return fmt.Errorf("failed to update payment status: %w", err)
	}

	return nil
}

func (ps *PaymentService) updatePaymentStatusWithReason(ctx context.Context, paymentID uint, status, reason string) error {
	updateData := map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
		"notes":      reason,
	}

	if err := ps.db.WithContext(ctx).Model(&model.Payment{}).Where("id = ?", paymentID).Updates(updateData).Error; err != nil {
		return fmt.Errorf("failed to update payment status: %w", err)
	}

	return nil
}

func (ps *PaymentService) generatePaymentReference() string {
	now := time.Now()
	timestamp := now.Format("20060102150405")
	random := now.Nanosecond() % 1000
	return fmt.Sprintf("PMT%s%03d", timestamp, random)
}

func (ps *PaymentService) generateRefundReference(originalRef string) string {
	now := time.Now()
	timestamp := now.Format("20060102150405")
	random := now.Nanosecond() % 1000
	return fmt.Sprintf("RFD%s%03d", timestamp, random)
}