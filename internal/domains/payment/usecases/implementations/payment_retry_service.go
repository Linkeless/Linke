package implementations

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"linke/internal/domains/payment/constants"
	"linke/internal/domains/payment/dto"
	"linke/internal/domains/payment/entities"
	"linke/internal/domains/payment/usecases/interfaces"
	"linke/internal/shared/logger"
	"linke/internal/shared/queue"
)

// TaskQueueInterface defines the interface for task queue operations needed by payment retry service
type TaskQueueInterface interface {
	Enqueue(ctx context.Context, queueName string, task *queue.Task) error
	EnqueueDelayed(ctx context.Context, queueName string, task *queue.Task, delay time.Duration) error
	EnqueueAt(ctx context.Context, queueName string, task *queue.Task, processAt time.Time) error
}

// paymentRetryService implements PaymentRetryService interface
type paymentRetryService struct {
	retryRepo        interfaces.PaymentRetryRepository
	retryHistoryRepo interfaces.PaymentRetryHistoryRepository
	paymentService   interfaces.PaymentService
	taskQueue        TaskQueueInterface
	config           *dto.RetryConfiguration
}

// NewPaymentRetryService creates a new payment retry service
func NewPaymentRetryService(
	retryRepo interfaces.PaymentRetryRepository,
	retryHistoryRepo interfaces.PaymentRetryHistoryRepository,
	paymentService interfaces.PaymentService,
	taskQueue TaskQueueInterface,
) interfaces.PaymentRetryService {
	service := &paymentRetryService{
		retryRepo:        retryRepo,
		retryHistoryRepo: retryHistoryRepo,
		paymentService:   paymentService,
		taskQueue:        taskQueue,
		config:           getDefaultRetryConfiguration(),
	}

	return service
}

// Retry management

func (s *paymentRetryService) InitiateRetry(ctx context.Context, paymentRecord *entities.PaymentRecord, failureType, failureCode, errorMessage string) (*entities.PaymentRetry, error) {
	logger.Info("Initiating payment retry",
		logger.Uint("payment_record_id", paymentRecord.ID),
		logger.String("failure_type", failureType),
		logger.String("failure_code", failureCode),
	)

	// Check if retry already exists
	existingRetry, err := s.retryRepo.GetByPaymentRecordID(ctx, paymentRecord.ID)
	if err == nil && existingRetry != nil {
		// Update existing retry for next attempt
		existingRetry.UpdateForNextAttempt(failureType, failureCode, errorMessage)
		if err := s.retryRepo.Update(ctx, existingRetry); err != nil {
			return nil, fmt.Errorf("failed to update existing retry: %w", err)
		}

		// Schedule next retry if should retry
		if existingRetry.ShouldRetry() {
			if err := s.scheduleRetryTask(ctx, existingRetry); err != nil {
				logger.Error("Failed to schedule retry task",
					logger.Error2("error", err),
					logger.Uint("retry_id", existingRetry.ID),
				)
			}
		}

		return existingRetry, nil
	}

	// Get retry strategy for this gateway/payment method
	strategy, err := s.GetRetryStrategy(ctx, paymentRecord.Gateway, paymentRecord.PaymentMethod)
	if err != nil {
		logger.Warn("Failed to get retry strategy, using default",
			logger.Error2("error", err),
			logger.String("gateway", paymentRecord.Gateway),
		)
		strategy = s.getDefaultStrategyForGateway(paymentRecord.Gateway)
	}

	// Create new retry record
	retry := &entities.PaymentRetry{
		PaymentRecordID:  paymentRecord.ID,
		AttemptNumber:    0,
		MaxAttempts:      strategy.MaxAttempts,
		LastAttemptAt:    time.Now(),
		RetryStrategy:    strategy.Strategy,
		InitialDelay:     strategy.InitialDelay,
		MaxDelay:         strategy.MaxDelay,
		BackoffFactor:    strategy.BackoffFactor,
		Status:           constants.PaymentRetryStatusPending,
		FailureType:      failureType,
		LastFailureCode:  failureCode,
		LastErrorMessage: errorMessage,
	}

	// Set gateway-specific configuration
	if err := retry.SetGatewayConfig(strategy); err != nil {
		logger.Warn("Failed to set gateway config",
			logger.Error2("error", err),
		)
	}

	// Calculate next retry time
	retry.NextRetryAt = retry.CalculateNextRetryTime()

	// Create retry record
	if err := s.retryRepo.Create(ctx, retry); err != nil {
		return nil, fmt.Errorf("failed to create retry record: %w", err)
	}

	// Schedule retry task if should retry
	if retry.ShouldRetry() {
		if err := s.scheduleRetryTask(ctx, retry); err != nil {
			logger.Error("Failed to schedule retry task",
				logger.Error2("error", err),
				logger.Uint("retry_id", retry.ID),
			)
		}
	}

	logger.Info("Payment retry initiated successfully",
		logger.Uint("retry_id", retry.ID),
		logger.Uint("payment_record_id", paymentRecord.ID),
		logger.String("next_retry_at", retry.NextRetryAt.String()),
	)

	return retry, nil
}

