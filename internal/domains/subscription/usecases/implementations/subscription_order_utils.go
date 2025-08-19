package implementations

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"linke/internal/domains/subscription/constants"
	"linke/internal/domains/subscription/entities"
	"linke/internal/shared/logger"
)

// ===============================================================================
// SHARED TYPES - Used across multiple subscription order service files
// ===============================================================================

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

// ProcessRefundRequest represents the request to process a refund
type ProcessRefundRequest struct {
	OrderID      uint    `json:"order_id"`
	Amount       float64 `json:"amount"`
	Reason       string  `json:"reason"`
	RefundMethod string  `json:"refund_method,omitempty"`
	Notes        string  `json:"notes,omitempty"`
	ProcessedBy  uint    `json:"processed_by"`
}

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
	Successful []uint              `json:"successful"`
	Failed     []BulkUpdateFailure `json:"failed"`
	Summary    map[string]any      `json:"summary"`
}

// BulkUpdateFailure represents a failed bulk operation
type BulkUpdateFailure struct {
	OrderID uint   `json:"order_id"`
	Error   string `json:"error"`
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

// SubscriptionOrderUtilsService provides utility functions for order processing
type SubscriptionOrderUtilsService struct {
	*SubscriptionOrderServiceBase
}

// NewSubscriptionOrderUtilsService creates a new utils service
func NewSubscriptionOrderUtilsService(base *SubscriptionOrderServiceBase) *SubscriptionOrderUtilsService {
	return &SubscriptionOrderUtilsService{
		SubscriptionOrderServiceBase: base,
	}
}

// generateOrderNumber generates a unique order number with enhanced security
func (sous *SubscriptionOrderUtilsService) generateOrderNumber() string {
	// Generate format: ORD + Unix timestamp + 16-character secure random hex
	// This provides better security than predictable nanosecond-based randomness
	now := time.Now()
	timestamp := now.Unix()

	// Generate 8 random bytes (16 hex characters) for strong uniqueness
	randomBytes := make([]byte, 8)
	if _, err := rand.Read(randomBytes); err != nil {
		// Fallback to time-based randomness if crypto/rand fails
		logger.Warn("Failed to generate cryptographically secure random bytes, using fallback", logger.ErrorField(err))
		fallbackRandom := now.UnixNano() % 1000000000 // 9 digits
		return fmt.Sprintf("ORD%d%09d", timestamp, fallbackRandom)
	}

	randomHex := strings.ToUpper(hex.EncodeToString(randomBytes))
	return fmt.Sprintf("ORD%d%s", timestamp, randomHex)
}

// calculateProration calculates the prorated credit for subscription changes
func calculateProration(currentSubscription *entities.UserSubscription, currentPlan, newPlan *entities.SubscriptionPlan) float64 {
	// Skip proration for lifetime subscriptions
	if currentPlan.BillingCycle == constants.BillingCycleLifetime || newPlan.BillingCycle == constants.BillingCycleLifetime {
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

// getApplyImmediately determines when to apply the subscription change
func getApplyImmediately(orderType string, userPreference *bool) bool {
	if userPreference != nil {
		return *userPreference
	}

	// Default behavior: upgrades apply immediately, downgrades apply at period end
	return orderType == constants.OrderTypeUpgrade
}

// checkDuplicateOrders prevents users from creating multiple pending orders for the same plan
func (sous *SubscriptionOrderUtilsService) checkDuplicateOrders(ctx context.Context, userID, planID uint, orderType string) error {
	// SECURITY: Check for pending orders with same plan and type
	var pendingCount int64
	if err := sous.db.WithContext(ctx).Model(&entities.SubscriptionOrder{}).
		Where("user_id = ? AND subscription_plan_id = ? AND order_type = ? AND status IN (?)",
			userID, planID, orderType, []string{constants.OrderStatusPending}).
		Count(&pendingCount).Error; err != nil {
		return fmt.Errorf("failed to check pending orders: %w", err)
	}

	if pendingCount > 0 {
		return fmt.Errorf("you already have a pending order for this subscription plan")
	}

	// SECURITY: Check for recent orders within cooldown period (5 minutes)
	cooldownTime := time.Now().Add(-5 * time.Minute)
	var recentCount int64
	if err := sous.db.WithContext(ctx).Model(&entities.SubscriptionOrder{}).
		Where("user_id = ? AND subscription_plan_id = ? AND order_type = ? AND created_at > ?",
			userID, planID, orderType, cooldownTime).
		Count(&recentCount).Error; err != nil {
		return fmt.Errorf("failed to check recent orders: %w", err)
	}

	if recentCount >= 3 {
		return fmt.Errorf("too many recent order attempts, please wait a few minutes before trying again")
	}

	// BUSINESS LOGIC: For new subscriptions, check plan-specific subscription limits
	if orderType == constants.OrderTypeNew {
		// Get subscription plan to check its limits configuration
		plan, err := sous.subscriptionPlanService.GetSubscriptionPlan(ctx, planID)
		if err != nil {
			return fmt.Errorf("failed to get subscription plan: %w", err)
		}

		// Check if the plan allows multiple active subscriptions
		if !plan.AllowsMultipleActiveSubscriptions() {
			var activeCount int64
			if err := sous.db.WithContext(ctx).Model(&entities.UserSubscription{}).
				Where("user_id = ? AND subscription_plan_id = ? AND status IN (?)",
					userID, planID, []string{constants.UserSubscriptionStatusActive, constants.UserSubscriptionStatusTrial}).
				Count(&activeCount).Error; err != nil {
				return fmt.Errorf("failed to check active subscriptions: %w", err)
			}

			if activeCount > 0 {
				return fmt.Errorf("you already have an active subscription for this plan")
			}
		} else {
			// If plan allows multiple subscriptions, check if user has reached the maximum limit
			maxActiveSubscriptions := plan.GetMaxActiveSubscriptions()
			if maxActiveSubscriptions > 0 { // 0 means unlimited
				var activeCount int64
				if err := sous.db.WithContext(ctx).Model(&entities.UserSubscription{}).
					Where("user_id = ? AND subscription_plan_id = ? AND status IN (?)",
						userID, planID, []string{constants.UserSubscriptionStatusActive, constants.UserSubscriptionStatusTrial}).
					Count(&activeCount).Error; err != nil {
					return fmt.Errorf("failed to check active subscriptions: %w", err)
				}

				if activeCount >= int64(maxActiveSubscriptions) {
					return fmt.Errorf("you have reached the maximum number of active subscriptions (%d) for this plan", maxActiveSubscriptions)
				}
			}
		}
	}

	return nil
}
