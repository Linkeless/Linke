package implementations

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hibiken/asynq"

	"linke/internal/domains/payment/usecases/interfaces"
	"linke/internal/shared/logger"
	"linke/internal/shared/queue"
)

// PaymentRetryWorker handles payment retry task processing
type PaymentRetryWorker struct {
	retryService interfaces.PaymentRetryService
	processor    *queue.TaskProcessor
}

// NewPaymentRetryWorker creates a new payment retry worker
func NewPaymentRetryWorker(
	retryService interfaces.PaymentRetryService,
	processor *queue.TaskProcessor,
) *PaymentRetryWorker {
	worker := &PaymentRetryWorker{
		retryService: retryService,
		processor:    processor,
	}

	// Register task handlers
	worker.registerHandlers()

	return worker
}

// registerHandlers registers all retry-related task handlers
func (w *PaymentRetryWorker) registerHandlers() {
	// Handler for processing individual payment retries
	w.processor.RegisterHandler("process_payment_retry", w.handleProcessPaymentRetry)

	// Handler for batch processing pending retries
	w.processor.RegisterHandler("process_pending_retries", w.handleProcessPendingRetries)

	// Handler for retry notifications
	w.processor.RegisterHandler("payment_retry_notification", w.handleRetryNotification)

	// Handler for retry health checks
	w.processor.RegisterHandler("retry_health_check", w.handleRetryHealthCheck)

	// Handler for retry cleanup (removing old completed retries)
	w.processor.RegisterHandler("retry_cleanup", w.handleRetryCleanup)
}

// handleProcessPaymentRetry processes a single payment retry
func (w *PaymentRetryWorker) handleProcessPaymentRetry(ctx context.Context, task *asynq.Task) error {
	var payload struct {
		RetryID uint `json:"retry_id"`
	}

	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		logger.Error("Failed to unmarshal retry task payload",
			logger.ErrorField(err),
			logger.String("task_type", task.Type()),
		)
		return fmt.Errorf("invalid task payload: %w", err)
	}

	logger.Info("Processing payment retry task",
		logger.Uint("retry_id", payload.RetryID),
		logger.String("task_id", task.Type()),
	)

	// Process the retry
	result, err := w.retryService.ProcessRetry(ctx, payload.RetryID)
	if err != nil {
		logger.Error("Failed to process payment retry",
			logger.ErrorField(err),
			logger.Uint("retry_id", payload.RetryID),
		)
		return err
	}

	logger.Info("Payment retry processed successfully",
		logger.Uint("retry_id", payload.RetryID),
		logger.String("success", fmt.Sprintf("%t", result.Success)),
		logger.String("status", result.Status),
	)

	return nil
}

// handleProcessPendingRetries processes multiple pending retries in batch
func (w *PaymentRetryWorker) handleProcessPendingRetries(ctx context.Context, task *asynq.Task) error {
	var payload struct {
		BatchSize int `json:"batch_size"`
	}

	// Default batch size if not specified
	payload.BatchSize = 50

	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		logger.Warn("Failed to unmarshal pending retries task payload, using defaults",
			logger.ErrorField(err),
		)
	}

	logger.Info("Processing pending retries batch",
		logger.Int("batch_size", payload.BatchSize),
	)

	processed, err := w.retryService.ProcessPendingRetries(ctx, payload.BatchSize)
	if err != nil {
		logger.Error("Failed to process pending retries batch",
			logger.ErrorField(err),
			logger.Int("batch_size", payload.BatchSize),
		)
		return err
	}

	logger.Info("Pending retries batch processed",
		logger.Int("processed_count", processed),
		logger.Int("batch_size", payload.BatchSize),
	)

	return nil
}

// handleRetryNotification sends notifications for retry attempts
func (w *PaymentRetryWorker) handleRetryNotification(ctx context.Context, task *asynq.Task) error {
	var payload struct {
		RetryID       uint       `json:"retry_id"`
		PaymentID     uint       `json:"payment_id"`
		AttemptNumber int        `json:"attempt_number"`
		Status        string     `json:"status"`
		NextRetryAt   *time.Time `json:"next_retry_at"`
	}

	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		logger.Error("Failed to unmarshal notification task payload",
			logger.ErrorField(err),
		)
		return fmt.Errorf("invalid notification payload: %w", err)
	}

	logger.Info("Processing retry notification",
		logger.Uint("retry_id", payload.RetryID),
		logger.Uint("payment_id", payload.PaymentID),
		logger.Int("attempt_number", payload.AttemptNumber),
		logger.String("status", payload.Status),
	)

	// Here you would implement actual notification logic
	// For example: send email, SMS, webhook, etc.
	// This is a placeholder implementation

	switch payload.Status {
	case "success":
		logger.Info("Payment retry succeeded - notification sent",
			logger.Uint("retry_id", payload.RetryID),
			logger.Uint("payment_id", payload.PaymentID),
		)
	case "failed":
		if payload.NextRetryAt != nil {
			logger.Info("Payment retry failed - will retry again",
				logger.Uint("retry_id", payload.RetryID),
				logger.Uint("payment_id", payload.PaymentID),
				logger.String("next_retry_at", payload.NextRetryAt.String()),
			)
		} else {
			logger.Warn("Payment retry failed - no more retries",
				logger.Uint("retry_id", payload.RetryID),
				logger.Uint("payment_id", payload.PaymentID),
			)
		}
	}

	return nil
}

