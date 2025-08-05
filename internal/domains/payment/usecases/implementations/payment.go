package implementations

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	invoiceInterfaces "linke/internal/domains/invoice/usecases/interfaces"
	"linke/internal/domains/payment/entities"
	"linke/internal/domains/payment/usecases/interfaces"
	"linke/internal/shared/logger"

	"gorm.io/gorm"
)

// PaymentService represents the unified payment service
type PaymentService struct {
	db                       *gorm.DB
	gateways                 map[string]interfaces.PaymentGateway
	subscriptionOrderService interfaces.SubscriptionOrderServiceInterface
	invoiceService           invoiceInterfaces.InvoiceService
}

// NewPaymentService creates a new payment service instance
func NewPaymentService(db *gorm.DB) *PaymentService {
	return &PaymentService{
		db:       db,
		gateways: make(map[string]interfaces.PaymentGateway),
	}
}

// RegisterGateway registers a payment gateway
func (ps *PaymentService) RegisterGateway(name string, gateway interfaces.PaymentGateway) error {
	if err := gateway.ValidateConfig(); err != nil {
		return fmt.Errorf("gateway config validation failed: %w", err)
	}

	ps.gateways[name] = gateway
	logger.Info("Payment gateway registered", logger.String("gateway", name))
	return nil
}

// GetGateway gets a payment gateway by name
func (ps *PaymentService) GetGateway(name string) (interfaces.PaymentGateway, error) {
	gateway, exists := ps.gateways[name]
	if !exists {
		return nil, fmt.Errorf("payment gateway '%s' not found", name)
	}
	return gateway, nil
}

// GeneratePaymentNo generates a unique payment number
func (ps *PaymentService) GeneratePaymentNo() (string, error) {
	// Generate format: PAY + YYYYMMDD + 8-digit random hex
	now := time.Now()
	dateStr := now.Format("20060102")

	// Generate 4 random bytes (8 hex characters)
	randomBytes := make([]byte, 4)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	randomStr := strings.ToUpper(hex.EncodeToString(randomBytes))
	paymentNo := fmt.Sprintf("PAY%s%s", dateStr, randomStr)

	// Check if already exists (very unlikely but possible)
	var existingRecord entities.PaymentRecord
	if err := ps.db.Where("payment_no = ?", paymentNo).First(&existingRecord).Error; err == nil {
		// Payment number exists, try again (recursive call)
		return ps.GeneratePaymentNo()
	}

	return paymentNo, nil
}

// CreatePaymentOrder creates a new payment order
func (ps *PaymentService) CreatePaymentOrder(ctx context.Context, req *interfaces.CreatePaymentOrderRequest) (*entities.PaymentRecord, error) {
	// Get gateway
	gateway, err := ps.GetGateway(req.Gateway)
	if err != nil {
		return nil, err
	}

	// Validate payment method
	supportedMethods := gateway.GetSupportedPaymentMethods()
	methodSupported := false
	for _, method := range supportedMethods {
		if method == req.PaymentMethod {
			methodSupported = true
			break
		}
	}
	if !methodSupported {
		return nil, fmt.Errorf("payment method '%s' not supported by gateway '%s'", req.PaymentMethod, req.Gateway)
	}

	// Generate payment number and out trade number
	paymentNo, err := ps.GeneratePaymentNo()
	if err != nil {
		return nil, fmt.Errorf("failed to generate payment number: %w", err)
	}

	outTradeNo := paymentNo // Use payment number as out trade number

	// Calculate expiration time
	expiredAt := time.Now().Add(time.Duration(req.ExpiredMinutes) * time.Minute)
	if req.ExpiredMinutes <= 0 {
		expiredAt = time.Now().Add(30 * time.Minute) // Default 30 minutes
	}

	// Create gateway order request
	gatewayReq := &interfaces.CreatePaymentOrderRequest{
		UserID:              req.UserID,
		SubscriptionOrderID: req.SubscriptionOrderID,
		InvoiceID:           req.InvoiceID,
		Gateway:             req.Gateway,
		PaymentMethod:       req.PaymentMethod,
		Amount:              req.Amount,
		Currency:            req.Currency,
		Subject:             req.Subject,
		Body:                req.Body,
		ClientIP:            req.ClientIP,
		NotifyURL:           req.NotifyURL,
		ReturnURL:           req.ReturnURL,
	}

	// Create order through gateway using unified interface
	gatewayResp, err := gateway.CreatePaymentOrder(gatewayReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment order through gateway: %w", err)
	}

	// Store raw gateway response
	rawData, _ := json.Marshal(gatewayResp)
	gatewayData := string(rawData)

	// Create payment record
	paymentRecord := &entities.PaymentRecord{
		UserID:              req.UserID,
		SubscriptionOrderID: req.SubscriptionOrderID,
		InvoiceID:           req.InvoiceID,
		PaymentNo:           paymentNo,
		OutTradeNo:          outTradeNo,
		Gateway:             req.Gateway,
		PaymentMethod:       req.PaymentMethod,
		Amount:              req.Amount,
		Currency:            req.Currency,
		Status:              entities.PaymentRecordStatusPending,
		GatewayResponse:     gatewayData,
		PaymentURL:          gatewayResp.PaymentURL,
		QRCodeURL:           gatewayResp.QRCodeURL,
		ExpiredAt:           &expiredAt,
		ClientIP:            req.ClientIP,
		NotifyURL:           req.NotifyURL,
		ReturnURL:           req.ReturnURL,
		Metadata:            req.Metadata,
	}

	// Save to database
	if err := ps.db.WithContext(ctx).Create(paymentRecord).Error; err != nil {
		logger.Error("Failed to create payment record", logger.Error2("error", err))
		return nil, fmt.Errorf("failed to create payment record: %w", err)
	}

	logger.Info("Payment order created successfully",
		logger.String("payment_no", paymentNo),
		logger.String("gateway", req.Gateway),
		logger.String("method", req.PaymentMethod),
		logger.Uint("user_id", req.UserID))

	return paymentRecord, nil
}

