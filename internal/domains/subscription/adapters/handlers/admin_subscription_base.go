package handlers

import (
	"linke/internal/domains/subscription/usecases/interfaces"
)

// AdminSubscriptionHandlerBase contains shared dependencies and functionality
// for all admin subscription handlers
type AdminSubscriptionHandlerBase struct {
	subscriptionPlanService  interfaces.SubscriptionPlanService
	userSubscriptionService  interfaces.UserSubscriptionService
	subscriptionOrderService interfaces.SubscriptionOrderService
	usageTrackingService     interfaces.UsageTrackingService
	usageAlertService        interfaces.UsageAlertService
}

// NewAdminSubscriptionHandlerBase creates a new base handler with all dependencies
func NewAdminSubscriptionHandlerBase(
	subscriptionPlanService interfaces.SubscriptionPlanService,
	userSubscriptionService interfaces.UserSubscriptionService,
	subscriptionOrderService interfaces.SubscriptionOrderService,
	usageTrackingService interfaces.UsageTrackingService,
	usageAlertService interfaces.UsageAlertService,
) *AdminSubscriptionHandlerBase {
	return &AdminSubscriptionHandlerBase{
		subscriptionPlanService:  subscriptionPlanService,
		userSubscriptionService:  userSubscriptionService,
		subscriptionOrderService: subscriptionOrderService,
		usageTrackingService:     usageTrackingService,
		usageAlertService:        usageAlertService,
	}
}