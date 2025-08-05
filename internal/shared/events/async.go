package events

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"linke/internal/shared/logger"
	"linke/internal/shared/queue"

	"github.com/hibiken/asynq"
)

// AsyncEventProcessor handles asynchronous event processing using the queue system
type AsyncEventProcessor struct {
	taskQueue     *queue.TaskQueue
	eventStore    EventStore
	eventBus      EventBus
	logger        logger.Logger
	retryConfig   RetryConfig
	deadLetterBus EventBus // For events that failed all retries
}

// RetryConfig defines retry behavior for failed events
type RetryConfig struct {
	MaxRetries   int           `json:"max_retries"`
	InitialDelay time.Duration `json:"initial_delay"`
	MaxDelay     time.Duration `json:"max_delay"`
	Multiplier   float64       `json:"multiplier"`
}

// DefaultRetryConfig returns a sensible default retry configuration
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:   3,
		InitialDelay: time.Second * 2,
		MaxDelay:     time.Minute * 5,
		Multiplier:   2.0,
	}
}

// NewAsyncEventProcessor creates a new async event processor
func NewAsyncEventProcessor(
	taskQueue *queue.TaskQueue,
	eventStore EventStore,
	eventBus EventBus,
	retryConfig RetryConfig,
) *AsyncEventProcessor {
	return &AsyncEventProcessor{
		taskQueue:     taskQueue,
		eventStore:    eventStore,
		eventBus:      eventBus,
		logger:        logger.GetGlobalLogger(),
		retryConfig:   retryConfig,
		deadLetterBus: NewInMemoryEventBus(), // Simple in-memory bus for dead letters
	}
}

// AsyncEventBus wraps EventBus with async processing capabilities
type AsyncEventBus struct {
	EventBus
	asyncProcessor *AsyncEventProcessor
}

// NewAsyncEventBus creates a new event bus with async processing
func NewAsyncEventBus(eventBus EventBus, asyncProcessor *AsyncEventProcessor) *AsyncEventBus {
	return &AsyncEventBus{
		EventBus:       eventBus,
		asyncProcessor: asyncProcessor,
	}
}

// PublishAsync publishes an event asynchronously using the queue system
func (bus *AsyncEventBus) PublishAsync(ctx context.Context, event Event) error {
	// Store event first for audit trail
	if bus.asyncProcessor.eventStore != nil {
		if err := bus.asyncProcessor.eventStore.Store(ctx, event); err != nil {
			bus.asyncProcessor.logger.Error("Failed to store event before async processing",
				logger.String("event_id", event.EventID()),
				logger.String("event_type", event.EventType()),
				logger.ErrorField(err),
			)
			// Continue with async processing even if storage fails
		}
	}

	// Serialize the event for the queue
	eventData, err := SerializeEvent(event)
	if err != nil {
		return fmt.Errorf("failed to serialize event for async processing: %w", err)
	}

	// Create a task for the event
	task := &queue.Task{
		ID:       event.EventID(),
		Type:     TaskTypeEventProcessing,
		MaxRetry: bus.asyncProcessor.retryConfig.MaxRetries,
		Payload: map[string]any{
			"event_data":     string(eventData),
			"event_id":       event.EventID(),
			"event_type":     event.EventType(),
			"correlation_id": extractCorrelationID(event),
		},
	}

	// Enqueue the task
	if err := bus.asyncProcessor.taskQueue.Enqueue(ctx, "events", task); err != nil {
		bus.asyncProcessor.logger.Error("Failed to enqueue event for async processing",
			logger.String("event_id", event.EventID()),
			logger.String("event_type", event.EventType()),
			logger.ErrorField(err),
		)
		return fmt.Errorf("failed to enqueue event: %w", err)
	}

	bus.asyncProcessor.logger.Debug("Event enqueued for async processing",
		logger.String("event_id", event.EventID()),
		logger.String("event_type", event.EventType()),
	)

	return nil
}

// Task type constants for event processing
const (
	TaskTypeEventProcessing      = "event:process"
	TaskTypeEventReprocessing    = "event:reprocess"
	TaskTypeDeadLetterProcessing = "event:dead_letter"
)

