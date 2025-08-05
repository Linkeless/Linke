package events

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"linke/internal/shared/logger"
	"linke/internal/shared/queue"

	"github.com/hibiken/asynq"
)

// DeadLetterReason represents the reason why an event was sent to dead letter queue
type DeadLetterReason string

const (
	DeadLetterReasonMaxRetriesExceeded DeadLetterReason = "max_retries_exceeded"
	DeadLetterReasonTimeout            DeadLetterReason = "timeout"
	DeadLetterReasonCircuitBreakerOpen DeadLetterReason = "circuit_breaker_open"
	DeadLetterReasonValidationFailed   DeadLetterReason = "validation_failed"
	DeadLetterReasonHandlerNotFound    DeadLetterReason = "handler_not_found"
	DeadLetterReasonSystemError        DeadLetterReason = "system_error"
)

// String returns the string representation of the dead letter reason
func (r DeadLetterReason) String() string {
	return string(r)
}

// DeadLetterEvent represents an event in the dead letter queue
type DeadLetterEvent struct {
	ID                string            `json:"id"`
	OriginalEventID   string            `json:"original_event_id"`
	OriginalEventType string            `json:"original_event_type"`
	EventData         string            `json:"event_data"`
	Reason            DeadLetterReason  `json:"reason"`
	Error             string            `json:"error"`
	RetryCount        int               `json:"retry_count"`
	MaxRetries        int               `json:"max_retries"`
	CreatedAt         time.Time         `json:"created_at"`
	LastRetryAt       *time.Time        `json:"last_retry_at,omitempty"`
	NextRetryAt       *time.Time        `json:"next_retry_at,omitempty"`
	Metadata          map[string]string `json:"metadata"`
	Status            DeadLetterStatus  `json:"status"`
}

// DeadLetterStatus represents the status of a dead letter event
type DeadLetterStatus string

const (
	DeadLetterStatusPending    DeadLetterStatus = "pending"
	DeadLetterStatusProcessing DeadLetterStatus = "processing"
	DeadLetterStatusRetrying   DeadLetterStatus = "retrying"
	DeadLetterStatusResolved   DeadLetterStatus = "resolved"
	DeadLetterStatusAbandoned  DeadLetterStatus = "abandoned"
)

// DeadLetterQueue manages dead letter events
type DeadLetterQueue interface {
	Add(ctx context.Context, dlEvent *DeadLetterEvent) error
	Get(ctx context.Context, id string) (*DeadLetterEvent, error)
	List(ctx context.Context, filters DeadLetterFilters) ([]*DeadLetterEvent, error)
	Retry(ctx context.Context, id string) error
	Resolve(ctx context.Context, id string) error
	Abandon(ctx context.Context, id string) error
	GetStats(ctx context.Context) (*DeadLetterStats, error)
}

// DeadLetterFilters provides filtering options for dead letter queries
type DeadLetterFilters struct {
	Status    []DeadLetterStatus `json:"status,omitempty"`
	Reason    []DeadLetterReason `json:"reason,omitempty"`
	EventType string             `json:"event_type,omitempty"`
	FromDate  *time.Time         `json:"from_date,omitempty"`
	ToDate    *time.Time         `json:"to_date,omitempty"`
	Limit     int                `json:"limit,omitempty"`
	Offset    int                `json:"offset,omitempty"`
}

// DeadLetterStats provides statistics about the dead letter queue
type DeadLetterStats struct {
	TotalEvents        int64                      `json:"total_events"`
	EventsByStatus     map[DeadLetterStatus]int64 `json:"events_by_status"`
	EventsByReason     map[DeadLetterReason]int64 `json:"events_by_reason"`
	EventsByType       map[string]int64           `json:"events_by_type"`
	AverageRetryCount  float64                    `json:"average_retry_count"`
	OldestPendingEvent *time.Time                 `json:"oldest_pending_event,omitempty"`
	RecentEventsCount  int64                      `json:"recent_events_count"`
}

// InMemoryDeadLetterQueue provides an in-memory implementation of DeadLetterQueue
type InMemoryDeadLetterQueue struct {
	events map[string]*DeadLetterEvent
	mutex  sync.RWMutex
	logger logger.Logger
}

