package query

import (
	"context"
	"fmt"

	"linke/internal/ticket/domain/model"
	"linke/internal/ticket/domain/repository"
	"linke/internal/ticket/domain/service"
	"linke/internal/ticket/domain/valueobject"
)

// TicketQueryHandler handles ticket-related queries
type TicketQueryHandler struct {
	ticketRepo        repository.TicketRepository
	ticketMessageRepo repository.TicketMessageRepository
	domainService     *service.TicketDomainService
}

// NewTicketQueryHandler creates a new ticket query handler
func NewTicketQueryHandler(
	ticketRepo repository.TicketRepository,
	ticketMessageRepo repository.TicketMessageRepository,
	domainService *service.TicketDomainService,
) *TicketQueryHandler {
	return &TicketQueryHandler{
		ticketRepo:        ticketRepo,
		ticketMessageRepo: ticketMessageRepo,
		domainService:     domainService,
	}
}

// GetTicket handles the get ticket query
func (h *TicketQueryHandler) GetTicket(ctx context.Context, query GetTicketQuery) (*model.Ticket, error) {
	// Check access permissions using domain service
	canAccess, err := h.domainService.CanUserAccessTicket(ctx, query.UserID, query.TicketID, query.IsAdmin)
	if err != nil {
		return nil, fmt.Errorf("failed to check access permissions: %w", err)
	}
	
	if !canAccess {
		return nil, fmt.Errorf("access denied: cannot access this ticket")
	}
	
	// Load the ticket
	ticket, err := h.ticketRepo.FindByID(ctx, query.TicketID)
	if err != nil {
		return nil, fmt.Errorf("failed to load ticket: %w", err)
	}
	
	return ticket, nil
}

// GetTicketByNumber handles the get ticket by number query
func (h *TicketQueryHandler) GetTicketByNumber(ctx context.Context, query GetTicketByNumberQuery) (*model.Ticket, error) {
	// Load the ticket first
	ticket, err := h.ticketRepo.FindByTicketNumber(ctx, query.TicketNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to load ticket: %w", err)
	}
	
	// Check access permissions
	canAccess, err := h.domainService.CanUserAccessTicket(ctx, query.UserID, ticket.ID(), query.IsAdmin)
	if err != nil {
		return nil, fmt.Errorf("failed to check access permissions: %w", err)
	}
	
	if !canAccess {
		return nil, fmt.Errorf("access denied: cannot access this ticket")
	}
	
	return ticket, nil
}

// ListTickets handles the list tickets query
func (h *TicketQueryHandler) ListTickets(ctx context.Context, query ListTicketsQuery) ([]*model.Ticket, int64, error) {
	// Build search criteria
	criteria := repository.SearchCriteria{
		Limit:     query.Limit,
		Offset:    query.Offset,
		SortBy:    query.SortBy,
		SortOrder: query.SortOrder,
	}
	
	// Set default sort if not specified
	if criteria.SortBy == "" {
		criteria.SortBy = "created_at"
	}
	if criteria.SortOrder == "" {
		criteria.SortOrder = "desc"
	}
	
	// Apply filters based on user permissions
	if !query.IsAdmin {
		// Non-admin users can only see their own tickets
		criteria.UserID = &query.RequestedBy
	} else {
		// Admin users can filter by any user
		if query.UserID != nil {
			criteria.UserID = query.UserID
		}
	}
	
	// Apply other filters
	if query.AssignedToID != nil {
		criteria.AssignedToID = query.AssignedToID
	}
	
	if query.Status != nil {
		status, err := valueobject.NewTicketStatus(*query.Status)
		if err != nil {
			return nil, 0, fmt.Errorf("invalid status: %w", err)
		}
		criteria.Status = &status
	}
	
	if query.Priority != nil {
		priority, err := valueobject.NewTicketPriority(*query.Priority)
		if err != nil {
			return nil, 0, fmt.Errorf("invalid priority: %w", err)
		}
		criteria.Priority = &priority
	}
	
	if query.Category != nil {
		category, err := valueobject.NewTicketCategory(*query.Category)
		if err != nil {
			return nil, 0, fmt.Errorf("invalid category: %w", err)
		}
		criteria.Category = &category
	}
	
	if query.SearchTerm != "" {
		criteria.SearchTerm = query.SearchTerm
	}
	
	if len(query.Tags) > 0 {
		criteria.Tags = query.Tags
	}
	
	// Execute search
	tickets, total, err := h.ticketRepo.Search(ctx, criteria)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search tickets: %w", err)
	}
	
	return tickets, total, nil
}

// GetTicketMessages handles the get ticket messages query
func (h *TicketQueryHandler) GetTicketMessages(ctx context.Context, query GetTicketMessagesQuery) ([]*model.TicketMessage, int64, error) {
	// Check access permissions
	canAccess, err := h.domainService.CanUserAccessTicket(ctx, query.RequestedBy, query.TicketID, query.IsAdmin)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to check access permissions: %w", err)
	}
	
	if !canAccess {
		return nil, 0, fmt.Errorf("access denied: cannot access this ticket")
	}
	
	// Non-admin users cannot see internal messages
	includeInternal := query.IncludeInternal && query.IsAdmin
	
	// Load messages
	messages, total, err := h.ticketMessageRepo.FindByTicketID(ctx, query.TicketID, includeInternal, query.Limit, query.Offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to load ticket messages: %w", err)
	}
	
	return messages, total, nil
}

// GetTicketStatistics handles the get ticket statistics query
func (h *TicketQueryHandler) GetTicketStatistics(ctx context.Context, query GetTicketStatisticsQuery) (*repository.TicketStatistics, error) {
	// Only admins can access statistics
	if !query.IsAdmin {
		return nil, fmt.Errorf("access denied: only admins can access ticket statistics")
	}
	
	stats, err := h.ticketRepo.GetStatistics(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get ticket statistics: %w", err)
	}
	
	return stats, nil
}