// GetPaymentRecord gets a payment record by payment number
func (ps *PaymentService) GetPaymentRecord(ctx context.Context, paymentNo string) (*entities.PaymentRecord, error) {
	var record entities.PaymentRecord
	if err := ps.db.WithContext(ctx).Where("payment_no = ?", paymentNo).First(&record).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("payment record not found")
		}
		logger.Error("Failed to get payment record", logger.Error2("error", err))
		return nil, fmt.Errorf("failed to get payment record: %w", err)
	}
	return &record, nil
}

// GetPaymentRecordByOutTradeNo gets a payment record by out trade number
func (ps *PaymentService) GetPaymentRecordByOutTradeNo(ctx context.Context, outTradeNo string) (*entities.PaymentRecord, error) {
	var record entities.PaymentRecord
	if err := ps.db.WithContext(ctx).Where("out_trade_no = ?", outTradeNo).First(&record).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("payment record not found")
		}
		logger.Error("Failed to get payment record by out trade no", logger.Error2("error", err))
		return nil, fmt.Errorf("failed to get payment record: %w", err)
	}
	return &record, nil
}

// UpdatePaymentStatus updates payment record status
func (ps *PaymentService) UpdatePaymentStatus(ctx context.Context, paymentNo string, status string, transactionID string, paidAt *time.Time) error {
	updates := map[string]any{
		"status":      status,
		"updated_at":  time.Now(),
		"notified_at": time.Now(),
	}

	if transactionID != "" {
		updates["transaction_id"] = transactionID
	}

	if paidAt != nil {
		updates["paid_at"] = paidAt
	}

	if err := ps.db.WithContext(ctx).Model(&entities.PaymentRecord{}).
		Where("payment_no = ?", paymentNo).
		Updates(updates).Error; err != nil {
		logger.Error("Failed to update payment status", logger.Error2("error", err))
		return fmt.Errorf("failed to update payment status: %w", err)
	}

	logger.Info("Payment status updated",
		logger.String("payment_no", paymentNo),
		logger.String("status", status))

	return nil
}

