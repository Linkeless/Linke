package implementations

import (
	"context"

	"linke/internal/domains/ticket/entities"
	"linke/internal/domains/ticket/usecases/interfaces"
	userInterfaces "linke/internal/domains/user/usecases/interfaces"
	"linke/internal/shared/events"
	"linke/internal/shared/logger"
)

// EventAwareTicketService wraps the ticket service with event publishing capabilities
type EventAwareTicketService struct {
	ticketService interfaces.TicketService
	userService   userInterfaces.UserService
	eventBus      events.EventBus
}

// NewEventAwareTicketService creates a new event-aware ticket service
func NewEventAwareTicketService(
	ticketService interfaces.TicketService,
	userService userInterfaces.UserService,
	eventBus events.EventBus,
) interfaces.TicketService {
	return &EventAwareTicketService{
		ticketService: ticketService,
		userService:   userService,
		eventBus:      eventBus,
	}
}

// CreateTicket creates a new ticket and publishes an event
func (s *EventAwareTicketService) CreateTicket(ctx context.Context, userID uint, req *interfaces.CreateTicketRequest) (*entities.Ticket, error) {
	// Create the ticket
	ticket, err := s.ticketService.CreateTicket(ctx, userID, req)
	if err != nil {
		return nil, err
	}
	
	// Get user information for notification
	user, userErr := s.userService.GetUserByID(ctx, userID)
	
	// Prepare event data
	eventData := map[string]interface{}{
		"ticket_id":    ticket.ID,
		"ticket_no":    ticket.TicketNo,
		"title":        ticket.Title,
		"description":  ticket.Description,
		"category":     ticket.Category,
		"priority":     ticket.Priority,
		"status":       ticket.Status,
		"user_id":      userID,
	}
	
	// Add user information if available
	if userErr == nil && user != nil {
		eventData["user_name"] = user.Username
		if user.TelegramID != nil && *user.TelegramID != "" {
			eventData["user_telegram_id"] = *user.TelegramID
		}
	}
	
	// Publish ticket created event
	event := events.NewBaseEvent("ticket.created", "ticket-service", eventData)
	if err := s.eventBus.Publish(ctx, event); err != nil {
		logger.Error("Failed to publish ticket created event",
			logger.Uint("ticket_id", ticket.ID),
			logger.Error2("error", err))
	}
	
	return ticket, nil
}

// AssignTicket assigns a ticket to an agent and publishes an event
func (s *EventAwareTicketService) AssignTicket(ctx context.Context, ticketID uint, req *interfaces.AssignTicketRequest) (*entities.Ticket, error) {
	// Get ticket before assignment for comparison
	oldTicket, _ := s.ticketService.GetTicket(ctx, ticketID)
	
	// Assign the ticket
	ticket, err := s.ticketService.AssignTicket(ctx, ticketID, req)
	if err != nil {
		return nil, err
	}
	
	// Get assigned user information
	assignedUser, assignedErr := s.userService.GetUserByID(ctx, req.AssignedToID)
	
	// Get ticket creator information
	user, userErr := s.userService.GetUserByID(ctx, ticket.UserID)
	
	// Prepare event data
	eventData := map[string]interface{}{
		"ticket_id":      ticket.ID,
		"ticket_no":      ticket.TicketNo,
		"title":          ticket.Title,
		"priority":       ticket.Priority,
		"assigned_to_id": req.AssignedToID,
		"user_id":        ticket.UserID,
	}
	
	// Add assigned user information
	if assignedErr == nil && assignedUser != nil {
		eventData["assigned_to_name"] = assignedUser.Username
		if assignedUser.TelegramID != nil && *assignedUser.TelegramID != "" {
			eventData["assigned_telegram_id"] = *assignedUser.TelegramID
		}
	}
	
	// Add user information
	if userErr == nil && user != nil {
		eventData["user_name"] = user.Username
		if user.TelegramID != nil && *user.TelegramID != "" {
			eventData["user_telegram_id"] = *user.TelegramID
		}
	}
	
	// Add old assigned ID if changed
	if oldTicket != nil && oldTicket.AssignedToID != nil {
		eventData["old_assigned_to_id"] = *oldTicket.AssignedToID
	}
	
	// Publish ticket assigned event
	event := events.NewBaseEvent("ticket.assigned", "ticket-service", eventData)
	if err := s.eventBus.Publish(ctx, event); err != nil {
		logger.Error("Failed to publish ticket assigned event",
			logger.Uint("ticket_id", ticket.ID),
			logger.Error2("error", err))
	}
	
	return ticket, nil
}