func (s *paymentRetryService) ProcessPendingRetries(ctx context.Context, batchSize int) (int, error) {
	logger.Debug("Processing pending retries", logger.Int("batch_size", batchSize))

	// Get retries that are due for processing
	retries, err := s.retryRepo.GetRetriesDueForProcessing(ctx, time.Now(), batchSize)
	if err != nil {
		return 0, fmt.Errorf("failed to get retries due for processing: %w", err)
	}

	if len(retries) == 0 {
		return 0, nil
	}

	processed := 0
	for _, retry := range retries {
		result, err := s.ProcessRetry(ctx, retry.ID)
		if err != nil {
			logger.Error("Failed to process retry",
				logger.Error2("error", err),
				logger.Uint("retry_id", retry.ID),
			)
			continue
		}

		logger.Info("Processed retry",
			logger.Uint("retry_id", retry.ID),
			logger.String("success", fmt.Sprintf("%t", result.Success)),
			logger.String("status", result.Status),
		)

		processed++
	}

	return processed, nil
}

func (s *paymentRetryService) ProcessRetry(ctx context.Context, retryID uint) (*dto.RetryResult, error) {
	// Get retry record
	retry, err := s.retryRepo.GetByID(ctx, retryID)
	if err != nil {
		return nil, fmt.Errorf("failed to get retry record: %w", err)
	}

	// Check if retry is still valid
	if !retry.ShouldRetry() {
		return &dto.RetryResult{
			Success:      false,
			RetryID:      retryID,
			Status:       retry.Status,
			ErrorMessage: "retry no longer valid",
		}, nil
	}

	// Mark as in progress
	retry.Status = constants.PaymentRetryStatusInProgress
	if err := s.retryRepo.Update(ctx, retry); err != nil {
		return nil, fmt.Errorf("failed to mark retry as in progress: %w", err)
	}

	// Get payment record
	paymentRecord, err := s.paymentService.GetPaymentRecord(ctx, "")
	if err != nil {
		// If we can't get payment record, mark retry as failed
		retry.Status = constants.PaymentRetryStatusFailed
		s.retryRepo.Update(ctx, retry)
		return nil, fmt.Errorf("failed to get payment record: %w", err)
	}

	// Create retry attempt history
	startTime := time.Now()
	attemptHistory := &entities.PaymentRetryHistory{
		PaymentRetryID:  retry.ID,
		PaymentRecordID: retry.PaymentRecordID,
		AttemptNumber:   retry.AttemptNumber + 1,
		AttemptedAt:     startTime,
		Status:          "in_progress",
	}

	// Process the payment retry
	var result *dto.RetryResult
	var paymentResult *dto.PaymentProcessResult

	// Simulate payment processing - in reality this would call the payment gateway
	paymentResult, err = s.processPaymentRetry(ctx, paymentRecord, retry)

	// Calculate duration
	duration := int(time.Since(startTime).Milliseconds())
	attemptHistory.Duration = duration

	if err != nil {
		// Handle retry failure
		attemptHistory.Status = constants.AttemptStatusFailed
		attemptHistory.ErrorType = s.ClassifyFailure(ctx, paymentRecord.Gateway, paymentRecord.PaymentMethod, paymentResult.ErrorCode, err.Error())
		attemptHistory.FailureReason = err.Error()
		attemptHistory.ResponseCode = paymentResult.ErrorCode

		// Update retry for next attempt
		retry.UpdateForNextAttempt(attemptHistory.ErrorType, paymentResult.ErrorCode, err.Error())

		result = &dto.RetryResult{
			Success:       false,
			RetryID:       retryID,
			AttemptNumber: retry.AttemptNumber,
			Status:        retry.Status,
			ErrorMessage:  err.Error(),
			PaymentResult: paymentResult,
		}

		// Schedule next retry if applicable
		if retry.ShouldRetry() {
			attemptHistory.NextRetryAt = &retry.NextRetryAt
			attemptHistory.DelayFromPrevious = int(retry.NextRetryAt.Sub(time.Now()).Seconds())
			result.NextRetryAt = &retry.NextRetryAt

			// Schedule next retry
			if scheduleErr := s.scheduleRetryTask(ctx, retry); scheduleErr != nil {
				logger.Error("Failed to schedule next retry",
					logger.Error2("error", scheduleErr),
					logger.Uint("retry_id", retry.ID),
				)
			}
		} else {
			result.CompletedAt = retry.CompletedAt
		}
	} else {
		// Handle retry success
		attemptHistory.Status = constants.AttemptStatusSuccess
		attemptHistory.ResponseCode = "SUCCESS"

		retry.MarkAsSuccessful()

		result = &dto.RetryResult{
			Success:       true,
			RetryID:       retryID,
			AttemptNumber: retry.AttemptNumber,
			Status:        retry.Status,
			CompletedAt:   retry.CompletedAt,
			PaymentResult: paymentResult,
		}
	}

	// Save retry update
	if err := s.retryRepo.Update(ctx, retry); err != nil {
		logger.Error("Failed to update retry record",
			logger.Error2("error", err),
			logger.Uint("retry_id", retry.ID),
		)
	}

	// Save attempt history
	if err := s.retryHistoryRepo.Create(ctx, attemptHistory); err != nil {
		logger.Error("Failed to create retry history",
			logger.Error2("error", err),
			logger.Uint("retry_id", retry.ID),
		)
	}

	result.History = dto.ToPaymentRetryHistoryResponse(attemptHistory)

	// Send notification
	if err := s.NotifyRetryAttempt(ctx, retry, attemptHistory); err != nil {
		logger.Warn("Failed to send retry notification",
			logger.Error2("error", err),
			logger.Uint("retry_id", retry.ID),
		)
	}

	return result, nil
}

