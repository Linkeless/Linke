package event

import (
	"time"

	"linke/internal/payment/domain/valueobject"
	"linke/internal/shared/domain"
)

// PaymentConfigCreatedEvent represents the event when a payment config is created
type PaymentConfigCreatedEvent struct {
	eventID             string
	occurredAt          time.Time
	gateway             valueobject.PaymentGateway
	name                string
	supportedCurrencies []valueobject.Currency
}

// NewPaymentConfigCreatedEvent creates a new PaymentConfigCreatedEvent
func NewPaymentConfigCreatedEvent(
	gateway valueobject.PaymentGateway,
	name string,
	supportedCurrencies []valueobject.Currency,
	occurredAt time.Time,
) *PaymentConfigCreatedEvent {
	return &PaymentConfigCreatedEvent{
		eventID:             domain.NewEventID(),
		occurredAt:          occurredAt,
		gateway:             gateway,
		name:                name,
		supportedCurrencies: supportedCurrencies,
	}
}

func (e *PaymentConfigCreatedEvent) EventID() string {
	return e.eventID
}

func (e *PaymentConfigCreatedEvent) OccurredAt() time.Time {
	return e.occurredAt
}

func (e *PaymentConfigCreatedEvent) EventType() string {
	return "PaymentConfigCreated"
}

func (e *PaymentConfigCreatedEvent) Gateway() valueobject.PaymentGateway {
	return e.gateway
}

func (e *PaymentConfigCreatedEvent) Name() string {
	return e.name
}

func (e *PaymentConfigCreatedEvent) SupportedCurrencies() []valueobject.Currency {
	return e.supportedCurrencies
}

func (e *PaymentConfigCreatedEvent) AggregateID() string {
	return e.gateway.String()
}

func (e *PaymentConfigCreatedEvent) EventData() interface{} {
	return e
}

// PaymentConfigUpdatedEvent represents the event when a payment config is updated
type PaymentConfigUpdatedEvent struct {
	eventID    string
	occurredAt time.Time
	gateway    valueobject.PaymentGateway
	name       string
	isEnabled  bool
}

// NewPaymentConfigUpdatedEvent creates a new PaymentConfigUpdatedEvent
func NewPaymentConfigUpdatedEvent(
	gateway valueobject.PaymentGateway,
	name string,
	isEnabled bool,
	occurredAt time.Time,
) *PaymentConfigUpdatedEvent {
	return &PaymentConfigUpdatedEvent{
		eventID:    domain.NewEventID(),
		occurredAt: occurredAt,
		gateway:    gateway,
		name:       name,
		isEnabled:  isEnabled,
	}
}

func (e *PaymentConfigUpdatedEvent) EventID() string {
	return e.eventID
}

func (e *PaymentConfigUpdatedEvent) OccurredAt() time.Time {
	return e.occurredAt
}

func (e *PaymentConfigUpdatedEvent) EventType() string {
	return "PaymentConfigUpdated"
}

func (e *PaymentConfigUpdatedEvent) Gateway() valueobject.PaymentGateway {
	return e.gateway
}

func (e *PaymentConfigUpdatedEvent) Name() string {
	return e.name
}

func (e *PaymentConfigUpdatedEvent) IsEnabled() bool {
	return e.isEnabled
}

func (e *PaymentConfigUpdatedEvent) AggregateID() string {
	return e.gateway.String()
}

func (e *PaymentConfigUpdatedEvent) EventData() interface{} {
	return e
}

// PaymentConfigEnabledEvent represents the event when a payment config is enabled
type PaymentConfigEnabledEvent struct {
	eventID    string
	occurredAt time.Time
	gateway    valueobject.PaymentGateway
	name       string
}

// NewPaymentConfigEnabledEvent creates a new PaymentConfigEnabledEvent
func NewPaymentConfigEnabledEvent(
	gateway valueobject.PaymentGateway,
	name string,
	occurredAt time.Time,
) *PaymentConfigEnabledEvent {
	return &PaymentConfigEnabledEvent{
		eventID:    domain.NewEventID(),
		occurredAt: occurredAt,
		gateway:    gateway,
		name:       name,
	}
}

func (e *PaymentConfigEnabledEvent) EventID() string {
	return e.eventID
}

func (e *PaymentConfigEnabledEvent) OccurredAt() time.Time {
	return e.occurredAt
}

func (e *PaymentConfigEnabledEvent) EventType() string {
	return "PaymentConfigEnabled"
}

func (e *PaymentConfigEnabledEvent) Gateway() valueobject.PaymentGateway {
	return e.gateway
}

func (e *PaymentConfigEnabledEvent) Name() string {
	return e.name
}

func (e *PaymentConfigEnabledEvent) AggregateID() string {
	return e.gateway.String()
}

func (e *PaymentConfigEnabledEvent) EventData() interface{} {
	return e
}

// PaymentConfigDisabledEvent represents the event when a payment config is disabled
type PaymentConfigDisabledEvent struct {
	eventID    string
	occurredAt time.Time
	gateway    valueobject.PaymentGateway
	name       string
}

// NewPaymentConfigDisabledEvent creates a new PaymentConfigDisabledEvent
func NewPaymentConfigDisabledEvent(
	gateway valueobject.PaymentGateway,
	name string,
	occurredAt time.Time,
) *PaymentConfigDisabledEvent {
	return &PaymentConfigDisabledEvent{
		eventID:    domain.NewEventID(),
		occurredAt: occurredAt,
		gateway:    gateway,
		name:       name,
	}
}

func (e *PaymentConfigDisabledEvent) EventID() string {
	return e.eventID
}

func (e *PaymentConfigDisabledEvent) OccurredAt() time.Time {
	return e.occurredAt
}

func (e *PaymentConfigDisabledEvent) EventType() string {
	return "PaymentConfigDisabled"
}

func (e *PaymentConfigDisabledEvent) Gateway() valueobject.PaymentGateway {
	return e.gateway
}

func (e *PaymentConfigDisabledEvent) Name() string {
	return e.name
}

func (e *PaymentConfigDisabledEvent) AggregateID() string {
	return e.gateway.String()
}

func (e *PaymentConfigDisabledEvent) EventData() interface{} {
	return e
}

// PaymentConfigDeletedEvent represents the event when a payment config is deleted
type PaymentConfigDeletedEvent struct {
	eventID    string
	occurredAt time.Time
	gateway    valueobject.PaymentGateway
	name       string
}

// NewPaymentConfigDeletedEvent creates a new PaymentConfigDeletedEvent
func NewPaymentConfigDeletedEvent(
	gateway valueobject.PaymentGateway,
	name string,
	occurredAt time.Time,
) *PaymentConfigDeletedEvent {
	return &PaymentConfigDeletedEvent{
		eventID:    domain.NewEventID(),
		occurredAt: occurredAt,
		gateway:    gateway,
		name:       name,
	}
}

func (e *PaymentConfigDeletedEvent) EventID() string {
	return e.eventID
}

func (e *PaymentConfigDeletedEvent) OccurredAt() time.Time {
	return e.occurredAt
}

func (e *PaymentConfigDeletedEvent) EventType() string {
	return "PaymentConfigDeleted"
}

func (e *PaymentConfigDeletedEvent) Gateway() valueobject.PaymentGateway {
	return e.gateway
}

func (e *PaymentConfigDeletedEvent) Name() string {
	return e.name
}

func (e *PaymentConfigDeletedEvent) AggregateID() string {
	return e.gateway.String()
}

func (e *PaymentConfigDeletedEvent) EventData() interface{} {
	return e
}