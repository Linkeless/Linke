package service

import (
	"context"
	"fmt"
	"time"

	"linke/internal/logger"
	"linke/internal/model"

	"gorm.io/gorm"
)

type SubscriptionService struct {
	db           *gorm.DB
	orderService *OrderService
}

func NewSubscriptionService(db *gorm.DB, orderService *OrderService) *SubscriptionService {
	return &SubscriptionService{
		db:           db,
		orderService: orderService,
	}
}

// CreateActiveSubscriptionRequest represents the request to create a subscription
type CreateActiveSubscriptionRequest struct {
	OrderID         uint     `json:"order_id" binding:"required"`
	UserID          uint     `json:"user_id" binding:"required"`
	PlanID          uint     `json:"plan_id" binding:"required"`
	ServerGroupIDs  []uint   `json:"server_group_ids,omitempty"`
	StartDate       string   `json:"start_date,omitempty"`
	TrialDays       *int     `json:"trial_days,omitempty"`
	AutoRenew       *bool    `json:"auto_renew,omitempty"`
	Notes           string   `json:"notes,omitempty"`
}

// CreateSubscription creates a new active subscription from a fulfilled order
func (ss *SubscriptionService) CreateSubscription(ctx context.Context, req *CreateActiveSubscriptionRequest) (*model.Subscription, error) {
	// Get order with plan details
	order, err := ss.orderService.GetOrderWithRelations(ctx, req.OrderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	// Validate order belongs to user and is fulfilled
	if order.UserID != req.UserID {
		return nil, fmt.Errorf("order does not belong to the specified user")
	}

	if !order.IsFulfilled() {
		return nil, fmt.Errorf("can only create subscriptions for fulfilled orders, current status: %s", order.Status)
	}

	// Check if subscription already exists for this order
	var existingSubscription model.Subscription
	if err := ss.db.WithContext(ctx).Where("order_id = ?", req.OrderID).First(&existingSubscription).Error; err == nil {
		return nil, fmt.Errorf("subscription already exists for this order")
	}

	// Start transaction
	tx := ss.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Calculate subscription periods
	var startDate, endDate, currentPeriodStart, currentPeriodEnd time.Time
	var nextBillingDate *time.Time

	now := time.Now()
	if req.StartDate != "" {
		if parsed, err := time.Parse("2006-01-02", req.StartDate); err == nil {
			startDate = parsed
		} else {
			startDate = now
		}
	} else {
		startDate = now
	}

	// Set trial period if specified
	var trialEndDate *time.Time
	if req.TrialDays != nil && *req.TrialDays > 0 {
		trialEnd := startDate.AddDate(0, 0, *req.TrialDays)
		trialEndDate = &trialEnd
		currentPeriodStart = startDate
		currentPeriodEnd = trialEnd
	} else {
		currentPeriodStart = startDate
		currentPeriodEnd = ss.calculatePeriodEnd(startDate, order.BillingCycle, order.BillingInterval)
	}

	// Calculate service end date
	if order.ServiceEndDate != nil {
		endDate = *order.ServiceEndDate
	} else if order.BillingCycle == model.BillingCycleLifetime {
		// Lifetime subscription - set end date to far future
		endDate = startDate.AddDate(99, 0, 0)
	} else {
		// Calculate based on service period
		endDate = ss.calculateServiceEndDate(startDate, order.BillingCycle, order.ServicePeriod)
	}

	// Set next billing date for recurring subscriptions
	if order.BillingCycle != model.BillingCycleLifetime {
		nextBilling := ss.calculateNextBillingDate(currentPeriodEnd, order.BillingCycle, order.BillingInterval)
		nextBillingDate = &nextBilling
	}

	// Set auto-renew default
	autoRenew := true
	if req.AutoRenew != nil {
		autoRenew = *req.AutoRenew
	}
	// Disable auto-renew for lifetime subscriptions
	if order.BillingCycle == model.BillingCycleLifetime {
		autoRenew = false
	}

	// Create subscription
	subscription := &model.Subscription{
		OrderID:            req.OrderID,
		UserID:             req.UserID,
		PlanID:             req.PlanID,
		Status:             model.SubscriptionStatusActive,
		StartDate:          startDate,
		EndDate:            endDate,
		CurrentPeriodStart: currentPeriodStart,
		CurrentPeriodEnd:   currentPeriodEnd,
		BillingCycle:       order.BillingCycle,
		BillingInterval:    order.BillingInterval,
		Price:              order.TotalAmount,
		Currency:           order.Currency,
		AutoRenew:          autoRenew,
		NextBillingDate:    nextBillingDate,
		TrialEndDate:       trialEndDate,
		Notes:              req.Notes,
	}

	// Set server group permissions
	if len(req.ServerGroupIDs) > 0 {
		if err := subscription.SetServerGroupIDs(req.ServerGroupIDs); err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to set server group IDs: %w", err)
		}
	}

	// Save subscription
	if err := tx.Create(subscription).Error; err != nil {
		tx.Rollback()
		logger.Error("Failed to create subscription", logger.Error2("error", err))
		return nil, fmt.Errorf("failed to create subscription: %w", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit subscription creation: %w", err)
	}

	logger.Info("Subscription created successfully",
		logger.Uint("subscription_id", subscription.ID),
		logger.String("uuid", subscription.UUID),
		logger.Uint("order_id", req.OrderID),
		logger.Uint("user_id", req.UserID))

	return subscription, nil
}

// RenewSubscription renews a subscription for the next billing period
func (ss *SubscriptionService) RenewSubscription(ctx context.Context, subscriptionID uint) error {
	// Start transaction
	tx := ss.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Get subscription with lock
	var subscription model.Subscription
	if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&subscription, subscriptionID).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("subscription not found")
		}
		return fmt.Errorf("failed to get subscription: %w", err)
	}

	// Validate subscription can be renewed
	if !subscription.ShouldRenew() {
		tx.Rollback()
		return fmt.Errorf("subscription should not be renewed: auto_renew=%v, cancel_at_period_end=%v", 
			subscription.AutoRenew, subscription.CancelAtPeriodEnd)
	}

	if !subscription.CanAttemptRenewal() {
		tx.Rollback()
		return fmt.Errorf("renewal attempt too soon after last failure")
	}

	// Calculate new period dates
	newPeriodStart := subscription.CurrentPeriodEnd
	newPeriodEnd := ss.calculatePeriodEnd(newPeriodStart, subscription.BillingCycle, subscription.BillingInterval)
	newNextBillingDate := ss.calculateNextBillingDate(newPeriodEnd, subscription.BillingCycle, subscription.BillingInterval)

	// Update subscription
	now := time.Now()
	updateData := map[string]interface{}{
		"current_period_start": newPeriodStart,
		"current_period_end":   newPeriodEnd,
		"next_billing_date":    newNextBillingDate,
		"renewal_attempts":     0, // Reset renewal attempts on successful renewal
		"last_renewal_failed":  nil,
		"renewal_fail_reason":  "",
		"updated_at":           now,
	}

	if err := tx.Model(&subscription).Updates(updateData).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit subscription renewal: %w", err)
	}

	logger.Info("Subscription renewed successfully",
		logger.Uint("subscription_id", subscriptionID),
		logger.String("uuid", subscription.UUID),
		logger.String("new_period_start", newPeriodStart.Format(time.RFC3339)),
		logger.String("new_period_end", newPeriodEnd.Format(time.RFC3339)))

	return nil
}

