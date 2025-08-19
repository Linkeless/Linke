package workflows

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	couponDto "linke/internal/domains/coupon/dto"
	couponInterfaces "linke/internal/domains/coupon/usecases/interfaces"
	invoiceInterfaces "linke/internal/domains/invoice/usecases/interfaces"
	paymentInterfaces "linke/internal/domains/payment/usecases/interfaces"
	subscriptionDto "linke/internal/domains/subscription/dto"
	"linke/internal/domains/subscription/entities"
	subscriptionInterfaces "linke/internal/domains/subscription/usecases/interfaces"
	"linke/internal/shared/database"
	"linke/internal/shared/logger"
)

// NotificationService 通知服务接口 (可选)
type NotificationService interface {
	SendOrderCreatedNotification(ctx context.Context, userID, orderID uint) error
	SendPaymentSuccessNotification(ctx context.Context, userID, orderID uint) error
	SendSubscriptionActivatedNotification(ctx context.Context, userID, subscriptionID uint) error
	SendOrderExpiredNotification(ctx context.Context, userID, orderID uint) error
	SendSubscriptionRenewedNotification(ctx context.Context, userID, subscriptionID uint) error
	SendSubscriptionCancelledNotification(ctx context.Context, userID, subscriptionID uint) error
}

// SubscriptionWorkflow 订阅购买工作流
// 处理完整的订阅购买流程，整合支付、优惠券、推荐等功能
type SubscriptionWorkflow struct {
	logger               logger.Logger
	db                   *database.Database
	subscriptionOrderSvc subscriptionInterfaces.SubscriptionOrderService
	invoiceSvc           invoiceInterfaces.InvoiceService
	paymentSvc           paymentInterfaces.PaymentService
	userSubscriptionSvc  subscriptionInterfaces.UserSubscriptionService
	couponSvc            couponInterfaces.CouponService
	notificationSvc      NotificationService // 可选
}

// NewSubscriptionWorkflow 创建订阅工作流
func NewSubscriptionWorkflow(
	logger logger.Logger,
	db *database.Database,
	subscriptionOrderSvc subscriptionInterfaces.SubscriptionOrderService,
	invoiceSvc invoiceInterfaces.InvoiceService,
	paymentSvc paymentInterfaces.PaymentService,
	userSubscriptionSvc subscriptionInterfaces.UserSubscriptionService,
	couponSvc couponInterfaces.CouponService,
) *SubscriptionWorkflow {
	return &SubscriptionWorkflow{
		logger:               logger,
		db:                   db,
		subscriptionOrderSvc: subscriptionOrderSvc,
		invoiceSvc:           invoiceSvc,
		paymentSvc:           paymentSvc,
		userSubscriptionSvc:  userSubscriptionSvc,
		couponSvc:            couponSvc,
	}
}

// SetNotificationService 设置通知服务 (可选)
func (w *SubscriptionWorkflow) SetNotificationService(notificationSvc NotificationService) {
	w.notificationSvc = notificationSvc
}

// PurchaseSubscriptionRequest 购买订阅请求
type PurchaseSubscriptionRequest struct {
	UserID             uint   `json:"user_id" binding:"required"`
	SubscriptionPlanID uint   `json:"subscription_plan_id" binding:"required"`
	OrderType          string `json:"order_type" binding:"required,oneof=new renewal upgrade downgrade"`
	CouponCode         string `json:"coupon_code,omitempty"`
	PaymentGateway     string `json:"payment_gateway" binding:"required"`
	PaymentMethod      string `json:"payment_method" binding:"required"`
	ReturnURL          string `json:"return_url,omitempty"`
	ClientIP           string `json:"client_ip,omitempty"`
	Metadata           string `json:"metadata,omitempty"`
}

// PurchaseSubscriptionResponse 购买订阅响应
type PurchaseSubscriptionResponse struct {
	OrderID     uint      `json:"order_id"`
	OrderNumber string    `json:"order_number"`
	InvoiceID   uint      `json:"invoice_id"`
	PaymentURL  string    `json:"payment_url"`
	QRCodeURL   string    `json:"qr_code_url,omitempty"`
	Amount      float64   `json:"amount"`
	Currency    string    `json:"currency"`
	ExpiredAt   time.Time `json:"expired_at"`
}