// NewInMemoryDeadLetterQueue creates a new in-memory dead letter queue
func NewInMemoryDeadLetterQueue() *InMemoryDeadLetterQueue {
	return &InMemoryDeadLetterQueue{
		events: make(map[string]*DeadLetterEvent),
		logger: logger.GetGlobalLogger(),
	}
}

// Add adds an event to the dead letter queue
func (dlq *InMemoryDeadLetterQueue) Add(ctx context.Context, dlEvent *DeadLetterEvent) error {
	dlq.mutex.Lock()
	defer dlq.mutex.Unlock()

	if dlEvent.ID == "" {
		dlEvent.ID = generateEventID()
	}

	dlEvent.CreatedAt = time.Now()
	dlEvent.Status = DeadLetterStatusPending

	dlq.events[dlEvent.ID] = dlEvent

	dlq.logger.Warn("Event added to dead letter queue",
		logger.String("dead_letter_id", dlEvent.ID),
		logger.String("original_event_id", dlEvent.OriginalEventID),
		logger.String("original_event_type", dlEvent.OriginalEventType),
		logger.String("reason", dlEvent.Reason.String()),
		logger.String("error", dlEvent.Error),
	)

	return nil
}

// Get retrieves a dead letter event by ID
func (dlq *InMemoryDeadLetterQueue) Get(ctx context.Context, id string) (*DeadLetterEvent, error) {
	dlq.mutex.RLock()
	defer dlq.mutex.RUnlock()

	event, exists := dlq.events[id]
	if !exists {
		return nil, fmt.Errorf("dead letter event not found: %s", id)
	}

	return event, nil
}

// List retrieves dead letter events based on filters
func (dlq *InMemoryDeadLetterQueue) List(ctx context.Context, filters DeadLetterFilters) ([]*DeadLetterEvent, error) {
	dlq.mutex.RLock()
	defer dlq.mutex.RUnlock()

	var filteredEvents []*DeadLetterEvent

	for _, event := range dlq.events {
		if dlq.matchesFilters(event, filters) {
			filteredEvents = append(filteredEvents, event)
		}
	}

	// Apply limit and offset
	start := filters.Offset
	if start > len(filteredEvents) {
		start = len(filteredEvents)
	}

	end := len(filteredEvents)
	if filters.Limit > 0 && start+filters.Limit < end {
		end = start + filters.Limit
	}

	return filteredEvents[start:end], nil
}

// Retry marks a dead letter event for retry
func (dlq *InMemoryDeadLetterQueue) Retry(ctx context.Context, id string) error {
	dlq.mutex.Lock()
	defer dlq.mutex.Unlock()

	event, exists := dlq.events[id]
	if !exists {
		return fmt.Errorf("dead letter event not found: %s", id)
	}

	if event.Status == DeadLetterStatusResolved || event.Status == DeadLetterStatusAbandoned {
		return fmt.Errorf("cannot retry event with status: %s", event.Status)
	}

	event.Status = DeadLetterStatusRetrying
	now := time.Now()
	event.LastRetryAt = &now
	event.RetryCount++

	dlq.logger.Info("Dead letter event marked for retry",
		logger.String("dead_letter_id", id),
		logger.String("original_event_id", event.OriginalEventID),
		logger.Int("retry_count", event.RetryCount),
	)

	return nil
}

// Resolve marks a dead letter event as resolved
func (dlq *InMemoryDeadLetterQueue) Resolve(ctx context.Context, id string) error {
	dlq.mutex.Lock()
	defer dlq.mutex.Unlock()

	event, exists := dlq.events[id]
	if !exists {
		return fmt.Errorf("dead letter event not found: %s", id)
	}

	event.Status = DeadLetterStatusResolved

	dlq.logger.Info("Dead letter event resolved",
		logger.String("dead_letter_id", id),
		logger.String("original_event_id", event.OriginalEventID),
	)

	return nil
}

// Abandon marks a dead letter event as abandoned
func (dlq *InMemoryDeadLetterQueue) Abandon(ctx context.Context, id string) error {
	dlq.mutex.Lock()
	defer dlq.mutex.Unlock()

	event, exists := dlq.events[id]
	if !exists {
		return fmt.Errorf("dead letter event not found: %s", id)
	}

	event.Status = DeadLetterStatusAbandoned

	dlq.logger.Info("Dead letter event abandoned",
		logger.String("dead_letter_id", id),
		logger.String("original_event_id", event.OriginalEventID),
	)

	return nil
}

