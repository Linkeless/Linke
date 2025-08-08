package notification

import (
	"context"
	"fmt"
	"time"

	"linke/internal/shared/logger"
)

// TrackingNotificationService wraps a notification service with comprehensive status tracking
type TrackingNotificationService struct {
	baseService    NotificationService
	statusTracker  *NotificationStatusTracker
	logger         logger.Logger
	enableTracking bool
}

// NewTrackingNotificationService creates a new tracking notification service
func NewTrackingNotificationService(
	baseService NotificationService,
	statusTracker *NotificationStatusTracker,
	logger logger.Logger,
) *TrackingNotificationService {
	return &TrackingNotificationService{
		baseService:    baseService,
		statusTracker:  statusTracker,
		logger:         logger,
		enableTracking: true,
	}
}

// SetTrackingEnabled enables or disables tracking
func (s *TrackingNotificationService) SetTrackingEnabled(enabled bool) {
	s.enableTracking = enabled
}

// Send sends notifications with comprehensive tracking
func (s *TrackingNotificationService) Send(ctx context.Context, req *NotificationRequest) ([]*NotificationResult, error) {
	requestID := s.generateRequestID(req)

	// Initialize tracking if enabled
	if s.enableTracking && s.statusTracker != nil {
		if err := s.statusTracker.InitializeTracking(ctx, req, requestID); err != nil {
			s.logger.Error("Failed to initialize tracking", 
				logger.String("request_id", requestID),
				logger.ErrorField(err))
		}
	}

	s.logger.Info("Sending notification with tracking",
		logger.String("request_id", requestID),
		logger.Uint("user_id", req.UserID),
		logger.String("event_type", req.EventType),
		logger.Int("channels_count", len(req.Channels)))

	// Update status to sending
	if s.enableTracking {
		s.updateStatus(ctx, requestID, StatusSending, nil)
	}

	// Process each channel individually for detailed tracking
	var results []*NotificationResult
	
	for _, channel := range req.Channels {
		channelReq := &NotificationRequest{
			UserID:    req.UserID,
			EventType: req.EventType,
			EventID:   req.EventID,
			Channels:  []NotificationChannel{channel},
			Priority:  req.Priority,
			Data:      req.Data,
			Metadata:  req.Metadata,
			Tags:      req.Tags,
		}

		result := s.processChannelWithTracking(ctx, channelReq, channel, requestID)
		results = append(results, result)
	}

	// Update overall status based on results
	if s.enableTracking {
		s.updateOverallStatusFromResults(ctx, requestID, results)
	}

	return results, nil
}

// processChannelWithTracking processes a single channel with detailed tracking
func (s *TrackingNotificationService) processChannelWithTracking(
	ctx context.Context,
	req *NotificationRequest,
	channel NotificationChannel,
	requestID string,
) *NotificationResult {
	attemptNumber := 1 // This would be tracked properly in retry scenarios
	
	// Create delivery attempt record
	attempt := &DeliveryAttempt{
		AttemptNumber: attemptNumber,
		Channel:       channel,
		Status:        StatusSending,
		StartedAt:     time.Now(),
	}

	// Record the start of the attempt
	if s.enableTracking {
		s.recordAttempt(ctx, requestID, channel, attempt)
	}

	// Send using base service
	results, err := s.baseService.Send(ctx, req)
	
	// Complete the attempt record
	completedAt := time.Now()
	attempt.CompletedAt = &completedAt
	attempt.Duration = completedAt.Sub(attempt.StartedAt)

	var result *NotificationResult
	if err != nil {
		// Service-level error
		attempt.Status = StatusFailed
		attempt.Error = err.Error()
		
		result = &NotificationResult{
			Channel:   channel,
			Success:   false,
			Error:     err.Error(),
			Timestamp: time.Now(),
			Metadata: map[string]interface{}{
				"request_id": requestID,
				"error_type": "service_error",
			},
		}
	} else if len(results) == 0 {
		// No results returned
		attempt.Status = StatusFailed
		attempt.Error = "no results returned from provider"
		
		result = &NotificationResult{
			Channel:   channel,
			Success:   false,
			Error:     "no results returned from provider",
			Timestamp: time.Now(),
			Metadata: map[string]interface{}{
				"request_id": requestID,
			},
		}
	} else {
		// Use the first (and should be only) result
		result = results[0]
		
		if result.Success {
			attempt.Status = StatusSent
			
			// Check for delivery confirmation in metadata
			if deliveryInfo, ok := result.Metadata["delivery_info"]; ok {
				attempt.Status = StatusDelivered
				if deliveryMap, ok := deliveryInfo.(map[string]interface{}); ok {
					attempt.Metadata = deliveryMap
				}
			}
		} else {
			attempt.Status = StatusFailed
			attempt.Error = result.Error
		}

		// Extract provider response information
		if result.Metadata != nil {
			providerResp := &ProviderResponse{}
			
			if providerID, ok := result.Metadata["provider_id"].(string); ok {
				providerResp.ProviderID = providerID
			}
			if messageID, ok := result.Metadata["message_id"].(string); ok {
				providerResp.MessageID = messageID
			}
			if statusCode, ok := result.Metadata["status_code"].(int); ok {
				providerResp.StatusCode = statusCode
			}
			if responseBody, ok := result.Metadata["response_body"].(string); ok {
				providerResp.ResponseBody = responseBody
			}
			if headers, ok := result.Metadata["headers"].(map[string]string); ok {
				providerResp.Headers = headers
			}
			if deliveryInfo, ok := result.Metadata["delivery_info"].(map[string]interface{}); ok {
				providerResp.DeliveryInfo = deliveryInfo
			}

			attempt.ProviderResp = providerResp
		}

		// Update result with tracking information
		if result.Metadata == nil {
			result.Metadata = make(map[string]interface{})
		}
		result.Metadata["request_id"] = requestID
		result.Metadata["attempt_number"] = attemptNumber
		result.Metadata["tracking_enabled"] = s.enableTracking
	}

	// Record the completed attempt
	if s.enableTracking {
		s.recordAttempt(ctx, requestID, channel, attempt)
	}

	return result
}

