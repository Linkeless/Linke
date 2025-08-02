package interfaces

import (
	"context"
	"linke/internal/domains/ticket/entities"
	"time"
)

// TicketRepository defines the interface for ticket data access operations
type TicketRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, ticket *entities.Ticket) error
	GetByID(ctx context.Context, id uint) (*entities.Ticket, error)
	Update(ctx context.Context, ticket *entities.Ticket) error
	Delete(ctx context.Context, id uint) error

	// User operations
	ListByUser(ctx context.Context, userID uint, limit, offset int) ([]*entities.Ticket, int64, error)
	CountByUser(ctx context.Context, userID uint) (int64, error)

	// Status operations
	ListByStatus(ctx context.Context, status string, limit, offset int) ([]*entities.Ticket, int64, error)
	ListOpen(ctx context.Context, limit, offset int) ([]*entities.Ticket, int64, error)
	ListClosed(ctx context.Context, limit, offset int) ([]*entities.Ticket, int64, error)
	UpdateStatus(ctx context.Context, id uint, status string) error

	// Priority operations
	ListByPriority(ctx context.Context, priority string, limit, offset int) ([]*entities.Ticket, int64, error)
	ListHighPriority(ctx context.Context, limit, offset int) ([]*entities.Ticket, int64, error)

	// Assignment operations
	ListUnassigned(ctx context.Context, limit, offset int) ([]*entities.Ticket, int64, error)
	ListByAssignee(ctx context.Context, assigneeID uint, limit, offset int) ([]*entities.Ticket, int64, error)
	UpdateAssignee(ctx context.Context, id uint, assigneeID *uint) error

	// Time-based operations
	ListByDateRange(ctx context.Context, start, end time.Time, limit, offset int) ([]*entities.Ticket, int64, error)
	ListRecentTickets(ctx context.Context, since time.Time, limit, offset int) ([]*entities.Ticket, int64, error)

	// List operations
	List(ctx context.Context, limit, offset int) ([]*entities.Ticket, int64, error)
	CountTotal(ctx context.Context) (int64, error)

	// Statistics
	GetStatusStats(ctx context.Context) (map[string]int64, error)
	GetPriorityStats(ctx context.Context) (map[string]int64, error)

	// Search operations
	Search(ctx context.Context, query string, limit, offset int) ([]*entities.Ticket, int64, error)
}
