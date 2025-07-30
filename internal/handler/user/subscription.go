package user

import (
	"fmt"
	"strconv"
	"time"

	"linke/internal/logger"
	"linke/internal/middleware"
	"linke/internal/model"
	"linke/internal/response"
	"linke/internal/service"

	"github.com/gin-gonic/gin"
)

type UserSubscriptionPublicHandler struct {
	subscriptionPlanService  *service.SubscriptionPlanService
	userSubscriptionService  *service.UserSubscriptionService
	subscriptionOrderService *service.SubscriptionOrderService
	couponService            *service.CouponService
	subscriptionExpiryService *service.SubscriptionExpiryService
}

func NewUserSubscriptionPublicHandler(
	subscriptionPlanService *service.SubscriptionPlanService,
	userSubscriptionService *service.UserSubscriptionService,
	subscriptionOrderService *service.SubscriptionOrderService,
	couponService *service.CouponService,
	subscriptionExpiryService *service.SubscriptionExpiryService,
) *UserSubscriptionPublicHandler {
	return &UserSubscriptionPublicHandler{
		subscriptionPlanService:  subscriptionPlanService,
		userSubscriptionService:  userSubscriptionService,
		subscriptionOrderService: subscriptionOrderService,
		couponService:            couponService,
		subscriptionExpiryService: subscriptionExpiryService,
	}
}

// ============= User Subscription Plan View =============

// GetPublicSubscriptionPlans godoc
// @Summary [Public] Get public subscription plans
// @Description Get visible and active subscription plans for public display
// @Tags User-Subscription
// @Accept json
// @Produce json
// @Param currency query string false "Filter by currency" example(USD)
// @Success 200 {object} response.StandardResponse{data=[]model.SubscriptionPlanResponse}
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /subscription-plans [get]
func (h *UserSubscriptionPublicHandler) GetPublicSubscriptionPlans(c *gin.Context) {
	currency := c.Query("currency")

	// Get public subscription plans
	plans, err := h.subscriptionPlanService.GetPublicSubscriptionPlans(c.Request.Context(), currency)
	if err != nil {
		logger.Error("Failed to get public subscription plans", logger.Error2("error", err))
		response.InternalServerError(c, "Failed to get subscription plans", err.Error())
		return
	}

	// Convert to public response format
	var planResponses []*model.SubscriptionPlanResponse
	for _, plan := range plans {
		planResponses = append(planResponses, plan.ToPublicResponse())
	}

	response.OK(c, "Public subscription plans retrieved successfully", planResponses)
}

// GetSubscriptionPlan godoc
// @Summary [Public] Get subscription plan by ID
// @Description Get a subscription plan by ID (public information only)
// @Tags User-Subscription
// @Accept json
// @Produce json
// @Param id path int true "Subscription plan ID"
// @Success 200 {object} response.StandardResponse{data=model.SubscriptionPlanResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /subscription-plans/{id} [get]
func (h *UserSubscriptionPublicHandler) GetSubscriptionPlan(c *gin.Context) {
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

	// Return public response
	response.OK(c, "Subscription plan retrieved successfully", plan.ToPublicResponse())
}

// GetSubscriptionPlanByCode godoc
// @Summary [Public] Get subscription plan by code
// @Description Get a subscription plan by code (public information only)
// @Tags User-Subscription
// @Accept json
// @Produce json
// @Param code path string true "Subscription plan code"
// @Success 200 {object} response.StandardResponse{data=model.SubscriptionPlanResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /subscription-plans/code/{code} [get]
func (h *UserSubscriptionPublicHandler) GetSubscriptionPlanByCode(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		response.BadRequest(c, "Invalid plan code", "Plan code cannot be empty")
		return
	}

	// Get subscription plan
	plan, err := h.subscriptionPlanService.GetSubscriptionPlanByCode(c.Request.Context(), code)
	if err != nil {
		if err.Error() == "subscription plan not found" {
			response.NotFound(c, "Subscription plan not found")
			return
		}
		logger.Error("Failed to get subscription plan by code", logger.Error2("error", err), logger.String("code", code))
		response.InternalServerError(c, "Failed to get subscription plan", err.Error())
		return
	}

	// Return public response
	response.OK(c, "Subscription plan retrieved successfully", plan.ToPublicResponse())
}

// ============= User Subscription Management =============

// GetMySubscriptions godoc
// @Summary [User] Get my subscriptions
// @Description Get current user's subscriptions
// @Tags User-Subscription
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param status query string false "Filter by status" Enums(active, inactive, cancelled, expired, trialing, past_due)
// @Success 200 {object} response.StandardResponse{data=[]model.UserSubscriptionResponse}
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /user/subscriptions [get]
func (h *UserSubscriptionPublicHandler) GetMySubscriptions(c *gin.Context) {
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

	// Get query parameters
	status := c.Query("status")

	// Create request for current user
	req := &service.GetUserSubscriptionsRequest{
		UserID: user.ID,
		Status: status,
	}

	// Get user subscriptions
	subscriptions, _, err := h.userSubscriptionService.GetUserSubscriptions(c.Request.Context(), req)
	if err != nil {
		logger.Error("Failed to get user subscriptions", logger.Error2("error", err), logger.Uint("user_id", user.ID))
		response.InternalServerError(c, "Failed to get user subscriptions", err.Error())
		return
	}

	// Convert to user response format (without admin details)
	var subscriptionResponses []*model.UserSubscriptionResponse
	for _, subscription := range subscriptions {
		subscriptionResponses = append(subscriptionResponses, subscription.ToUserResponse())
	}

	response.OK(c, "My subscriptions retrieved successfully", subscriptionResponses)
}

