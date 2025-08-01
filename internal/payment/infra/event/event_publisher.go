package event

import (
	"context"
	"fmt"
	"log"

	"linke/internal/shared/domain"
)

// EventPublisher implements domain.EventPublisher for payment events
type EventPublisher struct {
	handlers map[string][]EventHandler
}

// EventHandler defines the interface for event handlers
type EventHandler interface {
	Handle(ctx context.Context, event domain.DomainEvent) error
}

// NewEventPublisher creates a new EventPublisher
func NewEventPublisher() *EventPublisher {
	return &EventPublisher{
		handlers: make(map[string][]EventHandler),
	}
}

// RegisterHandler registers an event handler for a specific event type
func (p *EventPublisher) RegisterHandler(eventType string, handler EventHandler) {
	p.handlers[eventType] = append(p.handlers[eventType], handler)
}

// Publish publishes a domain event
func (p *EventPublisher) Publish(ctx context.Context, event domain.DomainEvent) error {
	handlers, exists := p.handlers[event.EventType()]
	if !exists {
		// Log that no handlers are registered for this event type
		log.Printf("No handlers registered for event type: %s", event.EventType())
		return nil
	}

	// Execute all handlers for this event type
	for _, handler := range handlers {
		if err := handler.Handle(ctx, event); err != nil {
			// Log the error but continue with other handlers
			log.Printf("Error handling event %s with handler %T: %v", event.EventType(), handler, err)
			// You might want to implement retry logic or circuit breaker here
		}
	}

	return nil
}

// PublishBatch publishes multiple domain events
func (p *EventPublisher) PublishBatch(ctx context.Context, events []domain.DomainEvent) error {
	var lastError error
	
	for _, event := range events {
		if err := p.Publish(ctx, event); err != nil {
			lastError = err
			log.Printf("Error publishing event %s: %v", event.EventType(), err)
		}
	}

	return lastError
}

// EventBus implements domain.EventBus for payment events
type EventBus struct {
	publisher *EventPublisher
}

// NewEventBus creates a new EventBus
func NewEventBus(publisher *EventPublisher) *EventBus {
	return &EventBus{
		publisher: publisher,
	}
}

// Publish publishes one or more domain events
func (b *EventBus) Publish(ctx context.Context, events ...domain.DomainEvent) error {
	if len(events) == 0 {
		return nil
	}

	if len(events) == 1 {
		return b.publisher.Publish(ctx, events[0])
	}

	return b.publisher.PublishBatch(ctx, events)
}

// Example event handlers

// PaymentCreatedHandler handles PaymentCreated events
type PaymentCreatedHandler struct{}

// Handle handles the PaymentCreated event
func (h *PaymentCreatedHandler) Handle(ctx context.Context, event domain.DomainEvent) error {
	log.Printf("Handling PaymentCreated event: %s", event.EventID())
	
	// Add your business logic here, such as:
	// - Sending notifications
	// - Updating invoice status
	// - Triggering payment processing
	// - Analytics/metrics collection
	
	return nil
}

// PaymentCompletedHandler handles PaymentCompleted events
type PaymentCompletedHandler struct{}

// Handle handles the PaymentCompleted event
func (h *PaymentCompletedHandler) Handle(ctx context.Context, event domain.DomainEvent) error {
	log.Printf("Handling PaymentCompleted event: %s", event.EventID())
	
	// Add your business logic here, such as:
	// - Activating services
	// - Sending confirmation emails
	// - Updating subscription status
	// - Generating receipts
	
	return nil
}

// PaymentFailedHandler handles PaymentFailed events
type PaymentFailedHandler struct{}

// Handle handles the PaymentFailed event
func (h *PaymentFailedHandler) Handle(ctx context.Context, event domain.DomainEvent) error {
	log.Printf("Handling PaymentFailed event: %s", event.EventID())
	
	// Add your business logic here, such as:
	// - Sending failure notifications
	// - Cleaning up resources
	// - Updating payment status
	// - Triggering retry mechanisms
	
	return nil
}

// PaymentRefundedHandler handles PaymentRefunded events
type PaymentRefundedHandler struct{}

// Handle handles the PaymentRefunded event
func (h *PaymentRefundedHandler) Handle(ctx context.Context, event domain.DomainEvent) error {
	log.Printf("Handling PaymentRefunded event: %s", event.EventID())
	
	// Add your business logic here, such as:
	// - Processing actual refund through gateway
	// - Updating service status
	// - Sending refund notifications
	// - Updating accounting records
	
	return nil
}

// PaymentConfigCreatedHandler handles PaymentConfigCreated events
type PaymentConfigCreatedHandler struct{}

// Handle handles the PaymentConfigCreated event
func (h *PaymentConfigCreatedHandler) Handle(ctx context.Context, event domain.DomainEvent) error {
	log.Printf("Handling PaymentConfigCreated event: %s", event.EventID())
	
	// Add your business logic here, such as:
	// - Refreshing payment method cache
	// - Notifying administrators
	// - Updating payment UI
	
	return nil
}

// SetupEventHandlers sets up all event handlers
func SetupEventHandlers(publisher *EventPublisher) {
	// Register payment event handlers
	publisher.RegisterHandler("PaymentCreated", &PaymentCreatedHandler{})
	publisher.RegisterHandler("PaymentCompleted", &PaymentCompletedHandler{})
	publisher.RegisterHandler("PaymentFailed", &PaymentFailedHandler{})
	publisher.RegisterHandler("PaymentRefunded", &PaymentRefundedHandler{})
	
	// Register payment config event handlers
	publisher.RegisterHandler("PaymentConfigCreated", &PaymentConfigCreatedHandler{})
	
	// Add more handlers as needed...
}

// AsyncEventPublisher implements asynchronous event publishing
type AsyncEventPublisher struct {
	publisher *EventPublisher
	eventChan chan EventWithContext
	workers   int
}

// EventWithContext wraps an event with its context
type EventWithContext struct {
	Event   domain.DomainEvent
	Context context.Context
}

// NewAsyncEventPublisher creates a new AsyncEventPublisher
func NewAsyncEventPublisher(publisher *EventPublisher, workers int) *AsyncEventPublisher {
	if workers <= 0 {
		workers = 1
	}
	
	return &AsyncEventPublisher{
		publisher: publisher,
		eventChan: make(chan EventWithContext, 1000), // Buffer size
		workers:   workers,
	}
}

// Start starts the async event publisher workers
func (p *AsyncEventPublisher) Start() {
	for i := 0; i < p.workers; i++ {
		go p.worker()
	}
}

// worker processes events from the channel
func (p *AsyncEventPublisher) worker() {
	for eventWithCtx := range p.eventChan {
		if err := p.publisher.Publish(eventWithCtx.Context, eventWithCtx.Event); err != nil {
			log.Printf("Error publishing event asynchronously: %v", err)
		}
	}
}

// Publish publishes an event asynchronously
func (p *AsyncEventPublisher) Publish(ctx context.Context, event domain.DomainEvent) error {
	select {
	case p.eventChan <- EventWithContext{Event: event, Context: ctx}:
		return nil
	default:
		return fmt.Errorf("event channel is full, dropping event: %s", event.EventID())
	}
}

// PublishBatch publishes multiple events asynchronously
func (p *AsyncEventPublisher) PublishBatch(ctx context.Context, events []domain.DomainEvent) error {
	for _, event := range events {
		if err := p.Publish(ctx, event); err != nil {
			return err
		}
	}
	return nil
}

// Stop stops the async event publisher
func (p *AsyncEventPublisher) Stop() {
	close(p.eventChan)
}