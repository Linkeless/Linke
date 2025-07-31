package shared

import (
	"linke/internal/service"
)

// BaseHandler provides common dependencies for ticket handlers
type BaseHandler struct {
	TicketService        *service.TicketService
	TicketMessageService *service.TicketMessageService
	Validator            *TicketValidator
}

// NewBaseHandler creates a new base handler with common dependencies
func NewBaseHandler(ticketService *service.TicketService, ticketMessageService *service.TicketMessageService) *BaseHandler {
	return &BaseHandler{
		TicketService:        ticketService,
		TicketMessageService: ticketMessageService,
		Validator:            NewTicketValidator(),
	}
}