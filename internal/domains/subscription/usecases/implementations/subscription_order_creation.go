package implementations

import (
	"context"
	"fmt"
	"strings"
	"time"

	couponDto "linke/internal/domains/coupon/dto"
	invoiceDto "linke/internal/domains/invoice/dto"
	invoiceEntities "linke/internal/domains/invoice/entities"
	paymentDto "linke/internal/domains/payment/dto"
	"linke/internal/domains/subscription/constants"
	"linke/internal/domains/subscription/dto"
	"linke/internal/domains/subscription/entities"
	"linke/internal/shared/events"
	"linke/internal/shared/logger"
)

// SubscriptionOrderCreationService handles all order creation operations
type SubscriptionOrderCreationService struct {
	*SubscriptionOrderServiceBase
	utilsService *SubscriptionOrderUtilsService
}

// NewSubscriptionOrderCreationService creates a new order creation service
func NewSubscriptionOrderCreationService(base *SubscriptionOrderServiceBase, utilsService *SubscriptionOrderUtilsService) *SubscriptionOrderCreationService {
	return &SubscriptionOrderCreationService{
		SubscriptionOrderServiceBase: base,
		utilsService:                 utilsService,
	}
}

// CreateOrderWithInvoice creates a complete order -> invoice -> payment flow
func (socs *SubscriptionOrderCreationService) CreateOrderWithInvoice(ctx context.Context, req *dto.CreateOrderRequest) (*dto.CreateOrderResponse, error) {
	// Convert to legacy request for now - this uses the existing battle-tested logic
	legacyReq := &dto.CreateSubscriptionOrderRequest{
		UserID:             req.UserID,
		SubscriptionPlanID: req.SubscriptionPlanID,
		OrderType:          req.OrderType,
		CouponCode:         req.CouponCode,
		PaymentGateway:     req.PaymentGateway,
		PaymentMethod:      req.PaymentMethod,
		PaymentMethodID:    req.PaymentMethodID,
		UseDefaultPayment:  req.UseDefaultPayment,
		ReturnURL:          req.ReturnURL,
		Metadata:           req.Metadata,
	}

	// Call the existing implementation
	legacyResp, err := socs.CreateSubscriptionOrder(ctx, legacyReq)
	if err != nil {
		return nil, err
	}

	// Convert to new response format
	response := &dto.CreateOrderResponse{
		Order:         legacyResp.Order,
		Invoice:       legacyResp.Invoice,
		PaymentRecord: legacyResp.PaymentRecord,
		PaymentURL:    legacyResp.PaymentURL,
		QRCodeURL:     legacyResp.QRCodeURL,
		ExpiredAt:     legacyResp.ExpiredAt,
	}

	// Publish order created event
	if socs.eventPublisher != nil {
		event := events.NewBaseEvent("subscription.order.created", "subscription", map[string]any{
			"order_id":   legacyResp.Order.ID,
			"user_id":    req.UserID,
			"plan_id":    req.SubscriptionPlanID,
			"order_type": req.OrderType,
			"amount":     legacyResp.Order.TotalAmount,
			"currency":   legacyResp.Order.Currency,
		})
		if err := socs.eventPublisher.Publish(ctx, event); err != nil {
			logger.Error("Failed to publish order created event", logger.ErrorField(err))
		}
	}

	return response, nil
}

