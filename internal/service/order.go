package service

import (
	"context"
	"fmt"
	"time"

	"linke/internal/logger"
	"linke/internal/model"

	"gorm.io/gorm"
)

type OrderService struct {
	db                      *gorm.DB
	subscriptionPlanService *SubscriptionPlanService
	couponService           *CouponService
}

func NewOrderService(db *gorm.DB, subscriptionPlanService *SubscriptionPlanService, couponService *CouponService) *OrderService {
	return &OrderService{
		db:                      db,
		subscriptionPlanService: subscriptionPlanService,
		couponService:           couponService,
	}
}

// CreateOrderRequest represents the request to create an order
type CreateOrderRequest struct {
	UserID      uint   `json:"user_id" binding:"required"`
	PlanID      uint   `json:"plan_id" binding:"required"`
	OrderType   string `json:"order_type" binding:"required,oneof=new renewal upgrade downgrade"`
	CouponCode  string `json:"coupon_code,omitempty"`
	Notes       string `json:"notes,omitempty"`
	
	// Service Configuration Overrides (optional)
	BillingCycle    string `json:"billing_cycle,omitempty"`
	BillingInterval *int   `json:"billing_interval,omitempty"`
	ServicePeriod   *int   `json:"service_period,omitempty"`
	
	// Time Configuration (optional)
	ServiceStartDate *string `json:"service_start_date,omitempty"` // ISO 8601 format
}

// CreateOrder creates a new order
func (os *OrderService) CreateOrder(ctx context.Context, req *CreateOrderRequest) (*model.Order, error) {
	// Get subscription plan
	plan, err := os.subscriptionPlanService.GetSubscriptionPlan(ctx, req.PlanID)
	if err != nil {
		return nil, fmt.Errorf("invalid subscription plan: %w", err)
	}

	if !plan.IsAvailableForPurchase() {
		return nil, fmt.Errorf("subscription plan is not available for purchase")
	}

	// Start database transaction
	tx := os.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Generate order number
	orderNumber := os.generateOrderNumber()

	// Use plan configuration or request overrides
	billingCycle := plan.BillingCycle
	if req.BillingCycle != "" {
		billingCycle = req.BillingCycle
	}

	billingInterval := plan.BillingInterval
	if req.BillingInterval != nil {
		billingInterval = *req.BillingInterval
	}

	servicePeriod := os.calculateServicePeriod(billingCycle, billingInterval)
	if req.ServicePeriod != nil {
		servicePeriod = *req.ServicePeriod
	}

	// Calculate service time
	var serviceStartDate, serviceEndDate *time.Time
	now := time.Now()
	
	if req.ServiceStartDate != nil && *req.ServiceStartDate != "" {
		if parsed, parseErr := time.Parse(time.RFC3339, *req.ServiceStartDate); parseErr == nil {
			serviceStartDate = &parsed
		} else {
			serviceStartDate = &now
		}
	} else {
		serviceStartDate = &now
	}

	if billingCycle != model.BillingCycleLifetime {
		endDate := serviceStartDate.AddDate(0, servicePeriod, 0)
		serviceEndDate = &endDate
	}

	// Calculate amounts
	baseAmount := plan.Price
	setupFee := plan.SetupFee
	var discountAmount, couponDiscountAmount float64 = 0.0, 0.0

	// Apply coupon discount if provided
	if req.CouponCode != "" {
		validateReq := &ValidateCouponRequest{
			Code:        req.CouponCode,
			UserID:      uint64(req.UserID),
			OrderAmount: baseAmount + setupFee,
			PlanID:      uint64(req.PlanID),
			Currency:    plan.Currency,
		}

		validateResp, err := os.couponService.ValidateCoupon(ctx, validateReq)
		if err != nil {
			logger.Error("Failed to validate coupon", logger.Error2("error", err), logger.String("coupon_code", req.CouponCode))
			return nil, fmt.Errorf("failed to validate coupon: %w", err)
		}

		if !validateResp.Valid {
			return nil, fmt.Errorf("coupon validation failed: %s", validateResp.Message)
		}

		couponDiscountAmount = validateResp.DiscountAmount
		discountAmount = couponDiscountAmount
	}

	totalAmount := baseAmount + setupFee - discountAmount

	// SECURITY: Price protection - validate final amount is reasonable
	if totalAmount < 0 {
		return nil, fmt.Errorf("invalid total amount: negative value not allowed")
	}

	// Prevent excessive discounts (more than 95% off)
	originalTotal := baseAmount + setupFee
	if originalTotal > 0 && discountAmount > (originalTotal*0.95) {
		return nil, fmt.Errorf("discount amount %.2f exceeds maximum allowed (95%% of original price)", discountAmount)
	}

	// Validate amounts are within reasonable limits
	maxAllowedAmount := 10000.0 // $10,000 max
	if totalAmount > maxAllowedAmount {
		return nil, fmt.Errorf("total amount %.2f exceeds maximum allowed amount %.2f", totalAmount, maxAllowedAmount)
	}

	// Create order
	order := &model.Order{
		UserID:           req.UserID,
		PlanID:           req.PlanID,
		OrderNumber:      orderNumber,
		OrderType:        req.OrderType,
		Status:           model.NewOrderStatusDraft,
		BillingCycle:     billingCycle,
		BillingInterval:  billingInterval,
		ServicePeriod:    servicePeriod,
		BaseAmount:       baseAmount,
		DiscountAmount:   discountAmount,
		SetupFee:         setupFee,
		TotalAmount:      totalAmount,
		Currency:         plan.Currency,
		CouponCode:       req.CouponCode,
		CouponDiscount:   couponDiscountAmount,
		ServiceStartDate: serviceStartDate,
		ServiceEndDate:   serviceEndDate,
		Notes:            req.Notes,
	}

	// Save order to database
	if err := tx.Create(order).Error; err != nil {
		tx.Rollback()
		logger.Error("Failed to create order", logger.Error2("error", err))
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	// Record coupon usage if coupon was applied
	if req.CouponCode != "" && couponDiscountAmount > 0 {
		if coupon, getCouponErr := os.couponService.GetCouponByCode(ctx, req.CouponCode); getCouponErr == nil {
			if useCouponErr := os.couponService.UseCoupon(ctx, coupon.ID, uint64(req.UserID), uint64(order.ID), couponDiscountAmount, originalTotal, plan.Currency); useCouponErr != nil {
				logger.Error("Failed to record coupon usage",
					logger.Error2("error", useCouponErr),
					logger.Any("coupon_id", coupon.ID),
					logger.Uint("order_id", order.ID))
				// Don't fail the order creation, just log the error
			}
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		logger.Error("Failed to commit order creation transaction", logger.Error2("error", err))
		return nil, fmt.Errorf("failed to commit order creation transaction: %w", err)
	}

	logger.Info("Order created successfully",
		logger.Uint("order_id", order.ID),
		logger.String("order_number", orderNumber),
		logger.Uint("user_id", req.UserID),
		logger.Uint("plan_id", req.PlanID))

	return order, nil
}

// ConfirmOrder confirms an order and generates an invoice
func (os *OrderService) ConfirmOrder(ctx context.Context, orderID uint) (*model.Order, error) {
	// Start transaction
	tx := os.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Get order with row-level lock
	var order model.Order
	if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&order, orderID).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("order not found")
		}
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	// SECURITY: Validate order can be confirmed
	if !order.CanBeConfirmed() {
		tx.Rollback()
		return nil, fmt.Errorf("order cannot be confirmed in status: %s", order.Status)
	}

	// Update order status
	now := time.Now()
	if err := tx.Model(&order).Updates(map[string]interface{}{
		"status":       model.NewOrderStatusConfirmed,
		"confirmed_at": now,
		"updated_at":   now,
	}).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to update order status: %w", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit order confirmation: %w", err)
	}

	// Reload order with updated data
	if err := os.db.WithContext(ctx).First(&order, orderID).Error; err != nil {
		return nil, fmt.Errorf("failed to reload order: %w", err)
	}

	logger.Info("Order confirmed successfully", logger.Uint("order_id", orderID))
	return &order, nil
}

