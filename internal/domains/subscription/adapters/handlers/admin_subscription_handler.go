package handlers

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"linke/internal/domains/subscription/entities"
	"linke/internal/domains/subscription/usecases/interfaces"
	"linke/internal/shared/logger"
	"linke/internal/shared/response"

	"github.com/gin-gonic/gin"
)

// AdminSubscriptionHandler handles admin subscription management operations
type AdminSubscriptionHandler struct {
	subscriptionPlanService  interfaces.SubscriptionPlanService
	userSubscriptionService  interfaces.UserSubscriptionService
	subscriptionOrderService interfaces.SubscriptionOrderService
	usageTrackingService     interfaces.UsageTrackingService
	usageAlertService        interfaces.UsageAlertService
}

// NewAdminSubscriptionHandler creates a new admin subscription handler
func NewAdminSubscriptionHandler(
	subscriptionPlanService interfaces.SubscriptionPlanService,
	userSubscriptionService interfaces.UserSubscriptionService,
	subscriptionOrderService interfaces.SubscriptionOrderService,
	usageTrackingService interfaces.UsageTrackingService,
	usageAlertService interfaces.UsageAlertService,
) *AdminSubscriptionHandler {
	return &AdminSubscriptionHandler{
		subscriptionPlanService:  subscriptionPlanService,
		userSubscriptionService:  userSubscriptionService,
		subscriptionOrderService: subscriptionOrderService,
		usageTrackingService:     usageTrackingService,
		usageAlertService:        usageAlertService,
	}
}

// Request/Response structures for admin operations

// CreatePlanRequest represents the request body for creating a subscription plan
type CreatePlanRequest struct {
	Name            string  `json:"name" binding:"required,min=1,max=100" example:"Premium Plan"`
	Code            string  `json:"code" binding:"required,min=1,max=50" example:"premium-monthly"`
	Description     string  `json:"description" binding:"max=1000" example:"Premium features with monthly billing"`
	Price           float64 `json:"price" binding:"required,min=0" example:"29.99"`
	Currency        string  `json:"currency" binding:"required,len=3" example:"USD"`
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

// SUBSCRIPTION PLANS MANAGEMENT

// CreateSubscriptionPlan godoc
// @Summary Create subscription plan
// @Description Create a new subscription plan (Admin only)
// @Tags Admin-Subscription
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param plan body CreatePlanRequest true "Plan creation data" example({"name":"Premium Plan","code":"premium-monthly","description":"Premium features with monthly billing","price":29.99,"currency":"USD","billing_cycle":"monthly","billing_interval":1,"trial_period_days":7,"features":"{\"max_projects\": 10, \"storage_gb\": 100}","limits":"{\"api_calls_per_month\": 10000}","is_visible":true,"sort_order":1,"is_popular":false,"is_recommended":true,"setup_fee":0,"cancellation_fee":0,"traffic_limit":107374182400,"traffic_reset_cycle":"monthly","default_server_group_ids":[1]})
// @Success 201 {object} response.StandardResponse{data=entities.SubscriptionPlanResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 409 {object} response.ConflictResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/subscriptions/plans [post]
func (h *AdminSubscriptionHandler) CreateSubscriptionPlan(c *gin.Context) {
	var createReq CreatePlanRequest
	if err := c.ShouldBindJSON(&createReq); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Convert to service request
	serviceReq := &interfaces.CreateSubscriptionPlanRequest{
		Name:                  createReq.Name,
		Code:                  createReq.Code,
		Description:           createReq.Description,
		Price:                 createReq.Price,
		Currency:              createReq.Currency,
		BillingCycle:          createReq.BillingCycle,
		BillingInterval:       createReq.BillingInterval,
		TrialPeriodDays:       createReq.TrialPeriodDays,
		Features:              createReq.Features,
		Limits:                createReq.Limits,
		IsVisible:             createReq.IsVisible,
		SortOrder:             createReq.SortOrder,
		IsPopular:             createReq.IsPopular,
		IsRecommended:         createReq.IsRecommended,
		SetupFee:              createReq.SetupFee,
		CancellationFee:       createReq.CancellationFee,
		TrafficLimit:          createReq.TrafficLimit,
		TrafficResetCycle:     createReq.TrafficResetCycle,
		DefaultServerGroupIDs: createReq.DefaultServerGroupIDs,
	}

	// Create the plan (use admin user ID from context if available)
	// For now, use 0 as creator ID - this should be enhanced to get actual admin user ID
	plan, err := h.subscriptionPlanService.CreateSubscriptionPlan(c.Request.Context(), 0, serviceReq)
	if err != nil {
		logger.Error("Admin failed to create subscription plan",
			logger.String("code", createReq.Code),
			logger.ErrorField(err),
		)

		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			response.Conflict(c, "Plan with this code already exists")
			return
		}

		response.InternalServerError(c, "Failed to create subscription plan")
		return
	}

	logger.Info("Admin created subscription plan",
		logger.Uint("plan_id", plan.ID),
		logger.String("code", plan.Code),
		logger.String("admin_action", "create_plan"),
	)

	response.Created(c, plan.ToResponse())
}

// GetSubscriptionPlan godoc
// @Summary Get subscription plan
// @Description Get subscription plan details by ID (Admin only)
// @Tags Admin-Subscription
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Plan ID"
// @Success 200 {object} response.StandardResponse{data=entities.SubscriptionPlanResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Router /admin/subscriptions/plans/{id} [get]
func (h *AdminSubscriptionHandler) GetSubscriptionPlan(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid plan ID")
		return
	}

	plan, err := h.subscriptionPlanService.GetSubscriptionPlan(c.Request.Context(), uint(id))
	if err != nil {
		logger.Error("Admin failed to get subscription plan",
			logger.Uint("plan_id", uint(id)),
			logger.ErrorField(err),
		)
		response.NotFound(c, "Subscription plan not found")
		return
	}

	response.Success(c, plan.ToResponse())
}

