package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"linke/internal/logger"
	"linke/internal/model"
)

// ============= Subscription Statistics and Analytics =============

// SubscriptionStatsResponse represents subscription statistics
type SubscriptionStatsResponse struct {
	TotalSubscriptions     int                           `json:"total_subscriptions" example:"5"`
	ActiveSubscriptions    int                           `json:"active_subscriptions" example:"3"`
	PausedSubscriptions    int                           `json:"paused_subscriptions" example:"1"`
	CancelledSubscriptions int                           `json:"cancelled_subscriptions" example:"1"`
	ExpiredSubscriptions   int                           `json:"expired_subscriptions" example:"0"`
	TotalSpent            float64                       `json:"total_spent" example:"149.95"`
	CurrentMonthlyCost    float64                       `json:"current_monthly_cost" example:"29.99"`
	NextBillingDate       *time.Time                    `json:"next_billing_date,omitempty" example:"2024-02-01T00:00:00Z"`
	SubscriptionsByPlan   map[string]int                `json:"subscriptions_by_plan"`
	UsageStats           *SubscriptionUsageStats       `json:"usage_stats,omitempty"`
}

// SubscriptionUsageStats represents usage statistics
type SubscriptionUsageStats struct {
	TotalTrafficUsed    int64   `json:"total_traffic_used" example:"10737418240"`    // bytes
	TotalTrafficLimit   int64   `json:"total_traffic_limit" example:"107374182400"`   // bytes  
	TrafficUsagePercent float64 `json:"traffic_usage_percent" example:"10.0"`
	CurrentMonthTraffic int64   `json:"current_month_traffic" example:"5368709120"`
	AverageSessionTime  int64   `json:"average_session_time" example:"3600"`         // seconds
	TotalSessions      int     `json:"total_sessions" example:"156"`
}

// GetUserSubscriptionStats gets comprehensive statistics for a user's subscriptions
func (s *UserSubscriptionService) GetUserSubscriptionStats(ctx context.Context, userID uint) (*SubscriptionStatsResponse, error) {
	stats := &SubscriptionStatsResponse{
		SubscriptionsByPlan: make(map[string]int),
	}

	// Get subscription counts by status
	var statusCounts []struct {
		Status string
		Count  int
	}
	
	if err := s.db.WithContext(ctx).
		Model(&model.UserSubscription{}).
		Select("status, COUNT(*) as count").
		Where("user_id = ?", userID).
		Group("status").
		Find(&statusCounts).Error; err != nil {
		return nil, fmt.Errorf("failed to get subscription status counts: %w", err)
	}

	// Process status counts
	for _, sc := range statusCounts {
		stats.TotalSubscriptions += sc.Count
		switch sc.Status {
		case model.UserSubscriptionStatusActive:
			stats.ActiveSubscriptions = sc.Count
		case model.UserSubscriptionStatusPaused:
			stats.PausedSubscriptions = sc.Count
		case model.UserSubscriptionStatusCancelled:
			stats.CancelledSubscriptions = sc.Count
		case model.UserSubscriptionStatusExpired:
			stats.ExpiredSubscriptions = sc.Count
		}
	}

	// Get subscriptions by plan
	var planCounts []struct {
		PlanName string `gorm:"column:plan_name"`
		Count    int
	}
	
	if err := s.db.WithContext(ctx).
		Model(&model.UserSubscription{}).
		Select("sp.name as plan_name, COUNT(*) as count").
		Joins("JOIN subscription_plans sp ON user_subscriptions.subscription_plan_id = sp.id").
		Where("user_subscriptions.user_id = ?", userID).
		Group("sp.name").
		Find(&planCounts).Error; err != nil {
		return nil, fmt.Errorf("failed to get subscription plan counts: %w", err)
	}

	for _, pc := range planCounts {
		stats.SubscriptionsByPlan[pc.PlanName] = pc.Count
	}

	// Calculate total spent (sum of all payment records for this user)
	var totalSpent float64
	if err := s.db.WithContext(ctx).
		Model(&model.SubscriptionOrder{}).
		Select("COALESCE(SUM(final_amount), 0)").
		Where("user_id = ? AND status = ?", userID, "completed").
		Scan(&totalSpent).Error; err != nil {
		logger.Error("Failed to calculate total spent", logger.Error2("error", err))
		// Don't fail the whole operation
	}
	stats.TotalSpent = totalSpent

	// Calculate current monthly cost (active subscriptions only)
	var monthlyCost float64
	if err := s.db.WithContext(ctx).
		Model(&model.UserSubscription{}).
		Select("COALESCE(SUM(CASE WHEN billing_cycle = 'monthly' THEN price WHEN billing_cycle = 'yearly' THEN price/12 ELSE 0 END), 0)").
		Where("user_id = ? AND status = ?", userID, model.UserSubscriptionStatusActive).
		Scan(&monthlyCost).Error; err != nil {
		logger.Error("Failed to calculate current monthly cost", logger.Error2("error", err))
	}
	stats.CurrentMonthlyCost = monthlyCost

	// Get next billing date (earliest among active subscriptions)
	var nextBilling time.Time
	if err := s.db.WithContext(ctx).
		Model(&model.UserSubscription{}).
		Select("MIN(next_billing_date)").
		Where("user_id = ? AND status = ? AND next_billing_date IS NOT NULL", userID, model.UserSubscriptionStatusActive).
		Scan(&nextBilling).Error; err == nil && !nextBilling.IsZero() {
		stats.NextBillingDate = &nextBilling
	}

	// Get usage statistics
	usageStats, err := s.getUserUsageStats(ctx, userID)
	if err != nil {
		logger.Error("Failed to get usage stats", logger.Error2("error", err))
		// Don't fail the whole operation
	} else {
		stats.UsageStats = usageStats
	}

	return stats, nil
}