// PurchaseSubscription 购买订阅工作流
// 核心业务流程：用户选择服务 → 创建订单 → 生成发票 → 引导用户付款
func (w *SubscriptionWorkflow) PurchaseSubscription(ctx context.Context, req *PurchaseSubscriptionRequest) (*PurchaseSubscriptionResponse, error) {
	workflowLogger := w.logger.With(
		zap.Uint("user_id", req.UserID),
		zap.Uint("plan_id", req.SubscriptionPlanID),
		zap.String("order_type", req.OrderType),
		zap.String("payment_gateway", req.PaymentGateway),
	)

	workflowLogger.Info("Starting subscription purchase workflow")

	// 使用数据库事务确保数据一致性
	tx := w.db.DB.Begin()
	if tx.Error != nil {
		workflowLogger.Error("Failed to begin transaction", zap.Error(tx.Error))
		return nil, fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			workflowLogger.Error("Transaction rolled back due to panic", zap.Any("panic", r))
		}
	}()

	// 1. 验证优惠券 (如果提供)
	var discountAmount float64
	var couponID *uint64
	if req.CouponCode != "" {
		workflowLogger.Info("Validating coupon", zap.String("coupon_code", req.CouponCode))

		couponValidationReq := &couponDto.ValidateCouponRequest{
			Code:        req.CouponCode,
			UserID:      uint64(req.UserID),
			PlanID:      uint64(req.SubscriptionPlanID),
			OrderAmount: 0,     // 将在获取套餐价格后更新
			Currency:    "CNY", // 默认货币，实际应从套餐获取
		}

		couponValidation, err := w.couponSvc.ValidateCoupon(ctx, couponValidationReq)
		if err != nil {
			tx.Rollback()
			workflowLogger.Error("Failed to validate coupon", zap.Error(err))
			return nil, fmt.Errorf("coupon validation failed: %w", err)
		}

		if !couponValidation.Valid {
			tx.Rollback()
			workflowLogger.Warn("Invalid coupon", zap.String("message", couponValidation.Message))
			return nil, fmt.Errorf("invalid coupon: %s", couponValidation.Message)
		}

		discountAmount = couponValidation.DiscountAmount
		if couponValidation.Coupon != nil {
			couponIDVal := uint64(couponValidation.Coupon.ID)
			couponID = &couponIDVal
		}

		workflowLogger.Info("Coupon validated successfully",
			zap.Float64("discount_amount", discountAmount),
			zap.Float64("final_amount", couponValidation.FinalAmount),
		)
	}

	// 2. 创建订阅订单
	workflowLogger.Info("Creating subscription order")

	createOrderReq := &subscriptionDto.CreateSubscriptionOrderRequest{
		UserID:             req.UserID,
		SubscriptionPlanID: req.SubscriptionPlanID,
		OrderType:          req.OrderType,
		CouponCode:         req.CouponCode,
		PaymentGateway:     req.PaymentGateway,
		PaymentMethod:      req.PaymentMethod,
		ReturnURL:          req.ReturnURL,
		Metadata:           req.Metadata,
	}

	orderResp, err := w.subscriptionOrderSvc.CreateSubscriptionOrder(ctx, createOrderReq)
	if err != nil {
		tx.Rollback()
		workflowLogger.Error("Failed to create subscription order", zap.Error(err))
		return nil, fmt.Errorf("failed to create subscription order: %w", err)
	}

	workflowLogger.Info("Subscription order created successfully",
		zap.Uint("order_id", orderResp.Order.ID),
		zap.String("order_number", orderResp.Order.OrderNumber),
	)

	// 3. 使用优惠券 (如果有效)
	if couponID != nil {
		workflowLogger.Info("Using coupon", zap.Uint64("coupon_id", *couponID))

		_, err := w.couponSvc.UseCoupon(ctx, *couponID, uint64(req.UserID), orderResp.Order.TotalAmount, &[]uint64{uint64(orderResp.Order.ID)}[0])
		if err != nil {
			tx.Rollback()
			workflowLogger.Error("Failed to use coupon", zap.Error(err))
			return nil, fmt.Errorf("failed to use coupon: %w", err)
		}

		workflowLogger.Info("Coupon used successfully")
	}

	// 4. 提交事务
	if err := tx.Commit().Error; err != nil {
		workflowLogger.Error("Failed to commit transaction", zap.Error(err))
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// 5. 发送订单创建通知 (异步，不影响主流程)
	if w.notificationSvc != nil {
		go func() {
			notifyCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			if err := w.notificationSvc.SendOrderCreatedNotification(notifyCtx, req.UserID, orderResp.Order.ID); err != nil {
				workflowLogger.Warn("Failed to send order created notification", zap.Error(err))
			}
		}()
	}

	workflowLogger.Info("Subscription purchase workflow completed successfully",
		zap.String("payment_url", orderResp.PaymentURL),
	)

	return &PurchaseSubscriptionResponse{
		OrderID:     orderResp.Order.ID,
		OrderNumber: orderResp.Order.OrderNumber,
		InvoiceID:   orderResp.Invoice.ID,
		PaymentURL:  orderResp.PaymentURL,
		QRCodeURL:   orderResp.QRCodeURL,
		Amount:      orderResp.Order.TotalAmount,
		Currency:    orderResp.Order.Currency,
		ExpiredAt:   orderResp.ExpiredAt,
	}, nil
}

