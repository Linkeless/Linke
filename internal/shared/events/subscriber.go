package events

import (
	"context"
	"fmt"
	"sync"
	"time"

	"linke/internal/shared/logger"
)

// SubscriberManager manages event subscribers and their lifecycles
type SubscriberManager struct {
	subscribers map[string]*EventSubscriber
	mutex       sync.RWMutex
	logger      logger.Logger
}

// EventSubscriber represents a single event subscriber
type EventSubscriber struct {
	ID          string
	EventTypes  []string
	Handler     EventHandler
	Config      SubscriberConfig
	IsActive    bool
	CreatedAt   time.Time
	LastEventAt time.Time
	ProcessedCount int64
	FailedCount    int64
	mutex       sync.RWMutex
}

// SubscriberConfig contains configuration for an event subscriber
type SubscriberConfig struct {
	MaxRetries      int           `json:"max_retries"`
	RetryDelay      time.Duration `json:"retry_delay"`
	TimeoutDuration time.Duration `json:"timeout_duration"`
	BufferSize      int           `json:"buffer_size"`
	Async           bool          `json:"async"`
}

// DefaultSubscriberConfig returns a default subscriber configuration
func DefaultSubscriberConfig() SubscriberConfig {
	return SubscriberConfig{
		MaxRetries:      3,
		RetryDelay:      time.Second * 2,
		TimeoutDuration: time.Second * 30,
		BufferSize:      100,
		Async:           true,
	}
}

// NewSubscriberManager creates a new subscriber manager
func NewSubscriberManager() *SubscriberManager {
	return &SubscriberManager{
		subscribers: make(map[string]*EventSubscriber),
		logger:      logger.GetGlobalLogger(),
	}
}

// AddSubscriber adds a new event subscriber
func (sm *SubscriberManager) AddSubscriber(id string, eventTypes []string, handler EventHandler, config SubscriberConfig) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	if _, exists := sm.subscribers[id]; exists {
		return fmt.Errorf("subscriber with ID %s already exists", id)
	}

	subscriber := &EventSubscriber{
		ID:         id,
		EventTypes: eventTypes,
		Handler:    handler,
		Config:     config,
		IsActive:   true,
		CreatedAt:  time.Now(),
	}

	sm.subscribers[id] = subscriber

	sm.logger.Info("Event subscriber added",
		logger.String("subscriber_id", id),
		logger.Any("event_types", eventTypes))

	return nil
}

// RemoveSubscriber removes an event subscriber
func (sm *SubscriberManager) RemoveSubscriber(id string) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	if _, exists := sm.subscribers[id]; !exists {
		return fmt.Errorf("subscriber with ID %s not found", id)
	}

	delete(sm.subscribers, id)

	sm.logger.Info("Event subscriber removed",
		logger.String("subscriber_id", id))

	return nil
}

// GetSubscriber returns a subscriber by ID
func (sm *SubscriberManager) GetSubscriber(id string) (*EventSubscriber, error) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	subscriber, exists := sm.subscribers[id]
	if !exists {
		return nil, fmt.Errorf("subscriber with ID %s not found", id)
	}

	return subscriber, nil
}

// GetSubscribersForEvent returns all active subscribers for a specific event type
func (sm *SubscriberManager) GetSubscribersForEvent(eventType string) []*EventSubscriber {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	var matchingSubscribers []*EventSubscriber
	for _, subscriber := range sm.subscribers {
		if subscriber.IsActive && subscriber.HandlesEventType(eventType) {
			matchingSubscribers = append(matchingSubscribers, subscriber)
		}
	}

	return matchingSubscribers
}

// ListSubscribers returns all subscribers
func (sm *SubscriberManager) ListSubscribers() map[string]*EventSubscriber {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	result := make(map[string]*EventSubscriber)
	for id, subscriber := range sm.subscribers {
		result[id] = subscriber
	}

	return result
}

// DeactivateSubscriber deactivates a subscriber without removing it
func (sm *SubscriberManager) DeactivateSubscriber(id string) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	subscriber, exists := sm.subscribers[id]
	if !exists {
		return fmt.Errorf("subscriber with ID %s not found", id)
	}

	subscriber.IsActive = false

	sm.logger.Info("Event subscriber deactivated",
		logger.String("subscriber_id", id))

	return nil
}

// ActivateSubscriber activates a previously deactivated subscriber
func (sm *SubscriberManager) ActivateSubscriber(id string) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	subscriber, exists := sm.subscribers[id]
	if !exists {
		return fmt.Errorf("subscriber with ID %s not found", id)
	}

	subscriber.IsActive = true

	sm.logger.Info("Event subscriber activated",
		logger.String("subscriber_id", id))

	return nil
}

// EventSubscriber methods

// HandlesEventType checks if the subscriber handles a specific event type
func (es *EventSubscriber) HandlesEventType(eventType string) bool {
	for _, et := range es.EventTypes {
		if et == eventType {
			return true
		}
	}
	return false
}

