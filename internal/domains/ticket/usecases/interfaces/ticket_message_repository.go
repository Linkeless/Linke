package interfaces

import (
	"context"
	"linke/internal/domains/ticket/entities"
)

// TicketMessageRepository defines the interface for ticket message data access operations
type TicketMessageRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, message *entities.TicketMessage) error
	GetByID(ctx context.Context, messageID uint) (*entities.TicketMessage, error)
	Update(ctx context.Context, message *entities.TicketMessage) error
	Delete(ctx context.Context, messageID uint) error
	BatchDelete(ctx context.Context, messageIDs []uint) (int, []uint, error)

	// Message listing operations
	GetTicketMessages(ctx context.Context, ticketID uint, includeInternal bool, limit, offset int) ([]*entities.TicketMessage, int64, error)
	GetLatestTicketMessages(ctx context.Context, ticketID uint, limit int) ([]*entities.TicketMessage, error)
	GetInternalMessages(ctx context.Context, ticketID uint) ([]*entities.TicketMessage, error)

	// Message management
	MarkAsRead(ctx context.Context, messageID uint, userID uint) error
	MarkTicketMessagesAsRead(ctx context.Context, ticketID uint, userID uint) error

	// Statistics
	GetMessageStatistics(ctx context.Context, ticketID uint) (map[string]any, error)
}