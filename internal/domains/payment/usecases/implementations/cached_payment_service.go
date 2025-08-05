package implementations

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"linke/internal/domains/payment/entities"
	"linke/internal/domains/payment/usecases/interfaces"
	"linke/internal/shared/cache"
	"linke/internal/shared/logger"
)

// CachedPaymentService wraps PaymentService with caching capabilities
type CachedPaymentService struct {
	base         *PaymentService
	cacheManager cache.CacheManager
	cacheKeys    *cache.PaymentCacheKeys
}

// NewCachedPaymentService creates a new cached payment service
func NewCachedPaymentService(
	base *PaymentService,
	cacheManager cache.CacheManager,
	allKeys *cache.AllCacheKeys,
) *CachedPaymentService {
	return &CachedPaymentService{
		base:         base,
		cacheManager: cacheManager,
		cacheKeys:    allKeys.Payment,
	}
}

// RegisterGateway registers a payment gateway
func (cs *CachedPaymentService) RegisterGateway(name string, gateway interfaces.PaymentGateway) error {
	return cs.base.RegisterGateway(name, gateway)
}

// GetGateway gets a payment gateway by name
func (cs *CachedPaymentService) GetGateway(name string) (interfaces.PaymentGateway, error) {
	return cs.base.GetGateway(name)
}

// GeneratePaymentNo generates a unique payment number
func (cs *CachedPaymentService) GeneratePaymentNo() (string, error) {
	return cs.base.GeneratePaymentNo()
}

// CreatePaymentOrder creates a new payment order
func (cs *CachedPaymentService) CreatePaymentOrder(ctx context.Context, req *interfaces.CreatePaymentOrderRequest) (*entities.PaymentRecord, error) {
	record, err := cs.base.CreatePaymentOrder(ctx, req)
	if err != nil {
		return nil, err
	}

	// Cache the payment record with short TTL for security
	cs.cachePaymentRecord(ctx, record)

	return record, nil
}

// GetPaymentRecord gets a payment record by payment number with caching
func (cs *CachedPaymentService) GetPaymentRecord(ctx context.Context, paymentNo string) (*entities.PaymentRecord, error) {
	cacheKey := cs.cacheKeys.PaymentByNo(paymentNo)

	// Try to get from cache first
	cached, err := cs.cacheManager.GetCache().Get(ctx, cacheKey)
	if err == nil && cached != nil {
		var record entities.PaymentRecord
		if err := json.Unmarshal(cached, &record); err == nil {
			return &record, nil
		}
		// If unmarshal fails, continue to fetch from database
		logger.Warn("Failed to unmarshal cached payment record",
			logger.String("payment_no", paymentNo),
			logger.Error2("error", err))
	}

	// Fetch from database
	record, err := cs.base.GetPaymentRecord(ctx, paymentNo)
	if err != nil {
		return nil, err
	}

	// Cache the result
	cs.cachePaymentRecord(ctx, record)

	return record, nil
}

// GetPaymentRecordByOutTradeNo gets a payment record by out trade number
func (cs *CachedPaymentService) GetPaymentRecordByOutTradeNo(ctx context.Context, outTradeNo string) (*entities.PaymentRecord, error) {
	// For security reasons, we don't cache by transaction ID to avoid exposing sensitive data
	// Always fetch from database for transaction-based lookups
	return cs.base.GetPaymentRecordByOutTradeNo(ctx, outTradeNo)
}

// UpdatePaymentStatus updates payment record status and invalidates cache
func (cs *CachedPaymentService) UpdatePaymentStatus(ctx context.Context, paymentNo string, status string, transactionID string, paidAt *time.Time) error {
	err := cs.base.UpdatePaymentStatus(ctx, paymentNo, status, transactionID, paidAt)
	if err != nil {
		return err
	}

	// Invalidate cache entries for this payment
	cs.invalidatePaymentCache(ctx, paymentNo)

	return nil
}

// ProcessNotification processes payment notification with idempotency caching
func (cs *CachedPaymentService) ProcessNotification(ctx context.Context, gateway string, data map[string]any) error {
	// Generate idempotency key for this notification
	idempotencyKey := cs.generateIdempotencyKey(gateway, data)
	cacheKey := cs.cacheKeys.IdempotencyKey(gateway, idempotencyKey)

	// Check if this notification was already processed (idempotency check)
	exists, err := cs.cacheManager.GetCache().Exists(ctx, cacheKey)
	if err == nil && exists {
		logger.Info("Duplicate notification detected via idempotency cache, ignoring",
			logger.String("gateway", gateway),
			logger.String("idempotency_key", idempotencyKey))
		return nil
	}

	// Process the notification
	err = cs.base.ProcessNotification(ctx, gateway, data)
	if err != nil {
		return err
	}

	// Store idempotency key in cache to prevent duplicate processing
	// Use 30 minutes TTL for idempotency keys (replay attack protection)
	idempotencyData := map[string]any{
		"processed_at": time.Now(),
		"gateway":      gateway,
	}

	if data, err := json.Marshal(idempotencyData); err == nil {
		_ = cs.cacheManager.GetCache().Set(ctx, cacheKey, data, 30*time.Minute)
	}

	return nil
}

