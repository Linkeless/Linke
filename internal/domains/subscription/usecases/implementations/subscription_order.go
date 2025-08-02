package implementations

import (
	"context"
	"fmt"
	"time"

	couponInterfaces "linke/internal/domains/coupon/usecases/interfaces"
	paymentInterfaces "linke/internal/domains/payment/usecases/interfaces"
	"linke/internal/domains/subscription/entities"
	"linke/internal/domains/subscription/usecases/interfaces"
	"linke/internal/shared/logger"

	"gorm.io/gorm"
)

type SubscriptionOrderService struct {
	db                      *gorm.DB
	subscriptionPlanService interfaces.SubscriptionPlanService
	userSubscriptionService interfaces.UserSubscriptionService
	paymentService          paymentInterfaces.PaymentService
	couponService           couponInterfaces.CouponService
}

func NewSubscriptionOrderService(db *gorm.DB, subscriptionPlanService interfaces.SubscriptionPlanService, userSubscriptionService interfaces.UserSubscriptionService, paymentService paymentInterfaces.PaymentService, couponService couponInterfaces.CouponService) *SubscriptionOrderService {
	return &SubscriptionOrderService{
		db:                      db,
		subscriptionPlanService: subscriptionPlanService,
		userSubscriptionService: userSubscriptionService,
		paymentService:          paymentService,
		couponService:           couponService,
	}
}


// CreateSubscriptionOrder creates a new subscription order with payment
func (sos *SubscriptionOrderService) CreateSubscriptionOrder(ctx context.Context, req *interfaces.CreateSubscriptionOrderRequest) (*interfaces.CreateSubscriptionOrderResponse, error) {
	// SECURITY: Check for duplicate pending orders to prevent order spam
	if err := sos.checkDuplicateOrders(ctx, req.UserID, req.SubscriptionPlanID, req.OrderType); err != nil {
		return nil, fmt.Errorf("duplicate order check failed: %w", err)
	}

	// Get subscription plan
	plan, err := sos.subscriptionPlanService.GetSubscriptionPlan(ctx, req.SubscriptionPlanID)
	if err != nil {
		return nil, fmt.Errorf("invalid subscription plan: %w", err)
	}

	if !plan.IsAvailableForPurchase() {
		return nil, fmt.Errorf("subscription plan is not available for purchase")
	}

	// Start database transaction
	tx := sos.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Generate order number
	orderNumber := sos.generateOrderNumber()

	// Calculate billing period
	var billingPeriodStart, billingPeriodEnd *time.Time
	now := time.Now()

	if plan.BillingCycle != entities.BillingCycleLifetime {
		periodStart := now
		billingPeriodStart = &periodStart

		var periodEnd time.Time
		switch plan.BillingCycle {
		case entities.BillingCycleMonthly:
			periodEnd = periodStart.AddDate(0, plan.BillingInterval, 0)
		case entities.BillingCycleYearly:
			periodEnd = periodStart.AddDate(plan.BillingInterval, 0, 0)
		}
		billingPeriodEnd = &periodEnd
	}

	// Calculate amounts
	amount := plan.Price
	setupFee := plan.SetupFee
	var discountAmount, couponDiscountAmount float64 = 0.0, 0.0
	var appliedCouponID uint64 // Store only the ID to avoid cross-domain dependency

	// Apply coupon discount if provided
	if req.CouponCode != "" {
		validateReq := &couponInterfaces.ValidateCouponRequest{
			Code:        req.CouponCode,
			UserID:      uint64(req.UserID),
			OrderAmount: amount + setupFee,
			PlanID:      uint64(req.SubscriptionPlanID),
			Currency:    plan.Currency,
		}

		validateResp, err := sos.couponService.ValidateCoupon(ctx, validateReq)
		if err != nil {
			logger.Error("Failed to validate coupon", logger.Error2("error", err), logger.String("coupon_code", req.CouponCode))
			return nil, fmt.Errorf("failed to validate coupon: %w", err)
		}

		if !validateResp.Valid {
			return nil, fmt.Errorf("coupon validation failed: %s", validateResp.Message)
		}

		couponDiscountAmount = validateResp.DiscountAmount
		discountAmount = couponDiscountAmount

		// Store coupon code for later use
		if couponTemp, err := sos.couponService.GetCouponByCode(ctx, req.CouponCode); err == nil {
			if couponTemp != nil {
				appliedCouponID = couponTemp.ID
			}
		}
	}

	totalAmount := amount + setupFee - discountAmount

	// SECURITY: Price protection - validate final amount is reasonable
	if totalAmount < 0 {
		return nil, fmt.Errorf("invalid total amount: negative value not allowed")
	}

	// Prevent excessive discounts (more than 95% off)
	originalTotal := amount + setupFee
	if originalTotal > 0 && discountAmount > (originalTotal*0.95) {
		return nil, fmt.Errorf("discount amount %.2f exceeds maximum allowed (95%% of original price)", discountAmount)
	}

	// Validate amounts are within reasonable limits
	maxAllowedAmount := 10000.0 // $10,000 max
	if totalAmount > maxAllowedAmount {
		return nil, fmt.Errorf("total amount %.2f exceeds maximum allowed amount %.2f", totalAmount, maxAllowedAmount)
	}

	// Create subscription order
	order := &entities.SubscriptionOrder{
		UserID:             req.UserID,
		SubscriptionPlanID: req.SubscriptionPlanID,
		OrderNumber:        orderNumber,
		OrderType:          req.OrderType,
		Status:             entities.OrderStatusPending,
		Amount:             amount,
		Currency:           plan.Currency,
		SetupFee:           setupFee,
		DiscountAmount:     discountAmount,
		TotalAmount:        totalAmount,
		BillingPeriodStart: billingPeriodStart,
		BillingPeriodEnd:   billingPeriodEnd,
		PaymentGateway:     req.PaymentGateway,
		PaymentMethod:      req.PaymentMethod,
		CouponCode:         req.CouponCode,
		Metadata:           req.Metadata,
	}

	// Save order to database
	if err := tx.Create(order).Error; err != nil {
		tx.Rollback()
		logger.Error("Failed to create subscription order", logger.Error2("error", err))
		return nil, fmt.Errorf("failed to create subscription order: %w", err)
	}

	// Create payment order within the same transaction context
	paymentReq := &paymentInterfaces.CreatePaymentOrderRequest{
		UserID:              req.UserID,
		SubscriptionOrderID: &order.ID,
		Gateway:             req.PaymentGateway,
		PaymentMethod:       req.PaymentMethod,
		Amount:              totalAmount,
		Currency:            plan.Currency,
		Subject:             fmt.Sprintf("Subscription: %s", plan.Name),
		Body:                fmt.Sprintf("Payment for %s subscription plan", plan.Name),
		ReturnURL:           req.ReturnURL,
		ExpiredMinutes:      30, // 30 minutes expiration
	}

	paymentRecord, err := sos.paymentService.CreatePaymentOrder(ctx, paymentReq)
	if err != nil {
		// SECURITY: Rollback transaction if payment creation fails
		tx.Rollback()
		logger.Error("Failed to create payment order", logger.Error2("error", err))
		return nil, fmt.Errorf("failed to create payment order: %w", err)
	}

	// Commit transaction only after both order and payment are successfully created
	if err := tx.Commit().Error; err != nil {
		logger.Error("Failed to commit subscription order transaction", logger.Error2("error", err))
		return nil, fmt.Errorf("failed to commit subscription order transaction: %w", err)
	}

	// Record coupon usage if coupon was applied (outside of transaction since order is already committed)
	if appliedCouponID > 0 && couponDiscountAmount > 0 {
		// Update coupon service call to match correct signature: (ctx, couponID, userID, usedAmount, orderID)
		orderIDUint64 := uint64(order.ID)
		_, err := sos.couponService.UseCoupon(ctx, appliedCouponID, uint64(req.UserID), couponDiscountAmount, &orderIDUint64)
		if err != nil {
			logger.Error("Failed to record coupon usage",
				logger.Error2("error", err),
				logger.Any("coupon_id", appliedCouponID),
				logger.Uint("order_id", order.ID))
			// Don't fail the entire order creation, just log the error
		}
	}

	// Update order with payment information
	if err := sos.db.WithContext(ctx).Model(order).Updates(map[string]interface{}{
		"transaction_id": paymentRecord.PaymentNo,
		"updated_at":     time.Now(),
	}).Error; err != nil {
		logger.Error("Failed to update order with payment info", logger.Error2("error", err))
		// Don't fail the entire process, just log the error
	}

	logger.Info("Subscription order created successfully",
		logger.Uint("order_id", order.ID),
		logger.String("order_number", orderNumber),
		logger.Uint("user_id", req.UserID),
		logger.Uint("plan_id", req.SubscriptionPlanID))

	// Return response
	response := &interfaces.CreateSubscriptionOrderResponse{
		Order:         order.ToResponse(),
		PaymentRecord: paymentRecord.ToUserResponse(),
		PaymentURL:    paymentRecord.PaymentURL,
		QRCodeURL:     paymentRecord.QRCodeURL,
		ExpiredAt:     *paymentRecord.ExpiredAt,
	}

	return response, nil
}

