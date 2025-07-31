package operation

import (
	"linke/internal/handler/admin/ticket/shared"
	"linke/internal/logger"
	"linke/internal/response"
	"linke/internal/service"

	"github.com/gin-gonic/gin"
)

// TicketActionHandler handles ticket action operations
type TicketActionHandler struct {
	*shared.BaseHandler
}

// NewTicketActionHandler creates a new ticket action handler
func NewTicketActionHandler(ticketService *service.TicketService, ticketMessageService *service.TicketMessageService) *TicketActionHandler {
	return &TicketActionHandler{
		BaseHandler: shared.NewBaseHandler(ticketService, ticketMessageService),
	}
}

// DeleteTicket deletes a ticket
// @Summary Delete a ticket (Admin)
// @Description Soft delete a ticket (admin only)
// @Tags Admin-Ticket-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Ticket ID"
// @Success 200 {object} response.Response "Ticket deleted successfully"
// @Failure 400 {object} response.Response "Invalid ticket ID"
// @Failure 401 {object} response.Response "Unauthorized"
// @Failure 403 {object} response.Response "Access denied"
// @Failure 404 {object} response.Response "Ticket not found"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /admin/tickets/{id} [delete]
func (h *TicketActionHandler) DeleteTicket(c *gin.Context) {
	// Validate ticket ID
	ticketID, err := h.Validator.ValidateTicketID(c)
	if err != nil {
		return // Response already sent by validator
	}

	// Delete ticket
	if err := h.TicketService.DeleteTicket(c.Request.Context(), ticketID); err != nil {
		logger.Error("Failed to delete ticket", logger.Error2("error", err))
		if err.Error() == "ticket not found" {
			response.NotFound(c, "Ticket not found")
			return
		}
		response.InternalServerError(c, "Failed to delete ticket", err.Error())
		return
	}

	response.OK(c, "Ticket deleted successfully", nil)
}