// CancelOrder cancels an order
func (os *OrderService) CancelOrder(ctx context.Context, orderID uint, reason string) error {
	// Start transaction
	tx := os.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Get order with row-level lock
	var order model.Order
	if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&order, orderID).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("order not found")
		}
		return fmt.Errorf("failed to get order: %w", err)
	}

	// SECURITY: Validate order can be cancelled
	if !order.CanBeCancelled() {
		tx.Rollback()
		return fmt.Errorf("order cannot be cancelled in status: %s", order.Status)
	}

	// Update order status
	now := time.Now()
	updateData := map[string]interface{}{
		"status":       model.NewOrderStatusCancelled,
		"cancelled_at": now,
		"updated_at":   now,
	}

	// Add cancellation reason to notes
	if reason != "" {
		cancellationNote := fmt.Sprintf("[%s] Order cancelled: %s", now.Format("2006-01-02 15:04:05"), reason)
		if order.Notes != "" {
			updateData["notes"] = order.Notes + "\n" + cancellationNote
		} else {
			updateData["notes"] = cancellationNote
		}
	}

	if err := tx.Model(&order).Updates(updateData).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update order status: %w", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit order cancellation: %w", err)
	}

	logger.Info("Order cancelled successfully", logger.Uint("order_id", orderID), logger.String("reason", reason))
	return nil
}