func (s *paymentRetryService) CancelRetry(ctx context.Context, retryID uint, reason string) error {
	return s.retryRepo.CancelRetry(ctx, retryID, reason)
}

func (s *paymentRetryService) ResetRetry(ctx context.Context, retryID uint) error {
	return s.retryRepo.ResetRetry(ctx, retryID)
}

// Configuration and strategy

func (s *paymentRetryService) GetRetryStrategy(ctx context.Context, gateway, paymentMethod string) (*entities.RetryStrategyConfig, error) {
	// First check for gateway-specific configuration
	if strategy, exists := s.config.GatewayStrategies[gateway]; exists {
		return strategy, nil
	}

	// Fall back to default strategy for the gateway
	if defaultStrategy, exists := entities.DefaultRetryStrategies[gateway]; exists {
		return &defaultStrategy, nil
	}

	// Use system default
	return s.config.DefaultStrategy, nil
}

func (s *paymentRetryService) UpdateRetryStrategy(ctx context.Context, gateway, paymentMethod string, config *entities.RetryStrategyConfig) error {
	// Update in-memory configuration
	if s.config.GatewayStrategies == nil {
		s.config.GatewayStrategies = make(map[string]*entities.RetryStrategyConfig)
	}
	s.config.GatewayStrategies[gateway] = config

	// In a real implementation, this would persist to database or configuration store
	logger.Info("Updated retry strategy",
		logger.String("gateway", gateway),
		logger.String("payment_method", paymentMethod),
		logger.Int("max_attempts", config.MaxAttempts),
	)

	return nil
}

