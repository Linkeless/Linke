package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"linke/internal/domains/subscription/usecases/interfaces"
	"linke/internal/shared/response"
)

// UsageAlertHandler handles usage alert API endpoints
type UsageAlertHandler struct {
	usageAlertService interfaces.UsageAlertService
}

// NewUsageAlertHandler creates a new usage alert handler instance
func NewUsageAlertHandler(usageAlertService interfaces.UsageAlertService) *UsageAlertHandler {
	return &UsageAlertHandler{
		usageAlertService: usageAlertService,
	}
}

// Alert Configuration Management

// GetAlertConfigurations godoc
// @Summary Get alert configurations for a subscription
// @Description Retrieve all alert configurations for a specific subscription
// @Tags alerts
// @Accept json
// @Produce json
// @Param subscription_id path int true "Subscription ID"
// @Param usage_type query string false "Usage Type Filter"
// @Param is_enabled query bool false "Enabled Status Filter"
// @Param priority query string false "Priority Filter" Enums(low,medium,high,critical)
// @Param limit query int false "Limit" default(50) maximum(1000)
// @Param offset query int false "Offset" default(0)
// @Param order_by query string false "Order By" Enums(name,threshold,priority,created_at) default(created_at)
// @Param order_direction query string false "Order Direction" Enums(asc,desc) default(desc)
// @Success 200 {object} response.Response{data=interfaces.GetAlertConfigsResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /usage-alerts/subscription/{subscription_id}/configurations [get]
func (h *UsageAlertHandler) GetAlertConfigurations(c *gin.Context) {
	subscriptionID, err := strconv.ParseUint(c.Param("subscription_id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid subscription ID")
		return
	}

	req := &interfaces.GetAlertConfigsRequest{
		UserSubscriptionID: uint(subscriptionID),
		UsageType:          c.Query("usage_type"),
		Priority:           c.Query("priority"),
		Limit:              50,
		Offset:             0,
		OrderBy:            c.DefaultQuery("order_by", "created_at"),
		OrderDirection:     c.DefaultQuery("order_direction", "desc"),
	}

	// Parse is_enabled filter
	if isEnabledStr := c.Query("is_enabled"); isEnabledStr != "" {
		if isEnabled, err := strconv.ParseBool(isEnabledStr); err == nil {
			req.IsEnabled = &isEnabled
		}
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

	configs, err := h.usageAlertService.GetAlertConfigurations(c.Request.Context(), req)
	if err != nil {
		response.InternalServerError(c, "Failed to get alert configurations", err.Error())
		return
	}

	response.SuccessWithMessage(c, "Alert configurations retrieved successfully", configs)
}

// CreateAlertConfiguration godoc
// @Summary Create a new alert configuration
// @Description Create a new alert configuration for usage monitoring
// @Tags alerts
// @Accept json
// @Produce json
// @Param request body interfaces.CreateAlertConfigRequest true "Alert Configuration Request"
// @Success 201 {object} response.Response{data=entities.AlertConfigurationResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /usage-alerts/configurations [post]
func (h *UsageAlertHandler) CreateAlertConfiguration(c *gin.Context) {
	var req interfaces.CreateAlertConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request format")
		return
	}

	config, err := h.usageAlertService.CreateAlertConfiguration(c.Request.Context(), &req)
	if err != nil {
		response.InternalServerError(c, "Failed to create alert configuration", err.Error())
		return
	}

	response.CreatedWithMessage(c, "Alert configuration created successfully", config.ToResponse())
}

// UpdateAlertConfiguration godoc
// @Summary Update an alert configuration
// @Description Update an existing alert configuration
// @Tags alerts
// @Accept json
// @Produce json
// @Param config_id path int true "Configuration ID"
// @Param request body interfaces.UpdateAlertConfigRequest true "Update Request"
// @Success 200 {object} response.Response{data=entities.AlertConfigurationResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /usage-alerts/config/{config_id} [put]
func (h *UsageAlertHandler) UpdateAlertConfiguration(c *gin.Context) {
	configID, err := strconv.ParseUint(c.Param("config_id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid configuration ID")
		return
	}

	var req interfaces.UpdateAlertConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request format")
		return
	}

	req.ID = uint(configID)

	config, err := h.usageAlertService.UpdateAlertConfiguration(c.Request.Context(), &req)
	if err != nil {
		response.InternalServerError(c, "Failed to update alert configuration", err.Error())
		return
	}

	response.SuccessWithMessage(c, "Alert configuration updated successfully", config.ToResponse())
}