// HandleEvent processes an event with retry logic and statistics
func (es *EventSubscriber) HandleEvent(ctx context.Context, event Event) error {
	es.mutex.Lock()
	es.LastEventAt = time.Now()
	es.mutex.Unlock()

	// Create a context with timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, es.Config.TimeoutDuration)
	defer cancel()

	var err error
	for attempt := 0; attempt <= es.Config.MaxRetries; attempt++ {
		err = es.Handler.Handle(timeoutCtx, event)
		if err == nil {
			es.mutex.Lock()
			es.ProcessedCount++
			es.mutex.Unlock()
			return nil
		}

		if attempt < es.Config.MaxRetries {
			logger.Warn("Event handling failed, retrying",
				logger.String("subscriber_id", es.ID),
				logger.String("event_type", event.EventType()),
				logger.String("event_id", event.EventID()),
				logger.Int("attempt", attempt+1),
				logger.Int("max_retries", es.Config.MaxRetries),
				logger.Error2("error", err))

			// Wait before retry
			select {
			case <-timeoutCtx.Done():
				return timeoutCtx.Err()
			case <-time.After(es.Config.RetryDelay):
				// Continue to next retry
			}
		}
	}

	es.mutex.Lock()
	es.FailedCount++
	es.mutex.Unlock()

	logger.Error("Event handling failed after all retries",
		logger.String("subscriber_id", es.ID),
		logger.String("event_type", event.EventType()),
		logger.String("event_id", event.EventID()),
		logger.Int("max_retries", es.Config.MaxRetries),
		logger.Error2("error", err))

	return err
}

// GetStats returns statistics for the subscriber
func (es *EventSubscriber) GetStats() SubscriberStats {
	es.mutex.RLock()
	defer es.mutex.RUnlock()

	return SubscriberStats{
		ID:             es.ID,
		EventTypes:     es.EventTypes,
		IsActive:       es.IsActive,
		CreatedAt:      es.CreatedAt,
		LastEventAt:    es.LastEventAt,
		ProcessedCount: es.ProcessedCount,
		FailedCount:    es.FailedCount,
		SuccessRate:    es.calculateSuccessRate(),
	}
}

func (es *EventSubscriber) calculateSuccessRate() float64 {
	total := es.ProcessedCount + es.FailedCount
	if total == 0 {
		return 0
	}
	return float64(es.ProcessedCount) / float64(total) * 100
}

// SubscriberStats contains statistics for a subscriber
type SubscriberStats struct {
	ID             string    `json:"id"`
	EventTypes     []string  `json:"event_types"`
	IsActive       bool      `json:"is_active"`
	CreatedAt      time.Time `json:"created_at"`
	LastEventAt    time.Time `json:"last_event_at"`
	ProcessedCount int64     `json:"processed_count"`
	FailedCount    int64     `json:"failed_count"`
	SuccessRate    float64   `json:"success_rate"`
}

// Enhanced EventBus with subscriber management
type EnhancedEventBus struct {
	*InMemoryEventBus
	subscriberManager *SubscriberManager
}

// NewEnhancedEventBus creates a new enhanced event bus with subscriber management
func NewEnhancedEventBus() *EnhancedEventBus {
	return &EnhancedEventBus{
		InMemoryEventBus:  NewInMemoryEventBus(),
		subscriberManager: NewSubscriberManager(),
	}
}

// SubscribeWithConfig subscribes with configuration
func (bus *EnhancedEventBus) SubscribeWithConfig(id string, eventTypes []string, handler EventHandler, config SubscriberConfig) error {
	// Add to subscriber manager
	if err := bus.subscriberManager.AddSubscriber(id, eventTypes, handler, config); err != nil {
		return err
	}

	// Subscribe to event bus
	return bus.InMemoryEventBus.Subscribe(eventTypes, handler)
}

// UnsubscribeByID unsubscribes a subscriber by ID
func (bus *EnhancedEventBus) UnsubscribeByID(id string) error {
	subscriber, err := bus.subscriberManager.GetSubscriber(id)
	if err != nil {
		return err
	}

	// Unsubscribe from event bus
	if err := bus.InMemoryEventBus.Unsubscribe(subscriber.EventTypes, subscriber.Handler); err != nil {
		return err
	}

	// Remove from subscriber manager
	return bus.subscriberManager.RemoveSubscriber(id)
}

// GetSubscriberStats returns statistics for all subscribers
func (bus *EnhancedEventBus) GetSubscriberStats() map[string]SubscriberStats {
	subscribers := bus.subscriberManager.ListSubscribers()
	stats := make(map[string]SubscriberStats)

	for id, subscriber := range subscribers {
		stats[id] = subscriber.GetStats()
	}

	return stats
}

// HealthCheck performs a health check on all subscribers
func (bus *EnhancedEventBus) HealthCheck() map[string]bool {
	subscribers := bus.subscriberManager.ListSubscribers()
	health := make(map[string]bool)

	for id, subscriber := range subscribers {
		// Consider a subscriber healthy if it's active and has processed events recently
		isHealthy := subscriber.IsActive &&
			(subscriber.ProcessedCount > 0 || time.Since(subscriber.CreatedAt) < time.Minute*5)
		health[id] = isHealthy
	}

	return health
}