func (s *paymentRetryService) CalculateNextRetryTime(ctx context.Context, retry *entities.PaymentRetry) time.Time {
	return retry.CalculateNextRetryTime()
}

// Query operations

func (s *paymentRetryService) GetRetryByPaymentID(ctx context.Context, paymentRecordID uint) (*entities.PaymentRetry, error) {
	return s.retryRepo.GetByPaymentRecordID(ctx, paymentRecordID)
}

func (s *paymentRetryService) GetRetryWithHistory(ctx context.Context, retryID uint) (*dto.RetryWithHistory, error) {
	retry, err := s.retryRepo.GetByID(ctx, retryID)
	if err != nil {
		return nil, err
	}

	history, err := s.retryHistoryRepo.GetByRetryID(ctx, retryID)
	if err != nil {
		return nil, err
	}

	return &dto.RetryWithHistory{
		Retry:   dto.ToPaymentRetryResponse(retry),
		History: convertRetryHistorySlice(history),
	}, nil
}

func (s *paymentRetryService) GetActiveRetries(ctx context.Context, filters *dto.RetryFilters, limit, offset int) ([]*entities.PaymentRetry, int64, error) {
	return s.retryRepo.GetAllRetries(ctx, filters, limit, offset)
}

func (s *paymentRetryService) GetRetryHistory(ctx context.Context, retryID uint) ([]*entities.PaymentRetryHistory, error) {
	return s.retryHistoryRepo.GetByRetryID(ctx, retryID)
}

// Statistics and monitoring

func (s *paymentRetryService) GetRetryStatistics(ctx context.Context, gateway string, days int) (*dto.RetryStatistics, error) {
	fromDate := time.Now().AddDate(0, 0, -days)
	toDate := time.Now()
	return s.retryRepo.GetRetryStatsByGateway(ctx, gateway, fromDate, toDate)
}

func (s *paymentRetryService) GetFailureAnalysis(ctx context.Context, gateway string, days int) (*dto.FailureAnalysis, error) {
	// Get failure patterns
	patterns, err := s.retryHistoryRepo.GetFailurePatterns(ctx, gateway, days)
	if err != nil {
		return nil, err
	}

	// Get basic statistics
	stats, err := s.GetRetryStatistics(ctx, gateway, days)
	if err != nil {
		return nil, err
	}

	// Analyze failure types and create recommendations
	analysis := &dto.FailureAnalysis{
		Gateway:            gateway,
		TotalFailures:      stats.FailedRetries,
		FailurePatterns:    patterns,
		RecoveryRate:       stats.SuccessRate,
		RecommendedActions: s.generateFailureRecommendations(patterns),
	}

	return analysis, nil
}

func (s *paymentRetryService) GetRetryHealthMetrics(ctx context.Context) (*dto.RetryHealthMetrics, error) {
	// Get overall metrics
	activeRetries, err := s.retryRepo.GetPendingRetries(ctx, 0)
	if err != nil {
		return nil, err
	}

	overdueRetries, err := s.retryRepo.GetRetriesDueForProcessing(ctx, time.Now().Add(-time.Hour), 0)
	if err != nil {
		return nil, err
	}

	successRate24h, _ := s.retryRepo.GetRetrySuccessRate(ctx, "", 1)
	successRate7d, _ := s.retryRepo.GetRetrySuccessRate(ctx, "", 7)

	// Get gateway-specific health
	gatewayHealth := []*dto.GatewayHealthMetric{}
	for gateway := range entities.DefaultRetryStrategies {
		activeGatewayRetries, _ := s.retryRepo.GetActiveRetriesForGateway(ctx, gateway)
		successRate, _ := s.retryRepo.GetRetrySuccessRate(ctx, gateway, 7)

		health := &dto.GatewayHealthMetric{
			Gateway:       gateway,
			ActiveRetries: int64(len(activeGatewayRetries)),
			SuccessRate:   successRate,
			HealthStatus:  s.determineGatewayHealth(successRate, len(activeGatewayRetries)),
		}
		gatewayHealth = append(gatewayHealth, health)
	}

	// Generate alerts and recommendations
	alerts := s.generateHealthAlerts(int64(len(activeRetries)), int64(len(overdueRetries)), successRate24h)
	recommendations := s.generateSystemRecommendations(gatewayHealth)

	return &dto.RetryHealthMetrics{
		TotalActiveRetries:    int64(len(activeRetries)),
		OverdueRetries:        int64(len(overdueRetries)),
		SuccessRate24h:        successRate24h,
		SuccessRate7d:         successRate7d,
		GatewayHealth:         gatewayHealth,
		AlertsTriggered:       alerts,
		SystemRecommendations: recommendations,
	}, nil
}