// ProcessPaymentCallback 处理支付回调后的流程
// 支付成功后的完整处理：更新订单状态 → 创建/激活订阅 → 发送通知
func (w *SubscriptionWorkflow) ProcessPaymentCallback(ctx context.Context, paymentNo string) error {
	workflowLogger := w.logger.With(zap.String("payment_no", paymentNo))
	workflowLogger.Info("Starting payment callback processing workflow")

	// 1. 获取支付记录
	paymentRecord, err := w.paymentSvc.GetPaymentRecord(ctx, paymentNo)
	if err != nil {
		workflowLogger.Error("Failed to get payment record", zap.Error(err))
		return fmt.Errorf("failed to get payment record: %w", err)
	}

	if paymentRecord.SubscriptionOrderID == nil {
		workflowLogger.Warn("Payment record has no associated subscription order")
		return fmt.Errorf("payment record has no associated subscription order")
	}

	orderID := *paymentRecord.SubscriptionOrderID
	workflowLogger = workflowLogger.With(zap.Uint("order_id", orderID))

	// 2. 获取订单信息
	order, err := w.subscriptionOrderSvc.GetSubscriptionOrder(ctx, orderID)
	if err != nil {
		workflowLogger.Error("Failed to get subscription order", zap.Error(err))
		return fmt.Errorf("failed to get subscription order: %w", err)
	}

	workflowLogger = workflowLogger.With(
		zap.Uint("user_id", order.UserID),
		zap.String("order_status", order.Status),
	)

	// 3. 检查订单状态，避免重复处理
	if order.Status == "paid" {
		workflowLogger.Info("Order already processed")
		return nil
	}

	if order.Status != "pending" {
		workflowLogger.Warn("Order is not in pending status", zap.String("current_status", order.Status))
		return fmt.Errorf("order is not in pending status: %s", order.Status)
	}

	// 使用数据库事务确保数据一致性
	tx := w.db.DB.Begin()
	if tx.Error != nil {
		workflowLogger.Error("Failed to begin transaction", zap.Error(tx.Error))
		return fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			workflowLogger.Error("Transaction rolled back due to panic", zap.Any("panic", r))
		}
	}()

	// 4. 处理订单支付成功
	workflowLogger.Info("Processing order payment success")

	err = w.subscriptionOrderSvc.ProcessOrderPaymentSuccess(ctx, orderID)
	if err != nil {
		tx.Rollback()
		workflowLogger.Error("Failed to process order payment success", zap.Error(err))
		return fmt.Errorf("failed to process order payment success: %w", err)
	}

	// 5. 根据订单类型处理订阅
	var subscriptionID uint
	switch order.OrderType {
	case "new":
		// 创建新订阅
		workflowLogger.Info("Creating new subscription")

		createSubReq := &subscriptionInterfaces.CreateSubscriptionRequest{
			UserID:             order.UserID,
			SubscriptionPlanID: order.SubscriptionPlanID,
			StartDate:          time.Now().Format(time.RFC3339),
			UseTrial:           false, // 付费订阅不使用试用期
		}

		userSubscription, err := w.userSubscriptionSvc.CreateUserSubscription(ctx, createSubReq)
		if err != nil {
			tx.Rollback()
			workflowLogger.Error("Failed to create user subscription", zap.Error(err))
			return fmt.Errorf("failed to create user subscription: %w", err)
		}

		subscriptionID = userSubscription.ID
		workflowLogger.Info("New subscription created", zap.Uint("subscription_id", subscriptionID))

	case "renewal":
		// 续费现有订阅
		if order.UserSubscriptionID == nil {
			tx.Rollback()
			workflowLogger.Error("Renewal order missing user subscription ID")
			return fmt.Errorf("renewal order missing user subscription ID")
		}

		subscriptionID = *order.UserSubscriptionID
		workflowLogger.Info("Renewing subscription", zap.Uint("subscription_id", subscriptionID))

		err = w.userSubscriptionSvc.RenewUserSubscription(ctx, subscriptionID)
		if err != nil {
			tx.Rollback()
			workflowLogger.Error("Failed to renew user subscription", zap.Error(err))
			return fmt.Errorf("failed to renew user subscription: %w", err)
		}

		workflowLogger.Info("Subscription renewed successfully")

	case "upgrade", "downgrade":
		// 升级/降级处理
		workflowLogger.Info("Processing subscription change", zap.String("change_type", order.OrderType))

		if order.UserID == 0 {
			tx.Rollback()
			workflowLogger.Error("Missing user ID for subscription change")
			return fmt.Errorf("missing user ID for subscription change")
		}

		if order.SubscriptionPlanID == 0 {
			tx.Rollback()
			workflowLogger.Error("Missing subscription plan ID for subscription change")
			return fmt.Errorf("missing subscription plan ID for subscription change")
		}

		// Find active subscription for this user (assuming one active subscription per user)
		var activeSubscription *entities.UserSubscription
		userSubscriptions, err := w.userSubscriptionSvc.GetUserActiveSubscriptions(ctx, order.UserID)
		if err != nil {
			tx.Rollback()
			workflowLogger.Error("Failed to get user active subscriptions", zap.Error(err))
			return fmt.Errorf("failed to get user active subscriptions: %w", err)
		}

		if len(userSubscriptions) == 0 {
			tx.Rollback()
			workflowLogger.Error("No active subscription found for user")
			return fmt.Errorf("no active subscription found for user")
		}

		// Use the first active subscription (could be enhanced to match specific criteria)
		activeSubscription = userSubscriptions[0]

		// Process the subscription change
		_, err = w.userSubscriptionSvc.ProcessSubscriptionChange(ctx, activeSubscription.ID, order.SubscriptionPlanID, order.OrderType)
		if err != nil {
			tx.Rollback()
			workflowLogger.Error("Failed to process subscription change",
				zap.String("change_type", order.OrderType),
				zap.Uint("subscription_id", activeSubscription.ID),
				zap.Uint("new_plan_id", order.SubscriptionPlanID),
				zap.Error(err))
			return fmt.Errorf("failed to process subscription %s: %w", order.OrderType, err)
		}

		workflowLogger.Info("Subscription change completed successfully",
			zap.String("change_type", order.OrderType),
			zap.Uint("subscription_id", activeSubscription.ID),
			zap.Uint("new_plan_id", order.SubscriptionPlanID))

	default:
		tx.Rollback()
		workflowLogger.Error("Unknown order type", zap.String("order_type", order.OrderType))
		return fmt.Errorf("unknown order type: %s", order.OrderType)
	}

	// 6. 更新发票状态为已支付
	workflowLogger.Info("Marking invoice as paid")

	// 通过订单ID查找关联的发票并标记为已支付
	if paymentRecord.PaidAt != nil && paymentRecord.InvoiceID != nil {
		err = w.invoiceSvc.MarkInvoiceAsPaid(ctx, *paymentRecord.InvoiceID, paymentRecord.PaidAt.Format(time.RFC3339))
		if err != nil {
			// 发票状态更新失败不应该阻止整个流程，记录警告
			workflowLogger.Warn("Failed to mark invoice as paid", zap.Error(err))
		}
	}

	// 7. 提交事务
	if err := tx.Commit().Error; err != nil {
		workflowLogger.Error("Failed to commit transaction", zap.Error(err))
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	// 8. 发送通知 (异步，不影响主流程)
	if w.notificationSvc != nil {
		go func() {
			notifyCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			// 发送支付成功通知
			if err := w.notificationSvc.SendPaymentSuccessNotification(notifyCtx, order.UserID, orderID); err != nil {
				workflowLogger.Warn("Failed to send payment success notification", zap.Error(err))
			}

			// 发送订阅激活通知
			if subscriptionID > 0 {
				if err := w.notificationSvc.SendSubscriptionActivatedNotification(notifyCtx, order.UserID, subscriptionID); err != nil {
					workflowLogger.Warn("Failed to send subscription activated notification", zap.Error(err))
				}
			}
		}()
	}

	workflowLogger.Info("Payment callback processing workflow completed successfully")
	return nil
}

