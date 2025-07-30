package admin

import (
	"strconv"
	"time"

	"linke/internal/logger"
	"linke/internal/middleware"
	"linke/internal/model"
	"linke/internal/response"
	"linke/internal/service"

	"github.com/gin-gonic/gin"
)

type AdminSubscriptionHandler struct {
	subscriptionPlanService *service.SubscriptionPlanService
	userSubscriptionService *service.UserSubscriptionService
	subscriptionOrderService *service.SubscriptionOrderService
}

func NewAdminSubscriptionHandler(
	subscriptionPlanService *service.SubscriptionPlanService,
	userSubscriptionService *service.UserSubscriptionService,
	subscriptionOrderService *service.SubscriptionOrderService,
) *AdminSubscriptionHandler {
	return &AdminSubscriptionHandler{
		subscriptionPlanService:  subscriptionPlanService,
		userSubscriptionService: userSubscriptionService,
		subscriptionOrderService: subscriptionOrderService,
	}
}

// ============= Admin Subscription Plan Management =============

// CreateSubscriptionPlan godoc
// @Summary [Admin] Create subscription plan
// @Description Create a new subscription plan (Admin only)
// @Tags Admin-Subscription-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param plan body service.CreateSubscriptionPlanRequest true "Subscription plan data"
// @Success 201 {object} response.StandardResponse{data=model.SubscriptionPlanResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/subscriptions/plans [post]
func (h *AdminSubscriptionHandler) CreateSubscriptionPlan(c *gin.Context) {
	// Get current user from context
	userValue, exists := c.Get(middleware.AuthContextKey)
	if !exists {
		response.Unauthorized(c, "Authentication required")
		return
	}

	user, ok := userValue.(*model.User)
	if !ok {
		response.Unauthorized(c, "Invalid user context")
		return
	}

	// Check if user is admin
	if !user.IsAdmin() {
		response.Forbidden(c, "Admin access required")
		return
	}

	// Bind request
	var req service.CreateSubscriptionPlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request data", err.Error())
		return
	}

	// Create subscription plan
	plan, err := h.subscriptionPlanService.CreateSubscriptionPlan(c.Request.Context(), user.ID, &req)
	if err != nil {
		logger.Error("Failed to create subscription plan", logger.Error2("error", err), logger.Uint("admin_id", user.ID))
		response.InternalServerError(c, "Failed to create subscription plan", err.Error())
		return
	}

	response.CreatedWithMessage(c, "Subscription plan created successfully", plan.ToResponse())
}

// ListSubscriptionPlans godoc
// @Summary [Admin] List all subscription plans
// @Description Get all subscription plans with full details (Admin only)
// @Tags Admin-Subscription-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param status query string false "Filter by status" Enums(active, inactive, archived)
// @Param currency query string false "Filter by currency" example(USD)
// @Param limit query int false "Limit results" minimum(1) maximum(100) example(10)
// @Param offset query int false "Offset results" minimum(0) example(0)
// @Success 200 {object} response.PaginatedResponse{data=[]model.SubscriptionPlanResponse}
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/subscriptions/plans [get]
func (h *AdminSubscriptionHandler) ListSubscriptionPlans(c *gin.Context) {
	// Get current user from context
	userValue, exists := c.Get(middleware.AuthContextKey)
	if !exists {
		response.Unauthorized(c, "Authentication required")
		return
	}

	user, ok := userValue.(*model.User)
	if !ok {
		response.Unauthorized(c, "Invalid user context")
		return
	}

	// Check if user is admin
	if !user.IsAdmin() {
		response.Forbidden(c, "Admin access required")
		return
	}

	// Bind query parameters
	var req service.GetSubscriptionPlansRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "Invalid query parameters", err.Error())
		return
	}

	// Get all subscription plans (including hidden ones)
	plans, totalCount, err := h.subscriptionPlanService.GetSubscriptionPlans(c.Request.Context(), &req)
	if err != nil {
		logger.Error("Failed to get subscription plans", logger.Error2("error", err))
		response.InternalServerError(c, "Failed to get subscription plans", err.Error())
		return
	}

	// Convert to admin response format (full details)
	var planResponses []*model.SubscriptionPlanResponse
	for _, plan := range plans {
		planResponses = append(planResponses, plan.ToResponse())
	}

	response.OKPaginated(c, "Subscription plans retrieved successfully", planResponses, totalCount, req.Limit, req.Offset)
}

