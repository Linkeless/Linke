package implementations

import (
	"context"

	"linke/internal/domains/ticket/entities"
	"linke/internal/domains/ticket/usecases/interfaces"
	userInterfaces "linke/internal/domains/user/usecases/interfaces"
	"linke/internal/shared/events"
	"linke/internal/shared/logger"
)

// EventAwareTicketMessageService wraps the ticket message service with event publishing capabilities
type EventAwareTicketMessageService struct {
	ticketMessageService interfaces.TicketMessageService
	ticketService        interfaces.TicketService
	userService          userInterfaces.UserService
	eventBus             events.EventBus
}

// NewEventAwareTicketMessageService creates a new event-aware ticket message service
func NewEventAwareTicketMessageService(
	ticketMessageService interfaces.TicketMessageService,
	ticketService interfaces.TicketService,
	userService userInterfaces.UserService,
	eventBus events.EventBus,
) interfaces.TicketMessageService {
	return &EventAwareTicketMessageService{
		ticketMessageService: ticketMessageService,
		ticketService:        ticketService,
		userService:          userService,
		eventBus:             eventBus,
	}
}

// CreateTicketMessage creates a new ticket message and publishes an event
func (s *EventAwareTicketMessageService) CreateTicketMessage(ctx context.Context, ticketID uint, userID uint, req *interfaces.CreateTicketMessageRequest) (*entities.TicketMessage, error) {
	// Create the message
	message, err := s.ticketMessageService.CreateTicketMessage(ctx, ticketID, userID, req)
	if err != nil {
		return nil, err
	}

	// Get ticket information
	ticket, ticketErr := s.ticketService.GetTicket(ctx, ticketID)
	
	// Get message author information
	user, userErr := s.userService.GetUserByID(ctx, userID)
	
	// Prepare event data
	eventData := map[string]interface{}{
		"ticket_id":      ticketID,
		"message_id":     message.ID,
		"message_type":   message.MessageType,
		"reply_content":  message.Content,
		"replied_by_id":  userID,
		"is_internal":    message.IsInternal,
		"user_id":        userID, // Will be overridden with ticket owner ID below if ticket info is available
	}

	// Add ticket information if available
	if ticketErr == nil && ticket != nil {
		eventData["ticket_no"] = ticket.TicketNo
		eventData["title"] = ticket.Title
		eventData["priority"] = ticket.Priority
		eventData["status"] = ticket.Status
		eventData["user_id"] = ticket.UserID // Ticket owner's ID
		
		// Get ticket owner information for notifications
		ticketOwner, ticketOwnerErr := s.userService.GetUserByID(ctx, ticket.UserID)
		if ticketOwnerErr == nil && ticketOwner != nil {
			eventData["user_name"] = ticketOwner.Username
			if ticketOwner.TelegramID != nil && *ticketOwner.TelegramID != "" {
				eventData["user_telegram_id"] = *ticketOwner.TelegramID
			}
		}
		
		// Add assigned user information if exists
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
	}

	// Add message author information
	if userErr == nil && user != nil {
		eventData["replied_by_name"] = user.Username
		if user.TelegramID != nil && *user.TelegramID != "" {
			eventData["replied_by_telegram_id"] = *user.TelegramID
		}
	}

	// Determine if this is an admin reply or user reply for notification routing
	isAdminReply := message.MessageType == "admin" || message.MessageType == "system"
	eventData["is_admin_reply"] = isAdminReply

	// Only publish event for non-internal messages or if explicitly needed
	if !message.IsInternal || isAdminReply {
		// Create event with metadata for event handler
		baseEvent := events.NewBaseEvent("ticket.replied", "ticket-message-service", eventData)
		
		// Add metadata for event handler routing
		if baseEvent.Metadata == nil {
			baseEvent.Metadata = make(map[string]interface{})
		}
		baseEvent.Metadata["is_admin_reply"] = isAdminReply
		if ticket != nil && ticket.UserID > 0 {
			if userTelegramID, exists := eventData["user_telegram_id"]; exists {
				baseEvent.Metadata["user_telegram_id"] = userTelegramID
			}
		}
		if assignedTelegramID, exists := eventData["assigned_telegram_id"]; exists {
			baseEvent.Metadata["assigned_telegram_id"] = assignedTelegramID
		}
		
		if err := s.eventBus.Publish(ctx, baseEvent); err != nil {
			logger.Error("Failed to publish ticket replied event",
				logger.Uint("message_id", message.ID),
				logger.Uint("ticket_id", ticketID),
				logger.Error2("error", err))
		}
	}

	logger.Info("Ticket message created with event published",
		logger.Uint("message_id", message.ID),
		logger.Uint("ticket_id", ticketID),
		logger.Uint("user_id", userID),
		logger.String("message_type", message.MessageType))

	return message, nil
}