// GetMySubscription godoc
// @Summary [User] Get my subscription by ID
// @Description Get current user's subscription by ID
// @Tags User-Subscription
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
// @Router /user/subscriptions/{id} [get]
func (h *UserSubscriptionPublicHandler) GetMySubscription(c *gin.Context) {
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
			response.NotFound(c, "Subscription not found")
			return
		}
		logger.Error("Failed to get user subscription", logger.Error2("error", err), logger.Uint("subscription_id", uint(subscriptionID)))
		response.InternalServerError(c, "Failed to get user subscription", err.Error())
		return
	}

	// Check if user owns this subscription
	if subscription.UserID != user.ID {
		response.Forbidden(c, "You can only access your own subscriptions")
		return
	}

	response.OK(c, "Subscription retrieved successfully", subscription.ToUserResponse())
}

// ============= User Subscription Purchase =============

// PurchaseSubscriptionRequest represents the request to purchase a subscription
type PurchaseSubscriptionRequest struct {
	PlanID          uint   `json:"plan_id" binding:"required" example:"1"`
	PaymentMethod   string `json:"payment_method" binding:"required" example:"credit_card"`
	PaymentGateway  string `json:"payment_gateway" binding:"required" example:"stripe"`
	CouponCode      string `json:"coupon_code,omitempty" example:"SAVE10"`
	AutoRenew       *bool  `json:"auto_renew,omitempty" example:"true"`
	PaymentMetadata string `json:"payment_metadata,omitempty" example:"{\"card_last4\": \"1234\"}"`
}

// PurchaseSubscription godoc
// @Summary [User] Purchase subscription
// @Description Create subscription order with payment - returns payment URL for completion
// @Tags User-Subscription
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param purchase_request body PurchaseSubscriptionRequest true "Purchase request data"
// @Success 201 {object} response.StandardResponse{data=service.CreateSubscriptionOrderResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /user/subscriptions/purchase [post]
func (h *UserSubscriptionPublicHandler) PurchaseSubscription(c *gin.Context) {
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

	// Bind request
	var req PurchaseSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request data", err.Error())
		return
	}

	// Check if subscription plan exists and is available
	plan, err := h.subscriptionPlanService.GetSubscriptionPlan(c.Request.Context(), req.PlanID)
	if err != nil {
		if err.Error() == "subscription plan not found" {
			response.BadRequest(c, "Invalid subscription plan", "Selected plan not found")
			return
		}
		logger.Error("Failed to get subscription plan", logger.Error2("error", err), logger.Uint("plan_id", req.PlanID))
		response.InternalServerError(c, "Failed to validate subscription plan", err.Error())
		return
	}

	// Check if plan is available for purchase
	if plan.Status != model.SubscriptionPlanStatusActive || !plan.IsVisible {
		response.BadRequest(c, "Subscription plan not available", "Selected plan is not available for purchase")
		return
	}

	// Create subscription order request
	orderReq := &service.CreateSubscriptionOrderRequest{
		UserID:             user.ID,
		SubscriptionPlanID: req.PlanID,
		OrderType:          model.OrderTypeNew, // New subscription purchase
		CouponCode:         req.CouponCode,
		PaymentGateway:     req.PaymentGateway,
		PaymentMethod:      req.PaymentMethod,
		ReturnURL:          c.Request.Header.Get("Referer"), // Use referer as return URL
		Metadata:           req.PaymentMetadata,
	}

	// Create subscription order with payment
	orderResponse, err := h.subscriptionOrderService.CreateSubscriptionOrder(c.Request.Context(), orderReq)
	if err != nil {
		logger.Error("Failed to create subscription order", logger.Error2("error", err), logger.Uint("user_id", user.ID), logger.Uint("plan_id", req.PlanID))
		response.InternalServerError(c, "Failed to create subscription order", err.Error())
		return
	}

	response.CreatedWithMessage(c, "Subscription order created successfully. Complete payment to activate subscription.", orderResponse)
}

// CancelSubscriptionRequest represents the request to cancel a subscription
type CancelSubscriptionRequest struct {
	Reason      string `json:"reason" binding:"required,max=255" example:"No longer needed"`
	Immediately bool   `json:"immediately,omitempty" example:"false"`
}

