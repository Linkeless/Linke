package notification

import (
	"context"
	"fmt"
	"sync"
	"time"

	"linke/internal/shared/cache"
	"linke/internal/shared/logger"
)

// NotificationStatus represents the current status of a notification
type NotificationStatus int

const (
	StatusPending NotificationStatus = iota
	StatusSending
	StatusSent
	StatusDelivered
	StatusFailed
	StatusRetrying
	StatusExpired
	StatusCancelled
)

func (s NotificationStatus) String() string {
	switch s {
	case StatusPending:
		return "pending"
	case StatusSending:
		return "sending"
	case StatusSent:
		return "sent"
	case StatusDelivered:
		return "delivered"
	case StatusFailed:
		return "failed"
	case StatusRetrying:
		return "retrying"
	case StatusExpired:
		return "expired"
	case StatusCancelled:
		return "cancelled"
	default:
		return "unknown"
	}
}

// DeliveryAttempt represents a single delivery attempt
type DeliveryAttempt struct {
	AttemptNumber int                    `json:"attempt_number"`
	Channel       NotificationChannel    `json:"channel"`
	Status        NotificationStatus     `json:"status"`
	Error         string                 `json:"error,omitempty"`
	StartedAt     time.Time              `json:"started_at"`
	CompletedAt   *time.Time             `json:"completed_at,omitempty"`
	Duration      time.Duration          `json:"duration" swaggertype:"string"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	ProviderResp  *ProviderResponse      `json:"provider_response,omitempty"`
}

// ProviderResponse contains response details from notification providers
type ProviderResponse struct {
	ProviderID   string                 `json:"provider_id"`
	MessageID    string                 `json:"message_id,omitempty"`
	StatusCode   int                    `json:"status_code,omitempty"`
	ResponseBody string                 `json:"response_body,omitempty"`
	Headers      map[string]string      `json:"headers,omitempty"`
	DeliveryInfo map[string]interface{} `json:"delivery_info,omitempty"`
}

// NotificationTrackingInfo contains complete tracking information for a notification
type NotificationTrackingInfo struct {
	RequestID          string                                     `json:"request_id"`
	UserID             uint                                       `json:"user_id"`
	EventType          string                                     `json:"event_type"`
	EventID            string                                     `json:"event_id,omitempty"`
	OverallStatus      NotificationStatus                         `json:"overall_status"`
	CreatedAt          time.Time                                  `json:"created_at"`
	UpdatedAt          time.Time                                  `json:"updated_at"`
	ExpiresAt          *time.Time                                 `json:"expires_at,omitempty"`
	Priority           NotificationPriority                       `json:"priority"`
	ChannelAttempts    map[NotificationChannel][]*DeliveryAttempt `json:"channel_attempts"`
	TotalAttempts      int                                        `json:"total_attempts"`
	SuccessfulChannels []NotificationChannel                      `json:"successful_channels"`
	FailedChannels     []NotificationChannel                      `json:"failed_channels"`
	RetryingChannels   []NotificationChannel                      `json:"retrying_channels"`
	Metadata           map[string]interface{}                     `json:"metadata,omitempty"`
	Tags               []string                                   `json:"tags,omitempty"`
}

// NotificationStatusTracker provides comprehensive tracking for notifications
type NotificationStatusTracker struct {
	cache      cache.CacheStore
	logger     logger.Logger
	mu         sync.RWMutex
	keyPrefix  string
	defaultTTL time.Duration
}

// StatusTrackerConfig contains configuration for the status tracker
type StatusTrackerConfig struct {
	KeyPrefix             string
	DefaultTTL            time.Duration
	EnableRealTimeUpdates bool
	BatchUpdateSize       int
	CleanupInterval       time.Duration
}

// DefaultStatusTrackerConfig returns default configuration
func DefaultStatusTrackerConfig() *StatusTrackerConfig {
	return &StatusTrackerConfig{
		KeyPrefix:             "notification_status:",
		DefaultTTL:            7 * 24 * time.Hour, // Keep for 7 days
		EnableRealTimeUpdates: true,
		BatchUpdateSize:       100,
		CleanupInterval:       1 * time.Hour,
	}
}

// NewNotificationStatusTracker creates a new status tracker
func NewNotificationStatusTracker(
	cache cache.CacheStore,
	logger logger.Logger,
	config *StatusTrackerConfig,
) *NotificationStatusTracker {
	if config == nil {
		config = DefaultStatusTrackerConfig()
	}

	return &NotificationStatusTracker{
		cache:      cache,
		logger:     logger,
		keyPrefix:  config.KeyPrefix,
		defaultTTL: config.DefaultTTL,
	}
}

// InitializeTracking initializes tracking for a new notification request
func (t *NotificationStatusTracker) InitializeTracking(
	ctx context.Context,
	req *NotificationRequest,
	requestID string,
) error {
	trackingInfo := &NotificationTrackingInfo{
		RequestID:          requestID,
		UserID:             req.UserID,
		EventType:          req.EventType,
		EventID:            req.EventID,
		OverallStatus:      StatusPending,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
		Priority:           req.Priority,
		ChannelAttempts:    make(map[NotificationChannel][]*DeliveryAttempt),
		TotalAttempts:      0,
		SuccessfulChannels: make([]NotificationChannel, 0),
		FailedChannels:     make([]NotificationChannel, 0),
		RetryingChannels:   make([]NotificationChannel, 0),
		Metadata:           req.Metadata,
		Tags:               req.Tags,
	}

	// Set expiration time based on priority
	if req.Priority == PriorityHigh {
		expiresAt := time.Now().Add(24 * time.Hour)
		trackingInfo.ExpiresAt = &expiresAt
	} else if req.Priority == PriorityNormal {
		expiresAt := time.Now().Add(72 * time.Hour)
		trackingInfo.ExpiresAt = &expiresAt
	}

	// Initialize channel attempts
	for _, channel := range req.Channels {
		trackingInfo.ChannelAttempts[channel] = make([]*DeliveryAttempt, 0)
	}

	return t.storeTrackingInfo(ctx, trackingInfo)
}

// RecordDeliveryAttempt records a delivery attempt for a specific channel
func (t *NotificationStatusTracker) RecordDeliveryAttempt(
	ctx context.Context,
	requestID string,
	channel NotificationChannel,
	attempt *DeliveryAttempt,
) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	trackingInfo, err := t.getTrackingInfo(ctx, requestID)
	if err != nil {
		return fmt.Errorf("failed to get tracking info: %w", err)
	}

	// Add the attempt to the channel's attempt list
	if trackingInfo.ChannelAttempts[channel] == nil {
		trackingInfo.ChannelAttempts[channel] = make([]*DeliveryAttempt, 0)
	}

	trackingInfo.ChannelAttempts[channel] = append(
		trackingInfo.ChannelAttempts[channel],
		attempt,
	)

	trackingInfo.TotalAttempts++
	trackingInfo.UpdatedAt = time.Now()

	// Update channel status lists
	t.updateChannelStatusLists(trackingInfo, channel, attempt.Status)

	// Update overall status
	t.updateOverallStatus(trackingInfo)

	t.logger.Debug("Recorded delivery attempt",
		logger.String("request_id", requestID),
		logger.String("channel", string(channel)),
		logger.String("status", attempt.Status.String()),
		logger.Int("attempt_number", attempt.AttemptNumber),
		logger.String("duration", attempt.Duration.String()))

	return t.storeTrackingInfo(ctx, trackingInfo)
}

// UpdateNotificationStatus updates the overall status of a notification
func (t *NotificationStatusTracker) UpdateNotificationStatus(
	ctx context.Context,
	requestID string,
	status NotificationStatus,
	metadata map[string]interface{},
) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	trackingInfo, err := t.getTrackingInfo(ctx, requestID)
	if err != nil {
		return fmt.Errorf("failed to get tracking info: %w", err)
	}

	trackingInfo.OverallStatus = status
	trackingInfo.UpdatedAt = time.Now()

	// Merge metadata
	if trackingInfo.Metadata == nil {
		trackingInfo.Metadata = make(map[string]interface{})
	}
	for k, v := range metadata {
		trackingInfo.Metadata[k] = v
	}

	t.logger.Info("Updated notification status",
		logger.String("request_id", requestID),
		logger.String("status", status.String()))

	return t.storeTrackingInfo(ctx, trackingInfo)
}

// GetNotificationStatus retrieves the current status of a notification
func (t *NotificationStatusTracker) GetNotificationStatus(
	ctx context.Context,
	requestID string,
) (*NotificationTrackingInfo, error) {
	return t.getTrackingInfo(ctx, requestID)
}

// GetUserNotificationHistory retrieves notification history for a user
func (t *NotificationStatusTracker) GetUserNotificationHistory(
	ctx context.Context,
	userID uint,
	limit int,
	eventType string,
) ([]*NotificationTrackingInfo, error) {
	// This would require a separate index in a production system
	// For now, we'll return a placeholder implementation
	t.logger.Debug("Getting user notification history",
		logger.Uint("user_id", userID),
		logger.Int("limit", limit),
		logger.String("event_type", eventType))

	// In a real implementation, you would:
	// 1. Maintain a separate index of notifications by user
	// 2. Use a database or search engine for complex queries
	// 3. Implement pagination and filtering

	return []*NotificationTrackingInfo{}, nil
}

// GetNotificationAnalytics provides analytics data for notifications
func (t *NotificationStatusTracker) GetNotificationAnalytics(
	ctx context.Context,
	timeRange TimeRange,
	filters AnalyticsFilters,
) (*NotificationAnalytics, error) {
	// Placeholder implementation - in production, this would aggregate data
	// from the tracking store and provide comprehensive analytics

	t.logger.Debug("Getting notification analytics",
		logger.String("time_range", timeRange.String()),
		logger.String("filters", fmt.Sprintf("%+v", filters)))

	analytics := &NotificationAnalytics{
		TimeRange:          timeRange,
		TotalNotifications: 0,
		SuccessRate:        0.0,
		ChannelStats:       make(map[NotificationChannel]*ChannelAnalytics),
		EventTypeStats:     make(map[string]*EventTypeAnalytics),
		GeneratedAt:        time.Now(),
	}

	return analytics, nil
}

// CleanupExpiredTracking removes expired tracking information
func (t *NotificationStatusTracker) CleanupExpiredTracking(ctx context.Context) error {
	// In a real implementation, this would:
	// 1. Scan for expired tracking entries
	// 2. Archive important data before deletion
	// 3. Clean up in batches to avoid performance impact

	t.logger.Info("Cleaning up expired notification tracking data")

	// This is a simplified implementation
	// Real cleanup would need pattern matching or separate indexes

	return nil
}

// Internal helper methods

func (t *NotificationStatusTracker) getTrackingInfo(
	ctx context.Context,
	requestID string,
) (*NotificationTrackingInfo, error) {
	key := t.keyPrefix + requestID

	var trackingInfo NotificationTrackingInfo
	err := t.cache.GetJSON(ctx, key, &trackingInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to get tracking info from cache: %w", err)
	}

	return &trackingInfo, nil
}

func (t *NotificationStatusTracker) storeTrackingInfo(
	ctx context.Context,
	trackingInfo *NotificationTrackingInfo,
) error {
	key := t.keyPrefix + trackingInfo.RequestID

	err := t.cache.SetJSON(ctx, key, trackingInfo, t.defaultTTL)
	if err != nil {
		return fmt.Errorf("failed to store tracking info: %w", err)
	}

	return nil
}

func (t *NotificationStatusTracker) updateChannelStatusLists(
	trackingInfo *NotificationTrackingInfo,
	channel NotificationChannel,
	status NotificationStatus,
) {
	// Remove from all lists first
	trackingInfo.SuccessfulChannels = t.removeChannelFromList(trackingInfo.SuccessfulChannels, channel)
	trackingInfo.FailedChannels = t.removeChannelFromList(trackingInfo.FailedChannels, channel)
	trackingInfo.RetryingChannels = t.removeChannelFromList(trackingInfo.RetryingChannels, channel)

	// Add to appropriate list
	switch status {
	case StatusSent, StatusDelivered:
		trackingInfo.SuccessfulChannels = append(trackingInfo.SuccessfulChannels, channel)
	case StatusFailed, StatusExpired:
		trackingInfo.FailedChannels = append(trackingInfo.FailedChannels, channel)
	case StatusRetrying, StatusPending:
		trackingInfo.RetryingChannels = append(trackingInfo.RetryingChannels, channel)
	}
}

func (t *NotificationStatusTracker) removeChannelFromList(
	channels []NotificationChannel,
	channel NotificationChannel,
) []NotificationChannel {
	for i, ch := range channels {
		if ch == channel {
			return append(channels[:i], channels[i+1:]...)
		}
	}
	return channels
}

func (t *NotificationStatusTracker) updateOverallStatus(trackingInfo *NotificationTrackingInfo) {
	totalChannels := len(trackingInfo.ChannelAttempts)
	successCount := len(trackingInfo.SuccessfulChannels)
	failedCount := len(trackingInfo.FailedChannels)
	retryingCount := len(trackingInfo.RetryingChannels)

	// Determine overall status based on channel states
	if successCount == totalChannels {
		trackingInfo.OverallStatus = StatusDelivered
	} else if failedCount == totalChannels {
		trackingInfo.OverallStatus = StatusFailed
	} else if retryingCount > 0 {
		trackingInfo.OverallStatus = StatusRetrying
	} else if successCount > 0 {
		trackingInfo.OverallStatus = StatusSent // Partial success
	} else {
		trackingInfo.OverallStatus = StatusPending
	}
}

// Analytics types and structures

// TimeRange represents a time range for analytics
type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

func (tr TimeRange) String() string {
	return fmt.Sprintf("%s to %s", tr.Start.Format(time.RFC3339), tr.End.Format(time.RFC3339))
}

// AnalyticsFilters contains filters for analytics queries
type AnalyticsFilters struct {
	UserIDs    []uint                 `json:"user_ids,omitempty"`
	EventTypes []string               `json:"event_types,omitempty"`
	Channels   []NotificationChannel  `json:"channels,omitempty"`
	Statuses   []NotificationStatus   `json:"statuses,omitempty"`
	Priorities []NotificationPriority `json:"priorities,omitempty"`
	Tags       []string               `json:"tags,omitempty"`
}

// NotificationAnalytics contains analytics data
type NotificationAnalytics struct {
	TimeRange           TimeRange                                 `json:"time_range"`
	TotalNotifications  int                                       `json:"total_notifications"`
	SuccessRate         float64                                   `json:"success_rate"`
	ChannelStats        map[NotificationChannel]*ChannelAnalytics `json:"channel_stats"`
	EventTypeStats      map[string]*EventTypeAnalytics            `json:"event_type_stats"`
	HourlyDistribution  []HourlyStats                             `json:"hourly_distribution"`
	TopFailureReasons   []FailureReason                           `json:"top_failure_reasons"`
	AverageDeliveryTime time.Duration                             `json:"average_delivery_time" swaggertype:"string"`
	GeneratedAt         time.Time                                 `json:"generated_at"`
}

// ChannelAnalytics contains analytics for a specific channel
type ChannelAnalytics struct {
	Channel             NotificationChannel `json:"channel"`
	TotalSent           int                 `json:"total_sent"`
	Successful          int                 `json:"successful"`
	Failed              int                 `json:"failed"`
	SuccessRate         float64             `json:"success_rate"`
	AverageDeliveryTime time.Duration       `json:"average_delivery_time" swaggertype:"string"`
	TopFailureReasons   []FailureReason     `json:"top_failure_reasons"`
}

// EventTypeAnalytics contains analytics for a specific event type
type EventTypeAnalytics struct {
	EventType           string                `json:"event_type"`
	TotalNotifications  int                   `json:"total_notifications"`
	SuccessRate         float64               `json:"success_rate"`
	AverageDeliveryTime time.Duration         `json:"average_delivery_time" swaggertype:"string"`
	PreferredChannels   []NotificationChannel `json:"preferred_channels"`
}

// HourlyStats represents notification stats for an hour
type HourlyStats struct {
	Hour        int     `json:"hour"`
	Count       int     `json:"count"`
	SuccessRate float64 `json:"success_rate"`
}

// FailureReason represents a failure reason with count
type FailureReason struct {
	Reason string `json:"reason"`
	Count  int    `json:"count"`
}
