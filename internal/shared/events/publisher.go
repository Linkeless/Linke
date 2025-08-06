package events

import (
	"context"
	"fmt"
	"sync"

	"linke/internal/shared/logger"
)

// Publisher defines the interface for event publishers
type Publisher interface {
	Publish(ctx context.Context, event Event) error
	PublishAsync(ctx context.Context, event Event) error
	Close() error
}

// Subscriber defines the interface for event subscribers
type Subscriber interface {
	Subscribe(eventTypes []string, handler EventHandler) error
	Unsubscribe(eventTypes []string, handler EventHandler) error
	Close() error
}

// EventBus combines publisher and subscriber interfaces
type EventBus interface {
	Publisher
	Subscriber
}

// InMemoryEventBus provides an in-memory event bus implementation
type InMemoryEventBus struct {
	handlers map[string][]EventHandler
	mutex    sync.RWMutex
	logger   logger.Logger
}

// NewInMemoryEventBus creates a new in-memory event bus
func NewInMemoryEventBus() *InMemoryEventBus {
	return &InMemoryEventBus{
		handlers: make(map[string][]EventHandler),
		logger:   logger.GetGlobalLogger(),
	}
}

// Publish publishes an event synchronously
func (bus *InMemoryEventBus) Publish(ctx context.Context, event Event) error {
	bus.mutex.RLock()
	handlers, exists := bus.handlers[event.EventType()]
	bus.mutex.RUnlock()

	if !exists || len(handlers) == 0 {
		bus.logger.Debug("No handlers registered for event type",
			logger.String("event_type", event.EventType()),
			logger.String("event_id", event.EventID()))
		return nil
	}

	bus.logger.Info("Publishing event",
		logger.String("event_type", event.EventType()),
		logger.String("event_id", event.EventID()),
		logger.Int("handler_count", len(handlers)))

	var errors []error
	for _, handler := range handlers {
		if err := handler.Handle(ctx, event); err != nil {
			bus.logger.Error("Event handler failed",
				logger.String("event_type", event.EventType()),
				logger.String("event_id", event.EventID()),
				logger.ErrorField(err))
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("event handling failed: %v", errors)
	}

	return nil
}

// PublishAsync publishes an event asynchronously
func (bus *InMemoryEventBus) PublishAsync(ctx context.Context, event Event) error {
	go func() {
		if err := bus.Publish(ctx, event); err != nil {
			bus.logger.Error("Async event publishing failed",
				logger.String("event_type", event.EventType()),
				logger.String("event_id", event.EventID()),
				logger.ErrorField(err))
		}
	}()
	return nil
}

// Subscribe registers an event handler for specific event types
func (bus *InMemoryEventBus) Subscribe(eventTypes []string, handler EventHandler) error {
	bus.mutex.Lock()
	defer bus.mutex.Unlock()

	for _, eventType := range eventTypes {
		if _, exists := bus.handlers[eventType]; !exists {
			bus.handlers[eventType] = nil
		}
		bus.handlers[eventType] = append(bus.handlers[eventType], handler)

		bus.logger.Debug("Event handler subscribed",
			logger.String("event_type", eventType))
	}

	return nil
}

// Unsubscribe removes an event handler for specific event types
func (bus *InMemoryEventBus) Unsubscribe(eventTypes []string, handler EventHandler) error {
	bus.mutex.Lock()
	defer bus.mutex.Unlock()

	for _, eventType := range eventTypes {
		if handlers, exists := bus.handlers[eventType]; exists {
			for i, h := range handlers {
				if h.ID() == handler.ID() {
					// Remove handler from slice
					bus.handlers[eventType] = append(handlers[:i], handlers[i+1:]...)
					bus.logger.Debug("Event handler unsubscribed",
						logger.String("event_type", eventType),
						logger.String("handler_id", handler.ID()))
					break
				}
			}
		}
	}

	return nil
}

// Close closes the event bus
func (bus *InMemoryEventBus) Close() error {
	bus.mutex.Lock()
	defer bus.mutex.Unlock()

	bus.handlers = make(map[string][]EventHandler)
	bus.logger.Info("Event bus closed")
	return nil
}

// GetHandlerCount returns the number of handlers for an event type
func (bus *InMemoryEventBus) GetHandlerCount(eventType string) int {
	bus.mutex.RLock()
	defer bus.mutex.RUnlock()

	if handlers, exists := bus.handlers[eventType]; exists {
		return len(handlers)
	}
	return 0
}

// ListEventTypes returns all registered event types
func (bus *InMemoryEventBus) ListEventTypes() []string {
	bus.mutex.RLock()
	defer bus.mutex.RUnlock()

	var eventTypes []string
	for eventType := range bus.handlers {
		eventTypes = append(eventTypes, eventType)
	}
	return eventTypes
}

// RedisEventBus provides a Redis-based event bus implementation
type RedisEventBus struct {
	// This would implement Redis pub/sub for distributed events
	// For now, we'll use the in-memory implementation
	*InMemoryEventBus
}

// NewRedisEventBus creates a new Redis-based event bus
func NewRedisEventBus() *RedisEventBus {
	return &RedisEventBus{
		InMemoryEventBus: NewInMemoryEventBus(),
	}
}

// EventMiddleware provides middleware functionality for events
type EventMiddleware interface {
	Process(ctx context.Context, event Event, next func(context.Context, Event) error) error
}

// EventMiddlewareFunc is an adapter for event middleware functions
type EventMiddlewareFunc func(ctx context.Context, event Event, next func(context.Context, Event) error) error

func (f EventMiddlewareFunc) Process(ctx context.Context, event Event, next func(context.Context, Event) error) error {
	return f(ctx, event, next)
}

// LoggingMiddleware logs events
func LoggingMiddleware() EventMiddleware {
	return EventMiddlewareFunc(func(ctx context.Context, event Event, next func(context.Context, Event) error) error {
		logger.Info("Processing event",
			logger.String("event_type", event.EventType()),
			logger.String("event_id", event.EventID()),
			logger.String("event_source", event.EventSource()))

		err := next(ctx, event)

		if err != nil {
			logger.Error("Event processing failed",
				logger.String("event_type", event.EventType()),
				logger.String("event_id", event.EventID()),
				logger.ErrorField(err))
		} else {
			logger.Info("Event processed successfully",
				logger.String("event_type", event.EventType()),
				logger.String("event_id", event.EventID()))
		}

		return err
	})
}

// RetryMiddleware provides retry functionality for failed events
func RetryMiddleware(maxRetries int) EventMiddleware {
	return EventMiddlewareFunc(func(ctx context.Context, event Event, next func(context.Context, Event) error) error {
		var err error
		for i := 0; i <= maxRetries; i++ {
			err = next(ctx, event)
			if err == nil {
				return nil
			}

			if i < maxRetries {
				logger.Warn("Event processing failed, retrying",
					logger.String("event_type", event.EventType()),
					logger.String("event_id", event.EventID()),
					logger.Int("attempt", i+1),
					logger.Int("max_retries", maxRetries),
					logger.ErrorField(err))
			}
		}

		logger.Error("Event processing failed after all retries",
			logger.String("event_type", event.EventType()),
			logger.String("event_id", event.EventID()),
			logger.Int("max_retries", maxRetries),
			logger.Error2("error", err))

		return err
	})
}

// Global event bus instance
var globalEventBus EventBus

// InitEventBus initializes the global event bus
func InitEventBus(bus EventBus) {
	globalEventBus = bus
}

// GetEventBus returns the global event bus
func GetEventBus() EventBus {
	if globalEventBus == nil {
		globalEventBus = NewInMemoryEventBus()
	}
	return globalEventBus
}

// Convenience functions for global event bus

// Publish publishes an event using the global event bus
func Publish(ctx context.Context, event Event) error {
	return GetEventBus().Publish(ctx, event)
}

// PublishAsync publishes an event asynchronously using the global event bus
func PublishAsync(ctx context.Context, event Event) error {
	return GetEventBus().PublishAsync(ctx, event)
}

// Subscribe subscribes to events using the global event bus
func Subscribe(eventTypes []string, handler EventHandler) error {
	return GetEventBus().Subscribe(eventTypes, handler)
}