// GetStats returns statistics about the dead letter queue
func (dlq *InMemoryDeadLetterQueue) GetStats(ctx context.Context) (*DeadLetterStats, error) {
	dlq.mutex.RLock()
	defer dlq.mutex.RUnlock()

	stats := &DeadLetterStats{
		EventsByStatus: make(map[DeadLetterStatus]int64),
		EventsByReason: make(map[DeadLetterReason]int64),
		EventsByType:   make(map[string]int64),
	}

	var totalRetries int64
	var oldestPending *time.Time
	recentCutoff := time.Now().Add(-24 * time.Hour)

	for _, event := range dlq.events {
		stats.TotalEvents++
		stats.EventsByStatus[event.Status]++
		stats.EventsByReason[event.Reason]++
		stats.EventsByType[event.OriginalEventType]++

		totalRetries += int64(event.RetryCount)

		if event.Status == DeadLetterStatusPending {
			if oldestPending == nil || event.CreatedAt.Before(*oldestPending) {
				oldestPending = &event.CreatedAt
			}
		}

		if event.CreatedAt.After(recentCutoff) {
			stats.RecentEventsCount++
		}
	}

	if stats.TotalEvents > 0 {
		stats.AverageRetryCount = float64(totalRetries) / float64(stats.TotalEvents)
	}

	stats.OldestPendingEvent = oldestPending

	return stats, nil
}

// matchesFilters checks if an event matches the given filters
func (dlq *InMemoryDeadLetterQueue) matchesFilters(event *DeadLetterEvent, filters DeadLetterFilters) bool {
	// Status filter
	if len(filters.Status) > 0 {
		statusMatch := false
		for _, status := range filters.Status {
			if event.Status == status {
				statusMatch = true
				break
			}
		}
		if !statusMatch {
			return false
		}
	}

	// Reason filter
	if len(filters.Reason) > 0 {
		reasonMatch := false
		for _, reason := range filters.Reason {
			if event.Reason == reason {
				reasonMatch = true
				break
			}
		}
		if !reasonMatch {
			return false
		}
	}

	// Event type filter
	if filters.EventType != "" && event.OriginalEventType != filters.EventType {
		return false
	}

	// Date filters
	if filters.FromDate != nil && event.CreatedAt.Before(*filters.FromDate) {
		return false
	}
	if filters.ToDate != nil && event.CreatedAt.After(*filters.ToDate) {
		return false
	}

	return true
}

// DeadLetterHandler handles events that have been moved to the dead letter queue
type DeadLetterHandler struct {
	deadLetterQueue DeadLetterQueue
	eventBus        EventBus
	taskQueue       *queue.TaskQueue
	retryPolicy     DeadLetterRetryPolicy
	logger          logger.Logger
}

// DeadLetterRetryPolicy defines retry behavior for dead letter events
type DeadLetterRetryPolicy struct {
	MaxRetries        int           `json:"max_retries"`
	InitialDelay      time.Duration `json:"initial_delay"`
	MaxDelay          time.Duration `json:"max_delay"`
	BackoffMultiplier float64       `json:"backoff_multiplier"`
	EnableJitter      bool          `json:"enable_jitter"`
}

// DefaultDeadLetterRetryPolicy returns a sensible default retry policy
func DefaultDeadLetterRetryPolicy() DeadLetterRetryPolicy {
	return DeadLetterRetryPolicy{
		MaxRetries:        3,
		InitialDelay:      time.Minute * 5,
		MaxDelay:          time.Hour * 2,
		BackoffMultiplier: 2.0,
		EnableJitter:      true,
	}
}

// NewDeadLetterHandler creates a new dead letter handler
func NewDeadLetterHandler(deadLetterQueue DeadLetterQueue, eventBus EventBus, taskQueue *queue.TaskQueue, retryPolicy DeadLetterRetryPolicy) *DeadLetterHandler {
	return &DeadLetterHandler{
		deadLetterQueue: deadLetterQueue,
		eventBus:        eventBus,
		taskQueue:       taskQueue,
		retryPolicy:     retryPolicy,
		logger:          logger.GetGlobalLogger(),
	}
}

