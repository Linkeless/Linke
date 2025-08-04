package events

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"linke/internal/shared/logger"
)

// DeduplicationStrategy defines different strategies for event deduplication
type DeduplicationStrategy string

const (
	DeduplicationByEventID   DeduplicationStrategy = "event_id"
	DeduplicationByContent   DeduplicationStrategy = "content"
	DeduplicationBySignature DeduplicationStrategy = "signature"
)

// EventDeduplicator provides event deduplication capabilities
type EventDeduplicator interface {
	IsDuplicate(ctx context.Context, event Event) (bool, error)
	MarkProcessed(ctx context.Context, event Event) error
	GetProcessedCount(ctx context.Context) (int64, error)
	Clean(ctx context.Context, olderThan time.Time) error
}

// DeduplicationConfig contains configuration for event deduplication
type DeduplicationConfig struct {
	Strategy        DeduplicationStrategy `json:"strategy"`
	TTL             time.Duration         `json:"ttl"`              // How long to keep deduplication records
	CleanupInterval time.Duration         `json:"cleanup_interval"` // How often to run cleanup
	UseSignature    bool                  `json:"use_signature"`    // Whether to use content signature for deduplication
}

// DefaultDeduplicationConfig returns a sensible default configuration
func DefaultDeduplicationConfig() DeduplicationConfig {
	return DeduplicationConfig{
		Strategy:        DeduplicationByEventID,
		TTL:             time.Hour * 24, // Keep records for 24 hours
		CleanupInterval: time.Hour,      // Cleanup every hour
		UseSignature:    false,
	}
}

// ProcessedEvent represents a processed event record for deduplication
type ProcessedEvent struct {
	EventID     string    `json:"event_id"`
	EventType   string    `json:"event_type"`
	EventSource string    `json:"event_source"`
	Signature   string    `json:"signature,omitempty"`
	ProcessedAt time.Time `json:"processed_at"`
	HandlerName string    `json:"handler_name,omitempty"`
}

// InMemoryEventDeduplicator provides in-memory event deduplication
type InMemoryEventDeduplicator struct {
	config          DeduplicationConfig
	processedEvents map[string]*ProcessedEvent
	mutex           sync.RWMutex
	logger          logger.Logger
	stopCleanup     chan struct{}
	cleanupOnce     sync.Once
}

// NewInMemoryEventDeduplicator creates a new in-memory event deduplicator
func NewInMemoryEventDeduplicator(config DeduplicationConfig) *InMemoryEventDeduplicator {
	deduplicator := &InMemoryEventDeduplicator{
		config:          config,
		processedEvents: make(map[string]*ProcessedEvent),
		logger:          logger.GetGlobalLogger(),
		stopCleanup:     make(chan struct{}),
	}

	// Start cleanup goroutine
	go deduplicator.cleanupRoutine()

	return deduplicator
}

// IsDuplicate checks if an event has already been processed
func (ded *InMemoryEventDeduplicator) IsDuplicate(ctx context.Context, event Event) (bool, error) {
	key, err := ded.generateDeduplicationKey(event)
	if err != nil {
		return false, fmt.Errorf("failed to generate deduplication key: %w", err)
	}

	ded.mutex.RLock()
	defer ded.mutex.RUnlock()

	processedEvent, exists := ded.processedEvents[key]
	if !exists {
		return false, nil
	}

	// Check if the record has expired
	if time.Since(processedEvent.ProcessedAt) > ded.config.TTL {
		// Record has expired, consider it not a duplicate
		return false, nil
	}

	ded.logger.Debug("Duplicate event detected",
		logger.String("event_id", event.EventID()),
		logger.String("event_type", event.EventType()),
		logger.String("dedup_key", key),
		logger.Any("processed_at", processedEvent.ProcessedAt),
	)

	return true, nil
}

