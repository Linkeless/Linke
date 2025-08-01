package event

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"linke/internal/shared/domain"
)

// EventPublisher handles publishing domain events
type EventPublisher struct {
	// In a real application, this would connect to a message broker
	// like Redis, RabbitMQ, Kafka, etc.
	handlers map[string][]EventHandler
}

// EventHandler represents an event handler function
type EventHandler func(ctx context.Context, event domain.DomainEvent) error

// NewEventPublisher creates a new event publisher
func NewEventPublisher() *EventPublisher {
	return &EventPublisher{
		handlers: make(map[string][]EventHandler),
	}
}

// Subscribe subscribes a handler to an event type
func (p *EventPublisher) Subscribe(eventType string, handler EventHandler) {
	if p.handlers[eventType] == nil {
		p.handlers[eventType] = make([]EventHandler, 0)
	}
	p.handlers[eventType] = append(p.handlers[eventType], handler)
}

// Publish publishes a domain event
func (p *EventPublisher) Publish(ctx context.Context, event domain.DomainEvent) error {
	eventType := event.EventType()
	
	// Log the event for debugging
	eventJSON, _ := json.Marshal(event)
	log.Printf("Publishing event: %s - %s", eventType, string(eventJSON))

	// Get handlers for this event type
	handlers, exists := p.handlers[eventType]
	if !exists {
		// No handlers registered, which is fine
		return nil
	}

	// Call all handlers
	var lastError error
	for _, handler := range handlers {
		if err := handler(ctx, event); err != nil {
			// Log error but continue with other handlers
			log.Printf("Event handler error for %s: %v", eventType, err)
			lastError = err
		}
	}

	// Return last error if any occurred
	if lastError != nil {
		return fmt.Errorf("one or more event handlers failed: %w", lastError)
	}

	return nil
}

// PublishBatch publishes multiple events
func (p *EventPublisher) PublishBatch(ctx context.Context, events []domain.DomainEvent) error {
	for _, event := range events {
		if err := p.Publish(ctx, event); err != nil {
			return fmt.Errorf("failed to publish event %s: %w", event.EventType(), err)
		}
	}
	return nil
}

// InvoiceEventHandlers contains common invoice event handlers
type InvoiceEventHandlers struct{}

// NewInvoiceEventHandlers creates new invoice event handlers
func NewInvoiceEventHandlers() *InvoiceEventHandlers {
	return &InvoiceEventHandlers{}
}

// HandleInvoiceCreated handles invoice created events
func (h *InvoiceEventHandlers) HandleInvoiceCreated(ctx context.Context, event domain.DomainEvent) error {
	// In a real application, this might:
	// - Send notification to accounting team
	// - Update analytics/metrics
	// - Trigger PDF generation
	// - Log to audit trail
	
	log.Printf("Invoice created: %s", event.AggregateID())
	return nil
}

// HandleInvoiceSent handles invoice sent events
func (h *InvoiceEventHandlers) HandleInvoiceSent(ctx context.Context, event domain.DomainEvent) error {
	// In a real application, this might:
	// - Send email to customer
	// - Schedule payment reminders
	// - Update external systems
	// - Track delivery metrics
	
	log.Printf("Invoice sent: %s", event.AggregateID())
	return nil
}

// HandleInvoicePaid handles invoice paid events
func (h *InvoiceEventHandlers) HandleInvoicePaid(ctx context.Context, event domain.DomainEvent) error {
	// In a real application, this might:
	// - Send payment confirmation
	// - Update subscription status
	// - Trigger service activation
	// - Update financial reports
	
	log.Printf("Invoice paid: %s", event.AggregateID())
	return nil
}

// HandleInvoiceVoided handles invoice voided events
func (h *InvoiceEventHandlers) HandleInvoiceVoided(ctx context.Context, event domain.DomainEvent) error {
	// In a real application, this might:
	// - Send void notification
	// - Update accounting records
	// - Reverse any automated processes
	// - Log for compliance
	
	log.Printf("Invoice voided: %s", event.AggregateID())
	return nil
}

// HandleInvoiceOverdue handles invoice overdue events
func (h *InvoiceEventHandlers) HandleInvoiceOverdue(ctx context.Context, event domain.DomainEvent) error {
	// In a real application, this might:
	// - Send overdue reminders
	// - Escalate to collections
	// - Update customer status
	// - Generate reports
	
	log.Printf("Invoice overdue: %s", event.AggregateID())
	return nil
}

// RegisterInvoiceEventHandlers registers all invoice event handlers
func RegisterInvoiceEventHandlers(publisher *EventPublisher) {
	handlers := NewInvoiceEventHandlers()
	
	publisher.Subscribe("invoice.created", handlers.HandleInvoiceCreated)
	publisher.Subscribe("invoice.sent", handlers.HandleInvoiceSent)
	publisher.Subscribe("invoice.paid", handlers.HandleInvoicePaid)
	publisher.Subscribe("invoice.voided", handlers.HandleInvoiceVoided)
	publisher.Subscribe("invoice.overdue", handlers.HandleInvoiceOverdue)
}