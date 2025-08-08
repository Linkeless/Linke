package notification

import (
	"context"
	"fmt"
	"sync"
	"time"

	"linke/internal/shared/logger"
	"linke/internal/shared/queue"
)

// BatchNotificationRequest represents a batch notification request
type BatchNotificationRequest struct {
	Recipients   []NotificationRecipient `json:"recipients" validate:"required,dive"`
	Template     string                  `json:"template" validate:"required"`
	GlobalVars   map[string]string       `json:"global_variables,omitempty"`
	Priority     NotificationPriority    `json:"priority"`
	ScheduleTime *time.Time              `json:"schedule_time,omitempty"`
	RateLimit    *RateLimitConfig        `json:"rate_limit,omitempty"`
	BatchID      string                  `json:"batch_id,omitempty"`
	Description  string                  `json:"description,omitempty"`
}

// NotificationRecipient represents a single recipient in a batch
type NotificationRecipient struct {
	UserID           uint                  `json:"user_id" validate:"required"`
	Email            string                `json:"email,omitempty"`
	Phone            string                `json:"phone,omitempty"`
	TelegramChatID   int64                 `json:"telegram_chat_id,omitempty"`
	TelegramUsername string                `json:"telegram_username,omitempty"`
	Channels         []NotificationChannel `json:"channels" validate:"required"`
	Variables        map[string]string     `json:"variables,omitempty"`
	Preferences      *UserPreferences      `json:"preferences,omitempty"`
}

// UserPreferences represents user notification preferences
type UserPreferences struct {
	Enabled     bool                         `json:"enabled"`
	Channels    map[NotificationChannel]bool `json:"channels"`
	QuietHours  *QuietHoursConfig            `json:"quiet_hours,omitempty"`
	Frequency   string                       `json:"frequency"` // immediate, hourly, daily
	MinPriority NotificationPriority         `json:"min_priority"`
	Language    string                       `json:"language"`
	Timezone    string                       `json:"timezone"`
}

// QuietHoursConfig represents quiet hours settings
type QuietHoursConfig struct {
	Enabled   bool   `json:"enabled"`
	StartHour int    `json:"start_hour" validate:"min=0,max=23"`
	EndHour   int    `json:"end_hour" validate:"min=0,max=23"`
	Timezone  string `json:"timezone"`
}

// RateLimitConfig represents rate limiting configuration
type RateLimitConfig struct {
	MaxPerSecond   int `json:"max_per_second" validate:"min=1"`
	MaxPerMinute   int `json:"max_per_minute" validate:"min=1"`
	BatchSize      int `json:"batch_size" validate:"min=1,max=1000"`
	MaxConcurrency int `json:"max_concurrency" validate:"min=1,max=100"`
}

// BatchNotificationResult represents the result of a batch operation
type BatchNotificationResult struct {
	BatchID           string         `json:"batch_id"`
	TotalRecipients   int            `json:"total_recipients"`
	TotalBatches      int            `json:"total_batches"`
	SuccessfulBatches int            `json:"successful_batches"`
	FailedBatches     int            `json:"failed_batches"`
	TotalSent         int            `json:"total_sent"`
	TotalFailed       int            `json:"total_failed"`
	StartTime         time.Time      `json:"start_time"`
	EndTime           time.Time      `json:"end_time"`
	Duration          time.Duration  `json:"duration" swaggertype:"string"`
	Errors            []error        `json:"errors,omitempty"`
	BatchResults      []*BatchResult `json:"batch_results,omitempty"`
}

// BatchResult represents the result of processing a single batch
type BatchResult struct {
	BatchIndex     int                     `json:"batch_index"`
	Recipients     []NotificationRecipient `json:"recipients"`
	Sent           int                     `json:"sent"`
	Failed         int                     `json:"failed"`
	ProcessingTime time.Duration           `json:"processing_time" swaggertype:"string"`
	Results        []*NotificationResult   `json:"results,omitempty"`
	Error          string                  `json:"error,omitempty"`
}

// BatchNotificationService provides batch notification processing
type BatchNotificationService struct {
	baseService NotificationService
	taskQueue   *queue.TaskQueue
	logger      logger.Logger
}

// NewBatchNotificationService creates a new batch notification service
func NewBatchNotificationService(
	baseService NotificationService,
	taskQueue *queue.TaskQueue,
	logger logger.Logger,
) *BatchNotificationService {
	return &BatchNotificationService{
		baseService: baseService,
		taskQueue:   taskQueue,
		logger:      logger,
	}
}

