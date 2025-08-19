package implementations

import (
	"context"
	"fmt"
	"slices"
	"time"

	"linke/internal/domains/subscription/constants"
	"linke/internal/domains/subscription/entities"
	"linke/internal/domains/subscription/usecases/interfaces"
	"linke/internal/shared/logger"

	"gorm.io/gorm"
)

// SubscriptionOrderOperationsService handles order status updates, refunds, and other operations
type SubscriptionOrderOperationsService struct {
	*SubscriptionOrderServiceBase
}

// NewSubscriptionOrderOperationsService creates a new order operations service
func NewSubscriptionOrderOperationsService(base *SubscriptionOrderServiceBase) *SubscriptionOrderOperationsService {
	return &SubscriptionOrderOperationsService{
		SubscriptionOrderServiceBase: base,
	}
}

// ProcessOrderPaymentSuccess processes successful payment for an order
func (soos *SubscriptionOrderOperationsService) ProcessOrderPaymentSuccess(ctx context.Context, orderID uint) error {
	// Start transaction
	tx := soos.db.WithContext(ctx).Begin()
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
	if order.Status == constants.OrderStatusPaid {
		tx.Rollback()
		return fmt.Errorf("order %d is already paid", orderID)
	}
	if order.Status == constants.OrderStatusCancelled || order.Status == constants.OrderStatusFailed {
		tx.Rollback()
		return fmt.Errorf("order %d cannot be processed in status: %s", orderID, order.Status)
	}

	// Only allow processing from pending or confirmed status
	if order.Status != constants.OrderStatusPending && order.Status != constants.OrderStatusConfirmed {
		tx.Rollback()
		return fmt.Errorf("order %d can only be processed from pending or confirmed status, current status: %s", orderID, order.Status)
	}

	// Get subscription plan
	var plan entities.SubscriptionPlan
	if err := tx.First(&plan, order.SubscriptionPlanID).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get subscription plan: %w", err)
	}

	// Update order status
	if err := tx.Model(&order).Updates(map[string]any{
		"status":     constants.OrderStatusPaid,
		"paid_at":    time.Now(),
		"updated_at": time.Now(),
	}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update order status: %w", err)
	}

	// Create or update user subscription based on order type
	switch order.OrderType {
	case constants.OrderTypeNew:
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

		subscription, err := soos.userSubscriptionService.CreateUserSubscription(ctx, subscriptionReq)
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

	case constants.OrderTypeRenewal:
		// Find and renew existing subscription
		var subscription entities.UserSubscription
		if err := tx.Where("user_id = ? AND subscription_plan_id = ?", order.UserID, order.SubscriptionPlanID).
			First(&subscription).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to find subscription for renewal: %w", err)
		}

		// Renew subscription
		if err := soos.userSubscriptionService.RenewUserSubscription(ctx, subscription.ID); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to renew subscription: %w", err)
		}

		logger.Info("Subscription renewed for paid order",
			logger.Uint("order_id", order.ID),
			logger.Uint("subscription_id", subscription.ID))

	case constants.OrderTypeUpgrade, constants.OrderTypeDowngrade:
		// Find existing subscription for this user and current active plan
		var currentSubscription entities.UserSubscription
		if err := tx.Where("user_id = ? AND status IN (?)",
			order.UserID, []string{constants.UserSubscriptionStatusActive, constants.UserSubscriptionStatusTrial}).
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
			if err := tx.Model(&order).Updates(map[string]any{
				"discount_amount": order.DiscountAmount + proratedCredit,
				"total_amount":    updatedTotalAmount,
				"metadata":        fmt.Sprintf(`{"proration_credit": %.2f, "original_total": %.2f}`, proratedCredit, order.TotalAmount),
			}).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update order with proration: %w", err)
			}
		}

		// Cancel current subscription (mark as cancelled but keep active until period end)
		if err := tx.Model(&currentSubscription).Updates(map[string]any{
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
		if order.OrderType == constants.OrderTypeUpgrade {
			subscriptionReq.StartDate = time.Now().Format(time.RFC3339)
		} else if currentSubscription.CurrentPeriodEnd != nil {
			subscriptionReq.StartDate = currentSubscription.CurrentPeriodEnd.Format(time.RFC3339)
		}

		// Inherit server group access from current subscription
		if currentSubscription.ServerGroupIDs != "" {
			subscriptionReq.ServerGroupIDs = currentSubscription.GetServerGroupIDs()
		}

		newSubscription, err := soos.userSubscriptionService.CreateUserSubscription(ctx, subscriptionReq)
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
		if order.OrderType == constants.OrderTypeUpgrade {
			// Deactivate current subscription immediately
			if err := tx.Model(&currentSubscription).Update("status", constants.UserSubscriptionStatusCancelled).Error; err != nil {
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

// UpdateOrderStatus updates order status with admin notes and payment verification
func (soos *SubscriptionOrderOperationsService) UpdateOrderStatus(ctx context.Context, orderID uint, status string, adminID uint, notes, reason string) (*entities.SubscriptionOrder, error) {
	return soos.UpdateOrderStatusWithEvidence(ctx, orderID, &UpdateOrderStatusRequest{
		Status: status,
		Notes:  notes,
		Reason: reason,
	}, adminID)
}

// UpdateOrderStatusWithEvidence updates order status with payment evidence verification
func (soos *SubscriptionOrderOperationsService) UpdateOrderStatusWithEvidence(ctx context.Context, orderID uint, req *UpdateOrderStatusRequest, adminID uint) (*entities.SubscriptionOrder, error) {
	// Start transaction
	tx := soos.db.WithContext(ctx).Begin()
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
	if req.Status == constants.OrderStatusPaid {
		// Require payment evidence for manual payment confirmation
		if req.PaymentEvidence == nil {
			tx.Rollback()
			return nil, fmt.Errorf("payment evidence is required when marking order as paid")
		}

		// Validate payment evidence
		if err := soos.validatePaymentEvidence(&order, req.PaymentEvidence); err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("payment evidence validation failed: %w", err)
		}
	}

	// SECURITY: Require admin confirmation for critical operations
	if soos.isCriticalOperation(order.Status, req.Status) && !req.AdminConfirm {
		tx.Rollback()
		return nil, fmt.Errorf("admin confirmation is required for this status change")
	}

	// Prepare update data
	updateData := map[string]any{
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

	if req.Status == constants.OrderStatusPaid && req.PaymentEvidence != nil {
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
	if req.Status == constants.OrderStatusPaid && order.PaidAt == nil {
		updateData["paid_at"] = time.Now()
	}

	// Update order
	if err := tx.Model(&order).Updates(updateData).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to update order status: %w", err)
	}

	// If changing to paid status, process the order
	if req.Status == constants.OrderStatusPaid && order.Status != constants.OrderStatusPaid {
		if err := soos.ProcessOrderPaymentSuccess(ctx, orderID); err != nil {
			logger.Error("Failed to process order after status update",
				logger.ErrorField(err),
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
	if err := soos.db.WithContext(ctx).First(&order, orderID).Error; err != nil {
		return nil, fmt.Errorf("failed to reload order: %w", err)
	}

	return &order, nil
}

// validatePaymentEvidence validates payment evidence for manual payment confirmation
func (soos *SubscriptionOrderOperationsService) validatePaymentEvidence(order *entities.SubscriptionOrder, evidence *PaymentEvidence) error {
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
	if err := soos.db.Where("transaction_id = ? AND id != ?", evidence.TransactionID, order.ID).First(&existingOrder).Error; err == nil {
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
func (soos *SubscriptionOrderOperationsService) isCriticalOperation(oldStatus, newStatus string) bool {
	criticalChanges := map[string][]string{
		constants.OrderStatusPending:   {constants.OrderStatusPaid},                                     // Manual payment confirmation
		constants.OrderStatusConfirmed: {constants.OrderStatusPaid},                                     // Manual payment confirmation for confirmed orders
		constants.OrderStatusFailed:    {constants.OrderStatusPaid},                                     // Retry failed payment
		constants.OrderStatusCancelled: {constants.OrderStatusPaid, constants.OrderStatusPending},       // Reactivate cancelled order
		constants.OrderStatusPaid:      {constants.OrderStatusRefunded, constants.OrderStatusCancelled}, // Reverse paid order
	}

	allowedTargets, exists := criticalChanges[oldStatus]
	if !exists {
		return false
	}

	return slices.Contains(allowedTargets, newStatus)
}

// ProcessRefund processes a refund for an order
func (soos *SubscriptionOrderOperationsService) ProcessRefund(ctx context.Context, req *ProcessRefundRequest) (*entities.SubscriptionOrder, error) {
	// Start transaction
	tx := soos.db.WithContext(ctx).Begin()
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
	updateData := map[string]any{
		"refund_amount": order.RefundAmount + req.Amount,
		"refund_reason": req.Reason,
		"refunded_at":   time.Now(),
		"updated_at":    time.Now(),
	}

	// If full refund, update status
	if order.RefundAmount+req.Amount >= order.TotalAmount {
		updateData["status"] = constants.OrderStatusRefunded
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
	if err := soos.db.WithContext(ctx).First(&order, req.OrderID).Error; err != nil {
		return nil, fmt.Errorf("failed to reload order: %w", err)
	}

	return &order, nil
}

// CancelSubscriptionOrder cancels a subscription order
func (soos *SubscriptionOrderOperationsService) CancelSubscriptionOrder(ctx context.Context, orderID uint, reason string) error {
	// Start transaction
	tx := soos.db.WithContext(ctx).Begin()
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
	if order.Status != constants.OrderStatusPending {
		tx.Rollback()
		return fmt.Errorf("order cannot be cancelled in status: %s", order.Status)
	}

	// Update order status
	updateData := map[string]any{
		"status":              constants.OrderStatusCancelled,
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