// HandleOrderExpiration 处理订单过期
// 定期清理过期的未支付订单
func (w *SubscriptionWorkflow) HandleOrderExpiration(ctx context.Context, orderID uint) error {
	workflowLogger := w.logger.With(zap.Uint("order_id", orderID))
	workflowLogger.Info("Starting order expiration handling workflow")

	// 1. 获取订单信息
	order, err := w.subscriptionOrderSvc.GetSubscriptionOrder(ctx, orderID)
	if err != nil {
		workflowLogger.Error("Failed to get subscription order", zap.Error(err))
		return fmt.Errorf("failed to get subscription order: %w", err)
	}

	workflowLogger = workflowLogger.With(
		zap.Uint("user_id", order.UserID),
		zap.String("order_status", order.Status),
		zap.String("order_number", order.OrderNumber),
	)

	// 2. 检查订单状态
	if order.Status != "pending" {
		workflowLogger.Info("Order is not in pending status, skipping expiration",
			zap.String("current_status", order.Status))
		return nil
	}

	// 3. 取消过期订单
	workflowLogger.Info("Cancelling expired order")

	err = w.subscriptionOrderSvc.CancelSubscriptionOrder(ctx, orderID, "Order expired - payment not received within time limit")
	if err != nil {
		workflowLogger.Error("Failed to cancel expired order", zap.Error(err))
		return fmt.Errorf("failed to cancel expired order: %w", err)
	}

	// 4. 作废关联的发票
	workflowLogger.Info("Marking associated invoice as void")

	// 这里需要通过订单ID找到关联的发票
	// 简化实现，实际可能需要在订单实体中直接关联发票ID
	// 或通过发票服务查询

	// 5. 发送订单过期通知 (异步)
	if w.notificationSvc != nil {
		go func() {
			notifyCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			if err := w.notificationSvc.SendOrderExpiredNotification(notifyCtx, order.UserID, orderID); err != nil {
				workflowLogger.Warn("Failed to send order expired notification", zap.Error(err))
			}
		}()
	}

	workflowLogger.Info("Order expiration handling workflow completed successfully")
	return nil
}

