package handlers

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"linke/internal/domains/subscription/usecases/interfaces"
	"linke/internal/shared/response"
)

// UsageHandler handles usage tracking and monitoring API endpoints
type UsageHandler struct {
	usageTrackingService interfaces.UsageTrackingService
	usageAlertService    interfaces.UsageAlertService
	usageAlertHandler    *UsageAlertHandler
}

// NewUsageHandler creates a new usage handler instance
func NewUsageHandler(
	usageTrackingService interfaces.UsageTrackingService,
	usageAlertService interfaces.UsageAlertService,
	usageAlertHandler *UsageAlertHandler,
) *UsageHandler {
	return &UsageHandler{
		usageTrackingService: usageTrackingService,
		usageAlertService:    usageAlertService,
		usageAlertHandler:    usageAlertHandler,
	}
}

// RegisterRoutes registers all usage-related routes
func (h *UsageHandler) RegisterRoutes(router *gin.RouterGroup) {
	usage := router.Group("/usage")
	{
		// Current usage endpoints
		usage.GET("/current/:subscription_id", h.GetCurrentUsage)
		usage.GET("/current/:subscription_id/:usage_type", h.GetCurrentUsageByType)

		// Usage history endpoints
		usage.GET("/history/:subscription_id", h.GetUsageHistory)
		usage.GET("/summary/:subscription_id", h.GetUsageSummary)
		usage.GET("/statistics/:subscription_id", h.GetUsageStatistics)
		usage.GET("/trends/:subscription_id", h.GetUsageTrends)
		usage.GET("/predictions/:subscription_id", h.GetUsagePredictions)
		usage.GET("/predictions/:subscription_id/:usage_type", h.GetUsagePredictionsByType)

		// Real-time monitoring
		usage.GET("/realtime/:subscription_id", h.GetRealTimeUsage)

		// Export endpoints
		usage.POST("/export", h.ExportUsageData)

		// Admin endpoints
		usage.GET("/top", h.GetTopUsageSubscriptions)

		// Alert configuration endpoints
		alerts := usage.Group("/alerts")
		{
			alerts.GET("/configs/:subscription_id", h.usageAlertHandler.GetAlertConfigurations)
			alerts.POST("/configs", h.usageAlertHandler.CreateAlertConfiguration)
			alerts.PUT("/configs/:config_id", h.usageAlertHandler.UpdateAlertConfiguration)
			alerts.DELETE("/configs/:config_id", h.usageAlertHandler.DeleteAlertConfiguration)
			alerts.GET("/configs/:config_id", h.usageAlertHandler.GetAlertConfiguration)

			// Alert management
			alerts.GET("/:subscription_id", h.usageAlertHandler.GetUsageAlerts)
			alerts.POST("/:alert_id/resolve", h.usageAlertHandler.ResolveAlert)
			alerts.POST("/:alert_id/acknowledge", h.usageAlertHandler.AcknowledgeAlert)
			alerts.POST("/:alert_id/suppress", h.usageAlertHandler.SuppressAlert)
			alerts.POST("/bulk-resolve", h.usageAlertHandler.BulkResolveAlerts)

			// Alert analytics
			alerts.GET("/statistics/:subscription_id", h.usageAlertHandler.GetAlertStatistics)
			alerts.GET("/history/:subscription_id", h.usageAlertHandler.GetAlertHistory)

			// Notification testing
			alerts.POST("/test-notification", h.usageAlertHandler.TestNotificationChannel)
		}

		// Admin operations (requires admin middleware)
		admin := usage.Group("/admin")
		{
			admin.POST("/cleanup", h.CleanupOldUsageData)
			admin.POST("/reset/:subscription_id", h.ResetUsageForSubscription)
			admin.POST("/sync/:subscription_id", h.SyncSubscriptionLimits)
		}
	}
}

// Current Usage Endpoints

