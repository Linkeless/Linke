package telegram

import (
	"context"
	"encoding/json"
	"fmt"

	"linke/internal/shared/events"
	"linke/internal/shared/logger"
)

// TicketEventHandler handles ticket-related events and sends Telegram notifications
type TicketEventHandler struct {
	bot         *BotEnhanced
	eventTypes  []string
	handlerID   string
	enabled     bool
}

// NewTicketEventHandler creates a new ticket event handler
func NewTicketEventHandler(bot *BotEnhanced) *TicketEventHandler {
	return &TicketEventHandler{
		bot:       bot,
		handlerID: "telegram-ticket-handler",
		enabled:   bot.cfg.Telegram.EnableTicketNotifications,
		eventTypes: []string{
			"ticket.created",
			"ticket.assigned",
			"ticket.status_changed",
			"ticket.replied",
			"ticket.resolved",
			"ticket.closed",
			"ticket.escalated",
		},
	}
}

// Handle processes ticket events and sends notifications
func (h *TicketEventHandler) Handle(ctx context.Context, event events.Event) error {
	if !h.enabled {
		return nil
	}
	
	logger.Info("Handling ticket event",
		logger.String("event_type", event.EventType()),
		logger.String("event_id", event.EventID()))
	
	// Parse event data
	notification, err := h.parseEventToNotification(event)
	if err != nil {
		logger.Error("Failed to parse ticket event",
			logger.String("event_type", event.EventType()),
			logger.ErrorField(err))
		return err
	}
	
	// Send notifications based on event type
	switch event.EventType() {
	case "ticket.created":
		// Notify admins about new ticket
		if err := h.bot.SendTicketNotificationToAdmins(notification); err != nil {
			logger.Error("Failed to send ticket created notification to admins",
				logger.ErrorField(err))
		}
		
		// Notify user about ticket creation
		if userTelegramID := h.getUserTelegramID(event); userTelegramID != "" {
			if err := h.bot.SendTicketNotificationToUser(userTelegramID, notification); err != nil {
				logger.Error("Failed to send ticket created notification to user",
					logger.String("telegram_id", userTelegramID),
					logger.ErrorField(err))
			}
		}
		
	case "ticket.assigned":
		// Notify assigned agent
		if assignedTelegramID := h.getAssignedTelegramID(event); assignedTelegramID != "" {
			chatID := h.parseChatID(assignedTelegramID)
			if chatID != 0 {
				if err := h.bot.SendTicketNotification(chatID, notification); err != nil {
					logger.Error("Failed to send ticket assigned notification",
						logger.Int64("chat_id", chatID),
						logger.ErrorField(err))
				}
			}
		}
		
	case "ticket.replied":
		// Determine if reply is from admin or user and notify accordingly
		if isAdminReply := h.isAdminReply(event); isAdminReply {
			// Notify user about admin reply
			if userTelegramID := h.getUserTelegramID(event); userTelegramID != "" {
				if err := h.bot.SendTicketNotificationToUser(userTelegramID, notification); err != nil {
					logger.Error("Failed to send reply notification to user",
						logger.String("telegram_id", userTelegramID),
						logger.ErrorField(err))
				}
			}
		} else {
			// Notify admins about user reply
			if err := h.bot.SendTicketNotificationToAdmins(notification); err != nil {
				logger.Error("Failed to send reply notification to admins",
					logger.ErrorField(err))
			}
			
			// Also notify assigned agent if exists
			if assignedTelegramID := h.getAssignedTelegramID(event); assignedTelegramID != "" {
				chatID := h.parseChatID(assignedTelegramID)
				if chatID != 0 {
					if err := h.bot.SendTicketNotification(chatID, notification); err != nil {
						logger.Error("Failed to send reply notification to assigned agent",
							logger.Int64("chat_id", chatID),
							logger.ErrorField(err))
					}
				}
			}
		}
		
	case "ticket.resolved", "ticket.closed":
		// Notify user about resolution/closure
		if userTelegramID := h.getUserTelegramID(event); userTelegramID != "" {
			if err := h.bot.SendTicketNotificationToUser(userTelegramID, notification); err != nil {
				logger.Error("Failed to send resolution/closure notification to user",
					logger.String("telegram_id", userTelegramID),
					logger.ErrorField(err))
			}
		}
		
	case "ticket.escalated":
		// Notify all admins about escalation
		notification.Priority = "urgent" // Override priority for escalated tickets
		if err := h.bot.SendTicketNotificationToAdmins(notification); err != nil {
			logger.Error("Failed to send escalation notification to admins",
				logger.ErrorField(err))
		}
		
	case "ticket.status_changed":
		// Notify assigned agent about status change
		if assignedTelegramID := h.getAssignedTelegramID(event); assignedTelegramID != "" {
			chatID := h.parseChatID(assignedTelegramID)
			if chatID != 0 {
				if err := h.bot.SendTicketNotification(chatID, notification); err != nil {
					logger.Error("Failed to send status change notification",
						logger.Int64("chat_id", chatID),
						logger.ErrorField(err))
				}
			}
		}
	}
	
	return nil
}

