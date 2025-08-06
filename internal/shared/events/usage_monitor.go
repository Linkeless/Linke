package events

import (
	"context"
	"fmt"

	"linke/internal/shared/logger"
	subscriptionInterfaces "linke/internal/domains/subscription/usecases/interfaces"
	subscriptionEntities "linke/internal/domains/subscription/entities"
)

// UsageMonitor handles real-time usage monitoring and traffic limit checking
type UsageMonitor struct {
	logger                   logger.Logger
	userSubscriptionService  subscriptionInterfaces.UserSubscriptionService
	eventBus                 EventBus
	
	// Configuration
	warningThresholds []float64 // e.g., [80.0, 90.0] for 80% and 90% warnings
}

// NewUsageMonitor creates a new usage monitor
func NewUsageMonitor(
	userSubscriptionService subscriptionInterfaces.UserSubscriptionService,
	eventBus EventBus,
) *UsageMonitor {
	return &UsageMonitor{
		logger:                  logger.GetGlobalLogger(),
		userSubscriptionService: userSubscriptionService,
		eventBus:               eventBus,
		warningThresholds:      []float64{80.0, 90.0}, // Default warning at 80% and 90%
	}
}

// MonitorTrafficUsage checks a subscription's traffic usage and triggers events if thresholds are exceeded
func (m *UsageMonitor) MonitorTrafficUsage(ctx context.Context, subscriptionID uint, newUsageBytes int64) error {
	// Get current subscription state
	subscription, err := m.userSubscriptionService.GetUserSubscription(ctx, subscriptionID)
	if err != nil {
		m.logger.Error("Failed to get subscription for traffic monitoring",
			logger.Uint("subscription_id", subscriptionID),
			logger.ErrorField(err),
		)
		return fmt.Errorf("failed to get subscription: %w", err)
	}

	// Skip monitoring for unlimited traffic subscriptions
	if !subscription.HasTrafficLimit() {
		return nil
	}

	// Calculate previous and new usage percentages
	oldUsagePercentage := subscription.GetTrafficUsagePercentage()
	
	// Update traffic usage
	limitExceeded := subscription.AddTrafficUsage(newUsageBytes)
	newUsagePercentage := subscription.GetTrafficUsagePercentage()

	// Update the subscription in database
	if err := m.userSubscriptionService.UpdateTrafficUsage(ctx, subscriptionID, subscription.TrafficUsed); err != nil {
		m.logger.Error("Failed to update traffic usage",
			logger.Uint("subscription_id", subscriptionID),
			logger.Int64("new_usage", subscription.TrafficUsed),
			logger.ErrorField(err),
		)
		return fmt.Errorf("failed to update traffic usage: %w", err)
	}

	m.logger.Info("Traffic usage updated",
		logger.Uint("subscription_id", subscriptionID),
		logger.Uint("user_id", subscription.UserID),
		logger.Int64("usage_bytes", newUsageBytes),
		logger.Float64("old_percentage", oldUsagePercentage),
		logger.Float64("new_percentage", newUsagePercentage),
		logger.Bool("limit_exceeded", limitExceeded),
	)

	// Check for warning thresholds crossed
	for _, threshold := range m.warningThresholds {
		if oldUsagePercentage < threshold && newUsagePercentage >= threshold && newUsagePercentage < 100 {
			// Crossed a warning threshold
			if err := m.publishTrafficWarningEvent(subscription, threshold, newUsagePercentage); err != nil {
				m.logger.Error("Failed to publish traffic warning event",
					logger.Uint("subscription_id", subscriptionID),
					logger.Float64("threshold", threshold),
					logger.ErrorField(err),
				)
			}
		}
	}

	// Check if traffic limit was exceeded
	if limitExceeded {
		if err := m.publishTrafficLimitExceededEvent(subscription); err != nil {
			m.logger.Error("Failed to publish traffic limit exceeded event",
				logger.Uint("subscription_id", subscriptionID),
				logger.ErrorField(err),
			)
		}
	}

	return nil
}

// MonitorAllActiveSubscriptions checks all active subscriptions for expired ones
func (m *UsageMonitor) MonitorAllActiveSubscriptions(ctx context.Context) error {
	// This would typically be called by a scheduled job
	m.logger.Info("Starting monitoring cycle for all active subscriptions")

	// Check for expired subscriptions
	if err := m.userSubscriptionService.CheckAndProcessExpiredSubscriptions(ctx); err != nil {
		m.logger.Error("Failed to check expired subscriptions",
			logger.ErrorField(err),
		)
		return fmt.Errorf("failed to check expired subscriptions: %w", err)
	}

	// Check for subscriptions that need traffic reset
	// This would require additional service methods to get subscriptions that need reset
	// For now, we'll log that this check would happen
	m.logger.Info("Traffic reset check completed (implementation pending)")

	m.logger.Info("Monitoring cycle completed successfully")
	return nil
}

// publishTrafficWarningEvent publishes a traffic usage warning event
func (m *UsageMonitor) publishTrafficWarningEvent(subscription *subscriptionEntities.UserSubscription, threshold, currentUsage float64) error {
	warningEvent := NewSubscriptionEvent(
		"subscription.traffic_usage_warning",
		subscription.ID,
		subscription.UserID,
		map[string]any{
			"threshold":        threshold,
			"usage_percentage": currentUsage,
			"traffic_used":     subscription.TrafficUsed,
			"traffic_limit":    subscription.TrafficLimit,
			"warning_level":    m.getWarningLevel(threshold),
		},
	)

	return m.eventBus.Publish(context.Background(), warningEvent)
}

// publishTrafficLimitExceededEvent publishes a traffic limit exceeded event
func (m *UsageMonitor) publishTrafficLimitExceededEvent(subscription *subscriptionEntities.UserSubscription) error {
	limitEvent := NewSubscriptionEvent(
		"subscription.traffic_limit_exceeded",
		subscription.ID,
		subscription.UserID,
		map[string]any{
			"traffic_used":     subscription.TrafficUsed,
			"traffic_limit":    subscription.TrafficLimit,
			"usage_percentage": subscription.GetTrafficUsagePercentage(),
			"suspended":        subscription.TrafficSuspended,
		},
	)

	return m.eventBus.Publish(context.Background(), limitEvent)
}

// getWarningLevel returns a descriptive warning level based on threshold
func (m *UsageMonitor) getWarningLevel(threshold float64) string {
	switch {
	case threshold >= 90:
		return "critical"
	case threshold >= 80:
		return "high"
	case threshold >= 70:
		return "medium"
	default:
		return "low"
	}
}

// SetWarningThresholds updates the warning thresholds
func (m *UsageMonitor) SetWarningThresholds(thresholds []float64) {
	m.warningThresholds = thresholds
	m.logger.Info("Warning thresholds updated",
		logger.Any("thresholds", thresholds),
	)
}