// DeleteAlertConfiguration godoc
// @Summary Delete an alert configuration
// @Description Delete (soft delete) an alert configuration
// @Tags alerts
// @Accept json
// @Produce json
// @Param config_id path int true "Configuration ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /usage-alerts/config/{config_id} [delete]
func (h *UsageAlertHandler) DeleteAlertConfiguration(c *gin.Context) {
	configID, err := strconv.ParseUint(c.Param("config_id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid configuration ID")
		return
	}

	err = h.usageAlertService.DeleteAlertConfiguration(c.Request.Context(), uint(configID))
	if err != nil {
		response.InternalServerError(c, "Failed to delete alert configuration", err.Error())
		return
	}

	response.SuccessWithMessage(c, "Alert configuration deleted successfully", nil)
}

// GetAlertConfiguration godoc
// @Summary Get a specific alert configuration
// @Description Retrieve details of a specific alert configuration
// @Tags alerts
// @Accept json
// @Produce json
// @Param config_id path int true "Configuration ID"
// @Success 200 {object} response.Response{data=entities.AlertConfigurationResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /usage-alerts/config/{config_id} [get]
func (h *UsageAlertHandler) GetAlertConfiguration(c *gin.Context) {
	configID, err := strconv.ParseUint(c.Param("config_id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid configuration ID")
		return
	}

	config, err := h.usageAlertService.GetAlertConfiguration(c.Request.Context(), uint(configID))
	if err != nil {
		response.InternalServerError(c, "Failed to get alert configuration", err.Error())
		return
	}

	response.SuccessWithMessage(c, "Alert configuration retrieved successfully", config.ToResponse())
}

// Alert Management

// GetUsageAlerts godoc
// @Summary Get usage alerts for a subscription
// @Description Retrieve usage alerts with optional filtering
// @Tags alerts
// @Accept json
// @Produce json
// @Param subscription_id path int true "Subscription ID"
// @Param alert_configuration_id query int false "Alert Configuration ID Filter"
// @Param usage_type query string false "Usage Type Filter"
// @Param status query string false "Status Filter" Enums(fired,resolved,suppressed,acknowledged)
// @Param severity query string false "Severity Filter" Enums(info,warning,error,critical)
// @Param is_active query bool false "Active Status Filter"
// @Param is_resolved query bool false "Resolved Status Filter"
// @Param start_time query string false "Start Time (RFC3339)" format(date-time)
// @Param end_time query string false "End Time (RFC3339)" format(date-time)
// @Param limit query int false "Limit" default(50) maximum(1000)
// @Param offset query int false "Offset" default(0)
// @Param order_by query string false "Order By" Enums(fired_at,resolved_at,severity,usage_percent) default(fired_at)
// @Param order_direction query string false "Order Direction" Enums(asc,desc) default(desc)
// @Success 200 {object} response.Response{data=interfaces.GetUsageAlertsResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /usage-alerts/subscription/{subscription_id} [get]
func (h *UsageAlertHandler) GetUsageAlerts(c *gin.Context) {
	subscriptionID, err := strconv.ParseUint(c.Param("subscription_id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid subscription ID")
		return
	}

	req := &interfaces.GetUsageAlertsRequest{
		UserSubscriptionID: uint(subscriptionID),
		UsageType:          c.Query("usage_type"),
		Status:             c.Query("status"),
		Severity:           c.Query("severity"),
		Limit:              50,
		Offset:             0,
		OrderBy:            c.DefaultQuery("order_by", "fired_at"),
		OrderDirection:     c.DefaultQuery("order_direction", "desc"),
	}

	// Parse alert_configuration_id
	if configIDStr := c.Query("alert_configuration_id"); configIDStr != "" {
		if configID, err := strconv.ParseUint(configIDStr, 10, 32); err == nil {
			req.AlertConfigurationID = uint(configID)
		}
	}

	// Parse boolean filters
	if isActiveStr := c.Query("is_active"); isActiveStr != "" {
		if isActive, err := strconv.ParseBool(isActiveStr); err == nil {
			req.IsActive = &isActive
		}
	}

	if isResolvedStr := c.Query("is_resolved"); isResolvedStr != "" {
		if isResolved, err := strconv.ParseBool(isResolvedStr); err == nil {
			req.IsResolved = &isResolved
		}
	}

	// Parse time filters
	if startTimeStr := c.Query("start_time"); startTimeStr != "" {
		if startTime, err := time.Parse(time.RFC3339, startTimeStr); err == nil {
			req.StartTime = &startTime
		}
	}

	if endTimeStr := c.Query("end_time"); endTimeStr != "" {
		if endTime, err := time.Parse(time.RFC3339, endTimeStr); err == nil {
			req.EndTime = &endTime
		}
	}

	// Parse pagination
	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			req.Limit = limit
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			req.Offset = offset
		}
	}

	alerts, err := h.usageAlertService.GetUsageAlerts(c.Request.Context(), req)
	if err != nil {
		response.InternalServerError(c, "Failed to get usage alerts", err.Error())
		return
	}

	response.SuccessWithMessage(c, "Usage alerts retrieved successfully", alerts)
}