// EventTypes returns the event types this handler listens to
func (h *TicketEventHandler) EventTypes() []string {
	return h.eventTypes
}

// ID returns the handler ID
func (h *TicketEventHandler) ID() string {
	return h.handlerID
}

// parseEventToNotification converts an event to a TicketNotification
func (h *TicketEventHandler) parseEventToNotification(event events.Event) (*TicketNotification, error) {
	notification := NewTicketNotification(TicketNotificationType(event.EventType()[7:])) // Remove "ticket." prefix
	
	// Get event data
	data := event.EventData()
	
	// Try to parse as JSON if it's a string
	var eventData map[string]interface{}
	switch v := data.(type) {
	case string:
		if err := json.Unmarshal([]byte(v), &eventData); err != nil {
			return nil, fmt.Errorf("failed to unmarshal event data: %w", err)
		}
	case map[string]interface{}:
		eventData = v
	default:
		// Try to convert to JSON and back to map
		jsonBytes, err := json.Marshal(data)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal event data: %w", err)
		}
		if err := json.Unmarshal(jsonBytes, &eventData); err != nil {
			return nil, fmt.Errorf("failed to unmarshal event data: %w", err)
		}
	}
	
	// Fill notification fields from event data
	// Handle ticket_id with multiple possible types (uint, int, float64)
	if ticketIDVal, ok := eventData["ticket_id"]; ok {
		switch v := ticketIDVal.(type) {
		case uint:
			notification.TicketID = v
		case int:
			notification.TicketID = uint(v)
		case float64:
			notification.TicketID = uint(v)
		case int64:
			notification.TicketID = uint(v)
		case uint64:
			notification.TicketID = uint(v)
		default:
			logger.Warn("Unexpected ticket_id type",
				logger.String("type", fmt.Sprintf("%T", v)),
				logger.Any("value", v))
		}
	}
	
	if ticketNo, ok := eventData["ticket_no"].(string); ok {
		notification.TicketNo = ticketNo
	}
	
	if title, ok := eventData["title"].(string); ok {
		notification.Title = title
	}
	
	if description, ok := eventData["description"].(string); ok {
		notification.Description = description
	}
	
	if category, ok := eventData["category"].(string); ok {
		notification.Category = category
	}
	
	if priority, ok := eventData["priority"].(string); ok {
		notification.Priority = priority
	}
	
	if status, ok := eventData["status"].(string); ok {
		notification.Status = status
	}
	
	if oldStatus, ok := eventData["old_status"].(string); ok {
		notification.OldStatus = oldStatus
	}
	
	// Handle user_id with multiple possible types
	if userIDVal, ok := eventData["user_id"]; ok {
		switch v := userIDVal.(type) {
		case uint:
			notification.UserID = v
		case int:
			notification.UserID = uint(v)
		case float64:
			notification.UserID = uint(v)
		case int64:
			notification.UserID = uint(v)
		case uint64:
			notification.UserID = uint(v)
		}
	}
	
	if userName, ok := eventData["user_name"].(string); ok {
		notification.UserName = userName
	}
	
	// Handle assigned_to_id with multiple possible types
	if assignedToIDVal, ok := eventData["assigned_to_id"]; ok {
		switch v := assignedToIDVal.(type) {
		case uint:
			notification.AssignedToID = v
		case int:
			notification.AssignedToID = uint(v)
		case float64:
			notification.AssignedToID = uint(v)
		case int64:
			notification.AssignedToID = uint(v)
		case uint64:
			notification.AssignedToID = uint(v)
		}
	}
	
	if assignedToName, ok := eventData["assigned_to_name"].(string); ok {
		notification.AssignedToName = assignedToName
	}
	
	// Handle replied_by_id with multiple possible types
	if repliedByIDVal, ok := eventData["replied_by_id"]; ok {
		switch v := repliedByIDVal.(type) {
		case uint:
			notification.RepliedByID = v
		case int:
			notification.RepliedByID = uint(v)
		case float64:
			notification.RepliedByID = uint(v)
		case int64:
			notification.RepliedByID = uint(v)
		case uint64:
			notification.RepliedByID = uint(v)
		}
	}
	
	if repliedByName, ok := eventData["replied_by_name"].(string); ok {
		notification.RepliedByName = repliedByName
	}
	
	if replyContent, ok := eventData["reply_content"].(string); ok {
		notification.ReplyContent = replyContent
	}
	
	if resolution, ok := eventData["resolution"].(string); ok {
		notification.Resolution = resolution
	}
	
	if closedReason, ok := eventData["closed_reason"].(string); ok {
		notification.ClosedReason = closedReason
	}
	
	// Store additional metadata
	notification.Metadata = eventData
	
	return notification, nil
}