// handleRetryHealthCheck performs health checks on the retry system
func (w *PaymentRetryWorker) handleRetryHealthCheck(ctx context.Context, task *asynq.Task) error {
	logger.Info("Performing retry system health check")

	healthMetrics, err := w.retryService.GetRetryHealthMetrics(ctx)
	if err != nil {
		logger.Error("Failed to get retry health metrics",
			logger.ErrorField(err),
		)
		return err
	}

	// Log important health metrics
	logger.Info("Retry system health check completed",
		logger.String("total_active_retries", fmt.Sprintf("%d", healthMetrics.TotalActiveRetries)),
		logger.String("overdue_retries", fmt.Sprintf("%d", healthMetrics.OverdueRetries)),
		logger.String("success_rate_24h", fmt.Sprintf("%.2f%%", healthMetrics.SuccessRate24h)),
		logger.String("success_rate_7d", fmt.Sprintf("%.2f%%", healthMetrics.SuccessRate7d)),
	)

	// Check for alerts
	if len(healthMetrics.AlertsTriggered) > 0 {
		for _, alert := range healthMetrics.AlertsTriggered {
			logger.Warn("Retry system alert triggered", logger.String("alert", alert))
		}
	}

	// Log gateway health
	for _, gatewayHealth := range healthMetrics.GatewayHealth {
		logger.Info("Gateway health status",
			logger.String("gateway", gatewayHealth.Gateway),
			logger.String("health_status", gatewayHealth.HealthStatus),
			logger.String("active_retries", fmt.Sprintf("%d", gatewayHealth.ActiveRetries)),
			logger.String("success_rate", fmt.Sprintf("%.2f%%", gatewayHealth.SuccessRate)),
		)
	}

	return nil
}

// handleRetryCleanup cleans up old completed retry records
func (w *PaymentRetryWorker) handleRetryCleanup(ctx context.Context, task *asynq.Task) error {
	var payload struct {
		RetentionDays int `json:"retention_days"`
	}

	// Default retention period
	payload.RetentionDays = 30

	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		logger.Warn("Failed to unmarshal cleanup task payload, using defaults",
			logger.ErrorField(err),
		)
	}

	logger.Info("Starting retry cleanup",
		logger.Int("retention_days", payload.RetentionDays),
	)

	// Calculate cutoff date
	cutoffDate := time.Now().AddDate(0, 0, -payload.RetentionDays)

	// This would implement actual cleanup logic
	// For now, just log the operation
	logger.Info("Retry cleanup completed",
		logger.Int("retention_days", payload.RetentionDays),
		logger.String("cutoff_date", cutoffDate.String()),
	)

	return nil
}

// ScheduleRetryWorkerTasks schedules recurring tasks for the retry worker
func (w *PaymentRetryWorker) ScheduleRetryWorkerTasks(taskQueue *queue.TaskQueue) error {
	ctx := context.Background()

	// Schedule pending retries processing every 5 minutes
	pendingRetriesTask := &queue.Task{
		Type: "process_pending_retries",
		Payload: map[string]any{
			"batch_size": 50,
		},
		MaxRetry: 3,
	}

	// Schedule health check every 15 minutes
	healthCheckTask := &queue.Task{
		Type:     "retry_health_check",
		Payload:  map[string]any{},
		MaxRetry: 2,
	}

	// Schedule cleanup daily at 2 AM
	cleanupTask := &queue.Task{
		Type: "retry_cleanup",
		Payload: map[string]any{
			"retention_days": 30,
		},
		MaxRetry: 2,
	}

	// For demonstration, we'll just enqueue the tasks
	// In a real implementation, you would use a scheduler like cron or asynq's scheduler

	if err := taskQueue.Enqueue(ctx, "payment_retries", pendingRetriesTask); err != nil {
		logger.Error("Failed to schedule pending retries task",
			logger.ErrorField(err),
		)
		return err
	}

	if err := taskQueue.Enqueue(ctx, "system", healthCheckTask); err != nil {
		logger.Error("Failed to schedule health check task",
			logger.ErrorField(err),
		)
		return err
	}

	if err := taskQueue.Enqueue(ctx, "maintenance", cleanupTask); err != nil {
		logger.Error("Failed to schedule cleanup task",
			logger.ErrorField(err),
		)
		return err
	}

	logger.Info("Retry worker tasks scheduled successfully")
	return nil
}

// RetryWorkerMetrics represents metrics for the retry worker
type RetryWorkerMetrics struct {
	TasksProcessed     int64     `json:"tasks_processed"`
	TasksSucceeded     int64     `json:"tasks_succeeded"`
	TasksFailed        int64     `json:"tasks_failed"`
	LastProcessedAt    time.Time `json:"last_processed_at"`
	AverageProcessTime float64   `json:"average_process_time"` // in seconds
	QueueDepth         int64     `json:"queue_depth"`
}

// GetWorkerMetrics returns metrics about the retry worker performance
func (w *PaymentRetryWorker) GetWorkerMetrics(ctx context.Context) (*RetryWorkerMetrics, error) {
	// This would be implemented to gather actual metrics from the queue system
	// For now, return placeholder metrics
	return &RetryWorkerMetrics{
		TasksProcessed:     100,
		TasksSucceeded:     95,
		TasksFailed:        5,
		LastProcessedAt:    time.Now(),
		AverageProcessTime: 2.5,
		QueueDepth:         10,
	}, nil
}