// ProcessOrderPaymentSuccess processes successful payment for an order
func (sos *SubscriptionOrderService) ProcessOrderPaymentSuccess(ctx context.Context, orderID uint) error {
	// Start transaction
	tx := sos.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Get order with row-level lock to prevent race conditions
	var order entities.SubscriptionOrder
	if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&order, orderID).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get subscription order: %w", err)
	}

	// SECURITY: Validate order status to prevent duplicate processing
	if order.Status == entities.OrderStatusPaid {
		tx.Rollback()
		return fmt.Errorf("order %d is already paid", orderID)
	}
	if order.Status == entities.OrderStatusCancelled || order.Status == entities.OrderStatusFailed {
		tx.Rollback()
		return fmt.Errorf("order %d cannot be processed in status: %s", orderID, order.Status)
	}

	// Get subscription plan
	var plan entities.SubscriptionPlan
	if err := tx.First(&plan, order.SubscriptionPlanID).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get subscription plan: %w", err)
	}

	// Update order status
	if err := tx.Model(&order).Updates(map[string]interface{}{
		"status":     entities.OrderStatusPaid,
		"paid_at":    time.Now(),
		"updated_at": time.Now(),
	}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update order status: %w", err)
	}

	// Create or update user subscription based on order type
	switch order.OrderType {
	case entities.OrderTypeNew:
		// Create new subscription
		subscriptionReq := &interfaces.CreateSubscriptionRequest{
			UserID:             order.UserID,
			SubscriptionPlanID: order.SubscriptionPlanID,
			UseTrial:           false, // Paid subscription, no trial
		}

		// Use the billing period from order if available
		if order.BillingPeriodStart != nil {
			subscriptionReq.StartDate = order.BillingPeriodStart.Format(time.RFC3339)
		}

		subscription, err := sos.userSubscriptionService.CreateUserSubscription(ctx, subscriptionReq)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to create user subscription: %w", err)
		}

		// Update order with subscription ID
		if err := tx.Model(&order).Update("user_subscription_id", subscription.ID).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to update order with subscription ID: %w", err)
		}

		logger.Info("New subscription created for paid order",
			logger.Uint("order_id", order.ID),
			logger.Uint("subscription_id", subscription.ID))

	case entities.OrderTypeRenewal:
		// Find and renew existing subscription
		var subscription entities.UserSubscription
		if err := tx.Where("user_id = ? AND subscription_plan_id = ?", order.UserID, order.SubscriptionPlanID).
			First(&subscription).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to find subscription for renewal: %w", err)
		}

		// Renew subscription
		if err := sos.userSubscriptionService.RenewUserSubscription(ctx, subscription.ID); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to renew subscription: %w", err)
		}

		logger.Info("Subscription renewed for paid order",
			logger.Uint("order_id", order.ID),
			logger.Uint("subscription_id", subscription.ID))

	case entities.OrderTypeUpgrade, entities.OrderTypeDowngrade:
		// Find existing subscription for this user and current active plan
		var currentSubscription entities.UserSubscription
		if err := tx.Where("user_id = ? AND status IN (?)",
			order.UserID, []string{entities.UserSubscriptionStatusActive, entities.UserSubscriptionStatusTrial}).
			First(&currentSubscription).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to find current subscription for upgrade/downgrade: %w", err)
		}

		// Get current plan to calculate proration
		var currentPlan entities.SubscriptionPlan
		if err := tx.First(&currentPlan, currentSubscription.SubscriptionPlanID).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to get current subscription plan: %w", err)
		}

		// Calculate proration amount based on remaining billing period
		proratedCredit := calculateProration(&currentSubscription, &currentPlan, &plan)

		// Update order with proration details
		if proratedCredit > 0 {
			updatedTotalAmount := order.TotalAmount - proratedCredit
			if updatedTotalAmount < 0 {
				updatedTotalAmount = 0
			}
			if err := tx.Model(&order).Updates(map[string]interface{}{
				"discount_amount": order.DiscountAmount + proratedCredit,
				"total_amount":    updatedTotalAmount,
				"metadata":        fmt.Sprintf(`{"proration_credit": %.2f, "original_total": %.2f}`, proratedCredit, order.TotalAmount),
			}).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update order with proration: %w", err)
			}
		}

		// Cancel current subscription (mark as cancelled but keep active until period end)
		if err := tx.Model(&currentSubscription).Updates(map[string]interface{}{
			"cancel_at_period_end": true,
			"cancelled_at":         time.Now(),
			"cancellation_reason":  fmt.Sprintf("Subscription %s to plan %s", order.OrderType, plan.Name),
		}).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to cancel current subscription: %w", err)
		}

		// Create new subscription with the new plan
		subscriptionReq := &interfaces.CreateSubscriptionRequest{
			UserID:             order.UserID,
			SubscriptionPlanID: order.SubscriptionPlanID,
			UseTrial:           false, // Paid subscription, no trial for upgrades/downgrades
		}

		// For upgrade: start immediately
		// For downgrade: start at end of current billing period
		if order.OrderType == entities.OrderTypeUpgrade {
			subscriptionReq.StartDate = time.Now().Format(time.RFC3339)
		} else if currentSubscription.CurrentPeriodEnd != nil {
			subscriptionReq.StartDate = currentSubscription.CurrentPeriodEnd.Format(time.RFC3339)
		}

		// Inherit server group access from current subscription
		if currentSubscription.ServerGroupIDs != "" {
			subscriptionReq.ServerGroupIDs = currentSubscription.GetServerGroupIDs()
		}

		newSubscription, err := sos.userSubscriptionService.CreateUserSubscription(ctx, subscriptionReq)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to create new subscription for %s: %w", order.OrderType, err)
		}

		// Update order with new subscription ID
		if err := tx.Model(&order).Update("user_subscription_id", newSubscription.ID).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to update order with new subscription ID: %w", err)
		}

		// For upgrades, immediately activate the new subscription and deactivate the old one
		if order.OrderType == entities.OrderTypeUpgrade {
			// Deactivate current subscription immediately
			if err := tx.Model(&currentSubscription).Update("status", entities.UserSubscriptionStatusCancelled).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to deactivate current subscription: %w", err)
			}
		}

		logger.Info("Subscription upgrade/downgrade processed successfully",
			logger.Uint("order_id", order.ID),
			logger.String("order_type", order.OrderType),
			logger.Uint("old_subscription_id", currentSubscription.ID),
			logger.Uint("new_subscription_id", newSubscription.ID),
			logger.Any("proration_credit", proratedCredit))
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit order payment processing: %w", err)
	}

	logger.Info("Order payment processed successfully", logger.Uint("order_id", order.ID))
	return nil
}