// EventProcessingTaskHandler handles event processing tasks
func EventProcessingTaskHandler(asyncProcessor *AsyncEventProcessor) queue.TaskHandler {
	return func(ctx context.Context, task *asynq.Task) error {
		var payload map[string]any
		if err := json.Unmarshal(task.Payload(), &payload); err != nil {
			return fmt.Errorf("failed to unmarshal event processing task payload: %w", err)
		}

		eventData, ok := payload["event_data"].(string)
		if !ok {
			return fmt.Errorf("missing or invalid 'event_data' field in event processing task")
		}

		eventID, ok := payload["event_id"].(string)
		if !ok {
			return fmt.Errorf("missing or invalid 'event_id' field in event processing task")
		}

		eventType, ok := payload["event_type"].(string)
		if !ok {
			return fmt.Errorf("missing or invalid 'event_type' field in event processing task")
		}

		correlationID, _ := payload["correlation_id"].(string)

		// Add correlation ID to context if available
		if correlationID != "" {
			ctx = context.WithValue(ctx, "correlation_id", correlationID)
		}

		asyncProcessor.logger.Info("Processing event asynchronously",
			logger.String("event_id", eventID),
			logger.String("event_type", eventType),
			logger.String("correlation_id", correlationID),
		)

		// Deserialize the event
		envelope, err := DeserializeEvent([]byte(eventData))
		if err != nil {
			asyncProcessor.logger.Error("Failed to deserialize event",
				logger.String("event_id", eventID),
				logger.ErrorField(err),
			)
			return fmt.Errorf("failed to deserialize event: %w", err)
		}

		// Process the event synchronously through the event bus
		if err := asyncProcessor.eventBus.Publish(ctx, envelope.Event); err != nil {
			asyncProcessor.logger.Error("Failed to process event asynchronously",
				logger.String("event_id", eventID),
				logger.String("event_type", eventType),
				logger.ErrorField(err),
			)

			// Note: Dead letter handling would be done by asynq's built-in retry mechanism
			// For now, we'll log the error and let asynq handle retries
			// In a production setup, you'd configure asynq with a dead letter queue

			return fmt.Errorf("failed to process event: %w", err)
		}

		asyncProcessor.logger.Info("Event processed successfully",
			logger.String("event_id", eventID),
			logger.String("event_type", eventType),
		)

		return nil
	}
}

// handleDeadLetter handles events that failed all retry attempts
func (p *AsyncEventProcessor) handleDeadLetter(ctx context.Context, event Event, originalError error) {
	p.logger.Warn("Event moved to dead letter queue",
		logger.String("event_id", event.EventID()),
		logger.String("event_type", event.EventType()),
		logger.ErrorField(originalError),
	)

	// Create a dead letter event
	deadLetterEvent := NewBaseEvent(
		"event.dead_letter",
		"event-processor",
		map[string]any{
			"original_event_id":   event.EventID(),
			"original_event_type": event.EventType(),
			"failure_reason":      originalError.Error(),
			"failed_at":           time.Now(),
		},
	)

	// Publish to dead letter bus (this should never fail with in-memory bus)
	if err := p.deadLetterBus.Publish(ctx, deadLetterEvent); err != nil {
		p.logger.Error("Failed to publish dead letter event",
			logger.String("original_event_id", event.EventID()),
			logger.ErrorField(err),
		)
	}

	// Store dead letter event if event store is available
	if p.eventStore != nil {
		if err := p.eventStore.Store(ctx, deadLetterEvent); err != nil {
			p.logger.Error("Failed to store dead letter event",
				logger.String("original_event_id", event.EventID()),
				logger.ErrorField(err),
			)
		}
	}
}

// RegisterEventHandlers registers event processing handlers with the task processor
func RegisterEventHandlers(processor *queue.TaskProcessor, asyncProcessor *AsyncEventProcessor) {
	processor.RegisterHandler(TaskTypeEventProcessing, EventProcessingTaskHandler(asyncProcessor))
}

// EventReplayHandler provides functionality to replay events asynchronously
type EventReplayHandler struct {
	eventStore     EventStore
	asyncProcessor *AsyncEventProcessor
	logger         logger.Logger
}

// NewEventReplayHandler creates a new event replay handler
func NewEventReplayHandler(eventStore EventStore, asyncProcessor *AsyncEventProcessor) *EventReplayHandler {
	return &EventReplayHandler{
		eventStore:     eventStore,
		asyncProcessor: asyncProcessor,
		logger:         logger.GetGlobalLogger(),
	}
}