// GenerateInvoiceForOrder creates an invoice for an existing order
func (socs *SubscriptionOrderCreationService) GenerateInvoiceForOrder(ctx context.Context, orderID uint) (any, error) {
	// Get the order
	var order entities.SubscriptionOrder
	if err := socs.db.WithContext(ctx).First(&order, orderID).Error; err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	// Check if order can have an invoice generated
	if order.Status != constants.OrderStatusPending && order.Status != constants.OrderStatusConfirmed {
		return nil, fmt.Errorf("invoice can only be generated for pending or confirmed orders")
	}

	// Get plan for invoice details
	plan, err := socs.subscriptionPlanService.GetSubscriptionPlan(ctx, order.SubscriptionPlanID)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription plan: %w", err)
	}

	// Create invoice request
	invoiceReq := &invoiceDto.CreateInvoiceRequest{
		UserID:              order.UserID,
		SubscriptionOrderID: order.ID,
		Amount:              order.TotalAmount,
		Currency:            order.Currency,
		Description:         fmt.Sprintf("Subscription: %s", plan.Name),
		BillingName:         fmt.Sprintf("User %d", order.UserID),              // TODO: Get actual user name
		BillingEmail:        fmt.Sprintf("user%d@example.com", order.UserID),   // TODO: Get actual user email
		DueDate:             time.Now().AddDate(0, 0, 30).Format("2006-01-02"), // 30 days from now
	}

	// Create invoice
	invoice, err := socs.invoiceService.CreateInvoice(ctx, invoiceReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create invoice: %w", err)
	}

	// Update order with invoice information
	if err := socs.db.WithContext(ctx).Model(&order).Updates(map[string]any{
		"invoice_number": invoice.InvoiceNumber,
		"invoice_status": invoice.Status,
		"invoiced_at":    time.Now(),
		"updated_at":     time.Now(),
	}).Error; err != nil {
		logger.Error("Failed to update order with invoice info", logger.ErrorField(err))
		// Don't fail the entire process, just log the error
	}

	// Publish invoice generated event
	if socs.eventPublisher != nil {
		event := events.NewBaseEvent("subscription.invoice.generated", "subscription", map[string]any{
			"order_id":       order.ID,
			"invoice_id":     invoice.ID,
			"invoice_number": invoice.InvoiceNumber,
			"user_id":        order.UserID,
			"amount":         invoice.TotalAmount,
		})
		if err := socs.eventPublisher.Publish(ctx, event); err != nil {
			logger.Error("Failed to publish invoice generated event", logger.ErrorField(err))
		}
	}

	return invoiceDto.ToResponse(invoice), nil
}

// CreatePaymentForOrder creates a payment record for an existing order
func (socs *SubscriptionOrderCreationService) CreatePaymentForOrder(ctx context.Context, req *dto.PayOrderRequest) (*dto.PayOrderResponse, error) {
	// Get the order
	var order entities.SubscriptionOrder
	if err := socs.db.WithContext(ctx).First(&order, req.OrderID).Error; err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	// Check if order can be paid
	if order.Status != constants.OrderStatusPending && order.Status != constants.OrderStatusConfirmed {
		return nil, fmt.Errorf("payment can only be created for pending or confirmed orders")
	}

	// Get plan for payment details
	plan, err := socs.subscriptionPlanService.GetSubscriptionPlan(ctx, order.SubscriptionPlanID)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription plan: %w", err)
	}

	// Resolve payment method if using saved payment methods
	finalPaymentGateway := req.PaymentGateway
	finalPaymentMethod := req.PaymentMethod

	if req.UseDefaultPayment {
		// Use default payment method for the specified gateway
		defaultPaymentMethod, err := socs.paymentMethodService.GetDefaultPaymentMethodByGateway(ctx, order.UserID, req.PaymentGateway)
		if err != nil {
			return nil, fmt.Errorf("failed to get default payment method: %w", err)
		}
		finalPaymentMethod = defaultPaymentMethod.Method
	} else if req.PaymentMethodID != nil {
		// Use specific saved payment method
		paymentMethod, err := socs.paymentMethodService.GetPaymentMethod(ctx, order.UserID, *req.PaymentMethodID)
		if err != nil {
			return nil, fmt.Errorf("failed to get payment method: %w", err)
		}
		finalPaymentGateway = paymentMethod.Gateway
		finalPaymentMethod = paymentMethod.Method
	}

	// Check if there's already an invoice for this order
	var invoice *invoiceEntities.Invoice
	var invoiceResp *invoiceDto.InvoiceResponse
	if err := socs.db.WithContext(ctx).Where("subscription_order_id = ?", order.ID).First(&invoice).Error; err == nil {
		invoiceResp = invoiceDto.ToResponse(invoice)
	}

	// Create payment request
	paymentReq := &paymentDto.CreatePaymentOrderRequest{
		UserID:              order.UserID,
		SubscriptionOrderID: &order.ID,
		Gateway:             finalPaymentGateway,
		PaymentMethod:       finalPaymentMethod,
		Amount:              order.TotalAmount,
		Currency:            order.Currency,
		Subject:             strings.ReplaceAll(plan.Name, " ", ""),
		Body:                fmt.Sprintf("Payment for order %s - %s subscription plan", order.OrderNumber, plan.Name),
		ClientIP:            req.ClientIP,
		ReturnURL:           req.ReturnURL,
		ExpiredMinutes:      30, // 30 minutes expiration
		Metadata:            req.Metadata,
	}

	// Add invoice reference if exists
	if invoice != nil {
		paymentReq.InvoiceID = &invoice.ID
	}

	// Create payment record
	paymentRecord, err := socs.paymentService.CreatePaymentOrder(ctx, paymentReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment order: %w", err)
	}

	// Update order with payment information
	if err := socs.db.WithContext(ctx).Model(&order).Updates(map[string]any{
		"transaction_id":  paymentRecord.PaymentNo,
		"payment_gateway": finalPaymentGateway,
		"payment_method":  finalPaymentMethod,
		"payment_status":  paymentRecord.Status,
		"updated_at":      time.Now(),
	}).Error; err != nil {
		logger.Error("Failed to update order with payment info", logger.ErrorField(err))
		// Don't fail the entire process, just log the error
	}

	// Build response
	response := &dto.PayOrderResponse{
		Order:         order.ToResponse(),
		Invoice:       invoiceResp,
		PaymentRecord: paymentDto.ToPaymentRecordUserResponse(paymentRecord),
		PaymentURL:    paymentRecord.PaymentURL,
		QRCodeURL:     paymentRecord.QRCodeURL,
		ExpiredAt:     *paymentRecord.ExpiredAt,
	}

	// Publish payment created event
	if socs.eventPublisher != nil {
		event := events.NewBaseEvent("subscription.payment.created", "subscription", map[string]any{
			"order_id":   order.ID,
			"payment_id": paymentRecord.ID,
			"payment_no": paymentRecord.PaymentNo,
			"user_id":    order.UserID,
			"amount":     paymentRecord.Amount,
			"gateway":    finalPaymentGateway,
			"method":     finalPaymentMethod,
		})
		if err := socs.eventPublisher.Publish(ctx, event); err != nil {
			logger.Error("Failed to publish payment created event", logger.ErrorField(err))
		}
	}

	return response, nil
}

