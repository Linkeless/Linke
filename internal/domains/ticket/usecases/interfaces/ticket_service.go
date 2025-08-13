package interfaces

import (
	"context"
	"linke/internal/domains/ticket/dto"
	"linke/internal/domains/ticket/entities"
)

// TicketService defines the interface for ticket operations
type TicketService interface {
	// Ticket CRUD operations
	CreateTicket(ctx context.Context, userID uint, req *dto.CreateTicketRequest) (*entities.Ticket, error)
	GetTicket(ctx context.Context, ticketID uint) (*entities.Ticket, error)
	GetTicketByNumber(ctx context.Context, ticketNo string) (*entities.Ticket, error)
	UpdateTicket(ctx context.Context, ticketID uint, req *dto.UpdateTicketRequest) (*entities.Ticket, error)
	DeleteTicket(ctx context.Context, ticketID uint) error

	// Ticket listing and filtering
	GetTickets(ctx context.Context, req *dto.GetTicketsRequest) ([]*entities.Ticket, int64, error)
	GetUserTickets(ctx context.Context, userID uint, limit, offset int) ([]*entities.Ticket, int64, error)
	GetAssignedTickets(ctx context.Context, assignedToID uint, limit, offset int) ([]*entities.Ticket, int64, error)

	// Ticket assignment and status management
	AssignTicket(ctx context.Context, ticketID uint, req *dto.AssignTicketRequest) (*entities.Ticket, error)
	AutoAssignTicket(ctx context.Context, ticketID uint) (*entities.Ticket, error)
	UnassignTicket(ctx context.Context, ticketID uint) (*entities.Ticket, error)
	UpdateTicketStatus(ctx context.Context, ticketID uint, status string) (*entities.Ticket, error)
	UpdateTicketPriority(ctx context.Context, ticketID uint, priority string) (*entities.Ticket, error)

	// Agent workload management
	GetAgentWorkload(ctx context.Context, agentID uint) (int, error)
	GetAvailableAgents(ctx context.Context, category string) ([]*dto.AgentInfo, error)
	FindBestAgentForTicket(ctx context.Context, ticket *entities.Ticket) (uint, error)

	// Ticket resolution
	ResolveTicket(ctx context.Context, ticketID uint, resolvedByID uint, req *dto.ResolveTicketRequest) (*entities.Ticket, error)
	CloseTicket(ctx context.Context, ticketID uint, reason string) (*entities.Ticket, error)
	ReopenTicket(ctx context.Context, ticketID uint, reason string) (*entities.Ticket, error)

	// Ticket statistics and reporting
	GetTicketStatistics(ctx context.Context, fromDate, toDate string) (map[string]any, error)
	GetUserTicketStatistics(ctx context.Context, userID uint) (map[string]any, error)
	GetAgentTicketStatistics(ctx context.Context, agentID uint) (map[string]any, error)

	// Bulk operations
	BulkAssignTickets(ctx context.Context, ticketIDs []uint, assignedToID uint) error
	BulkUpdateTicketStatus(ctx context.Context, ticketIDs []uint, status string) error
}