// ListSubscriptionPlans godoc
// @Summary List subscription plans
// @Description Get paginated list of all subscription plans (Admin only)
// @Tags Admin-Subscription
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param status query string false "Filter by status" Enums(active,inactive,archived)
// @Param currency query string false "Filter by currency" example("USD")
// @Param visible query bool false "Filter by visibility"
// @Param popular query bool false "Filter by popular flag"
// @Param recommended query bool false "Filter by recommended flag"
// @Param limit query int false "Items per page" default(10)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} response.StandardListResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/subscriptions/plans [get]
func (h *AdminSubscriptionHandler) ListSubscriptionPlans(c *gin.Context) {
	// Parse query parameters
	req := &interfaces.GetSubscriptionPlansRequest{}

	if err := c.ShouldBindQuery(req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Set defaults
	if req.Limit == 0 {
		req.Limit = 10
	}
	if req.Limit > 100 {
		req.Limit = 100
	}

	plans, total, err := h.subscriptionPlanService.GetSubscriptionPlans(c.Request.Context(), req)
	if err != nil {
		logger.Error("Admin failed to list subscription plans", logger.ErrorField(err))
		response.InternalServerError(c, "Failed to list subscription plans")
		return
	}

	// Convert to response format
	planResponses := make([]*entities.SubscriptionPlanResponse, len(plans))
	for i, plan := range plans {
		planResponses[i] = plan.ToResponse()
	}

	page := (req.Offset / req.Limit) + 1
	response.SuccessList(c, planResponses, page, req.Limit, total)
}

// UpdateSubscriptionPlan godoc
// @Summary Update subscription plan
// @Description Update subscription plan details (Admin only)
// @Tags Admin-Subscription
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Plan ID"
// @Param plan body UpdatePlanRequest true "Plan update data"
// @Success 200 {object} response.StandardResponse{data=entities.SubscriptionPlanResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/subscriptions/plans/{id} [put]
func (h *AdminSubscriptionHandler) UpdateSubscriptionPlan(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid plan ID")
		return
	}

	var updateReq UpdatePlanRequest
	if err := c.ShouldBindJSON(&updateReq); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Convert to service request
	serviceReq := &interfaces.UpdateSubscriptionPlanRequest{
		Name:                  updateReq.Name,
		Description:           updateReq.Description,
		Price:                 updateReq.Price,
		TrialPeriodDays:       updateReq.TrialPeriodDays,
		Features:              updateReq.Features,
		Limits:                updateReq.Limits,
		Status:                updateReq.Status,
		IsVisible:             updateReq.IsVisible,
		SortOrder:             updateReq.SortOrder,
		IsPopular:             updateReq.IsPopular,
		IsRecommended:         updateReq.IsRecommended,
		SetupFee:              updateReq.SetupFee,
		CancellationFee:       updateReq.CancellationFee,
		TrafficLimit:          updateReq.TrafficLimit,
		TrafficResetCycle:     updateReq.TrafficResetCycle,
		DefaultServerGroupIDs: updateReq.DefaultServerGroupIDs,
	}

	plan, err := h.subscriptionPlanService.UpdateSubscriptionPlan(c.Request.Context(), uint(id), serviceReq)
	if err != nil {
		logger.Error("Admin failed to update subscription plan",
			logger.Uint("plan_id", uint(id)),
			logger.ErrorField(err),
		)
		response.InternalServerError(c, "Failed to update subscription plan")
		return
	}

	logger.Info("Admin updated subscription plan",
		logger.Uint("plan_id", uint(id)),
		logger.String("admin_action", "update_plan"),
	)

	response.Success(c, plan.ToResponse())
}

// DeleteSubscriptionPlan godoc
// @Summary Delete subscription plan
// @Description Soft delete a subscription plan (Admin only)
// @Tags Admin-Subscription
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Plan ID"
// @Success 200 {object} response.MessageOnlyResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Router /admin/subscriptions/plans/{id} [delete]
func (h *AdminSubscriptionHandler) DeleteSubscriptionPlan(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid plan ID")
		return
	}

	if err := h.subscriptionPlanService.DeleteSubscriptionPlan(c.Request.Context(), uint(id)); err != nil {
		logger.Error("Admin failed to delete subscription plan",
			logger.Uint("plan_id", uint(id)),
			logger.ErrorField(err),
		)
		response.NotFound(c, "Subscription plan not found")
		return
	}

	logger.Info("Admin deleted subscription plan",
		logger.Uint("plan_id", uint(id)),
		logger.String("admin_action", "delete_plan"),
	)

	response.SuccessWithMessage(c, "Subscription plan deleted successfully", nil)
}

// ToggleSubscriptionPlanStatus godoc
// @Summary Toggle subscription plan status
// @Description Toggle subscription plan active/inactive status (Admin only)
// @Tags Admin-Subscription
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Plan ID"
// @Success 200 {object} response.StandardResponse{data=entities.SubscriptionPlanResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Router /admin/subscriptions/plans/{id}/toggle-status [put]
func (h *AdminSubscriptionHandler) ToggleSubscriptionPlanStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid plan ID")
		return
	}

	plan, err := h.subscriptionPlanService.ToggleSubscriptionPlanStatus(c.Request.Context(), uint(id))
	if err != nil {
		logger.Error("Admin failed to toggle subscription plan status",
			logger.Uint("plan_id", uint(id)),
			logger.ErrorField(err),
		)
		response.NotFound(c, "Subscription plan not found")
		return
	}

	logger.Info("Admin toggled subscription plan status",
		logger.Uint("plan_id", uint(id)),
		logger.String("new_status", plan.Status),
		logger.String("admin_action", "toggle_plan_status"),
	)

	response.Success(c, plan.ToResponse())
}