// CreateSubscriptionOrder creates a new subscription order with payment
func (socs *SubscriptionOrderCreationService) CreateSubscriptionOrder(ctx context.Context, req *dto.CreateSubscriptionOrderRequest) (*dto.CreateSubscriptionOrderResponse, error) {
	// SECURITY: Check for duplicate pending orders to prevent order spam
	if err := socs.utilsService.checkDuplicateOrders(ctx, req.UserID, req.SubscriptionPlanID, req.OrderType); err != nil {
		return nil, fmt.Errorf("duplicate order check failed: %w", err)
	}

	// Resolve payment method if using saved payment methods
	finalPaymentGateway := req.PaymentGateway
	finalPaymentMethod := req.PaymentMethod
	var usedPaymentMethodID *uint

	if req.UseDefaultPayment {
		// Use default payment method for the specified gateway
		defaultPaymentMethod, err := socs.paymentMethodService.GetDefaultPaymentMethodByGateway(ctx, req.UserID, req.PaymentGateway)
		if err != nil {
			return nil, fmt.Errorf("failed to get default payment method: %w", err)
		}
		finalPaymentMethod = defaultPaymentMethod.Method
		usedPaymentMethodID = &defaultPaymentMethod.ID
	} else if req.PaymentMethodID != nil {
		// Use specific saved payment method
		paymentMethod, err := socs.paymentMethodService.GetPaymentMethod(ctx, req.UserID, *req.PaymentMethodID)
		if err != nil {
			return nil, fmt.Errorf("failed to get payment method: %w", err)
		}
		finalPaymentGateway = paymentMethod.Gateway
		finalPaymentMethod = paymentMethod.Method
		usedPaymentMethodID = req.PaymentMethodID
	}

	// Get subscription plan
	plan, err := socs.subscriptionPlanService.GetSubscriptionPlan(ctx, req.SubscriptionPlanID)
	if err != nil {
		return nil, fmt.Errorf("invalid subscription plan: %w", err)
	}

	if !plan.IsAvailableForPurchase() {
		return nil, fmt.Errorf("subscription plan is not available for purchase")
	}

	// Start database transaction
	tx := socs.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Generate order number
	orderNumber := socs.utilsService.generateOrderNumber()

	// Calculate billing period
	var billingPeriodStart, billingPeriodEnd *time.Time
	now := time.Now()

	if plan.BillingCycle != constants.BillingCycleLifetime {
		periodStart := now
		billingPeriodStart = &periodStart

		var periodEnd time.Time
		switch plan.BillingCycle {
		case constants.BillingCycleMonthly:
			periodEnd = periodStart.AddDate(0, plan.BillingInterval, 0)
		case constants.BillingCycleYearly:
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
		validateReq := &couponDto.ValidateCouponRequest{
			Code:        req.CouponCode,
			UserID:      uint64(req.UserID),
			OrderAmount: amount + setupFee,
			PlanID:      uint64(req.SubscriptionPlanID),
			Currency:    plan.Currency,
		}

		validateResp, err := socs.couponService.ValidateCoupon(ctx, validateReq)
		if err != nil {
			logger.Error("Failed to validate coupon", logger.ErrorField(err), logger.String("coupon_code", req.CouponCode))
			return nil, fmt.Errorf("failed to validate coupon: %w", err)
		}

		if !validateResp.Valid {
			return nil, fmt.Errorf("coupon validation failed: %s", validateResp.Message)
		}

		couponDiscountAmount = validateResp.DiscountAmount
		discountAmount = couponDiscountAmount

		// Store coupon code for later use
		if couponTemp, err := socs.couponService.GetCouponByCode(ctx, req.CouponCode); err == nil {
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
		Status:             constants.OrderStatusPending,
		Amount:             amount,
		Currency:           plan.Currency,
		SetupFee:           setupFee,
		DiscountAmount:     discountAmount,
		TotalAmount:        totalAmount,
		BillingPeriodStart: billingPeriodStart,
		BillingPeriodEnd:   billingPeriodEnd,
		PaymentGateway:     finalPaymentGateway,
		PaymentMethod:      finalPaymentMethod,
		CouponCode:         req.CouponCode,
		Metadata:           req.Metadata,
	}

	// Set order status to confirmed (ready for invoice generation)
	order.Status = constants.OrderStatusConfirmed

	// Save order to database
	if err := tx.Create(order).Error; err != nil {
		tx.Rollback()
		logger.Error("Failed to create subscription order", logger.ErrorField(err))
		return nil, fmt.Errorf("failed to create subscription order: %w", err)
	}

	// Generate invoice from order
	invoiceReq := &invoiceDto.CreateInvoiceRequest{
		UserID:              req.UserID,
		SubscriptionOrderID: order.ID,
		Amount:              totalAmount,
		Currency:            plan.Currency,
		Description:         fmt.Sprintf("Subscription: %s", plan.Name),
		BillingName:         fmt.Sprintf("User %d", req.UserID),                // TODO: Get actual user name
		BillingEmail:        fmt.Sprintf("user%d@example.com", req.UserID),     // TODO: Get actual user email
		DueDate:             time.Now().AddDate(0, 0, 30).Format("2006-01-02"), // 30 days from now
	}

	invoice, err := socs.invoiceService.CreateInvoice(ctx, invoiceReq)
	if err != nil {
		tx.Rollback()
		logger.Error("Failed to create invoice", logger.ErrorField(err))
		return nil, fmt.Errorf("failed to create invoice: %w", err)
	}

	// Create payment order from invoice
	paymentReq := &paymentDto.CreatePaymentOrderRequest{
		UserID:              req.UserID,
		SubscriptionOrderID: &order.ID,   // Keep the order reference for service activation
		InvoiceID:           &invoice.ID, // Add invoice reference for invoice payment tracking
		Gateway:             finalPaymentGateway,
		PaymentMethod:       finalPaymentMethod,
		Amount:              totalAmount,
		Currency:            plan.Currency,
		Subject:             strings.ReplaceAll(plan.Name, " ", ""),
		Body:                fmt.Sprintf("Payment for invoice %s - %s subscription plan", invoice.InvoiceNumber, plan.Name),
		ReturnURL:           req.ReturnURL,
		ExpiredMinutes:      30, // 30 minutes expiration
	}

	paymentRecord, err := socs.paymentService.CreatePaymentOrder(ctx, paymentReq)
	if err != nil {
		// SECURITY: Rollback transaction if payment creation fails
		tx.Rollback()
		logger.Error("Failed to create payment order", logger.ErrorField(err))
		return nil, fmt.Errorf("failed to create payment order: %w", err)
	}

	// Commit transaction only after both order and payment are successfully created
	if err := tx.Commit().Error; err != nil {
		logger.Error("Failed to commit subscription order transaction", logger.ErrorField(err))
		return nil, fmt.Errorf("failed to commit subscription order transaction: %w", err)
	}

	// Record coupon usage if coupon was applied (outside of transaction since order is already committed)
	if appliedCouponID > 0 && couponDiscountAmount > 0 {
		// Update coupon service call to match correct signature: (ctx, couponID, userID, usedAmount, orderID)
		orderIDUint64 := uint64(order.ID)
		_, err := socs.couponService.UseCoupon(ctx, appliedCouponID, uint64(req.UserID), couponDiscountAmount, &orderIDUint64)
		if err != nil {
			logger.Error("Failed to record coupon usage",
				logger.ErrorField(err),
				logger.Any("coupon_id", appliedCouponID),
				logger.Uint("order_id", order.ID))
			// Don't fail the entire order creation, just log the error
		}
	}

	// Update order with payment information
	if err := socs.db.WithContext(ctx).Model(order).Updates(map[string]any{
		"transaction_id": paymentRecord.PaymentNo,
		"updated_at":     time.Now(),
	}).Error; err != nil {
		logger.Error("Failed to update order with payment info", logger.ErrorField(err))
		// Don't fail the entire process, just log the error
	}

	// Update payment method usage statistics if a saved payment method was used
	if usedPaymentMethodID != nil {
		// Note: We mark as successful here since the payment record was created successfully
		// The actual payment completion will be tracked separately in payment webhooks
		if _, err := socs.paymentMethodService.GetPaymentMethodUsageStats(ctx, req.UserID, *usedPaymentMethodID); err == nil {
			// Payment method exists and can be tracked - we could add this to a queue for async processing
			logger.Info("Payment method usage tracked",
				logger.Uint("payment_method_id", *usedPaymentMethodID),
				logger.Uint("user_id", req.UserID))
		}
	}

	logger.Info("Subscription order created successfully",
		logger.Uint("order_id", order.ID),
		logger.String("order_number", orderNumber),
		logger.Uint("user_id", req.UserID),
		logger.Uint("plan_id", req.SubscriptionPlanID),
		logger.Any("used_payment_method_id", usedPaymentMethodID))

	// Return response
	response := &dto.CreateSubscriptionOrderResponse{
		Order:         order.ToResponse(),
		Invoice:       invoiceDto.ToResponse(invoice),
		PaymentRecord: paymentDto.ToPaymentRecordUserResponse(paymentRecord),
		PaymentURL:    paymentRecord.PaymentURL,
		QRCodeURL:     paymentRecord.QRCodeURL,
		ExpiredAt:     *paymentRecord.ExpiredAt,
	}

	return response, nil
}

// CreateUpgradeDowngradeOrder creates an order for subscription upgrade or downgrade
func (socs *SubscriptionOrderCreationService) CreateUpgradeDowngradeOrder(ctx context.Context, userID uint, req *UpgradeDowngradeRequest) (*dto.CreateSubscriptionOrderResponse, error) {
	// Get current subscription
	currentSubscription, err := socs.userSubscriptionService.GetUserSubscription(ctx, req.CurrentSubscriptionID)
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
	currentPlan, err := socs.subscriptionPlanService.GetSubscriptionPlan(ctx, currentSubscription.SubscriptionPlanID)
	if err != nil {
		return nil, fmt.Errorf("failed to get current plan: %w", err)
	}

	// Get new plan
	newPlan, err := socs.subscriptionPlanService.GetSubscriptionPlan(ctx, req.NewPlanID)
	if err != nil {
		return nil, fmt.Errorf("invalid new subscription plan: %w", err)
	}

	if !newPlan.IsAvailableForPurchase() {
		return nil, fmt.Errorf("new subscription plan is not available for purchase")
	}

	// Determine order type based on plan prices
	var orderType string
	if newPlan.Price > currentPlan.Price {
		orderType = constants.OrderTypeUpgrade
	} else if newPlan.Price < currentPlan.Price {
		orderType = constants.OrderTypeDowngrade
	} else {
		return nil, fmt.Errorf("cannot change to plan with same price - use regular renewal instead")
	}

	// Create order request
	orderReq := &dto.CreateSubscriptionOrderRequest{
		UserID:             userID,
		SubscriptionPlanID: req.NewPlanID,
		OrderType:          orderType,
		PaymentGateway:     req.PaymentGateway,
		PaymentMethod:      req.PaymentMethod,
		ReturnURL:          req.ReturnURL,
		Metadata:           fmt.Sprintf(`{"current_subscription_id": %d, "apply_immediately": %t}`, req.CurrentSubscriptionID, getApplyImmediately(orderType, req.ApplyImmediately)),
	}

	// Create the order with proration calculation
	return socs.CreateSubscriptionOrder(ctx, orderReq)
}