// GetSubscriptionPlan godoc
// @Summary [Admin] Get subscription plan by ID
// @Description Get a subscription plan by ID with full details (Admin only)
// @Tags Admin-Subscription-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Subscription plan ID"
// @Success 200 {object} response.StandardResponse{data=model.SubscriptionPlanResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/subscriptions/plans/{id} [get]
func (h *AdminSubscriptionHandler) GetSubscriptionPlan(c *gin.Context) {
	// Get current user from context
	userValue, exists := c.Get(middleware.AuthContextKey)
	if !exists {
		response.Unauthorized(c, "Authentication required")
		return
	}

	user, ok := userValue.(*model.User)
	if !ok {
		response.Unauthorized(c, "Invalid user context")
		return
	}

	// Check if user is admin
	if !user.IsAdmin() {
		response.Forbidden(c, "Admin access required")
		return
	}

	// Parse plan ID
	planIDStr := c.Param("id")
	planID, err := strconv.ParseUint(planIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid plan ID", "Plan ID must be a valid number")
		return
	}

	// Get subscription plan
	plan, err := h.subscriptionPlanService.GetSubscriptionPlan(c.Request.Context(), uint(planID))
	if err != nil {
		if err.Error() == "subscription plan not found" {
			response.NotFound(c, "Subscription plan not found")
			return
		}
		logger.Error("Failed to get subscription plan", logger.Error2("error", err), logger.Uint("plan_id", uint(planID)))
		response.InternalServerError(c, "Failed to get subscription plan", err.Error())
		return
	}

	response.OK(c, "Subscription plan retrieved successfully", plan.ToResponse())
}

// UpdateSubscriptionPlan godoc
// @Summary [Admin] Update subscription plan
// @Description Update a subscription plan (Admin only)
// @Tags Admin-Subscription-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Subscription plan ID"
// @Param plan body service.UpdateSubscriptionPlanRequest true "Updated subscription plan data"
// @Success 200 {object} response.StandardResponse{data=model.SubscriptionPlanResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/subscriptions/plans/{id} [put]
func (h *AdminSubscriptionHandler) UpdateSubscriptionPlan(c *gin.Context) {
	// Get current user from context
	userValue, exists := c.Get(middleware.AuthContextKey)
	if !exists {
		response.Unauthorized(c, "Authentication required")
		return
	}

	user, ok := userValue.(*model.User)
	if !ok {
		response.Unauthorized(c, "Invalid user context")
		return
	}

	// Check if user is admin
	if !user.IsAdmin() {
		response.Forbidden(c, "Admin access required")
		return
	}

	// Parse plan ID
	planIDStr := c.Param("id")
	planID, err := strconv.ParseUint(planIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid plan ID", "Plan ID must be a valid number")
		return
	}

	// Bind request
	var req service.UpdateSubscriptionPlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request data", err.Error())
		return
	}

	// Update subscription plan
	plan, err := h.subscriptionPlanService.UpdateSubscriptionPlan(c.Request.Context(), uint(planID), &req)
	if err != nil {
		if err.Error() == "subscription plan not found" {
			response.NotFound(c, "Subscription plan not found")
			return
		}
		logger.Error("Failed to update subscription plan", logger.Error2("error", err), logger.Uint("plan_id", uint(planID)), logger.Uint("admin_id", user.ID))
		response.InternalServerError(c, "Failed to update subscription plan", err.Error())
		return
	}

	response.OK(c, "Subscription plan updated successfully", plan.ToResponse())
}