// FulfillOrder marks an order as fulfilled
func (os *OrderService) FulfillOrder(ctx context.Context, orderID uint) error {
	// Start transaction
	tx := os.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Get order with row-level lock
	var order model.Order
	if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&order, orderID).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("order not found")
		}
		return fmt.Errorf("failed to get order: %w", err)
	}

	// SECURITY: Validate order can be fulfilled
	if !order.CanBeFulfilled() {
		tx.Rollback()
		return fmt.Errorf("order cannot be fulfilled in status: %s", order.Status)
	}

	// Update order status
	now := time.Now()
	if err := tx.Model(&order).Updates(map[string]interface{}{
		"status":       model.NewOrderStatusFulfilled,
		"fulfilled_at": now,
		"updated_at":   now,
	}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update order status: %w", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit order fulfillment: %w", err)
	}

	logger.Info("Order fulfilled successfully", logger.Uint("order_id", orderID))
	return nil
}

// GetOrder gets an order by ID
func (os *OrderService) GetOrder(ctx context.Context, orderID uint) (*model.Order, error) {
	var order model.Order
	if err := os.db.WithContext(ctx).First(&order, orderID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("order not found")
		}
		logger.Error("Failed to get order", logger.Error2("error", err), logger.Uint("order_id", orderID))
		return nil, fmt.Errorf("failed to get order: %w", err)
	}
	return &order, nil
}

// GetOrderByNumber gets an order by order number
func (os *OrderService) GetOrderByNumber(ctx context.Context, orderNumber string) (*model.Order, error) {
	var order model.Order
	if err := os.db.WithContext(ctx).Where("order_number = ?", orderNumber).First(&order).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("order not found")
		}
		logger.Error("Failed to get order by number", logger.Error2("error", err), logger.String("order_number", orderNumber))
		return nil, fmt.Errorf("failed to get order: %w", err)
	}
	return &order, nil
}

// OrderFilters represents filters for order listing
type OrderFilters struct {
	UserID    *uint  `form:"user_id"`
	PlanID    *uint  `form:"plan_id"`
	Status    string `form:"status"`
	OrderType string `form:"order_type"`
	Currency  string `form:"currency"`
	Search    string `form:"search"`
	StartDate string `form:"start_date"`
	EndDate   string `form:"end_date"`
	SortBy    string `form:"sort_by"`
	SortOrder string `form:"sort_order"`
	Limit     int    `form:"limit"`
	Offset    int    `form:"offset"`
}

// ListOrders lists orders with filtering and pagination
func (os *OrderService) ListOrders(ctx context.Context, filters *OrderFilters) ([]*model.Order, int64, error) {
	query := os.db.WithContext(ctx).Model(&model.Order{})

	// Apply filters
	if filters.UserID != nil {
		query = query.Where("user_id = ?", *filters.UserID)
	}

	if filters.PlanID != nil {
		query = query.Where("plan_id = ?", *filters.PlanID)
	}

	if filters.Status != "" {
		query = query.Where("status = ?", filters.Status)
	}

	if filters.OrderType != "" {
		query = query.Where("order_type = ?", filters.OrderType)
	}

	if filters.Currency != "" {
		query = query.Where("currency = ?", filters.Currency)
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
			"order_number LIKE ? OR notes LIKE ?",
			searchPattern, searchPattern,
		)
	}

	// Get total count
	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count orders: %w", err)
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

	// Validate sort fields
	validSortFields := map[string]bool{
		"created_at":    true,
		"updated_at":    true,
		"confirmed_at":  true,
		"cancelled_at":  true,
		"fulfilled_at":  true,
		"total_amount":  true,
		"status":        true,
		"order_type":    true,
		"order_number":  true,
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

	var orders []*model.Order
	if err := query.Find(&orders).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get orders: %w", err)
	}

	return orders, totalCount, nil
}

// GetOrderWithRelations gets an order with related data
func (os *OrderService) GetOrderWithRelations(ctx context.Context, orderID uint) (*model.Order, error) {
	var order model.Order
	if err := os.db.WithContext(ctx).
		Preload("User").
		Preload("Plan").
		First(&order, orderID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("order not found")
		}
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	return &order, nil
}

// generateOrderNumber generates a unique order number
func (os *OrderService) generateOrderNumber() string {
	// Generate format: ORD + YYYYMMDDHHMMSS + 3-digit random
	now := time.Now()
	timestamp := now.Format("20060102150405")
	random := now.Nanosecond() % 1000
	return fmt.Sprintf("ORD%s%03d", timestamp, random)
}

// calculateServicePeriod calculates service period in months based on billing cycle and interval
func (os *OrderService) calculateServicePeriod(billingCycle string, billingInterval int) int {
	switch billingCycle {
	case model.BillingCycleMonthly:
		return billingInterval
	case model.BillingCycleYearly:
		return billingInterval * 12
	case model.BillingCycleLifetime:
		return 0 // Lifetime has no end date
	default:
		return billingInterval // Default to monthly
	}
}