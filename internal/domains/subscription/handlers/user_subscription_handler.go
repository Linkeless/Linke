package handlers

import (
	"net/http"
	"strconv"

	"linke/internal/domains/subscription/entities"
	"linke/internal/domains/subscription/usecases/interfaces"
	userEntities "linke/internal/domains/user/entities"
	"linke/internal/shared/logger"
	"linke/internal/shared/middleware"
	"linke/internal/shared/response"

	"github.com/gin-gonic/gin"
)

type UserSubscriptionHandler struct {
	userSubscriptionService interfaces.UserSubscriptionService
}

func NewUserSubscriptionHandler(userSubscriptionService interfaces.UserSubscriptionService) *UserSubscriptionHandler {
	return &UserSubscriptionHandler{
		userSubscriptionService: userSubscriptionService,
	}
}

// GetMySubscriptions godoc
// @Summary [User] Get my subscriptions
// @Description Get current user's subscriptions with optional filtering
// @Tags User-Subscription
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param status query string false "Filter by status" Enums(active, paused, cancelled, expired, trial)
// @Param limit query int false "Limit results" minimum(1) maximum(100) example(10)
// @Param offset query int false "Offset results" minimum(0) example(0)
// @Success 200 {object} response.PaginatedResponse{data=[]entities.UserSubscriptionResponse}
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /subscriptions/my [get]
func (h *UserSubscriptionHandler) GetMySubscriptions(c *gin.Context) {
	// Get current user from context
	userValue, exists := c.Get(middleware.AuthContextKey)
	if !exists {
		response.Unauthorized(c, "Authentication required")
		return
	}

	user, ok := userValue.(*userEntities.User)
	if !ok {
		response.Unauthorized(c, "Invalid user context")
		return
	}

	// Parse query parameters
	var req interfaces.GetUserSubscriptionsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "Invalid query parameters", err.Error())
		return
	}

	// Force user ID to current user
	req.UserID = user.ID

	// Set defaults
	if req.Limit == 0 {
		req.Limit = 10
	}

	// Get user subscriptions
	subscriptions, totalCount, err := h.userSubscriptionService.GetUserSubscriptions(c.Request.Context(), &req)
	if err != nil {
		logger.Error("Failed to get user subscriptions", logger.Error2("error", err), logger.Uint("user_id", user.ID))
		response.InternalServerError(c, "Failed to get subscriptions", err.Error())
		return
	}

	// Convert to response format
	var subscriptionResponses []*entities.UserSubscriptionResponse
	for _, sub := range subscriptions {
		subscriptionResponses = append(subscriptionResponses, sub.ToResponse())
	}

	response.OKPaginated(c, "My subscriptions retrieved successfully", subscriptionResponses, totalCount, req.Limit, req.Offset)
}

// GetMyActiveSubscriptions godoc
// @Summary [User] Get my active subscriptions
// @Description Get current user's active subscriptions
// @Tags User-Subscription
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.StandardResponse{data=[]entities.UserSubscriptionResponse}
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /subscriptions/my/active [get]
func (h *UserSubscriptionHandler) GetMyActiveSubscriptions(c *gin.Context) {
	// Get current user from context
	userValue, exists := c.Get(middleware.AuthContextKey)
	if !exists {
		response.Unauthorized(c, "Authentication required")
		return
	}

	user, ok := userValue.(*userEntities.User)
	if !ok {
		response.Unauthorized(c, "Invalid user context")
		return
	}

	// Get active subscriptions
	subscriptions, err := h.userSubscriptionService.GetUserActiveSubscriptions(c.Request.Context(), user.ID)
	if err != nil {
		logger.Error("Failed to get active subscriptions", logger.Error2("error", err), logger.Uint("user_id", user.ID))
		response.InternalServerError(c, "Failed to get active subscriptions", err.Error())
		return
	}

	// Convert to response format
	var subscriptionResponses []*entities.UserSubscriptionResponse
	for _, sub := range subscriptions {
		subscriptionResponses = append(subscriptionResponses, sub.ToResponse())
	}

	response.OK(c, "Active subscriptions retrieved successfully", subscriptionResponses)
}

// GetSubscription godoc
// @Summary [User] Get subscription details
// @Description Get details of a specific subscription
// @Tags User-Subscription
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Subscription ID"
// @Success 200 {object} response.StandardResponse{data=entities.UserSubscriptionResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /subscriptions/{id} [get]
func (h *UserSubscriptionHandler) GetSubscription(c *gin.Context) {
	// Get current user from context
	userValue, exists := c.Get(middleware.AuthContextKey)
	if !exists {
		response.Unauthorized(c, "Authentication required")
		return
	}

	user, ok := userValue.(*userEntities.User)
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

	// Get subscription
	subscription, err := h.userSubscriptionService.GetUserSubscription(c.Request.Context(), uint(subscriptionID))
	if err != nil {
		if err.Error() == "subscription not found" {
			response.NotFound(c, "Subscription not found")
			return
		}
		logger.Error("Failed to get subscription", logger.Error2("error", err), logger.Uint("subscription_id", uint(subscriptionID)))
		response.InternalServerError(c, "Failed to get subscription", err.Error())
		return
	}

	// Check if user has access to this subscription
	if !user.IsAdmin() && subscription.UserID != user.ID {
		response.Forbidden(c, "You can only access your own subscriptions")
		return
	}

	response.OK(c, "Subscription retrieved successfully", subscription.ToResponse())
}