// getUserUsageStats gets usage statistics for a user
func (s *UserSubscriptionService) getUserUsageStats(ctx context.Context, userID uint) (*SubscriptionUsageStats, error) {
	stats := &SubscriptionUsageStats{}

	// Get total traffic used and limits from active subscriptions
	var trafficData struct {
		TotalUsed  int64
		TotalLimit int64
	}
	
	if err := s.db.WithContext(ctx).
		Model(&model.UserSubscription{}).
		Select("COALESCE(SUM(traffic_used), 0) as total_used, COALESCE(SUM(CASE WHEN traffic_limit > 0 THEN traffic_limit ELSE 0 END), 0) as total_limit").
		Where("user_id = ? AND status = ?", userID, model.UserSubscriptionStatusActive).
		Scan(&trafficData).Error; err != nil {
		return nil, fmt.Errorf("failed to get traffic data: %w", err)
	}

	stats.TotalTrafficUsed = trafficData.TotalUsed
	stats.TotalTrafficLimit = trafficData.TotalLimit

	if trafficData.TotalLimit > 0 {
		stats.TrafficUsagePercent = float64(trafficData.TotalUsed) / float64(trafficData.TotalLimit) * 100
	}

	// Traffic logging is not implemented yet, using placeholder values
	stats.CurrentMonthTraffic = 0
	stats.AverageSessionTime = 0
	stats.TotalSessions = 0

	return stats, nil
}

// ============= Traffic and Usage History =============


// ============= Notification Preferences =============

// NotificationPreferences represents user notification preferences
type NotificationPreferences struct {
	EmailNotifications bool `json:"email_notifications"`
	SMSNotifications   bool `json:"sms_notifications"`
	PushNotifications  bool `json:"push_notifications"`
	RenewalReminders   bool `json:"renewal_reminders"`
	ExpirationWarnings bool `json:"expiration_warnings"`
	UsageAlerts        bool `json:"usage_alerts"`
	PromotionalOffers  bool `json:"promotional_offers"`
}

// UpdateNotificationPreferences updates notification preferences for a user
func (s *UserSubscriptionService) UpdateNotificationPreferences(ctx context.Context, userID uint, prefs *NotificationPreferences) error {
	// Convert preferences to JSON
	prefsJSON, err := json.Marshal(prefs)
	if err != nil {
		return fmt.Errorf("failed to marshal preferences: %w", err)
	}

	// Update user provider_data with notification preferences
	if err := s.db.WithContext(ctx).
		Model(&model.User{}).
		Where("id = ?", userID).
		Update("provider_data", fmt.Sprintf(`{"notification_preferences": %s}`, string(prefsJSON))).Error; err != nil {
		return fmt.Errorf("failed to update notification preferences: %w", err)
	}

	logger.Info("Notification preferences updated", logger.Uint("user_id", userID))
	return nil
}

// GetNotificationPreferences gets notification preferences for a user
func (s *UserSubscriptionService) GetNotificationPreferences(ctx context.Context, userID uint) (*NotificationPreferences, error) {
	var user model.User
	if err := s.db.WithContext(ctx).First(&user, userID).Error; err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Default preferences
	prefs := &NotificationPreferences{
		EmailNotifications: true,
		SMSNotifications:   false,
		PushNotifications:  true,
		RenewalReminders:   true,
		ExpirationWarnings: true,
		UsageAlerts:        true,
		PromotionalOffers:  false,
	}

	// Parse existing preferences from provider_data if available
	if user.ProviderData != nil && *user.ProviderData != "" {
		var providerData map[string]interface{}
		if err := json.Unmarshal([]byte(*user.ProviderData), &providerData); err == nil {
			if notifPrefs, exists := providerData["notification_preferences"]; exists {
				if prefsData, err := json.Marshal(notifPrefs); err == nil {
					json.Unmarshal(prefsData, prefs)
				}
			}
		}
	}

	return prefs, nil
}

// ============= Additional Subscription Management Methods =============