// Admin operations

func (s *paymentRetryService) GetRetriesForAdmin(ctx context.Context, filters *dto.AdminRetryFilters) (*dto.AdminRetryResponse, error) {
	retries, total, err := s.retryRepo.GetAllRetries(ctx, filters.RetryFilters, filters.Limit, filters.Offset)
	if err != nil {
		return nil, err
	}

	// Calculate pagination
	page := (filters.Offset / filters.Limit) + 1
	totalPages := int((total + int64(filters.Limit) - 1) / int64(filters.Limit))

	response := &dto.AdminRetryResponse{
		Retries:    convertRetrySliceToDTO(retries),
		TotalCount: total,
		Page:       page,
		PageSize:   filters.Limit,
		TotalPages: totalPages,
	}

	// Add statistics if requested
	if filters.IncludeHistory {
		// Calculate statistics
		// This would be implemented based on specific requirements
	}

	return response, nil
}

func (s *paymentRetryService) BulkCancelRetries(ctx context.Context, retryIDs []uint, reason string) error {
	for _, id := range retryIDs {
		if err := s.CancelRetry(ctx, id, reason); err != nil {
			logger.Error("Failed to cancel retry in bulk operation",
				logger.Error2("error", err),
				logger.Uint("retry_id", id),
			)
		}
	}
	return nil
}

func (s *paymentRetryService) BulkResetRetries(ctx context.Context, retryIDs []uint) error {
	for _, id := range retryIDs {
		if err := s.ResetRetry(ctx, id); err != nil {
			logger.Error("Failed to reset retry in bulk operation",
				logger.Error2("error", err),
				logger.Uint("retry_id", id),
			)
		}
	}
	return nil
}

// Integration points

func (s *paymentRetryService) ClassifyFailure(ctx context.Context, gateway, paymentMethod, errorCode, errorMessage string) string {
	// Check configuration rules first
	if s.config.FailureClassificationRules != nil {
		for _, rule := range s.config.FailureClassificationRules {
			if rule.Gateway != gateway {
				continue
			}
			if rule.PaymentMethod != "" && rule.PaymentMethod != paymentMethod {
				continue
			}

			// Check error codes
			for _, code := range rule.ErrorCodes {
				if code == errorCode {
					return rule.FailureType
				}
			}

			// Check error patterns
			for _, pattern := range rule.ErrorPatterns {
				if matched, _ := regexp.MatchString(pattern, errorMessage); matched {
					return rule.FailureType
				}
			}
		}
	}

	// Default classification logic
	return s.classifyFailureDefault(errorCode, errorMessage)
}

func (s *paymentRetryService) ShouldRetryPayment(ctx context.Context, paymentRecord *entities.PaymentRecord, errorCode string) bool {
	// Classify the failure
	failureType := s.ClassifyFailure(ctx, paymentRecord.Gateway, paymentRecord.PaymentMethod, errorCode, "")

	// Don't retry permanent failures
	return failureType != constants.FailureTypePermanent
}

