package command

import (
	"context"
	"errors"

	"linke/internal/server/domain/model"
	"linke/internal/server/domain/repository"
	"linke/internal/server/domain/service"
	"linke/internal/server/domain/valueobject"
)

var (
	ErrServerGroupAlreadyExists     = errors.New("server group with this name already exists")
	ErrServerGroupNotFound          = errors.New("server group not found")
	ErrCannotDeleteServerGroupWithServers = errors.New("cannot delete server group that contains servers")
)

// ServerGroupCommandHandler handles server group commands
type ServerGroupCommandHandler struct {
	serverGroupRepo    repository.ServerGroupRepository
	domainService      *service.ServerGroupDomainService
	eventPublisher     EventPublisher
}

// EventPublisher defines the interface for publishing domain events
type EventPublisher interface {
	PublishEvents(ctx context.Context, events []interface{}) error
}

// NewServerGroupCommandHandler creates a new server group command handler
func NewServerGroupCommandHandler(
	serverGroupRepo repository.ServerGroupRepository,
	domainService *service.ServerGroupDomainService,
	eventPublisher EventPublisher,
) *ServerGroupCommandHandler {
	return &ServerGroupCommandHandler{
		serverGroupRepo: serverGroupRepo,
		domainService:   domainService,
		eventPublisher:  eventPublisher,
	}
}

// HandleCreateServerGroup handles the create server group command
func (h *ServerGroupCommandHandler) HandleCreateServerGroup(ctx context.Context, cmd CreateServerGroupCommand) (*model.ServerGroup, error) {
	// Create value objects
	name, err := valueobject.NewServerGroupName(cmd.Name)
	if err != nil {
		return nil, err
	}
	
	// Check if name is unique
	isUnique, err := h.domainService.IsServerGroupNameUnique(ctx, name)
	if err != nil {
		return nil, err
	}
	
	if !isUnique {
		return nil, ErrServerGroupAlreadyExists
	}
	
	// Create the server group
	group, err := model.NewServerGroup(name)
	if err != nil {
		return nil, err
	}
	
	// Save to repository
	if err := h.serverGroupRepo.Save(ctx, group); err != nil {
		return nil, err
	}
	
	// Publish domain events
	events := group.DomainEvents()
	if len(events) > 0 {
		eventInterfaces := make([]interface{}, len(events))
		for i, evt := range events {
			eventInterfaces[i] = evt
		}
		if err := h.eventPublisher.PublishEvents(ctx, eventInterfaces); err != nil {
			// Log error but don't fail the transaction
			// In a production system, you might want to retry or use outbox pattern
		}
		group.ClearEvents()
	}
	
	return group, nil
}

// HandleUpdateServerGroup handles the update server group command
func (h *ServerGroupCommandHandler) HandleUpdateServerGroup(ctx context.Context, cmd UpdateServerGroupCommand) (*model.ServerGroup, error) {
	// Create value objects
	id, err := valueobject.NewServerGroupID(cmd.ID)
	if err != nil {
		return nil, err
	}
	
	// Find existing server group
	group, err := h.serverGroupRepo.FindByID(ctx, id)
	if err != nil {
		return nil, ErrServerGroupNotFound
	}
	
	// Update name if provided
	if cmd.Name != nil {
		name, err := valueobject.NewServerGroupName(*cmd.Name)
		if err != nil {
			return nil, err
		}
		
		// Check if name is unique for update
		isUnique, err := h.domainService.IsServerGroupNameUniqueForUpdate(ctx, name, id)
		if err != nil {
			return nil, err
		}
		
		if !isUnique {
			return nil, ErrServerGroupAlreadyExists
		}
		
		// Change the name
		if err := group.ChangeGroupName(name); err != nil {
			return nil, err
		}
	}
	
	// Save changes
	if err := h.serverGroupRepo.Save(ctx, group); err != nil {
		return nil, err
	}
	
	// Publish domain events
	events := group.DomainEvents()
	if len(events) > 0 {
		eventInterfaces := make([]interface{}, len(events))
		for i, evt := range events {
			eventInterfaces[i] = evt
		}
		if err := h.eventPublisher.PublishEvents(ctx, eventInterfaces); err != nil {
			// Log error but don't fail the transaction
		}
		group.ClearEvents()
	}
	
	return group, nil
}

// HandleDeleteServerGroup handles the delete server group command
func (h *ServerGroupCommandHandler) HandleDeleteServerGroup(ctx context.Context, cmd DeleteServerGroupCommand) error {
	// Create value objects
	id, err := valueobject.NewServerGroupID(cmd.ID)
	if err != nil {
		return err
	}
	
	// Check if server group can be deleted
	if err := h.domainService.CanDeleteServerGroup(ctx, id); err != nil {
		if errors.Is(err, service.ErrServerGroupHasServers) {
			return ErrCannotDeleteServerGroupWithServers
		}
		return err
	}
	
	// Find existing server group to get domain events
	group, err := h.serverGroupRepo.FindByID(ctx, id)
	if err != nil {
		return ErrServerGroupNotFound
	}
	
	// Mark as deleted
	group.MarkAsDeleted()
	
	// Delete from repository
	if err := h.serverGroupRepo.Delete(ctx, id); err != nil {
		return err
	}
	
	// Publish domain events
	events := group.DomainEvents()
	if len(events) > 0 {
		eventInterfaces := make([]interface{}, len(events))
		for i, evt := range events {
			eventInterfaces[i] = evt
		}
		if err := h.eventPublisher.PublishEvents(ctx, eventInterfaces); err != nil {
			// Log error but don't fail the transaction
		}
	}
	
	return nil
}