// FailRenewal marks a subscription renewal as failed
func (ss *SubscriptionService) FailRenewal(ctx context.Context, subscriptionID uint, reason string) error {
	// Start transaction
	tx := ss.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Get subscription with lock
	var subscription model.Subscription
	if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&subscription, subscriptionID).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("subscription not found")
		}
		return fmt.Errorf("failed to get subscription: %w", err)
	}

	// Update subscription with renewal failure info
	now := time.Now()
	updateData := map[string]interface{}{
		"renewal_attempts":     subscription.RenewalAttempts + 1,
		"last_renewal_failed":  now,
		"renewal_fail_reason":  reason,
		"updated_at":           now,
	}

	// If too many attempts, cancel subscription
	maxAttempts := 3
	if subscription.RenewalAttempts+1 >= maxAttempts {
		updateData["status"] = model.SubscriptionStatusCancelled
		updateData["cancelled_at"] = now
		updateData["cancellation_reason"] = fmt.Sprintf("Renewal failed after %d attempts: %s", maxAttempts, reason)
	}

	if err := tx.Model(&subscription).Updates(updateData).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update subscription renewal failure: %w", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit subscription renewal failure: %w", err)
	}

	logger.Info("Subscription renewal failed",
		logger.Uint("subscription_id", subscriptionID),
		logger.Int("attempt", subscription.RenewalAttempts+1),
		logger.String("reason", reason))

	return nil
}

