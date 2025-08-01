package command

import (
	"context"
	"fmt"

	"linke/internal/ticket/domain/model"
	"linke/internal/ticket/domain/repository"
	"linke/internal/ticket/domain/service"
	"linke/internal/ticket/domain/valueobject"
)

// TicketCommandHandler handles ticket-related commands
type TicketCommandHandler struct {
	ticketRepo        repository.TicketRepository
	ticketMessageRepo repository.TicketMessageRepository
	domainService     *service.TicketDomainService
	eventPublisher    EventPublisher
}

// EventPublisher interface for publishing domain events
type EventPublisher interface {
	Publish(ctx context.Context, events ...interface{}) error
}

// NewTicketCommandHandler creates a new ticket command handler
func NewTicketCommandHandler(
	ticketRepo repository.TicketRepository,
	ticketMessageRepo repository.TicketMessageRepository,
	domainService *service.TicketDomainService,
	eventPublisher EventPublisher,
) *TicketCommandHandler {
	return &TicketCommandHandler{
		ticketRepo:        ticketRepo,
		ticketMessageRepo: ticketMessageRepo,
		domainService:     domainService,
		eventPublisher:    eventPublisher,
	}
}

// CreateTicket handles the create ticket command
func (h *TicketCommandHandler) CreateTicket(ctx context.Context, cmd CreateTicketCommand) (*model.Ticket, error) {
	// Validate and create value objects
	category, err := valueobject.NewTicketCategory(cmd.Category)
	if err != nil {
		return nil, fmt.Errorf("invalid category: %w", err)
	}
	
	priority := valueobject.DefaultTicketPriority()
	if cmd.Priority != "" {
		priority, err = valueobject.NewTicketPriority(cmd.Priority)
		if err != nil {
			return nil, fmt.Errorf("invalid priority: %w", err)
		}
	}
	
	// Generate unique ticket number
	_, err = h.domainService.GenerateUniqueTicketNumber(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to generate ticket number: %w", err)
	}
	
	// Generate ticket ID (in a real system, this might come from the repository)
	ticketID := valueobject.NewTicketID(0) // Will be set by repository
	
	// Create the ticket aggregate
	ticket, err := model.NewTicket(
		ticketID,
		cmd.UserID,
		cmd.Title,
		cmd.Description,
		category,
		priority,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create ticket: %w", err)
	}
	
	// Set additional properties
	if len(cmd.Tags) > 0 {
		ticket.SetTags(cmd.Tags)
	}
	
	for key, value := range cmd.Metadata {
		ticket.SetMetadata(key, value)
	}
	
	// Save the ticket
	if err := h.ticketRepo.Save(ctx, ticket); err != nil {
		return nil, fmt.Errorf("failed to save ticket: %w", err)
	}
	
	// Publish domain events
	if err := h.publishDomainEvents(ctx, ticket); err != nil {
		// Log error but don't fail the operation
		// In production, you might want to use an outbox pattern
	}
	
	return ticket, nil
}