// PatchSubscriptionPlan godoc
// @Summary [Admin] Partially update subscription plan
// @Description Partially update a subscription plan with only provided fields (Admin only)
// @Tags Admin-Subscription-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Subscription plan ID"
// @Param plan body service.UpdateSubscriptionPlanRequest true "Subscription plan fields to update"
// @Success 200 {object} response.StandardResponse{data=model.SubscriptionPlanResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/subscriptions/plans/{id} [patch]
func (h *AdminSubscriptionHandler) PatchSubscriptionPlan(c *gin.Context) {
	// Get current user from context
	userValue, exists := c.Get(middleware.AuthContextKey)
	if !exists {
		response.Unauthorized(c, "Authentication required")
		return
	}

	user, ok := userValue.(*model.User)
	if !ok {
		response.Unauthorized(c, "Invalid user context")
		return
	}

	// Check if user is admin
	if !user.IsAdmin() {
		response.Forbidden(c, "Admin access required")
		return
	}

	// Parse plan ID
	planIDStr := c.Param("id")
	planID, err := strconv.ParseUint(planIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid plan ID", "Plan ID must be a valid number")
		return
	}

	// Bind request
	var req service.UpdateSubscriptionPlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request data", err.Error())
		return
	}

	// Update subscription plan (same logic as PUT since service layer already handles partial updates)
	plan, err := h.subscriptionPlanService.UpdateSubscriptionPlan(c.Request.Context(), uint(planID), &req)
	if err != nil {
		if err.Error() == "subscription plan not found" {
			response.NotFound(c, "Subscription plan not found")
			return
		}
		logger.Error("Failed to patch subscription plan", logger.Error2("error", err), logger.Uint("plan_id", uint(planID)), logger.Uint("admin_id", user.ID))
		response.InternalServerError(c, "Failed to patch subscription plan", err.Error())
		return
	}

	response.OK(c, "Subscription plan patched successfully", plan.ToResponse())
}

// DeleteSubscriptionPlan godoc
// @Summary [Admin] Delete subscription plan
// @Description Soft delete a subscription plan (Admin only)
// @Tags Admin-Subscription-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Subscription plan ID"
// @Success 200 {object} response.StandardResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/subscriptions/plans/{id} [delete]
func (h *AdminSubscriptionHandler) DeleteSubscriptionPlan(c *gin.Context) {
	// Get current user from context
	userValue, exists := c.Get(middleware.AuthContextKey)
	if !exists {
		response.Unauthorized(c, "Authentication required")
		return
	}

	user, ok := userValue.(*model.User)
	if !ok {
		response.Unauthorized(c, "Invalid user context")
		return
	}

	// Check if user is admin
	if !user.IsAdmin() {
		response.Forbidden(c, "Admin access required")
		return
	}

	// Parse plan ID
	planIDStr := c.Param("id")
	planID, err := strconv.ParseUint(planIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid plan ID", "Plan ID must be a valid number")
		return
	}

	// Delete subscription plan
	if err := h.subscriptionPlanService.DeleteSubscriptionPlan(c.Request.Context(), uint(planID)); err != nil {
		if err.Error() == "subscription plan not found" {
			response.NotFound(c, "Subscription plan not found")
			return
		}
		logger.Error("Failed to delete subscription plan", logger.Error2("error", err), logger.Uint("plan_id", uint(planID)), logger.Uint("admin_id", user.ID))
		response.InternalServerError(c, "Failed to delete subscription plan", err.Error())
		return
	}

	response.OK(c, "Subscription plan deleted successfully", nil)
}

// ============= Admin User Subscription Management =============

// CreateUserSubscriptionRequest represents the request to create a user subscription
type CreateUserSubscriptionRequest struct {
	UserID             uint    `json:"user_id" binding:"required" example:"1"`
	SubscriptionPlanID uint    `json:"subscription_plan_id" binding:"required" example:"1"`
	StartDate          string  `json:"start_date,omitempty" example:"2024-01-01T00:00:00Z"`
	EndDate            string  `json:"end_date,omitempty" example:"2024-12-31T23:59:59Z"`
	UseTrial           bool    `json:"use_trial,omitempty" example:"false"`
	ServerGroupIDs     []uint  `json:"server_group_ids,omitempty"`
	Notes              string  `json:"notes,omitempty" example:"Admin assigned subscription"`
	SkipPayment        bool    `json:"skip_payment,omitempty" example:"true"`
}