// CancelSubscription cancels a subscription
func (ss *SubscriptionService) CancelSubscription(ctx context.Context, subscriptionID uint, reason string, immediately bool) error {
	// Start transaction
	tx := ss.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Get subscription with lock
	var subscription model.Subscription
	if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&subscription, subscriptionID).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("subscription not found")
		}
		return fmt.Errorf("failed to get subscription: %w", err)
	}

	// Validate subscription can be cancelled
	if subscription.IsCancelled() {
		tx.Rollback()
		return fmt.Errorf("subscription is already cancelled")
	}

	now := time.Now()
	updateData := map[string]interface{}{
		"auto_renew":           false,
		"cancellation_reason":  reason,
		"updated_at":           now,
	}

	if immediately {
		// Cancel immediately
		updateData["status"] = model.SubscriptionStatusCancelled
		updateData["cancelled_at"] = now
	} else {
		// Cancel at period end
		updateData["cancel_at_period_end"] = true
	}

	if err := tx.Model(&subscription).Updates(updateData).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update subscription cancellation: %w", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit subscription cancellation: %w", err)
	}

	cancelType := "at period end"
	if immediately {
		cancelType = "immediately"
	}

	logger.Info("Subscription cancelled",
		logger.Uint("subscription_id", subscriptionID),
		logger.String("uuid", subscription.UUID),
		logger.String("type", cancelType),
		logger.String("reason", reason))

	return nil
}

// PauseSubscription pauses a subscription
func (ss *SubscriptionService) PauseSubscription(ctx context.Context, subscriptionID uint, reason string) error {
	return ss.updateSubscriptionStatus(ctx, subscriptionID, model.SubscriptionStatusPaused, reason)
}

// ResumeSubscription resumes a paused subscription
func (ss *SubscriptionService) ResumeSubscription(ctx context.Context, subscriptionID uint) error {
	return ss.updateSubscriptionStatus(ctx, subscriptionID, model.SubscriptionStatusActive, "Subscription resumed")
}

// UpdateUsage updates the last used timestamp for a subscription
func (ss *SubscriptionService) UpdateUsage(ctx context.Context, subscriptionID uint) error {
	now := time.Now()
	if err := ss.db.WithContext(ctx).Model(&model.Subscription{}).
		Where("id = ?", subscriptionID).
		Update("last_used_at", now).Error; err != nil {
		return fmt.Errorf("failed to update subscription usage: %w", err)
	}

	return nil
}

// GetSubscription gets a subscription by ID
func (ss *SubscriptionService) GetSubscription(ctx context.Context, subscriptionID uint) (*model.Subscription, error) {
	var subscription model.Subscription
	if err := ss.db.WithContext(ctx).First(&subscription, subscriptionID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("subscription not found")
		}
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}
	return &subscription, nil
}

// GetSubscriptionByUUID gets a subscription by UUID
func (ss *SubscriptionService) GetSubscriptionByUUID(ctx context.Context, uuid string) (*model.Subscription, error) {
	var subscription model.Subscription
	if err := ss.db.WithContext(ctx).Where("uuid = ?", uuid).First(&subscription).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("subscription not found")
		}
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}
	return &subscription, nil
}

// GetSubscriptionWithRelations gets a subscription with related data
func (ss *SubscriptionService) GetSubscriptionWithRelations(ctx context.Context, subscriptionID uint) (*model.Subscription, error) {
	var subscription model.Subscription
	if err := ss.db.WithContext(ctx).
		Preload("Order").
		Preload("User").
		Preload("Plan").
		First(&subscription, subscriptionID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("subscription not found")
		}
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}
	return &subscription, nil
}

// SubscriptionFilters represents filters for subscription listing
type SubscriptionFilters struct {
	UserID      *uint  `form:"user_id"`
	PlanID      *uint  `form:"plan_id"`
	Status      string `form:"status"`
	Currency    string `form:"currency"`
	AutoRenew   *bool  `form:"auto_renew"`
	InTrial     *bool  `form:"in_trial"`
	Expired     *bool  `form:"expired"`
	Overdue     *bool  `form:"overdue"`
	StartDate   string `form:"start_date"`
	EndDate     string `form:"end_date"`
	Search      string `form:"search"`
	SortBy      string `form:"sort_by"`
	SortOrder   string `form:"sort_order"`
	Limit       int    `form:"limit"`
	Offset      int    `form:"offset"`
}

