package implementations

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	couponInterfaces "linke/internal/domains/coupon/usecases/interfaces"
	invoiceInterfaces "linke/internal/domains/invoice/usecases/interfaces"
	paymentInterfaces "linke/internal/domains/payment/usecases/interfaces"
	"linke/internal/domains/subscription/entities"
	"linke/internal/domains/subscription/usecases/interfaces"
	"linke/internal/shared/cache"
	"linke/internal/shared/logger"

	"gorm.io/gorm"
)

type CachedSubscriptionOrderService struct {
	*SubscriptionOrderService
	cacheManager cache.CacheManager
	cacheKeys    *cache.AllCacheKeys
	orderCache   *cache.CacheAside[entities.SubscriptionOrder]
}

func NewSubscriptionOrderServiceWithCache(
	db *gorm.DB,
	subscriptionPlanService interfaces.SubscriptionPlanService,
	userSubscriptionService interfaces.UserSubscriptionService,
	paymentService paymentInterfaces.PaymentService,
	paymentMethodService paymentInterfaces.PaymentMethodService,
	couponService couponInterfaces.CouponService,
	invoiceService invoiceInterfaces.InvoiceService,
	cacheManager cache.CacheManager,
	cacheKeys *cache.AllCacheKeys,
) interfaces.SubscriptionOrderService {
	baseService := NewSubscriptionOrderService(
		db,
		subscriptionPlanService,
		userSubscriptionService,
		paymentService,
		paymentMethodService,
		couponService,
		invoiceService,
	)

	orderCache := cache.NewCacheAside[entities.SubscriptionOrder](
		cacheManager.GetCache(),
		cache.CachePrefixSubscription,
		func(order entities.SubscriptionOrder) string {
			return fmt.Sprintf("order:id:%d", order.ID)
		},
		cache.MediumCacheTTL, // 15 minutes TTL for orders
	)

	return &CachedSubscriptionOrderService{
		SubscriptionOrderService: baseService,
		cacheManager:             cacheManager,
		cacheKeys:                cacheKeys,
		orderCache:               orderCache,
	}
}

// CreateSubscriptionOrder creates order with cache management
func (sos *CachedSubscriptionOrderService) CreateSubscriptionOrder(ctx context.Context, req *interfaces.CreateSubscriptionOrderRequest) (*interfaces.CreateSubscriptionOrderResponse, error) {
	response, err := sos.SubscriptionOrderService.CreateSubscriptionOrder(ctx, req)
	if err != nil {
		return nil, err
	}

	// Write-through: cache the new order after successful creation
	if response != nil && response.Order != nil {
		// Convert response back to entity for caching
		order := &entities.SubscriptionOrder{
			ID:                 response.Order.ID,
			UserID:             response.Order.UserID,
			SubscriptionPlanID: response.Order.SubscriptionPlanID,
			OrderNumber:        response.Order.OrderNumber,
			OrderType:          response.Order.OrderType,
			Status:             response.Order.Status,
			Amount:             response.Order.Amount,
			Currency:           response.Order.Currency,
			SetupFee:           response.Order.SetupFee,
			DiscountAmount:     response.Order.DiscountAmount,
			TotalAmount:        response.Order.TotalAmount,
			PaymentGateway:     response.Order.PaymentGateway,
			PaymentMethod:      response.Order.PaymentMethod,
			CouponCode:         response.Order.CouponCode,
			TransactionID:      response.Order.TransactionID,
		}

		if err := sos.orderCache.Set(ctx, order); err != nil {
			logger.Error("Failed to cache new order",
				logger.Uint("order_id", order.ID),
				logger.Error2("error", err))
		}

		// Cache order by number
		orderByNumberKey := sos.cacheKeys.Subscription.OrderByNumber(order.OrderNumber)
		if data, err := json.Marshal(order); err == nil {
			_ = sos.cacheManager.GetCache().Set(ctx, orderByNumberKey, data, cache.MediumCacheTTL)
		}

		// Invalidate user order caches
		sos.invalidateUserOrderCaches(ctx, req.UserID)
	}

	return response, nil
}

