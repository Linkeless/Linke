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
	ErrShadowsocksServerNotFound = errors.New("shadowsocks server not found")
	ErrInvalidServerGroup        = errors.New("invalid server group")
)

// ShadowsocksServerCommandHandler handles shadowsocks server commands
type ShadowsocksServerCommandHandler struct {
	shadowsocksServerRepo repository.ShadowsocksServerRepository
	domainService         *service.ShadowsocksServerDomainService
	eventPublisher        EventPublisher
}

// NewShadowsocksServerCommandHandler creates a new shadowsocks server command handler
func NewShadowsocksServerCommandHandler(
	shadowsocksServerRepo repository.ShadowsocksServerRepository,
	domainService *service.ShadowsocksServerDomainService,
	eventPublisher EventPublisher,
) *ShadowsocksServerCommandHandler {
	return &ShadowsocksServerCommandHandler{
		shadowsocksServerRepo: shadowsocksServerRepo,
		domainService:         domainService,
		eventPublisher:        eventPublisher,
	}
}

// HandleCreateShadowsocksServer handles the create shadowsocks server command
func (h *ShadowsocksServerCommandHandler) HandleCreateShadowsocksServer(ctx context.Context, cmd CreateShadowsocksServerCommand) (*model.ShadowsocksServer, error) {
	// Create value objects
	groupID, err := valueobject.NewServerGroupID(cmd.GroupID)
	if err != nil {
		return nil, err
	}
	
	name, err := valueobject.NewServerName(cmd.Name)
	if err != nil {
		return nil, err
	}
	
	host, err := valueobject.NewServerHost(cmd.Host)
	if err != nil {
		return nil, err
	}
	
	port, err := valueobject.NewServerPort(cmd.Port)
	if err != nil {
		return nil, err
	}
	
	serverPort, err := valueobject.NewServerPort(cmd.ServerPort)
	if err != nil {
		return nil, err
	}
	
	cipher, err := valueobject.NewCipher(cmd.Cipher)
	if err != nil {
		return nil, err
	}
	
	rate, err := valueobject.NewRate(cmd.Rate)
	if err != nil {
		return nil, err
	}
	
	// Validate server group exists
	if err := h.domainService.ValidateServerGroup(ctx, groupID); err != nil {
		return nil, ErrInvalidServerGroup
	}
	
	// Create the shadowsocks server
	server, err := model.NewShadowsocksServer(groupID, name, host, port, serverPort, cipher, rate)
	if err != nil {
		return nil, err
	}
	
	// Update metadata
	server.UpdateMetadata(cmd.Tags, cmd.Obfs, cmd.ObfsSettings, cmd.Excludes, cmd.IPs, cmd.RouteID, cmd.ParentID, cmd.Sort)
	
	// Set visibility
	server.ChangeVisibility(cmd.Show == 1)
	
	// Save to repository
	if err := h.shadowsocksServerRepo.Save(ctx, server); err != nil {
		return nil, err
	}
	
	// Publish domain events
	events := server.DomainEvents()
	if len(events) > 0 {
		eventInterfaces := make([]interface{}, len(events))
		for i, evt := range events {
			eventInterfaces[i] = evt
		}
		if err := h.eventPublisher.PublishEvents(ctx, eventInterfaces); err != nil {
			// Log error but don't fail the transaction
		}
		server.ClearEvents()
	}
	
	return server, nil
}

// HandleUpdateShadowsocksServer handles the update shadowsocks server command
func (h *ShadowsocksServerCommandHandler) HandleUpdateShadowsocksServer(ctx context.Context, cmd UpdateShadowsocksServerCommand) (*model.ShadowsocksServer, error) {
	// Create value objects
	id, err := valueobject.NewServerID(cmd.ID)
	if err != nil {
		return nil, err
	}
	
	// Find existing server
	server, err := h.shadowsocksServerRepo.FindByID(ctx, id)
	if err != nil {
		return nil, ErrShadowsocksServerNotFound
	}
	
	// Update configuration if provided
	if cmd.Name != nil || cmd.Host != nil || cmd.Port != nil || cmd.ServerPort != nil || cmd.Cipher != nil || cmd.Rate != nil {
		name := server.Name()
		if cmd.Name != nil {
			name, err = valueobject.NewServerName(*cmd.Name)
			if err != nil {
				return nil, err
			}
		}
		
		host := server.Host()
		if cmd.Host != nil {
			host, err = valueobject.NewServerHost(*cmd.Host)
			if err != nil {
				return nil, err
			}
		}
		
		port := server.Port()
		if cmd.Port != nil {
			port, err = valueobject.NewServerPort(*cmd.Port)
			if err != nil {
				return nil, err
			}
		}
		
		serverPort := server.ServerPort()
		if cmd.ServerPort != nil {
			serverPort, err = valueobject.NewServerPort(*cmd.ServerPort)
			if err != nil {
				return nil, err
			}
		}
		
		cipher := server.Cipher()
		if cmd.Cipher != nil {
			cipher, err = valueobject.NewCipher(*cmd.Cipher)
			if err != nil {
				return nil, err
			}
		}
		
		rate := server.Rate()
		if cmd.Rate != nil {
			rate, err = valueobject.NewRate(*cmd.Rate)
			if err != nil {
				return nil, err
			}
		}
		
		if err := server.UpdateConfiguration(name, host, port, serverPort, cipher, rate); err != nil {
			return nil, err
		}
	}
	
	// Update group if provided
	if cmd.GroupID != nil {
		groupID, err := valueobject.NewServerGroupID(*cmd.GroupID)
		if err != nil {
			return nil, err
		}
		
		// Validate new group exists
		if err := h.domainService.ValidateServerGroup(ctx, groupID); err != nil {
			return nil, ErrInvalidServerGroup
		}
		
		server.MoveToGroup(groupID)
	}
	
	// Update visibility if provided
	if cmd.Show != nil {
		server.ChangeVisibility(*cmd.Show == 1)
	}
	
	// Update metadata
	tags := server.Tags()
	if cmd.Tags != nil {
		tags = *cmd.Tags
	}
	
	obfs := server.Obfs()
	if cmd.Obfs != nil {
		obfs = *cmd.Obfs
	}
	
	obfsSettings := server.ObfsSettings()
	if cmd.ObfsSettings != nil {
		obfsSettings = *cmd.ObfsSettings
	}
	
	excludes := server.Excludes()
	if cmd.Excludes != nil {
		excludes = *cmd.Excludes
	}
	
	ips := server.IPs()
	if cmd.IPs != nil {
		ips = *cmd.IPs
	}
	
	routeID := server.RouteID()
	if cmd.RouteID != nil {
		routeID = *cmd.RouteID
	}
	
	parentID := server.ParentID()
	if cmd.ParentID != nil {
		parentID = cmd.ParentID
	}
	
	sort := server.Sort()
	if cmd.Sort != nil {
		sort = cmd.Sort
	}
	
	server.UpdateMetadata(tags, obfs, obfsSettings, excludes, ips, routeID, parentID, sort)
	
	// Save changes
	if err := h.shadowsocksServerRepo.Save(ctx, server); err != nil {
		return nil, err
	}
	
	// Publish domain events
	events := server.DomainEvents()
	if len(events) > 0 {
		eventInterfaces := make([]interface{}, len(events))
		for i, evt := range events {
			eventInterfaces[i] = evt
		}
		if err := h.eventPublisher.PublishEvents(ctx, eventInterfaces); err != nil {
			// Log error but don't fail the transaction
		}
		server.ClearEvents()
	}
	
	return server, nil
}