// ProcessNotification processes payment notification from gateway
func (ps *PaymentService) ProcessNotification(ctx context.Context, gateway string, data map[string]any) error {
	// Get gateway instance
	gatewayInstance, err := ps.GetGateway(gateway)
	if err != nil {
		return err
	}

	// Verify notification
	isValid, notifyData := gatewayInstance.VerifyPaymentNotify(data)
	if !isValid {
		return fmt.Errorf("notification signature verification failed")
	}

	// Get payment record with row lock to prevent race conditions
	var paymentRecord *entities.PaymentRecord
	if err := ps.db.WithContext(ctx).Set("gorm:query_option", "FOR UPDATE").
		Where("out_trade_no = ?", notifyData.OutTradeNo).
		First(&paymentRecord).Error; err != nil {
		return fmt.Errorf("payment record not found: %w", err)
	}

	// SECURITY: Enhanced idempotency check
	notifyHash := ps.generateNotificationHash(data)

	// Safely get client IP from context (set by middleware)
	var clientIP string
	if ip := ctx.Value("client_ip"); ip != nil {
		if ipStr, ok := ip.(string); ok {
			clientIP = ipStr
		}
	}
	if clientIP == "" {
		clientIP = "unknown" // Fallback
	}

	// Check for exact duplicate (same hash)
	if paymentRecord.LastNotifyHash == notifyHash {
		logger.Warn("Duplicate notification detected, ignoring",
			logger.String("payment_no", paymentRecord.PaymentNo),
			logger.String("gateway", gateway),
			logger.String("notify_hash", notifyHash),
			logger.String("client_ip", clientIP))
		return nil // Silently ignore duplicate notifications
	}

	// Check for time-based replay attack protection
	if paymentRecord.LastNotifyTime != nil {
		timeSinceLastNotify := time.Since(*paymentRecord.LastNotifyTime)
		if timeSinceLastNotify < 30*time.Second {
			logger.Warn("Notification received too soon after last notification",
				logger.String("payment_no", paymentRecord.PaymentNo),
				logger.String("gateway", gateway),
				logger.Duration("time_since_last", timeSinceLastNotify))
			return fmt.Errorf("notification rate limit exceeded")
		}
	}

	// Check for suspicious IP changes (optional security measure)
	if paymentRecord.NotifySource != "" && paymentRecord.NotifySource != clientIP {
		logger.Warn("Notification source IP changed",
			logger.String("payment_no", paymentRecord.PaymentNo),
			logger.String("previous_ip", paymentRecord.NotifySource),
			logger.String("current_ip", clientIP))
		// Don't block, but log for monitoring
	}

	// SECURITY: Check for status downgrade attempts
	if paymentRecord.Status == entities.PaymentRecordStatusCompleted &&
		!gatewayInstance.IsPaymentCompleted(notifyData.Status) {
		logger.Warn("Attempted status downgrade from completed, ignoring",
			logger.String("payment_no", paymentRecord.PaymentNo),
			logger.String("current_status", paymentRecord.Status),
			logger.String("notify_status", notifyData.Status))
		return nil
	}

	// Update notification tracking with enhanced security fields
	now := time.Now()
	updateFields := map[string]any{
		"last_notify_hash": notifyHash,
		"notify_count":     paymentRecord.NotifyCount + 1,
		"notified_at":      &now,
		"last_notify_time": &now,
		"notify_source":    clientIP,
	}

	// Check if payment is completed
	if gatewayInstance.IsPaymentCompleted(notifyData.Status) {
		paidAt := time.Now()
		if notifyData.PaidAt != "" {
			if parsedTime, parseErr := time.Parse("2006-01-02 15:04:05", notifyData.PaidAt); parseErr == nil {
				paidAt = parsedTime
			}
		}

		// Update payment status with notification tracking
		updateFields["status"] = entities.PaymentRecordStatusCompleted
		updateFields["transaction_id"] = notifyData.TransactionID
		updateFields["paid_at"] = &paidAt

		if err := ps.db.WithContext(ctx).Model(paymentRecord).Updates(updateFields).Error; err != nil {
			return fmt.Errorf("failed to update payment status: %w", err)
		}

		// Process order completion (integrate with subscription system)
		if err := ps.processOrderCompletion(ctx, paymentRecord); err != nil {
			logger.Error("Failed to process order completion", logger.Error2("error", err))
			// Don't return error here as payment is already processed
		}
	} else {
		// Update status based on notification
		var status string
		switch notifyData.Status {
		case "TRADE_CLOSED", "TRADE_CANCELLED":
			status = entities.PaymentRecordStatusCancelled
		case "TRADE_FAILED":
			status = entities.PaymentRecordStatusFailed
		default:
			status = entities.PaymentRecordStatusProcessing
		}

		updateFields["status"] = status
		updateFields["transaction_id"] = notifyData.TransactionID

		if err := ps.db.WithContext(ctx).Model(paymentRecord).Updates(updateFields).Error; err != nil {
			return fmt.Errorf("failed to update payment status: %w", err)
		}
	}

	return nil
}

// SetSubscriptionOrderService sets the subscription order service for payment processing
func (ps *PaymentService) SetSubscriptionOrderService(subscriptionOrderService interfaces.SubscriptionOrderServiceInterface) {
	ps.subscriptionOrderService = subscriptionOrderService
}