// MarkProcessed marks an event as processed
func (ded *InMemoryEventDeduplicator) MarkProcessed(ctx context.Context, event Event) error {
	key, err := ded.generateDeduplicationKey(event)
	if err != nil {
		return fmt.Errorf("failed to generate deduplication key: %w", err)
	}

	ded.mutex.Lock()
	defer ded.mutex.Unlock()

	signature := ""
	if ded.config.UseSignature {
		signature, err = ded.generateEventSignature(event)
		if err != nil {
			ded.logger.Warn("Failed to generate event signature",
				logger.String("event_id", event.EventID()),
				logger.ErrorField(err),
			)
		}
	}

	ded.processedEvents[key] = &ProcessedEvent{
		EventID:     event.EventID(),
		EventType:   event.EventType(),
		EventSource: event.EventSource(),
		Signature:   signature,
		ProcessedAt: time.Now(),
	}

	ded.logger.Debug("Event marked as processed",
		logger.String("event_id", event.EventID()),
		logger.String("event_type", event.EventType()),
		logger.String("dedup_key", key),
	)

	return nil
}

// GetProcessedCount returns the number of processed events
func (ded *InMemoryEventDeduplicator) GetProcessedCount(ctx context.Context) (int64, error) {
	ded.mutex.RLock()
	defer ded.mutex.RUnlock()

	return int64(len(ded.processedEvents)), nil
}

// Clean removes processed events older than the specified time
func (ded *InMemoryEventDeduplicator) Clean(ctx context.Context, olderThan time.Time) error {
	ded.mutex.Lock()
	defer ded.mutex.Unlock()

	var removedCount int
	for key, processedEvent := range ded.processedEvents {
		if processedEvent.ProcessedAt.Before(olderThan) {
			delete(ded.processedEvents, key)
			removedCount++
		}
	}

	if removedCount > 0 {
		ded.logger.Info("Cleaned up expired deduplication records",
			logger.Int("removed_count", removedCount),
			logger.Any("older_than", olderThan),
		)
	}

	return nil
}

// Close stops the cleanup routine
func (ded *InMemoryEventDeduplicator) Close() error {
	ded.cleanupOnce.Do(func() {
		close(ded.stopCleanup)
	})
	return nil
}

// generateDeduplicationKey generates a key for deduplication based on the configured strategy
func (ded *InMemoryEventDeduplicator) generateDeduplicationKey(event Event) (string, error) {
	switch ded.config.Strategy {
	case DeduplicationByEventID:
		return event.EventID(), nil

	case DeduplicationByContent:
		signature, err := ded.generateEventSignature(event)
		if err != nil {
			return "", fmt.Errorf("failed to generate content signature: %w", err)
		}
		return fmt.Sprintf("%s:%s", event.EventType(), signature), nil

	case DeduplicationBySignature:
		signature, err := ded.generateEventSignature(event)
		if err != nil {
			return "", fmt.Errorf("failed to generate signature: %w", err)
		}
		return signature, nil

	default:
		return "", fmt.Errorf("unknown deduplication strategy: %s", ded.config.Strategy)
	}
}

// generateEventSignature generates a signature for an event based on its content
func (ded *InMemoryEventDeduplicator) generateEventSignature(event Event) (string, error) {
	// Create a content-based signature that excludes timestamps and IDs
	// to focus only on the logical content of the event
	contentData := map[string]any{
		"type":    event.EventType(),
		"source":  event.EventSource(),
		"version": event.EventVersion(),
		"data":    event.EventData(),
	}

	// Serialize content data for signature generation
	eventData, err := json.Marshal(contentData)
	if err != nil {
		return "", fmt.Errorf("failed to serialize event content for signature: %w", err)
	}

	// Generate SHA-256 hash
	hasher := sha256.New()
	hasher.Write(eventData)
	hash := hasher.Sum(nil)

	return hex.EncodeToString(hash), nil
}

// cleanupRoutine runs periodic cleanup of expired records
func (ded *InMemoryEventDeduplicator) cleanupRoutine() {
	ticker := time.NewTicker(ded.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			cutoff := time.Now().Add(-ded.config.TTL)
			if err := ded.Clean(context.Background(), cutoff); err != nil {
				ded.logger.Error("Failed to clean up deduplication records",
					logger.ErrorField(err),
				)
			}
		case <-ded.stopCleanup:
			return
		}
	}
}