// GetSubscriptionOrder gets a subscription order by ID
func (sos *SubscriptionOrderService) GetSubscriptionOrder(ctx context.Context, orderID uint) (*entities.SubscriptionOrder, error) {
	var order entities.SubscriptionOrder
	if err := sos.db.WithContext(ctx).First(&order, orderID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("subscription order not found")
		}
		logger.Error("Failed to get subscription order", logger.Error2("error", err), logger.Uint("order_id", orderID))
		return nil, fmt.Errorf("failed to get subscription order: %w", err)
	}
	return &order, nil
}

// GetSubscriptionOrderByNumber gets a subscription order by order number
func (sos *SubscriptionOrderService) GetSubscriptionOrderByNumber(ctx context.Context, orderNumber string) (*entities.SubscriptionOrder, error) {
	var order entities.SubscriptionOrder
	if err := sos.db.WithContext(ctx).Where("order_number = ?", orderNumber).First(&order).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("subscription order not found")
		}
		logger.Error("Failed to get subscription order by number", logger.Error2("error", err), logger.String("order_number", orderNumber))
		return nil, fmt.Errorf("failed to get subscription order: %w", err)
	}
	return &order, nil
}

// GetUserSubscriptionOrders gets subscription orders for a user with pagination
func (sos *SubscriptionOrderService) GetUserSubscriptionOrders(ctx context.Context, userID uint, limit, offset int) ([]*entities.SubscriptionOrder, int64, error) {
	var orders []*entities.SubscriptionOrder
	var totalCount int64

	// Get total count
	if err := sos.db.WithContext(ctx).Model(&entities.SubscriptionOrder{}).
		Where("user_id = ?", userID).
		Count(&totalCount).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count subscription orders: %w", err)
	}

	// Get orders with pagination
	query := sos.db.WithContext(ctx).Where("user_id = ?", userID).
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