// CancelMySubscription godoc
// @Summary [User] Cancel my subscription
// @Description Cancel current user's subscription
// @Tags User-Subscription
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User subscription ID"
// @Param cancel_request body CancelSubscriptionRequest true "Cancellation data"
// @Success 200 {object} response.StandardResponse{data=model.UserSubscriptionResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /user/subscriptions/{id}/cancel [post]
func (h *UserSubscriptionPublicHandler) CancelMySubscription(c *gin.Context) {
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

	// Parse subscription ID
	subscriptionIDStr := c.Param("id")
	subscriptionID, err := strconv.ParseUint(subscriptionIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid subscription ID", "Subscription ID must be a valid number")
		return
	}

	// Get existing subscription to check ownership
	existingSubscription, err := h.userSubscriptionService.GetUserSubscription(c.Request.Context(), uint(subscriptionID))
	if err != nil {
		if err.Error() == "subscription not found" {
			response.NotFound(c, "Subscription not found")
			return
		}
		logger.Error("Failed to get user subscription", logger.Error2("error", err), logger.Uint("subscription_id", uint(subscriptionID)))
		response.InternalServerError(c, "Failed to get user subscription", err.Error())
		return
	}

	// Check if user owns this subscription
	if existingSubscription.UserID != user.ID {
		response.Forbidden(c, "You can only cancel your own subscriptions")
		return
	}

	// Bind cancellation request
	var cancelReq CancelSubscriptionRequest
	if err := c.ShouldBindJSON(&cancelReq); err != nil {
		response.BadRequest(c, "Invalid request data", err.Error())
		return
	}

	// Cancel user subscription
	subscription, err := h.userSubscriptionService.CancelUserSubscription(c.Request.Context(), uint(subscriptionID), cancelReq.Reason, cancelReq.Immediately)
	if err != nil {
		logger.Error("Failed to cancel user subscription", logger.Error2("error", err), logger.Uint("subscription_id", uint(subscriptionID)), logger.Uint("user_id", user.ID))
		response.InternalServerError(c, "Failed to cancel subscription", err.Error())
		return
	}

	response.OK(c, "Subscription cancelled successfully", subscription.ToUserResponse())
}

// ============= User Subscription History =============

// GetMySubscriptionHistory godoc
// @Summary [User] Get my subscription history
// @Description Get current user's subscription history including past subscriptions
// @Tags User-Subscription
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Limit results" minimum(1) maximum(100) example(10)
// @Param offset query int false "Offset results" minimum(0) example(0)
// @Success 200 {object} response.PaginatedResponse{data=[]model.UserSubscriptionResponse}
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /user/subscriptions/history [get]
func (h *UserSubscriptionPublicHandler) GetMySubscriptionHistory(c *gin.Context) {
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

	// Bind query parameters
	var req service.GetUserSubscriptionsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "Invalid query parameters", err.Error())
		return
	}

	// Set user ID to current user
	req.UserID = user.ID

	// Get user subscription history
	subscriptions, totalCount, err := h.userSubscriptionService.GetUserSubscriptions(c.Request.Context(), &req)
	if err != nil {
		logger.Error("Failed to get user subscription history", logger.Error2("error", err), logger.Uint("user_id", user.ID))
		response.InternalServerError(c, "Failed to get subscription history", err.Error())
		return
	}

	// Convert to user response format
	var subscriptionResponses []*model.UserSubscriptionResponse
	for _, subscription := range subscriptions {
		subscriptionResponses = append(subscriptionResponses, subscription.ToUserResponse())
	}

	response.OKPaginated(c, "Subscription history retrieved successfully", subscriptionResponses, totalCount, req.Limit, req.Offset)
}

// ============= User Subscription Upgrade/Downgrade =============

// UpgradeDowngradeSubscription godoc
// @Summary [User] Upgrade/Downgrade subscription
// @Description Upgrade or downgrade current user's subscription to a different plan
// @Tags User-Subscription
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param upgrade_request body service.UpgradeDowngradeRequest true "Upgrade/downgrade request data"
// @Success 201 {object} response.StandardResponse{data=service.CreateSubscriptionOrderResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /user/subscriptions/upgrade-downgrade [post]
func (h *UserSubscriptionPublicHandler) UpgradeDowngradeSubscription(c *gin.Context) {
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

	// Bind request
	var req service.UpgradeDowngradeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request data", err.Error())
		return
	}

	// Verify user owns the current subscription
	currentSubscription, err := h.userSubscriptionService.GetUserSubscription(c.Request.Context(), req.CurrentSubscriptionID)
	if err != nil {
		if err.Error() == "subscription not found" {
			response.NotFound(c, "Current subscription not found")
			return
		}
		logger.Error("Failed to get current subscription", logger.Error2("error", err), logger.Uint("subscription_id", req.CurrentSubscriptionID))
		response.InternalServerError(c, "Failed to get current subscription", err.Error())
		return
	}

	if currentSubscription.UserID != user.ID {
		response.Forbidden(c, "You can only upgrade/downgrade your own subscriptions")
		return
	}

	// Create upgrade/downgrade order
	orderResponse, err := h.subscriptionOrderService.CreateUpgradeDowngradeOrder(c.Request.Context(), user.ID, &req)
	if err != nil {
		logger.Error("Failed to create upgrade/downgrade order", 
			logger.Error2("error", err), 
			logger.Uint("user_id", user.ID),
			logger.Uint("current_subscription_id", req.CurrentSubscriptionID),
			logger.Uint("new_plan_id", req.NewPlanID))
		response.InternalServerError(c, "Failed to create upgrade/downgrade order", err.Error())
		return
	}

	response.CreatedWithMessage(c, "Upgrade/downgrade order created successfully", orderResponse)
}

