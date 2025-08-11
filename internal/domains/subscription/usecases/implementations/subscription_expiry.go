package implementations

import (
	"context"
	"fmt"
	"time"

	"linke/internal/domains/subscription/constants"
	"linke/internal/domains/subscription/entities"
	"linke/internal/shared/logger"

	"gorm.io/gorm"
)

type SubscriptionExpiryService struct {
	db                      *gorm.DB
	userSubscriptionService *UserSubscriptionService
}

func NewSubscriptionExpiryService(db *gorm.DB, userSubscriptionService *UserSubscriptionService) *SubscriptionExpiryService {
	return &SubscriptionExpiryService{
		db:                      db,
		userSubscriptionService: userSubscriptionService,
	}
}

// ProcessExpiredSubscriptions processes all expired subscriptions
func (s *SubscriptionExpiryService) ProcessExpiredSubscriptions(ctx context.Context) (int, error) {
	now := time.Now()

	// Find subscriptions that have expired but are still active
	var expiredSubscriptions []entities.UserSubscription
	if err := s.db.WithContext(ctx).
		Where("status IN (?) AND end_date IS NOT NULL AND end_date <= ?",
			[]string{constants.UserSubscriptionStatusActive, constants.UserSubscriptionStatusTrial}, now).
		Find(&expiredSubscriptions).Error; err != nil {
		logger.Error("Failed to find expired subscriptions", logger.Error2("error", err))
		return 0, fmt.Errorf("failed to find expired subscriptions: %w", err)
	}

	var processedCount int
	for _, subscription := range expiredSubscriptions {
		if err := s.ExpireSubscription(ctx, subscription.ID); err != nil {
			logger.Error("Failed to expire subscription",
				logger.Uint("subscription_id", subscription.ID),
				logger.Error2("error", err))
			continue
		}
		processedCount++
	}

	logger.Info("Expired subscriptions processed", logger.Int("processed_count", processedCount))
	return processedCount, nil
}

// ProcessCancelledSubscriptions processes subscriptions that should be cancelled at period end
func (s *SubscriptionExpiryService) ProcessCancelledSubscriptions(ctx context.Context) (int, error) {
	now := time.Now()

	// Find subscriptions marked for cancellation at period end
	var subscriptions []entities.UserSubscription
	if err := s.db.WithContext(ctx).
		Where("cancel_at_period_end = ? AND status IN (?) AND current_period_end IS NOT NULL AND current_period_end <= ?",
			true, []string{constants.UserSubscriptionStatusActive, constants.UserSubscriptionStatusTrial}, now).
		Find(&subscriptions).Error; err != nil {
		logger.Error("Failed to find subscriptions to cancel", logger.Error2("error", err))
		return 0, fmt.Errorf("failed to find subscriptions to cancel: %w", err)
	}

	var processedCount int
	for _, subscription := range subscriptions {
		if err := s.CancelSubscriptionAtPeriodEnd(ctx, subscription.ID); err != nil {
			logger.Error("Failed to cancel subscription at period end",
				logger.Uint("subscription_id", subscription.ID),
				logger.Error2("error", err))
			continue
		}
		processedCount++
	}

	logger.Info("Period-end cancellations processed", logger.Int("processed_count", processedCount))
	return processedCount, nil
}

// ProcessOverdueSubscriptions processes subscriptions that are overdue for renewal
func (s *SubscriptionExpiryService) ProcessOverdueSubscriptions(ctx context.Context) (int, error) {
	// Find subscriptions that are overdue (7 days past billing date) and still active
	overdueThreshold := time.Now().Add(-7 * 24 * time.Hour)

	var overdueSubscriptions []entities.UserSubscription
	if err := s.db.WithContext(ctx).
		Where("status IN (?) AND next_billing_date IS NOT NULL AND next_billing_date <= ? AND auto_renew = ?",
			[]string{constants.UserSubscriptionStatusActive, constants.UserSubscriptionStatusTrial}, overdueThreshold, false).
		Find(&overdueSubscriptions).Error; err != nil {
		logger.Error("Failed to find overdue subscriptions", logger.Error2("error", err))
		return 0, fmt.Errorf("failed to find overdue subscriptions: %w", err)
	}

	var processedCount int
	for _, subscription := range overdueSubscriptions {
		if err := s.SuspendOverdueSubscription(ctx, subscription.ID); err != nil {
			logger.Error("Failed to suspend overdue subscription",
				logger.Uint("subscription_id", subscription.ID),
				logger.Error2("error", err))
			continue
		}
		processedCount++
	}

	logger.Info("Overdue subscriptions processed", logger.Int("processed_count", processedCount))
	return processedCount, nil
}

// ExpireSubscription marks a subscription as expired
func (s *SubscriptionExpiryService) ExpireSubscription(ctx context.Context, subscriptionID uint) error {
	// Start transaction
	tx := s.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Get subscription with lock
	var subscription entities.UserSubscription
	if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&subscription, subscriptionID).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get subscription: %w", err)
	}

	// Validate subscription is eligible for expiry
	if subscription.Status == constants.UserSubscriptionStatusExpired {
		tx.Rollback()
		return fmt.Errorf("subscription is already expired")
	}

	// Update subscription status
	now := time.Now()
	oldStatus := subscription.Status
	updates := map[string]any{
		"status":     constants.UserSubscriptionStatusExpired,
		"updated_at": now,
	}

	if err := tx.Model(&subscription).Updates(updates).Error; err != nil {
		tx.Rollback()
		logger.Error("Failed to update expired subscription",
			logger.Uint("subscription_id", subscriptionID),
			logger.Error2("error", err))
		return fmt.Errorf("failed to update expired subscription: %w", err)
	}

	// Log subscription expiration
	logger.Info("Subscription expired",
		logger.Uint("subscription_id", subscriptionID),
		logger.Uint("user_id", subscription.UserID),
		logger.String("old_status", string(oldStatus)))

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit subscription expiry: %w", err)
	}

	logger.Info("Subscription expired successfully",
		logger.Uint("subscription_id", subscriptionID),
		logger.Uint("user_id", subscription.UserID))

	return nil
}