// SendEmail sends an email notification with tracking
func (s *TrackingNotificationService) SendEmail(ctx context.Context, to, subject, body string) error {
	return s.baseService.SendEmail(ctx, to, subject, body)
}

// SendSMS sends an SMS notification with tracking
func (s *TrackingNotificationService) SendSMS(ctx context.Context, to, message string) error {
	return s.baseService.SendSMS(ctx, to, message)
}

// SendPush sends a push notification with tracking
func (s *TrackingNotificationService) SendPush(ctx context.Context, userID uint, title, body string) error {
	return s.baseService.SendPush(ctx, userID, title, body)
}

// SendTelegram sends a Telegram notification with tracking
func (s *TrackingNotificationService) SendTelegram(ctx context.Context, chatID int64, message string) error {
	return s.baseService.SendTelegram(ctx, chatID, message)
}

// SendTelegramByUsername sends a Telegram notification by username with tracking
func (s *TrackingNotificationService) SendTelegramByUsername(ctx context.Context, username, message string) error {
	return s.baseService.SendTelegramByUsername(ctx, username, message)
}

// GetTemplate gets a notification template
func (s *TrackingNotificationService) GetTemplate(templateName string) (string, error) {
	return s.baseService.GetTemplate(templateName)
}

// HealthCheck performs health check on the base service
func (s *TrackingNotificationService) HealthCheck(ctx context.Context) error {
	return s.baseService.HealthCheck(ctx)
}

// GetNotificationStatus retrieves the status of a notification
func (s *TrackingNotificationService) GetNotificationStatus(
	ctx context.Context,
	requestID string,
) (*NotificationTrackingInfo, error) {
	if !s.enableTracking || s.statusTracker == nil {
		return nil, fmt.Errorf("tracking is not enabled")
	}

	return s.statusTracker.GetNotificationStatus(ctx, requestID)
}

// GetUserNotificationHistory retrieves notification history for a user
func (s *TrackingNotificationService) GetUserNotificationHistory(
	ctx context.Context,
	userID uint,
	limit int,
	eventType string,
) ([]*NotificationTrackingInfo, error) {
	if !s.enableTracking || s.statusTracker == nil {
		return nil, fmt.Errorf("tracking is not enabled")
	}

	return s.statusTracker.GetUserNotificationHistory(ctx, userID, limit, eventType)
}

// GetNotificationAnalytics provides analytics for notifications
func (s *TrackingNotificationService) GetNotificationAnalytics(
	ctx context.Context,
	timeRange TimeRange,
	filters AnalyticsFilters,
) (*NotificationAnalytics, error) {
	if !s.enableTracking || s.statusTracker == nil {
		return nil, fmt.Errorf("tracking is not enabled")
	}

	return s.statusTracker.GetNotificationAnalytics(ctx, timeRange, filters)
}

// Helper methods

func (s *TrackingNotificationService) generateRequestID(req *NotificationRequest) string {
	return fmt.Sprintf("track_%d_%s_%d", req.UserID, req.EventType, time.Now().UnixNano())
}

