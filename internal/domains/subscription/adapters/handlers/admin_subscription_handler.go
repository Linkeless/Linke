package handlers

import (
	"time"

	"linke/internal/domains/subscription/usecases/interfaces"

	"github.com/gin-gonic/gin"
)

// Request/Response structures for admin operations

// CreatePlanRequest represents the request body for creating a subscription plan
type CreatePlanRequest struct {
	Name            string  `json:"name" binding:"required,min=1,max=100" example:"Premium Plan"`
	Code            string  `json:"code" binding:"required,min=1,max=50" example:"premium-monthly"`
	Description     string  `json:"description" binding:"max=1000" example:"Premium features with monthly billing"`
	Price           float64 `json:"price" binding:"required,min=0" example:"29.99"`
	Currency        string  `json:"currency" binding:"required,len=3" example:"CNY"`
	BillingCycle    string  `json:"billing_cycle" binding:"required,oneof=monthly yearly lifetime" example:"monthly"`
	BillingInterval int     `json:"billing_interval" binding:"min=1,max=12" example:"1"`
	TrialPeriodDays int     `json:"trial_period_days" binding:"min=0,max=365" example:"7"`
	Features        string  `json:"features,omitempty" example:"{\"max_projects\": 10, \"storage_gb\": 100}"`
	Limits          string  `json:"limits,omitempty" example:"{\"api_calls_per_month\": 10000}"`
	IsVisible       *bool   `json:"is_visible,omitempty" example:"true"`
	SortOrder       int     `json:"sort_order,omitempty" example:"1"`
	IsPopular       *bool   `json:"is_popular,omitempty" example:"false"`
	IsRecommended   *bool   `json:"is_recommended,omitempty" example:"true"`
	SetupFee        float64 `json:"setup_fee,omitempty" example:"0"`
	CancellationFee float64 `json:"cancellation_fee,omitempty" example:"0"`

	// Traffic Configuration (Required)
	TrafficLimit      int64  `json:"traffic_limit" binding:"required,min=0" example:"107374182400"`
	TrafficResetCycle string `json:"traffic_reset_cycle" binding:"required,oneof=monthly never" example:"monthly"`

	// Server Group Configuration (Required)
	DefaultServerGroupIDs []uint `json:"default_server_group_ids" binding:"required,min=1"`
}

// UpdatePlanRequest represents the request body for updating a subscription plan
type UpdatePlanRequest struct {
	Name            *string  `json:"name,omitempty" binding:"omitempty,min=1,max=100" example:"Premium Plan Updated"`
	Description     *string  `json:"description,omitempty" binding:"omitempty,max=1000" example:"Updated description"`
	Price           *float64 `json:"price,omitempty" binding:"omitempty,min=0" example:"39.99"`
	TrialPeriodDays *int     `json:"trial_period_days,omitempty" binding:"omitempty,min=0,max=365" example:"14"`
	Features        *string  `json:"features,omitempty" example:"{\"max_projects\": 20}"`
	Limits          *string  `json:"limits,omitempty" example:"{\"api_calls_per_month\": 20000}"`
	Status          *string  `json:"status,omitempty" binding:"omitempty,oneof=active inactive archived" example:"active"`
	IsVisible       *bool    `json:"is_visible,omitempty" example:"true"`
	SortOrder       *int     `json:"sort_order,omitempty" example:"2"`
	IsPopular       *bool    `json:"is_popular,omitempty" example:"true"`
	IsRecommended   *bool    `json:"is_recommended,omitempty" example:"false"`
	SetupFee        *float64 `json:"setup_fee,omitempty" example:"10"`
	CancellationFee *float64 `json:"cancellation_fee,omitempty" example:"25"`

	// Traffic Configuration
	TrafficLimit      *int64  `json:"traffic_limit,omitempty" binding:"omitempty,min=0" example:"107374182400"`
	TrafficResetCycle *string `json:"traffic_reset_cycle,omitempty" binding:"omitempty,oneof=monthly never" example:"monthly"`

	// Server Group Configuration
	DefaultServerGroupIDs *[]uint `json:"default_server_group_ids,omitempty"`
}

// AdminUpdateUserSubscriptionRequest represents the request body for admin subscription updates
type AdminUpdateUserSubscriptionRequest struct {
	Status             *string    `json:"status,omitempty" binding:"omitempty,oneof=active paused cancelled expired trial" example:"active"`
	EndDate            *time.Time `json:"end_date,omitempty" example:"2024-12-31T23:59:59Z"`
	CancellationReason *string    `json:"cancellation_reason,omitempty" binding:"omitempty,max=255" example:"Admin action"`
	CancelAtPeriodEnd  *bool      `json:"cancel_at_period_end,omitempty" example:"true"`
	AutoRenew          *bool      `json:"auto_renew,omitempty" example:"true"`
	Notes              *string    `json:"notes,omitempty" binding:"omitempty,max=1000" example:"Admin notes"`
	ServerGroupIDs     *[]uint    `json:"server_group_ids,omitempty"`

	// Traffic configuration overrides
	TrafficLimit     *int64 `json:"traffic_limit,omitempty" binding:"omitempty,min=0" example:"107374182400"`
	ResetTraffic     *bool  `json:"reset_traffic,omitempty" example:"false"`
	TrafficSuspended *bool  `json:"traffic_suspended,omitempty" example:"false"`
}

