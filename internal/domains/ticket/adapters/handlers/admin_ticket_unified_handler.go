package handlers

import (
	ticketInterfaces "linke/internal/domains/ticket/usecases/interfaces"
	userInterfaces "linke/internal/domains/user/usecases/interfaces"
)

// AdminTicketHandler provides comprehensive admin ticket management functionality
// This is a composition of all specialized handlers to maintain backward compatibility
type AdminTicketHandler struct {
	*AdminTicketBaseHandler
	*AdminTicketStatusHandler
	*AdminTicketMessageHandler
	*AdminTicketSearchHandler
	*AdminTicketBulkHandler
}

// NewAdminTicketHandler creates a new admin ticket handler with all components
func NewAdminTicketHandler(
	ticketService ticketInterfaces.TicketService,
	ticketMessageService ticketInterfaces.TicketMessageService,
	userService userInterfaces.UserService,
) *AdminTicketHandler {
	// Create the base handler
	base := NewAdminTicketHandlerBase(ticketService, ticketMessageService, userService)

	// Create all specialized handlers
	baseHandler := NewAdminTicketBaseHandler(base)
	statusHandler := NewAdminTicketStatusHandler(base)
	messageHandler := NewAdminTicketMessageHandler(base)
	searchHandler := NewAdminTicketSearchHandler(base)
	bulkHandler := NewAdminTicketBulkHandler(base)

	return &AdminTicketHandler{
		baseHandler,
		statusHandler,
		messageHandler,
		searchHandler,
		bulkHandler,
	}
}

// All methods are now available through embedded structs:
//
// Base Operations:
// - CreateTicket(c *gin.Context)
// - ListTickets(c *gin.Context)
// - GetTicket(c *gin.Context)
// - UpdateTicket(c *gin.Context)
// - DeleteTicket(c *gin.Context)
//
// Status Management:
// - AssignTicket(c *gin.Context)
// - EscalateTicket(c *gin.Context)
// - CloseTicket(c *gin.Context)
// - ReopenTicket(c *gin.Context)
// - GetAgents(c *gin.Context)
//
// Message Management:
// - GetTicketMessages(c *gin.Context)
// - AddMessage(c *gin.Context)
// - GetMessage(c *gin.Context)
// - UpdateMessage(c *gin.Context)
// - DeleteMessage(c *gin.Context)
// - AddInternalNote(c *gin.Context)
//
// Search & Analytics:
// - SearchTickets(c *gin.Context)
// - GetStatistics(c *gin.Context)
// - GetAnalytics(c *gin.Context)
//
// Bulk Operations:
// - BulkAssignTickets(c *gin.Context)
// - BulkUpdateStatus(c *gin.Context)
// - BulkCloseTickets(c *gin.Context)