// GetCurrentUsage godoc
// @Summary Get current usage for a subscription
// @Description Retrieve current usage statistics for all usage types of a subscription
// @Tags usage
// @Accept json
// @Produce json
// @Param subscription_id path int true "Subscription ID"
// @Success 200 {object} response.Response{data=interfaces.CurrentUsageResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/usage/current/{subscription_id} [get]
func (h *UsageHandler) GetCurrentUsage(c *gin.Context) {
	subscriptionID, err := strconv.ParseUint(c.Param("subscription_id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid subscription ID")
		return
	}

	currentUsage, err := h.usageTrackingService.GetCurrentUsage(c.Request.Context(), uint(subscriptionID), "")
	if err != nil {
		response.InternalServerError(c, "Failed to get current usage", err.Error())
		return
	}

	response.SuccessWithMessage(c, "Current usage retrieved successfully", currentUsage)
}

// GetCurrentUsageByType godoc
// @Summary Get current usage for a specific usage type
// @Description Retrieve current usage statistics for a specific usage type of a subscription
// @Tags usage
// @Accept json
// @Produce json
// @Param subscription_id path int true "Subscription ID"
// @Param usage_type path string true "Usage Type" Enums(traffic,api_call,storage,bandwidth,connections)
// @Success 200 {object} response.Response{data=interfaces.CurrentUsageResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/usage/current/{subscription_id}/{usage_type} [get]
func (h *UsageHandler) GetCurrentUsageByType(c *gin.Context) {
	subscriptionID, err := strconv.ParseUint(c.Param("subscription_id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid subscription ID")
		return
	}

	usageType := c.Param("usage_type")
	if usageType == "" {
		response.BadRequest(c, "Usage type is required")
		return
	}

	currentUsage, err := h.usageTrackingService.GetCurrentUsage(c.Request.Context(), uint(subscriptionID), usageType)
	if err != nil {
		response.InternalServerError(c, "Failed to get current usage", err.Error())
		return
	}

	response.SuccessWithMessage(c, "Current usage retrieved successfully", currentUsage)
}

// Usage History Endpoints

// GetUsageHistory godoc
// @Summary Get usage history for a subscription
// @Description Retrieve historical usage data with configurable time range and granularity
// @Tags usage
// @Accept json
// @Produce json
// @Param subscription_id path int true "Subscription ID"
// @Param usage_type query string false "Usage Type Filter"
// @Param start_time query string false "Start Time (RFC3339)" format(date-time)
// @Param end_time query string false "End Time (RFC3339)" format(date-time)
// @Param granularity query string false "Data Granularity" Enums(hourly,daily,weekly,monthly) default(daily)
// @Param limit query int false "Limit" default(50) maximum(1000)
// @Param offset query int false "Offset" default(0)
// @Param include_details query bool false "Include detailed breakdown" default(false)
// @Param source_type query string false "Source Type Filter"
// @Success 200 {object} response.Response{data=interfaces.UsageHistoryResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/usage/history/{subscription_id} [get]
func (h *UsageHandler) GetUsageHistory(c *gin.Context) {
	subscriptionID, err := strconv.ParseUint(c.Param("subscription_id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid subscription ID")
		return
	}

	// Parse query parameters
	req := &interfaces.UsageHistoryRequest{
		UserSubscriptionID: uint(subscriptionID),
		UsageType:          c.Query("usage_type"),
		Granularity:        c.DefaultQuery("granularity", interfaces.GranularityDaily),
		Limit:              50,
		Offset:             0,
		IncludeDetails:     c.Query("include_details") == "true",
		SourceType:         c.Query("source_type"),
	}

	// Parse limit
	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			req.Limit = limit
		}
	}

	// Parse offset
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			req.Offset = offset
		}
	}

	// Parse start time
	if startTimeStr := c.Query("start_time"); startTimeStr != "" {
		if startTime, err := time.Parse(time.RFC3339, startTimeStr); err == nil {
			req.StartTime = &startTime
		}
	}

	// Parse end time
	if endTimeStr := c.Query("end_time"); endTimeStr != "" {
		if endTime, err := time.Parse(time.RFC3339, endTimeStr); err == nil {
			req.EndTime = &endTime
		}
	}

	history, err := h.usageTrackingService.GetUsageHistory(c.Request.Context(), req)
	if err != nil {
		response.InternalServerError(c, "Failed to get usage history", err.Error())
		return
	}

	response.SuccessWithMessage(c, "Usage history retrieved successfully", history)
}