// ListSubscriptions lists subscriptions with filtering and pagination
func (ss *SubscriptionService) ListSubscriptions(ctx context.Context, filters *SubscriptionFilters) ([]*model.Subscription, int64, error) {
	query := ss.db.WithContext(ctx).Model(&model.Subscription{})

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

	if filters.Currency != "" {
		query = query.Where("currency = ?", filters.Currency)
	}

	if filters.AutoRenew != nil {
		query = query.Where("auto_renew = ?", *filters.AutoRenew)
	}

	// Date range filtering
	if filters.StartDate != "" {
		if startDate, err := time.Parse("2006-01-02", filters.StartDate); err == nil {
			query = query.Where("start_date >= ?", startDate)
		}
	}

	if filters.EndDate != "" {
		if endDate, err := time.Parse("2006-01-02", filters.EndDate); err == nil {
			query = query.Where("end_date <= ?", endDate.Add(24*time.Hour))
		}
	}

	// Special filters
	now := time.Now()
	if filters.InTrial != nil {
		if *filters.InTrial {
			query = query.Where("trial_end_date IS NOT NULL AND trial_end_date > ?", now)
		} else {
			query = query.Where("trial_end_date IS NULL OR trial_end_date <= ?", now)
		}
	}

	if filters.Expired != nil {
		if *filters.Expired {
			query = query.Where("end_date < ? OR status = ?", now, model.SubscriptionStatusExpired)
		} else {
			query = query.Where("end_date >= ? AND status != ?", now, model.SubscriptionStatusExpired)
		}
	}

	if filters.Overdue != nil {
		if *filters.Overdue {
			query = query.Where("next_billing_date < ? AND auto_renew = true AND status = ?", 
				now.AddDate(0, 0, -7), model.SubscriptionStatusActive)
		}
	}

	// Search functionality
	if filters.Search != "" {
		searchPattern := "%" + filters.Search + "%"
		query = query.Where("uuid LIKE ? OR notes LIKE ?", searchPattern, searchPattern)
	}

	// Get total count
	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count subscriptions: %w", err)
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

	validSortFields := map[string]bool{
		"created_at":           true,
		"updated_at":           true,
		"start_date":           true,
		"end_date":             true,
		"current_period_end":   true,
		"next_billing_date":    true,
		"price":                true,
		"status":               true,
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

	var subscriptions []*model.Subscription
	if err := query.Find(&subscriptions).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get subscriptions: %w", err)
	}

	return subscriptions, totalCount, nil
}

// Helper methods

func (ss *SubscriptionService) updateSubscriptionStatus(ctx context.Context, subscriptionID uint, status, reason string) error {
	updateData := map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}

	if reason != "" {
		updateData["notes"] = reason
	}

	if err := ss.db.WithContext(ctx).Model(&model.Subscription{}).
		Where("id = ?", subscriptionID).
		Updates(updateData).Error; err != nil {
		return fmt.Errorf("failed to update subscription status: %w", err)
	}

	return nil
}

func (ss *SubscriptionService) calculatePeriodEnd(start time.Time, billingCycle string, billingInterval int) time.Time {
	switch billingCycle {
	case model.BillingCycleMonthly:
		return start.AddDate(0, billingInterval, 0)
	case model.BillingCycleYearly:
		return start.AddDate(billingInterval, 0, 0)
	case model.BillingCycleLifetime:
		return start.AddDate(99, 0, 0) // Far future
	default:
		return start.AddDate(0, billingInterval, 0) // Default to monthly
	}
}

func (ss *SubscriptionService) calculateServiceEndDate(start time.Time, billingCycle string, servicePeriod int) time.Time {
	switch billingCycle {
	case model.BillingCycleMonthly:
		return start.AddDate(0, servicePeriod, 0)
	case model.BillingCycleYearly:
		return start.AddDate(servicePeriod, 0, 0)
	case model.BillingCycleLifetime:
		return start.AddDate(99, 0, 0) // Far future
	default:
		return start.AddDate(0, servicePeriod, 0) // Default to monthly
	}
}

func (ss *SubscriptionService) calculateNextBillingDate(periodEnd time.Time, billingCycle string, billingInterval int) time.Time {
	switch billingCycle {
	case model.BillingCycleMonthly:
		return periodEnd.AddDate(0, billingInterval, 0)
	case model.BillingCycleYearly:
		return periodEnd.AddDate(billingInterval, 0, 0)
	default:
		return periodEnd.AddDate(0, billingInterval, 0) // Default to monthly
	}
}