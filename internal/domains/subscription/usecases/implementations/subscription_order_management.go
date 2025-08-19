package implementations

import (
	"context"
	"fmt"
	"time"

	invoiceDto "linke/internal/domains/invoice/dto"
	invoiceEntities "linke/internal/domains/invoice/entities"
	paymentDto "linke/internal/domains/payment/dto"
	paymentEntities "linke/internal/domains/payment/entities"
	"linke/internal/domains/subscription/dto"
	"linke/internal/domains/subscription/entities"
	"linke/internal/shared/logger"

	"gorm.io/gorm"
)

// SubscriptionOrderManagementService handles all order retrieval and management operations
type SubscriptionOrderManagementService struct {
	*SubscriptionOrderServiceBase
}

// NewSubscriptionOrderManagementService creates a new order management service
func NewSubscriptionOrderManagementService(base *SubscriptionOrderServiceBase) *SubscriptionOrderManagementService {
	return &SubscriptionOrderManagementService{
		SubscriptionOrderServiceBase: base,
	}
}

// GetSubscriptionOrderSummary aggregates order + latest payment + latest invoice
func (soms *SubscriptionOrderManagementService) GetSubscriptionOrderSummary(ctx context.Context, orderID uint) (map[string]any, error) {
	// Load order
	var order entities.SubscriptionOrder
	if err := soms.db.WithContext(ctx).First(&order, orderID).Error; err != nil {
		return nil, fmt.Errorf("subscription order not found")
	}

	// Latest payment by paid_at then created_at
	var payment paymentEntities.PaymentRecord
	_ = soms.db.WithContext(ctx).
		Where("subscription_order_id = ?", orderID).
		Order("paid_at DESC, created_at DESC").
		Limit(1).
		Find(&payment).Error

	// Latest invoice by issued_at then created_at
	var invoice invoiceEntities.Invoice
	_ = soms.db.WithContext(ctx).
		Where("subscription_order_id = ?", orderID).
		Order("issued_at DESC, created_at DESC").
		Limit(1).
		Find(&invoice).Error

	result := map[string]any{
		"order":   order.ToResponse(),
		"payment": nil,
		"invoice": nil,
	}
	if payment.ID != 0 {
		result["payment"] = paymentDto.ToPaymentRecordSecureResponse(&payment)
	}
	if invoice.ID != 0 {
		result["invoice"] = invoiceDto.ToResponse(&invoice)
	}
	return result, nil
}

// GetSubscriptionOrder gets a subscription order by ID
func (soms *SubscriptionOrderManagementService) GetSubscriptionOrder(ctx context.Context, orderID uint) (*entities.SubscriptionOrder, error) {
	var order entities.SubscriptionOrder
	if err := soms.db.WithContext(ctx).First(&order, orderID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("subscription order not found")
		}
		logger.Error("Failed to get subscription order", logger.Uint("orderID", uint(orderID)))
		return nil, fmt.Errorf("failed to get subscription order: %w", err)
	}
	return &order, nil
}

// GetSubscriptionOrderByNumber gets a subscription order by order number
func (soms *SubscriptionOrderManagementService) GetSubscriptionOrderByNumber(ctx context.Context, orderNumber string) (*entities.SubscriptionOrder, error) {
	var order entities.SubscriptionOrder
	if err := soms.db.WithContext(ctx).Where("order_number = ?", orderNumber).First(&order).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("subscription order not found")
		}
		logger.Error("Failed to get subscription order by number", logger.String("order_number", orderNumber), logger.ErrorField(err))
		return nil, fmt.Errorf("failed to get subscription order: %w", err)
	}
	return &order, nil
}

// GetUserSubscriptionOrders gets subscription orders for a user with pagination
func (soms *SubscriptionOrderManagementService) GetUserSubscriptionOrders(ctx context.Context, userID uint, limit, offset int) ([]*entities.SubscriptionOrder, int64, error) {
	var orders []*entities.SubscriptionOrder
	var totalCount int64

	// Get total count
	if err := soms.db.WithContext(ctx).Model(&entities.SubscriptionOrder{}).
		Where("user_id = ?", userID).
		Count(&totalCount).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count subscription orders: %w", err)
	}

	// Get orders with pagination
	query := soms.db.WithContext(ctx).Where("user_id = ?", userID).
		Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&orders).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get subscription orders: %w", err)
	}

	return orders, totalCount, nil
}