// CancelSubscriptionAtPeriodEnd cancels a subscription that was marked for period-end cancellation
func (s *SubscriptionExpiryService) CancelSubscriptionAtPeriodEnd(ctx context.Context, subscriptionID uint) error {
	// Start transaction
	tx := s.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Get subscription with lock
	var subscription entities.UserSubscription
	if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&subscription, subscriptionID).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get subscription: %w", err)
	}

	// Validate subscription is eligible for period-end cancellation
	if !subscription.CancelAtPeriodEnd {
		tx.Rollback()
		return fmt.Errorf("subscription is not marked for period-end cancellation")
	}

	if subscription.Status == constants.UserSubscriptionStatusCancelled {
		tx.Rollback()
		return fmt.Errorf("subscription is already cancelled")
	}

	// Update subscription status
	now := time.Now()
	oldStatus := subscription.Status
	updates := map[string]any{
		"status":               constants.UserSubscriptionStatusCancelled,
		"cancelled_at":         &now,
		"cancel_at_period_end": false,
		"updated_at":           now,
	}

	if subscription.CancellationReason == "" {
		updates["cancellation_reason"] = "Cancelled at period end"
	}

	if err := tx.Model(&subscription).Updates(updates).Error; err != nil {
		tx.Rollback()
		logger.Error("Failed to cancel subscription at period end",
			logger.Uint("subscription_id", subscriptionID),
			logger.Error2("error", err))
		return fmt.Errorf("failed to cancel subscription at period end: %w", err)
	}

	// Log subscription cancellation
	logger.Info("Subscription cancelled at period end",
		logger.Uint("subscription_id", subscriptionID),
		logger.Uint("user_id", subscription.UserID),
		logger.String("old_status", string(oldStatus)))

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit period-end cancellation: %w", err)
	}

	logger.Info("Subscription cancelled at period end",
		logger.Uint("subscription_id", subscriptionID),
		logger.Uint("user_id", subscription.UserID))

	return nil
}

// SuspendOverdueSubscription suspends a subscription that is overdue for payment
func (s *SubscriptionExpiryService) SuspendOverdueSubscription(ctx context.Context, subscriptionID uint) error {
	// Start transaction
	tx := s.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Get subscription with lock
	var subscription entities.UserSubscription
	if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&subscription, subscriptionID).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get subscription: %w", err)
	}

	// Validate subscription is eligible for suspension
	if subscription.Status == constants.UserSubscriptionStatusPaused {
		tx.Rollback()
		return fmt.Errorf("subscription is already suspended")
	}

	// Update subscription status
	now := time.Now()
	oldStatus := subscription.Status
	updates := map[string]any{
		"status":     constants.UserSubscriptionStatusPaused,
		"updated_at": now,
	}

	if err := tx.Model(&subscription).Updates(updates).Error; err != nil {
		tx.Rollback()
		logger.Error("Failed to suspend overdue subscription",
			logger.Uint("subscription_id", subscriptionID),
			logger.Error2("error", err))
		return fmt.Errorf("failed to suspend overdue subscription: %w", err)
	}

	// Log subscription suspension
	logger.Info("Subscription suspended for overdue payment",
		logger.Uint("subscription_id", subscriptionID),
		logger.Uint("user_id", subscription.UserID),
		logger.String("old_status", string(oldStatus)))

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit overdue suspension: %w", err)
	}

	logger.Info("Subscription suspended for overdue payment",
		logger.Uint("subscription_id", subscriptionID),
		logger.Uint("user_id", subscription.UserID))

	return nil
}

// ProcessAllSubscriptionMaintenance runs all subscription maintenance tasks
func (s *SubscriptionExpiryService) ProcessAllSubscriptionMaintenance(ctx context.Context) (map[string]int, error) {
	results := make(map[string]int)
	var errors []error

	// Process expired subscriptions
	if count, err := s.ProcessExpiredSubscriptions(ctx); err != nil {
		errors = append(errors, fmt.Errorf("expired subscriptions: %w", err))
	} else {
		results["expired"] = count
	}

	// Process period-end cancellations
	if count, err := s.ProcessCancelledSubscriptions(ctx); err != nil {
		errors = append(errors, fmt.Errorf("period-end cancellations: %w", err))
	} else {
		results["cancelled"] = count
	}

	// Process overdue subscriptions
	if count, err := s.ProcessOverdueSubscriptions(ctx); err != nil {
		errors = append(errors, fmt.Errorf("overdue subscriptions: %w", err))
	} else {
		results["suspended"] = count
	}

	// Process auto-renewals
	if count, err := s.userSubscriptionService.ProcessAutoRenewals(ctx); err != nil {
		errors = append(errors, fmt.Errorf("auto-renewals: %w", err))
	} else {
		results["auto_renewed"] = count
	}

	if len(errors) > 0 {
		logger.Error("Some subscription maintenance tasks failed", logger.Any("errors", errors))
		return results, fmt.Errorf("maintenance tasks failed: %v", errors)
	}

	logger.Info("All subscription maintenance tasks completed successfully", logger.Any("results", results))
	return results, nil
}