// CancelSubscription godoc
// @Summary [User] Cancel subscription
// @Description Cancel a subscription
// @Tags User-Subscription
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Subscription ID"
// @Param request body CancelSubscriptionRequest true "Cancel request"
// @Success 200 {object} response.StandardResponse{data=entities.UserSubscriptionResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /subscriptions/{id}/cancel [post]
func (h *UserSubscriptionHandler) CancelSubscription(c *gin.Context) {
	// Get current user from context
	userValue, exists := c.Get(middleware.AuthContextKey)
	if !exists {
		response.Unauthorized(c, "Authentication required")
		return
	}

	user, ok := userValue.(*userEntities.User)
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

	// Bind request
	var req CancelSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request data", err.Error())
		return
	}

	// Get subscription to check ownership
	subscription, err := h.userSubscriptionService.GetUserSubscription(c.Request.Context(), uint(subscriptionID))
	if err != nil {
		if err.Error() == "subscription not found" {
			response.NotFound(c, "Subscription not found")
			return
		}
		logger.Error("Failed to get subscription", logger.Error2("error", err), logger.Uint("subscription_id", uint(subscriptionID)))
		response.InternalServerError(c, "Failed to get subscription", err.Error())
		return
	}

	// Check if user has access to cancel this subscription
	if !user.IsAdmin() && subscription.UserID != user.ID {
		response.Forbidden(c, "You can only cancel your own subscriptions")
		return
	}

	// Cancel subscription
	if err := h.userSubscriptionService.CancelUserSubscription(c.Request.Context(), uint(subscriptionID), req.Reason, req.CancelAtPeriodEnd); err != nil {
		logger.Error("Failed to cancel subscription", logger.Error2("error", err), logger.Uint("subscription_id", uint(subscriptionID)))
		response.InternalServerError(c, "Failed to cancel subscription", err.Error())
		return
	}

	// Get updated subscription
	updatedSubscription, err := h.userSubscriptionService.GetUserSubscription(c.Request.Context(), uint(subscriptionID))
	if err != nil {
		logger.Error("Failed to get updated subscription", logger.Error2("error", err), logger.Uint("subscription_id", uint(subscriptionID)))
		response.InternalServerError(c, "Subscription cancelled but failed to get updated details", err.Error())
		return
	}

	response.OK(c, "Subscription cancelled successfully", updatedSubscription.ToResponse())
}

// GetSubscriptionTrafficStats godoc
// @Summary [User] Get subscription traffic statistics
// @Description Get traffic statistics for a subscription
// @Tags User-Subscription
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Subscription ID"
// @Success 200 {object} response.StandardResponse{data=map[string]interface{}}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /subscriptions/{id}/traffic-stats [get]
func (h *UserSubscriptionHandler) GetSubscriptionTrafficStats(c *gin.Context) {
	// Get current user from context
	userValue, exists := c.Get(middleware.AuthContextKey)
	if !exists {
		response.Unauthorized(c, "Authentication required")
		return
	}

	user, ok := userValue.(*userEntities.User)
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

	// Get subscription to check ownership
	subscription, err := h.userSubscriptionService.GetUserSubscription(c.Request.Context(), uint(subscriptionID))
	if err != nil {
		if err.Error() == "subscription not found" {
			response.NotFound(c, "Subscription not found")
			return
		}
		logger.Error("Failed to get subscription", logger.Error2("error", err), logger.Uint("subscription_id", uint(subscriptionID)))
		response.InternalServerError(c, "Failed to get subscription", err.Error())
		return
	}

	// Check if user has access to this subscription
	if !user.IsAdmin() && subscription.UserID != user.ID {
		response.Forbidden(c, "You can only access your own subscription traffic stats")
		return
	}

	// Get traffic statistics
	stats, err := h.userSubscriptionService.GetSubscriptionTrafficStats(c.Request.Context(), uint(subscriptionID))
	if err != nil {
		logger.Error("Failed to get traffic stats", logger.Error2("error", err), logger.Uint("subscription_id", uint(subscriptionID)))
		response.InternalServerError(c, "Failed to get traffic statistics", err.Error())
		return
	}

	response.OK(c, "Traffic statistics retrieved successfully", stats)
}

// CancelSubscriptionRequest represents the request to cancel a subscription
type CancelSubscriptionRequest struct {
	Reason            string `json:"reason" binding:"required,min=1,max=255" example:"No longer needed"`
	CancelAtPeriodEnd bool   `json:"cancel_at_period_end" example:"true"`
}