// DeduplicatingEventHandler wraps an event handler with deduplication
type DeduplicatingEventHandler struct {
	handler      EventHandler
	deduplicator EventDeduplicator
	handlerName  string
	logger       logger.Logger
	id           string
}

// NewDeduplicatingEventHandler creates a new deduplicating event handler
func NewDeduplicatingEventHandler(handlerName string, handler EventHandler, deduplicator EventDeduplicator) *DeduplicatingEventHandler {
	return &DeduplicatingEventHandler{
		handler:      handler,
		deduplicator: deduplicator,
		handlerName:  handlerName,
		logger:       logger.GetGlobalLogger(),
		id:           generateEventID(),
	}
}

// Handle implements the EventHandler interface with deduplication
func (deh *DeduplicatingEventHandler) Handle(ctx context.Context, event Event) error {
	// Check if event is a duplicate
	isDuplicate, err := deh.deduplicator.IsDuplicate(ctx, event)
	if err != nil {
		deh.logger.Error("Failed to check for duplicate event",
			logger.String("handler", deh.handlerName),
			logger.String("event_id", event.EventID()),
			logger.String("event_type", event.EventType()),
			logger.ErrorField(err),
		)
		// Continue processing even if deduplication check fails
	} else if isDuplicate {
		deh.logger.Info("Skipping duplicate event",
			logger.String("handler", deh.handlerName),
			logger.String("event_id", event.EventID()),
			logger.String("event_type", event.EventType()),
		)
		return nil // Skip processing duplicate events
	}

	// Process the event
	err = deh.handler.Handle(ctx, event)
	if err != nil {
		return err // Don't mark as processed if handling failed
	}

	// Mark event as processed
	if markErr := deh.deduplicator.MarkProcessed(ctx, event); markErr != nil {
		deh.logger.Error("Failed to mark event as processed",
			logger.String("handler", deh.handlerName),
			logger.String("event_id", event.EventID()),
			logger.String("event_type", event.EventType()),
			logger.ErrorField(markErr),
		)
		// Don't fail the overall processing because of this
	}

	return nil
}

// EventTypes returns the event types this handler processes
func (deh *DeduplicatingEventHandler) EventTypes() []string {
	return deh.handler.EventTypes()
}

// ID returns the unique identifier for this handler
func (deh *DeduplicatingEventHandler) ID() string {
	return deh.id
}

// GetDeduplicator returns the underlying deduplicator
func (deh *DeduplicatingEventHandler) GetDeduplicator() EventDeduplicator {
	return deh.deduplicator
}

// AtLeastOnceEventBus provides at-least-once delivery guarantees with deduplication
type AtLeastOnceEventBus struct {
	EventBus
	deduplicator EventDeduplicator
	retryPolicy  AtLeastOnceRetryPolicy
	logger       logger.Logger
}

// AtLeastOnceRetryPolicy defines retry behavior for failed event processing
type AtLeastOnceRetryPolicy struct {
	MaxRetries   int           `json:"max_retries"`
	InitialDelay time.Duration `json:"initial_delay"`
	MaxDelay     time.Duration `json:"max_delay"`
	Multiplier   float64       `json:"multiplier"`
	EnableJitter bool          `json:"enable_jitter"`
}

// DefaultAtLeastOnceRetryPolicy returns a sensible default retry policy
func DefaultAtLeastOnceRetryPolicy() AtLeastOnceRetryPolicy {
	return AtLeastOnceRetryPolicy{
		MaxRetries:   3,
		InitialDelay: time.Second * 2,
		MaxDelay:     time.Minute * 5,
		Multiplier:   2.0,
		EnableJitter: true,
	}
}

