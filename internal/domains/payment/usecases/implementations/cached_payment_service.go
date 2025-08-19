package implementations

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"linke/internal/domains/payment/dto"
	"linke/internal/domains/payment/entities"
	"linke/internal/domains/payment/usecases/interfaces"
	"linke/internal/shared/cache"
	"linke/internal/shared/logger"
)

// CachedPaymentService wraps PaymentService with minimal caching for security and performance
// Only caches: idempotency keys (security) and payment methods (static config data)
// NO payment record caching - all payment queries are real-time for transactional consistency
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
func (cs *CachedPaymentService) CreatePaymentOrder(ctx context.Context, req *dto.CreatePaymentOrderRequest) (*entities.PaymentRecord, error) {
	// No caching - all payment operations are transactional and require real-time consistency
	return cs.base.CreatePaymentOrder(ctx, req)
}

// GetPaymentRecord gets a payment record by payment number (real-time, no caching)
func (cs *CachedPaymentService) GetPaymentRecord(ctx context.Context, paymentNo string) (*entities.PaymentRecord, error) {
	// No caching - payment records must be real-time for transactional consistency
	return cs.base.GetPaymentRecord(ctx, paymentNo)
}

// GetPaymentRecordByOutTradeNo gets a payment record by out trade number (real-time, no caching)
func (cs *CachedPaymentService) GetPaymentRecordByOutTradeNo(ctx context.Context, outTradeNo string) (*entities.PaymentRecord, error) {
	// No caching - payment records must be real-time for transactional consistency
	return cs.base.GetPaymentRecordByOutTradeNo(ctx, outTradeNo)
}

// UpdatePaymentStatus updates payment record status (real-time, no caching)
func (cs *CachedPaymentService) UpdatePaymentStatus(ctx context.Context, paymentNo, status, transactionID string, paidAt *time.Time) error {
	// No caching involved - direct database update for transactional consistency
	return cs.base.UpdatePaymentStatus(ctx, paymentNo, status, transactionID, paidAt)
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

// GetUserPaymentRecords gets payment records for a user (real-time, no caching)
func (cs *CachedPaymentService) GetUserPaymentRecords(ctx context.Context, userID uint, limit, offset int) ([]*entities.PaymentRecord, int64, error) {
	// No caching - payment records must be real-time for transactional consistency
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