// Helper methods to extract information from events

func (h *TicketEventHandler) getUserTelegramID(event events.Event) string {
	// Extract user telegram ID from event metadata
	if metadata, ok := event.(*events.BaseEvent).Metadata["user_telegram_id"]; ok {
		if telegramID, ok := metadata.(string); ok {
			return telegramID
		}
	}
	
	// Try to get from event data
	if data, ok := event.EventData().(map[string]interface{}); ok {
		if telegramID, ok := data["user_telegram_id"].(string); ok {
			return telegramID
		}
	}
	
	return ""
}

func (h *TicketEventHandler) getAssignedTelegramID(event events.Event) string {
	// Extract assigned user telegram ID from event metadata
	if metadata, ok := event.(*events.BaseEvent).Metadata["assigned_telegram_id"]; ok {
		if telegramID, ok := metadata.(string); ok {
			return telegramID
		}
	}
	
	// Try to get from event data
	if data, ok := event.EventData().(map[string]interface{}); ok {
		if telegramID, ok := data["assigned_telegram_id"].(string); ok {
			return telegramID
		}
	}
	
	return ""
}

func (h *TicketEventHandler) isAdminReply(event events.Event) bool {
	// Check if reply is from an admin
	if metadata, ok := event.(*events.BaseEvent).Metadata["is_admin_reply"]; ok {
		if isAdmin, ok := metadata.(bool); ok {
			return isAdmin
		}
	}
	
	// Try to get from event data
	if data, ok := event.EventData().(map[string]interface{}); ok {
		if messageType, ok := data["message_type"].(string); ok {
			return messageType == "admin" || messageType == "system"
		}
		if isAdmin, ok := data["is_admin_reply"].(bool); ok {
			return isAdmin
		}
	}
	
	return false
}

func (h *TicketEventHandler) parseChatID(telegramID string) int64 {
	// Parse telegram ID to int64
	var chatID int64
	if _, err := fmt.Sscanf(telegramID, "%d", &chatID); err == nil {
		return chatID
	}
	return 0
}