// ExtendSubscriptionRequest represents the request body for extending subscriptions
type ExtendSubscriptionRequest struct {
	ExtendByDays     int    `json:"extend_by_days" binding:"required,min=1,max=3650" example:"30"`
	Reason           string `json:"reason" binding:"required,max=255" example:"Customer loyalty bonus"`
	SendNotification *bool  `json:"send_notification,omitempty" example:"true"`
}

// BulkSubscriptionActionRequest represents bulk operations on subscriptions
type BulkSubscriptionActionRequest struct {
	SubscriptionIDs []uint  `json:"subscription_ids" binding:"required,min=1,max=100"`
	Action          string  `json:"action" binding:"required,oneof=pause resume cancel extend reset_traffic" example:"pause"`
	Reason          *string `json:"reason,omitempty" binding:"omitempty,max=255" example:"Bulk admin action"`
	ExtendByDays    *int    `json:"extend_by_days,omitempty" binding:"omitempty,min=1,max=365"`
}

// RefundOrderRequest represents the request body for refunding orders
type RefundOrderRequest struct {
	RefundAmount       *float64 `json:"refund_amount,omitempty" binding:"omitempty,min=0" example:"29.99"`
	RefundReason       string   `json:"refund_reason" binding:"required,max=255" example:"Customer requested refund"`
	NotifyCustomer     *bool    `json:"notify_customer,omitempty" example:"true"`
	CancelSubscription *bool    `json:"cancel_subscription,omitempty" example:"false"`
}

// AdminUsageResetRequest represents the request to reset usage for a subscription
type AdminUsageResetRequest struct {
	UsageType        *string `json:"usage_type,omitempty" example:"traffic"`
	SendNotification *bool   `json:"send_notification,omitempty" example:"true"`
	Reason           string  `json:"reason" binding:"required,max=255" example:"Admin reset per customer request"`
}

// AdminCreateUserSubscriptionRequest represents the request body for creating a user subscription (Admin only)
type AdminCreateUserSubscriptionRequest struct {
	UserID             uint   `json:"user_id" binding:"required" example:"1"`
	SubscriptionPlanID uint   `json:"subscription_plan_id" binding:"required" example:"1"`
	StartDate          string `json:"start_date,omitempty" example:"2024-01-01T00:00:00Z"`
	UseTrial           *bool  `json:"use_trial,omitempty" example:"false"`
	ServerGroupIDs     []uint `json:"server_group_ids,omitempty"`
	Reason             string `json:"reason" binding:"required,max=255" example:"Admin granted subscription"`

	// Custom Traffic Configuration (optional, overrides plan defaults)
	CustomTrafficLimit      *int64  `json:"custom_traffic_limit,omitempty" example:"107374182400"`  // Custom traffic limit in bytes
	CustomTrafficResetCycle *string `json:"custom_traffic_reset_cycle,omitempty" example:"monthly"` // Custom reset cycle
	DisableTrafficLimit     *bool   `json:"disable_traffic_limit,omitempty" example:"false"`        // Disable traffic limit for this subscription

	// Administrative overrides
	SkipPayment      *bool   `json:"skip_payment,omitempty" example:"true"`        // Skip payment requirement
	SendNotification *bool   `json:"send_notification,omitempty" example:"true"`   // Send notification to user
	Notes            *string `json:"notes,omitempty" binding:"omitempty,max=1000"` // Admin notes
}

// AdminSubscriptionHandler provides backward compatibility by wrapping the unified handler
type AdminSubscriptionHandler struct {
	*AdminSubscriptionUnifiedHandler
}

// NewAdminSubscriptionHandler creates a new admin subscription handler that maintains compatibility
func NewAdminSubscriptionHandler(
	subscriptionPlanService interfaces.SubscriptionPlanService,
	userSubscriptionService interfaces.UserSubscriptionService,
	subscriptionOrderService interfaces.SubscriptionOrderService,
	usageTrackingService interfaces.UsageTrackingService,
	usageAlertService interfaces.UsageAlertService,
) *AdminSubscriptionHandler {
	unified := NewAdminSubscriptionUnifiedHandler(
		subscriptionPlanService,
		userSubscriptionService,
		subscriptionOrderService,
		usageTrackingService,
		usageAlertService,
	)

	return &AdminSubscriptionHandler{
		AdminSubscriptionUnifiedHandler: unified,
	}
}