// SetInvoiceService sets the invoice service for payment processing
func (ps *PaymentService) SetInvoiceService(invoiceService invoiceInterfaces.InvoiceService) {
	ps.invoiceService = invoiceService
}

// processOrderCompletion processes the completion of an order after successful payment
func (ps *PaymentService) processOrderCompletion(ctx context.Context, paymentRecord *entities.PaymentRecord) error {
	// If this payment is for an invoice, mark the invoice as paid first
	if paymentRecord.InvoiceID != nil {
		if ps.invoiceService != nil {
			// Mark invoice as paid
			paidAt := time.Now().Format("2006-01-02")
			if err := ps.invoiceService.MarkInvoiceAsPaid(ctx, *paymentRecord.InvoiceID, paidAt); err != nil {
				return fmt.Errorf("failed to mark invoice as paid: %w", err)
			}

			logger.Info("Invoice marked as paid",
				logger.Uint("invoice_id", *paymentRecord.InvoiceID),
				logger.String("payment_no", paymentRecord.PaymentNo))

			// For invoice payments, we need to find the associated order and process it
			// TODO: Add method to get invoice with order details
			// For now, assume invoice payment should also trigger subscription activation
		}
	}

	// If this payment is for a subscription order, process the order completion
	if paymentRecord.SubscriptionOrderID != nil {
		// If subscription order service is available, use it for processing
		if ps.subscriptionOrderService != nil {
			if err := ps.subscriptionOrderService.ProcessOrderPaymentSuccess(ctx, *paymentRecord.SubscriptionOrderID); err != nil {
				return fmt.Errorf("failed to process subscription order payment: %w", err)
			}
		} else {
			// TODO: Update subscription order status through subscription service interface
			// This should be handled by the subscription domain, not payment domain
			logger.Info("Payment completed for subscription order",
				logger.Uint("order_id", *paymentRecord.SubscriptionOrderID),
				logger.String("payment_no", paymentRecord.PaymentNo))
		}

		logger.Info("Subscription order payment processed",
			logger.Uint("order_id", *paymentRecord.SubscriptionOrderID),
			logger.String("payment_no", paymentRecord.PaymentNo))
	}

	// If we have both invoice and subscription order, ensure both are processed
	// This handles the new business flow: Order -> Invoice -> Payment -> Service Activation
	if paymentRecord.InvoiceID != nil && paymentRecord.SubscriptionOrderID != nil {
		logger.Info("Complete business flow processed: Order -> Invoice -> Payment -> Service will be activated",
			logger.Uint("order_id", *paymentRecord.SubscriptionOrderID),
			logger.Uint("invoice_id", *paymentRecord.InvoiceID),
			logger.String("payment_no", paymentRecord.PaymentNo))
	}

	return nil
}

// GetUserPaymentRecords gets payment records for a user with pagination
func (ps *PaymentService) GetUserPaymentRecords(ctx context.Context, userID uint, limit, offset int) ([]*entities.PaymentRecord, int64, error) {
	var records []*entities.PaymentRecord
	var totalCount int64

	// Get total count
	if err := ps.db.WithContext(ctx).Model(&entities.PaymentRecord{}).
		Where("user_id = ?", userID).
		Count(&totalCount).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count payment records: %w", err)
	}

	// Get records with pagination
	query := ps.db.WithContext(ctx).Where("user_id = ?", userID).
		Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&records).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get payment records: %w", err)
	}

	return records, totalCount, nil
}

// GetAvailablePaymentMethods gets available payment methods
func (ps *PaymentService) GetAvailablePaymentMethods(ctx context.Context) (map[string][]string, error) {
	methods := make(map[string][]string)

	for gatewayName, gateway := range ps.gateways {
		methods[gatewayName] = gateway.GetSupportedPaymentMethods()
	}

	return methods, nil
}

// generateNotificationHash generates a hash for notification data to detect duplicates
func (ps *PaymentService) generateNotificationHash(data map[string]any) string {
	// Import crypto/sha256 at the top of the file if not already imported
	// Create a consistent string representation of the notification data
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	// Sort keys for consistent hash generation
	for i := 0; i < len(keys); i++ {
		for j := i + 1; j < len(keys); j++ {
			if keys[i] > keys[j] {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}

	var hashContent string
	for _, key := range keys {
		hashContent += key + ":" + fmt.Sprintf("%v", data[key]) + "|"
	}

	// Generate SHA256 hash
	h := sha256.Sum256([]byte(hashContent))
	return fmt.Sprintf("%x", h)
}