// UpdateTicketStatus updates ticket status and publishes an event
func (s *EventAwareTicketService) UpdateTicketStatus(ctx context.Context, ticketID uint, status string) (*entities.Ticket, error) {
	// Get old ticket for comparison
	oldTicket, _ := s.ticketService.GetTicket(ctx, ticketID)
	oldStatus := ""
	if oldTicket != nil {
		oldStatus = oldTicket.Status
	}
	
	// Update the status
	ticket, err := s.ticketService.UpdateTicketStatus(ctx, ticketID, status)
	if err != nil {
		return nil, err
	}
	
	// Get user information
	user, userErr := s.userService.GetUserByID(ctx, ticket.UserID)
	
	// Prepare event data
	eventData := map[string]interface{}{
		"ticket_id":  ticket.ID,
		"ticket_no":  ticket.TicketNo,
		"title":      ticket.Title,
		"status":     status,
		"old_status": oldStatus,
		"user_id":    ticket.UserID,
	}
	
	// Add user information
	if userErr == nil && user != nil {
		eventData["user_name"] = user.Username
		if user.TelegramID != nil && *user.TelegramID != "" {
			eventData["user_telegram_id"] = *user.TelegramID
		}
	}
	
	// Add assigned user information if assigned
	if ticket.AssignedToID != nil {
		assignedUser, assignedErr := s.userService.GetUserByID(ctx, *ticket.AssignedToID)
		if assignedErr == nil && assignedUser != nil {
			eventData["assigned_to_id"] = *ticket.AssignedToID
			eventData["assigned_to_name"] = assignedUser.Username
			if assignedUser.TelegramID != nil && *assignedUser.TelegramID != "" {
				eventData["assigned_telegram_id"] = *assignedUser.TelegramID
			}
		}
	}
	
	// Publish status changed event
	event := events.NewBaseEvent("ticket.status_changed", "ticket-service", eventData)
	if err := s.eventBus.Publish(ctx, event); err != nil {
		logger.Error("Failed to publish ticket status changed event",
			logger.Uint("ticket_id", ticket.ID),
			logger.Error2("error", err))
	}
	
	return ticket, nil
}

// ResolveTicket resolves a ticket and publishes an event
func (s *EventAwareTicketService) ResolveTicket(ctx context.Context, ticketID uint, resolvedByID uint, req *interfaces.ResolveTicketRequest) (*entities.Ticket, error) {
	// Resolve the ticket
	ticket, err := s.ticketService.ResolveTicket(ctx, ticketID, resolvedByID, req)
	if err != nil {
		return nil, err
	}
	
	// Get user information
	user, userErr := s.userService.GetUserByID(ctx, ticket.UserID)
	
	// Get resolver information
	resolver, resolverErr := s.userService.GetUserByID(ctx, resolvedByID)
	
	// Prepare event data
	eventData := map[string]interface{}{
		"ticket_id":      ticket.ID,
		"ticket_no":      ticket.TicketNo,
		"title":          ticket.Title,
		"resolution":     req.Resolution,
		"resolved_by_id": resolvedByID,
		"user_id":        ticket.UserID,
	}
	
	// Add user information
	if userErr == nil && user != nil {
		eventData["user_name"] = user.Username
		if user.TelegramID != nil && *user.TelegramID != "" {
			eventData["user_telegram_id"] = *user.TelegramID
		}
	}
	
	// Add resolver information
	if resolverErr == nil && resolver != nil {
		eventData["resolved_by_name"] = resolver.Username
		if resolver.TelegramID != nil && *resolver.TelegramID != "" {
			eventData["resolved_telegram_id"] = *resolver.TelegramID
		}
	}
	
	// Publish ticket resolved event
	event := events.NewBaseEvent("ticket.resolved", "ticket-service", eventData)
	if err := s.eventBus.Publish(ctx, event); err != nil {
		logger.Error("Failed to publish ticket resolved event",
			logger.Uint("ticket_id", ticket.ID),
			logger.Error2("error", err))
	}
	
	return ticket, nil
}

