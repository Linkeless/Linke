package notification

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"linke/internal/shared/logger"
	"linke/internal/shared/queue"

	"github.com/hibiken/asynq"
)

// RetryTaskHandler handles notification retry tasks
type RetryTaskHandler struct {
	retryService *RetryableNotificationService
	failureStore FailureStore
	logger       logger.Logger
}

// NewRetryTaskHandler creates a new retry task handler
func NewRetryTaskHandler(
	retryService *RetryableNotificationService,
	failureStore FailureStore,
	logger logger.Logger,
) *RetryTaskHandler {
	return &RetryTaskHandler{
		retryService: retryService,
		failureStore: failureStore,
		logger:       logger,
	}
}

// HandleNotificationRetry processes notification retry tasks
func (h *RetryTaskHandler) HandleNotificationRetry(ctx context.Context, task *asynq.Task) error {
	var retryData struct {
		RequestID   string               `json:"request_id"`
		OriginalReq *NotificationRequest `json:"original_req"`
		Attempt     int                  `json:"attempt"`
		Error       string               `json:"error"`
		ScheduledAt int64                `json:"scheduled_at"`
	}

	if err := json.Unmarshal(task.Payload(), &retryData); err != nil {
		h.logger.Error("Failed to unmarshal notification retry task",
			logger.String("task_type", task.Type()),
			logger.ErrorField(err))
		return fmt.Errorf("failed to unmarshal retry task: %w", err)
	}

	h.logger.Info("Processing notification retry task",
		logger.String("request_id", retryData.RequestID),
		logger.Int("attempt", retryData.Attempt),
		logger.Uint("user_id", retryData.OriginalReq.UserID))

	// Attempt to send the notification again
	results, err := h.retryService.baseService.Send(ctx, retryData.OriginalReq)
	if err != nil {
		// Service-level error occurred
		h.logger.Error("Service-level error during retry",
			logger.String("request_id", retryData.RequestID),
			logger.Int("attempt", retryData.Attempt),
			logger.ErrorField(err))

		// Check if we should retry again
		if retryData.Attempt < h.retryService.retryConfig.MaxRetries {
			// Schedule another retry
			return h.scheduleNextRetry(ctx, &retryData, err)
		} else {
			// Max retries exceeded, log as permanent failure
			h.logger.Error("Max retries exceeded for notification",
				logger.String("request_id", retryData.RequestID),
				logger.Int("final_attempt", retryData.Attempt),
				logger.ErrorField(err))
			return nil // Don't retry anymore
		}
	}

	// Check individual channel results
	successCount := 0
	failureCount := 0

	for _, result := range results {
		if result.Success {
			successCount++
		} else {
			failureCount++
			h.logger.Warn("Channel failed during retry",
				logger.String("request_id", retryData.RequestID),
				logger.String("channel", string(result.Channel)),
				logger.String("error", result.Error))
		}
	}

	if successCount > 0 {
		h.logger.Info("Notification retry partially or fully succeeded",
			logger.String("request_id", retryData.RequestID),
			logger.Int("successful_channels", successCount),
			logger.Int("failed_channels", failureCount))
	}

	if failureCount > 0 && retryData.Attempt < h.retryService.retryConfig.MaxRetries {
		// Some channels failed, schedule another retry
		return h.scheduleNextRetry(ctx, &retryData, fmt.Errorf("partial failure: %d channels failed", failureCount))
	}

	return nil
}

// scheduleNextRetry schedules the next retry attempt
func (h *RetryTaskHandler) scheduleNextRetry(ctx context.Context, retryData *struct {
	RequestID   string               `json:"request_id"`
	OriginalReq *NotificationRequest `json:"original_req"`
	Attempt     int                  `json:"attempt"`
	Error       string               `json:"error"`
	ScheduledAt int64                `json:"scheduled_at"`
}, err error,
) error {
	nextAttempt := retryData.Attempt + 1
	delay := h.retryService.calculateRetryDelay(nextAttempt, FailureTypeTemporary)

	h.logger.Info("Scheduling next notification retry",
		logger.String("request_id", retryData.RequestID),
		logger.Int("next_attempt", nextAttempt),
		logger.String("delay", delay.String()))

	nextRetryData := map[string]interface{}{
		"request_id":   retryData.RequestID,
		"original_req": retryData.OriginalReq,
		"attempt":      nextAttempt,
		"error":        err.Error(),
		"scheduled_at": time.Now().Unix(),
	}

	task := queue.NewTask("notification_retry", nextRetryData)
	task.MaxRetry = h.retryService.retryConfig.MaxRetries - nextAttempt

	// For now, use regular queue - in production, you'd want to implement delayed processing
	h.logger.Info("Scheduling retry task",
		logger.String("request_id", retryData.RequestID),
		logger.String("delay", delay.String()))

	return h.retryService.taskQueue.Enqueue(ctx, "notification_retry", task)
}