// ResolveAlert godoc
// @Summary Resolve a usage alert
// @Description Mark a usage alert as resolved
// @Tags alerts
// @Accept json
// @Produce json
// @Param alert_id path int true "Alert ID"
// @Param request body map[string]string false "Resolve Request" example({"reason":"Issue fixed"})
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /usage-alerts/alert/{alert_id}/resolve [post]
func (h *UsageAlertHandler) ResolveAlert(c *gin.Context) {
	alertID, err := strconv.ParseUint(c.Param("alert_id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid alert ID")
		return
	}

	var req map[string]string
	if err := c.ShouldBindJSON(&req); err != nil {
		// If no body provided, use empty reason
		req = make(map[string]string)
	}

	reason := req["reason"]
	err = h.usageAlertService.ResolveAlert(c.Request.Context(), uint(alertID), reason)
	if err != nil {
		response.InternalServerError(c, "Failed to resolve alert", err.Error())
		return
	}

	response.SuccessWithMessage(c, "Alert resolved successfully", nil)
}

// AcknowledgeAlert godoc
// @Summary Acknowledge a usage alert
// @Description Mark a usage alert as acknowledged
// @Tags alerts
// @Accept json
// @Produce json
// @Param alert_id path int true "Alert ID"
// @Param request body map[string]int false "Acknowledge Request" example({"acknowledged_by":1})
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /usage-alerts/alert/{alert_id}/acknowledge [post]
func (h *UsageAlertHandler) AcknowledgeAlert(c *gin.Context) {
	alertID, err := strconv.ParseUint(c.Param("alert_id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid alert ID")
		return
	}

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request format")
		return
	}

	acknowledgedBy := uint(0)
	if ackBy, ok := req["acknowledged_by"].(float64); ok {
		acknowledgedBy = uint(ackBy)
	}

	err = h.usageAlertService.AcknowledgeAlert(c.Request.Context(), uint(alertID), acknowledgedBy)
	if err != nil {
		response.InternalServerError(c, "Failed to acknowledge alert", err.Error())
		return
	}

	response.SuccessWithMessage(c, "Alert acknowledged successfully", nil)
}

// SuppressAlert godoc
// @Summary Suppress a usage alert
// @Description Suppress a usage alert for a specified duration
// @Tags alerts
// @Accept json
// @Produce json
// @Param alert_id path int true "Alert ID"
// @Param request body map[string]interface{} true "Suppress Request" example({"duration_minutes":60,"reason":"Maintenance window"})
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /usage-alerts/alert/{alert_id}/suppress [post]
func (h *UsageAlertHandler) SuppressAlert(c *gin.Context) {
	alertID, err := strconv.ParseUint(c.Param("alert_id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid alert ID")
		return
	}

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request format")
		return
	}

	durationMinutes := 60 // Default to 1 hour
	if durMins, ok := req["duration_minutes"].(float64); ok {
		durationMinutes = int(durMins)
	}

	reason := ""
	if r, ok := req["reason"].(string); ok {
		reason = r
	}

	duration := time.Duration(durationMinutes) * time.Minute
	err = h.usageAlertService.SuppressAlert(c.Request.Context(), uint(alertID), duration, reason)
	if err != nil {
		response.InternalServerError(c, "Failed to suppress alert", err.Error())
		return
	}

	response.SuccessWithMessage(c, "Alert suppressed successfully", nil)
}

