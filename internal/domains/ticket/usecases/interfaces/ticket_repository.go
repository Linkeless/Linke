package interfaces

import (
	"context"
	"time"

	"linke/internal/domains/ticket/entities"
	"linke/internal/shared/framework"
)

// TicketRepository defines the interface for ticket data access operations
type TicketRepository interface {
	// Combines UserScopedRepository and TimeBasedRepository functionality
	framework.UserScopedRepository[entities.Ticket, uint]
	framework.TimeBasedRepository[entities.Ticket, uint]

	// Status operations
	ListOpen(ctx context.Context, limit, offset int) ([]*entities.Ticket, int64, error)
	ListClosed(ctx context.Context, limit, offset int) ([]*entities.Ticket, int64, error)

	// Priority operations
	ListByPriority(ctx context.Context, priority string, limit, offset int) ([]*entities.Ticket, int64, error)
	ListHighPriority(ctx context.Context, limit, offset int) ([]*entities.Ticket, int64, error)

	// Assignment operations
	ListUnassigned(ctx context.Context, limit, offset int) ([]*entities.Ticket, int64, error)
	ListByAssignee(ctx context.Context, assigneeID uint, limit, offset int) ([]*entities.Ticket, int64, error)
	UpdateAssignee(ctx context.Context, id uint, assigneeID *uint) error

	// Time-based operations
	ListRecentTickets(ctx context.Context, since time.Time, limit, offset int) ([]*entities.Ticket, int64, error)

	// Statistics
	GetStatusStats(ctx context.Context) (map[string]int64, error)
	GetPriorityStats(ctx context.Context) (map[string]int64, error)
}
