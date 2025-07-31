package management

import (
	"linke/internal/handler/admin/ticket/shared"
	"linke/internal/logger"
	"linke/internal/response"
	"linke/internal/service"

	"github.com/gin-gonic/gin"
)

// TicketCRUDHandler handles ticket CRUD operations
type TicketCRUDHandler struct {
	*shared.BaseHandler
}

// NewTicketCRUDHandler creates a new ticket CRUD handler
func NewTicketCRUDHandler(ticketService *service.TicketService, ticketMessageService *service.TicketMessageService) *TicketCRUDHandler {
	return &TicketCRUDHandler{
		BaseHandler: shared.NewBaseHandler(ticketService, ticketMessageService),
	}
}

// UpdateTicket updates a ticket
// @Summary Update a ticket (Admin)
// @Description Update ticket details (admin only)
// @Tags Admin-Ticket-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Ticket ID"
// @Param request body service.UpdateTicketRequest true "Ticket update request"
// @Success 200 {object} response.Response{data=model.TicketResponse} "Ticket updated successfully"
// @Failure 400 {object} response.Response "Invalid request"
// @Failure 401 {object} response.Response "Unauthorized"
// @Failure 403 {object} response.Response "Access denied"
// @Failure 404 {object} response.Response "Ticket not found"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /admin/tickets/{id} [put]
func (h *TicketCRUDHandler) UpdateTicket(c *gin.Context) {
	// Validate ticket ID
	ticketID, err := h.Validator.ValidateTicketID(c)
	if err != nil {
		return // Response already sent by validator
	}

	var req service.UpdateTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Failed to bind update ticket request", logger.Error2("error", err))
		response.BadRequest(c, "Invalid request format", err.Error())
		return
	}

	// Validate request fields if provided
	if req.Status != nil {
		if err := h.Validator.ValidateTicketStatus(*req.Status); err != nil {
			response.BadRequest(c, "Invalid status", err.Error())
			return
		}
	}

	if req.Priority != nil {
		if err := h.Validator.ValidateTicketPriority(*req.Priority); err != nil {
			response.BadRequest(c, "Invalid priority", err.Error())
			return
		}
	}

	if req.Category != nil {
		if err := h.Validator.ValidateTicketCategory(*req.Category); err != nil {
			response.BadRequest(c, "Invalid category", err.Error())
			return
		}
	}

	// Update ticket
	ticket, err := h.TicketService.UpdateTicket(c.Request.Context(), ticketID, &req)
	if err != nil {
		logger.Error("Failed to update ticket", logger.Error2("error", err))
		if err.Error() == "ticket not found" {
			response.NotFound(c, "Ticket not found")
			return
		}
		response.InternalServerError(c, "Failed to update ticket", err.Error())
		return
	}

	// Convert to admin response format
	adminResponse := ticket.ToResponse()
	response.OK(c, "Ticket updated successfully", adminResponse)
}