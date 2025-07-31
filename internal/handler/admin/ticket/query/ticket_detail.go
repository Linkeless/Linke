package query

import (
	"linke/internal/handler/admin/ticket/shared"
	"linke/internal/logger"
	"linke/internal/response"
	"linke/internal/service"

	"github.com/gin-gonic/gin"
)

// TicketDetailHandler handles ticket detail operations
type TicketDetailHandler struct {
	*shared.BaseHandler
}

// NewTicketDetailHandler creates a new ticket detail handler
func NewTicketDetailHandler(ticketService *service.TicketService, ticketMessageService *service.TicketMessageService) *TicketDetailHandler {
	return &TicketDetailHandler{
		BaseHandler: shared.NewBaseHandler(ticketService, ticketMessageService),
	}
}

// GetTicket gets a specific ticket by ID (admin view)
// @Summary Get a specific ticket (Admin)
// @Description Get details of a specific ticket by ID with full admin access
// @Tags Admin-Ticket-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Ticket ID"
// @Success 200 {object} response.Response{data=model.TicketResponse} "Ticket retrieved successfully"
// @Failure 400 {object} response.Response "Invalid ticket ID"
// @Failure 401 {object} response.Response "Unauthorized"
// @Failure 403 {object} response.Response "Access denied"
// @Failure 404 {object} response.Response "Ticket not found"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /admin/tickets/{id} [get]
func (h *TicketDetailHandler) GetTicket(c *gin.Context) {
	// Validate ticket ID
	ticketID, err := h.Validator.ValidateTicketID(c)
	if err != nil {
		return // Response already sent by validator
	}

	// Get ticket
	ticket, err := h.TicketService.GetTicket(c.Request.Context(), ticketID)
	if err != nil {
		logger.Error("Failed to get ticket", logger.Error2("error", err))
		if err.Error() == "ticket not found" {
			response.NotFound(c, "Ticket not found")
			return
		}
		response.InternalServerError(c, "Failed to get ticket", err.Error())
		return
	}

	// Convert to admin response format
	adminResponse := ticket.ToResponse()
	response.OK(c, "Ticket retrieved successfully", adminResponse)
}

// GetTicketByNumber gets a specific ticket by ticket number (admin)
// @Summary Get a ticket by ticket number (Admin)
// @Description Get details of a specific ticket by ticket number with full admin access
// @Tags Admin-Ticket-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param ticket_no path string true "Ticket Number"
// @Success 200 {object} response.Response{data=model.TicketResponse} "Ticket retrieved successfully"
// @Failure 400 {object} response.Response "Invalid ticket number"
// @Failure 401 {object} response.Response "Unauthorized"
// @Failure 403 {object} response.Response "Access denied"
// @Failure 404 {object} response.Response "Ticket not found"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /admin/tickets/number/{ticket_no} [get]
func (h *TicketDetailHandler) GetTicketByNumber(c *gin.Context) {
	// Get ticket number from path
	ticketNo := c.Param("ticket_no")
	
	// Validate ticket number
	if err := h.Validator.ValidateTicketNumber(ticketNo); err != nil {
		logger.Error("Invalid ticket number", logger.Error2("error", err))
		response.BadRequest(c, "Invalid ticket number", err.Error())
		return
	}

	// Get ticket
	ticket, err := h.TicketService.GetTicketByNumber(c.Request.Context(), ticketNo)
	if err != nil {
		logger.Error("Failed to get ticket", logger.Error2("error", err))
		if err.Error() == "ticket not found" {
			response.NotFound(c, "Ticket not found")
			return
		}
		response.InternalServerError(c, "Failed to get ticket", err.Error())
		return
	}

	// Convert to admin response format
	adminResponse := ticket.ToResponse()
	response.OK(c, "Ticket retrieved successfully", adminResponse)
}