// CloseTicket closes a ticket and publishes an event
func (s *EventAwareTicketService) CloseTicket(ctx context.Context, ticketID uint, reason string) (*entities.Ticket, error) {
	// Close the ticket
	ticket, err := s.ticketService.CloseTicket(ctx, ticketID, reason)
	if err != nil {
		return nil, err
	}
	
	// Get user information
	user, userErr := s.userService.GetUserByID(ctx, ticket.UserID)
	
	// Prepare event data
	eventData := map[string]interface{}{
		"ticket_id":     ticket.ID,
		"ticket_no":     ticket.TicketNo,
		"title":         ticket.Title,
		"closed_reason": reason,
		"user_id":       ticket.UserID,
	}
	
	// Add user information
	if userErr == nil && user != nil {
		eventData["user_name"] = user.Username
		if user.TelegramID != nil && *user.TelegramID != "" {
			eventData["user_telegram_id"] = *user.TelegramID
		}
	}
	
	// Publish ticket closed event
	event := events.NewBaseEvent("ticket.closed", "ticket-service", eventData)
	if err := s.eventBus.Publish(ctx, event); err != nil {
		logger.Error("Failed to publish ticket closed event",
			logger.Uint("ticket_id", ticket.ID),
			logger.Error2("error", err))
	}
	
	return ticket, nil
}

// UpdateTicketPriority updates ticket priority and publishes an event if escalated
func (s *EventAwareTicketService) UpdateTicketPriority(ctx context.Context, ticketID uint, priority string) (*entities.Ticket, error) {
	// Get old ticket for comparison
	oldTicket, _ := s.ticketService.GetTicket(ctx, ticketID)
	oldPriority := ""
	if oldTicket != nil {
		oldPriority = oldTicket.Priority
	}
	
	// Update the priority
	ticket, err := s.ticketService.UpdateTicketPriority(ctx, ticketID, priority)
	if err != nil {
		return nil, err
	}
	
	// Check if this is an escalation
	isEscalation := s.isEscalation(oldPriority, priority)
	
	if isEscalation {
		// Get user information
		user, userErr := s.userService.GetUserByID(ctx, ticket.UserID)
		
		// Prepare event data
		eventData := map[string]interface{}{
			"ticket_id":    ticket.ID,
			"ticket_no":    ticket.TicketNo,
			"title":        ticket.Title,
			"priority":     priority,
			"old_priority": oldPriority,
			"user_id":      ticket.UserID,
		}
		
		// Add user information
		if userErr == nil && user != nil {
			eventData["user_name"] = user.Username
			if user.TelegramID != nil && *user.TelegramID != "" {
				eventData["user_telegram_id"] = *user.TelegramID
			}
		}
		
		// Publish ticket escalated event
		event := events.NewBaseEvent("ticket.escalated", "ticket-service", eventData)
		if err := s.eventBus.Publish(ctx, event); err != nil {
			logger.Error("Failed to publish ticket escalated event",
				logger.Uint("ticket_id", ticket.ID),
				logger.Error2("error", err))
		}
	}
	
	return ticket, nil
}

// isEscalation checks if priority change is an escalation
func (s *EventAwareTicketService) isEscalation(oldPriority, newPriority string) bool {
	priorityLevels := map[string]int{
		entities.TicketPriorityLow:      1,
		entities.TicketPriorityNormal:   2,
		entities.TicketPriorityHigh:     3,
		entities.TicketPriorityUrgent:   4,
		entities.TicketPriorityCritical: 5,
	}
	
	oldLevel := priorityLevels[oldPriority]
	newLevel := priorityLevels[newPriority]
	
	return newLevel > oldLevel
}