// NewAtLeastOnceEventBus creates a new at-least-once event bus
func NewAtLeastOnceEventBus(eventBus EventBus, deduplicator EventDeduplicator, retryPolicy AtLeastOnceRetryPolicy) *AtLeastOnceEventBus {
	return &AtLeastOnceEventBus{
		EventBus:     eventBus,
		deduplicator: deduplicator,
		retryPolicy:  retryPolicy,
		logger:       logger.GetGlobalLogger(),
	}
}

// Publish publishes an event with at-least-once delivery guarantee
func (alob *AtLeastOnceEventBus) Publish(ctx context.Context, event Event) error {
	var lastErr error

	for attempt := 0; attempt <= alob.retryPolicy.MaxRetries; attempt++ {
		// Check if event was already processed (deduplication)
		isDuplicate, err := alob.deduplicator.IsDuplicate(ctx, event)
		if err != nil {
			alob.logger.Warn("Failed to check for duplicate event",
				logger.String("event_id", event.EventID()),
				logger.String("event_type", event.EventType()),
				logger.Int("attempt", attempt+1),
				logger.ErrorField(err),
			)
		} else if isDuplicate {
			alob.logger.Info("Event already processed, skipping",
				logger.String("event_id", event.EventID()),
				logger.String("event_type", event.EventType()),
			)
			return nil
		}

		// Attempt to publish
		err = alob.EventBus.Publish(ctx, event)
		if err == nil {
			// Success, mark as processed
			if markErr := alob.deduplicator.MarkProcessed(ctx, event); markErr != nil {
				alob.logger.Warn("Failed to mark event as processed after successful publish",
					logger.String("event_id", event.EventID()),
					logger.String("event_type", event.EventType()),
					logger.ErrorField(markErr),
				)
			}
			return nil
		}

		lastErr = err

		// If this is the last attempt, don't wait
		if attempt == alob.retryPolicy.MaxRetries {
			break
		}

		// Calculate retry delay
		delay := alob.calculateRetryDelay(attempt)

		alob.logger.Warn("Event publishing failed, retrying",
			logger.String("event_id", event.EventID()),
			logger.String("event_type", event.EventType()),
			logger.Int("attempt", attempt+1),
			logger.Int("max_retries", alob.retryPolicy.MaxRetries),
			logger.Any("retry_delay", delay),
			logger.ErrorField(err),
		)

		// Wait before retrying
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
			// Continue to next retry
		}
	}

	alob.logger.Error("Event publishing failed after all retries",
		logger.String("event_id", event.EventID()),
		logger.String("event_type", event.EventType()),
		logger.Int("max_retries", alob.retryPolicy.MaxRetries),
		logger.ErrorField(lastErr),
	)

	return fmt.Errorf("failed to publish event after %d retries: %w", alob.retryPolicy.MaxRetries, lastErr)
}

// calculateRetryDelay calculates the delay for a retry attempt
func (alob *AtLeastOnceEventBus) calculateRetryDelay(attempt int) time.Duration {
	delay := alob.retryPolicy.InitialDelay

	// Apply exponential backoff
	for i := 0; i < attempt; i++ {
		delay = time.Duration(float64(delay) * alob.retryPolicy.Multiplier)
	}

	// Apply maximum delay
	if delay > alob.retryPolicy.MaxDelay {
		delay = alob.retryPolicy.MaxDelay
	}

	// Apply jitter if enabled
	if alob.retryPolicy.EnableJitter {
		jitter := time.Duration(float64(delay) * 0.1) // 10% jitter
		jitterFactor := (2.0*float64(time.Now().UnixNano()%1000)/1000.0 - 1.0)
		delay += time.Duration(float64(jitter) * jitterFactor)
	}

	return delay
}

// DeduplicationStats provides statistics about deduplication
type DeduplicationStats struct {
	TotalProcessedEvents int64     `json:"total_processed_events"`
	DuplicatesDetected   int64     `json:"duplicates_detected"`
	DeduplicationRate    float64   `json:"deduplication_rate"`
	OldestRecord         time.Time `json:"oldest_record"`
	NewestRecord         time.Time `json:"newest_record"`
}