// HandleFailedEvent handles an event that failed processing
func (dlh *DeadLetterHandler) HandleFailedEvent(ctx context.Context, event Event, reason DeadLetterReason, err error, retryCount int) error {
	// Serialize the event
	eventData, serializeErr := SerializeEvent(event)
	if serializeErr != nil {
		dlh.logger.Error("Failed to serialize failed event",
			logger.String("event_id", event.EventID()),
			logger.String("event_type", event.EventType()),
			logger.ErrorField(serializeErr),
		)
		return fmt.Errorf("failed to serialize event for dead letter queue: %w", serializeErr)
	}

	// Create dead letter event
	dlEvent := &DeadLetterEvent{
		OriginalEventID:   event.EventID(),
		OriginalEventType: event.EventType(),
		EventData:         string(eventData),
		Reason:            reason,
		Error:             err.Error(),
		RetryCount:        retryCount,
		MaxRetries:        dlh.retryPolicy.MaxRetries,
		Metadata:          make(map[string]string),
	}

	// Add correlation ID if available
	if baseEvent, ok := event.(*BaseEvent); ok {
		if correlationID, exists := baseEvent.GetMetadata("correlation_id"); exists {
			if id, ok := correlationID.(string); ok {
				dlEvent.Metadata["correlation_id"] = id
			}
		}
	}

	// Add to dead letter queue
	if err := dlh.deadLetterQueue.Add(ctx, dlEvent); err != nil {
		dlh.logger.Error("Failed to add event to dead letter queue",
			logger.String("event_id", event.EventID()),
			logger.ErrorField(err),
		)
		return fmt.Errorf("failed to add event to dead letter queue: %w", err)
	}

	// Schedule retry if applicable
	if dlEvent.RetryCount < dlh.retryPolicy.MaxRetries {
		if err := dlh.scheduleRetry(ctx, dlEvent); err != nil {
			dlh.logger.Error("Failed to schedule retry for dead letter event",
				logger.String("dead_letter_id", dlEvent.ID),
				logger.ErrorField(err),
			)
		}
	}

	return nil
}

// scheduleRetry schedules a retry for a dead letter event
func (dlh *DeadLetterHandler) scheduleRetry(ctx context.Context, dlEvent *DeadLetterEvent) error {
	// Calculate retry delay
	delay := dlh.calculateRetryDelay(dlEvent.RetryCount)
	nextRetryAt := time.Now().Add(delay)
	dlEvent.NextRetryAt = &nextRetryAt

	// Create retry task
	task := &queue.Task{
		ID:       fmt.Sprintf("dead-letter-retry-%s", dlEvent.ID),
		Type:     TaskTypeDeadLetterRetry,
		MaxRetry: 1, // Don't retry the retry task itself
		Payload: map[string]any{
			"dead_letter_id": dlEvent.ID,
		},
	}

	// Schedule the task (using a simple approach since EnqueueWithDelay doesn't exist)
	// In a real implementation, you would use a proper delayed queue mechanism
	if err := dlh.taskQueue.Enqueue(ctx, "dead-letter-retry", task); err != nil {
		return fmt.Errorf("failed to schedule retry task: %w", err)
	}

	dlh.logger.Info("Dead letter retry scheduled",
		logger.String("dead_letter_id", dlEvent.ID),
		logger.String("original_event_id", dlEvent.OriginalEventID),
		logger.Any("retry_delay", delay),
		logger.Any("next_retry_at", nextRetryAt),
	)

	return nil
}

// calculateRetryDelay calculates the delay for a retry attempt
func (dlh *DeadLetterHandler) calculateRetryDelay(retryCount int) time.Duration {
	delay := dlh.retryPolicy.InitialDelay

	// Apply exponential backoff
	for i := 0; i < retryCount; i++ {
		delay = time.Duration(float64(delay) * dlh.retryPolicy.BackoffMultiplier)
	}

	// Apply maximum delay
	if delay > dlh.retryPolicy.MaxDelay {
		delay = dlh.retryPolicy.MaxDelay
	}

	// Apply jitter if enabled
	if dlh.retryPolicy.EnableJitter {
		jitter := time.Duration(float64(delay) * 0.1) // 10% jitter
		jitterFactor := (2.0*float64(time.Now().UnixNano()%1000)/1000.0 - 1.0)
		delay += time.Duration(float64(jitter) * jitterFactor)
	}

	return delay
}