// CreateInternalMessage creates an internal message and publishes event if needed
func (s *EventAwareTicketMessageService) CreateInternalMessage(ctx context.Context, ticketID uint, userID uint, content string) (*entities.TicketMessage, error) {
	// Create the internal message using the regular service
	message, err := s.ticketMessageService.CreateInternalMessage(ctx, ticketID, userID, content)
	if err != nil {
		return nil, err
	}

	// Internal messages typically don't need events unless they are system messages
	// that should notify users. For now, we'll skip event publishing for internal messages
	// to avoid noise, but this can be customized based on business requirements.

	logger.Info("Internal ticket message created",
		logger.Uint("message_id", message.ID),
		logger.Uint("ticket_id", ticketID),
		logger.Uint("user_id", userID))

	return message, nil
}

// Delegate all other methods to the wrapped service

func (s *EventAwareTicketMessageService) GetTicketMessage(ctx context.Context, messageID uint) (*entities.TicketMessage, error) {
	return s.ticketMessageService.GetTicketMessage(ctx, messageID)
}

func (s *EventAwareTicketMessageService) UpdateTicketMessage(ctx context.Context, messageID uint, req *interfaces.UpdateTicketMessageRequest) (*entities.TicketMessage, error) {
	return s.ticketMessageService.UpdateTicketMessage(ctx, messageID, req)
}

func (s *EventAwareTicketMessageService) DeleteTicketMessage(ctx context.Context, messageID uint) error {
	return s.ticketMessageService.DeleteTicketMessage(ctx, messageID)
}

func (s *EventAwareTicketMessageService) GetTicketMessages(ctx context.Context, req *interfaces.GetTicketMessagesRequest) ([]*entities.TicketMessage, int64, error) {
	return s.ticketMessageService.GetTicketMessages(ctx, req)
}

func (s *EventAwareTicketMessageService) GetLatestTicketMessages(ctx context.Context, ticketID uint, limit int) ([]*entities.TicketMessage, error) {
	return s.ticketMessageService.GetLatestTicketMessages(ctx, ticketID, limit)
}

func (s *EventAwareTicketMessageService) MarkMessageAsRead(ctx context.Context, messageID uint, userID uint) error {
	return s.ticketMessageService.MarkMessageAsRead(ctx, messageID, userID)
}

func (s *EventAwareTicketMessageService) MarkTicketMessagesAsRead(ctx context.Context, ticketID uint, userID uint) error {
	return s.ticketMessageService.MarkTicketMessagesAsRead(ctx, ticketID, userID)
}

func (s *EventAwareTicketMessageService) GetInternalMessages(ctx context.Context, ticketID uint) ([]*entities.TicketMessage, error) {
	return s.ticketMessageService.GetInternalMessages(ctx, ticketID)
}

func (s *EventAwareTicketMessageService) GetMessageStatistics(ctx context.Context, ticketID uint) (map[string]any, error) {
	return s.ticketMessageService.GetMessageStatistics(ctx, ticketID)
}