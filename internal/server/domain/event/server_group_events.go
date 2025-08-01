package event

import (
	"time"

	"linke/internal/server/domain/valueobject"
)

// DomainEvent represents a domain event
type DomainEvent interface {
	OccurredOn() time.Time
	EventType() string
	AggregateID() string
}

// BaseDomainEvent provides common event functionality
type BaseDomainEvent struct {
	occurredOn time.Time
	eventType  string
}

// OccurredOn returns when the event occurred
func (e BaseDomainEvent) OccurredOn() time.Time {
	return e.occurredOn
}

// EventType returns the type of the event
func (e BaseDomainEvent) EventType() string {
	return e.eventType
}

// ServerGroupCreated is emitted when a server group is created
type ServerGroupCreated struct {
	BaseDomainEvent
	GroupID valueobject.ServerGroupID
	Name    valueobject.ServerGroupName
}

// NewServerGroupCreated creates a new ServerGroupCreated event
func NewServerGroupCreated(groupID valueobject.ServerGroupID, name valueobject.ServerGroupName) *ServerGroupCreated {
	return &ServerGroupCreated{
		BaseDomainEvent: BaseDomainEvent{
			occurredOn: time.Now(),
			eventType:  "ServerGroupCreated",
		},
		GroupID: groupID,
		Name:    name,
	}
}

// AggregateID returns the aggregate ID
func (e ServerGroupCreated) AggregateID() string {
	return e.GroupID.String()
}

// ServerGroupUpdated is emitted when a server group is updated
type ServerGroupUpdated struct {
	BaseDomainEvent
	GroupID valueobject.ServerGroupID
	Name    valueobject.ServerGroupName
}

// NewServerGroupUpdated creates a new ServerGroupUpdated event
func NewServerGroupUpdated(groupID valueobject.ServerGroupID, name valueobject.ServerGroupName) *ServerGroupUpdated {
	return &ServerGroupUpdated{
		BaseDomainEvent: BaseDomainEvent{
			occurredOn: time.Now(),
			eventType:  "ServerGroupUpdated",
		},
		GroupID: groupID,
		Name:    name,
	}
}

// AggregateID returns the aggregate ID
func (e ServerGroupUpdated) AggregateID() string {
	return e.GroupID.String()
}

// ServerGroupDeleted is emitted when a server group is deleted
type ServerGroupDeleted struct {
	BaseDomainEvent
	GroupID valueobject.ServerGroupID
}

// NewServerGroupDeleted creates a new ServerGroupDeleted event
func NewServerGroupDeleted(groupID valueobject.ServerGroupID) *ServerGroupDeleted {
	return &ServerGroupDeleted{
		BaseDomainEvent: BaseDomainEvent{
			occurredOn: time.Now(),
			eventType:  "ServerGroupDeleted",
		},
		GroupID: groupID,
	}
}

// AggregateID returns the aggregate ID
func (e ServerGroupDeleted) AggregateID() string {
	return e.GroupID.String()
}