// EventDeduplicationManager manages multiple deduplicators
type EventDeduplicationManager struct {
	deduplicators map[string]EventDeduplicator
	defaultConfig DeduplicationConfig
	mutex         sync.RWMutex
	logger        logger.Logger
}

// NewEventDeduplicationManager creates a new deduplication manager
func NewEventDeduplicationManager(defaultConfig DeduplicationConfig) *EventDeduplicationManager {
	return &EventDeduplicationManager{
		deduplicators: make(map[string]EventDeduplicator),
		defaultConfig: defaultConfig,
		logger:        logger.GetGlobalLogger(),
	}
}

// GetOrCreateDeduplicator gets an existing deduplicator or creates a new one
func (edm *EventDeduplicationManager) GetOrCreateDeduplicator(name string, config ...DeduplicationConfig) EventDeduplicator {
	edm.mutex.Lock()
	defer edm.mutex.Unlock()

	if deduplicator, exists := edm.deduplicators[name]; exists {
		return deduplicator
	}

	// Use provided config or default
	var dedupConfig DeduplicationConfig
	if len(config) > 0 {
		dedupConfig = config[0]
	} else {
		dedupConfig = edm.defaultConfig
	}

	deduplicator := NewInMemoryEventDeduplicator(dedupConfig)
	edm.deduplicators[name] = deduplicator

	edm.logger.Info("Event deduplicator created",
		logger.String("name", name),
		logger.Any("config", dedupConfig),
	)

	return deduplicator
}

// GetDeduplicator returns a deduplicator by name
func (edm *EventDeduplicationManager) GetDeduplicator(name string) (EventDeduplicator, bool) {
	edm.mutex.RLock()
	defer edm.mutex.RUnlock()

	deduplicator, exists := edm.deduplicators[name]
	return deduplicator, exists
}

// Close closes all deduplicators
func (edm *EventDeduplicationManager) Close() error {
	edm.mutex.Lock()
	defer edm.mutex.Unlock()

	for name, deduplicator := range edm.deduplicators {
		if closer, ok := deduplicator.(*InMemoryEventDeduplicator); ok {
			if err := closer.Close(); err != nil {
				edm.logger.Error("Failed to close deduplicator",
					logger.String("name", name),
					logger.ErrorField(err),
				)
			}
		}
	}

	edm.deduplicators = make(map[string]EventDeduplicator)
	return nil
}

// Global deduplication manager
var globalDeduplicationManager *EventDeduplicationManager

// InitDeduplicationManager initializes the global deduplication manager
func InitDeduplicationManager(config DeduplicationConfig) {
	globalDeduplicationManager = NewEventDeduplicationManager(config)
}

// GetDeduplicationManager returns the global deduplication manager
func GetDeduplicationManager() *EventDeduplicationManager {
	if globalDeduplicationManager == nil {
		globalDeduplicationManager = NewEventDeduplicationManager(DefaultDeduplicationConfig())
	}
	return globalDeduplicationManager
}

// WrapHandlerWithDeduplication wraps an event handler with deduplication
func WrapHandlerWithDeduplication(handlerName string, handler EventHandler, config ...DeduplicationConfig) EventHandler {
	manager := GetDeduplicationManager()
	deduplicator := manager.GetOrCreateDeduplicator(handlerName, config...)
	return NewDeduplicatingEventHandler(handlerName, handler, deduplicator)
}

// WrapEventBusWithAtLeastOnce wraps an event bus with at-least-once delivery guarantee
func WrapEventBusWithAtLeastOnce(eventBus EventBus, deduplicator EventDeduplicator, retryPolicy ...AtLeastOnceRetryPolicy) EventBus {
	var policy AtLeastOnceRetryPolicy
	if len(retryPolicy) > 0 {
		policy = retryPolicy[0]
	} else {
		policy = DefaultAtLeastOnceRetryPolicy()
	}

	return NewAtLeastOnceEventBus(eventBus, deduplicator, policy)
}