// USER SUBSCRIPTIONS MANAGEMENT

// CreateUserSubscription godoc
// @Summary Create user subscription
// @Description Create a subscription for a user directly (Admin only)
// @Tags Admin-Subscription
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param subscription body AdminCreateUserSubscriptionRequest true "Subscription creation data"
// @Success 201 {object} response.StandardResponse{data=entities.UserSubscriptionResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 409 {object} response.ConflictResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/subscriptions/users [post]
func (h *AdminSubscriptionHandler) CreateUserSubscription(c *gin.Context) {
	var createReq AdminCreateUserSubscriptionRequest
	if err := c.ShouldBindJSON(&createReq); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Convert to service request
	serviceReq := &interfaces.CreateSubscriptionRequest{
		UserID:                  createReq.UserID,
		SubscriptionPlanID:      createReq.SubscriptionPlanID,
		StartDate:               createReq.StartDate,
		UseTrial:                createReq.UseTrial != nil && *createReq.UseTrial,
		ServerGroupIDs:          createReq.ServerGroupIDs,
		CustomTrafficLimit:      createReq.CustomTrafficLimit,
		CustomTrafficResetCycle: createReq.CustomTrafficResetCycle,
		DisableTrafficLimit:     createReq.DisableTrafficLimit,
	}

	// Create the user subscription
	subscription, err := h.userSubscriptionService.CreateUserSubscription(c.Request.Context(), serviceReq)
	if err != nil {
		logger.Error("Admin failed to create user subscription",
			logger.Uint("user_id", createReq.UserID),
			logger.Uint("plan_id", createReq.SubscriptionPlanID),
			logger.String("reason", createReq.Reason),
			logger.ErrorField(err),
		)

		if strings.Contains(err.Error(), "not found") {
			response.NotFound(c, "User or subscription plan not found")
		} else if strings.Contains(err.Error(), "already has") || strings.Contains(err.Error(), "duplicate") {
			response.Conflict(c, "User already has an active subscription for this plan")
		} else {
			response.InternalServerError(c, "Failed to create user subscription")
		}
		return
	}

	logger.Info("Admin created user subscription",
		logger.Uint("subscription_id", subscription.ID),
		logger.Uint("user_id", createReq.UserID),
		logger.Uint("plan_id", createReq.SubscriptionPlanID),
		logger.String("reason", createReq.Reason),
		logger.String("admin_action", "create_user_subscription"),
	)

	response.Created(c, subscription.ToResponse())
}

// GetUserSubscription godoc
// @Summary Get user subscription
// @Description Get user subscription details by ID (Admin only)
// @Tags Admin-Subscription
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Subscription ID"
// @Success 200 {object} response.StandardResponse{data=entities.UserSubscriptionResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Router /admin/subscriptions/users/{id} [get]
func (h *AdminSubscriptionHandler) GetUserSubscription(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid subscription ID")
		return
	}

	subscription, err := h.userSubscriptionService.GetUserSubscription(c.Request.Context(), uint(id))
	if err != nil {
		logger.Error("Admin failed to get user subscription",
			logger.Uint("subscription_id", uint(id)),
			logger.ErrorField(err),
		)
		response.NotFound(c, "User subscription not found")
		return
	}

	response.Success(c, subscription.ToResponse())
}

// ListUserSubscriptions godoc
// @Summary List all user subscriptions
// @Description Get paginated list of all user subscriptions (Admin only)
// @Tags Admin-Subscription
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param user_id query int false "Filter by user ID"
// @Param status query string false "Filter by status" Enums(active,paused,cancelled,expired,trial)
// @Param limit query int false "Items per page" default(10)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} response.StandardListResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/subscriptions/users [get]
func (h *AdminSubscriptionHandler) ListUserSubscriptions(c *gin.Context) {
	// Parse query parameters
	req := &interfaces.GetUserSubscriptionsRequest{}

	if err := c.ShouldBindQuery(req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Set defaults
	if req.Limit == 0 {
		req.Limit = 10
	}
	if req.Limit > 100 {
		req.Limit = 100
	}

	subscriptionResponses, total, err := h.userSubscriptionService.GetUserSubscriptionsWithUserDataForAdmin(c.Request.Context(), req)
	if err != nil {
		logger.Error("Admin failed to list user subscriptions", logger.ErrorField(err))
		response.InternalServerError(c, "Failed to list user subscriptions")
		return
	}

	page := (req.Offset / req.Limit) + 1
	response.SuccessList(c, subscriptionResponses, page, req.Limit, total)
}

// UpdateUserSubscription godoc
// @Summary Update user subscription
// @Description Update user subscription details (Admin only)
// @Tags Admin-Subscription
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Subscription ID"
// @Param subscription body AdminUpdateUserSubscriptionRequest true "Subscription update data"
// @Success 200 {object} response.StandardResponse{data=entities.UserSubscriptionResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/subscriptions/users/{id} [put]
func (h *AdminSubscriptionHandler) UpdateUserSubscription(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid subscription ID")
		return
	}

	var updateReq AdminUpdateUserSubscriptionRequest
	if err := c.ShouldBindJSON(&updateReq); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Handle traffic reset if requested
	if updateReq.ResetTraffic != nil && *updateReq.ResetTraffic {
		// Reset traffic usage for this subscription
		if _, err := h.userSubscriptionService.ResetTrafficUsage(c.Request.Context(), uint(id), 0); err != nil {
			logger.Error("Admin failed to reset traffic usage",
				logger.Uint("subscription_id", uint(id)),
				logger.ErrorField(err),
			)
		}
	}

	// Convert to service request
	serviceReq := &interfaces.UpdateSubscriptionRequest{
		Status:             updateReq.Status,
		EndDate:            updateReq.EndDate,
		CancellationReason: updateReq.CancellationReason,
		CancelAtPeriodEnd:  updateReq.CancelAtPeriodEnd,
		AutoRenew:          updateReq.AutoRenew,
		Notes:              updateReq.Notes,
		ServerGroupIDs:     updateReq.ServerGroupIDs,
	}

	subscription, err := h.userSubscriptionService.UpdateUserSubscription(c.Request.Context(), uint(id), serviceReq)
	if err != nil {
		logger.Error("Admin failed to update user subscription",
			logger.Uint("subscription_id", uint(id)),
			logger.ErrorField(err),
		)
		response.InternalServerError(c, "Failed to update user subscription")
		return
	}

	logger.Info("Admin updated user subscription",
		logger.Uint("subscription_id", uint(id)),
		logger.String("admin_action", "update_subscription"),
	)

	response.Success(c, subscription.ToResponse())
}