// Delegate all other methods to the wrapped service

func (s *EventAwareTicketService) GetTicket(ctx context.Context, ticketID uint) (*entities.Ticket, error) {
	return s.ticketService.GetTicket(ctx, ticketID)
}

func (s *EventAwareTicketService) GetTicketByNumber(ctx context.Context, ticketNo string) (*entities.Ticket, error) {
	return s.ticketService.GetTicketByNumber(ctx, ticketNo)
}

func (s *EventAwareTicketService) UpdateTicket(ctx context.Context, ticketID uint, req *interfaces.UpdateTicketRequest) (*entities.Ticket, error) {
	return s.ticketService.UpdateTicket(ctx, ticketID, req)
}

func (s *EventAwareTicketService) DeleteTicket(ctx context.Context, ticketID uint) error {
	return s.ticketService.DeleteTicket(ctx, ticketID)
}

func (s *EventAwareTicketService) GetTickets(ctx context.Context, req *interfaces.GetTicketsRequest) ([]*entities.Ticket, int64, error) {
	return s.ticketService.GetTickets(ctx, req)
}

func (s *EventAwareTicketService) GetUserTickets(ctx context.Context, userID uint, limit, offset int) ([]*entities.Ticket, int64, error) {
	return s.ticketService.GetUserTickets(ctx, userID, limit, offset)
}

func (s *EventAwareTicketService) GetAssignedTickets(ctx context.Context, assignedToID uint, limit, offset int) ([]*entities.Ticket, int64, error) {
	return s.ticketService.GetAssignedTickets(ctx, assignedToID, limit, offset)
}

func (s *EventAwareTicketService) AutoAssignTicket(ctx context.Context, ticketID uint) (*entities.Ticket, error) {
	return s.ticketService.AutoAssignTicket(ctx, ticketID)
}

func (s *EventAwareTicketService) UnassignTicket(ctx context.Context, ticketID uint) (*entities.Ticket, error) {
	return s.ticketService.UnassignTicket(ctx, ticketID)
}

func (s *EventAwareTicketService) GetAgentWorkload(ctx context.Context, agentID uint) (int, error) {
	return s.ticketService.GetAgentWorkload(ctx, agentID)
}

func (s *EventAwareTicketService) GetAvailableAgents(ctx context.Context, category string) ([]*interfaces.AgentInfo, error) {
	return s.ticketService.GetAvailableAgents(ctx, category)
}

func (s *EventAwareTicketService) FindBestAgentForTicket(ctx context.Context, ticket *entities.Ticket) (uint, error) {
	return s.ticketService.FindBestAgentForTicket(ctx, ticket)
}

func (s *EventAwareTicketService) ReopenTicket(ctx context.Context, ticketID uint, reason string) (*entities.Ticket, error) {
	return s.ticketService.ReopenTicket(ctx, ticketID, reason)
}

func (s *EventAwareTicketService) GetTicketStatistics(ctx context.Context, fromDate, toDate string) (map[string]any, error) {
	return s.ticketService.GetTicketStatistics(ctx, fromDate, toDate)
}

func (s *EventAwareTicketService) GetUserTicketStatistics(ctx context.Context, userID uint) (map[string]any, error) {
	return s.ticketService.GetUserTicketStatistics(ctx, userID)
}

func (s *EventAwareTicketService) GetAgentTicketStatistics(ctx context.Context, agentID uint) (map[string]any, error) {
	return s.ticketService.GetAgentTicketStatistics(ctx, agentID)
}

func (s *EventAwareTicketService) BulkAssignTickets(ctx context.Context, ticketIDs []uint, assignedToID uint) error {
	return s.ticketService.BulkAssignTickets(ctx, ticketIDs, assignedToID)
}

func (s *EventAwareTicketService) BulkUpdateTicketStatus(ctx context.Context, ticketIDs []uint, status string) error {
	return s.ticketService.BulkUpdateTicketStatus(ctx, ticketIDs, status)
}