// PauseUserSubscription pauses a user subscription
func (s *UserSubscriptionService) PauseUserSubscription(ctx context.Context, subscriptionID uint, reason string, adminUserID uint) (*model.UserSubscription, error) {
	// Start transaction
	tx := s.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Get subscription with lock
	var subscription model.UserSubscription
	if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&subscription, subscriptionID).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}

	// Validate subscription can be paused
	if subscription.Status != model.UserSubscriptionStatusActive {
		tx.Rollback()
		return nil, fmt.Errorf("only active subscriptions can be paused")
	}

	// Update subscription status
	now := time.Now()
	oldStatus := subscription.Status
	updates := map[string]interface{}{
		"status":     model.UserSubscriptionStatusPaused,
		"updated_at": now,
		"notes":      reason,
	}

	if err := tx.Model(&subscription).Updates(updates).Error; err != nil {
		tx.Rollback()
		logger.Error("Failed to pause subscription", 
			logger.Uint("subscription_id", subscriptionID),
			logger.Error2("error", err))
		return nil, fmt.Errorf("failed to pause subscription: %w", err)
	}

	// Log subscription pause action
	logger.Info("Subscription paused successfully", 
		logger.Uint("subscription_id", subscriptionID),
		logger.Uint("user_id", subscription.UserID),
		logger.String("reason", reason),
		logger.String("old_status", string(oldStatus)))

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit subscription pause: %w", err)
	}

	// Reload subscription with updates
	subscription.Status = model.UserSubscriptionStatusPaused

	logger.Info("Subscription paused successfully", 
		logger.Uint("subscription_id", subscriptionID),
		logger.Uint("user_id", subscription.UserID))

	return &subscription, nil
}

// ResumeUserSubscription resumes a paused user subscription
func (s *UserSubscriptionService) ResumeUserSubscription(ctx context.Context, subscriptionID uint, adminUserID uint) (*model.UserSubscription, error) {
	// Start transaction
	tx := s.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Get subscription with lock
	var subscription model.UserSubscription
	if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&subscription, subscriptionID).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}

	// Validate subscription can be resumed
	if subscription.Status != model.UserSubscriptionStatusPaused {
		tx.Rollback()
		return nil, fmt.Errorf("only paused subscriptions can be resumed")
	}

	// Check if subscription has expired while paused
	if subscription.EndDate != nil && subscription.EndDate.Before(time.Now()) {
		tx.Rollback()
		return nil, fmt.Errorf("subscription has expired and cannot be resumed")
	}

	// Update subscription status
	now := time.Now()
	oldStatus := subscription.Status
	updates := map[string]interface{}{
		"status":           model.UserSubscriptionStatusActive,
		"traffic_suspended": false, // Clear traffic suspension on resume
		"updated_at":       now,
	}

	if err := tx.Model(&subscription).Updates(updates).Error; err != nil {
		tx.Rollback()
		logger.Error("Failed to resume subscription", 
			logger.Uint("subscription_id", subscriptionID),
			logger.Error2("error", err))
		return nil, fmt.Errorf("failed to resume subscription: %w", err)
	}

	// Log subscription resume action
	logger.Info("Subscription resumed successfully", 
		logger.Uint("subscription_id", subscriptionID),
		logger.Uint("user_id", subscription.UserID),
		logger.String("old_status", string(oldStatus)))

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit subscription resume: %w", err)
	}

	// Reload subscription with updates
	subscription.Status = model.UserSubscriptionStatusActive
	subscription.TrafficSuspended = false

	logger.Info("Subscription resumed successfully", 
		logger.Uint("subscription_id", subscriptionID),
		logger.Uint("user_id", subscription.UserID))

	return &subscription, nil
}

// ResetTrafficUsage resets traffic usage for a subscription
func (s *UserSubscriptionService) ResetTrafficUsage(ctx context.Context, subscriptionID uint, adminUserID uint) (*model.UserSubscription, error) {
	// Start transaction
	tx := s.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Get subscription with lock
	var subscription model.UserSubscription
	if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&subscription, subscriptionID).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}

	// Reset traffic usage
	now := time.Now()
	oldTrafficUsed := subscription.TrafficUsed

	// Calculate next reset date if applicable
	var nextResetDate *time.Time
	if subscription.TrafficResetCycle == model.TrafficResetCycleMonthly {
		nextReset := now.AddDate(0, 1, 0)
		nextResetDate = &nextReset
	}

	updates := map[string]interface{}{
		"traffic_used":      0,
		"traffic_suspended": false,
		"traffic_reset_date": nextResetDate,
		"updated_at":        now,
	}

	if err := tx.Model(&subscription).Updates(updates).Error; err != nil {
		tx.Rollback()
		logger.Error("Failed to reset traffic usage", 
			logger.Uint("subscription_id", subscriptionID),
			logger.Error2("error", err))
		return nil, fmt.Errorf("failed to reset traffic usage: %w", err)
	}

	// Log traffic reset action
	logger.Info("Traffic usage reset successfully", 
		logger.Uint("subscription_id", subscriptionID),
		logger.Uint("user_id", subscription.UserID),
		logger.Int64("previous_usage", oldTrafficUsed))

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit traffic reset: %w", err)
	}

	// Reload subscription with updates
	subscription.TrafficUsed = 0
	subscription.TrafficSuspended = false
	subscription.TrafficResetDate = nextResetDate

	logger.Info("Traffic usage reset successfully", 
		logger.Uint("subscription_id", subscriptionID),
		logger.Uint("user_id", subscription.UserID),
		logger.Int64("previous_usage", oldTrafficUsed))

	return &subscription, nil
}