// CreateUserSubscription godoc
// @Summary [Admin] Create user subscription
// @Description Create a new user subscription directly (Admin only)
// @Tags Admin-Subscription-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param subscription body CreateUserSubscriptionRequest true "User subscription data"
// @Success 201 {object} response.StandardResponse{data=model.UserSubscriptionResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/subscriptions/users [post]
func (h *AdminSubscriptionHandler) CreateUserSubscription(c *gin.Context) {
	// Get current user from context
	userValue, exists := c.Get(middleware.AuthContextKey)
	if !exists {
		response.Unauthorized(c, "Authentication required")
		return
	}

	user, ok := userValue.(*model.User)
	if !ok {
		response.Unauthorized(c, "Invalid user context")
		return
	}

	// Check if user is admin
	if !user.IsAdmin() {
		response.Forbidden(c, "Admin access required")
		return
	}

	// Bind request
	var req CreateUserSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request data", err.Error())
		return
	}

	// Create service request  
	createReq := &service.CreateSubscriptionRequest{
		UserID:             req.UserID,
		SubscriptionPlanID: req.SubscriptionPlanID,
		StartDate:          req.StartDate,
		UseTrial:           req.UseTrial,
		ServerGroupIDs:     req.ServerGroupIDs,
	}

	// Create user subscription
	subscription, err := h.userSubscriptionService.CreateUserSubscription(c.Request.Context(), createReq)
	if err != nil {
		logger.Error("Failed to create user subscription", logger.Error2("error", err), logger.Uint("admin_id", user.ID), logger.Uint("target_user_id", req.UserID))
		response.InternalServerError(c, "Failed to create user subscription", err.Error())
		return
	}

	// Update notes if provided
	if req.Notes != "" {
		updateReq := &service.UpdateSubscriptionRequest{
			Notes: &req.Notes,
		}
		if _, err := h.userSubscriptionService.UpdateUserSubscription(c.Request.Context(), subscription.ID, updateReq); err != nil {
			logger.Error("Failed to update subscription notes", logger.Error2("error", err), logger.Uint("subscription_id", subscription.ID))
			// Don't fail the entire operation for notes update
		}
	}

	// Set custom end date if provided
	if req.EndDate != "" {
		if endDate, err := time.Parse(time.RFC3339, req.EndDate); err == nil {
			updateReq := &service.UpdateSubscriptionRequest{
				EndDate: &endDate,
			}
			if _, err := h.userSubscriptionService.UpdateUserSubscription(c.Request.Context(), subscription.ID, updateReq); err != nil {
				logger.Error("Failed to update subscription end date", logger.Error2("error", err), logger.Uint("subscription_id", subscription.ID))
			}
		}
	}

	logger.Info("Admin created user subscription", 
		logger.Uint("admin_id", user.ID),
		logger.Uint("target_user_id", req.UserID),
		logger.Uint("subscription_id", subscription.ID),
		logger.Uint("plan_id", req.SubscriptionPlanID))

	response.CreatedWithMessage(c, "User subscription created successfully", subscription.ToResponse())
}

// ListUserSubscriptions godoc
// @Summary [Admin] List all user subscriptions
// @Description Get all user subscriptions with full details including user and subscription plan information (Admin only)
// @Tags Admin-Subscription-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param user_id query int false "Filter by user ID"
// @Param status query string false "Filter by status" Enums(active, inactive, cancelled, expired, trialing, past_due)
// @Param plan_id query int false "Filter by subscription plan ID"
// @Param limit query int false "Limit results" minimum(1) maximum(100) example(10)
// @Param offset query int false "Offset results" minimum(0) example(0)
// @Success 200 {object} response.PaginatedResponse{data=[]model.UserSubscriptionResponse}
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/subscriptions/users [get]
func (h *AdminSubscriptionHandler) ListUserSubscriptions(c *gin.Context) {
	// Get current user from context
	userValue, exists := c.Get(middleware.AuthContextKey)
	if !exists {
		response.Unauthorized(c, "Authentication required")
		return
	}

	user, ok := userValue.(*model.User)
	if !ok {
		response.Unauthorized(c, "Invalid user context")
		return
	}

	// Check if user is admin
	if !user.IsAdmin() {
		response.Forbidden(c, "Admin access required")
		return
	}

	// Bind query parameters
	var req service.GetUserSubscriptionsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "Invalid query parameters", err.Error())
		return
	}

	// Get all user subscriptions with related user and plan data
	subscriptions, totalCount, err := h.userSubscriptionService.GetUserSubscriptionsWithRelations(c.Request.Context(), &req)
	if err != nil {
		logger.Error("Failed to get user subscriptions", logger.Error2("error", err))
		response.InternalServerError(c, "Failed to get user subscriptions", err.Error())
		return
	}

	// Convert to admin response format (full details with relations)
	var subscriptionResponses []*model.UserSubscriptionResponse
	for _, subscription := range subscriptions {
		subscriptionResponses = append(subscriptionResponses, subscription.ToResponse())
	}

	response.OKPaginated(c, "User subscriptions retrieved successfully", subscriptionResponses, totalCount, req.Limit, req.Offset)
}

