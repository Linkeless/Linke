package model_test

import (
	"testing"
	"time"

	"linke/internal/server/domain/event"
	"linke/internal/server/domain/model"
	"linke/internal/server/domain/valueobject"
)

func TestNewServerGroup(t *testing.T) {
	name, err := valueobject.NewServerGroupName("Asia Pacific")
	if err != nil {
		t.Fatalf("Failed to create server group name: %v", err)
	}

	group, err := model.NewServerGroup(name)
	if err != nil {
		t.Fatalf("Failed to create server group: %v", err)
	}

	if group == nil {
		t.Fatal("Expected server group to be created")
	}

	if !group.Name().Equals(name) {
		t.Errorf("Expected name %v, got %v", name.Value(), group.Name().Value())
	}

	if group.ID().IsZero() == false {
		t.Error("Expected ID to be zero for new server group")
	}

	if group.CreatedAt().IsZero() {
		t.Error("Expected CreatedAt to be set")
	}

	if group.UpdatedAt().IsZero() {
		t.Error("Expected UpdatedAt to be set")
	}

	if len(group.DomainEvents()) != 0 {
		t.Error("Expected no domain events for new server group")
	}
}

func TestServerGroup_ChangeGroupName(t *testing.T) {
	// Create initial server group
	originalName, _ := valueobject.NewServerGroupName("Asia Pacific")
	group, _ := model.NewServerGroup(originalName)

	// Simulate persistence by setting ID
	id, _ := valueobject.NewServerGroupID(1)
	group.MarkAsCreated(id)
	group.ClearEvents() // Clear creation event

	// Change name
	newName, _ := valueobject.NewServerGroupName("Europe")
	err := group.ChangeGroupName(newName)

	if err != nil {
		t.Fatalf("Failed to change group name: %v", err)
	}

	if !group.Name().Equals(newName) {
		t.Errorf("Expected name %v, got %v", newName.Value(), group.Name().Value())
	}

	// Check domain events
	events := group.DomainEvents()
	if len(events) != 1 {
		t.Errorf("Expected 1 domain event, got %d", len(events))
	}

	if events[0].EventType() != "ServerGroupUpdated" {
		t.Errorf("Expected ServerGroupUpdated event, got %s", events[0].EventType())
	}
}

func TestServerGroup_ChangeGroupName_SameName(t *testing.T) {
	// Create initial server group
	name, _ := valueobject.NewServerGroupName("Asia Pacific")
	group, _ := model.NewServerGroup(name)

	// Simulate persistence by setting ID
	id, _ := valueobject.NewServerGroupID(1)
	group.MarkAsCreated(id)
	group.ClearEvents() // Clear creation event

	// Try to change to same name
	err := group.ChangeGroupName(name)

	if err != nil {
		t.Fatalf("Failed to change group name: %v", err)
	}

	// Should not generate domain events for no-op changes
	events := group.DomainEvents()
	if len(events) != 0 {
		t.Errorf("Expected no domain events for no-op change, got %d", len(events))
	}
}

func TestServerGroup_MarkAsCreated(t *testing.T) {
	name, _ := valueobject.NewServerGroupName("Asia Pacific")
	group, _ := model.NewServerGroup(name)

	id, _ := valueobject.NewServerGroupID(1)
	group.MarkAsCreated(id)

	if !group.ID().Equals(id) {
		t.Errorf("Expected ID %v, got %v", id.Value(), group.ID().Value())
	}

	// Check domain events
	events := group.DomainEvents()
	if len(events) != 1 {
		t.Errorf("Expected 1 domain event, got %d", len(events))
	}

	if events[0].EventType() != "ServerGroupCreated" {
		t.Errorf("Expected ServerGroupCreated event, got %s", events[0].EventType())
	}

	if createdEvent, ok := events[0].(*event.ServerGroupCreated); ok {
		if !createdEvent.GroupID.Equals(id) {
			t.Errorf("Expected event GroupID %v, got %v", id.Value(), createdEvent.GroupID.Value())
		}
		if !createdEvent.Name.Equals(name) {
			t.Errorf("Expected event Name %v, got %v", name.Value(), createdEvent.Name.Value())
		}
	} else {
		t.Error("Expected ServerGroupCreated event")
	}
}

func TestServerGroup_MarkAsDeleted(t *testing.T) {
	name, _ := valueobject.NewServerGroupName("Asia Pacific")
	group, _ := model.NewServerGroup(name)

	// First mark as created
	id, _ := valueobject.NewServerGroupID(1)
	group.MarkAsCreated(id)
	group.ClearEvents() // Clear creation event

	// Then mark as deleted
	group.MarkAsDeleted()

	// Check domain events
	events := group.DomainEvents()
	if len(events) != 1 {
		t.Errorf("Expected 1 domain event, got %d", len(events))
	}

	if events[0].EventType() != "ServerGroupDeleted" {
		t.Errorf("Expected ServerGroupDeleted event, got %s", events[0].EventType())
	}

	if deletedEvent, ok := events[0].(*event.ServerGroupDeleted); ok {
		if !deletedEvent.GroupID.Equals(id) {
			t.Errorf("Expected event GroupID %v, got %v", id.Value(), deletedEvent.GroupID.Value())
		}
	} else {
		t.Error("Expected ServerGroupDeleted event")
	}
}

func TestServerGroup_ClearEvents(t *testing.T) {
	name, _ := valueobject.NewServerGroupName("Asia Pacific")
	group, _ := model.NewServerGroup(name)

	id, _ := valueobject.NewServerGroupID(1)
	group.MarkAsCreated(id)

	// Verify events exist
	if len(group.DomainEvents()) != 1 {
		t.Error("Expected domain events to exist")
	}

	// Clear events
	group.ClearEvents()

	// Verify events are cleared
	if len(group.DomainEvents()) != 0 {
		t.Error("Expected domain events to be cleared")
	}
}

func TestReconstructServerGroup(t *testing.T) {
	id, _ := valueobject.NewServerGroupID(1)
	name, _ := valueobject.NewServerGroupName("Asia Pacific")
	
	// Use a fixed time for testing
	now := time.Now()
	
	group := model.ReconstructServerGroup(id, name, now, now)

	if !group.ID().Equals(id) {
		t.Errorf("Expected ID %v, got %v", id.Value(), group.ID().Value())
	}

	if !group.Name().Equals(name) {
		t.Errorf("Expected name %v, got %v", name.Value(), group.Name().Value())
	}

	if group.CreatedAt().Unix() != now.Unix() {
		t.Error("Expected CreatedAt to match provided time")
	}

	if group.UpdatedAt().Unix() != now.Unix() {
		t.Error("Expected UpdatedAt to match provided time")
	}

	if len(group.DomainEvents()) != 0 {
		t.Error("Expected no domain events for reconstructed server group")
	}
}