// generateOrderNumber generates a unique order number
func (sos *SubscriptionOrderService) generateOrderNumber() string {
	// Generate format: ORD + YYYYMMDDHHMMSS + 3-digit random
	now := time.Now()
	timestamp := now.Format("20060102150405")
	random := now.Nanosecond() % 1000
	return fmt.Sprintf("ORD%s%03d", timestamp, random)
}

// calculateProration calculates the prorated credit for subscription changes
func calculateProration(currentSubscription *entities.UserSubscription, currentPlan *entities.SubscriptionPlan, newPlan *entities.SubscriptionPlan) float64 {
	// Skip proration for lifetime subscriptions
	if currentPlan.BillingCycle == entities.BillingCycleLifetime || newPlan.BillingCycle == entities.BillingCycleLifetime {
		return 0.0
	}

	// Check if we have valid billing period
	if currentSubscription.CurrentPeriodStart == nil || currentSubscription.CurrentPeriodEnd == nil {
		return 0.0
	}

	now := time.Now()
	periodStart := *currentSubscription.CurrentPeriodStart
	periodEnd := *currentSubscription.CurrentPeriodEnd

	// If current period has ended, no proration
	if now.After(periodEnd) {
		return 0.0
	}

	// Calculate total period duration in days
	totalPeriodDays := periodEnd.Sub(periodStart).Hours() / 24

	// Calculate remaining days in current period
	remainingDays := periodEnd.Sub(now).Hours() / 24

	// If no remaining time, no proration
	if remainingDays <= 0 {
		return 0.0
	}

	// Calculate proration percentage
	prorationPercentage := remainingDays / totalPeriodDays

	// Calculate prorated credit based on current subscription price
	proratedCredit := currentSubscription.Price * prorationPercentage

	// For downgrades, apply full proration
	// For upgrades, apply partial proration to encourage upgrades
	if newPlan.Price < currentPlan.Price {
		// Downgrade: full proration credit
		return proratedCredit
	} else {
		// Upgrade: reduced proration credit (70% of calculated amount)
		return proratedCredit * 0.7
	}
}

// UpgradeDowngradeRequest represents the request to upgrade or downgrade a subscription
type UpgradeDowngradeRequest struct {
	CurrentSubscriptionID uint   `json:"current_subscription_id" binding:"required" example:"1"`
	NewPlanID             uint   `json:"new_plan_id" binding:"required" example:"2"`
	PaymentGateway        string `json:"payment_gateway" binding:"required" example:"epay"`
	PaymentMethod         string `json:"payment_method" binding:"required" example:"alipay"`
	ReturnURL             string `json:"return_url,omitempty" example:"https://example.com/payment/return"`
	ApplyImmediately      *bool  `json:"apply_immediately,omitempty" example:"true"` // For upgrades, default true; for downgrades, default false
}

// CreateUpgradeDowngradeOrder creates an order for subscription upgrade or downgrade
func (sos *SubscriptionOrderService) CreateUpgradeDowngradeOrder(ctx context.Context, userID uint, req *UpgradeDowngradeRequest) (*interfaces.CreateSubscriptionOrderResponse, error) {
	// Get current subscription
	currentSubscription, err := sos.userSubscriptionService.GetUserSubscription(ctx, req.CurrentSubscriptionID)
	if err != nil {
		return nil, fmt.Errorf("invalid current subscription: %w", err)
	}

	// Verify user owns this subscription
	if currentSubscription.UserID != userID {
		return nil, fmt.Errorf("unauthorized: subscription does not belong to user")
	}

	// Check if subscription is active
	if !currentSubscription.IsActive() {
		return nil, fmt.Errorf("current subscription is not active")
	}

	// Get current plan
	currentPlan, err := sos.subscriptionPlanService.GetSubscriptionPlan(ctx, currentSubscription.SubscriptionPlanID)
	if err != nil {
		return nil, fmt.Errorf("failed to get current plan: %w", err)
	}

	// Get new plan
	newPlan, err := sos.subscriptionPlanService.GetSubscriptionPlan(ctx, req.NewPlanID)
	if err != nil {
		return nil, fmt.Errorf("invalid new subscription plan: %w", err)
	}

	if !newPlan.IsAvailableForPurchase() {
		return nil, fmt.Errorf("new subscription plan is not available for purchase")
	}

	// Determine order type based on plan prices
	var orderType string
	if newPlan.Price > currentPlan.Price {
		orderType = entities.OrderTypeUpgrade
	} else if newPlan.Price < currentPlan.Price {
		orderType = entities.OrderTypeDowngrade
	} else {
		return nil, fmt.Errorf("cannot change to plan with same price - use regular renewal instead")
	}

	// Create order request
	orderReq := &interfaces.CreateSubscriptionOrderRequest{
		UserID:             userID,
		SubscriptionPlanID: req.NewPlanID,
		OrderType:          orderType,
		PaymentGateway:     req.PaymentGateway,
		PaymentMethod:      req.PaymentMethod,
		ReturnURL:          req.ReturnURL,
		Metadata:           fmt.Sprintf(`{"current_subscription_id": %d, "apply_immediately": %t}`, req.CurrentSubscriptionID, getApplyImmediately(orderType, req.ApplyImmediately)),
	}

	// Create the order with proration calculation
	return sos.CreateSubscriptionOrder(ctx, orderReq)
}