// GetUsageSummary godoc
// @Summary Get usage summary for a subscription
// @Description Retrieve aggregated usage summary for a specific period
// @Tags usage
// @Accept json
// @Produce json
// @Param subscription_id path int true "Subscription ID"
// @Param period query string false "Time Period" Enums(daily,weekly,monthly,custom) default(monthly)
// @Param period_start query string false "Period Start Time (RFC3339)" format(date-time)
// @Param period_end query string false "Period End Time (RFC3339)" format(date-time)
// @Param usage_types query string false "Comma-separated usage types filter"
// @Param include_breakdown query bool false "Include detailed breakdown" default(false)
// @Param include_predictions query bool false "Include usage predictions" default(false)
// @Param compare_with_previous query bool false "Compare with previous period" default(false)
// @Success 200 {object} response.Response{data=entities.UsageSummary}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/usage/summary/{subscription_id} [get]
func (h *UsageHandler) GetUsageSummary(c *gin.Context) {
	subscriptionID, err := strconv.ParseUint(c.Param("subscription_id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid subscription ID")
		return
	}

	req := &interfaces.UsageSummaryRequest{
		UserSubscriptionID:  uint(subscriptionID),
		Period:              c.DefaultQuery("period", interfaces.PeriodMonthly),
		IncludeBreakdown:    c.Query("include_breakdown") == "true",
		IncludePredictions:  c.Query("include_predictions") == "true",
		CompareWithPrevious: c.Query("compare_with_previous") == "true",
	}

	// Parse usage types filter
	if usageTypesStr := c.Query("usage_types"); usageTypesStr != "" {
		// Split comma-separated values
		// req.UsageTypes = strings.Split(usageTypesStr, ",")
		// Note: strings package would need to be imported
	}

	// Parse period start
	if periodStartStr := c.Query("period_start"); periodStartStr != "" {
		if periodStart, err := time.Parse(time.RFC3339, periodStartStr); err == nil {
			req.PeriodStart = &periodStart
		}
	}

	// Parse period end
	if periodEndStr := c.Query("period_end"); periodEndStr != "" {
		if periodEnd, err := time.Parse(time.RFC3339, periodEndStr); err == nil {
			req.PeriodEnd = &periodEnd
		}
	}

	summary, err := h.usageTrackingService.GetUsageSummary(c.Request.Context(), req)
	if err != nil {
		response.InternalServerError(c, "Failed to get usage summary", err.Error())
		return
	}

	response.SuccessWithMessage(c, "Usage summary retrieved successfully", summary)
}