// PauseUserSubscription godoc
// @Summary Pause user subscription
// @Description Pause a user subscription (Admin only)
// @Tags Admin-Subscription
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Subscription ID"
// @Param pause body interfaces.PauseSubscriptionRequest true "Pause request data"
// @Success 200 {object} response.StandardResponse{data=entities.UserSubscriptionResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/subscriptions/users/{id}/pause [post]
func (h *AdminSubscriptionHandler) PauseUserSubscription(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid subscription ID")
		return
	}

	var pauseReq interfaces.PauseSubscriptionRequest
	if err := c.ShouldBindJSON(&pauseReq); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// For now, use admin ID 0 - should be enhanced to get actual admin user ID from context
	subscription, err := h.userSubscriptionService.PauseUserSubscription(c.Request.Context(), uint(id), &pauseReq, 0)
	if err != nil {
		logger.Error("Admin failed to pause user subscription",
			logger.Uint("subscription_id", uint(id)),
			logger.ErrorField(err),
		)

		if strings.Contains(err.Error(), "not found") {
			response.NotFound(c, "User subscription not found")
		} else if strings.Contains(err.Error(), "cannot be paused") {
			response.BadRequest(c, err.Error())
		} else {
			response.InternalServerError(c, "Failed to pause subscription")
		}
		return
	}

	logger.Info("Admin paused user subscription",
		logger.Uint("subscription_id", uint(id)),
		logger.String("reason", pauseReq.Reason),
		logger.String("admin_action", "pause_subscription"),
	)

	response.Success(c, subscription.ToResponse())
}

// ResumeUserSubscription godoc
// @Summary Resume user subscription
// @Description Resume a paused user subscription (Admin only)
// @Tags Admin-Subscription
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Subscription ID"
// @Param resume body interfaces.ResumeSubscriptionRequest true "Resume request data"
// @Success 200 {object} response.StandardResponse{data=entities.UserSubscriptionResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/subscriptions/users/{id}/resume [post]
func (h *AdminSubscriptionHandler) ResumeUserSubscription(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid subscription ID")
		return
	}

	var resumeReq interfaces.ResumeSubscriptionRequest
	if err := c.ShouldBindJSON(&resumeReq); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// For now, use admin ID 0 - should be enhanced to get actual admin user ID from context
	subscription, err := h.userSubscriptionService.ResumeUserSubscription(c.Request.Context(), uint(id), &resumeReq, 0)
	if err != nil {
		logger.Error("Admin failed to resume user subscription",
			logger.Uint("subscription_id", uint(id)),
			logger.ErrorField(err),
		)

		if strings.Contains(err.Error(), "not found") {
			response.NotFound(c, "User subscription not found")
		} else if strings.Contains(err.Error(), "cannot be resumed") {
			response.BadRequest(c, err.Error())
		} else {
			response.InternalServerError(c, "Failed to resume subscription")
		}
		return
	}

	logger.Info("Admin resumed user subscription",
		logger.Uint("subscription_id", uint(id)),
		logger.String("admin_action", "resume_subscription"),
	)

	response.Success(c, subscription.ToResponse())
}

// ExtendUserSubscription godoc
// @Summary Extend user subscription
// @Description Extend a user subscription by specified days (Admin only)
// @Tags Admin-Subscription
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Subscription ID"
// @Param extend body ExtendSubscriptionRequest true "Extension data"
// @Success 200 {object} response.StandardResponse{data=entities.UserSubscriptionResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/subscriptions/users/{id}/extend [post]
func (h *AdminSubscriptionHandler) ExtendUserSubscription(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid subscription ID")
		return
	}

	var extendReq ExtendSubscriptionRequest
	if err := c.ShouldBindJSON(&extendReq); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.userSubscriptionService.ExtendSubscription(c.Request.Context(), uint(id), extendReq.ExtendByDays, extendReq.Reason); err != nil {
		logger.Error("Admin failed to extend user subscription",
			logger.Uint("subscription_id", uint(id)),
			logger.Int("extend_by_days", extendReq.ExtendByDays),
			logger.ErrorField(err),
		)

		if strings.Contains(err.Error(), "not found") {
			response.NotFound(c, "User subscription not found")
		} else {
			response.InternalServerError(c, "Failed to extend subscription")
		}
		return
	}

	// Get updated subscription
	subscription, err := h.userSubscriptionService.GetUserSubscription(c.Request.Context(), uint(id))
	if err != nil {
		logger.Error("Failed to retrieve extended subscription",
			logger.Uint("subscription_id", uint(id)),
			logger.ErrorField(err),
		)
		response.InternalServerError(c, "Subscription extended but failed to retrieve updated data")
		return
	}

	logger.Info("Admin extended user subscription",
		logger.Uint("subscription_id", uint(id)),
		logger.Int("extend_by_days", extendReq.ExtendByDays),
		logger.String("reason", extendReq.Reason),
		logger.String("admin_action", "extend_subscription"),
	)

	response.Success(c, subscription.ToResponse())
}