// GetSubscriptionOrder gets order by ID with caching
func (sos *CachedSubscriptionOrderService) GetSubscriptionOrder(ctx context.Context, orderID uint) (*entities.SubscriptionOrder, error) {
	cacheKey := sos.cacheKeys.Subscription.OrderByID(orderID)

	order, err := sos.orderCache.Get(ctx, cacheKey, func() (*entities.SubscriptionOrder, error) {
		return sos.SubscriptionOrderService.GetSubscriptionOrder(ctx, orderID)
	})

	if err != nil {
		return nil, err
	}

	return order, nil
}

// GetSubscriptionOrderByNumber gets order by order number with caching
func (sos *CachedSubscriptionOrderService) GetSubscriptionOrderByNumber(ctx context.Context, orderNumber string) (*entities.SubscriptionOrder, error) {
	cacheKey := sos.cacheKeys.Subscription.OrderByNumber(orderNumber)

	cached, err := sos.cacheManager.GetCache().Get(ctx, cacheKey)
	if err == nil && cached != nil {
		var order entities.SubscriptionOrder
		if err := json.Unmarshal(cached, &order); err == nil {
			return &order, nil
		}
	}

	// Cache miss - fetch from database
	order, err := sos.SubscriptionOrderService.GetSubscriptionOrderByNumber(ctx, orderNumber)
	if err != nil {
		return nil, err
	}

	// Cache the result
	if order != nil {
		if data, err := json.Marshal(order); err == nil {
			_ = sos.cacheManager.GetCache().Set(ctx, cacheKey, data, cache.MediumCacheTTL)
		}
	}

	return order, nil
}

// GetSubscriptionOrders gets orders with caching for list results
func (sos *CachedSubscriptionOrderService) GetSubscriptionOrders(ctx context.Context, req *interfaces.GetSubscriptionOrdersRequest) ([]*entities.SubscriptionOrder, int64, error) {
	cacheKey := sos.buildOrderListCacheKey(req)

	// Use cache decorator for list results
	cached, err := sos.cacheManager.GetCache().Get(ctx, cacheKey)
	if err == nil && cached != nil {
		var result struct {
			Orders []*entities.SubscriptionOrder `json:"orders"`
			Total  int64                         `json:"total"`
		}
		if err := json.Unmarshal(cached, &result); err == nil {
			return result.Orders, result.Total, nil
		}
	}

	// Cache miss - fetch from database
	orders, total, err := sos.SubscriptionOrderService.GetSubscriptionOrders(ctx, req)
	if err != nil {
		return nil, 0, err
	}

	// Cache the result
	result := struct {
		Orders []*entities.SubscriptionOrder `json:"orders"`
		Total  int64                         `json:"total"`
	}{
		Orders: orders,
		Total:  total,
	}

	if data, err := json.Marshal(result); err == nil {
		_ = sos.cacheManager.GetCache().Set(ctx, cacheKey, data, cache.MediumCacheTTL)
	}

	return orders, total, nil
}