// HandleDeleteShadowsocksServer handles the delete shadowsocks server command
func (h *ShadowsocksServerCommandHandler) HandleDeleteShadowsocksServer(ctx context.Context, cmd DeleteShadowsocksServerCommand) error {
	// Create value objects
	id, err := valueobject.NewServerID(cmd.ID)
	if err != nil {
		return err
	}
	
	// Find existing server to get domain events
	server, err := h.shadowsocksServerRepo.FindByID(ctx, id)
	if err != nil {
		return ErrShadowsocksServerNotFound
	}
	
	// Mark as deleted
	server.MarkAsDeleted()
	
	// Delete from repository
	if err := h.shadowsocksServerRepo.Delete(ctx, id); err != nil {
		return err
	}
	
	// Publish domain events
	events := server.DomainEvents()
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

// HandleChangeShadowsocksServerVisibility handles the change server visibility command
func (h *ShadowsocksServerCommandHandler) HandleChangeShadowsocksServerVisibility(ctx context.Context, cmd ChangeShadowsocksServerVisibilityCommand) error {
	// Create value objects
	id, err := valueobject.NewServerID(cmd.ID)
	if err != nil {
		return err
	}
	
	// Find existing server
	server, err := h.shadowsocksServerRepo.FindByID(ctx, id)
	if err != nil {
		return ErrShadowsocksServerNotFound
	}
	
	// Change visibility
	server.ChangeVisibility(cmd.IsVisible)
	
	// Save changes
	if err := h.shadowsocksServerRepo.Save(ctx, server); err != nil {
		return err
	}
	
	// Publish domain events
	events := server.DomainEvents()
	if len(events) > 0 {
		eventInterfaces := make([]interface{}, len(events))
		for i, evt := range events {
			eventInterfaces[i] = evt
		}
		if err := h.eventPublisher.PublishEvents(ctx, eventInterfaces); err != nil {
			// Log error but don't fail the transaction
		}
		server.ClearEvents()
	}
	
	return nil
}

// HandleMoveShadowsocksServerToGroup handles the move server to group command
func (h *ShadowsocksServerCommandHandler) HandleMoveShadowsocksServerToGroup(ctx context.Context, cmd MoveShadowsocksServerToGroupCommand) error {
	// Create value objects
	id, err := valueobject.NewServerID(cmd.ID)
	if err != nil {
		return err
	}
	
	groupID, err := valueobject.NewServerGroupID(cmd.GroupID)
	if err != nil {
		return err
	}
	
	// Find existing server
	server, err := h.shadowsocksServerRepo.FindByID(ctx, id)
	if err != nil {
		return ErrShadowsocksServerNotFound
	}
	
	// Validate new group exists
	if err := h.domainService.ValidateServerGroup(ctx, groupID); err != nil {
		return ErrInvalidServerGroup
	}
	
	// Move to group
	server.MoveToGroup(groupID)
	
	// Save changes
	if err := h.shadowsocksServerRepo.Save(ctx, server); err != nil {
		return err
	}
	
	// Publish domain events
	events := server.DomainEvents()
	if len(events) > 0 {
		eventInterfaces := make([]interface{}, len(events))
		for i, evt := range events {
			eventInterfaces[i] = evt
		}
		if err := h.eventPublisher.PublishEvents(ctx, eventInterfaces); err != nil {
			// Log error but don't fail the transaction
		}
		server.ClearEvents()
	}
	
	return nil
}