// getApplyImmediately determines when to apply the subscription change
func getApplyImmediately(orderType string, userPreference *bool) bool {
	if userPreference != nil {
		return *userPreference
	}

	// Default behavior: upgrades apply immediately, downgrades apply at period end
	return orderType == entities.OrderTypeUpgrade
}

// ================ Admin Order Management Methods ================

// GetOrdersRequest represents the request to get orders with advanced filtering
type GetOrdersRequest struct {
	UserID           *uint   `json:"user_id,omitempty"`
	Status           string  `json:"status,omitempty"`
	OrderType        string  `json:"order_type,omitempty"`
	PaymentMethod    string  `json:"payment_method,omitempty"`
	PaymentGateway   string  `json:"payment_gateway,omitempty"`
	MinAmount        float64 `json:"min_amount,omitempty"`
	MaxAmount        float64 `json:"max_amount,omitempty"`
	StartDate        string  `json:"start_date,omitempty"`
	EndDate          string  `json:"end_date,omitempty"`
	CouponCode       string  `json:"coupon_code,omitempty"`
	Search           string  `json:"search,omitempty"`
	SortBy           string  `json:"sort_by,omitempty"`
	SortOrder        string  `json:"sort_order,omitempty"`
	Limit            int     `json:"limit,omitempty"`
	Offset           int     `json:"offset,omitempty"`
	IncludeRelations bool    `json:"include_relations,omitempty"`
}

// GetOrdersWithFiltering gets orders with advanced filtering and search
func (sos *SubscriptionOrderService) GetOrdersWithFiltering(ctx context.Context, req *GetOrdersRequest) ([]*entities.SubscriptionOrder, int64, error) {
	query := sos.db.WithContext(ctx).Model(&entities.SubscriptionOrder{})

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
func (sos *SubscriptionOrderService) GetOrderWithRelations(ctx context.Context, orderID uint) (*entities.SubscriptionOrder, error) {
	var order entities.SubscriptionOrder
	if err := sos.db.WithContext(ctx).
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

// PaymentEvidence represents payment verification evidence
type PaymentEvidence struct {
	PaymentMethod    string  `json:"payment_method" binding:"required"`
	TransactionID    string  `json:"transaction_id" binding:"required"`
	PaymentReference string  `json:"payment_reference,omitempty"`
	PaymentProof     string  `json:"payment_proof,omitempty"`
	VerifiedAmount   float64 `json:"verified_amount" binding:"required,min=0"`
	Notes            string  `json:"notes,omitempty"`
}

// UpdateOrderStatusRequest represents the enhanced request to update order status
type UpdateOrderStatusRequest struct {
	Status          string           `json:"status" binding:"required"`
	Notes           string           `json:"notes,omitempty"`
	Reason          string           `json:"reason,omitempty"`
	PaymentEvidence *PaymentEvidence `json:"payment_evidence,omitempty"`
	AdminConfirm    bool             `json:"admin_confirm,omitempty"`
}

// UpdateOrderStatus updates order status with admin notes and payment verification
func (sos *SubscriptionOrderService) UpdateOrderStatus(ctx context.Context, orderID uint, status string, adminID uint, notes, reason string) (*entities.SubscriptionOrder, error) {
	return sos.UpdateOrderStatusWithEvidence(ctx, orderID, &UpdateOrderStatusRequest{
		Status: status,
		Notes:  notes,
		Reason: reason,
	}, adminID)
}

// UpdateOrderStatusWithEvidence updates order status with payment evidence verification
func (sos *SubscriptionOrderService) UpdateOrderStatusWithEvidence(ctx context.Context, orderID uint, req *UpdateOrderStatusRequest, adminID uint) (*entities.SubscriptionOrder, error) {
	// Start transaction
	tx := sos.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Get order with lock
	var order entities.SubscriptionOrder
	if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&order, orderID).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("subscription order not found")
		}
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	// SECURITY: Enhanced validation for status changes to "paid"
	if req.Status == entities.OrderStatusPaid {
		// Require payment evidence for manual payment confirmation
		if req.PaymentEvidence == nil {
			tx.Rollback()
			return nil, fmt.Errorf("payment evidence is required when marking order as paid")
		}

		// Validate payment evidence
		if err := sos.validatePaymentEvidence(&order, req.PaymentEvidence); err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("payment evidence validation failed: %w", err)
		}
	}

	// SECURITY: Require admin confirmation for critical operations
	if sos.isCriticalOperation(order.Status, req.Status) && !req.AdminConfirm {
		tx.Rollback()
		return nil, fmt.Errorf("admin confirmation is required for this status change")
	}

	// Prepare update data
	updateData := map[string]interface{}{
		"status":     req.Status,
		"updated_at": time.Now(),
	}

	// Add payment evidence data if provided
	if req.PaymentEvidence != nil {
		updateData["payment_method"] = req.PaymentEvidence.PaymentMethod
		updateData["transaction_id"] = req.PaymentEvidence.TransactionID
		if req.PaymentEvidence.PaymentReference != "" {
			updateData["payment_reference"] = req.PaymentEvidence.PaymentReference
		}
	}

	// Add admin notes with enhanced security logging
	adminNote := fmt.Sprintf("[%s] Admin (ID:%d)", time.Now().Format("2006-01-02 15:04:05"), adminID)

	if req.Status == entities.OrderStatusPaid && req.PaymentEvidence != nil {
		adminNote += fmt.Sprintf(" - Payment verified: Method=%s, TxnID=%s, Amount=%.2f",
			req.PaymentEvidence.PaymentMethod,
			req.PaymentEvidence.TransactionID,
			req.PaymentEvidence.VerifiedAmount)
	}

	if req.Notes != "" {
		adminNote += fmt.Sprintf(" - Notes: %s", req.Notes)
	}

	if req.Reason != "" {
		adminNote += fmt.Sprintf(" - Reason: %s", req.Reason)
	}

	if req.PaymentEvidence != nil && req.PaymentEvidence.Notes != "" {
		adminNote += fmt.Sprintf(" - Payment Notes: %s", req.PaymentEvidence.Notes)
	}

	currentNotes := order.Notes
	if currentNotes != "" {
		updateData["notes"] = currentNotes + "\n" + adminNote
	} else {
		updateData["notes"] = adminNote
	}

	// Set paid_at if status is paid
	if req.Status == entities.OrderStatusPaid && order.PaidAt == nil {
		updateData["paid_at"] = time.Now()
	}

	// Update order
	if err := tx.Model(&order).Updates(updateData).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to update order status: %w", err)
	}

	// If changing to paid status, process the order
	if req.Status == entities.OrderStatusPaid && order.Status != entities.OrderStatusPaid {
		if err := sos.ProcessOrderPaymentSuccess(ctx, orderID); err != nil {
			logger.Error("Failed to process order after status update",
				logger.Error2("error", err),
				logger.Uint("order_id", orderID))
			// Don't fail the status update, but log the error
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit status update: %w", err)
	}

	// Log security event for audit
	logger.Info("Order status updated by admin",
		logger.Uint("admin_id", adminID),
		logger.Uint("order_id", orderID),
		logger.String("old_status", order.Status),
		logger.String("new_status", req.Status),
		logger.Any("payment_verified", req.PaymentEvidence != nil),
		logger.Any("admin_confirmed", req.AdminConfirm))

	// Reload order with updated data
	if err := sos.db.WithContext(ctx).First(&order, orderID).Error; err != nil {
		return nil, fmt.Errorf("failed to reload order: %w", err)
	}

	return &order, nil
}