// CancelUserSubscription godoc
// @Summary Cancel user subscription
// @Description Cancel a user subscription (Admin only)
// @Tags Admin-Subscription
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Subscription ID"
// @Param cancel body object{reason=string,cancel_at_period_end=bool} true "Cancel data"
// @Success 200 {object} response.MessageOnlyResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/subscriptions/users/{id}/cancel [post]
func (h *AdminSubscriptionHandler) CancelUserSubscription(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid subscription ID")
		return
	}

	var cancelData struct {
		Reason            string `json:"reason" binding:"required,max=255"`
		CancelAtPeriodEnd bool   `json:"cancel_at_period_end"`
	}
	if err := c.ShouldBindJSON(&cancelData); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.userSubscriptionService.CancelUserSubscription(c.Request.Context(), uint(id), cancelData.Reason, cancelData.CancelAtPeriodEnd); err != nil {
		logger.Error("Admin failed to cancel user subscription",
			logger.Uint("subscription_id", uint(id)),
			logger.ErrorField(err),
		)

		if strings.Contains(err.Error(), "not found") {
			response.NotFound(c, "User subscription not found")
		} else {
			response.InternalServerError(c, "Failed to cancel subscription")
		}
		return
	}

	logger.Info("Admin cancelled user subscription",
		logger.Uint("subscription_id", uint(id)),
		logger.String("reason", cancelData.Reason),
		logger.Bool("cancel_at_period_end", cancelData.CancelAtPeriodEnd),
		logger.String("admin_action", "cancel_subscription"),
	)

	response.SuccessWithMessage(c, "Subscription cancelled successfully", nil)
}

// ResetTrafficUsage godoc
// @Summary Reset traffic usage
// @Description Reset traffic usage for a user subscription (Admin only)
// @Tags Admin-Subscription
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Subscription ID"
// @Param reset body AdminUsageResetRequest true "Reset data"
// @Success 200 {object} response.StandardResponse{data=entities.UserSubscriptionResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/subscriptions/users/{id}/reset-traffic [post]
func (h *AdminSubscriptionHandler) ResetTrafficUsage(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid subscription ID")
		return
	}

	var resetReq AdminUsageResetRequest
	if err := c.ShouldBindJSON(&resetReq); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// For now, use admin ID 0 - should be enhanced to get actual admin user ID from context
	subscription, err := h.userSubscriptionService.ResetTrafficUsage(c.Request.Context(), uint(id), 0)
	if err != nil {
		logger.Error("Admin failed to reset traffic usage",
			logger.Uint("subscription_id", uint(id)),
			logger.ErrorField(err),
		)

		if strings.Contains(err.Error(), "not found") {
			response.NotFound(c, "User subscription not found")
		} else {
			response.InternalServerError(c, "Failed to reset traffic usage")
		}
		return
	}

	logger.Info("Admin reset traffic usage",
		logger.Uint("subscription_id", uint(id)),
		logger.String("reason", resetReq.Reason),
		logger.String("admin_action", "reset_traffic"),
	)

	response.Success(c, subscription.ToResponse())
}

// UpgradeSubscription godoc
// @Summary Upgrade user subscription
// @Description Upgrade a user's subscription to a higher plan (Admin only)
// @Tags Admin-Subscription-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Subscription ID"
// @Param upgrade body interfaces.UpgradeSubscriptionRequest true "Upgrade request"
// @Success 200 {object} response.StandardResponse{data=entities.UserSubscriptionResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/subscriptions/users/{id}/upgrade [post]
func (h *AdminSubscriptionHandler) UpgradeSubscription(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid subscription ID")
		return
	}

	var req interfaces.UpgradeSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Set the subscription ID from the URL parameter
	req.SubscriptionID = uint(id)

	subscription, err := h.userSubscriptionService.UpgradeUserSubscription(c.Request.Context(), &req)
	if err != nil {
		logger.Error("Admin failed to upgrade subscription",
			logger.Uint("subscription_id", uint(id)),
			logger.Uint("new_plan_id", req.NewSubscriptionPlanID),
			logger.ErrorField(err),
		)

		if strings.Contains(err.Error(), "not found") {
			response.NotFound(c, "Subscription not found")
		} else if strings.Contains(err.Error(), "not active") {
			response.BadRequest(c, "Subscription is not active")
		} else if strings.Contains(err.Error(), "not available") {
			response.BadRequest(c, "New plan is not available")
		} else if strings.Contains(err.Error(), "price") {
			response.BadRequest(c, err.Error())
		} else {
			response.InternalServerError(c, "Failed to upgrade subscription")
		}
		return
	}

	logger.Info("Admin upgraded subscription",
		logger.Uint("subscription_id", uint(id)),
		logger.Uint("new_plan_id", req.NewSubscriptionPlanID),
		logger.String("admin_action", "upgrade_subscription"),
	)

	response.Success(c, subscription.ToResponse())
}