func (s *paymentRetryService) NotifyRetryAttempt(ctx context.Context, retry *entities.PaymentRetry, attempt *entities.PaymentRetryHistory) error {
	if !s.config.NotificationSettings.Enabled {
		return nil
	}

	// Create notification payload
	payload := map[string]any{
		"retry_id":       retry.ID,
		"payment_id":     retry.PaymentRecordID,
		"attempt_number": attempt.AttemptNumber,
		"status":         attempt.Status,
		"next_retry_at":  attempt.NextRetryAt,
	}

	// Send to notification queue
	task := &queue.Task{
		Type:     "payment_retry_notification",
		Payload:  payload,
		MaxRetry: 3,
	}

	return s.taskQueue.Enqueue(ctx, "notifications", task)
}

// Helper methods

func (s *paymentRetryService) scheduleRetryTask(ctx context.Context, retry *entities.PaymentRetry) error {
	delay := time.Until(retry.NextRetryAt)
	if delay < 0 {
		delay = 0 // Process immediately if overdue
	}

	task := &queue.Task{
		Type: "process_payment_retry",
		Payload: map[string]any{
			"retry_id": retry.ID,
		},
		MaxRetry: 1,
	}

	return s.taskQueue.EnqueueDelayed(ctx, "payment_retries", task, delay)
}

func (s *paymentRetryService) processPaymentRetry(ctx context.Context, paymentRecord *entities.PaymentRecord, retry *entities.PaymentRetry) (*dto.PaymentProcessResult, error) {
	// This is a simplified implementation
	// In reality, this would call the actual payment gateway

	result := &dto.PaymentProcessResult{
		PaymentRecordID: paymentRecord.ID,
	}

	// Simulate payment processing
	// For demonstration, we'll simulate different scenarios based on attempt number
	switch retry.AttemptNumber {
	case 0:
		// First retry - simulate temporary failure
		result.Status = "failed"
		result.ErrorCode = "NETWORK_ERROR"
		result.ErrorMessage = "Network timeout"
		return result, fmt.Errorf("network timeout")
	case 1:
		// Second retry - simulate gateway error
		result.Status = "failed"
		result.ErrorCode = "GATEWAY_UNAVAILABLE"
		result.ErrorMessage = "Gateway temporarily unavailable"
		return result, fmt.Errorf("gateway unavailable")
	default:
		// Third retry - simulate success
		result.Status = "completed"
		result.TransactionID = fmt.Sprintf("TXN_%d_%d", paymentRecord.ID, time.Now().Unix())
		now := time.Now()
		result.PaidAt = &now
		return result, nil
	}
}

func (s *paymentRetryService) getDefaultStrategyForGateway(gateway string) *entities.RetryStrategyConfig {
	if strategy, exists := entities.DefaultRetryStrategies[gateway]; exists {
		return &strategy
	}
	return s.config.DefaultStrategy
}

func (s *paymentRetryService) classifyFailureDefault(errorCode, errorMessage string) string {
	errorMessage = strings.ToLower(errorMessage)
	errorCode = strings.ToUpper(errorCode)

	// Permanent failure patterns
	permanentPatterns := []string{
		"invalid card", "card expired", "insufficient funds",
		"card declined", "invalid account", "account closed",
		"invalid cvv", "invalid expiry", "fraud",
	}

	for _, pattern := range permanentPatterns {
		if strings.Contains(errorMessage, pattern) {
			return constants.FailureTypePermanent
		}
	}

	// Network failure patterns
	networkPatterns := []string{
		"timeout", "network", "connection", "dns", "ssl",
	}

	for _, pattern := range networkPatterns {
		if strings.Contains(errorMessage, pattern) {
			return constants.FailureTypeNetwork
		}
	}

	// Gateway failure patterns
	gatewayPatterns := []string{
		"gateway", "service unavailable", "server error", "maintenance",
	}

	for _, pattern := range gatewayPatterns {
		if strings.Contains(errorMessage, pattern) {
			return constants.FailureTypeGateway
		}
	}

	// Default to temporary
	return constants.FailureTypeTemporary
}

func (s *paymentRetryService) generateFailureRecommendations(patterns []*dto.FailurePattern) []string {
	recommendations := []string{}

	for _, pattern := range patterns {
		if pattern.Count > 10 && pattern.SuccessRate < 50 {
			recommendations = append(recommendations,
				fmt.Sprintf("High failure rate for %s: %s - consider reviewing gateway configuration",
					pattern.ErrorType, pattern.FailureReason))
		}
	}

	return recommendations
}