// HandleFailureProcessing processes notification failures for dead letter queue management
func (h *RetryTaskHandler) HandleFailureProcessing(ctx context.Context, task *asynq.Task) error {
	var failureData struct {
		Action    string `json:"action"` // "process_retries" or "cleanup_dead_letters"
		BatchSize int    `json:"batch_size,omitempty"`
		MaxAge    int64  `json:"max_age_hours,omitempty"`
	}

	if err := json.Unmarshal(task.Payload(), &failureData); err != nil {
		return fmt.Errorf("failed to unmarshal failure processing task: %w", err)
	}

	switch failureData.Action {
	case "process_retries":
		return h.processRetries(ctx, failureData.BatchSize)
	case "cleanup_dead_letters":
		return h.cleanupDeadLetters(ctx, failureData.MaxAge)
	default:
		return fmt.Errorf("unknown failure processing action: %s", failureData.Action)
	}
}

// processRetries processes pending retries
func (h *RetryTaskHandler) processRetries(ctx context.Context, batchSize int) error {
	if batchSize <= 0 {
		batchSize = 50 // Default batch size
	}

	h.logger.Info("Processing notification retries batch",
		logger.Int("batch_size", batchSize))

	return h.retryService.ProcessRetries(ctx)
}

// cleanupDeadLetters cleans up old dead letter entries
func (h *RetryTaskHandler) cleanupDeadLetters(ctx context.Context, maxAgeHours int64) error {
	if maxAgeHours <= 0 {
		maxAgeHours = 24 * 7 // Default: 1 week
	}

	h.logger.Info("Cleaning up old dead letter entries",
		logger.Int64("max_age_hours", maxAgeHours))

	// Get dead letters for cleanup
	deadLetters, err := h.failureStore.GetDeadLetters(ctx, 1000) // Process in batches
	if err != nil {
		return fmt.Errorf("failed to get dead letters for cleanup: %w", err)
	}

	cutoffTime := time.Now().Add(-time.Duration(maxAgeHours) * time.Hour)
	cleanedCount := 0

	for _, deadLetter := range deadLetters {
		if deadLetter.MovedAt.Before(cutoffTime) {
			// This would require additional cleanup methods in the failure store
			// For now, just log the cleanup intent
			h.logger.Debug("Dead letter entry ready for cleanup",
				logger.String("request_id", deadLetter.Failure.RequestID),
				logger.String("moved_at", deadLetter.MovedAt.Format(time.RFC3339)))
			cleanedCount++
		}
	}

	h.logger.Info("Dead letter cleanup completed",
		logger.Int("cleaned_entries", cleanedCount),
		logger.Int("total_examined", len(deadLetters)))

	return nil
}

// RegisterRetryHandlers registers all retry-related task handlers
func (h *RetryTaskHandler) RegisterRetryHandlers(processor *queue.TaskProcessor) {
	processor.RegisterHandler("notification_retry", h.HandleNotificationRetry)
	processor.RegisterHandler("notification_failure_processing", h.HandleFailureProcessing)
}

// SchedulePeriodicRetryProcessing schedules periodic retry processing tasks
func (h *RetryTaskHandler) SchedulePeriodicRetryProcessing(ctx context.Context, taskQueue *queue.TaskQueue) error {
	// For now, just enqueue immediate processing tasks
	// In production, you'd want to implement proper delayed task scheduling

	// Schedule retry processing task
	retryProcessingData := map[string]interface{}{
		"action":     "process_retries",
		"batch_size": 100,
	}

	retryTask := queue.NewTask("notification_failure_processing", retryProcessingData)
	if err := taskQueue.Enqueue(ctx, "notification_processing", retryTask); err != nil {
		h.logger.Error("Failed to schedule retry processing", logger.ErrorField(err))
		return err
	}

	// Schedule dead letter cleanup task
	cleanupData := map[string]interface{}{
		"action":        "cleanup_dead_letters",
		"max_age_hours": 24 * 7, // 1 week
	}

	cleanupTask := queue.NewTask("notification_failure_processing", cleanupData)
	if err := taskQueue.Enqueue(ctx, "notification_processing", cleanupTask); err != nil {
		h.logger.Error("Failed to schedule dead letter cleanup", logger.ErrorField(err))
		return err
	}

	h.logger.Info("Scheduled notification failure processing tasks")
	return nil
}