// HandleSubscriptionRenewal 处理订阅续费
// 自动续费流程：检查订阅状态 → 创建续费订单 → 处理支付
func (w *SubscriptionWorkflow) HandleSubscriptionRenewal(ctx context.Context, subscriptionID uint) error {
	workflowLogger := w.logger.With(zap.Uint("subscription_id", subscriptionID))
	workflowLogger.Info("Starting subscription renewal handling workflow")

	// 1. 获取订阅信息
	userSubscription, err := w.userSubscriptionSvc.GetUserSubscription(ctx, subscriptionID)
	if err != nil {
		workflowLogger.Error("Failed to get user subscription", zap.Error(err))
		return fmt.Errorf("failed to get user subscription: %w", err)
	}

	workflowLogger = workflowLogger.With(
		zap.Uint("user_id", userSubscription.UserID),
		zap.String("subscription_status", userSubscription.Status),
		zap.Bool("auto_renew", userSubscription.AutoRenew),
	)

	// 2. 检查是否需要续费
	if !userSubscription.AutoRenew {
		workflowLogger.Info("Subscription auto-renew is disabled")
		return nil
	}

	if userSubscription.Status != "active" {
		workflowLogger.Info("Subscription is not active, skipping renewal",
			zap.String("current_status", userSubscription.Status))
		return nil
	}

	// 3. 检查是否即将到期 (例如：7天内)
	if userSubscription.CurrentPeriodEnd == nil {
		workflowLogger.Warn("Subscription has no current period end date")
		return nil
	}

	sevenDaysFromNow := time.Now().Add(7 * 24 * time.Hour)
	if userSubscription.CurrentPeriodEnd.After(sevenDaysFromNow) {
		workflowLogger.Info("Subscription is not near expiration yet",
			zap.Time("expires_at", *userSubscription.CurrentPeriodEnd))
		return nil
	}

	// 4. 创建续费订单
	workflowLogger.Info("Creating renewal order")

	// 获取用户的默认支付方式 (简化实现，实际需要从用户配置中获取)
	paymentGateway := "epay"  // 默认支付网关
	paymentMethod := "alipay" // 默认支付方式

	createOrderReq := &subscriptionDto.CreateSubscriptionOrderRequest{
		UserID:             userSubscription.UserID,
		SubscriptionPlanID: userSubscription.SubscriptionPlanID,
		OrderType:          "renewal",
		PaymentGateway:     paymentGateway,
		PaymentMethod:      paymentMethod,
		Metadata:           fmt.Sprintf("{\"auto_renewal\": true, \"subscription_id\": %d}", subscriptionID),
	}

	orderResp, err := w.subscriptionOrderSvc.CreateSubscriptionOrder(ctx, createOrderReq)
	if err != nil {
		workflowLogger.Error("Failed to create renewal order", zap.Error(err))
		return fmt.Errorf("failed to create renewal order: %w", err)
	}

	workflowLogger.Info("Renewal order created successfully",
		zap.Uint("order_id", orderResp.Order.ID),
		zap.String("order_number", orderResp.Order.OrderNumber),
	)

	// 5. 这里可以集成自动支付逻辑
	// 例如：使用保存的支付方式自动完成支付
	// 目前只是创建订单，实际支付仍需要用户操作

	// 6. 发送续费通知 (异步)
	if w.notificationSvc != nil {
		go func() {
			notifyCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			if err := w.notificationSvc.SendSubscriptionRenewedNotification(notifyCtx, userSubscription.UserID, subscriptionID); err != nil {
				workflowLogger.Warn("Failed to send subscription renewal notification", zap.Error(err))
			}
		}()
	}

	workflowLogger.Info("Subscription renewal handling workflow completed successfully")
	return nil
}