// GetUpgradeDowngradeOptions godoc
// @Summary [User] Get upgrade/downgrade options
// @Description Get available upgrade/downgrade options for a specific subscription
// @Tags User-Subscription
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User subscription ID"
// @Success 200 {object} response.StandardResponse{data=[]model.SubscriptionPlanResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /user/subscriptions/{id}/upgrade-downgrade-options [get]
func (h *UserSubscriptionPublicHandler) GetUpgradeDowngradeOptions(c *gin.Context) {
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

	// Parse subscription ID
	subscriptionIDStr := c.Param("id")
	subscriptionID, err := strconv.ParseUint(subscriptionIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid subscription ID", "Subscription ID must be a valid number")
		return
	}

	// Get current subscription with plan details
	subscription, err := h.userSubscriptionService.GetUserSubscriptionWithRelations(c.Request.Context(), uint(subscriptionID))
	if err != nil {
		if err.Error() == "subscription not found" {
			response.NotFound(c, "Subscription not found")
			return
		}
		logger.Error("Failed to get subscription", logger.Error2("error", err), logger.Uint("subscription_id", uint(subscriptionID)))
		response.InternalServerError(c, "Failed to get subscription", err.Error())
		return
	}

	// Check if user owns this subscription
	if subscription.UserID != user.ID {
		response.Forbidden(c, "You can only access your own subscriptions")
		return
	}

	// Check if subscription is active
	if !subscription.IsActive() {
		response.BadRequest(c, "Subscription is not active", "Only active subscriptions can be upgraded/downgraded")
		return
	}

	// Get all available plans for the same currency
	currency := subscription.Currency
	if subscription.SubscriptionPlan != nil {
		currency = subscription.SubscriptionPlan.Currency
	}

	availablePlans, err := h.subscriptionPlanService.GetPublicSubscriptionPlans(c.Request.Context(), currency)
	if err != nil {
		logger.Error("Failed to get available plans", logger.Error2("error", err))
		response.InternalServerError(c, "Failed to get available plans", err.Error())
		return
	}

	// Filter out current plan and prepare response
	var upgradeDowngradeOptions []*model.SubscriptionPlanResponse
	for _, plan := range availablePlans {
		// Skip current plan
		if plan.ID == subscription.SubscriptionPlanID {
			continue
		}

		// Skip plans with same price (no meaningful upgrade/downgrade)
		if plan.Price == subscription.Price {
			continue
		}

		planResponse := plan.ToPublicResponse()
		
		// Add upgrade/downgrade indicator
		if plan.Price > subscription.Price {
			planResponse.Description = fmt.Sprintf("[UPGRADE] %s", planResponse.Description)
		} else {
			planResponse.Description = fmt.Sprintf("[DOWNGRADE] %s", planResponse.Description)
		}

		upgradeDowngradeOptions = append(upgradeDowngradeOptions, planResponse)
	}

	response.OK(c, "Upgrade/downgrade options retrieved successfully", upgradeDowngradeOptions)
}

// ============= Coupon Validation =============

// ValidateCouponRequest represents the request to validate a coupon
type ValidateCouponRequest struct {
	CouponCode  string  `json:"coupon_code" binding:"required" example:"SAVE20"`
	PlanID      uint    `json:"plan_id" binding:"required" example:"1"`
	OrderAmount float64 `json:"order_amount" binding:"required,min=0" example:"29.99"`
}

// ValidateCoupon godoc
// @Summary [User] Validate coupon code
// @Description Validate a coupon code for a specific plan and order amount
// @Tags User-Subscription
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param validate_request body ValidateCouponRequest true "Coupon validation data"
// @Success 200 {object} response.StandardResponse{data=service.ValidateCouponResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /user/coupons/validate [post]
func (h *UserSubscriptionPublicHandler) ValidateCoupon(c *gin.Context) {
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

	// Bind request
	var req ValidateCouponRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request data", err.Error())
		return
	}

	// Get subscription plan to determine currency
	plan, err := h.subscriptionPlanService.GetSubscriptionPlan(c.Request.Context(), req.PlanID)
	if err != nil {
		if err.Error() == "subscription plan not found" {
			response.BadRequest(c, "Invalid subscription plan", "Selected plan not found")
			return
		}
		logger.Error("Failed to get subscription plan", logger.Error2("error", err), logger.Uint("plan_id", req.PlanID))
		response.InternalServerError(c, "Failed to validate subscription plan", err.Error())
		return
	}

	// Create coupon validation request
	validateReq := &service.ValidateCouponRequest{
		Code:        req.CouponCode,
		UserID:      uint64(user.ID),
		OrderAmount: req.OrderAmount,
		PlanID:      uint64(req.PlanID),
		Currency:    plan.Currency,
	}

	// Validate coupon
	validateResp, err := h.couponService.ValidateCoupon(c.Request.Context(), validateReq)
	if err != nil {
		logger.Error("Failed to validate coupon", logger.Error2("error", err), logger.String("coupon_code", req.CouponCode))
		response.InternalServerError(c, "Failed to validate coupon", err.Error())
		return
	}

	response.OK(c, "Coupon validation completed", validateResp)
}

// ============= User Subscription Pause/Resume =============

// PauseSubscriptionRequest represents the request to pause a subscription
type PauseSubscriptionRequest struct {
	Reason string `json:"reason" binding:"required,max=255" example:"Temporary financial difficulty"`
}