// UpdateTicket handles the update ticket command
func (h *TicketCommandHandler) UpdateTicket(ctx context.Context, cmd UpdateTicketCommand) (*model.Ticket, error) {
	// Load the ticket
	ticket, err := h.ticketRepo.FindByID(ctx, cmd.TicketID)
	if err != nil {
		return nil, fmt.Errorf("failed to load ticket: %w", err)
	}
	
	// Check access permissions
	if !cmd.IsAdmin && !ticket.UserID().Equals(cmd.UserID) {
		return nil, fmt.Errorf("access denied: can only update your own tickets")
	}
	
	// Update fields if provided
	if cmd.Title != nil {
		if err := ticket.UpdateTitle(*cmd.Title); err != nil {
			return nil, fmt.Errorf("failed to update title: %w", err)
		}
	}
	
	if cmd.Description != nil {
		if err := ticket.UpdateDescription(*cmd.Description); err != nil {
			return nil, fmt.Errorf("failed to update description: %w", err)
		}
	}
	
	if cmd.Category != nil {
		category, err := valueobject.NewTicketCategory(*cmd.Category)
		if err != nil {
			return nil, fmt.Errorf("invalid category: %w", err)
		}
		if err := ticket.UpdateCategory(category); err != nil {
			return nil, fmt.Errorf("failed to update category: %w", err)
		}
	}
	
	if cmd.Priority != nil {
		priority, err := valueobject.NewTicketPriority(*cmd.Priority)
		if err != nil {
			return nil, fmt.Errorf("invalid priority: %w", err)
		}
		if err := ticket.UpdatePriority(priority); err != nil {
			return nil, fmt.Errorf("failed to update priority: %w", err)
		}
	}
	
	if len(cmd.Tags) > 0 {
		ticket.SetTags(cmd.Tags)
	}
	
	// Save the ticket
	if err := h.ticketRepo.Save(ctx, ticket); err != nil {
		return nil, fmt.Errorf("failed to save ticket: %w", err)
	}
	
	// Publish domain events
	if err := h.publishDomainEvents(ctx, ticket); err != nil {
		// Log error but don't fail the operation
	}
	
	return ticket, nil
}

// AssignTicket handles the assign ticket command
func (h *TicketCommandHandler) AssignTicket(ctx context.Context, cmd AssignTicketCommand) (*model.Ticket, error) {
	// Load the ticket
	ticket, err := h.ticketRepo.FindByID(ctx, cmd.TicketID)
	if err != nil {
		return nil, fmt.Errorf("failed to load ticket: %w", err)
	}
	
	// Assign the ticket
	if err := ticket.AssignTo(cmd.AssignedToID); err != nil {
		return nil, fmt.Errorf("failed to assign ticket: %w", err)
	}
	
	// Save the ticket
	if err := h.ticketRepo.Save(ctx, ticket); err != nil {
		return nil, fmt.Errorf("failed to save ticket: %w", err)
	}
	
	// Publish domain events
	if err := h.publishDomainEvents(ctx, ticket); err != nil {
		// Log error but don't fail the operation
	}
	
	return ticket, nil
}

// ChangeTicketStatus handles the change ticket status command
func (h *TicketCommandHandler) ChangeTicketStatus(ctx context.Context, cmd ChangeTicketStatusCommand) (*model.Ticket, error) {
	// Load the ticket
	ticket, err := h.ticketRepo.FindByID(ctx, cmd.TicketID)
	if err != nil {
		return nil, fmt.Errorf("failed to load ticket: %w", err)
	}
	
	// Validate status
	newStatus, err := valueobject.NewTicketStatus(cmd.NewStatus)
	if err != nil {
		return nil, fmt.Errorf("invalid status: %w", err)
	}
	
	// Validate transition using domain service
	if err := h.domainService.ValidateTicketTransition(ticket, newStatus, cmd.ChangedBy, cmd.IsAdmin); err != nil {
		return nil, fmt.Errorf("invalid status transition: %w", err)
	}
	
	// Change the status
	if err := ticket.ChangeStatus(newStatus, cmd.ChangedBy); err != nil {
		return nil, fmt.Errorf("failed to change status: %w", err)
	}
	
	// Save the ticket
	if err := h.ticketRepo.Save(ctx, ticket); err != nil {
		return nil, fmt.Errorf("failed to save ticket: %w", err)
	}
	
	// Publish domain events
	if err := h.publishDomainEvents(ctx, ticket); err != nil {
		// Log error but don't fail the operation
	}
	
	return ticket, nil
}

// ResolveTicket handles the resolve ticket command
func (h *TicketCommandHandler) ResolveTicket(ctx context.Context, cmd ResolveTicketCommand) (*model.Ticket, error) {
	// Load the ticket
	ticket, err := h.ticketRepo.FindByID(ctx, cmd.TicketID)
	if err != nil {
		return nil, fmt.Errorf("failed to load ticket: %w", err)
	}
	
	// Resolve the ticket
	if err := ticket.Resolve(cmd.ResolvedBy, cmd.Resolution); err != nil {
		return nil, fmt.Errorf("failed to resolve ticket: %w", err)
	}
	
	// Save the ticket
	if err := h.ticketRepo.Save(ctx, ticket); err != nil {
		return nil, fmt.Errorf("failed to save ticket: %w", err)
	}
	
	// Publish domain events
	if err := h.publishDomainEvents(ctx, ticket); err != nil {
		// Log error but don't fail the operation
	}
	
	return ticket, nil
}

