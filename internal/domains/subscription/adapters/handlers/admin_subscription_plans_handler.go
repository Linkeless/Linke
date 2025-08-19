package handlers

import (
	"strconv"
	"strings"

	"linke/internal/domains/subscription/entities"
	"linke/internal/domains/subscription/usecases/interfaces"
	"linke/internal/shared/logger"
	"linke/internal/shared/response"

	"github.com/gin-gonic/gin"
)

// AdminSubscriptionPlansHandler handles subscription plan management operations
type AdminSubscriptionPlansHandler struct {
	*AdminSubscriptionHandlerBase
}

// NewAdminSubscriptionPlansHandler creates a new admin subscription plans handler
func NewAdminSubscriptionPlansHandler(base *AdminSubscriptionHandlerBase) *AdminSubscriptionPlansHandler {
	return &AdminSubscriptionPlansHandler{
		AdminSubscriptionHandlerBase: base,
	}
}

// Request/Response structures for plan operations are defined in the main handler file

// CreateSubscriptionPlan godoc
// @Summary Create subscription plan
// @Description Create a new subscription plan (Admin only)
// @Tags Admin-Subscription
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param plan body CreatePlanRequest true "Plan creation data" example({"name":"Premium Plan","code":"premium-monthly","description":"Premium features with monthly billing","price":29.99,"currency":"USD","billing_cycle":"monthly","billing_interval":1,"trial_period_days":7,"features":"{\"max_projects\": 10, \"storage_gb\": 100}","limits":"{\"api_calls_per_month\": 10000}","is_visible":true,"sort_order":1,"is_popular":false,"is_recommended":true,"setup_fee":0,"cancellation_fee":0,"traffic_limit":107374182400,"traffic_reset_cycle":"monthly","default_server_group_ids":[1]})
// @Success 201 {object} entities.SubscriptionPlanResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 409 {object} response.ConflictResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/subscriptions/plans [post]
func (h *AdminSubscriptionPlansHandler) CreateSubscriptionPlan(c *gin.Context) {
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
// @Success 200 {object} entities.SubscriptionPlanResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Router /admin/subscriptions/plans/{id} [get]
func (h *AdminSubscriptionPlansHandler) GetSubscriptionPlan(c *gin.Context) {
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

	response.OK(c, plan.ToResponse())
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
// @Success 200 {object} response.HALCollectionResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/subscriptions/plans [get]
func (h *AdminSubscriptionPlansHandler) ListSubscriptionPlans(c *gin.Context) {
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

	_ = (req.Offset / req.Limit) + 1 // page calculation for future use
	response.SendPaginatedResponse(c, planResponses, total)
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
// @Success 200 {object} entities.SubscriptionPlanResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/subscriptions/plans/{id} [put]
func (h *AdminSubscriptionPlansHandler) UpdateSubscriptionPlan(c *gin.Context) {
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

	response.OK(c, plan.ToResponse())
}

// DeleteSubscriptionPlan godoc
// @Summary Delete subscription plan
// @Description Soft delete a subscription plan (Admin only)
// @Tags Admin-Subscription
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Plan ID"
// @Success 200 {object} string
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Router /admin/subscriptions/plans/{id} [delete]
func (h *AdminSubscriptionPlansHandler) DeleteSubscriptionPlan(c *gin.Context) {
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

	response.NoContent(c)
}

// ToggleSubscriptionPlanStatus godoc
// @Summary Toggle subscription plan status
// @Description Toggle subscription plan active/inactive status (Admin only)
// @Tags Admin-Subscription
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Plan ID"
// @Success 200 {object} entities.SubscriptionPlanResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Router /admin/subscriptions/plans/{id}/toggle-status [put]
func (h *AdminSubscriptionPlansHandler) ToggleSubscriptionPlanStatus(c *gin.Context) {
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

	response.OK(c, plan.ToResponse())
}