// GetUserSubscriptionOrders gets user orders with caching
func (sos *CachedSubscriptionOrderService) GetUserSubscriptionOrders(ctx context.Context, userID uint, limit, offset int) ([]*entities.SubscriptionOrder, int64, error) {
	cacheKey := fmt.Sprintf("%s:limit:%d:offset:%d", sos.cacheKeys.Subscription.UserOrders(userID), limit, offset)

	cached, err := sos.cacheManager.GetCache().Get(ctx, cacheKey)
	if err == nil && cached != nil {
		var result struct {
			Orders []*entities.SubscriptionOrder `json:"orders"`
			Total  int64                         `json:"total"`
		}
		if err := json.Unmarshal(cached, &result); err == nil {
			return result.Orders, result.Total, nil
		}
	}

	// Cache miss - fetch from database
	orders, total, err := sos.SubscriptionOrderService.GetUserSubscriptionOrders(ctx, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	// Cache the result
	result := struct {
		Orders []*entities.SubscriptionOrder `json:"orders"`
		Total  int64                         `json:"total"`
	}{
		Orders: orders,
		Total:  total,
	}

	if data, err := json.Marshal(result); err == nil {
		_ = sos.cacheManager.GetCache().Set(ctx, cacheKey, data, cache.MediumCacheTTL)
	}

	return orders, total, nil
}

// Note: UpdateSubscriptionOrder is not implemented in the base service

// ProcessOrderPaymentSuccess processes payment success with cache invalidation
func (sos *CachedSubscriptionOrderService) ProcessOrderPaymentSuccess(ctx context.Context, orderID uint) error {
	// Get the order before processing to know userID for cache invalidation
	existingOrder, err := sos.GetSubscriptionOrder(ctx, orderID)
	if err != nil {
		return err
	}

	err = sos.SubscriptionOrderService.ProcessOrderPaymentSuccess(ctx, orderID)
	if err != nil {
		return err
	}

	// Invalidate caches after successful payment processing
	sos.invalidateOrderCaches(ctx, orderID, existingOrder.OrderNumber, existingOrder.UserID)

	return nil
}

// Note: ProcessOrderPaymentFailure is not implemented in the base service

// CancelSubscriptionOrder cancels order with cache invalidation
func (sos *CachedSubscriptionOrderService) CancelSubscriptionOrder(ctx context.Context, orderID uint, reason string) error {
	// Get the order before cancellation to know userID for cache invalidation
	existingOrder, err := sos.GetSubscriptionOrder(ctx, orderID)
	if err != nil {
		return err
	}

	err = sos.SubscriptionOrderService.CancelSubscriptionOrder(ctx, orderID, reason)
	if err != nil {
		return err
	}

	// Invalidate caches after cancellation
	sos.invalidateOrderCaches(ctx, orderID, existingOrder.OrderNumber, existingOrder.UserID)

	return nil
}

// Note: ExpireSubscriptionOrder, DeleteSubscriptionOrder, and ProcessExpiredOrders
// are not implemented in the base service

// GetOrderStatistics gets statistics with caching
func (sos *CachedSubscriptionOrderService) GetOrderStatistics(ctx context.Context, fromDate, toDate time.Time) (map[string]any, error) {
	cacheKey := fmt.Sprintf("stats:order:from:%s:to:%s", fromDate.Format("2006-01-02"), toDate.Format("2006-01-02"))

	cached, err := sos.cacheManager.GetCache().Get(ctx, cacheKey)
	if err == nil && cached != nil {
		var stats map[string]any
		if err := json.Unmarshal(cached, &stats); err == nil {
			return stats, nil
		}
	}

	stats, err := sos.SubscriptionOrderService.GetOrderStatistics(ctx, fromDate, toDate)
	if err != nil {
		return nil, err
	}

	// Cache statistics with medium TTL
	if data, err := json.Marshal(stats); err == nil {
		_ = sos.cacheManager.GetCache().Set(ctx, cacheKey, data, cache.MediumCacheTTL)
	}

	return stats, nil
}

// Cache invalidation helper methods

func (sos *CachedSubscriptionOrderService) invalidateOrderCaches(ctx context.Context, orderID uint, orderNumber string, userID uint) {
	// Invalidate specific order caches
	orderByIDKey := sos.cacheKeys.Subscription.OrderByID(orderID)
	if err := sos.orderCache.Invalidate(ctx, orderByIDKey); err != nil {
		logger.Error("Failed to invalidate order by ID cache",
			logger.Uint("order_id", orderID),
			logger.Error2("error", err))
	}

	orderByNumberKey := sos.cacheKeys.Subscription.OrderByNumber(orderNumber)
	if err := sos.cacheManager.GetCache().Delete(ctx, orderByNumberKey); err != nil {
		logger.Error("Failed to invalidate order by number cache",
			logger.String("order_number", orderNumber),
			logger.Error2("error", err))
	}

	// Invalidate user order caches
	sos.invalidateUserOrderCaches(ctx, userID)

	// Invalidate order list caches
	sos.invalidateOrderListCaches(ctx)
}

func (sos *CachedSubscriptionOrderService) invalidateUserOrderCaches(ctx context.Context, userID uint) {
	// Invalidate user-specific order caches
	userOrdersKey := sos.cacheKeys.Subscription.UserOrders(userID)
	if err := sos.cacheManager.GetCache().Delete(ctx, userOrdersKey); err != nil {
		logger.Error("Failed to invalidate user orders cache",
			logger.Uint("user_id", userID),
			logger.Error2("error", err))
	}

	// Invalidate list caches for this user
	patterns := []string{
		fmt.Sprintf("%s:*", userOrdersKey),
		fmt.Sprintf("list:*user:%d*", userID),
	}

	for _, pattern := range patterns {
		if err := sos.cacheManager.GetCache().DeleteByPattern(ctx, pattern); err != nil {
			logger.Error("Failed to invalidate user order list cache",
				logger.String("pattern", pattern),
				logger.Error2("error", err))
		}
	}
}

func (sos *CachedSubscriptionOrderService) invalidateOrderListCaches(ctx context.Context) {
	// Invalidate order list caches
	patterns := []string{
		cache.CachePrefixSubscription + "list:order:*",
		cache.CachePrefixSubscription + "stats:*",
	}

	for _, pattern := range patterns {
		if err := sos.cacheManager.GetCache().DeleteByPattern(ctx, pattern); err != nil {
			logger.Error("Failed to invalidate order list cache pattern",
				logger.String("pattern", pattern),
				logger.Error2("error", err))
		}
	}
}

func (sos *CachedSubscriptionOrderService) invalidateAllOrderCaches(ctx context.Context) {
	// Invalidate all order caches - use sparingly
	patterns := []string{
		cache.CachePrefixSubscription + "order:*",
		cache.CachePrefixSubscription + "list:order:*",
		cache.CachePrefixSubscription + "stats:*",
	}

	for _, pattern := range patterns {
		if err := sos.cacheManager.GetCache().DeleteByPattern(ctx, pattern); err != nil {
			logger.Error("Failed to invalidate all order caches",
				logger.String("pattern", pattern),
				logger.Error2("error", err))
		}
	}
}

// Cache key building helper methods

func (sos *CachedSubscriptionOrderService) buildOrderListCacheKey(req *interfaces.GetSubscriptionOrdersRequest) string {
	var keyParts []string
	keyParts = append(keyParts, "list", "order")

	if req.UserID > 0 {
		keyParts = append(keyParts, "user", fmt.Sprintf("%d", req.UserID))
	}
	if req.Status != "" {
		keyParts = append(keyParts, "status", req.Status)
	}
	if req.OrderType != "" {
		keyParts = append(keyParts, "type", req.OrderType)
	}
	if req.DateFrom != "" {
		keyParts = append(keyParts, "from", req.DateFrom)
	}
	if req.DateTo != "" {
		keyParts = append(keyParts, "to", req.DateTo)
	}

	keyParts = append(keyParts, fmt.Sprintf("limit:%d", req.Limit))
	keyParts = append(keyParts, fmt.Sprintf("offset:%d", req.Offset))

	return cache.CachePrefixSubscription + strings.Join(keyParts, ":")
}

// QuickPurchase delegates to base service (no caching needed for payment creation)
func (sos *CachedSubscriptionOrderService) QuickPurchase(ctx context.Context, req *interfaces.QuickPurchaseRequest) (*interfaces.QuickPurchaseResponse, error) {
	// Quick purchase doesn't need caching since it's just creating a payment
	// The actual order/invoice creation happens asynchronously after payment success
	return sos.SubscriptionOrderService.QuickPurchase(ctx, req)
}