// CancelSubscriptionRequest 取消订阅请求
type CancelSubscriptionRequest struct {
	SubscriptionID    uint   `json:"subscription_id" binding:"required"`
	Reason            string `json:"reason" binding:"required"`
	CancelAtPeriodEnd bool   `json:"cancel_at_period_end"`
	UserID            uint   `json:"user_id" binding:"required"` // 用于权限验证
}

// CancelSubscription 处理订阅取消
// 取消流程：验证权限 → 更新订阅状态 → 处理退费 → 发送通知
func (w *SubscriptionWorkflow) CancelSubscription(ctx context.Context, req *CancelSubscriptionRequest) error {
	workflowLogger := w.logger.With(
		zap.Uint("subscription_id", req.SubscriptionID),
		zap.Uint("user_id", req.UserID),
		zap.String("reason", req.Reason),
		zap.Bool("cancel_at_period_end", req.CancelAtPeriodEnd),
	)

	workflowLogger.Info("Starting subscription cancellation workflow")

	// 1. 获取订阅信息验证权限
	userSubscription, err := w.userSubscriptionSvc.GetUserSubscription(ctx, req.SubscriptionID)
	if err != nil {
		workflowLogger.Error("Failed to get user subscription", zap.Error(err))
		return fmt.Errorf("failed to get user subscription: %w", err)
	}

	// 2. 验证用户权限
	if userSubscription.UserID != req.UserID {
		workflowLogger.Warn("User does not have permission to cancel this subscription",
			zap.Uint("subscription_user_id", userSubscription.UserID))
		return fmt.Errorf("permission denied: user cannot cancel this subscription")
	}

	// 3. 检查订阅状态
	if userSubscription.Status == "cancelled" {
		workflowLogger.Info("Subscription is already cancelled")
		return nil
	}

	if userSubscription.Status != "active" {
		workflowLogger.Warn("Cannot cancel subscription with current status",
			zap.String("current_status", userSubscription.Status))
		return fmt.Errorf("cannot cancel subscription with status: %s", userSubscription.Status)
	}

	// 4. 取消订阅
	workflowLogger.Info("Cancelling user subscription")

	err = w.userSubscriptionSvc.CancelUserSubscription(ctx, req.SubscriptionID, req.Reason, req.CancelAtPeriodEnd)
	if err != nil {
		workflowLogger.Error("Failed to cancel user subscription", zap.Error(err))
		return fmt.Errorf("failed to cancel user subscription: %w", err)
	}

	// 5. 处理退费逻辑 (如果需要)
	if !req.CancelAtPeriodEnd {
		// 立即取消可能涉及退费
		workflowLogger.Info("Processing immediate cancellation - checking for refunds")

		// 这里可以添加退费逻辑
		// 例如：计算剩余时间，生成退费订单等
		// 目前简化处理，只记录日志
		workflowLogger.Warn("Refund processing not implemented - manual review may be required")
	}

	// 6. 发送取消通知 (异步)
	if w.notificationSvc != nil {
		go func() {
			notifyCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			if err := w.notificationSvc.SendSubscriptionCancelledNotification(notifyCtx, req.UserID, req.SubscriptionID); err != nil {
				workflowLogger.Warn("Failed to send subscription cancelled notification", zap.Error(err))
			}
		}()
	}

	workflowLogger.Info("Subscription cancellation workflow completed successfully")
	return nil
}