// BulkResolveAlerts godoc
// @Summary Bulk resolve usage alerts
// @Description Resolve multiple usage alerts at once
// @Tags alerts
// @Accept json
// @Produce json
// @Param request body interfaces.BulkResolveAlertsRequest true "Bulk Resolve Request"
// @Success 200 {object} response.Response{data=interfaces.BulkResolveAlertsResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /usage-alerts/bulk-resolve [post]
func (h *UsageAlertHandler) BulkResolveAlerts(c *gin.Context) {
	var req interfaces.BulkResolveAlertsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request format")
		return
	}

	result, err := h.usageAlertService.BulkResolveAlerts(c.Request.Context(), &req)
	if err != nil {
		response.InternalServerError(c, "Failed to bulk resolve alerts", err.Error())
		return
	}

	response.SuccessWithMessage(c, "Bulk resolve operation completed", result)
}

// Alert Analytics

// GetAlertStatistics godoc
// @Summary Get alert statistics for a subscription
// @Description Retrieve detailed alert statistics and trends
// @Tags alerts
// @Accept json
// @Produce json
// @Param subscription_id path int true "Subscription ID"
// @Param usage_type query string false "Usage Type Filter"
// @Param severity query string false "Severity Filter" Enums(info,warning,error,critical)
// @Param period query string false "Time Period" Enums(24h,7d,30d,90d,365d) default(30d)
// @Param start_time query string false "Start Time (RFC3339)" format(date-time)
// @Param end_time query string false "End Time (RFC3339)" format(date-time)
// @Param group_by query string false "Group By" Enums(hour,day,week,month,severity,usage_type)
// @Success 200 {object} response.Response{data=interfaces.AlertStatisticsResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /usage-alerts/subscription/{subscription_id}/statistics [get]
func (h *UsageAlertHandler) GetAlertStatistics(c *gin.Context) {
	subscriptionID, err := strconv.ParseUint(c.Param("subscription_id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid subscription ID")
		return
	}

	req := &interfaces.AlertStatsRequest{
		UserSubscriptionID: uint(subscriptionID),
		UsageType:          c.Query("usage_type"),
		Severity:           c.Query("severity"),
		Period:             c.DefaultQuery("period", "30d"),
		GroupBy:            c.Query("group_by"),
	}

	// Parse time filters
	if startTimeStr := c.Query("start_time"); startTimeStr != "" {
		if startTime, err := time.Parse(time.RFC3339, startTimeStr); err == nil {
			req.StartTime = &startTime
		}
	}

	if endTimeStr := c.Query("end_time"); endTimeStr != "" {
		if endTime, err := time.Parse(time.RFC3339, endTimeStr); err == nil {
			req.EndTime = &endTime
		}
	}

	stats, err := h.usageAlertService.GetAlertStatistics(c.Request.Context(), req)
	if err != nil {
		response.InternalServerError(c, "Failed to get alert statistics", err.Error())
		return
	}

	response.SuccessWithMessage(c, "Alert statistics retrieved successfully", stats)
}

