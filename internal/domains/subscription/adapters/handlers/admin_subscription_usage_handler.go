package handlers

import (
	"strconv"
	"strings"

	"linke/internal/domains/subscription/usecases/interfaces"
	"linke/internal/shared/logger"
	"linke/internal/shared/response"

	"github.com/gin-gonic/gin"
)

// AdminSubscriptionUsageHandler handles usage tracking and alerts management operations
type AdminSubscriptionUsageHandler struct {
	*AdminSubscriptionHandlerBase
}

// NewAdminSubscriptionUsageHandler creates a new admin subscription usage handler
func NewAdminSubscriptionUsageHandler(base *AdminSubscriptionHandlerBase) *AdminSubscriptionUsageHandler {
	return &AdminSubscriptionUsageHandler{
		AdminSubscriptionHandlerBase: base,
	}
}

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
func (h *AdminSubscriptionUsageHandler) GetUsageStatistics(c *gin.Context) {
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

	response.OK(c, stats)
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
func (h *AdminSubscriptionUsageHandler) GetCurrentUsage(c *gin.Context) {
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

	response.OK(c, usage)
}

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
// @Success 200 {object} response.HALCollectionResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/subscriptions/alerts [get]
func (h *AdminSubscriptionUsageHandler) GetUsageAlerts(c *gin.Context) {
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

	_ = (req.Offset / req.Limit) + 1 // page calculation for future use
	response.SendPaginatedResponse(c, alertsResponse.UsageAlerts, alertsResponse.TotalCount)
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
func (h *AdminSubscriptionUsageHandler) GetAlertStatistics(c *gin.Context) {
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

	response.OK(c, stats)
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
func (h *AdminSubscriptionUsageHandler) BulkResolveAlerts(c *gin.Context) {
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

	response.OK(c, result)
}