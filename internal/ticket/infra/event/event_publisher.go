package event

import (
	"context"
	"fmt"
	"log"
)

// EventPublisher publishes domain events
type EventPublisher struct {
	// In a real implementation, this might use a message broker like RabbitMQ, Kafka, etc.
	// For now, we'll use a simple in-memory implementation
	handlers map[string][]EventHandler
}

// EventHandler handles domain events
type EventHandler interface {
	Handle(ctx context.Context, event interface{}) error
}

// NewEventPublisher creates a new event publisher
func NewEventPublisher() *EventPublisher {
	return &EventPublisher{
		handlers: make(map[string][]EventHandler),
	}
}

// Subscribe subscribes a handler to events of a specific type
func (p *EventPublisher) Subscribe(eventType string, handler EventHandler) {
	p.handlers[eventType] = append(p.handlers[eventType], handler)
}

// Publish publishes events to registered handlers
func (p *EventPublisher) Publish(ctx context.Context, events ...interface{}) error {
	for _, event := range events {
		if err := p.publishSingle(ctx, event); err != nil {
			return fmt.Errorf("failed to publish event %T: %w", event, err)
		}
	}
	return nil
}

// publishSingle publishes a single event
func (p *EventPublisher) publishSingle(ctx context.Context, event interface{}) error {
	eventType := fmt.Sprintf("%T", event)
	handlers, exists := p.handlers[eventType]
	
	if !exists {
		// No handlers registered for this event type, that's okay
		log.Printf("No handlers registered for event type: %s", eventType)
		return nil
	}
	
	// Handle the event with all registered handlers
	for _, handler := range handlers {
		if err := handler.Handle(ctx, event); err != nil {
			// Log error but continue with other handlers
			log.Printf("Handler failed for event %s: %v", eventType, err)
			// In a production system, you might want to implement retry logic
		}
	}
	
	return nil
}