// GetUserSubscription godoc
// @Summary [Admin] Get user subscription by ID
// @Description Get a user subscription by ID with full details (Admin only)
// @Tags Admin-Subscription-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User subscription ID"
// @Success 200 {object} response.StandardResponse{data=model.UserSubscriptionResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/subscriptions/users/{id} [get]
func (h *AdminSubscriptionHandler) GetUserSubscription(c *gin.Context) {
	// Get current user from context
	userValue, exists := c.Get(middleware.AuthContextKey)
	if !exists {
		response.Unauthorized(c, "Authentication required")
		return
	}

	user, ok := userValue.(*model.User)
	if !ok {
		response.Unauthorized(c, "Invalid user context")
		return
	}

	// Check if user is admin
	if !user.IsAdmin() {
		response.Forbidden(c, "Admin access required")
		return
	}

	// Parse subscription ID
	subscriptionIDStr := c.Param("id")
	subscriptionID, err := strconv.ParseUint(subscriptionIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid subscription ID", "Subscription ID must be a valid number")
		return
	}

	// Get user subscription with relations
	subscription, err := h.userSubscriptionService.GetUserSubscriptionWithRelations(c.Request.Context(), uint(subscriptionID))
	if err != nil {
		if err.Error() == "subscription not found" {
			response.NotFound(c, "User subscription not found")
			return
		}
		logger.Error("Failed to get user subscription", logger.Error2("error", err), logger.Uint("subscription_id", uint(subscriptionID)))
		response.InternalServerError(c, "Failed to get user subscription", err.Error())
		return
	}

	response.OK(c, "User subscription retrieved successfully", subscription.ToResponse())
}

// UpdateUserSubscription godoc
// @Summary [Admin] Update user subscription
// @Description Update a user subscription (Admin only)
// @Tags Admin-Subscription-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User subscription ID"
// @Param subscription body service.UpdateSubscriptionRequest true "Updated subscription data"
// @Success 200 {object} response.StandardResponse{data=model.UserSubscriptionResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/subscriptions/users/{id} [put]
func (h *AdminSubscriptionHandler) UpdateUserSubscription(c *gin.Context) {
	// Get current user from context
	userValue, exists := c.Get(middleware.AuthContextKey)
	if !exists {
		response.Unauthorized(c, "Authentication required")
		return
	}

	user, ok := userValue.(*model.User)
	if !ok {
		response.Unauthorized(c, "Invalid user context")
		return
	}

	// Check if user is admin
	if !user.IsAdmin() {
		response.Forbidden(c, "Admin access required")
		return
	}

	// Parse subscription ID
	subscriptionIDStr := c.Param("id")
	subscriptionID, err := strconv.ParseUint(subscriptionIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid subscription ID", "Subscription ID must be a valid number")
		return
	}

	// Bind request
	var req service.UpdateSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request data", err.Error())
		return
	}

	// Update user subscription
	subscription, err := h.userSubscriptionService.UpdateUserSubscription(c.Request.Context(), uint(subscriptionID), &req)
	if err != nil {
		if err.Error() == "subscription not found" {
			response.NotFound(c, "User subscription not found")
			return
		}
		logger.Error("Failed to update user subscription", logger.Error2("error", err), logger.Uint("subscription_id", uint(subscriptionID)), logger.Uint("admin_id", user.ID))
		response.InternalServerError(c, "Failed to update user subscription", err.Error())
		return
	}

	response.OK(c, "User subscription updated successfully", subscription.ToResponse())
}