// PauseMySubscription godoc
// @Summary [User] Pause my subscription
// @Description Pause current user's active subscription
// @Tags User-Subscription
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User subscription ID"
// @Param pause_request body PauseSubscriptionRequest true "Pause request data"
// @Success 200 {object} response.StandardResponse{data=model.UserSubscriptionResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /user/subscriptions/{id}/pause [post]
func (h *UserSubscriptionPublicHandler) PauseMySubscription(c *gin.Context) {
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

	// Parse subscription ID
	subscriptionIDStr := c.Param("id")
	subscriptionID, err := strconv.ParseUint(subscriptionIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid subscription ID", "Subscription ID must be a valid number")
		return
	}

	// Get existing subscription to check ownership
	existingSubscription, err := h.userSubscriptionService.GetUserSubscription(c.Request.Context(), uint(subscriptionID))
	if err != nil {
		if err.Error() == "subscription not found" {
			response.NotFound(c, "Subscription not found")
			return
		}
		logger.Error("Failed to get user subscription", logger.Error2("error", err), logger.Uint("subscription_id", uint(subscriptionID)))
		response.InternalServerError(c, "Failed to get user subscription", err.Error())
		return
	}

	// Check if user owns this subscription
	if existingSubscription.UserID != user.ID {
		response.Forbidden(c, "You can only pause your own subscriptions")
		return
	}

	// Check if subscription can be paused
	if existingSubscription.Status != model.UserSubscriptionStatusActive {
		response.BadRequest(c, "Subscription cannot be paused", "Only active subscriptions can be paused")
		return
	}

	// Bind pause request
	var pauseReq PauseSubscriptionRequest
	if err := c.ShouldBindJSON(&pauseReq); err != nil {
		response.BadRequest(c, "Invalid request data", err.Error())
		return
	}

	// Pause subscription
	subscription, err := h.userSubscriptionService.PauseUserSubscription(c.Request.Context(), uint(subscriptionID), pauseReq.Reason, user.ID)
	if err != nil {
		logger.Error("Failed to pause user subscription", logger.Error2("error", err), logger.Uint("subscription_id", uint(subscriptionID)), logger.Uint("user_id", user.ID))
		response.InternalServerError(c, "Failed to pause subscription", err.Error())
		return
	}

	response.OK(c, "Subscription paused successfully", subscription.ToUserResponse())
}

// ResumeMySubscription godoc
// @Summary [User] Resume my subscription
// @Description Resume current user's paused subscription
// @Tags User-Subscription
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
// @Router /user/subscriptions/{id}/resume [post]
func (h *UserSubscriptionPublicHandler) ResumeMySubscription(c *gin.Context) {
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

	// Parse subscription ID
	subscriptionIDStr := c.Param("id")
	subscriptionID, err := strconv.ParseUint(subscriptionIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid subscription ID", "Subscription ID must be a valid number")
		return
	}

	// Get existing subscription to check ownership
	existingSubscription, err := h.userSubscriptionService.GetUserSubscription(c.Request.Context(), uint(subscriptionID))
	if err != nil {
		if err.Error() == "subscription not found" {
			response.NotFound(c, "Subscription not found")
			return
		}
		logger.Error("Failed to get user subscription", logger.Error2("error", err), logger.Uint("subscription_id", uint(subscriptionID)))
		response.InternalServerError(c, "Failed to get user subscription", err.Error())
		return
	}

	// Check if user owns this subscription
	if existingSubscription.UserID != user.ID {
		response.Forbidden(c, "You can only resume your own subscriptions")
		return
	}

	// Check if subscription can be resumed
	if existingSubscription.Status != model.UserSubscriptionStatusPaused {
		response.BadRequest(c, "Subscription cannot be resumed", "Only paused subscriptions can be resumed")
		return
	}

	// Resume subscription
	subscription, err := h.userSubscriptionService.ResumeUserSubscription(c.Request.Context(), uint(subscriptionID), user.ID)
	if err != nil {
		logger.Error("Failed to resume user subscription", logger.Error2("error", err), logger.Uint("subscription_id", uint(subscriptionID)), logger.Uint("user_id", user.ID))
		response.InternalServerError(c, "Failed to resume subscription", err.Error())
		return
	}

	response.OK(c, "Subscription resumed successfully", subscription.ToUserResponse())
}

// ============= Auto-Renewal Management =============

// UpdateAutoRenewalRequest represents the request to update auto-renewal settings
type UpdateAutoRenewalRequest struct {
	AutoRenew bool `json:"auto_renew" example:"true"`
}

// UpdateAutoRenewal godoc
// @Summary [User] Update auto-renewal settings
// @Description Enable or disable auto-renewal for a subscription
// @Tags User-Subscription
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User subscription ID"
// @Param update_request body UpdateAutoRenewalRequest true "Auto-renewal update data"
// @Success 200 {object} response.StandardResponse{data=model.UserSubscriptionResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /user/subscriptions/{id}/auto-renewal [put]
func (h *UserSubscriptionPublicHandler) UpdateAutoRenewal(c *gin.Context) {
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

	// Parse subscription ID
	subscriptionIDStr := c.Param("id")
	subscriptionID, err := strconv.ParseUint(subscriptionIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid subscription ID", "Subscription ID must be a valid number")
		return
	}

	// Get existing subscription to check ownership
	existingSubscription, err := h.userSubscriptionService.GetUserSubscription(c.Request.Context(), uint(subscriptionID))
	if err != nil {
		if err.Error() == "subscription not found" {
			response.NotFound(c, "Subscription not found")
			return
		}
		logger.Error("Failed to get user subscription", logger.Error2("error", err), logger.Uint("subscription_id", uint(subscriptionID)))
		response.InternalServerError(c, "Failed to get user subscription", err.Error())
		return
	}

	// Check if user owns this subscription
	if existingSubscription.UserID != user.ID {
		response.Forbidden(c, "You can only modify your own subscriptions")
		return
	}

	// Bind update request
	var updateReq UpdateAutoRenewalRequest
	if err := c.ShouldBindJSON(&updateReq); err != nil {
		response.BadRequest(c, "Invalid request data", err.Error())
		return
	}

	// Update auto-renewal setting
	var subscription *model.UserSubscription
	if updateReq.AutoRenew {
		err = h.userSubscriptionService.EnableAutoRenewal(c.Request.Context(), uint(subscriptionID))
		if err == nil {
			subscription, err = h.userSubscriptionService.GetUserSubscription(c.Request.Context(), uint(subscriptionID))
		}
	} else {
		err = h.userSubscriptionService.DisableAutoRenewal(c.Request.Context(), uint(subscriptionID))
		if err == nil {
			subscription, err = h.userSubscriptionService.GetUserSubscription(c.Request.Context(), uint(subscriptionID))
		}
	}

	if err != nil {
		logger.Error("Failed to update auto-renewal", logger.Error2("error", err), logger.Uint("subscription_id", uint(subscriptionID)), logger.Uint("user_id", user.ID))
		response.InternalServerError(c, "Failed to update auto-renewal setting", err.Error())
		return
	}

	action := "disabled"
	if updateReq.AutoRenew {
		action = "enabled"
	}
	response.OK(c, fmt.Sprintf("Auto-renewal %s successfully", action), subscription.ToUserResponse())
}