// SendBatch processes a batch notification request
func (s *BatchNotificationService) SendBatch(ctx context.Context, req *BatchNotificationRequest) (*BatchNotificationResult, error) {
	if len(req.Recipients) == 0 {
		return nil, fmt.Errorf("no recipients specified")
	}

	// Generate batch ID if not provided
	if req.BatchID == "" {
		req.BatchID = fmt.Sprintf("batch_%d", time.Now().Unix())
	}

	// Apply default rate limiting if not specified
	rateLimit := req.RateLimit
	if rateLimit == nil {
		rateLimit = &RateLimitConfig{
			MaxPerSecond:   10,
			MaxPerMinute:   100,
			BatchSize:      50,
			MaxConcurrency: 5,
		}
	}

	s.logger.Info("Starting batch notification processing",
		logger.String("batch_id", req.BatchID),
		logger.Int("total_recipients", len(req.Recipients)),
		logger.String("template", req.Template),
		logger.String("priority", req.Priority.String()))

	// Create result tracking
	result := &BatchNotificationResult{
		BatchID:         req.BatchID,
		TotalRecipients: len(req.Recipients),
		StartTime:       time.Now(),
		Errors:          make([]error, 0),
		BatchResults:    make([]*BatchResult, 0),
	}

	// Check if this should be scheduled
	if req.ScheduleTime != nil && req.ScheduleTime.After(time.Now()) {
		return s.scheduleForLater(ctx, req)
	}

	// Create batches
	batches := s.createBatches(req.Recipients, rateLimit.BatchSize)
	result.TotalBatches = len(batches)

	// Process batches with concurrency control
	return s.processBatchesConcurrently(ctx, batches, req, rateLimit, result)
}

// createBatches splits recipients into smaller batches
func (s *BatchNotificationService) createBatches(recipients []NotificationRecipient, batchSize int) [][]NotificationRecipient {
	var batches [][]NotificationRecipient

	for i := 0; i < len(recipients); i += batchSize {
		end := i + batchSize
		if end > len(recipients) {
			end = len(recipients)
		}
		batches = append(batches, recipients[i:end])
	}

	return batches
}

// processBatchesConcurrently processes multiple batches concurrently with rate limiting
func (s *BatchNotificationService) processBatchesConcurrently(
	ctx context.Context,
	batches [][]NotificationRecipient,
	req *BatchNotificationRequest,
	rateLimit *RateLimitConfig,
	result *BatchNotificationResult,
) (*BatchNotificationResult, error) {
	// Channel for rate limiting
	rateLimiter := make(chan struct{}, rateLimit.MaxConcurrency)

	// Channel for collecting results
	resultChan := make(chan *BatchResult, len(batches))
	errorChan := make(chan error, len(batches))

	var wg sync.WaitGroup

	// Process each batch
	for batchIndex, batch := range batches {
		wg.Add(1)

		go func(idx int, recipients []NotificationRecipient) {
			defer wg.Done()

			// Acquire rate limit token
			rateLimiter <- struct{}{}
			defer func() { <-rateLimiter }()

			// Process the batch
			batchResult, err := s.processBatch(ctx, recipients, req, idx)
			if err != nil {
				errorChan <- err
				return
			}

			resultChan <- batchResult

			// Apply inter-batch delay for rate limiting
			if rateLimit.MaxPerSecond > 0 {
				delay := time.Duration(len(recipients)*1000/rateLimit.MaxPerSecond) * time.Millisecond
				time.Sleep(delay)
			}
		}(batchIndex, batch)
	}

	// Close channels when all goroutines are done
	go func() {
		wg.Wait()
		close(resultChan)
		close(errorChan)
	}()

	// Collect results
	for {
		select {
		case batchResult, ok := <-resultChan:
			if !ok {
				resultChan = nil
			} else {
				result.BatchResults = append(result.BatchResults, batchResult)
				if batchResult.Error == "" {
					result.SuccessfulBatches++
				} else {
					result.FailedBatches++
				}
				result.TotalSent += batchResult.Sent
				result.TotalFailed += batchResult.Failed
			}
		case err, ok := <-errorChan:
			if !ok {
				errorChan = nil
			} else {
				result.Errors = append(result.Errors, err)
				result.FailedBatches++
			}
		}

		if resultChan == nil && errorChan == nil {
			break
		}
	}

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)

	s.logger.Info("Batch notification processing completed",
		logger.String("batch_id", req.BatchID),
		logger.Int("successful_batches", result.SuccessfulBatches),
		logger.Int("failed_batches", result.FailedBatches),
		logger.Int("total_sent", result.TotalSent),
		logger.Int("total_failed", result.TotalFailed),
		logger.String("duration", result.Duration.String()))

	return result, nil
}