// GetOrdersWithFiltering gets orders with advanced filtering and search
func (soms *SubscriptionOrderManagementService) GetOrdersWithFiltering(ctx context.Context, req *GetOrdersRequest) ([]*entities.SubscriptionOrder, int64, error) {
	query := soms.db.WithContext(ctx).Model(&entities.SubscriptionOrder{})

	// Apply filters
	if req.UserID != nil {
		query = query.Where("user_id = ?", *req.UserID)
	}

	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}

	if req.OrderType != "" {
		query = query.Where("order_type = ?", req.OrderType)
	}

	if req.PaymentMethod != "" {
		query = query.Where("payment_method = ?", req.PaymentMethod)
	}

	if req.PaymentGateway != "" {
		query = query.Where("payment_gateway = ?", req.PaymentGateway)
	}

	if req.MinAmount > 0 {
		query = query.Where("total_amount >= ?", req.MinAmount)
	}

	if req.MaxAmount > 0 {
		query = query.Where("total_amount <= ?", req.MaxAmount)
	}

	if req.CouponCode != "" {
		query = query.Where("coupon_code = ?", req.CouponCode)
	}

	// Date range filtering
	if req.StartDate != "" {
		if startDate, err := time.Parse("2006-01-02", req.StartDate); err == nil {
			query = query.Where("created_at >= ?", startDate)
		}
	}

	if req.EndDate != "" {
		if endDate, err := time.Parse("2006-01-02", req.EndDate); err == nil {
			// Add 24 hours to include the entire end date
			endDate = endDate.Add(24 * time.Hour)
			query = query.Where("created_at < ?", endDate)
		}
	}

	// Search functionality
	if req.Search != "" {
		searchPattern := "%" + req.Search + "%"
		query = query.Where(
			"order_number LIKE ? OR transaction_id LIKE ? OR notes LIKE ?",
			searchPattern, searchPattern, searchPattern,
		)
	}

	// Get total count before pagination
	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count orders: %w", err)
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

	// Validate sort fields
	validSortFields := map[string]bool{
		"created_at":   true,
		"updated_at":   true,
		"paid_at":      true,
		"amount":       true,
		"total_amount": true,
		"status":       true,
		"order_type":   true,
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

	// Load relations if requested
	if req.IncludeRelations {
		query = query.Preload("User").Preload("SubscriptionPlan").Preload("UserSubscription")
	}

	var orders []*entities.SubscriptionOrder
	if err := query.Find(&orders).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get orders: %w", err)
	}

	return orders, totalCount, nil
}

// GetOrderWithRelations gets an order with all related data
func (soms *SubscriptionOrderManagementService) GetOrderWithRelations(ctx context.Context, orderID uint) (*entities.SubscriptionOrder, error) {
	var order entities.SubscriptionOrder
	if err := soms.db.WithContext(ctx).
		Preload("User").
		Preload("SubscriptionPlan").
		Preload("UserSubscription").
		First(&order, orderID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("subscription order not found")
		}
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	return &order, nil
}

// GetSubscriptionOrders gets subscription orders with filtering
func (soms *SubscriptionOrderManagementService) GetSubscriptionOrders(ctx context.Context, req *dto.GetSubscriptionOrdersRequest) ([]*entities.SubscriptionOrder, int64, error) {
	query := soms.db.WithContext(ctx).Model(&entities.SubscriptionOrder{})

	// Apply filters
	if req.UserID != 0 {
		query = query.Where("user_id = ?", req.UserID)
	}

	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}

	if req.OrderType != "" {
		query = query.Where("order_type = ?", req.OrderType)
	}

	// Date range filtering
	if req.DateFrom != "" {
		if startDate, err := time.Parse("2006-01-02", req.DateFrom); err == nil {
			query = query.Where("created_at >= ?", startDate)
		}
	}

	if req.DateTo != "" {
		if endDate, err := time.Parse("2006-01-02", req.DateTo); err == nil {
			// Add 24 hours to include the entire end date
			endDate = endDate.Add(24 * time.Hour)
			query = query.Where("created_at < ?", endDate)
		}
	}

	// Get total count
	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count subscription orders: %w", err)
	}

	// Apply pagination
	if req.Limit > 0 {
		query = query.Limit(req.Limit)
	}

	if req.Offset > 0 {
		query = query.Offset(req.Offset)
	}

	// Apply ordering
	query = query.Order("created_at DESC")

	// Get orders
	var orders []*entities.SubscriptionOrder
	if err := query.Find(&orders).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get subscription orders: %w", err)
	}

	return orders, totalCount, nil
}
