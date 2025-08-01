package event

import (
	"context"
	"log"

	"linke/internal/server/domain/event"
)

// SimpleEventPublisher is a simple implementation of EventPublisher
// In a production system, this might publish to a message queue, event store, etc.
type SimpleEventPublisher struct {
	handlers map[string][]EventHandler
}

// EventHandler defines the interface for handling domain events
type EventHandler interface {
	Handle(ctx context.Context, event event.DomainEvent) error
}

// NewSimpleEventPublisher creates a new simple event publisher
func NewSimpleEventPublisher() *SimpleEventPublisher {
	return &SimpleEventPublisher{
		handlers: make(map[string][]EventHandler),
	}
}

// RegisterHandler registers an event handler for a specific event type
func (p *SimpleEventPublisher) RegisterHandler(eventType string, handler EventHandler) {
	if p.handlers[eventType] == nil {
		p.handlers[eventType] = make([]EventHandler, 0)
	}
	p.handlers[eventType] = append(p.handlers[eventType], handler)
}

// PublishEvents publishes domain events
func (p *SimpleEventPublisher) PublishEvents(ctx context.Context, events []interface{}) error {
	for _, evt := range events {
		if domainEvent, ok := evt.(event.DomainEvent); ok {
			if err := p.publishEvent(ctx, domainEvent); err != nil {
				// Log error but continue processing other events
				log.Printf("Failed to publish event %s: %v", domainEvent.EventType(), err)
			}
		}
	}
	return nil
}

// publishEvent publishes a single domain event
func (p *SimpleEventPublisher) publishEvent(ctx context.Context, evt event.DomainEvent) error {
	handlers, exists := p.handlers[evt.EventType()]
	if !exists {
		// No handlers registered for this event type
		return nil
	}
	
	for _, handler := range handlers {
		if err := handler.Handle(ctx, evt); err != nil {
			return err
		}
	}
	
	return nil
}