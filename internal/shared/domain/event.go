package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// EventPublisher defines the interface for publishing domain events
type EventPublisher interface {
	Publish(ctx context.Context, event DomainEvent) error
	PublishBatch(ctx context.Context, events []DomainEvent) error
}

// EventBus defines the interface for event bus operations
type EventBus interface {
	Publish(ctx context.Context, events ...DomainEvent) error
}

// DomainEvent represents a domain event
type DomainEvent interface {
	EventID() string
	EventType() string
	OccurredAt() time.Time
	AggregateID() string
	EventData() interface{}
}

// NewEventID generates a new unique event ID
func NewEventID() string {
	return uuid.New().String()
}

// BusinessError represents a domain business rule violation
type BusinessError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Error implements the error interface
func (e BusinessError) Error() string {
	return e.Message
}

// NewBusinessError creates a new BusinessError
func NewBusinessError(code, message string) BusinessError {
	return BusinessError{
		Code:    code,
		Message: message,
	}
}