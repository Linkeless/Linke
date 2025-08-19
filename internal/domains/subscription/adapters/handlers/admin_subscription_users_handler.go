package handlers

import (
	"strconv"
	"strings"

	"linke/internal/domains/subscription/usecases/interfaces"
	"linke/internal/shared/logger"
	"linke/internal/shared/response"

	"github.com/gin-gonic/gin"
)

// AdminSubscriptionUsersHandler handles user subscription lifecycle management operations
type AdminSubscriptionUsersHandler struct {
	*AdminSubscriptionHandlerBase
}

// NewAdminSubscriptionUsersHandler creates a new admin subscription users handler
func NewAdminSubscriptionUsersHandler(base *AdminSubscriptionHandlerBase) *AdminSubscriptionUsersHandler {
	return &AdminSubscriptionUsersHandler{
		AdminSubscriptionHandlerBase: base,
	}
}

// CreateUserSubscription godoc
// @Summary Create user subscription
// @Description Create a subscription for a user directly (Admin only)
// @Tags Admin-Subscription
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param subscription body AdminCreateUserSubscriptionRequest true "Subscription creation data"
// @Success 201 {object} entities.UserSubscriptionResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 409 {object} response.ConflictResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/subscriptions/users [post]
func (h *AdminSubscriptionUsersHandler) CreateUserSubscription(c *gin.Context) {
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
// @Success 200 {object} entities.UserSubscriptionResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Router /admin/subscriptions/users/{id} [get]
func (h *AdminSubscriptionUsersHandler) GetUserSubscription(c *gin.Context) {
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

	response.OK(c, subscription.ToResponse())
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
// @Success 200 {object} response.HALCollectionResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/subscriptions/users [get]
func (h *AdminSubscriptionUsersHandler) ListUserSubscriptions(c *gin.Context) {
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

	_ = (req.Offset / req.Limit) + 1 // page calculation for future use
	response.SendPaginatedResponse(c, subscriptionResponses, total)
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
// @Success 200 {object} entities.UserSubscriptionResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/subscriptions/users/{id} [put]
func (h *AdminSubscriptionUsersHandler) UpdateUserSubscription(c *gin.Context) {
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

	response.OK(c, subscription.ToResponse())
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
// @Success 200 {object} entities.UserSubscriptionResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/subscriptions/users/{id}/pause [post]
func (h *AdminSubscriptionUsersHandler) PauseUserSubscription(c *gin.Context) {
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

	response.OK(c, subscription.ToResponse())
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
// @Success 200 {object} entities.UserSubscriptionResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/subscriptions/users/{id}/resume [post]
func (h *AdminSubscriptionUsersHandler) ResumeUserSubscription(c *gin.Context) {
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

	response.OK(c, subscription.ToResponse())
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
// @Success 200 {object} entities.UserSubscriptionResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/subscriptions/users/{id}/extend [post]
func (h *AdminSubscriptionUsersHandler) ExtendUserSubscription(c *gin.Context) {
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

	response.OK(c, subscription.ToResponse())
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
// @Success 200 {object} string
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/subscriptions/users/{id}/cancel [post]
func (h *AdminSubscriptionUsersHandler) CancelUserSubscription(c *gin.Context) {
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

	response.NoContent(c)
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
// @Success 200 {object} entities.UserSubscriptionResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/subscriptions/users/{id}/reset-traffic [post]
func (h *AdminSubscriptionUsersHandler) ResetTrafficUsage(c *gin.Context) {
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

	response.OK(c, subscription.ToResponse())
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
// @Success 200 {object} entities.UserSubscriptionResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/subscriptions/users/{id}/upgrade [post]
func (h *AdminSubscriptionUsersHandler) UpgradeSubscription(c *gin.Context) {
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

	response.OK(c, subscription.ToResponse())
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
// @Success 200 {object} entities.UserSubscriptionResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/subscriptions/users/{id}/downgrade [post]
func (h *AdminSubscriptionUsersHandler) DowngradeSubscription(c *gin.Context) {
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

	response.OK(c, subscription.ToResponse())
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
func (h *AdminSubscriptionUsersHandler) GetSubscriptionStatistics(c *gin.Context) {
	stats, err := h.userSubscriptionService.GetSubscriptionStatistics(c.Request.Context())
	if err != nil {
		logger.Error("Admin failed to get subscription statistics", logger.ErrorField(err))
		response.InternalServerError(c, "Failed to get subscription statistics")
		return
	}

	response.OK(c, stats)
}