func (s *paymentRetryService) determineGatewayHealth(successRate float64, activeRetries int) string {
	if successRate < 70 || activeRetries > 100 {
		return "critical"
	} else if successRate < 85 || activeRetries > 50 {
		return "degraded"
	}
	return "healthy"
}

func (s *paymentRetryService) generateHealthAlerts(activeRetries, overdueRetries int64, successRate float64) []string {
	alerts := []string{}

	if activeRetries > int64(s.config.MonitoringSettings.AlertThresholds.MaxPendingRetries) {
		alerts = append(alerts, fmt.Sprintf("High number of pending retries: %d", activeRetries))
	}

	if overdueRetries > int64(s.config.MonitoringSettings.AlertThresholds.MaxOverdueRetries) {
		alerts = append(alerts, fmt.Sprintf("High number of overdue retries: %d", overdueRetries))
	}

	if successRate < s.config.MonitoringSettings.AlertThresholds.MinSuccessRate {
		alerts = append(alerts, fmt.Sprintf("Low retry success rate: %.2f%%", successRate))
	}

	return alerts
}

func (s *paymentRetryService) generateSystemRecommendations(gatewayHealth []*dto.GatewayHealthMetric) []string {
	recommendations := []string{}

	for _, health := range gatewayHealth {
		if health.HealthStatus == "critical" {
			recommendations = append(recommendations,
				fmt.Sprintf("Gateway %s requires immediate attention", health.Gateway))
		}
	}

	return recommendations
}

func convertRetryHistorySlice(history []*entities.PaymentRetryHistory) []*dto.PaymentRetryHistoryResponse {
	if history == nil {
		return nil
	}
	result := make([]*dto.PaymentRetryHistoryResponse, len(history))
	for i, h := range history {
		result[i] = dto.ToPaymentRetryHistoryResponse(h)
	}
	return result
}

// convertRetrySliceToDTO converts entity retry slice to DTO slice
func convertRetrySliceToDTO(retries []*entities.PaymentRetry) []*dto.PaymentRetryResponse {
	if retries == nil {
		return nil
	}
	result := make([]*dto.PaymentRetryResponse, len(retries))
	for i, r := range retries {
		result[i] = dto.ToPaymentRetryResponse(r)
	}
	return result
}
func getDefaultRetryConfiguration() *dto.RetryConfiguration {
	return &dto.RetryConfiguration{
		Enabled:              true,
		MaxConcurrentRetries: 100,
		ProcessingInterval:   (time.Minute * 5).String(),
		HealthCheckInterval:  (time.Minute * 15).String(),
		DefaultStrategy: &entities.RetryStrategyConfig{
			MaxAttempts:      3,
			InitialDelay:     3600,  // 1 hour
			MaxDelay:         86400, // 24 hours
			BackoffFactor:    2.0,
			Strategy:         constants.RetryStrategyExponential,
			FailureTypes:     []string{constants.FailureTypeTemporary, constants.FailureTypeNetwork, constants.FailureTypeGateway},
			TimeoutSeconds:   30,
			EnableAfterHours: true,
			MaxConcurrent:    5,
		},
		GatewayStrategies: make(map[string]*entities.RetryStrategyConfig),
		NotificationSettings: &dto.RetryNotificationSettings{
			Enabled:             true,
			NotifyOnFailure:     true,
			NotifyOnSuccess:     false,
			NotifyOnMaxAttempts: true,
		},
		MonitoringSettings: &dto.RetryMonitoringSettings{
			MetricsEnabled:     true,
			AlertsEnabled:      true,
			HealthCheckEnabled: true,
			LogLevel:           "info",
			RetentionPeriod:    (time.Hour * 24 * 30).String(), // 30 days
			AlertThresholds: &dto.AlertThresholds{
				MaxPendingRetries:  50,
				MaxOverdueRetries:  10,
				MinSuccessRate:     80.0,
				MaxAverageAttempts: 2.5,
				MaxProcessingDelay: 3600, // 1 hour
			},
		},
	}
}