// ============= Subscription Statistics and Analytics =============

// SubscriptionStatsResponse represents subscription statistics
type SubscriptionStatsResponse struct {
	TotalSubscriptions    int                           `json:"total_subscriptions" example:"5"`
	ActiveSubscriptions   int                           `json:"active_subscriptions" example:"3"`
	PausedSubscriptions   int                           `json:"paused_subscriptions" example:"1"`
	CancelledSubscriptions int                          `json:"cancelled_subscriptions" example:"1"`
	ExpiredSubscriptions  int                           `json:"expired_subscriptions" example:"0"`
	TotalSpent           float64                       `json:"total_spent" example:"149.95"`
	CurrentMonthlyCost   float64                       `json:"current_monthly_cost" example:"29.99"`
	NextBillingDate      *time.Time                    `json:"next_billing_date,omitempty" example:"2024-02-01T00:00:00Z"`
	SubscriptionsByPlan  map[string]int                `json:"subscriptions_by_plan"`
	UsageStats          *SubscriptionUsageStats       `json:"usage_stats,omitempty"`
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

// GetMySubscriptionStats godoc
// @Summary [User] Get my subscription statistics
// @Description Get comprehensive statistics about current user's subscriptions
// @Tags User-Subscription
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.StandardResponse{data=SubscriptionStatsResponse}
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /user/subscriptions/stats [get]
func (h *UserSubscriptionPublicHandler) GetMySubscriptionStats(c *gin.Context) {
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

	// Get user subscription statistics
	stats, err := h.userSubscriptionService.GetUserSubscriptionStats(c.Request.Context(), user.ID)
	if err != nil {
		logger.Error("Failed to get user subscription stats", logger.Error2("error", err), logger.Uint("user_id", user.ID))
		response.InternalServerError(c, "Failed to get subscription statistics", err.Error())
		return
	}

	response.OK(c, "Subscription statistics retrieved successfully", stats)
}


// ResetTrafficUsage godoc
// @Summary [User] Reset traffic usage
// @Description Reset traffic usage for current billing cycle (admin only or with user consent)
// @Tags User-Subscription
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
// @Router /user/subscriptions/{id}/reset-traffic [post]
func (h *UserSubscriptionPublicHandler) ResetTrafficUsage(c *gin.Context) {
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

	// Parse subscription ID
	subscriptionIDStr := c.Param("id")
	subscriptionID, err := strconv.ParseUint(subscriptionIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid subscription ID", "Subscription ID must be a valid number")
		return
	}

	// Get existing subscription to check ownership
	existingSubscription, err := h.userSubscriptionService.GetUserSubscription(c.Request.Context(), uint(subscriptionID))
	if err != nil {
		if err.Error() == "subscription not found" {
			response.NotFound(c, "Subscription not found")
			return
		}
		logger.Error("Failed to get user subscription", logger.Error2("error", err), logger.Uint("subscription_id", uint(subscriptionID)))
		response.InternalServerError(c, "Failed to get user subscription", err.Error())
		return
	}

	// Check if user owns this subscription
	if existingSubscription.UserID != user.ID {
		response.Forbidden(c, "You can only reset traffic for your own subscriptions")
		return
	}

	// Reset traffic usage
	subscription, err := h.userSubscriptionService.ResetTrafficUsage(c.Request.Context(), uint(subscriptionID), user.ID)
	if err != nil {
		logger.Error("Failed to reset traffic usage", logger.Error2("error", err), logger.Uint("subscription_id", uint(subscriptionID)), logger.Uint("user_id", user.ID))
		response.InternalServerError(c, "Failed to reset traffic usage", err.Error())
		return
	}

	response.OK(c, "Traffic usage reset successfully", subscription.ToUserResponse())
}


// ============= Subscription Notifications and Reminders =============

// NotificationPreferencesRequest represents notification preferences
type NotificationPreferencesRequest struct {
	EmailNotifications bool `json:"email_notifications" example:"true"`
	SMSNotifications   bool `json:"sms_notifications" example:"false"`
	PushNotifications  bool `json:"push_notifications" example:"true"`
	// Notification types
	RenewalReminders    bool `json:"renewal_reminders" example:"true"`
	ExpirationWarnings  bool `json:"expiration_warnings" example:"true"`
	UsageAlerts        bool `json:"usage_alerts" example:"true"`
	PromotionalOffers  bool `json:"promotional_offers" example:"false"`
}

// UpdateNotificationPreferences godoc
// @Summary [User] Update notification preferences
// @Description Update notification preferences for subscription-related communications
// @Tags User-Subscription
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param preferences_request body NotificationPreferencesRequest true "Notification preferences"
// @Success 200 {object} response.StandardResponse{data=NotificationPreferencesRequest}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /user/subscriptions/notification-preferences [put]
func (h *UserSubscriptionPublicHandler) UpdateNotificationPreferences(c *gin.Context) {
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

	// Bind preferences request
	var prefsReq NotificationPreferencesRequest
	if err := c.ShouldBindJSON(&prefsReq); err != nil {
		response.BadRequest(c, "Invalid request data", err.Error())
		return
	}

	// Update notification preferences
	err := h.userSubscriptionService.UpdateNotificationPreferences(c.Request.Context(), user.ID, &service.NotificationPreferences{
		EmailNotifications:  prefsReq.EmailNotifications,
		SMSNotifications:    prefsReq.SMSNotifications,
		PushNotifications:   prefsReq.PushNotifications,
		RenewalReminders:    prefsReq.RenewalReminders,
		ExpirationWarnings:  prefsReq.ExpirationWarnings,
		UsageAlerts:         prefsReq.UsageAlerts,
		PromotionalOffers:   prefsReq.PromotionalOffers,
	})

	if err != nil {
		logger.Error("Failed to update notification preferences", logger.Error2("error", err), logger.Uint("user_id", user.ID))
		response.InternalServerError(c, "Failed to update notification preferences", err.Error())
		return
	}

	response.OK(c, "Notification preferences updated successfully", prefsReq)
}

// GetNotificationPreferences godoc
// @Summary [User] Get notification preferences
// @Description Get current notification preferences for subscription communications
// @Tags User-Subscription
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.StandardResponse{data=NotificationPreferencesRequest}
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /user/subscriptions/notification-preferences [get]
func (h *UserSubscriptionPublicHandler) GetNotificationPreferences(c *gin.Context) {
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

	// Get notification preferences
	prefs, err := h.userSubscriptionService.GetNotificationPreferences(c.Request.Context(), user.ID)
	if err != nil {
		logger.Error("Failed to get notification preferences", logger.Error2("error", err), logger.Uint("user_id", user.ID))
		response.InternalServerError(c, "Failed to get notification preferences", err.Error())
		return
	}

	// Convert to response format
	prefsResponse := &NotificationPreferencesRequest{
		EmailNotifications:  prefs.EmailNotifications,
		SMSNotifications:    prefs.SMSNotifications,
		PushNotifications:   prefs.PushNotifications,
		RenewalReminders:    prefs.RenewalReminders,
		ExpirationWarnings:  prefs.ExpirationWarnings,
		UsageAlerts:         prefs.UsageAlerts,
		PromotionalOffers:   prefs.PromotionalOffers,
	}

	response.OK(c, "Notification preferences retrieved successfully", prefsResponse)
}

// ============= Server Group Management =============

// GetAvailableServerGroups godoc
// @Summary [User] Get available server groups
// @Description Get server groups available for a specific subscription
// @Tags User-Subscription
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param subscription_id path int true "User subscription ID"
// @Success 200 {object} response.StandardResponse{data=[]model.ServerGroupResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /user/subscriptions/{subscription_id}/available-server-groups [get]
func (h *UserSubscriptionPublicHandler) GetAvailableServerGroups(c *gin.Context) {
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

	// Parse subscription ID
	subscriptionIDStr := c.Param("subscription_id")
	subscriptionID, err := strconv.ParseUint(subscriptionIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid subscription ID", "Subscription ID must be a valid number")
		return
	}

	// Get available server groups
	availableGroups, err := h.userSubscriptionService.GetAvailableServerGroups(c.Request.Context(), user.ID, uint(subscriptionID))
	if err != nil {
		if err.Error() == "subscription does not belong to user" {
			response.Forbidden(c, "You can only access your own subscriptions")
			return
		}
		logger.Error("Failed to get available server groups", logger.Error2("error", err), logger.Uint("user_id", user.ID), logger.Uint("subscription_id", uint(subscriptionID)))
		response.InternalServerError(c, "Failed to get available server groups", err.Error())
		return
	}

	// Convert to response format
	var groupResponses []*model.ServerGroupResponse
	for _, group := range availableGroups {
		groupResponses = append(groupResponses, group.ToResponse())
	}

	response.OK(c, "Available server groups retrieved successfully", groupResponses)
}

// GetSubscriptionServerGroups godoc
// @Summary [User] Get subscription server groups
// @Description Get server groups assigned to a specific subscription
// @Tags User-Subscription
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param subscription_id path int true "User subscription ID"
// @Success 200 {object} response.StandardResponse{data=[]model.ServerGroupResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /user/subscriptions/{subscription_id}/server-groups [get]
func (h *UserSubscriptionPublicHandler) GetSubscriptionServerGroups(c *gin.Context) {
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

	// Parse subscription ID
	subscriptionIDStr := c.Param("subscription_id")
	subscriptionID, err := strconv.ParseUint(subscriptionIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid subscription ID", "Subscription ID must be a valid number")
		return
	}

	// Get assigned server groups
	assignedGroups, err := h.userSubscriptionService.GetSubscriptionServerGroups(c.Request.Context(), user.ID, uint(subscriptionID))
	if err != nil {
		if err.Error() == "subscription does not belong to user" {
			response.Forbidden(c, "You can only access your own subscriptions")
			return
		}
		logger.Error("Failed to get subscription server groups", logger.Error2("error", err), logger.Uint("user_id", user.ID), logger.Uint("subscription_id", uint(subscriptionID)))
		response.InternalServerError(c, "Failed to get subscription server groups", err.Error())
		return
	}

	// Convert to response format
	var groupResponses []*model.ServerGroupResponse
	for _, group := range assignedGroups {
		groupResponses = append(groupResponses, group.ToResponse())
	}

	response.OK(c, "Subscription server groups retrieved successfully", groupResponses)
}

// UpdateSubscriptionServerGroupsRequest represents the request to update server groups
type UpdateSubscriptionServerGroupsRequest struct {
	ServerGroupIDs []uint `json:"server_group_ids" binding:"required" example:"1,2,3"`
}

// UpdateSubscriptionServerGroups godoc
// @Summary [User] Update subscription server groups
// @Description Update server groups assigned to a subscription
// @Tags User-Subscription
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param subscription_id path int true "User subscription ID"
// @Param update_request body UpdateSubscriptionServerGroupsRequest true "Server groups update data"
// @Success 200 {object} response.StandardResponse{data=model.UserSubscriptionResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /user/subscriptions/{subscription_id}/server-groups [put]
func (h *UserSubscriptionPublicHandler) UpdateSubscriptionServerGroups(c *gin.Context) {
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

	// Parse subscription ID
	subscriptionIDStr := c.Param("subscription_id")
	subscriptionID, err := strconv.ParseUint(subscriptionIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid subscription ID", "Subscription ID must be a valid number")
		return
	}

	// Bind update request
	var updateReq UpdateSubscriptionServerGroupsRequest
	if err := c.ShouldBindJSON(&updateReq); err != nil {
		response.BadRequest(c, "Invalid request data", err.Error())
		return
	}

	// Create service request
	serviceReq := &service.UpdateSubscriptionServerGroupsRequest{
		SubscriptionID: uint(subscriptionID),
		ServerGroupIDs: updateReq.ServerGroupIDs,
	}

	// Update server groups
	subscription, err := h.userSubscriptionService.UpdateSubscriptionServerGroups(c.Request.Context(), user.ID, serviceReq)
	if err != nil {
		if err.Error() == "subscription does not belong to user" {
			response.Forbidden(c, "You can only modify your own subscriptions")
			return
		}
		logger.Error("Failed to update subscription server groups", logger.Error2("error", err), logger.Uint("user_id", user.ID), logger.Uint("subscription_id", uint(subscriptionID)))
		response.InternalServerError(c, "Failed to update server groups", err.Error())
		return
	}

	response.OK(c, "Server groups updated successfully", subscription.ToUserResponse())
}

// GetAccessibleServers godoc
// @Summary [User] Get accessible servers
// @Description Get all shadowsocks servers accessible through user's active subscriptions
// @Tags User-Subscription
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.StandardResponse{data=[]model.ShadowsocksServerResponse}
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /user/subscriptions/accessible-servers [get]
func (h *UserSubscriptionPublicHandler) GetAccessibleServers(c *gin.Context) {
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

	// Get accessible servers
	accessibleServers, err := h.userSubscriptionService.GetUserAccessibleServers(c.Request.Context(), user.ID)
	if err != nil {
		logger.Error("Failed to get accessible servers", logger.Error2("error", err), logger.Uint("user_id", user.ID))
		response.InternalServerError(c, "Failed to get accessible servers", err.Error())
		return
	}

	// Convert to response format
	var serverResponses []*model.ShadowsocksServerResponse
	for _, server := range accessibleServers {
		serverResponses = append(serverResponses, server.ToResponse())
	}

	response.OK(c, "Accessible servers retrieved successfully", serverResponses)
}

// GetServersBySubscription godoc
// @Summary [User] Get servers by subscription
// @Description Get shadowsocks servers accessible through a specific subscription
// @Tags User-Subscription
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param subscription_id path int true "User subscription ID"
// @Success 200 {object} response.StandardResponse{data=[]model.ShadowsocksServerResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /user/subscriptions/{subscription_id}/servers [get]
func (h *UserSubscriptionPublicHandler) GetServersBySubscription(c *gin.Context) {
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

	// Parse subscription ID
	subscriptionIDStr := c.Param("subscription_id")
	subscriptionID, err := strconv.ParseUint(subscriptionIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid subscription ID", "Subscription ID must be a valid number")
		return
	}

	// Get servers for this subscription
	servers, err := h.userSubscriptionService.GetUserServersBySubscription(c.Request.Context(), user.ID, uint(subscriptionID))
	if err != nil {
		if err.Error() == "subscription does not belong to user" {
			response.Forbidden(c, "You can only access your own subscriptions")
			return
		}
		logger.Error("Failed to get servers by subscription", logger.Error2("error", err), logger.Uint("user_id", user.ID), logger.Uint("subscription_id", uint(subscriptionID)))
		response.InternalServerError(c, "Failed to get servers", err.Error())
		return
	}

	// Convert to response format
	var serverResponses []*model.ShadowsocksServerResponse
	for _, server := range servers {
		serverResponses = append(serverResponses, server.ToResponse())
	}

	response.OK(c, "Subscription servers retrieved successfully", serverResponses)
}