// ReplayEvents replays events from a specific timestamp asynchronously
func (h *EventReplayHandler) ReplayEvents(ctx context.Context, fromTimestamp time.Time, eventTypes []string) error {
	filters := EventFilters{
		FromTime:   fromTimestamp,
		EventTypes: eventTypes,
		Limit:      1000, // Process in batches
	}

	events, err := h.eventStore.GetEvents(ctx, filters)
	if err != nil {
		return fmt.Errorf("failed to get events for replay: %w", err)
	}

	h.logger.Info("Starting event replay",
		logger.String("from_timestamp", fromTimestamp.Format(time.RFC3339)),
		logger.Any("event_types", eventTypes),
		logger.Int("event_count", len(events)),
	)

	// Replay events asynchronously
	for _, storedEvent := range events {
		// Convert stored event back to Event interface
		event, err := h.deserializeStoredEvent(storedEvent)
		if err != nil {
			h.logger.Error("Failed to deserialize event during replay",
				logger.String("event_id", storedEvent.EventID),
				logger.ErrorField(err),
			)
			continue
		}

		// Create a replay task
		eventData, err := SerializeEvent(event)
		if err != nil {
			h.logger.Error("Failed to serialize event for replay",
				logger.String("event_id", event.EventID()),
				logger.ErrorField(err),
			)
			continue
		}

		task := &queue.Task{
			ID:       fmt.Sprintf("replay-%s", event.EventID()),
			Type:     TaskTypeEventReprocessing,
			MaxRetry: 1, // Don't retry replay events
			Payload: map[string]any{
				"event_data": string(eventData),
				"event_id":   event.EventID(),
				"event_type": event.EventType(),
				"is_replay":  true,
			},
		}

		if err := h.asyncProcessor.taskQueue.Enqueue(ctx, "event-replay", task); err != nil {
			h.logger.Error("Failed to enqueue event for replay",
				logger.String("event_id", event.EventID()),
				logger.ErrorField(err),
			)
			continue
		}
	}

	h.logger.Info("Event replay tasks enqueued",
		logger.Int("event_count", len(events)),
	)

	return nil
}

// deserializeStoredEvent is a simplified version for replay
func (h *EventReplayHandler) deserializeStoredEvent(storedEvent *StoredEvent) (Event, error) {
	var eventData any
	if err := json.Unmarshal([]byte(storedEvent.EventData), &eventData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal event data: %w", err)
	}

	return &BaseEvent{
		ID:      storedEvent.EventID,
		Type:    storedEvent.EventType,
		Source:  storedEvent.EventSource,
		Time:    storedEvent.OccurredAt,
		Version: storedEvent.EventVersion,
		Data:    eventData,
	}, nil
}

// EventMetrics provides metrics for event processing
type EventMetrics struct {
	ProcessedEvents       int64         `json:"processed_events"`
	FailedEvents          int64         `json:"failed_events"`
	DeadLetterEvents      int64         `json:"dead_letter_events"`
	AverageProcessingTime time.Duration `json:"average_processing_time"`
	QueueLength           int64         `json:"queue_length"`
}

// GetMetrics returns event processing metrics
func (p *AsyncEventProcessor) GetMetrics(ctx context.Context) (*EventMetrics, error) {
	// This is a simplified implementation
	// In a real implementation, you would collect these metrics
	// from your monitoring system or maintain counters
	return &EventMetrics{
		ProcessedEvents:  0, // Would be tracked via counters
		FailedEvents:     0, // Would be tracked via counters
		DeadLetterEvents: 0, // Would be tracked via counters
		QueueLength:      0, // Would be queried from Redis
	}, nil
}

// extractCorrelationID extracts correlation ID from event metadata
func extractCorrelationID(event Event) string {
	if baseEvent, ok := event.(*BaseEvent); ok {
		if correlationID, exists := baseEvent.GetMetadata("correlation_id"); exists {
			if id, ok := correlationID.(string); ok {
				return id
			}
		}
	}
	return ""
}

// SetCorrelationID sets correlation ID in event metadata
func SetCorrelationID(event Event, correlationID string) {
	if baseEvent, ok := event.(*BaseEvent); ok {
		baseEvent.SetMetadata("correlation_id", correlationID)
	}
}

// EventProcessingMiddleware adds correlation ID tracking to events
func EventProcessingMiddleware() EventMiddleware {
	return EventMiddlewareFunc(func(ctx context.Context, event Event, next func(context.Context, Event) error) error {
		// Extract correlation ID from context
		if correlationID := ctx.Value("correlation_id"); correlationID != nil {
			if id, ok := correlationID.(string); ok {
				SetCorrelationID(event, id)
			}
		}

		// Add processing timestamp
		if baseEvent, ok := event.(*BaseEvent); ok {
			baseEvent.SetMetadata("processed_at", time.Now())
		}

		return next(ctx, event)
	})
}