func (s *TrackingNotificationService) recordAttempt(
	ctx context.Context,
	requestID string,
	channel NotificationChannel,
	attempt *DeliveryAttempt,
) {
	if err := s.statusTracker.RecordDeliveryAttempt(ctx, requestID, channel, attempt); err != nil {
		s.logger.Error("Failed to record delivery attempt",
			logger.String("request_id", requestID),
			logger.String("channel", string(channel)),
			logger.ErrorField(err))
	}
}

func (s *TrackingNotificationService) updateStatus(
	ctx context.Context,
	requestID string,
	status NotificationStatus,
	metadata map[string]interface{},
) {
	if err := s.statusTracker.UpdateNotificationStatus(ctx, requestID, status, metadata); err != nil {
		s.logger.Error("Failed to update notification status",
			logger.String("request_id", requestID),
			logger.String("status", status.String()),
			logger.ErrorField(err))
	}
}

func (s *TrackingNotificationService) updateOverallStatusFromResults(
	ctx context.Context,
	requestID string,
	results []*NotificationResult,
) {
	successCount := 0
	failureCount := 0

	for _, result := range results {
		if result.Success {
			successCount++
		} else {
			failureCount++
		}
	}

	var status NotificationStatus
	metadata := map[string]interface{}{
		"total_channels":      len(results),
		"successful_channels": successCount,
		"failed_channels":     failureCount,
		"completion_time":     time.Now().Format(time.RFC3339),
	}

	if successCount == len(results) {
		status = StatusDelivered
	} else if failureCount == len(results) {
		status = StatusFailed
	} else {
		status = StatusSent // Partial success
		metadata["partial_success"] = true
	}

	s.updateStatus(ctx, requestID, status, metadata)
}

// TrackingRetryableNotificationService extends the retryable service with tracking
type TrackingRetryableNotificationService struct {
	*RetryableNotificationService
	statusTracker *NotificationStatusTracker
	logger        logger.Logger
}

// NewTrackingRetryableNotificationService creates a new tracking retryable service
func NewTrackingRetryableNotificationService(
	baseService NotificationService,
	retryableService *RetryableNotificationService,
	statusTracker *NotificationStatusTracker,
	logger logger.Logger,
) *TrackingRetryableNotificationService {
	return &TrackingRetryableNotificationService{
		RetryableNotificationService: retryableService,
		statusTracker:                statusTracker,
		logger:                       logger,
	}
}

// Send overrides the retryable service send method to add tracking
func (s *TrackingRetryableNotificationService) Send(ctx context.Context, req *NotificationRequest) ([]*NotificationResult, error) {
	requestID := s.generateRequestID(req)

	// Initialize tracking
	if err := s.statusTracker.InitializeTracking(ctx, req, requestID); err != nil {
		s.logger.Error("Failed to initialize retry tracking", 
			logger.String("request_id", requestID),
			logger.ErrorField(err))
	}

	// Call the retryable service
	results, err := s.RetryableNotificationService.Send(ctx, req)

	// Update tracking based on retry results
	s.updateTrackingFromRetryResults(ctx, requestID, results, err)

	return results, err
}

func (s *TrackingRetryableNotificationService) updateTrackingFromRetryResults(
	ctx context.Context,
	requestID string,
	results []*NotificationResult,
	err error,
) {
	if err != nil {
		// Service-level error
		s.updateStatus(ctx, requestID, StatusFailed, map[string]interface{}{
			"service_error": err.Error(),
			"retry_enabled": true,
		})
		return
	}

	// Process individual channel results
	retryingCount := 0
	successCount := 0
	failedCount := 0

	for _, result := range results {
		if result.Success {
			successCount++
		} else if result.Error == "scheduled for retry" {
			retryingCount++
		} else {
			failedCount++
		}
	}

	var status NotificationStatus
	metadata := map[string]interface{}{
		"retry_enabled":       true,
		"successful_channels": successCount,
		"retrying_channels":   retryingCount,
		"failed_channels":     failedCount,
	}

	if retryingCount > 0 {
		status = StatusRetrying
	} else if successCount > 0 && failedCount == 0 {
		status = StatusDelivered
	} else if successCount > 0 {
		status = StatusSent // Partial success
	} else {
		status = StatusFailed
	}

	s.updateStatus(ctx, requestID, status, metadata)
}

func (s *TrackingRetryableNotificationService) updateStatus(
	ctx context.Context,
	requestID string,
	status NotificationStatus,
	metadata map[string]interface{},
) {
	if err := s.statusTracker.UpdateNotificationStatus(ctx, requestID, status, metadata); err != nil {
		s.logger.Error("Failed to update retry tracking status",
			logger.String("request_id", requestID),
			logger.String("status", status.String()),
			logger.ErrorField(err))
	}
}