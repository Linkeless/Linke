package notification

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"linke/internal/shared/logger"
	"linke/internal/shared/queue"
)

// RetryConfig defines configuration for retry mechanisms
type RetryConfig struct {
	MaxRetries        int           `json:"max_retries"`
	InitialDelay      time.Duration `json:"initial_delay" swaggertype:"string"`
	MaxDelay          time.Duration `json:"max_delay" swaggertype:"string"`
	BackoffMultiplier float64       `json:"backoff_multiplier"`
	EnableJitter      bool          `json:"enable_jitter"`
}

// DefaultRetryConfig returns the default retry configuration
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxRetries:        3,
		InitialDelay:      1 * time.Second,
		MaxDelay:          30 * time.Second,
		BackoffMultiplier: 2.0,
		EnableJitter:      true,
	}
}

// FailureType represents the type of failure
type FailureType int

const (
	FailureTypeTemporary FailureType = iota // Retryable failures (network issues, rate limits)
	FailureTypePermanent                    // Non-retryable failures (invalid email, auth failures)
	FailureTypeThrottled                    // Rate limited - requires longer backoff
)

// NotificationFailure represents a notification failure with details
type NotificationFailure struct {
	RequestID     string                 `json:"request_id"`
	UserID        uint                   `json:"user_id"`
	Channel       NotificationChannel    `json:"channel"`
	FailureType   FailureType            `json:"failure_type"`
	ErrorMessage  string                 `json:"error_message"`
	AttemptCount  int                    `json:"attempt_count"`
	LastAttemptAt time.Time              `json:"last_attempt_at"`
	NextRetryAt   *time.Time             `json:"next_retry_at,omitempty"`
	OriginalReq   *NotificationRequest   `json:"original_request"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// RetryableNotificationService provides notification service with retry and compensation mechanisms
type RetryableNotificationService struct {
	baseService  NotificationService
	taskQueue    *queue.TaskQueue
	retryConfig  *RetryConfig
	logger       logger.Logger
	failureStore FailureStore // Interface for storing failure information
}

// FailureStore interface for persisting notification failures
type FailureStore interface {
	StoreFailure(ctx context.Context, failure *NotificationFailure) error
	GetFailure(ctx context.Context, requestID string) (*NotificationFailure, error)
	UpdateFailure(ctx context.Context, failure *NotificationFailure) error
	GetPendingRetries(ctx context.Context, limit int) ([]*NotificationFailure, error)
	MarkAsDeadLetter(ctx context.Context, requestID string, reason string) error
	GetDeadLetters(ctx context.Context, limit int) ([]*DeadLetterEntry, error)
}

// NewRetryableNotificationService creates a new retryable notification service
func NewRetryableNotificationService(
	baseService NotificationService,
	taskQueue *queue.TaskQueue,
	failureStore FailureStore,
	logger logger.Logger,
) *RetryableNotificationService {
	return &RetryableNotificationService{
		baseService:  baseService,
		taskQueue:    taskQueue,
		retryConfig:  DefaultRetryConfig(),
		logger:       logger,
		failureStore: failureStore,
	}
}

// SetRetryConfig updates the retry configuration
func (s *RetryableNotificationService) SetRetryConfig(config *RetryConfig) {
	s.retryConfig = config
}

// Send sends notifications with retry logic and failure compensation
func (s *RetryableNotificationService) Send(ctx context.Context, req *NotificationRequest) ([]*NotificationResult, error) {
	requestID := s.generateRequestID(req)
	
	s.logger.Info("Sending notification with retry support",
		logger.String("request_id", requestID),
		logger.Uint("user_id", req.UserID),
		logger.String("event_type", req.EventType),
		logger.Int("channels_count", len(req.Channels)))

	// First attempt
	results, err := s.baseService.Send(ctx, req)
	if err != nil {
		// Service-level error, retry the entire request
		s.logger.Error("Service-level error, scheduling retry",
			logger.String("request_id", requestID),
			logger.ErrorField(err))
		
		if err := s.scheduleRetry(ctx, req, requestID, 1, err); err != nil {
			s.logger.Error("Failed to schedule retry", logger.ErrorField(err))
		}
		return nil, err
	}

	// Process individual channel results
	var finalResults []*NotificationResult
	hasFailures := false

	for _, result := range results {
		if result.Success {
			finalResults = append(finalResults, result)
			continue
		}

		// Channel-level failure, analyze and handle
		hasFailures = true
		failureType := s.classifyFailure(result.Error, result.Channel)
		
		s.logger.Warn("Channel notification failed",
			logger.String("request_id", requestID),
			logger.String("channel", string(result.Channel)),
			logger.String("error", result.Error),
			logger.String("failure_type", s.failureTypeToString(failureType)))

		// Handle based on failure type
		switch failureType {
		case FailureTypePermanent:
			// Don't retry permanent failures, but log them for analysis
			s.logPermanentFailure(ctx, req, result, requestID)
			finalResults = append(finalResults, result)

		case FailureTypeTemporary, FailureTypeThrottled:
			// Schedule retry for temporary failures
			if err := s.scheduleChannelRetry(ctx, req, result, requestID, 1, failureType); err != nil {
				s.logger.Error("Failed to schedule channel retry", 
					logger.String("channel", string(result.Channel)),
					logger.ErrorField(err))
				finalResults = append(finalResults, result)
			} else {
				// Mark as pending retry
				result.Error = "scheduled for retry"
				finalResults = append(finalResults, result)
			}

		default:
			finalResults = append(finalResults, result)
		}
	}

	// If we had failures but scheduled retries, log the compensation
	if hasFailures {
		s.logger.Info("Notification completed with compensation mechanisms activated",
			logger.String("request_id", requestID),
			logger.Int("total_channels", len(req.Channels)),
			logger.Int("immediate_successes", len(results)-s.countFailures(results)),
			logger.Int("scheduled_retries", s.countScheduledRetries(results)))
	}

	return finalResults, nil
}

// HealthCheck performs health check on the base service
func (s *RetryableNotificationService) HealthCheck(ctx context.Context) error {
	return s.baseService.HealthCheck(ctx)
}

// ProcessRetries processes pending notification retries
func (s *RetryableNotificationService) ProcessRetries(ctx context.Context) error {
	if s.failureStore == nil {
		s.logger.Warn("Failure store not configured, skipping retry processing")
		return nil
	}

	pendingRetries, err := s.failureStore.GetPendingRetries(ctx, 100)
	if err != nil {
		return fmt.Errorf("failed to get pending retries: %w", err)
	}

	s.logger.Info("Processing notification retries", 
		logger.Int("pending_count", len(pendingRetries)))

	for _, failure := range pendingRetries {
		if failure.NextRetryAt != nil && failure.NextRetryAt.After(time.Now()) {
			continue // Not ready for retry yet
		}

		if failure.AttemptCount >= s.retryConfig.MaxRetries {
			// Max retries exceeded, move to dead letter
			s.logger.Warn("Max retries exceeded, moving to dead letter queue",
				logger.String("request_id", failure.RequestID),
				logger.Int("attempts", failure.AttemptCount))
			
			if err := s.failureStore.MarkAsDeadLetter(ctx, failure.RequestID, 
				"max retries exceeded"); err != nil {
				s.logger.Error("Failed to mark as dead letter", logger.ErrorField(err))
			}
			continue
		}

		// Attempt retry
		if err := s.processRetry(ctx, failure); err != nil {
			s.logger.Error("Failed to process retry",
				logger.String("request_id", failure.RequestID),
				logger.ErrorField(err))
		}
	}

	return nil
}

// processRetry processes a single retry attempt
func (s *RetryableNotificationService) processRetry(ctx context.Context, failure *NotificationFailure) error {
	s.logger.Info("Processing notification retry",
		logger.String("request_id", failure.RequestID),
		logger.Int("attempt", failure.AttemptCount+1),
		logger.String("channel", string(failure.Channel)))

	// Create a single-channel request for retry
	retryReq := *failure.OriginalReq
	retryReq.Channels = []NotificationChannel{failure.Channel}

	// Attempt to send
	results, err := s.baseService.Send(ctx, &retryReq)
	failure.AttemptCount++
	failure.LastAttemptAt = time.Now()

	if err != nil || len(results) == 0 || !results[0].Success {
		// Retry failed, schedule next retry or mark as dead letter
		errorMsg := "unknown error"
		if err != nil {
			errorMsg = err.Error()
		} else if len(results) > 0 {
			errorMsg = results[0].Error
		}

		failure.ErrorMessage = errorMsg

		if failure.AttemptCount >= s.retryConfig.MaxRetries {
			// Max retries reached
			return s.failureStore.MarkAsDeadLetter(ctx, failure.RequestID, 
				fmt.Sprintf("max retries exceeded: %s", errorMsg))
		}

		// Schedule next retry
		nextRetry := s.calculateNextRetry(failure.AttemptCount, failure.FailureType)
		failure.NextRetryAt = &nextRetry
		
		return s.failureStore.UpdateFailure(ctx, failure)
	}

	// Retry succeeded!
	s.logger.Info("Notification retry succeeded",
		logger.String("request_id", failure.RequestID),
		logger.String("channel", string(failure.Channel)),
		logger.Int("attempts_taken", failure.AttemptCount))

	// Remove from failure store (success)
	return s.failureStore.MarkAsDeadLetter(ctx, failure.RequestID, "retry_succeeded")
}

// scheduleRetry schedules a full request retry
func (s *RetryableNotificationService) scheduleRetry(ctx context.Context, req *NotificationRequest, requestID string, attempt int, err error) error {
	if s.taskQueue == nil {
		return fmt.Errorf("task queue not configured for retries")
	}

	// Calculate retry delay (for logging purposes)
	delay := s.calculateRetryDelay(attempt, FailureTypeTemporary)
	s.logger.Info("Scheduling full request retry",
		logger.String("request_id", requestID),
		logger.Int("attempt", attempt),
		logger.String("delay", delay.String()))

	retryData := map[string]interface{}{
		"request_id":     requestID,
		"original_req":   req,
		"attempt":        attempt,
		"error":          err.Error(),
		"scheduled_at":   time.Now().Unix(),
	}

	task := queue.NewTask("notification_retry", retryData)
	task.MaxRetry = s.retryConfig.MaxRetries

	return s.taskQueue.Enqueue(ctx, "notification_retry", task)
}

// scheduleChannelRetry schedules a retry for a specific channel
func (s *RetryableNotificationService) scheduleChannelRetry(ctx context.Context, req *NotificationRequest, result *NotificationResult, requestID string, attempt int, failureType FailureType) error {
	if s.failureStore == nil {
		return fmt.Errorf("failure store not configured for channel retries")
	}

	nextRetry := s.calculateNextRetry(attempt, failureType)

	failure := &NotificationFailure{
		RequestID:     fmt.Sprintf("%s_%s", requestID, result.Channel),
		UserID:        req.UserID,
		Channel:       result.Channel,
		FailureType:   failureType,
		ErrorMessage:  result.Error,
		AttemptCount:  attempt,
		LastAttemptAt: time.Now(),
		NextRetryAt:   &nextRetry,
		OriginalReq:   req,
		Metadata: map[string]interface{}{
			"event_type": req.EventType,
			"event_id":   req.EventID,
		},
	}

	return s.failureStore.StoreFailure(ctx, failure)
}

// classifyFailure classifies the type of failure based on error message and channel
func (s *RetryableNotificationService) classifyFailure(errorMsg string, channel NotificationChannel) FailureType {
	errorLower := strings.ToLower(errorMsg)

	// Permanent failures - don't retry
	permanentKeywords := []string{
		"invalid email", "invalid phone", "invalid user", 
		"authentication failed", "unauthorized", "forbidden",
		"not found", "user blocked", "account disabled",
		"malformed", "invalid format", "permission denied",
	}

	for _, keyword := range permanentKeywords {
		if strings.Contains(errorLower, keyword) {
			return FailureTypePermanent
		}
	}

	// Throttled failures - need longer backoff
	throttledKeywords := []string{
		"rate limit", "too many requests", "quota exceeded",
		"throttled", "429", "rate exceeded",
	}

	for _, keyword := range throttledKeywords {
		if strings.Contains(errorLower, keyword) {
			return FailureTypeThrottled
		}
	}

	// Default to temporary failure (network issues, server errors, etc.)
	return FailureTypeTemporary
}

// calculateNextRetry calculates the next retry time based on attempt count and failure type
func (s *RetryableNotificationService) calculateNextRetry(attemptCount int, failureType FailureType) time.Time {
	delay := s.calculateRetryDelay(attemptCount, failureType)
	return time.Now().Add(delay)
}

// calculateRetryDelay calculates the retry delay using exponential backoff with jitter
func (s *RetryableNotificationService) calculateRetryDelay(attemptCount int, failureType FailureType) time.Duration {
	// Base delay calculation
	delay := time.Duration(float64(s.retryConfig.InitialDelay) * 
		math.Pow(s.retryConfig.BackoffMultiplier, float64(attemptCount-1)))

	// Apply failure type multipliers
	switch failureType {
	case FailureTypeThrottled:
		delay *= 3 // Longer delay for rate-limited failures
	case FailureTypePermanent:
		return 0 // No delay for permanent failures (shouldn't retry)
	}

	// Cap the delay
	if delay > s.retryConfig.MaxDelay {
		delay = s.retryConfig.MaxDelay
	}

	// Add jitter to prevent thundering herd
	if s.retryConfig.EnableJitter && delay > 0 {
		jitter := time.Duration(float64(delay) * 0.1) // 10% jitter
		jitterFactor := (2*float64(time.Now().UnixNano()%1000)/1000.0 - 1) // Random factor between -1 and 1
		delay += time.Duration(float64(jitter) * jitterFactor)
	}

	return delay
}

// Utility functions
func (s *RetryableNotificationService) generateRequestID(req *NotificationRequest) string {
	return fmt.Sprintf("notif_%d_%s_%d", req.UserID, req.EventType, time.Now().UnixNano())
}

func (s *RetryableNotificationService) failureTypeToString(ft FailureType) string {
	switch ft {
	case FailureTypeTemporary:
		return "temporary"
	case FailureTypePermanent:
		return "permanent"
	case FailureTypeThrottled:
		return "throttled"
	default:
		return "unknown"
	}
}

func (s *RetryableNotificationService) logPermanentFailure(ctx context.Context, req *NotificationRequest, result *NotificationResult, requestID string) {
	s.logger.Error("Permanent notification failure detected",
		logger.String("request_id", requestID),
		logger.Uint("user_id", req.UserID),
		logger.String("channel", string(result.Channel)),
		logger.String("error", result.Error),
		logger.String("event_type", req.EventType))
}

func (s *RetryableNotificationService) countFailures(results []*NotificationResult) int {
	count := 0
	for _, result := range results {
		if !result.Success {
			count++
		}
	}
	return count
}

func (s *RetryableNotificationService) countScheduledRetries(results []*NotificationResult) int {
	count := 0
	for _, result := range results {
		if !result.Success && strings.Contains(result.Error, "scheduled for retry") {
			count++
		}
	}
	return count
}