// DowngradeSubscription godoc
// @Summary Downgrade user subscription
// @Description Downgrade a user's subscription to a lower plan (Admin only)
// @Tags Admin-Subscription-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Subscription ID"
// @Param downgrade body interfaces.DowngradeSubscriptionRequest true "Downgrade request"
// @Success 200 {object} response.StandardResponse{data=entities.UserSubscriptionResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/subscriptions/users/{id}/downgrade [post]
func (h *AdminSubscriptionHandler) DowngradeSubscription(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid subscription ID")
		return
	}

	var req interfaces.DowngradeSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Set the subscription ID from the URL parameter
	req.SubscriptionID = uint(id)

	subscription, err := h.userSubscriptionService.DowngradeUserSubscription(c.Request.Context(), &req)
	if err != nil {
		logger.Error("Admin failed to downgrade subscription",
			logger.Uint("subscription_id", uint(id)),
			logger.Uint("new_plan_id", req.NewSubscriptionPlanID),
			logger.ErrorField(err),
		)

		if strings.Contains(err.Error(), "not found") {
			response.NotFound(c, "Subscription not found")
		} else if strings.Contains(err.Error(), "not active") {
			response.BadRequest(c, "Subscription is not active")
		} else if strings.Contains(err.Error(), "not available") {
			response.BadRequest(c, "New plan is not available")
		} else if strings.Contains(err.Error(), "price") {
			response.BadRequest(c, err.Error())
		} else {
			response.InternalServerError(c, "Failed to downgrade subscription")
		}
		return
	}

	logger.Info("Admin downgraded subscription",
		logger.Uint("subscription_id", uint(id)),
		logger.Uint("new_plan_id", req.NewSubscriptionPlanID),
		logger.String("admin_action", "downgrade_subscription"),
	)

	response.Success(c, subscription.ToResponse())
}

// GetSubscriptionStatistics godoc
// @Summary Get subscription statistics
// @Description Get overall subscription statistics (Admin only)
// @Tags Admin-Subscription-Analytics
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.StandardResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/subscriptions/analytics/statistics [get]
func (h *AdminSubscriptionHandler) GetSubscriptionStatistics(c *gin.Context) {
	stats, err := h.userSubscriptionService.GetSubscriptionStatistics(c.Request.Context())
	if err != nil {
		logger.Error("Admin failed to get subscription statistics", logger.ErrorField(err))
		response.InternalServerError(c, "Failed to get subscription statistics")
		return
	}

	response.Success(c, stats)
}

// SUBSCRIPTION ORDERS MANAGEMENT

// GetSubscriptionOrder godoc
// @Summary Get subscription order
// @Description Get subscription order details by ID (Admin only)
// @Tags Admin-Subscription-Orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Order ID"
// @Success 200 {object} response.StandardResponse{data=entities.SubscriptionOrderResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Router /admin/subscriptions/orders/{id} [get]
func (h *AdminSubscriptionHandler) GetSubscriptionOrder(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid order ID")
		return
	}

	order, err := h.subscriptionOrderService.GetSubscriptionOrder(c.Request.Context(), uint(id))
	if err != nil {
		logger.Error("Admin failed to get subscription order",
			logger.Uint("order_id", uint(id)),
			logger.ErrorField(err),
		)
		response.NotFound(c, "Subscription order not found")
		return
	}

	response.Success(c, order.ToResponse())
}

// ListSubscriptionOrders godoc
// @Summary List subscription orders
// @Description Get paginated list of all subscription orders (Admin only)
// @Tags Admin-Subscription-Orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param user_id query int false "Filter by user ID"
// @Param status query string false "Filter by status"
// @Param order_type query string false "Filter by order type"
// @Param date_from query string false "Filter from date (YYYY-MM-DD)"
// @Param date_to query string false "Filter to date (YYYY-MM-DD)"
// @Param limit query int false "Items per page" default(10)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} response.StandardListResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/subscriptions/orders [get]
func (h *AdminSubscriptionHandler) ListSubscriptionOrders(c *gin.Context) {
	// Parse query parameters
	req := &interfaces.GetSubscriptionOrdersRequest{}

	if err := c.ShouldBindQuery(req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Set defaults
	if req.Limit == 0 {
		req.Limit = 10
	}
	if req.Limit > 100 {
		req.Limit = 100
	}

	orders, total, err := h.subscriptionOrderService.GetSubscriptionOrders(c.Request.Context(), req)
	if err != nil {
		logger.Error("Admin failed to list subscription orders", logger.ErrorField(err))
		response.InternalServerError(c, "Failed to list subscription orders")
		return
	}

	// Convert to response format
	orderResponses := make([]*entities.SubscriptionOrderResponse, len(orders))
	for i, order := range orders {
		orderResponses[i] = order.ToResponse()
	}

	page := (req.Offset / req.Limit) + 1
	response.SuccessList(c, orderResponses, page, req.Limit, total)
}

// CancelSubscriptionOrder godoc
// @Summary Cancel subscription order
// @Description Cancel a pending subscription order (Admin only)
// @Tags Admin-Subscription-Orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Order ID"
// @Param cancel body object{reason=string} true "Cancel data"
// @Success 200 {object} response.MessageOnlyResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/subscriptions/orders/{id}/cancel [post]
func (h *AdminSubscriptionHandler) CancelSubscriptionOrder(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid order ID")
		return
	}

	var cancelData struct {
		Reason string `json:"reason" binding:"required,max=255"`
	}
	if err := c.ShouldBindJSON(&cancelData); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.subscriptionOrderService.CancelSubscriptionOrder(c.Request.Context(), uint(id), cancelData.Reason); err != nil {
		logger.Error("Admin failed to cancel subscription order",
			logger.Uint("order_id", uint(id)),
			logger.ErrorField(err),
		)

		if strings.Contains(err.Error(), "not found") {
			response.NotFound(c, "Subscription order not found")
		} else {
			response.InternalServerError(c, "Failed to cancel order")
		}
		return
	}

	logger.Info("Admin cancelled subscription order",
		logger.Uint("order_id", uint(id)),
		logger.String("reason", cancelData.Reason),
		logger.String("admin_action", "cancel_order"),
	)

	response.SuccessWithMessage(c, "Order cancelled successfully", nil)
}

