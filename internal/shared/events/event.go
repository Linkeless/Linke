package events

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// Event represents a domain event
type Event interface {
	// EventType returns the type of the event
	EventType() string
	// EventData returns the event data
	EventData() any
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
	ID       string         `json:"id"`
	Type     string         `json:"type"`
	Source   string         `json:"source"`
	Time     time.Time      `json:"time"`
	Version  string         `json:"version"`
	Data     any            `json:"data"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

func (e *BaseEvent) EventType() string {
	return e.Type
}

func (e *BaseEvent) EventData() any {
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
func NewBaseEvent(eventType, source string, data any) *BaseEvent {
	return &BaseEvent{
		ID:       generateEventID(),
		Type:     eventType,
		Source:   source,
		Time:     time.Now(),
		Version:  "1.0",
		Data:     data,
		Metadata: make(map[string]any),
	}
}

// SetMetadata sets metadata for the event
func (e *BaseEvent) SetMetadata(key string, value any) {
	if e.Metadata == nil {
		e.Metadata = make(map[string]any)
	}
	e.Metadata[key] = value
}

// GetMetadata gets metadata from the event
func (e *BaseEvent) GetMetadata(key string) (any, bool) {
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
	ID() string
}

// EventHandlerFunc is an adapter to allow the use of ordinary functions as EventHandlers
type EventHandlerFunc struct {
	HandlerFunc func(ctx context.Context, event Event) error
	Types       []string
	HandlerID   string // Unique identifier for comparison
}

func (h EventHandlerFunc) Handle(ctx context.Context, event Event) error {
	return h.HandlerFunc(ctx, event)
}

func (h EventHandlerFunc) EventTypes() []string {
	return h.Types
}

func (h EventHandlerFunc) ID() string {
	return h.HandlerID
}

// NewEventHandler creates a new event handler function
func NewEventHandler(eventTypes []string, handlerFunc func(ctx context.Context, event Event) error) EventHandler {
	return EventHandlerFunc{
		HandlerFunc: handlerFunc,
		Types:       eventTypes,
		HandlerID:   generateEventID(), // Generate unique ID for handler
	}
}

// EventEnvelope wraps an event with additional context information
type EventEnvelope struct {
	Event     Event             `json:"-"` // Skip JSON serialization, handled by custom methods
	Context   map[string]any    `json:"context,omitempty"`
	Headers   map[string]string `json:"headers,omitempty"`
	CreatedAt time.Time         `json:"created_at"`
}

// eventEnvelopeJSON is used for JSON serialization/deserialization
type eventEnvelopeJSON struct {
	EventType string            `json:"event_type"`
	EventData json.RawMessage   `json:"event_data"`
	Context   map[string]any    `json:"context,omitempty"`
	Headers   map[string]string `json:"headers,omitempty"`
	CreatedAt time.Time         `json:"created_at"`
}

// NewEventEnvelope creates a new event envelope
func NewEventEnvelope(event Event) *EventEnvelope {
	return &EventEnvelope{
		Event:     event,
		Context:   make(map[string]any),
		Headers:   make(map[string]string),
		CreatedAt: time.Now(),
	}
}

// MarshalJSON implements custom JSON marshaling for EventEnvelope
func (e *EventEnvelope) MarshalJSON() ([]byte, error) {
	if e.Event == nil {
		return nil, fmt.Errorf("event cannot be nil")
	}

	// Serialize the event to JSON
	eventData, err := json.Marshal(e.Event)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal event: %w", err)
	}

	// Create the JSON representation
	envJSON := eventEnvelopeJSON{
		EventType: e.Event.EventType(),
		EventData: eventData,
		Context:   e.Context,
		Headers:   e.Headers,
		CreatedAt: e.CreatedAt,
	}

	return json.Marshal(envJSON)
}

// UnmarshalJSON implements custom JSON unmarshaling for EventEnvelope
func (e *EventEnvelope) UnmarshalJSON(data []byte) error {
	var envJSON eventEnvelopeJSON
	if err := json.Unmarshal(data, &envJSON); err != nil {
		return fmt.Errorf("failed to unmarshal event envelope: %w", err)
	}

	// Reconstruct the event based on the raw event data
	// We'll unmarshal into a BaseEvent since we can't determine the exact type
	// This is a limitation, but the Event interface methods will still work
	var baseEvent BaseEvent
	if err := json.Unmarshal(envJSON.EventData, &baseEvent); err != nil {
		return fmt.Errorf("failed to unmarshal event data: %w", err)
	}

	// Set the envelope fields
	e.Event = &baseEvent
	e.Context = envJSON.Context
	e.Headers = envJSON.Headers
	e.CreatedAt = envJSON.CreatedAt

	return nil
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
func NewUserEvent(eventType string, userID uint, data any) *UserEvent {
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
func NewPaymentEvent(eventType, paymentID string, amount float64, userID uint, data any) *PaymentEvent {
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
func NewSubscriptionEvent(eventType string, subscriptionID, userID uint, data any) *SubscriptionEvent {
	return &SubscriptionEvent{
		BaseEvent:      NewBaseEvent(eventType, "subscription-service", data),
		SubscriptionID: subscriptionID,
		UserID:         userID,
	}
}

// Event type constants
const (
	// User events
	EventTypeUserCreated       = "user.created"
	EventTypeUserRegistered    = "user.registered"
	EventTypeUserUpdated       = "user.updated"
	EventTypeUserDeleted       = "user.deleted"
	EventTypeUserStatusChanged = "user.status_changed"
	EventTypeUserLoggedIn      = "user.logged_in"
	EventTypeUserLoggedOut     = "user.logged_out"
	EventTypeUserPasswordReset = "user.password_reset"

	// Payment events
	EventTypePaymentCreated   = "payment.created"
	EventTypePaymentCompleted = "payment.completed"
	EventTypePaymentFailed    = "payment.failed"
	EventTypePaymentRefunded  = "payment.refunded"

	// Subscription events
	EventTypeSubscriptionCreated   = "subscription.created"
	EventTypeSubscriptionUpdated   = "subscription.updated"
	EventTypeSubscriptionActivated = "subscription.activated"
	EventTypeSubscriptionPaused    = "subscription.paused"
	EventTypeSubscriptionResumed   = "subscription.resumed"
	EventTypeSubscriptionExpired   = "subscription.expired"
	EventTypeSubscriptionCancelled = "subscription.cancelled"
	EventTypeSubscriptionRenewed   = "subscription.renewed"
	EventTypeSubscriptionSuspended = "subscription.suspended"

	// Order events
	EventTypeOrderCreated   = "order.created"
	EventTypeOrderUpdated   = "order.updated"
	EventTypeOrderPaid      = "order.paid"
	EventTypeOrderCancelled = "order.cancelled"
	EventTypeOrderExpired   = "order.expired"
	EventTypeOrderRefunded  = "order.refunded"

	// Server events
	EventTypeServerCreated = "server.created"
	EventTypeServerUpdated = "server.updated"
	EventTypeServerDeleted = "server.deleted"

	// Referral events
	EventTypeReferralCreated   = "referral.created"
	EventTypeReferralCompleted = "referral.completed"

	// Invoice events
	EventTypeInvoiceCreated   = "invoice.created"
	EventTypeInvoiceGenerated = "invoice.generated"
	EventTypeInvoiceSent      = "invoice.sent"
	EventTypeInvoicePaid      = "invoice.paid"
	EventTypeInvoiceOverdue   = "invoice.overdue"
	EventTypeInvoiceCancelled = "invoice.cancelled"

	// Ticket events
	EventTypeTicketCreated = "ticket.created"
	EventTypeTicketUpdated = "ticket.updated"
	EventTypeTicketClosed  = "ticket.closed"
)

// Order Event represents order-related events
type OrderEvent struct {
	*BaseEvent
	OrderID uint `json:"order_id"`
	UserID  uint `json:"user_id"`
}

// NewOrderEvent creates a new order event
func NewOrderEvent(eventType string, orderID, userID uint, data any) *OrderEvent {
	return &OrderEvent{
		BaseEvent: NewBaseEvent(eventType, "order-service", data),
		OrderID:   orderID,
		UserID:    userID,
	}
}

// Invoice Event represents invoice-related events
type InvoiceEvent struct {
	*BaseEvent
	InvoiceID uint    `json:"invoice_id"`
	OrderID   uint    `json:"order_id"`
	UserID    uint    `json:"user_id"`
	Amount    float64 `json:"amount"`
}

// NewInvoiceEvent creates a new invoice event
func NewInvoiceEvent(eventType string, invoiceID, orderID, userID uint, amount float64, data any) *InvoiceEvent {
	return &InvoiceEvent{
		BaseEvent: NewBaseEvent(eventType, "invoice-service", data),
		InvoiceID: invoiceID,
		OrderID:   orderID,
		UserID:    userID,
		Amount:    amount,
	}
}

// ServerEvent represents server-related events
type ServerEvent struct {
	*BaseEvent
	ServerID uint `json:"server_id"`
}

// NewServerEvent creates a new server event
func NewServerEvent(eventType string, serverID uint, data any) *ServerEvent {
	return &ServerEvent{
		BaseEvent: NewBaseEvent(eventType, "server-service", data),
		ServerID:  serverID,
	}
}

// Helper function to generate event IDs
func generateEventID() string {
	// Generate ID with nanosecond precision to ensure uniqueness
	// Format: YYYYMMDDHHMMSS-nanoseconds-event
	now := time.Now()
	return now.Format("20060102150405") + "-" + fmt.Sprintf("%d", now.UnixNano()) + "-event"
}