// validatePaymentEvidence validates payment evidence for manual payment confirmation
func (sos *SubscriptionOrderService) validatePaymentEvidence(order *entities.SubscriptionOrder, evidence *PaymentEvidence) error {
	// SECURITY: Validate payment amount matches order total
	if evidence.VerifiedAmount != order.TotalAmount {
		return fmt.Errorf("verified amount %.2f does not match order total %.2f", evidence.VerifiedAmount, order.TotalAmount)
	}

	// SECURITY: Ensure transaction ID is not empty and unique
	if len(evidence.TransactionID) < 6 {
		return fmt.Errorf("transaction ID must be at least 6 characters")
	}

	// Check for duplicate transaction ID in other orders
	var existingOrder entities.SubscriptionOrder
	if err := sos.db.Where("transaction_id = ? AND id != ?", evidence.TransactionID, order.ID).First(&existingOrder).Error; err == nil {
		return fmt.Errorf("transaction ID %s is already used by order %d", evidence.TransactionID, existingOrder.ID)
	}

	// SECURITY: Validate payment method
	validMethods := map[string]bool{
		"bank_transfer": true,
		"alipay":        true,
		"wechat":        true,
		"credit_card":   true,
		"crypto":        true,
		"paypal":        true,
		"manual":        true,
	}

	if !validMethods[evidence.PaymentMethod] {
		return fmt.Errorf("invalid payment method: %s", evidence.PaymentMethod)
	}

	return nil
}

// isCriticalOperation determines if a status change requires additional confirmation
func (sos *SubscriptionOrderService) isCriticalOperation(oldStatus, newStatus string) bool {
	criticalChanges := map[string][]string{
		entities.OrderStatusPending:   {entities.OrderStatusPaid},                                    // Manual payment confirmation
		entities.OrderStatusFailed:    {entities.OrderStatusPaid},                                    // Retry failed payment
		entities.OrderStatusCancelled: {entities.OrderStatusPaid, entities.OrderStatusPending},       // Reactivate cancelled order
		entities.OrderStatusPaid:      {entities.OrderStatusRefunded, entities.OrderStatusCancelled}, // Reverse paid order
	}

	allowedTargets, exists := criticalChanges[oldStatus]
	if !exists {
		return false
	}

	for _, target := range allowedTargets {
		if target == newStatus {
			return true
		}
	}

	return false
}

// ProcessRefundRequest represents the request to process a refund
type ProcessRefundRequest struct {
	OrderID      uint    `json:"order_id"`
	Amount       float64 `json:"amount"`
	Reason       string  `json:"reason"`
	RefundMethod string  `json:"refund_method,omitempty"`
	Notes        string  `json:"notes,omitempty"`
	ProcessedBy  uint    `json:"processed_by"`
}

// ProcessRefund processes a refund for an order
func (sos *SubscriptionOrderService) ProcessRefund(ctx context.Context, req *ProcessRefundRequest) (*entities.SubscriptionOrder, error) {
	// Start transaction
	tx := sos.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Get order with lock
	var order entities.SubscriptionOrder
	if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&order, req.OrderID).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("subscription order not found")
		}
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	// Validate refund
	if !order.CanBeRefunded() {
		tx.Rollback()
		return nil, fmt.Errorf("order cannot be refunded")
	}

	// SECURITY: Validate refund time window (30 days from payment)
	if order.PaidAt != nil {
		refundDeadline := order.PaidAt.AddDate(0, 0, 30) // 30 days from payment
		if time.Now().After(refundDeadline) {
			tx.Rollback()
			return nil, fmt.Errorf("refund window has expired (30 days from payment date)")
		}
	}

	maxRefundable := order.GetRefundableAmount()
	if req.Amount > maxRefundable {
		tx.Rollback()
		return nil, fmt.Errorf("refund amount %.2f exceeds refundable amount %.2f", req.Amount, maxRefundable)
	}

	// Update order with refund information
	updateData := map[string]interface{}{
		"refund_amount": order.RefundAmount + req.Amount,
		"refund_reason": req.Reason,
		"refunded_at":   time.Now(),
		"updated_at":    time.Now(),
	}

	// If full refund, update status
	if order.RefundAmount+req.Amount >= order.TotalAmount {
		updateData["status"] = entities.OrderStatusRefunded
	}

	// Add admin notes
	if req.Notes != "" {
		currentNotes := order.Notes
		adminNote := fmt.Sprintf("[%s] Refund processed by Admin (ID:%d): Amount=%.2f, Reason=%s, Notes=%s",
			time.Now().Format("2006-01-02 15:04:05"), req.ProcessedBy, req.Amount, req.Reason, req.Notes)

		if currentNotes != "" {
			updateData["notes"] = currentNotes + "\n" + adminNote
		} else {
			updateData["notes"] = adminNote
		}
	}

	// Update order
	if err := tx.Model(&order).Updates(updateData).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to update order with refund: %w", err)
	}

	// TODO: Integrate with payment gateway to process actual refund
	// This would call the payment service to process the refund through the original payment gateway

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit refund: %w", err)
	}

	// Reload order with updated data
	if err := sos.db.WithContext(ctx).First(&order, req.OrderID).Error; err != nil {
		return nil, fmt.Errorf("failed to reload order: %w", err)
	}

	return &order, nil
}

