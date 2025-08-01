package repository

import (
	"context"

	"linke/internal/ticket/domain/model"
	"linke/internal/ticket/domain/valueobject"
	sharedvo "linke/internal/shared/valueobject"
)

// TicketRepository defines the interface for ticket persistence
type TicketRepository interface {
	// Save saves a ticket aggregate
	Save(ctx context.Context, ticket *model.Ticket) error
	
	// FindByID finds a ticket by ID
	FindByID(ctx context.Context, id valueobject.TicketID) (*model.Ticket, error)
	
	// FindByTicketNumber finds a ticket by ticket number
	FindByTicketNumber(ctx context.Context, ticketNumber valueobject.TicketNumber) (*model.Ticket, error)
	
	// FindByUserID finds tickets by user ID with pagination
	FindByUserID(ctx context.Context, userID sharedvo.UserID, limit, offset int) ([]*model.Ticket, int64, error)
	
	// FindByAssignedTo finds tickets assigned to a user with pagination
	FindByAssignedTo(ctx context.Context, assignedToID sharedvo.UserID, limit, offset int) ([]*model.Ticket, int64, error)
	
	// FindByStatus finds tickets by status with pagination
	FindByStatus(ctx context.Context, status valueobject.TicketStatus, limit, offset int) ([]*model.Ticket, int64, error)
	
	// FindByPriority finds tickets by priority with pagination
	FindByPriority(ctx context.Context, priority valueobject.TicketPriority, limit, offset int) ([]*model.Ticket, int64, error)
	
	// FindByCategory finds tickets by category with pagination
	FindByCategory(ctx context.Context, category valueobject.TicketCategory, limit, offset int) ([]*model.Ticket, int64, error)
	
	// Search searches tickets by various criteria
	Search(ctx context.Context, criteria SearchCriteria) ([]*model.Ticket, int64, error)
	
	// Delete soft deletes a ticket
	Delete(ctx context.Context, id valueobject.TicketID) error
	
	// ExistsByTicketNumber checks if a ticket number already exists
	ExistsByTicketNumber(ctx context.Context, ticketNumber valueobject.TicketNumber) (bool, error)
	
	// GetStatistics returns ticket statistics
	GetStatistics(ctx context.Context) (*TicketStatistics, error)
}

// SearchCriteria defines search criteria for tickets
type SearchCriteria struct {
	UserID       *sharedvo.UserID
	AssignedToID *sharedvo.UserID
	Status       *valueobject.TicketStatus
	Priority     *valueobject.TicketPriority
	Category     *valueobject.TicketCategory
	SearchTerm   string // Searches in title, description, and ticket number
	Tags         []string
	Limit        int
	Offset       int
	SortBy       string // "created_at", "updated_at", "priority", "status"
	SortOrder    string // "asc", "desc"
}

// TicketStatistics contains ticket statistics
type TicketStatistics struct {
	Total         int64
	ByStatus      map[string]int64
	ByPriority    map[string]int64
	ByCategory    map[string]int64
	Unassigned    int64
	OverdueCount  int64 // Tickets that need response based on SLA
	ResolvedToday int64
	CreatedToday  int64
}

// TicketMessageRepository defines the interface for ticket message persistence
type TicketMessageRepository interface {
	// Save saves a ticket message
	Save(ctx context.Context, message *model.TicketMessage) error
	
	// FindByID finds a message by ID
	FindByID(ctx context.Context, id uint) (*model.TicketMessage, error)
	
	// FindByTicketID finds messages by ticket ID
	FindByTicketID(ctx context.Context, ticketID valueobject.TicketID, includeInternal bool, limit, offset int) ([]*model.TicketMessage, int64, error)
	
	// Delete soft deletes a message
	Delete(ctx context.Context, id uint) error
}