// PauseUserSubscription godoc
// @Summary [Admin] Pause user subscription
// @Description Pause a user subscription (admin only)
// @Tags Admin-Subscription
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Subscription ID"
// @Param request body interfaces.PauseSubscriptionRequest true "Pause subscription request"
// @Success 200 {object} response.StandardResponse{data=entities.UserSubscriptionResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/subscriptions/{id}/pause [post]
func (h *UserSubscriptionHandler) PauseUserSubscription(c *gin.Context) {
	// Get current user from context
	userValue, exists := c.Get(middleware.AuthContextKey)
	if !exists {
		response.Unauthorized(c, "Authentication required")
		return
	}

	user, ok := userValue.(*userEntities.User)
	if !ok {
		response.Unauthorized(c, "Invalid user context")
		return
	}

	// Check admin permission
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

	// Parse request body
	var req interfaces.PauseSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body", err.Error())
		return
	}

	// Pause subscription
	subscription, err := h.userSubscriptionService.PauseUserSubscription(c.Request.Context(), uint(subscriptionID), &req, user.ID)
	if err != nil {
		logger.Error("Failed to pause subscription",
			logger.Error2("error", err),
			logger.Uint("subscription_id", uint(subscriptionID)),
			logger.Uint("admin_user_id", user.ID))

		if err.Error() == "failed to get subscription: record not found" {
			response.NotFound(c, "Subscription not found")
			return
		}

		if err.Error() == "subscription cannot be paused - only active subscriptions can be paused" ||
			err.Error() == "subscription is already paused" {
			response.BadRequest(c, "Cannot pause subscription", err.Error())
			return
		}

		response.InternalServerError(c, "Failed to pause subscription", err.Error())
		return
	}

	response.OK(c, "Subscription paused successfully", subscription.ToResponse())
}

// ResumeUserSubscription godoc
// @Summary [Admin] Resume user subscription
// @Description Resume a paused user subscription (admin only)
// @Tags Admin-Subscription
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Subscription ID"
// @Param request body interfaces.ResumeSubscriptionRequest true "Resume subscription request"
// @Success 200 {object} response.StandardResponse{data=entities.UserSubscriptionResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/subscriptions/{id}/resume [post]
func (h *UserSubscriptionHandler) ResumeUserSubscription(c *gin.Context) {
	// Get current user from context
	userValue, exists := c.Get(middleware.AuthContextKey)
	if !exists {
		response.Unauthorized(c, "Authentication required")
		return
	}

	user, ok := userValue.(*userEntities.User)
	if !ok {
		response.Unauthorized(c, "Invalid user context")
		return
	}

	// Check admin permission
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

	// Parse request body (optional)
	var req interfaces.ResumeSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Use default values if no body provided
		req = interfaces.ResumeSubscriptionRequest{
			AdjustBillingDate: true, // Default to adjusting billing dates
		}
	}

	// Resume subscription
	subscription, err := h.userSubscriptionService.ResumeUserSubscription(c.Request.Context(), uint(subscriptionID), &req, user.ID)
	if err != nil {
		logger.Error("Failed to resume subscription",
			logger.Error2("error", err),
			logger.Uint("subscription_id", uint(subscriptionID)),
			logger.Uint("admin_user_id", user.ID))

		if err.Error() == "failed to get subscription: record not found" {
			response.NotFound(c, "Subscription not found")
			return
		}

		if err.Error() == "subscription cannot be resumed - only paused subscriptions can be resumed" ||
			err.Error() == "subscription has expired and cannot be resumed" {
			response.BadRequest(c, "Cannot resume subscription", err.Error())
			return
		}

		response.InternalServerError(c, "Failed to resume subscription", err.Error())
		return
	}

	response.OK(c, "Subscription resumed successfully", subscription.ToResponse())
}

// RegisterRoutes registers all user subscription routes
func (h *UserSubscriptionHandler) RegisterRoutes(router *gin.RouterGroup) {
	// User subscription routes - accessible to authenticated users
	subscriptionGroup := router.Group("/subscriptions")
	{
		// Public endpoint for testing (should be protected in production)
		subscriptionGroup.GET("", func(c *gin.Context) {
			response.Error(c, http.StatusUnauthorized, 4001, "User not authenticated - please login to access subscriptions")
		})
		
		subscriptionGroup.GET("/my", h.GetMySubscriptions)
		subscriptionGroup.GET("/my/active", h.GetMyActiveSubscriptions)
		subscriptionGroup.GET("/:id", h.GetSubscription)
		subscriptionGroup.POST("/:id/cancel", h.CancelSubscription)
		subscriptionGroup.GET("/:id/traffic-stats", h.GetSubscriptionTrafficStats)
	}

	// Admin subscription management routes
	adminGroup := router.Group("/admin/subscriptions")
	{
		adminGroup.POST("/:id/pause", h.PauseUserSubscription)
		adminGroup.POST("/:id/resume", h.ResumeUserSubscription)
	}
}
