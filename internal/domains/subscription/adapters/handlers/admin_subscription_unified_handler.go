package handlers

import (
	"linke/internal/domains/subscription/usecases/interfaces"
)

// AdminSubscriptionUnifiedHandler provides backward compatibility by composing all specialized handlers
type AdminSubscriptionUnifiedHandler struct {
	// Specialized handlers
	PlansHandler  *AdminSubscriptionPlansHandler
	UsersHandler  *AdminSubscriptionUsersHandler
	OrdersHandler *AdminSubscriptionOrdersHandler
	UsageHandler  *AdminSubscriptionUsageHandler
	BulkHandler   *AdminSubscriptionBulkHandler

	// Direct access to base for constructor compatibility
	*AdminSubscriptionHandlerBase
}

// NewAdminSubscriptionUnifiedHandler creates a unified handler that maintains backward compatibility
func NewAdminSubscriptionUnifiedHandler(
	subscriptionPlanService interfaces.SubscriptionPlanService,
	userSubscriptionService interfaces.UserSubscriptionService,
	subscriptionOrderService interfaces.SubscriptionOrderService,
	usageTrackingService interfaces.UsageTrackingService,
	usageAlertService interfaces.UsageAlertService,
) *AdminSubscriptionUnifiedHandler {
	// Create the base handler
	base := NewAdminSubscriptionHandlerBase(
		subscriptionPlanService,
		userSubscriptionService,
		subscriptionOrderService,
		usageTrackingService,
		usageAlertService,
	)

	// Create specialized handlers
	plansHandler := NewAdminSubscriptionPlansHandler(base)
	usersHandler := NewAdminSubscriptionUsersHandler(base)
	ordersHandler := NewAdminSubscriptionOrdersHandler(base)
	usageHandler := NewAdminSubscriptionUsageHandler(base)
	bulkHandler := NewAdminSubscriptionBulkHandler(base)

	return &AdminSubscriptionUnifiedHandler{
		PlansHandler:                 plansHandler,
		UsersHandler:                 usersHandler,
		OrdersHandler:                ordersHandler,
		UsageHandler:                 usageHandler,
		BulkHandler:                  bulkHandler,
		AdminSubscriptionHandlerBase: base,
	}
}