// GetOrderStatistics godoc
// @Summary Get order statistics
// @Description Get subscription order statistics (Admin only)
// @Tags Admin-Subscription-Analytics
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param from_date query string false "Start date (YYYY-MM-DD)"
// @Param to_date query string false "End date (YYYY-MM-DD)"
// @Success 200 {object} response.StandardResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/subscriptions/analytics/orders [get]
func (h *AdminSubscriptionHandler) GetOrderStatistics(c *gin.Context) {
	fromDateStr := c.Query("from_date")
	toDateStr := c.Query("to_date")

	// Default to last 30 days if not specified
	var fromDate, toDate time.Time
	var err error

	if fromDateStr == "" {
		fromDate = time.Now().AddDate(0, 0, -30)
	} else {
		fromDate, err = time.Parse("2006-01-02", fromDateStr)
		if err != nil {
			response.BadRequest(c, "Invalid from_date format, use YYYY-MM-DD")
			return
		}
	}

	if toDateStr == "" {
		toDate = time.Now()
	} else {
		toDate, err = time.Parse("2006-01-02", toDateStr)
		if err != nil {
			response.BadRequest(c, "Invalid to_date format, use YYYY-MM-DD")
			return
		}
	}

	stats, err := h.subscriptionOrderService.GetOrderStatistics(c.Request.Context(), fromDate, toDate)
	if err != nil {
		logger.Error("Admin failed to get order statistics", logger.ErrorField(err))
		response.InternalServerError(c, "Failed to get order statistics")
		return
	}

	response.Success(c, stats)
}

// USAGE MANAGEMENT

// GetUsageStatistics godoc
// @Summary Get usage statistics
// @Description Get usage statistics for a subscription (Admin only)
// @Tags Admin-Usage-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Subscription ID"
// @Success 200 {object} response.StandardResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/subscriptions/usage/{id}/statistics [get]
func (h *AdminSubscriptionHandler) GetUsageStatistics(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid subscription ID")
		return
	}

	stats, err := h.userSubscriptionService.GetSubscriptionTrafficStats(c.Request.Context(), uint(id))
	if err != nil {
		logger.Error("Admin failed to get usage statistics",
			logger.Uint("subscription_id", uint(id)),
			logger.ErrorField(err),
		)

		if strings.Contains(err.Error(), "not found") {
			response.NotFound(c, "Subscription not found")
		} else {
			response.InternalServerError(c, "Failed to get usage statistics")
		}
		return
	}

	response.Success(c, stats)
}

// GetCurrentUsage godoc
// @Summary Get current usage
// @Description Get current usage for a subscription (Admin only)
// @Tags Admin-Usage-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Subscription ID"
// @Param usage_type query string false "Usage type filter"
// @Success 200 {object} response.StandardResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/subscriptions/usage/{id}/current [get]
func (h *AdminSubscriptionHandler) GetCurrentUsage(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid subscription ID")
		return
	}

	usageType := c.Query("usage_type")
	if usageType == "" {
		usageType = "traffic" // Default to traffic
	}

	usage, err := h.usageTrackingService.GetCurrentUsage(c.Request.Context(), uint(id), usageType)
	if err != nil {
		logger.Error("Admin failed to get current usage",
			logger.Uint("subscription_id", uint(id)),
			logger.String("usage_type", usageType),
			logger.ErrorField(err),
		)

		if strings.Contains(err.Error(), "not found") {
			response.NotFound(c, "Subscription not found")
		} else {
			response.InternalServerError(c, "Failed to get current usage")
		}
		return
	}

	response.Success(c, usage)
}

// BULK OPERATIONS

