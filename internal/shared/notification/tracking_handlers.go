package notification

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"linke/internal/shared/handlers"
	"linke/internal/shared/logger"
	"linke/internal/shared/response"
)

// NotificationTrackingHandlers provides HTTP handlers for notification tracking
type NotificationTrackingHandlers struct {
	statusTracker *NotificationStatusTracker
	logger        logger.Logger
}

// NewNotificationTrackingHandlers creates new tracking handlers
func NewNotificationTrackingHandlers(
	statusTracker *NotificationStatusTracker,
	logger logger.Logger,
) *NotificationTrackingHandlers {
	return &NotificationTrackingHandlers{
		statusTracker: statusTracker,
		logger:        logger,
	}
}

// GetNotificationStatus godoc
// @Summary Get notification status
// @Description Get detailed status information for a specific notification
// @Tags notifications,tracking
// @Accept json
// @Produce json
// @Param request_id path string true "Notification Request ID"
// @Success 200 {object} NotificationTrackingInfo
// @Failure 400 {object} response.BadRequestResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /api/v1/notifications/{request_id}/status [get]
func (h *NotificationTrackingHandlers) GetNotificationStatus(c *gin.Context) {
	requestID := c.Param("request_id")
	if requestID == "" {
		response.BadRequest(c, "Request ID is required")
		return
	}

	trackingInfo, err := h.statusTracker.GetNotificationStatus(c.Request.Context(), requestID)
	if err != nil {
		h.logger.Error("Failed to get notification status",
			logger.String("request_id", requestID),
			logger.ErrorField(err))
		
		response.NotFound(c, "Notification not found")
		return
	}

	response.Success(c, trackingInfo)
}

// GetUserNotificationHistory godoc
// @Summary Get user notification history
// @Description Get notification history for a specific user
// @Tags notifications,tracking,users
// @Accept json
// @Produce json
// @Param user_id path int true "User ID"
// @Param limit query int false "Number of notifications to return" default(20)
// @Param event_type query string false "Filter by event type"
// @Success 200 {object} UserNotificationHistoryResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /api/v1/users/{user_id}/notifications/history [get]
func (h *NotificationTrackingHandlers) GetUserNotificationHistory(c *gin.Context) {
	userID, ok := handlers.ParseIDParam(c, "user_id")
	if !ok {
		return // ParseIDParam already handles the error response
	}

	limit := 20
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 100 {
			limit = parsedLimit
		}
	}

	eventType := c.Query("event_type")

	history, err := h.statusTracker.GetUserNotificationHistory(c.Request.Context(), userID, limit, eventType)
	if err != nil {
		h.logger.Error("Failed to get user notification history",
			logger.Uint("user_id", userID),
			logger.ErrorField(err))
		
		response.InternalServerError(c, "Failed to retrieve notification history")
		return
	}

	responseData := UserNotificationHistoryResponse{
		UserID:        userID,
		Notifications: history,
		Limit:         limit,
		EventType:     eventType,
		RetrievedAt:   time.Now(),
	}

	response.Success(c, responseData)
}

// GetNotificationAnalytics godoc
// @Summary Get notification analytics
// @Description Get analytics and statistics for notifications within a time range
// @Tags notifications,tracking,analytics
// @Accept json
// @Produce json
// @Param start_date query string false "Start date (RFC3339 format)" default("24h ago")
// @Param end_date query string false "End date (RFC3339 format)" default("now")
// @Param user_ids query string false "Comma-separated user IDs to filter by"
// @Param event_types query string false "Comma-separated event types to filter by"
// @Param channels query string false "Comma-separated channels to filter by"
// @Param statuses query string false "Comma-separated statuses to filter by"
// @Success 200 {object} NotificationAnalytics
// @Failure 400 {object} response.BadRequestResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /api/v1/notifications/analytics [get]
func (h *NotificationTrackingHandlers) GetNotificationAnalytics(c *gin.Context) {
	// Parse time range
	timeRange, err := h.parseTimeRange(c)
	if err != nil {
		response.BadRequest(c, "Invalid time range: "+err.Error())
		return
	}

	// Parse filters
	filters, err := h.parseAnalyticsFilters(c)
	if err != nil {
		response.BadRequest(c, "Invalid filters: "+err.Error())
		return
	}

	analytics, err := h.statusTracker.GetNotificationAnalytics(c.Request.Context(), timeRange, filters)
	if err != nil {
		h.logger.Error("Failed to get notification analytics",
			logger.String("time_range", timeRange.String()),
			logger.ErrorField(err))
		
		response.InternalServerError(c, "Failed to retrieve analytics")
		return
	}

	response.Success(c, analytics)
}

// GetNotificationStatusSummary godoc
// @Summary Get notification status summary
// @Description Get a summary of notification statuses for the current user
// @Tags notifications,tracking
// @Accept json
// @Produce json
// @Success 200 {object} NotificationStatusSummary
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /api/v1/notifications/status/summary [get]
func (h *NotificationTrackingHandlers) GetNotificationStatusSummary(c *gin.Context) {
	// In a real implementation, you would get the current user from JWT token
	// For now, we'll return a placeholder summary
	
	summary := NotificationStatusSummary{
		TotalNotifications:   0,
		PendingNotifications: 0,
		SentNotifications:    0,
		FailedNotifications:  0,
		RetryingNotifications: 0,
		LastUpdated:          time.Now(),
		RecentActivity:       []RecentNotificationActivity{},
	}

	response.Success(c, summary)
}

// Helper methods