// processBatch processes a single batch of notifications
func (s *BatchNotificationService) processBatch(
	ctx context.Context,
	recipients []NotificationRecipient,
	req *BatchNotificationRequest,
	batchIndex int,
) (*BatchResult, error) {
	startTime := time.Now()

	batchResult := &BatchResult{
		BatchIndex: batchIndex,
		Recipients: recipients,
		Results:    make([]*NotificationResult, 0, len(recipients)),
	}

	// Process each recipient in the batch
	for _, recipient := range recipients {
		// Skip if user preferences don't allow this notification
		if !s.shouldSendToRecipient(recipient, req.Priority) {
			s.logger.Debug("Skipping notification due to user preferences",
				logger.Uint("user_id", recipient.UserID),
				logger.String("batch_id", req.BatchID))
			continue
		}

		// Merge global and recipient-specific variables
		variables := make(map[string]string)
		for k, v := range req.GlobalVars {
			variables[k] = v
		}
		for k, v := range recipient.Variables {
			variables[k] = v
		}

		// Create individual notification request
		notificationReq := &NotificationRequest{
			UserID:           recipient.UserID,
			Email:            recipient.Email,
			Phone:            recipient.Phone,
			TelegramChatID:   recipient.TelegramChatID,
			TelegramUsername: recipient.TelegramUsername,
			Channels:         recipient.Channels,
			Subject:          fmt.Sprintf("Batch: %s", req.BatchID),
			Template:         req.Template,
			Variables:        variables,
			Priority:         req.Priority,
			EventType:        "batch_notification",
			EventID:          fmt.Sprintf("%s_%d_%d", req.BatchID, batchIndex, recipient.UserID),
		}

		// Send notification
		results, err := s.baseService.Send(ctx, notificationReq)
		if err != nil {
			batchResult.Failed++
			s.logger.Error("Failed to send notification in batch",
				logger.Uint("user_id", recipient.UserID),
				logger.String("batch_id", req.BatchID),
				logger.ErrorField(err))
			continue
		}

		// Count successful notifications
		for _, result := range results {
			batchResult.Results = append(batchResult.Results, result)
			if result.Success {
				batchResult.Sent++
			} else {
				batchResult.Failed++
			}
		}
	}

	batchResult.ProcessingTime = time.Since(startTime)

	s.logger.Debug("Batch processed",
		logger.Int("batch_index", batchIndex),
		logger.String("batch_id", req.BatchID),
		logger.Int("sent", batchResult.Sent),
		logger.Int("failed", batchResult.Failed),
		logger.String("processing_time", batchResult.ProcessingTime.String()))

	return batchResult, nil
}

// shouldSendToRecipient checks if notification should be sent based on user preferences
func (s *BatchNotificationService) shouldSendToRecipient(recipient NotificationRecipient, priority NotificationPriority) bool {
	if recipient.Preferences == nil {
		return true // No preferences set, allow all
	}

	prefs := recipient.Preferences

	// Check if notifications are globally enabled
	if !prefs.Enabled {
		return false
	}

	// Check minimum priority
	if !s.isPriorityAllowed(priority, prefs.MinPriority) {
		return false
	}

	// Check quiet hours
	if s.isInQuietHours(prefs.QuietHours) {
		return priority == PriorityUrgent // Only urgent notifications during quiet hours
	}

	// Check if any of the recipient's channels are enabled
	hasEnabledChannel := false
	for _, channel := range recipient.Channels {
		if enabled, exists := prefs.Channels[channel]; !exists || enabled {
			hasEnabledChannel = true
			break
		}
	}

	return hasEnabledChannel
}

// isPriorityAllowed checks if the notification priority meets the minimum requirement
func (s *BatchNotificationService) isPriorityAllowed(notifPriority, minPriority NotificationPriority) bool {
	priorityLevels := map[NotificationPriority]int{
		PriorityLow:    1,
		PriorityNormal: 2,
		PriorityHigh:   3,
		PriorityUrgent: 4,
	}

	notifLevel, exists1 := priorityLevels[notifPriority]
	minLevel, exists2 := priorityLevels[minPriority]

	if !exists1 || !exists2 {
		return true // Default to allow if priority not found
	}

	return notifLevel >= minLevel
}

// isInQuietHours checks if current time is within user's quiet hours
func (s *BatchNotificationService) isInQuietHours(quietHours *QuietHoursConfig) bool {
	if quietHours == nil || !quietHours.Enabled {
		return false
	}

	// TODO: Implement proper timezone handling
	now := time.Now()
	currentHour := now.Hour()

	start := quietHours.StartHour
	end := quietHours.EndHour

	if start <= end {
		return currentHour >= start && currentHour < end
	} else {
		// Quiet hours span midnight
		return currentHour >= start || currentHour < end
	}
}

// scheduleForLater schedules a batch notification for later processing
func (s *BatchNotificationService) scheduleForLater(ctx context.Context, req *BatchNotificationRequest) (*BatchNotificationResult, error) {
	// For now, we'll return an error indicating scheduled notifications are not yet supported
	// In a full implementation, this would integrate with a job scheduler like asynq with delay support
	s.logger.Warn("Scheduled batch notifications not yet implemented, processing immediately",
		logger.String("batch_id", req.BatchID),
		logger.String("schedule_time", req.ScheduleTime.String()))

	// Process immediately instead of scheduling
	req.ScheduleTime = nil
	return s.SendBatch(ctx, req)
}