// GetUserPaymentRecords gets payment records for a user (no caching for user queries due to pagination)
func (cs *CachedPaymentService) GetUserPaymentRecords(ctx context.Context, userID uint, limit, offset int) ([]*entities.PaymentRecord, int64, error) {
	// Don't cache paginated user results due to complexity and freshness requirements
	return cs.base.GetUserPaymentRecords(ctx, userID, limit, offset)
}

// GetAvailablePaymentMethods gets available payment methods with caching
func (cs *CachedPaymentService) GetAvailablePaymentMethods(ctx context.Context) (map[string][]string, error) {
	cacheKey := cs.cacheKeys.PaymentMethods()

	// Try to get from cache first
	cached, err := cs.cacheManager.GetCache().Get(ctx, cacheKey)
	if err == nil && cached != nil {
		var methods map[string][]string
		if err := json.Unmarshal(cached, &methods); err == nil {
			return methods, nil
		}
	}

	// Fetch from base service
	methods, err := cs.base.GetAvailablePaymentMethods(ctx)
	if err != nil {
		return nil, err
	}

	// Cache the result for 1 hour (payment methods don't change frequently)
	if data, err := json.Marshal(methods); err == nil {
		_ = cs.cacheManager.GetCache().Set(ctx, cacheKey, data, cache.LongCacheTTL)
	}

	return methods, nil
}

// SetSubscriptionOrderService sets the subscription order service
func (cs *CachedPaymentService) SetSubscriptionOrderService(subscriptionOrderService interfaces.SubscriptionOrderServiceInterface) {
	cs.base.SetSubscriptionOrderService(subscriptionOrderService)
}

// Helper methods

// cachePaymentRecord caches a payment record with appropriate TTL
func (cs *CachedPaymentService) cachePaymentRecord(ctx context.Context, record *entities.PaymentRecord) {
	if record == nil {
		return
	}

	cacheKey := cs.cacheKeys.PaymentByNo(record.PaymentNo)

	// Create a sanitized version for caching (remove sensitive data)
	cachedRecord := &entities.PaymentRecord{
		ID:                  record.ID,
		UserID:              record.UserID,
		SubscriptionOrderID: record.SubscriptionOrderID,
		InvoiceID:           record.InvoiceID,
		PaymentNo:           record.PaymentNo,
		Gateway:             record.Gateway,
		PaymentMethod:       record.PaymentMethod,
		Amount:              record.Amount,
		Currency:            record.Currency,
		ExchangeRate:        record.ExchangeRate,
		Status:              record.Status,
		PaymentStatus:       record.PaymentStatus,
		PaymentURL:          record.PaymentURL,
		QRCodeURL:           record.QRCodeURL,
		ExpiredAt:           record.ExpiredAt,
		PaidAt:              record.PaidAt,
		RefundAmount:        record.RefundAmount,
		RefundStatus:        record.RefundStatus,
		RefundedAt:          record.RefundedAt,
		RefundReason:        record.RefundReason,
		ClientIP:            record.ClientIP,
		NotifyURL:           record.NotifyURL,
		ReturnURL:           record.ReturnURL,
		Remark:              record.Remark,
		CreatedAt:           record.CreatedAt,
		UpdatedAt:           record.UpdatedAt,
		// Deliberately exclude sensitive fields:
		// - OutTradeNo (sensitive transaction identifier)
		// - TransactionID (sensitive payment gateway identifier)
		// - GatewayResponse (may contain sensitive gateway data)
		// - Metadata (may contain sensitive user data)
		// - NotifySource, LastNotifyHash, NotifyCount, LastNotifyTime (internal security data)
	}

	if data, err := json.Marshal(cachedRecord); err == nil {
		// Use short TTL for payment records for security and freshness
		_ = cs.cacheManager.GetCache().Set(ctx, cacheKey, data, cache.DefaultCacheTTL)
	}
}

// invalidatePaymentCache invalidates all cache entries for a payment
func (cs *CachedPaymentService) invalidatePaymentCache(ctx context.Context, paymentNo string) {
	cacheKey := cs.cacheKeys.PaymentByNo(paymentNo)
	_ = cs.cacheManager.GetCache().Delete(ctx, cacheKey)
}

// generateIdempotencyKey generates a consistent key for notification idempotency
func (cs *CachedPaymentService) generateIdempotencyKey(gateway string, data map[string]any) string {
	// Create a consistent string representation of key notification data
	var keyParts []string

	// Include essential fields that should be consistent for the same notification
	if outTradeNo, ok := data["out_trade_no"].(string); ok && outTradeNo != "" {
		keyParts = append(keyParts, "out_trade_no:"+outTradeNo)
	}
	if txnID, ok := data["transaction_id"].(string); ok && txnID != "" {
		keyParts = append(keyParts, "txn_id:"+txnID)
	}
	if status, ok := data["status"].(string); ok && status != "" {
		keyParts = append(keyParts, "status:"+status)
	}
	if amount, ok := data["amount"]; ok {
		keyParts = append(keyParts, fmt.Sprintf("amount:%v", amount))
	}

	// Sort for consistency
	for i := 0; i < len(keyParts); i++ {
		for j := i + 1; j < len(keyParts); j++ {
			if keyParts[i] > keyParts[j] {
				keyParts[i], keyParts[j] = keyParts[j], keyParts[i]
			}
		}
	}

	return strings.Join(keyParts, "|")
}