// RetryDeadLetterEvent retries a specific dead letter event
func (dlh *DeadLetterHandler) RetryDeadLetterEvent(ctx context.Context, deadLetterID string) error {
	// Get the dead letter event
	dlEvent, err := dlh.deadLetterQueue.Get(ctx, deadLetterID)
	if err != nil {
		return fmt.Errorf("failed to get dead letter event: %w", err)
	}

	// Deserialize the original event
	envelope, err := DeserializeEvent([]byte(dlEvent.EventData))
	if err != nil {
		return fmt.Errorf("failed to deserialize dead letter event: %w", err)
	}

	// Mark as retrying
	if err := dlh.deadLetterQueue.Retry(ctx, deadLetterID); err != nil {
		return fmt.Errorf("failed to mark dead letter event as retrying: %w", err)
	}

	// Attempt to republish the event
	if err := dlh.eventBus.Publish(ctx, envelope.Event); err != nil {
		dlh.logger.Error("Dead letter event retry failed",
			logger.String("dead_letter_id", deadLetterID),
			logger.String("original_event_id", dlEvent.OriginalEventID),
			logger.ErrorField(err),
		)

		// Handle retry failure
		if dlEvent.RetryCount >= dlh.retryPolicy.MaxRetries {
			// Max retries exceeded, abandon the event
			if abandonErr := dlh.deadLetterQueue.Abandon(ctx, deadLetterID); abandonErr != nil {
				dlh.logger.Error("Failed to abandon dead letter event",
					logger.String("dead_letter_id", deadLetterID),
					logger.ErrorField(abandonErr),
				)
			}
		} else {
			// Schedule another retry
			if scheduleErr := dlh.scheduleRetry(ctx, dlEvent); scheduleErr != nil {
				dlh.logger.Error("Failed to schedule retry after failed attempt",
					logger.String("dead_letter_id", deadLetterID),
					logger.ErrorField(scheduleErr),
				)
			}
		}

		return err
	}

	// Retry successful, resolve the dead letter event
	if err := dlh.deadLetterQueue.Resolve(ctx, deadLetterID); err != nil {
		dlh.logger.Error("Failed to resolve dead letter event after successful retry",
			logger.String("dead_letter_id", deadLetterID),
			logger.ErrorField(err),
		)
	}

	dlh.logger.Info("Dead letter event retry successful",
		logger.String("dead_letter_id", deadLetterID),
		logger.String("original_event_id", dlEvent.OriginalEventID),
	)

	return nil
}

// Task type constants for dead letter processing
const (
	TaskTypeDeadLetterRetry = "dead_letter:retry"
)

// DeadLetterRetryTaskHandler handles dead letter retry tasks
func DeadLetterRetryTaskHandler(deadLetterHandler *DeadLetterHandler) queue.TaskHandler {
	return func(ctx context.Context, task *asynq.Task) error {
		var payload map[string]any
		if err := json.Unmarshal(task.Payload(), &payload); err != nil {
			return fmt.Errorf("failed to unmarshal dead letter retry task payload: %w", err)
		}

		deadLetterID, ok := payload["dead_letter_id"].(string)
		if !ok {
			return fmt.Errorf("missing or invalid 'dead_letter_id' field in dead letter retry task")
		}

		return deadLetterHandler.RetryDeadLetterEvent(ctx, deadLetterID)
	}
}

// RegisterDeadLetterHandlers registers dead letter processing handlers with the task processor
func RegisterDeadLetterHandlers(processor *queue.TaskProcessor, deadLetterHandler *DeadLetterHandler) {
	processor.RegisterHandler(TaskTypeDeadLetterRetry, DeadLetterRetryTaskHandler(deadLetterHandler))
}

// Global dead letter queue instance
var globalDeadLetterQueue DeadLetterQueue

// InitDeadLetterQueue initializes the global dead letter queue
func InitDeadLetterQueue(dlq DeadLetterQueue) {
	globalDeadLetterQueue = dlq
}

// GetDeadLetterQueue returns the global dead letter queue
func GetDeadLetterQueue() DeadLetterQueue {
	if globalDeadLetterQueue == nil {
		globalDeadLetterQueue = NewInMemoryDeadLetterQueue()
	}
	return globalDeadLetterQueue
}
