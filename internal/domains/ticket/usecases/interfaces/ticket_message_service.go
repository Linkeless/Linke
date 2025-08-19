package interfaces

import (
	"context"

	"linke/internal/domains/ticket/dto"
	"linke/internal/domains/ticket/entities"
)

// TicketMessageService defines the interface for ticket message operations
type TicketMessageService interface {
	// Message CRUD operations
	CreateTicketMessage(ctx context.Context, ticketID, userID uint, req *dto.CreateTicketMessageRequest) (*entities.TicketMessage, error)
	GetTicketMessage(ctx context.Context, messageID uint) (*entities.TicketMessage, error)
	UpdateTicketMessage(ctx context.Context, messageID uint, req *dto.UpdateTicketMessageRequest) (*entities.TicketMessage, error)
	DeleteTicketMessage(ctx context.Context, messageID uint) error

	// Message listing
	GetTicketMessages(ctx context.Context, req *dto.GetTicketMessagesRequest) ([]*entities.TicketMessage, int64, error)
	GetLatestTicketMessages(ctx context.Context, ticketID uint, limit int) ([]*entities.TicketMessage, error)

	// Message management
	MarkMessageAsRead(ctx context.Context, messageID, userID uint) error
	MarkTicketMessagesAsRead(ctx context.Context, ticketID, userID uint) error

	// Internal messages
	CreateInternalMessage(ctx context.Context, ticketID, userID uint, content string) (*entities.TicketMessage, error)
	GetInternalMessages(ctx context.Context, ticketID uint) ([]*entities.TicketMessage, error)

	// Message statistics
	GetMessageStatistics(ctx context.Context, ticketID uint) (map[string]any, error)
}