// BulkSubscriptionAction godoc
// @Summary Bulk subscription actions
// @Description Perform bulk actions on multiple subscriptions (Admin only)
// @Tags Admin-Subscription-Bulk
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param bulk body BulkSubscriptionActionRequest true "Bulk action data"
// @Success 200 {object} response.StandardResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/subscriptions/bulk/action [post]
func (h *AdminSubscriptionHandler) BulkSubscriptionAction(c *gin.Context) {
	var bulkReq BulkSubscriptionActionRequest
	if err := c.ShouldBindJSON(&bulkReq); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	successCount := 0
	failedIDs := make([]uint, 0)
	errors := make([]string, 0)

	for _, subID := range bulkReq.SubscriptionIDs {
		var err error

		switch bulkReq.Action {
		case "pause":
			reason := "Bulk admin pause"
			if bulkReq.Reason != nil {
				reason = *bulkReq.Reason
			}
			pauseReq := &interfaces.PauseSubscriptionRequest{
				Reason: reason,
			}
			_, err = h.userSubscriptionService.PauseUserSubscription(c.Request.Context(), subID, pauseReq, 0)

		case "resume":
			resumeReq := &interfaces.ResumeSubscriptionRequest{
				AdjustBillingDate: true,
			}
			_, err = h.userSubscriptionService.ResumeUserSubscription(c.Request.Context(), subID, resumeReq, 0)

		case "cancel":
			reason := "Bulk admin cancellation"
			if bulkReq.Reason != nil {
				reason = *bulkReq.Reason
			}
			err = h.userSubscriptionService.CancelUserSubscription(c.Request.Context(), subID, reason, false)

		case "extend":
			if bulkReq.ExtendByDays == nil {
				err = fmt.Errorf("extend_by_days is required for extend action")
			} else {
				reason := "Bulk admin extension"
				if bulkReq.Reason != nil {
					reason = *bulkReq.Reason
				}
				err = h.userSubscriptionService.ExtendSubscription(c.Request.Context(), subID, *bulkReq.ExtendByDays, reason)
			}

		case "reset_traffic":
			_, err = h.userSubscriptionService.ResetTrafficUsage(c.Request.Context(), subID, 0)

		default:
			err = fmt.Errorf("unknown action: %s", bulkReq.Action)
		}

		if err != nil {
			failedIDs = append(failedIDs, subID)
			errors = append(errors, fmt.Sprintf("ID %d: %s", subID, err.Error()))
		} else {
			successCount++
		}
	}

	logger.Info("Admin performed bulk subscription action",
		logger.String("action", bulkReq.Action),
		logger.Int("success_count", successCount),
		logger.Int("failed_count", len(failedIDs)),
		logger.String("admin_action", fmt.Sprintf("bulk_%s", bulkReq.Action)),
	)

	result := gin.H{
		"action":        bulkReq.Action,
		"total_count":   len(bulkReq.SubscriptionIDs),
		"success_count": successCount,
		"failed_count":  len(failedIDs),
		"failed_ids":    failedIDs,
		"errors":        errors,
	}

	if len(failedIDs) > 0 {
		response.SuccessWithMessage(c, fmt.Sprintf("Bulk action completed with %d successes and %d failures", successCount, len(failedIDs)), result)
	} else {
		response.SuccessWithMessage(c, fmt.Sprintf("Bulk action completed successfully for all %d subscriptions", successCount), result)
	}
}

// USAGE ALERTS MANAGEMENT

// GetUsageAlerts godoc
// @Summary Get usage alerts
// @Description Get usage alerts with filtering options (Admin only)
// @Tags Admin-Usage-Alerts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param user_subscription_id query int false "Filter by subscription ID"
// @Param usage_type query string false "Filter by usage type"
// @Param status query string false "Filter by status" Enums(fired,resolved,suppressed,acknowledged)
// @Param severity query string false "Filter by severity" Enums(info,warning,error,critical)
// @Param is_active query bool false "Filter by active status"
// @Param limit query int false "Items per page" default(50)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} response.StandardListResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/subscriptions/alerts [get]
func (h *AdminSubscriptionHandler) GetUsageAlerts(c *gin.Context) {
	req := &interfaces.GetUsageAlertsRequest{}
	if err := c.ShouldBindQuery(req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Set defaults
	if req.Limit == 0 {
		req.Limit = 50
	}
	if req.Limit > 1000 {
		req.Limit = 1000
	}

	alertsResponse, err := h.usageAlertService.GetUsageAlerts(c.Request.Context(), req)
	if err != nil {
		logger.Error("Admin failed to get usage alerts", logger.ErrorField(err))
		response.InternalServerError(c, "Failed to get usage alerts")
		return
	}

	page := (req.Offset / req.Limit) + 1
	response.SuccessList(c, alertsResponse.UsageAlerts, page, req.Limit, alertsResponse.TotalCount)
}

// GetAlertStatistics godoc
// @Summary Get alert statistics
// @Description Get usage alert statistics (Admin only)
// @Tags Admin-Usage-Alerts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param period query string false "Statistics period" Enums(24h,7d,30d,90d,365d) default(7d)
// @Param usage_type query string false "Filter by usage type"
// @Param severity query string false "Filter by severity"
// @Success 200 {object} response.StandardResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/subscriptions/alerts/statistics [get]
func (h *AdminSubscriptionHandler) GetAlertStatistics(c *gin.Context) {
	req := &interfaces.AlertStatsRequest{
		Period: c.DefaultQuery("period", "7d"),
	}

	if err := c.ShouldBindQuery(req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	stats, err := h.usageAlertService.GetAlertStatistics(c.Request.Context(), req)
	if err != nil {
		logger.Error("Admin failed to get alert statistics", logger.ErrorField(err))
		response.InternalServerError(c, "Failed to get alert statistics")
		return
	}

	response.Success(c, stats)
}

// BulkResolveAlerts godoc
// @Summary Bulk resolve alerts
// @Description Resolve multiple usage alerts (Admin only)
// @Tags Admin-Usage-Alerts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param bulk body interfaces.BulkResolveAlertsRequest true "Bulk resolve data"
// @Success 200 {object} response.StandardResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/subscriptions/alerts/bulk/resolve [post]
func (h *AdminSubscriptionHandler) BulkResolveAlerts(c *gin.Context) {
	var bulkReq interfaces.BulkResolveAlertsRequest
	if err := c.ShouldBindJSON(&bulkReq); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.usageAlertService.BulkResolveAlerts(c.Request.Context(), &bulkReq)
	if err != nil {
		logger.Error("Admin failed to bulk resolve alerts",
			logger.Any("alert_ids", bulkReq.AlertIDs),
			logger.ErrorField(err),
		)
		response.InternalServerError(c, "Failed to resolve alerts")
		return
	}

	logger.Info("Admin bulk resolved alerts",
		logger.Int64("resolved_count", result.ResolvedCount),
		logger.Int("failed_count", len(result.FailedIDs)),
		logger.String("admin_action", "bulk_resolve_alerts"),
	)

	response.Success(c, result)
}