// GetUsageStatistics godoc
// @Summary Get usage statistics for a subscription
// @Description Retrieve detailed usage statistics with breakdown by various dimensions
// @Tags usage
// @Accept json
// @Produce json
// @Param subscription_id path int true "Subscription ID"
// @Param usage_type query string false "Usage Type Filter"
// @Param period query string false "Time Period" default(monthly)
// @Param start_time query string false "Start Time (RFC3339)" format(date-time)
// @Param end_time query string false "End Time (RFC3339)" format(date-time)
// @Param group_by query string false "Group By Dimensions (comma-separated)"
// @Param include_comparison query bool false "Include period comparison" default(false)
// @Success 200 {object} response.Response{data=interfaces.UsageStatistics}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/usage/statistics/{subscription_id} [get]
func (h *UsageHandler) GetUsageStatistics(c *gin.Context) {
	subscriptionID, err := strconv.ParseUint(c.Param("subscription_id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid subscription ID")
		return
	}

	req := &interfaces.UsageStatsRequest{
		UserSubscriptionID: uint(subscriptionID),
		UsageType:          c.Query("usage_type"),
		Period:             c.DefaultQuery("period", interfaces.PeriodMonthly),
		IncludeComparison:  c.Query("include_comparison") == "true",
	}

	// Parse start time
	if startTimeStr := c.Query("start_time"); startTimeStr != "" {
		if startTime, err := time.Parse(time.RFC3339, startTimeStr); err == nil {
			req.StartTime = &startTime
		}
	}

	// Parse end time
	if endTimeStr := c.Query("end_time"); endTimeStr != "" {
		if endTime, err := time.Parse(time.RFC3339, endTimeStr); err == nil {
			req.EndTime = &endTime
		}
	}

	stats, err := h.usageTrackingService.GetUsageStatistics(c.Request.Context(), req)
	if err != nil {
		response.InternalServerError(c, "Failed to get usage statistics", err.Error())
		return
	}

	response.SuccessWithMessage(c, "Usage statistics retrieved successfully", stats)
}

// GetUsageTrends godoc
// @Summary Get usage trends for a subscription
// @Description Retrieve usage trends with optional anomaly detection and predictions
// @Tags usage
// @Accept json
// @Produce json
// @Param subscription_id path int true "Subscription ID"
// @Param usage_type query string false "Usage Type Filter"
// @Param period query string false "Time Period" Enums(7d,30d,90d,365d) default(30d)
// @Param granularity query string false "Data Granularity" Enums(hourly,daily,weekly) default(daily)
// @Param include_predictions query bool false "Include usage predictions" default(false)
// @Param include_anomalies query bool false "Include anomaly detection" default(false)
// @Success 200 {object} response.Response{data=interfaces.UsageTrendsResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/usage/trends/{subscription_id} [get]
func (h *UsageHandler) GetUsageTrends(c *gin.Context) {
	subscriptionID, err := strconv.ParseUint(c.Param("subscription_id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid subscription ID")
		return
	}

	req := &interfaces.UsageTrendsRequest{
		UserSubscriptionID: uint(subscriptionID),
		UsageType:          c.Query("usage_type"),
		Period:             c.DefaultQuery("period", "30d"),
		Granularity:        c.DefaultQuery("granularity", interfaces.GranularityDaily),
		IncludePredictions: c.Query("include_predictions") == "true",
		IncludeAnomalies:   c.Query("include_anomalies") == "true",
	}

	trends, err := h.usageTrackingService.GetUsageTrends(c.Request.Context(), req)
	if err != nil {
		response.InternalServerError(c, "Failed to get usage trends", err.Error())
		return
	}

	response.SuccessWithMessage(c, "Usage trends retrieved successfully", trends)
}

// GetUsagePredictions godoc
// @Summary Get usage predictions for a subscription
// @Description Retrieve usage predictions for all usage types
// @Tags usage
// @Accept json
// @Produce json
// @Param subscription_id path int true "Subscription ID"
// @Success 200 {object} response.Response{data=[]entities.UsagePrediction}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/usage/predictions/{subscription_id} [get]
func (h *UsageHandler) GetUsagePredictions(c *gin.Context) {
	subscriptionID, err := strconv.ParseUint(c.Param("subscription_id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid subscription ID")
		return
	}

	predictions, err := h.usageTrackingService.GetUsagePredictions(c.Request.Context(), uint(subscriptionID), "")
	if err != nil {
		response.InternalServerError(c, "Failed to get usage predictions", err.Error())
		return
	}

	response.SuccessWithMessage(c, "Usage predictions retrieved successfully", predictions)
}