// GetOrderStatsRequest represents the request for order statistics
type GetOrderStatsRequest struct {
	Period    string `json:"period,omitempty"`
	StartDate string `json:"start_date,omitempty"`
	EndDate   string `json:"end_date,omitempty"`
}

// OrderStatistics represents order statistics response
type OrderStatistics struct {
	TotalOrders     int64   `json:"total_orders"`
	PendingOrders   int64   `json:"pending_orders"`
	PaidOrders      int64   `json:"paid_orders"`
	FailedOrders    int64   `json:"failed_orders"`
	CancelledOrders int64   `json:"cancelled_orders"`
	RefundedOrders  int64   `json:"refunded_orders"`
	TotalRevenue    float64 `json:"total_revenue"`
	TotalRefunded   float64 `json:"total_refunded"`
	AvgOrderValue   float64 `json:"avg_order_value"`
	ConversionRate  float64 `json:"conversion_rate"`
}

// GetSubscriptionOrders gets subscription orders with filtering
func (sos *SubscriptionOrderService) GetSubscriptionOrders(ctx context.Context, req *interfaces.GetSubscriptionOrdersRequest) ([]*entities.SubscriptionOrder, int64, error) {
	query := sos.db.WithContext(ctx).Model(&entities.SubscriptionOrder{})

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

// CancelSubscriptionOrder cancels a subscription order
func (sos *SubscriptionOrderService) CancelSubscriptionOrder(ctx context.Context, orderID uint, reason string) error {
	// Start transaction
	tx := sos.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Get order with lock
	var order entities.SubscriptionOrder
	if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&order, orderID).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("subscription order not found")
		}
		return fmt.Errorf("failed to get subscription order: %w", err)
	}

	// Check if order can be cancelled
	if order.Status != entities.OrderStatusPending {
		tx.Rollback()
		return fmt.Errorf("order cannot be cancelled in status: %s", order.Status)
	}

	// Update order status
	updateData := map[string]interface{}{
		"status":              entities.OrderStatusCancelled,
		"cancellation_reason": reason,
		"cancelled_at":        time.Now(),
		"updated_at":          time.Now(),
	}

	if err := tx.Model(&order).Updates(updateData).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to cancel order: %w", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit order cancellation: %w", err)
	}

	logger.Info("Order cancelled successfully",
		logger.Uint("order_id", orderID),
		logger.String("reason", reason))

	return nil
}

// GetOrderStatistics gets comprehensive order statistics
func (sos *SubscriptionOrderService) GetOrderStatistics(ctx context.Context, fromDate, toDate time.Time) (map[string]interface{}, error) {
	query := sos.db.WithContext(ctx).Model(&entities.SubscriptionOrder{})

	// Apply date range
	if !fromDate.IsZero() {
		query = query.Where("created_at >= ?", fromDate)
	}

	if !toDate.IsZero() {
		query = query.Where("created_at <= ?", toDate)
	}

	stats := make(map[string]interface{})

	// Get total orders
	var totalOrders int64
	query.Count(&totalOrders)
	stats["total_orders"] = totalOrders

	// Get orders by status
	var statusCounts []struct {
		Status string
		Count  int64
	}

	if err := query.Select("status, count(*) as count").Group("status").Find(&statusCounts).Error; err != nil {
		return nil, fmt.Errorf("failed to get status counts: %w", err)
	}

	var pendingOrders, paidOrders, failedOrders, cancelledOrders, refundedOrders int64
	for _, sc := range statusCounts {
		switch sc.Status {
		case entities.OrderStatusPending:
			pendingOrders = sc.Count
		case entities.OrderStatusPaid:
			paidOrders = sc.Count
		case entities.OrderStatusFailed:
			failedOrders = sc.Count
		case entities.OrderStatusCancelled:
			cancelledOrders = sc.Count
		case entities.OrderStatusRefunded:
			refundedOrders = sc.Count
		}
	}

	stats["pending_orders"] = pendingOrders
	stats["paid_orders"] = paidOrders
	stats["failed_orders"] = failedOrders
	stats["cancelled_orders"] = cancelledOrders
	stats["refunded_orders"] = refundedOrders

	// Get revenue statistics
	var revenueStats struct {
		TotalRevenue  float64
		TotalRefunded float64
		AvgOrderValue float64
	}

	if err := query.Select(
		"COALESCE(SUM(CASE WHEN status = 'paid' THEN total_amount ELSE 0 END), 0) as total_revenue, " +
			"COALESCE(SUM(refund_amount), 0) as total_refunded, " +
			"COALESCE(AVG(CASE WHEN status = 'paid' THEN total_amount ELSE NULL END), 0) as avg_order_value",
	).Scan(&revenueStats).Error; err != nil {
		return nil, fmt.Errorf("failed to get revenue stats: %w", err)
	}

	stats["total_revenue"] = revenueStats.TotalRevenue
	stats["total_refunded"] = revenueStats.TotalRefunded
	stats["avg_order_value"] = revenueStats.AvgOrderValue

	// Calculate conversion rate (paid orders / total orders)
	if totalOrders > 0 {
		stats["conversion_rate"] = float64(paidOrders) / float64(totalOrders) * 100
	} else {
		stats["conversion_rate"] = 0.0
	}

	return stats, nil
}

