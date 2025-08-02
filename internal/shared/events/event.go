package events

import (
	"context"
	"encoding/json"
	"time"
)

// Event represents a domain event
type Event interface {
	// EventType returns the type of the event
	EventType() string
	// EventData returns the event data
	EventData() interface{}
	// EventTime returns when the event occurred
	EventTime() time.Time
	// EventID returns a unique identifier for the event
	EventID() string
	// EventVersion returns the event schema version
	EventVersion() string
	// EventSource returns the source that generated the event
	EventSource() string
}

// BaseEvent provides a basic implementation of Event interface
type BaseEvent struct {
	ID        string      `json:"id"`
	Type      string      `json:"type"`
	Source    string      `json:"source"`
	Time      time.Time   `json:"time"`
	Version   string      `json:"version"`
	Data      interface{} `json:"data"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

func (e *BaseEvent) EventType() string {
	return e.Type
}

func (e *BaseEvent) EventData() interface{} {
	return e.Data
}

func (e *BaseEvent) EventTime() time.Time {
	return e.Time
}

func (e *BaseEvent) EventID() string {
	return e.ID
}

func (e *BaseEvent) EventVersion() string {
	return e.Version
}

func (e *BaseEvent) EventSource() string {
	return e.Source
}

// NewBaseEvent creates a new base event
func NewBaseEvent(eventType, source string, data interface{}) *BaseEvent {
	return &BaseEvent{
		ID:       generateEventID(),
		Type:     eventType,
		Source:   source,
		Time:     time.Now(),
		Version:  "1.0",
		Data:     data,
		Metadata: make(map[string]interface{}),
	}
}

// SetMetadata sets metadata for the event
func (e *BaseEvent) SetMetadata(key string, value interface{}) {
	if e.Metadata == nil {
		e.Metadata = make(map[string]interface{})
	}
	e.Metadata[key] = value
}

// GetMetadata gets metadata from the event
func (e *BaseEvent) GetMetadata(key string) (interface{}, bool) {
	if e.Metadata == nil {
		return nil, false
	}
	value, exists := e.Metadata[key]
	return value, exists
}

// EventHandler defines the interface for event handlers
type EventHandler interface {
	Handle(ctx context.Context, event Event) error
	EventTypes() []string
}

// EventHandlerFunc is an adapter to allow the use of ordinary functions as EventHandlers
type EventHandlerFunc struct {
	HandlerFunc func(ctx context.Context, event Event) error
	Types       []string
}

func (h EventHandlerFunc) Handle(ctx context.Context, event Event) error {
	return h.HandlerFunc(ctx, event)
}

func (h EventHandlerFunc) EventTypes() []string {
	return h.Types
}

// NewEventHandler creates a new event handler function
func NewEventHandler(eventTypes []string, handlerFunc func(ctx context.Context, event Event) error) EventHandler {
	return EventHandlerFunc{
		HandlerFunc: handlerFunc,
		Types:       eventTypes,
	}
}

// EventEnvelope wraps an event with additional context information
type EventEnvelope struct {
	Event     Event                  `json:"event"`
	Context   map[string]interface{} `json:"context,omitempty"`
	Headers   map[string]string      `json:"headers,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
}

// NewEventEnvelope creates a new event envelope
func NewEventEnvelope(event Event) *EventEnvelope {
	return &EventEnvelope{
		Event:     event,
		Context:   make(map[string]interface{}),
		Headers:   make(map[string]string),
		CreatedAt: time.Now(),
	}
}

// SerializeEvent serializes an event to JSON
func SerializeEvent(event Event) ([]byte, error) {
	envelope := NewEventEnvelope(event)
	return json.Marshal(envelope)
}

// DeserializeEvent deserializes JSON to an event envelope
func DeserializeEvent(data []byte) (*EventEnvelope, error) {
	var envelope EventEnvelope
	err := json.Unmarshal(data, &envelope)
	if err != nil {
		return nil, err
	}
	return &envelope, nil
}

// Domain-specific event types

// UserEvent represents user-related events
type UserEvent struct {
	*BaseEvent
	UserID uint `json:"user_id"`
}

// NewUserEvent creates a new user event
func NewUserEvent(eventType string, userID uint, data interface{}) *UserEvent {
	return &UserEvent{
		BaseEvent: NewBaseEvent(eventType, "user-service", data),
		UserID:    userID,
	}
}

// PaymentEvent represents payment-related events
type PaymentEvent struct {
	*BaseEvent
	PaymentID string  `json:"payment_id"`
	Amount    float64 `json:"amount"`
	UserID    uint    `json:"user_id"`
}

// NewPaymentEvent creates a new payment event
func NewPaymentEvent(eventType string, paymentID string, amount float64, userID uint, data interface{}) *PaymentEvent {
	return &PaymentEvent{
		BaseEvent: NewBaseEvent(eventType, "payment-service", data),
		PaymentID: paymentID,
		Amount:    amount,
		UserID:    userID,
	}
}

// SubscriptionEvent represents subscription-related events
type SubscriptionEvent struct {
	*BaseEvent
	SubscriptionID uint `json:"subscription_id"`
	UserID         uint `json:"user_id"`
}

// NewSubscriptionEvent creates a new subscription event
func NewSubscriptionEvent(eventType string, subscriptionID uint, userID uint, data interface{}) *SubscriptionEvent {
	return &SubscriptionEvent{
		BaseEvent:      NewBaseEvent(eventType, "subscription-service", data),
		SubscriptionID: subscriptionID,
		UserID:         userID,
	}
}

// Event type constants
const (
	// User events
	EventTypeUserRegistered = "user.registered"
	EventTypeUserUpdated    = "user.updated"
	EventTypeUserDeleted    = "user.deleted"
	EventTypeUserLoggedIn   = "user.logged_in"
	EventTypeUserLoggedOut  = "user.logged_out"

	// Payment events
	EventTypePaymentCreated   = "payment.created"
	EventTypePaymentCompleted = "payment.completed"
	EventTypePaymentFailed    = "payment.failed"
	EventTypePaymentRefunded  = "payment.refunded"

	// Subscription events
	EventTypeSubscriptionCreated  = "subscription.created"
	EventTypeSubscriptionUpdated  = "subscription.updated"
	EventTypeSubscriptionExpired  = "subscription.expired"
	EventTypeSubscriptionCanceled = "subscription.canceled"
	EventTypeSubscriptionRenewed  = "subscription.renewed"

	// Server events
	EventTypeServerCreated = "server.created"
	EventTypeServerUpdated = "server.updated"
	EventTypeServerDeleted = "server.deleted"

	// Referral events
	EventTypeReferralCreated   = "referral.created"
	EventTypeReferralCompleted = "referral.completed"

	// Invoice events
	EventTypeInvoiceGenerated = "invoice.generated"
	EventTypeInvoiceSent      = "invoice.sent"

	// Ticket events
	EventTypeTicketCreated = "ticket.created"
	EventTypeTicketUpdated = "ticket.updated"
	EventTypeTicketClosed  = "ticket.closed"
)

// Helper function to generate event IDs
func generateEventID() string {
	// Simple ID generation - in production, use a proper UUID library
	return time.Now().Format("20060102150405") + "-" + "event"
}