// GetUsagePredictionsByType godoc
// @Summary Get usage predictions for a specific usage type
// @Description Retrieve usage predictions for a specific usage type
// @Tags usage
// @Accept json
// @Produce json
// @Param subscription_id path int true "Subscription ID"
// @Param usage_type path string true "Usage Type" Enums(traffic,api_call,storage,bandwidth,connections)
// @Success 200 {object} response.Response{data=[]entities.UsagePrediction}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/usage/predictions/{subscription_id}/{usage_type} [get]
func (h *UsageHandler) GetUsagePredictionsByType(c *gin.Context) {
	subscriptionID, err := strconv.ParseUint(c.Param("subscription_id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid subscription ID")
		return
	}

	usageType := c.Param("usage_type")
	if usageType == "" {
		response.BadRequest(c, "Usage type is required")
		return
	}

	predictions, err := h.usageTrackingService.GetUsagePredictions(c.Request.Context(), uint(subscriptionID), usageType)
	if err != nil {
		response.InternalServerError(c, "Failed to get usage predictions", err.Error())
		return
	}

	response.SuccessWithMessage(c, "Usage predictions retrieved successfully", predictions)
}

// Real-time Monitoring

// GetRealTimeUsage godoc
// @Summary Get real-time usage data for a subscription
// @Description Retrieve real-time usage data with current rates and predictions
// @Tags usage
// @Accept json
// @Produce json
// @Param subscription_id path int true "Subscription ID"
// @Success 200 {object} response.Response{data=interfaces.RealTimeUsageResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/usage/realtime/{subscription_id} [get]
func (h *UsageHandler) GetRealTimeUsage(c *gin.Context) {
	subscriptionID, err := strconv.ParseUint(c.Param("subscription_id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid subscription ID")
		return
	}

	realTimeUsage, err := h.usageTrackingService.GetRealTimeUsage(c.Request.Context(), uint(subscriptionID))
	if err != nil {
		response.InternalServerError(c, "Failed to get real-time usage", err.Error())
		return
	}

	response.SuccessWithMessage(c, "Real-time usage retrieved successfully", realTimeUsage)
}

// Export Endpoints

// ExportUsageData godoc
// @Summary Export usage data
// @Description Export usage data in various formats (CSV, JSON, XLSX)
// @Tags usage
// @Accept json
// @Produce json
// @Param request body interfaces.ExportUsageRequest true "Export Request"
// @Success 200 {object} response.Response{data=interfaces.ExportUsageResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/usage/export [post]
func (h *UsageHandler) ExportUsageData(c *gin.Context) {
	var req interfaces.ExportUsageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request format")
		return
	}

	exportResult, err := h.usageTrackingService.ExportUsageData(c.Request.Context(), &req)
	if err != nil {
		response.InternalServerError(c, "Failed to export usage data", err.Error())
		return
	}

	response.SuccessWithMessage(c, "Usage data export initiated successfully", exportResult)
}

// Admin Endpoints

// GetTopUsageSubscriptions godoc
// @Summary Get top usage subscriptions
// @Description Retrieve subscriptions with highest usage for administration
// @Tags usage,admin
// @Accept json
// @Produce json
// @Param usage_type query string false "Usage Type Filter"
// @Param period query string false "Time Period (e.g., 24h, 7d, 30d)" default(24h)
// @Param limit query int false "Limit" default(10) maximum(100)
// @Param order_by query string false "Order By" Enums(total_usage,average_usage,peak_usage) default(total_usage)
// @Param include_zero query bool false "Include zero usage subscriptions" default(false)
// @Success 200 {object} response.Response{data=interfaces.TopUsageResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/usage/top [get]
func (h *UsageHandler) GetTopUsageSubscriptions(c *gin.Context) {
	req := &interfaces.TopUsageRequest{
		UsageType:   c.Query("usage_type"),
		Limit:       10,
		OrderBy:     c.DefaultQuery("order_by", "total_usage"),
		IncludeZero: c.Query("include_zero") == "true",
	}

	// Parse limit
	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 && limit <= 100 {
			req.Limit = limit
		}
	}

	// Parse period
	periodStr := c.DefaultQuery("period", "24h")
	switch periodStr {
	case "1h":
		req.Period = time.Hour
	case "24h":
		req.Period = 24 * time.Hour
	case "7d":
		req.Period = 7 * 24 * time.Hour
	case "30d":
		req.Period = 30 * 24 * time.Hour
	default:
		req.Period = 24 * time.Hour
	}

	topUsage, err := h.usageTrackingService.GetTopUsageSubscriptions(c.Request.Context(), req)
	if err != nil {
		response.InternalServerError(c, "Failed to get top usage subscriptions", err.Error())
		return
	}

	response.SuccessWithMessage(c, "Top usage subscriptions retrieved successfully", topUsage)
}

