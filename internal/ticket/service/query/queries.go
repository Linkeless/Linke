package query

import (
	"linke/internal/ticket/domain/valueobject"
	sharedvo "linke/internal/shared/valueobject"
)

// GetTicketQuery represents a query to get a ticket by ID
type GetTicketQuery struct {
	TicketID valueobject.TicketID
	UserID   sharedvo.UserID // User making the request
	IsAdmin  bool
}

// GetTicketByNumberQuery represents a query to get a ticket by number
type GetTicketByNumberQuery struct {
	TicketNumber valueobject.TicketNumber
	UserID       sharedvo.UserID // User making the request
	IsAdmin      bool
}

// ListTicketsQuery represents a query to list tickets
type ListTicketsQuery struct {
	UserID       *sharedvo.UserID // Filter by user (nil for admin to see all)
	AssignedToID *sharedvo.UserID // Filter by assigned user
	Status       *string
	Priority     *string
	Category     *string
	SearchTerm   string
	Tags         []string
	Limit        int
	Offset       int
	SortBy       string // "created_at", "updated_at", "priority", "status"
	SortOrder    string // "asc", "desc"
	RequestedBy  sharedvo.UserID // User making the request
	IsAdmin      bool
}

// GetTicketMessagesQuery represents a query to get ticket messages
type GetTicketMessagesQuery struct {
	TicketID        valueobject.TicketID
	IncludeInternal bool
	Limit           int
	Offset          int
	RequestedBy     sharedvo.UserID // User making the request
	IsAdmin         bool
}

// GetTicketStatisticsQuery represents a query to get ticket statistics
type GetTicketStatisticsQuery struct {
	RequestedBy sharedvo.UserID // User making the request
	IsAdmin     bool
}