// PatchUserSubscription godoc
// @Summary [Admin] Partially update user subscription
// @Description Partially update a user subscription with only provided fields (Admin only)
// @Tags Admin-Subscription-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User subscription ID"
// @Param subscription body service.UpdateSubscriptionRequest true "Subscription fields to update"
// @Success 200 {object} response.StandardResponse{data=model.UserSubscriptionResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/subscriptions/users/{id} [patch]
func (h *AdminSubscriptionHandler) PatchUserSubscription(c *gin.Context) {
	// Get current user from context
	userValue, exists := c.Get(middleware.AuthContextKey)
	if !exists {
		response.Unauthorized(c, "Authentication required")
		return
	}

	user, ok := userValue.(*model.User)
	if !ok {
		response.Unauthorized(c, "Invalid user context")
		return
	}

	// Check if user is admin
	if !user.IsAdmin() {
		response.Forbidden(c, "Admin access required")
		return
	}

	// Parse subscription ID
	subscriptionIDStr := c.Param("id")
	subscriptionID, err := strconv.ParseUint(subscriptionIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid subscription ID", "Subscription ID must be a valid number")
		return
	}

	// Bind request
	var req service.UpdateSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request data", err.Error())
		return
	}

	// Update user subscription (same logic as PUT since service layer already handles partial updates)
	subscription, err := h.userSubscriptionService.UpdateUserSubscription(c.Request.Context(), uint(subscriptionID), &req)
	if err != nil {
		if err.Error() == "subscription not found" {
			response.NotFound(c, "User subscription not found")
			return
		}
		logger.Error("Failed to patch user subscription", logger.Error2("error", err), logger.Uint("subscription_id", uint(subscriptionID)), logger.Uint("admin_id", user.ID))
		response.InternalServerError(c, "Failed to patch user subscription", err.Error())
		return
	}

	response.OK(c, "User subscription patched successfully", subscription.ToResponse())
}

// RenewUserSubscription godoc
// @Summary [Admin] Renew user subscription
// @Description Manually renew a user subscription (Admin only)
// @Tags Admin-Subscription-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User subscription ID"
// @Success 200 {object} response.StandardResponse{data=model.UserSubscriptionResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/subscriptions/users/{id}/renew [post]
func (h *AdminSubscriptionHandler) RenewUserSubscription(c *gin.Context) {
	// Get current user from context
	userValue, exists := c.Get(middleware.AuthContextKey)
	if !exists {
		response.Unauthorized(c, "Authentication required")
		return
	}

	user, ok := userValue.(*model.User)
	if !ok {
		response.Unauthorized(c, "Invalid user context")
		return
	}

	// Check if user is admin
	if !user.IsAdmin() {
		response.Forbidden(c, "Admin access required")
		return
	}

	// Parse subscription ID
	subscriptionIDStr := c.Param("id")
	subscriptionID, err := strconv.ParseUint(subscriptionIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid subscription ID", "Subscription ID must be a valid number")
		return
	}

	// Renew user subscription
	subscription, err := h.userSubscriptionService.RenewUserSubscription(c.Request.Context(), uint(subscriptionID))
	if err != nil {
		if err.Error() == "subscription not found" {
			response.NotFound(c, "User subscription not found")
			return
		}
		logger.Error("Failed to renew user subscription", logger.Error2("error", err), logger.Uint("subscription_id", uint(subscriptionID)), logger.Uint("admin_id", user.ID))
		response.InternalServerError(c, "Failed to renew user subscription", err.Error())
		return
	}

	response.OK(c, "User subscription renewed successfully", subscription.ToResponse())
}

// DeleteUserSubscription godoc
// @Summary [Admin] Delete user subscription
// @Description Soft delete a user subscription (Admin only)
// @Tags Admin-Subscription-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User subscription ID"
// @Success 200 {object} response.StandardResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/subscriptions/users/{id} [delete]
func (h *AdminSubscriptionHandler) DeleteUserSubscription(c *gin.Context) {
	// Get current user from context
	userValue, exists := c.Get(middleware.AuthContextKey)
	if !exists {
		response.Unauthorized(c, "Authentication required")
		return
	}

	user, ok := userValue.(*model.User)
	if !ok {
		response.Unauthorized(c, "Invalid user context")
		return
	}

	// Check if user is admin
	if !user.IsAdmin() {
		response.Forbidden(c, "Admin access required")
		return
	}

	// Parse subscription ID
	subscriptionIDStr := c.Param("id")
	subscriptionID, err := strconv.ParseUint(subscriptionIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid subscription ID", "Subscription ID must be a valid number")
		return
	}

	// Delete user subscription
	if err := h.userSubscriptionService.DeleteUserSubscription(c.Request.Context(), uint(subscriptionID)); err != nil {
		if err.Error() == "subscription not found" {
			response.NotFound(c, "User subscription not found")
			return
		}
		logger.Error("Failed to delete user subscription", logger.Error2("error", err), logger.Uint("subscription_id", uint(subscriptionID)), logger.Uint("admin_id", user.ID))
		response.InternalServerError(c, "Failed to delete user subscription", err.Error())
		return
	}

	response.OK(c, "User subscription deleted successfully", nil)
}