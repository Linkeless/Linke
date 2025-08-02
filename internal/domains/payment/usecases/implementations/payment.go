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

	"linke/internal/shared/logger"
	"linke/internal/domains/payment/entities"

	"gorm.io/gorm"
)

// PaymentGateway interface defines the methods that all payment gateways must implement
type PaymentGateway interface {
	CreatePaymentOrder(req *CreatePaymentOrderRequest) (*CreatePaymentOrderResponse, error)
	QueryPaymentOrder(outTradeNo string) (*QueryPaymentOrderResponse, error)
	VerifyPaymentNotify(data map[string]interface{}) (bool, *NotifyData)
	IsPaymentCompleted(status string) bool
	GetSupportedPaymentMethods() []string
	GetPaymentMethodName(method string) string
	ValidateConfig() error
	TestConnection() error
}

// PaymentService represents the unified payment service
type PaymentService struct {
	db                       *gorm.DB
	gateways                 map[string]PaymentGateway
	subscriptionOrderService SubscriptionOrderServiceInterface
}

// NewPaymentService creates a new payment service instance
func NewPaymentService(db *gorm.DB) *PaymentService {
	return &PaymentService{
		db:       db,
		gateways: make(map[string]PaymentGateway),
	}
}

// CreatePaymentOrderRequest represents the unified request to create a payment order
type CreatePaymentOrderRequest struct {
	UserID              uint    `json:"user_id"`
	SubscriptionOrderID *uint   `json:"subscription_order_id,omitempty"`
	Gateway             string  `json:"gateway"`             // epay, epusdt
	PaymentMethod       string  `json:"payment_method"`      // alipay, wechat, usdt, etc.
	Amount              float64 `json:"amount"`              // Amount in specified currency
	Currency            string  `json:"currency"`            // CNY, USD, USDT
	Subject             string  `json:"subject"`             // Order subject
	Body                string  `json:"body"`                // Order description
	ClientIP            string  `json:"client_ip"`           // Client IP
	NotifyURL           string  `json:"notify_url"`          // Async notification URL
	ReturnURL           string  `json:"return_url"`          // Sync return URL
	ExpiredMinutes      int     `json:"expired_minutes"`     // Expiration time in minutes
	Metadata            string  `json:"metadata,omitempty"`  // Additional metadata
}

// CreatePaymentOrderResponse represents the unified response from payment order creation
type CreatePaymentOrderResponse struct {
	PaymentNo    string    `json:"payment_no"`              // Internal payment number
	PaymentURL   string    `json:"payment_url"`             // Payment URL
	QRCodeURL    string    `json:"qr_code_url"`             // QR code URL
	Amount       float64   `json:"amount"`                  // Payment amount
	Currency     string    `json:"currency"`                // Currency
	ExpiredAt    time.Time `json:"expired_at"`              // Expiration time
	GatewayData  string    `json:"gateway_data,omitempty"`  // Raw gateway response
}

// QueryPaymentOrderResponse represents the unified response from payment order query
type QueryPaymentOrderResponse struct {
	PaymentNo     string `json:"payment_no"`
	Status        string `json:"status"`
	PaidAmount    string `json:"paid_amount,omitempty"`
	TransactionID string `json:"transaction_id,omitempty"`
	PaidAt        string `json:"paid_at,omitempty"`
}

// NotifyData represents the unified notification data
type NotifyData struct {
	PaymentNo     string  `json:"payment_no"`
	OutTradeNo    string  `json:"out_trade_no"`
	TransactionID string  `json:"transaction_id"`
	Amount        float64 `json:"amount"`
	Status        string  `json:"status"`
	PaidAt        string  `json:"paid_at,omitempty"`
}

// RegisterGateway registers a payment gateway
func (ps *PaymentService) RegisterGateway(name string, gateway PaymentGateway) error {
	if err := gateway.ValidateConfig(); err != nil {
		return fmt.Errorf("gateway config validation failed: %w", err)
	}

	ps.gateways[name] = gateway
	logger.Info("Payment gateway registered", logger.String("gateway", name))
	return nil
}

// GetGateway gets a payment gateway by name
func (ps *PaymentService) GetGateway(name string) (PaymentGateway, error) {
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
func (ps *PaymentService) CreatePaymentOrder(ctx context.Context, req *CreatePaymentOrderRequest) (*entities.PaymentRecord, error) {
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
	gatewayReq := &CreatePaymentOrderRequest{
		UserID:              req.UserID,
		SubscriptionOrderID: req.SubscriptionOrderID,
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
	updates := map[string]interface{}{
		"status":       status,
		"updated_at":   time.Now(),
		"notified_at":  time.Now(),
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
func (ps *PaymentService) ProcessNotification(ctx context.Context, gateway string, data map[string]interface{}) error {
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

	// SECURITY: Idempotency check - generate hash of notification data
	notifyHash := ps.generateNotificationHash(data)
	if paymentRecord.LastNotifyHash == notifyHash {
		logger.Warn("Duplicate notification detected, ignoring",
			logger.String("payment_no", paymentRecord.PaymentNo),
			logger.String("gateway", gateway),
			logger.String("notify_hash", notifyHash))
		return nil // Silently ignore duplicate notifications
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

	// Update notification tracking
	now := time.Now()
	updateFields := map[string]interface{}{
		"last_notify_hash": notifyHash,
		"notify_count":     paymentRecord.NotifyCount + 1,
		"notified_at":      &now,
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
func (ps *PaymentService) SetSubscriptionOrderService(subscriptionOrderService SubscriptionOrderServiceInterface) {
	ps.subscriptionOrderService = subscriptionOrderService
}

// SubscriptionOrderServiceInterface defines the interface for subscription order service
type SubscriptionOrderServiceInterface interface {
	ProcessOrderPaymentSuccess(ctx context.Context, orderID uint) error
}

// processOrderCompletion processes the completion of an order after successful payment
func (ps *PaymentService) processOrderCompletion(ctx context.Context, paymentRecord *entities.PaymentRecord) error {
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
func (ps *PaymentService) generateNotificationHash(data map[string]interface{}) string {
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