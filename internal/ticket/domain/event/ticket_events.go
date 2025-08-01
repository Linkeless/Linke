package event

import (
	"fmt"
	"time"

	"linke/internal/ticket/domain/valueobject"
	sharedvo "linke/internal/shared/valueobject"
)

// DomainEvent represents a domain event
type DomainEvent interface {
	EventID() string
	AggregateID() string
	EventType() string
	OccurredOn() time.Time
	Version() int
}

// BaseDomainEvent provides common event functionality
type BaseDomainEvent struct {
	eventID     string
	aggregateID string
	eventType   string
	occurredOn  time.Time
	version     int
}

// NewBaseDomainEvent creates a new base domain event
func NewBaseDomainEvent(aggregateID, eventType string, version int) BaseDomainEvent {
	return BaseDomainEvent{
		eventID:     generateEventID(),
		aggregateID: aggregateID,
		eventType:   eventType,
		occurredOn:  time.Now(),
		version:     version,
	}
}

// EventID returns the event ID
func (e BaseDomainEvent) EventID() string {
	return e.eventID
}

// AggregateID returns the aggregate ID
func (e BaseDomainEvent) AggregateID() string {
	return e.aggregateID
}

// EventType returns the event type
func (e BaseDomainEvent) EventType() string {
	return e.eventType
}

// OccurredOn returns when the event occurred
func (e BaseDomainEvent) OccurredOn() time.Time {
	return e.occurredOn
}

// Version returns the event version
func (e BaseDomainEvent) Version() int {
	return e.version
}

// generateEventID generates a unique event ID
func generateEventID() string {
	// In a real implementation, you might use UUID
	return fmt.Sprintf("evt_%d", time.Now().UnixNano())
}

// TicketCreated represents a ticket creation event
type TicketCreated struct {
	BaseDomainEvent
	TicketID    valueobject.TicketID
	TicketNumber valueobject.TicketNumber
	UserID      sharedvo.UserID
	Title       string
	Description string
	Category    valueobject.TicketCategory
	Priority    valueobject.TicketPriority
}

// NewTicketCreated creates a new TicketCreated event
func NewTicketCreated(
	ticketID valueobject.TicketID,
	ticketNumber valueobject.TicketNumber,
	userID sharedvo.UserID,
	title, description string,
	category valueobject.TicketCategory,
	priority valueobject.TicketPriority,
) TicketCreated {
	return TicketCreated{
		BaseDomainEvent: NewBaseDomainEvent(ticketID.String(), "TicketCreated", 1),
		TicketID:        ticketID,
		TicketNumber:    ticketNumber,
		UserID:          userID,
		Title:           title,
		Description:     description,
		Category:        category,
		Priority:        priority,
	}
}

// TicketAssigned represents a ticket assignment event
type TicketAssigned struct {
	BaseDomainEvent
	TicketID     valueobject.TicketID
	AssignedToID sharedvo.UserID
	AssignedAt   time.Time
}

// NewTicketAssigned creates a new TicketAssigned event
func NewTicketAssigned(ticketID valueobject.TicketID, assignedToID sharedvo.UserID, assignedAt time.Time) TicketAssigned {
	return TicketAssigned{
		BaseDomainEvent: NewBaseDomainEvent(ticketID.String(), "TicketAssigned", 1),
		TicketID:        ticketID,
		AssignedToID:    assignedToID,
		AssignedAt:      assignedAt,
	}
}

// TicketStatusChanged represents a ticket status change event
type TicketStatusChanged struct {
	BaseDomainEvent
	TicketID  valueobject.TicketID
	OldStatus valueobject.TicketStatus
	NewStatus valueobject.TicketStatus
	ChangedBy sharedvo.UserID
}

// NewTicketStatusChanged creates a new TicketStatusChanged event
func NewTicketStatusChanged(
	ticketID valueobject.TicketID,
	oldStatus, newStatus valueobject.TicketStatus,
	changedBy sharedvo.UserID,
) TicketStatusChanged {
	return TicketStatusChanged{
		BaseDomainEvent: NewBaseDomainEvent(ticketID.String(), "TicketStatusChanged", 1),
		TicketID:        ticketID,
		OldStatus:       oldStatus,
		NewStatus:       newStatus,
		ChangedBy:       changedBy,
	}
}

// TicketResolved represents a ticket resolution event
type TicketResolved struct {
	BaseDomainEvent
	TicketID     valueobject.TicketID
	ResolvedBy   sharedvo.UserID
	ResolvedAt   time.Time
	Resolution   string
}

// NewTicketResolved creates a new TicketResolved event
func NewTicketResolved(
	ticketID valueobject.TicketID,
	resolvedBy sharedvo.UserID,
	resolvedAt time.Time,
	resolution string,
) TicketResolved {
	return TicketResolved{
		BaseDomainEvent: NewBaseDomainEvent(ticketID.String(), "TicketResolved", 1),
		TicketID:        ticketID,
		ResolvedBy:      resolvedBy,
		ResolvedAt:      resolvedAt,
		Resolution:      resolution,
	}
}

// TicketClosed represents a ticket closure event
type TicketClosed struct {
	BaseDomainEvent
	TicketID valueobject.TicketID
	ClosedAt time.Time
}

// NewTicketClosed creates a new TicketClosed event
func NewTicketClosed(ticketID valueobject.TicketID, closedAt time.Time) TicketClosed {
	return TicketClosed{
		BaseDomainEvent: NewBaseDomainEvent(ticketID.String(), "TicketClosed", 1),
		TicketID:        ticketID,
		ClosedAt:        closedAt,
	}
}

// TicketMessageAdded represents a ticket message addition event
type TicketMessageAdded struct {
	BaseDomainEvent
	TicketID    valueobject.TicketID
	MessageID   uint
	UserID      sharedvo.UserID
	Content     string
	MessageType string
	IsInternal  bool
}

// NewTicketMessageAdded creates a new TicketMessageAdded event
func NewTicketMessageAdded(
	ticketID valueobject.TicketID,
	messageID uint,
	userID sharedvo.UserID,
	content, messageType string,
	isInternal bool,
) TicketMessageAdded {
	return TicketMessageAdded{
		BaseDomainEvent: NewBaseDomainEvent(ticketID.String(), "TicketMessageAdded", 1),
		TicketID:        ticketID,
		MessageID:       messageID,
		UserID:          userID,
		Content:         content,
		MessageType:     messageType,
		IsInternal:      isInternal,
	}
}