// CloseTicket handles the close ticket command
func (h *TicketCommandHandler) CloseTicket(ctx context.Context, cmd CloseTicketCommand) (*model.Ticket, error) {
	// Load the ticket
	ticket, err := h.ticketRepo.FindByID(ctx, cmd.TicketID)
	if err != nil {
		return nil, fmt.Errorf("failed to load ticket: %w", err)
	}
	
	// Validate access
	if !cmd.IsAdmin && !ticket.UserID().Equals(cmd.ClosedBy) {
		return nil, fmt.Errorf("access denied: only ticket owner or admins can close tickets")
	}
	
	// Close the ticket
	if err := ticket.Close(); err != nil {
		return nil, fmt.Errorf("failed to close ticket: %w", err)
	}
	
	// Save the ticket
	if err := h.ticketRepo.Save(ctx, ticket); err != nil {
		return nil, fmt.Errorf("failed to save ticket: %w", err)
	}
	
	// Publish domain events
	if err := h.publishDomainEvents(ctx, ticket); err != nil {
		// Log error but don't fail the operation
	}
	
	return ticket, nil
}

// AddTicketMessage handles the add ticket message command
func (h *TicketCommandHandler) AddTicketMessage(ctx context.Context, cmd AddTicketMessageCommand) (*model.Ticket, error) {
	// Load the ticket
	ticket, err := h.ticketRepo.FindByID(ctx, cmd.TicketID)
	if err != nil {
		return nil, fmt.Errorf("failed to load ticket: %w", err)
	}
	
	// Convert attachments
	attachments := make([]model.Attachment, len(cmd.Attachments))
	for i, att := range cmd.Attachments {
		attachments[i] = model.Attachment{
			Name: att.Name,
			URL:  att.URL,
			Size: att.Size,
			Type: att.Type,
		}
	}
	
	// Create the message
	message, err := model.NewTicketMessage(
		0, // Will be set by repository
		cmd.TicketID,
		cmd.UserID,
		cmd.Content,
		model.MessageType(cmd.MessageType),
		cmd.IsInternal,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}
	
	// Set attachments and metadata
	message.SetAttachments(attachments)
	for key, value := range cmd.Metadata {
		message.SetMetadata(key, value)
	}
	
	// Save the message first
	if err := h.ticketMessageRepo.Save(ctx, message); err != nil {
		return nil, fmt.Errorf("failed to save message: %w", err)
	}
	
	// Add message to ticket
	if err := ticket.AddMessage(*message); err != nil {
		return nil, fmt.Errorf("failed to add message to ticket: %w", err)
	}
	
	// Save the ticket (to update timing fields)
	if err := h.ticketRepo.Save(ctx, ticket); err != nil {
		return nil, fmt.Errorf("failed to save ticket: %w", err)
	}
	
	// Publish domain events
	if err := h.publishDomainEvents(ctx, ticket); err != nil {
		// Log error but don't fail the operation
	}
	
	return ticket, nil
}

// publishDomainEvents publishes domain events and clears them from the aggregate
func (h *TicketCommandHandler) publishDomainEvents(ctx context.Context, ticket *model.Ticket) error {
	events := ticket.DomainEvents()
	if len(events) == 0 {
		return nil
	}
	
	// Convert domain events to interface{}
	eventInterfaces := make([]interface{}, len(events))
	for i, event := range events {
		eventInterfaces[i] = event
	}
	
	// Publish events
	if err := h.eventPublisher.Publish(ctx, eventInterfaces...); err != nil {
		return fmt.Errorf("failed to publish domain events: %w", err)
	}
	
	// Clear events from aggregate
	ticket.ClearDomainEvents()
	
	return nil
}