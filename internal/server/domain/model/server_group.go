package model

import (
	"time"

	"linke/internal/server/domain/event"
	"linke/internal/server/domain/valueobject"
)

// ServerGroup represents a server group aggregate root
type ServerGroup struct {
	id        valueobject.ServerGroupID
	name      valueobject.ServerGroupName
	createdAt time.Time
	updatedAt time.Time
	events    []event.DomainEvent
}

// NewServerGroup creates a new server group
func NewServerGroup(name valueobject.ServerGroupName) (*ServerGroup, error) {
	group := &ServerGroup{
		name:      name,
		createdAt: time.Now(),
		updatedAt: time.Now(),
		events:    make([]event.DomainEvent, 0),
	}
	
	return group, nil
}

// ReconstructServerGroup reconstructs a server group from persistence
func ReconstructServerGroup(
	id valueobject.ServerGroupID,
	name valueobject.ServerGroupName,
	createdAt, updatedAt time.Time,
) *ServerGroup {
	return &ServerGroup{
		id:        id,
		name:      name,
		createdAt: createdAt,
		updatedAt: updatedAt,
		events:    make([]event.DomainEvent, 0),
	}
}

// ID returns the server group ID
func (sg *ServerGroup) ID() valueobject.ServerGroupID {
	return sg.id
}

// Name returns the server group name
func (sg *ServerGroup) Name() valueobject.ServerGroupName {
	return sg.name
}

// CreatedAt returns when the server group was created
func (sg *ServerGroup) CreatedAt() time.Time {
	return sg.createdAt
}

// UpdatedAt returns when the server group was last updated
func (sg *ServerGroup) UpdatedAt() time.Time {
	return sg.updatedAt
}

// ChangeGroupName changes the server group name
func (sg *ServerGroup) ChangeGroupName(newName valueobject.ServerGroupName) error {
	if sg.name.Equals(newName) {
		return nil // No change required
	}
	
	sg.name = newName
	sg.updatedAt = time.Now()
	
	// Only raise event if ID is set (persisted aggregate)
	if !sg.id.IsZero() {
		sg.raiseEvent(event.NewServerGroupUpdated(sg.id, sg.name))
	}
	
	return nil
}

// MarkAsCreated marks the server group as created (called after persistence)
func (sg *ServerGroup) MarkAsCreated(id valueobject.ServerGroupID) {
	sg.id = id
	sg.raiseEvent(event.NewServerGroupCreated(sg.id, sg.name))
}

// MarkAsDeleted marks the server group as deleted
func (sg *ServerGroup) MarkAsDeleted() {
	if !sg.id.IsZero() {
		sg.raiseEvent(event.NewServerGroupDeleted(sg.id))
	}
}

// DomainEvents returns all domain events
func (sg *ServerGroup) DomainEvents() []event.DomainEvent {
	return sg.events
}

// ClearEvents clears all domain events
func (sg *ServerGroup) ClearEvents() {
	sg.events = make([]event.DomainEvent, 0)
}

// raiseEvent adds a domain event
func (sg *ServerGroup) raiseEvent(evt event.DomainEvent) {
	sg.events = append(sg.events, evt)
}