package interfaces

import (
	"context"
	"linke/internal/domains/ticket/entities"
)

// TicketService defines the interface for ticket operations
type TicketService interface {
	// Ticket CRUD operations
	CreateTicket(ctx context.Context, userID uint, req *CreateTicketRequest) (*entities.Ticket, error)
	GetTicket(ctx context.Context, ticketID uint) (*entities.Ticket, error)
	GetTicketByNumber(ctx context.Context, ticketNo string) (*entities.Ticket, error)
	UpdateTicket(ctx context.Context, ticketID uint, req *UpdateTicketRequest) (*entities.Ticket, error)
	DeleteTicket(ctx context.Context, ticketID uint) error

	// Ticket listing and filtering
	GetTickets(ctx context.Context, req *GetTicketsRequest) ([]*entities.Ticket, int64, error)
	GetUserTickets(ctx context.Context, userID uint, limit, offset int) ([]*entities.Ticket, int64, error)
	GetAssignedTickets(ctx context.Context, assignedToID uint, limit, offset int) ([]*entities.Ticket, int64, error)

	// Ticket assignment and status management
	AssignTicket(ctx context.Context, ticketID uint, req *AssignTicketRequest) (*entities.Ticket, error)
	UnassignTicket(ctx context.Context, ticketID uint) (*entities.Ticket, error)
	UpdateTicketStatus(ctx context.Context, ticketID uint, status string) (*entities.Ticket, error)
	UpdateTicketPriority(ctx context.Context, ticketID uint, priority string) (*entities.Ticket, error)

	// Ticket resolution
	ResolveTicket(ctx context.Context, ticketID uint, req *ResolveTicketRequest) (*entities.Ticket, error)
	CloseTicket(ctx context.Context, ticketID uint, reason string) (*entities.Ticket, error)
	ReopenTicket(ctx context.Context, ticketID uint, reason string) (*entities.Ticket, error)

	// Ticket statistics and reporting
	GetTicketStatistics(ctx context.Context, fromDate, toDate string) (map[string]interface{}, error)
	GetUserTicketStatistics(ctx context.Context, userID uint) (map[string]interface{}, error)
	GetAgentTicketStatistics(ctx context.Context, agentID uint) (map[string]interface{}, error)

	// Bulk operations
	BulkAssignTickets(ctx context.Context, ticketIDs []uint, assignedToID uint) error
	BulkUpdateTicketStatus(ctx context.Context, ticketIDs []uint, status string) error
}

// CreateTicketRequest represents the request to create a ticket
type CreateTicketRequest struct {
	Title       string `json:"title" binding:"required,min=5,max=255" example:"Unable to access my account"`
	Description string `json:"description" binding:"required,min=10,max=5000" example:"I cannot log in to my account even with correct credentials"`
	Category    string `json:"category" binding:"required,oneof=general technical billing account feature bug subscription payment" example:"technical"`
	Priority    string `json:"priority" binding:"omitempty,oneof=low normal high urgent critical" example:"normal"`
	Tags        string `json:"tags,omitempty" example:"login,authentication"`
	Metadata    string `json:"metadata,omitempty" example:"{\"browser\": \"Chrome\", \"os\": \"Windows\"}"`
}

// UpdateTicketRequest represents the request to update a ticket
type UpdateTicketRequest struct {
	Title       *string `json:"title,omitempty" binding:"omitempty,min=5,max=255" example:"Updated ticket title"`
	Description *string `json:"description,omitempty" binding:"omitempty,min=10,max=5000" example:"Updated description"`
	Category    *string `json:"category,omitempty" binding:"omitempty,oneof=general technical billing account feature bug subscription payment" example:"billing"`
	Priority    *string `json:"priority,omitempty" binding:"omitempty,oneof=low normal high urgent critical" example:"high"`
	Status      *string `json:"status,omitempty" binding:"omitempty,oneof=open in_progress pending resolved closed" example:"in_progress"`
	Tags        *string `json:"tags,omitempty" example:"urgent,billing"`
	Metadata    *string `json:"metadata,omitempty" example:"{\"updated_by\": \"admin\"}"`
}

// AssignTicketRequest represents the request to assign a ticket
type AssignTicketRequest struct {
	AssignedToID uint `json:"assigned_to_id" binding:"required" example:"2"`
}

// ResolveTicketRequest represents the request to resolve a ticket
type ResolveTicketRequest struct {
	Resolution string `json:"resolution" binding:"required,min=10,max=5000" example:"Issue resolved by updating user permissions"`
}

// GetTicketsRequest represents the request to get tickets with filters
type GetTicketsRequest struct {
	UserID       uint   `form:"user_id" example:"1"`
	AssignedToID *uint  `form:"assigned_to_id" example:"2"`
	Status       string `form:"status" binding:"omitempty,oneof=open in_progress pending resolved closed" example:"open"`
	Priority     string `form:"priority" binding:"omitempty,oneof=low normal high urgent critical" example:"high"`
	Category     string `form:"category" binding:"omitempty,oneof=general technical billing account feature bug subscription payment" example:"technical"`
	Search       string `form:"search" example:"login issue"`
	Limit        int    `form:"limit" binding:"omitempty,min=1,max=100" example:"10"`
	Offset       int    `form:"offset" binding:"omitempty,min=0" example:"0"`
}