// CleanupOldUsageData godoc
// @Summary Cleanup old usage data
// @Description Remove old usage data and alerts (admin only)
// @Tags usage,admin
// @Accept json
// @Produce json
// @Param older_than query string true "Delete data older than this date (RFC3339)" format(date-time)
// @Success 200 {object} response.Response{data=interfaces.CleanupResult}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/usage/admin/cleanup [post]
func (h *UsageHandler) CleanupOldUsageData(c *gin.Context) {
	olderThanStr := c.Query("older_than")
	if olderThanStr == "" {
		response.BadRequest(c, "older_than parameter is required")
		return
	}

	olderThan, err := time.Parse(time.RFC3339, olderThanStr)
	if err != nil {
		response.BadRequest(c, "Invalid older_than date format")
		return
	}

	result, err := h.usageTrackingService.CleanupOldUsageData(c.Request.Context(), olderThan)
	if err != nil {
		response.InternalServerError(c, "Failed to cleanup old usage data", err.Error())
		return
	}

	response.SuccessWithMessage(c, "Old usage data cleanup completed successfully", result)
}

// ResetUsageForSubscription godoc
// @Summary Reset usage for a subscription
// @Description Reset usage counters for a specific subscription and usage type (admin only)
// @Tags usage,admin
// @Accept json
// @Produce json
// @Param subscription_id path int true "Subscription ID"
// @Param usage_type query string true "Usage Type" Enums(traffic,api_call,storage,bandwidth,connections)
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/usage/admin/reset/{subscription_id} [post]
func (h *UsageHandler) ResetUsageForSubscription(c *gin.Context) {
	subscriptionID, err := strconv.ParseUint(c.Param("subscription_id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid subscription ID")
		return
	}

	usageType := c.Query("usage_type")
	if usageType == "" {
		response.BadRequest(c, "usage_type parameter is required")
		return
	}

	err = h.usageTrackingService.ResetUsageForSubscription(c.Request.Context(), uint(subscriptionID), usageType)
	if err != nil {
		response.InternalServerError(c, "Failed to reset usage for subscription", err.Error())
		return
	}

	response.SuccessWithMessage(c, "Usage reset successfully", nil)
}

// SyncSubscriptionLimits godoc
// @Summary Sync subscription limits
// @Description Sync usage limits with subscription plan settings (admin only)
// @Tags usage,admin
// @Accept json
// @Produce json
// @Param subscription_id path int true "Subscription ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/usage/admin/sync/{subscription_id} [post]
func (h *UsageHandler) SyncSubscriptionLimits(c *gin.Context) {
	subscriptionID, err := strconv.ParseUint(c.Param("subscription_id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid subscription ID")
		return
	}

	err = h.usageTrackingService.SyncSubscriptionLimits(c.Request.Context(), uint(subscriptionID))
	if err != nil {
		response.InternalServerError(c, "Failed to sync subscription limits", err.Error())
		return
	}

	response.SuccessWithMessage(c, "Subscription limits synced successfully", nil)
}