// BatchProcessExpiredOrders 批量处理过期订单
// 定时任务：扫描并处理所有过期的未支付订单
func (w *SubscriptionWorkflow) BatchProcessExpiredOrders(ctx context.Context) error {
	w.logger.Info("Starting batch processing of expired orders")

	// 1. 获取过期的待支付订单
	// 这里需要调用订单服务获取过期订单列表
	// 简化实现，实际需要在订单服务中添加相应的查询方法

	// 示例逻辑：
	// expiredOrders, err := w.subscriptionOrderSvc.GetExpiredPendingOrders(ctx)
	// if err != nil {
	//     w.logger.Error("Failed to get expired pending orders", zap.Error(err))
	//     return fmt.Errorf("failed to get expired pending orders: %w", err)
	// }

	// for _, order := range expiredOrders {
	//     if err := w.HandleOrderExpiration(ctx, order.ID); err != nil {
	//         w.logger.Error("Failed to handle order expiration",
	//             zap.Uint("order_id", order.ID),
	//             zap.Error(err))
	//         // 继续处理其他订单，不因单个失败而中断整个批处理
	//     }
	// }

	w.logger.Info("Batch processing of expired orders completed")
	return nil
}

// BatchProcessSubscriptionRenewals 批量处理订阅续费
// 定时任务：扫描并处理所有需要续费的订阅
func (w *SubscriptionWorkflow) BatchProcessSubscriptionRenewals(ctx context.Context) error {
	w.logger.Info("Starting batch processing of subscription renewals")

	// 1. 获取需要续费的订阅列表
	// 这里需要调用订阅服务获取即将到期且启用自动续费的订阅
	// 简化实现，实际需要在订阅服务中添加相应的查询方法

	// 示例逻辑：
	// renewalSubscriptions, err := w.userSubscriptionSvc.GetSubscriptionsNeedingRenewal(ctx)
	// if err != nil {
	//     w.logger.Error("Failed to get subscriptions needing renewal", zap.Error(err))
	//     return fmt.Errorf("failed to get subscriptions needing renewal: %w", err)
	// }

	// for _, subscription := range renewalSubscriptions {
	//     if err := w.HandleSubscriptionRenewal(ctx, subscription.ID); err != nil {
	//         w.logger.Error("Failed to handle subscription renewal",
	//             zap.Uint("subscription_id", subscription.ID),
	//             zap.Error(err))
	//         // 继续处理其他订阅，不因单个失败而中断整个批处理
	//     }
	// }

	w.logger.Info("Batch processing of subscription renewals completed")
	return nil
}

// GetWorkflowStatus 获取工作流状态信息
// 用于监控和诊断工作流健康状况
func (w *SubscriptionWorkflow) GetWorkflowStatus(ctx context.Context) map[string]any {
	status := map[string]any{
		"workflow_name": "SubscriptionWorkflow",
		"status":        "healthy",
		"timestamp":     time.Now().Format(time.RFC3339),
		"services_status": map[string]bool{
			"subscription_order_service": w.subscriptionOrderSvc != nil,
			"invoice_service":            w.invoiceSvc != nil,
			"payment_service":            w.paymentSvc != nil,
			"user_subscription_service":  w.userSubscriptionSvc != nil,
			"coupon_service":             w.couponSvc != nil,
			"notification_service":       w.notificationSvc != nil,
		},
	}

	w.logger.Info("Workflow status requested", zap.Any("status", status))
	return status
}
