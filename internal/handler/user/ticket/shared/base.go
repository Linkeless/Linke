package ticket

import (
	"linke/internal/service"
)

// BaseTicketHandler provides common dependencies for all ticket handlers
type BaseTicketHandler struct {
	TicketService        *service.TicketService
	TicketMessageService *service.TicketMessageService
}

// NewBaseTicketHandler creates a new base ticket handler
func NewBaseTicketHandler(ticketService *service.TicketService, ticketMessageService *service.TicketMessageService) *BaseTicketHandler {
	return &BaseTicketHandler{
		TicketService:        ticketService,
		TicketMessageService: ticketMessageService,
	}
}