func (h *NotificationTrackingHandlers) parseTimeRange(c *gin.Context) (TimeRange, error) {
	now := time.Now()
	defaultStart := now.Add(-24 * time.Hour)
	
	var timeRange TimeRange
	var err error

	// Parse start date
	if startStr := c.Query("start_date"); startStr != "" {
		timeRange.Start, err = time.Parse(time.RFC3339, startStr)
		if err != nil {
			return timeRange, err
		}
	} else {
		timeRange.Start = defaultStart
	}

	// Parse end date
	if endStr := c.Query("end_date"); endStr != "" {
		timeRange.End, err = time.Parse(time.RFC3339, endStr)
		if err != nil {
			return timeRange, err
		}
	} else {
		timeRange.End = now
	}

	// Validate range
	if timeRange.End.Before(timeRange.Start) {
		return timeRange, fmt.Errorf("end date must be after start date")
	}

	return timeRange, nil
}

func (h *NotificationTrackingHandlers) parseAnalyticsFilters(c *gin.Context) (AnalyticsFilters, error) {
	var filters AnalyticsFilters

	// Parse user IDs
	if userIDsStr := c.Query("user_ids"); userIDsStr != "" {
		userIDStrs := strings.Split(userIDsStr, ",")
		for _, userIDStr := range userIDStrs {
			userID, err := strconv.ParseUint(strings.TrimSpace(userIDStr), 10, 32)
			if err != nil {
				return filters, fmt.Errorf("invalid user ID: %s", userIDStr)
			}
			filters.UserIDs = append(filters.UserIDs, uint(userID))
		}
	}

	// Parse event types
	if eventTypesStr := c.Query("event_types"); eventTypesStr != "" {
		eventTypes := strings.Split(eventTypesStr, ",")
		for i, eventType := range eventTypes {
			eventTypes[i] = strings.TrimSpace(eventType)
		}
		filters.EventTypes = eventTypes
	}

	// Parse channels
	if channelsStr := c.Query("channels"); channelsStr != "" {
		channelStrs := strings.Split(channelsStr, ",")
		for _, channelStr := range channelStrs {
			channel := NotificationChannel(strings.TrimSpace(channelStr))
			filters.Channels = append(filters.Channels, channel)
		}
	}

	// Parse statuses
	if statusesStr := c.Query("statuses"); statusesStr != "" {
		statusStrs := strings.Split(statusesStr, ",")
		for _, statusStr := range statusStrs {
			status := h.parseNotificationStatus(strings.TrimSpace(statusStr))
			if status != -1 {
				filters.Statuses = append(filters.Statuses, status)
			}
		}
	}

	// Parse priorities
	if prioritiesStr := c.Query("priorities"); prioritiesStr != "" {
		priorityStrs := strings.Split(prioritiesStr, ",")
		for _, priorityStr := range priorityStrs {
			priority := h.parseNotificationPriority(strings.TrimSpace(priorityStr))
			if priority != -1 {
				filters.Priorities = append(filters.Priorities, priority)
			}
		}
	}

	// Parse tags
	if tagsStr := c.Query("tags"); tagsStr != "" {
		tags := strings.Split(tagsStr, ",")
		for i, tag := range tags {
			tags[i] = strings.TrimSpace(tag)
		}
		filters.Tags = tags
	}

	return filters, nil
}

func (h *NotificationTrackingHandlers) parseNotificationStatus(statusStr string) NotificationStatus {
	switch strings.ToLower(statusStr) {
	case "pending":
		return StatusPending
	case "sending":
		return StatusSending
	case "sent":
		return StatusSent
	case "delivered":
		return StatusDelivered
	case "failed":
		return StatusFailed
	case "retrying":
		return StatusRetrying
	case "expired":
		return StatusExpired
	case "cancelled":
		return StatusCancelled
	default:
		return -1
	}
}

func (h *NotificationTrackingHandlers) parseNotificationPriority(priorityStr string) NotificationPriority {
	switch strings.ToLower(priorityStr) {
	case "low":
		return PriorityLow
	case "normal":
		return PriorityNormal
	case "high":
		return PriorityHigh
	case "urgent":
		return PriorityUrgent
	default:
		return -1
	}
}

// Response types

// UserNotificationHistoryResponse represents the response for user notification history
type UserNotificationHistoryResponse struct {
	UserID        uint                        `json:"user_id"`
	Notifications []*NotificationTrackingInfo `json:"notifications"`
	Limit         int                         `json:"limit"`
	EventType     string                      `json:"event_type,omitempty"`
	RetrievedAt   time.Time                   `json:"retrieved_at"`
}

// NotificationStatusSummary provides a summary of notification statuses
type NotificationStatusSummary struct {
	TotalNotifications    int                           `json:"total_notifications"`
	PendingNotifications  int                           `json:"pending_notifications"`
	SentNotifications     int                           `json:"sent_notifications"`
	FailedNotifications   int                           `json:"failed_notifications"`
	RetryingNotifications int                           `json:"retrying_notifications"`
	LastUpdated           time.Time                     `json:"last_updated"`
	RecentActivity        []RecentNotificationActivity  `json:"recent_activity"`
}

// RecentNotificationActivity represents recent notification activity
type RecentNotificationActivity struct {
	RequestID   string                 `json:"request_id"`
	EventType   string                 `json:"event_type"`
	Status      NotificationStatus     `json:"status"`
	Channel     NotificationChannel    `json:"channel"`
	Timestamp   time.Time              `json:"timestamp"`
	Description string                 `json:"description"`
}