// BulkUpdateOrdersRequest represents bulk update request
type BulkUpdateOrdersRequest struct {
	OrderIDs    []uint `json:"order_ids"`
	Operation   string `json:"operation"`
	Reason      string `json:"reason,omitempty"`
	Notes       string `json:"notes,omitempty"`
	ProcessedBy uint   `json:"processed_by"`
}

// BulkUpdateResult represents the result of bulk operations
type BulkUpdateResult struct {
	Successful []uint                 `json:"successful"`
	Failed     []BulkUpdateFailure    `json:"failed"`
	Summary    map[string]interface{} `json:"summary"`
}

// BulkUpdateFailure represents a failed bulk operation
type BulkUpdateFailure struct {
	OrderID uint   `json:"order_id"`
	Error   string `json:"error"`
}

// BulkUpdateOrders performs bulk operations on orders
func (sos *SubscriptionOrderService) BulkUpdateOrders(ctx context.Context, req *BulkUpdateOrdersRequest) (*BulkUpdateResult, error) {
	result := &BulkUpdateResult{
		Successful: make([]uint, 0),
		Failed:     make([]BulkUpdateFailure, 0),
	}

	for _, orderID := range req.OrderIDs {
		var err error

		switch req.Operation {
		case "cancel":
			_, err = sos.UpdateOrderStatus(ctx, orderID, entities.OrderStatusCancelled, req.ProcessedBy, req.Notes, req.Reason)
		case "refund":
			// For bulk refund, we'll refund the full amount
			order, getErr := sos.GetSubscriptionOrder(ctx, orderID)
			if getErr != nil {
				err = getErr
			} else if order.CanBeRefunded() {
				refundAmount := order.GetRefundableAmount()
				_, err = sos.ProcessRefund(ctx, &ProcessRefundRequest{
					OrderID:     orderID,
					Amount:      refundAmount,
					Reason:      req.Reason,
					Notes:       req.Notes,
					ProcessedBy: req.ProcessedBy,
				})
			} else {
				err = fmt.Errorf("order cannot be refunded")
			}
		default:
			err = fmt.Errorf("unsupported operation: %s", req.Operation)
		}

		if err != nil {
			result.Failed = append(result.Failed, BulkUpdateFailure{
				OrderID: orderID,
				Error:   err.Error(),
			})
		} else {
			result.Successful = append(result.Successful, orderID)
		}
	}

	// Generate summary
	result.Summary = map[string]interface{}{
		"total_processed": len(req.OrderIDs),
		"successful":      len(result.Successful),
		"failed":          len(result.Failed),
		"operation":       req.Operation,
	}

	return result, nil
}

// checkDuplicateOrders prevents users from creating multiple pending orders for the same plan
func (sos *SubscriptionOrderService) checkDuplicateOrders(ctx context.Context, userID uint, planID uint, orderType string) error {
	// SECURITY: Check for pending orders with same plan and type
	var pendingCount int64
	if err := sos.db.WithContext(ctx).Model(&entities.SubscriptionOrder{}).
		Where("user_id = ? AND subscription_plan_id = ? AND order_type = ? AND status IN (?)",
			userID, planID, orderType, []string{entities.OrderStatusPending}).
		Count(&pendingCount).Error; err != nil {
		return fmt.Errorf("failed to check pending orders: %w", err)
	}

	if pendingCount > 0 {
		return fmt.Errorf("you already have a pending order for this subscription plan")
	}

	// SECURITY: Check for recent orders within cooldown period (5 minutes)
	cooldownTime := time.Now().Add(-5 * time.Minute)
	var recentCount int64
	if err := sos.db.WithContext(ctx).Model(&entities.SubscriptionOrder{}).
		Where("user_id = ? AND subscription_plan_id = ? AND order_type = ? AND created_at > ?",
			userID, planID, orderType, cooldownTime).
		Count(&recentCount).Error; err != nil {
		return fmt.Errorf("failed to check recent orders: %w", err)
	}

	if recentCount >= 3 {
		return fmt.Errorf("too many recent order attempts, please wait a few minutes before trying again")
	}

	// SECURITY: For new subscriptions, check if user already has active subscription
	if orderType == entities.OrderTypeNew {
		var activeCount int64
		if err := sos.db.WithContext(ctx).Model(&entities.UserSubscription{}).
			Where("user_id = ? AND subscription_plan_id = ? AND status IN (?)",
				userID, planID, []string{entities.UserSubscriptionStatusActive, entities.UserSubscriptionStatusTrial}).
			Count(&activeCount).Error; err != nil {
			return fmt.Errorf("failed to check active subscriptions: %w", err)
		}

		if activeCount > 0 {
			return fmt.Errorf("you already have an active subscription for this plan")
		}
	}

	return nil
}

// checkRefundTimeWindow validates if refund is within allowed time window
func (sos *SubscriptionOrderService) checkRefundTimeWindow(order *entities.SubscriptionOrder) error {
	// SECURITY: Enforce refund time limits
	if order.PaidAt == nil {
		return fmt.Errorf("order has not been paid")
	}

	// Allow refunds within 30 days of payment
	refundDeadline := order.PaidAt.Add(30 * 24 * time.Hour)
	if time.Now().After(refundDeadline) {
		return fmt.Errorf("refund window has expired (30 days maximum)")
	}

	return nil
}