// Delegated methods for backward compatibility - Plans

func (h *AdminSubscriptionHandler) CreateSubscriptionPlan(c *gin.Context) {
	h.PlansHandler.CreateSubscriptionPlan(c)
}

func (h *AdminSubscriptionHandler) GetSubscriptionPlan(c *gin.Context) {
	h.PlansHandler.GetSubscriptionPlan(c)
}

func (h *AdminSubscriptionHandler) ListSubscriptionPlans(c *gin.Context) {
	h.PlansHandler.ListSubscriptionPlans(c)
}

func (h *AdminSubscriptionHandler) UpdateSubscriptionPlan(c *gin.Context) {
	h.PlansHandler.UpdateSubscriptionPlan(c)
}

func (h *AdminSubscriptionHandler) DeleteSubscriptionPlan(c *gin.Context) {
	h.PlansHandler.DeleteSubscriptionPlan(c)
}

func (h *AdminSubscriptionHandler) ToggleSubscriptionPlanStatus(c *gin.Context) {
	h.PlansHandler.ToggleSubscriptionPlanStatus(c)
}

// Delegated methods for backward compatibility - User Subscriptions

func (h *AdminSubscriptionHandler) CreateUserSubscription(c *gin.Context) {
	h.UsersHandler.CreateUserSubscription(c)
}

func (h *AdminSubscriptionHandler) GetUserSubscription(c *gin.Context) {
	h.UsersHandler.GetUserSubscription(c)
}

func (h *AdminSubscriptionHandler) ListUserSubscriptions(c *gin.Context) {
	h.UsersHandler.ListUserSubscriptions(c)
}

func (h *AdminSubscriptionHandler) UpdateUserSubscription(c *gin.Context) {
	h.UsersHandler.UpdateUserSubscription(c)
}

func (h *AdminSubscriptionHandler) PauseUserSubscription(c *gin.Context) {
	h.UsersHandler.PauseUserSubscription(c)
}

func (h *AdminSubscriptionHandler) ResumeUserSubscription(c *gin.Context) {
	h.UsersHandler.ResumeUserSubscription(c)
}

func (h *AdminSubscriptionHandler) ExtendUserSubscription(c *gin.Context) {
	h.UsersHandler.ExtendUserSubscription(c)
}

func (h *AdminSubscriptionHandler) CancelUserSubscription(c *gin.Context) {
	h.UsersHandler.CancelUserSubscription(c)
}

func (h *AdminSubscriptionHandler) ResetTrafficUsage(c *gin.Context) {
	h.UsersHandler.ResetTrafficUsage(c)
}

func (h *AdminSubscriptionHandler) UpgradeSubscription(c *gin.Context) {
	h.UsersHandler.UpgradeSubscription(c)
}

func (h *AdminSubscriptionHandler) DowngradeSubscription(c *gin.Context) {
	h.UsersHandler.DowngradeSubscription(c)
}

func (h *AdminSubscriptionHandler) GetSubscriptionStatistics(c *gin.Context) {
	h.UsersHandler.GetSubscriptionStatistics(c)
}

// Delegated methods for backward compatibility - Orders

func (h *AdminSubscriptionHandler) GetSubscriptionOrder(c *gin.Context) {
	h.OrdersHandler.GetSubscriptionOrder(c)
}

func (h *AdminSubscriptionHandler) ListSubscriptionOrders(c *gin.Context) {
	h.OrdersHandler.ListSubscriptionOrders(c)
}

func (h *AdminSubscriptionHandler) CancelSubscriptionOrder(c *gin.Context) {
	h.OrdersHandler.CancelSubscriptionOrder(c)
}

func (h *AdminSubscriptionHandler) GetOrderStatistics(c *gin.Context) {
	h.OrdersHandler.GetOrderStatistics(c)
}

// Delegated methods for backward compatibility - Usage

func (h *AdminSubscriptionHandler) GetUsageStatistics(c *gin.Context) {
	h.UsageHandler.GetUsageStatistics(c)
}

func (h *AdminSubscriptionHandler) GetCurrentUsage(c *gin.Context) {
	h.UsageHandler.GetCurrentUsage(c)
}

func (h *AdminSubscriptionHandler) GetUsageAlerts(c *gin.Context) {
	h.UsageHandler.GetUsageAlerts(c)
}

func (h *AdminSubscriptionHandler) GetAlertStatistics(c *gin.Context) {
	h.UsageHandler.GetAlertStatistics(c)
}

func (h *AdminSubscriptionHandler) BulkResolveAlerts(c *gin.Context) {
	h.UsageHandler.BulkResolveAlerts(c)
}

// Delegated methods for backward compatibility - Bulk Operations

func (h *AdminSubscriptionHandler) BulkSubscriptionAction(c *gin.Context) {
	h.BulkHandler.BulkSubscriptionAction(c)
}