// GetAlertHistory godoc
// @Summary Get alert history for a subscription
// @Description Retrieve detailed alert history with optional filters
// @Tags alerts
// @Accept json
// @Produce json
// @Param subscription_id path int true "Subscription ID"
// @Param alert_configuration_id query int false "Alert Configuration ID Filter"
// @Param usage_type query string false "Usage Type Filter"
// @Param start_time query string false "Start Time (RFC3339)" format(date-time)
// @Param end_time query string false "End Time (RFC3339)" format(date-time)
// @Param include_resolved query bool false "Include resolved alerts" default(true)
// @Param include_notifications query bool false "Include notification history" default(false)
// @Param limit query int false "Limit" default(50) maximum(1000)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} response.Response{data=interfaces.AlertHistoryResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /usage-alerts/subscription/{subscription_id}/history [get]
func (h *UsageAlertHandler) GetAlertHistory(c *gin.Context) {
	subscriptionID, err := strconv.ParseUint(c.Param("subscription_id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid subscription ID")
		return
	}

	req := &interfaces.AlertHistoryRequest{
		UserSubscriptionID:   uint(subscriptionID),
		UsageType:            c.Query("usage_type"),
		IncludeResolved:      c.DefaultQuery("include_resolved", "true") == "true",
		IncludeNotifications: c.Query("include_notifications") == "true",
		Limit:                50,
		Offset:               0,
	}

	// Parse alert_configuration_id
	if configIDStr := c.Query("alert_configuration_id"); configIDStr != "" {
		if configID, err := strconv.ParseUint(configIDStr, 10, 32); err == nil {
			req.AlertConfigurationID = uint(configID)
		}
	}

	// Parse time filters
	if startTimeStr := c.Query("start_time"); startTimeStr != "" {
		if startTime, err := time.Parse(time.RFC3339, startTimeStr); err == nil {
			req.StartTime = &startTime
		}
	}

	if endTimeStr := c.Query("end_time"); endTimeStr != "" {
		if endTime, err := time.Parse(time.RFC3339, endTimeStr); err == nil {
			req.EndTime = &endTime
		}
	}

	// Parse pagination
	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			req.Limit = limit
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			req.Offset = offset
		}
	}

	history, err := h.usageAlertService.GetAlertHistory(c.Request.Context(), req)
	if err != nil {
		response.InternalServerError(c, "Failed to get alert history", err.Error())
		return
	}

	response.SuccessWithMessage(c, "Alert history retrieved successfully", history)
}

// Notification Testing

// TestNotificationChannel godoc
// @Summary Test a notification channel
// @Description Send a test notification to verify channel configuration
// @Tags alerts
// @Accept json
// @Produce json
// @Param request body interfaces.TestNotificationRequest true "Test Notification Request"
// @Success 200 {object} response.Response{data=interfaces.TestNotificationResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /usage-alerts/test-notification [post]
func (h *UsageAlertHandler) TestNotificationChannel(c *gin.Context) {
	var req interfaces.TestNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request format")
		return
	}

	result, err := h.usageAlertService.TestNotificationChannel(c.Request.Context(), &req)
	if err != nil {
		response.InternalServerError(c, "Failed to test notification channel", err.Error())
		return
	}

	if result.Success {
		response.SuccessWithMessage(c, "Test notification sent successfully", result)
	} else {
		response.Error(c, http.StatusBadRequest, 4000, "Test notification failed")
	}
}

// RegisterRoutes registers all usage alert routes
func (h *UsageAlertHandler) RegisterRoutes(router *gin.RouterGroup) {
	usageAlertGroup := router.Group("/usage-alerts")
	{
		// Alert configuration routes - specific paths first
		usageAlertGroup.POST("/configurations", h.CreateAlertConfiguration)
		usageAlertGroup.GET("/config/:config_id", h.GetAlertConfiguration)
		usageAlertGroup.PUT("/config/:config_id", h.UpdateAlertConfiguration)
		usageAlertGroup.DELETE("/config/:config_id", h.DeleteAlertConfiguration)
		
		// Alert configuration routes - by subscription (more specific path)
		usageAlertGroup.GET("/subscription/:subscription_id/configurations", h.GetAlertConfigurations)

		// Alert management routes - by subscription
		usageAlertGroup.GET("/subscription/:subscription_id", h.GetUsageAlerts)
		usageAlertGroup.GET("/subscription/:subscription_id/statistics", h.GetAlertStatistics)
		usageAlertGroup.GET("/subscription/:subscription_id/history", h.GetAlertHistory)
		
		// Alert management routes - by alert ID
		usageAlertGroup.POST("/alert/:alert_id/resolve", h.ResolveAlert)
		usageAlertGroup.POST("/alert/:alert_id/acknowledge", h.AcknowledgeAlert)
		usageAlertGroup.POST("/alert/:alert_id/suppress", h.SuppressAlert)
		
		// Bulk operations
		usageAlertGroup.POST("/bulk/resolve", h.BulkResolveAlerts)

		// Test notification
		usageAlertGroup.POST("/test-notification", h.TestNotificationChannel)
	}
}
