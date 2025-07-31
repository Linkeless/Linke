package status

import (
	"linke/internal/handler/admin/ticket/shared"
	"linke/internal/logger"
	"linke/internal/middleware"
	"linke/internal/model"
	"linke/internal/response"
	"linke/internal/service"

	"github.com/gin-gonic/gin"
)

// TicketStatusHandler handles ticket status management operations
type TicketStatusHandler struct {
	*shared.BaseHandler
}

// NewTicketStatusHandler creates a new ticket status handler
func NewTicketStatusHandler(ticketService *service.TicketService, ticketMessageService *service.TicketMessageService) *TicketStatusHandler {
	return &TicketStatusHandler{
		BaseHandler: shared.NewBaseHandler(ticketService, ticketMessageService),
	}
}

// AssignTicket assigns a ticket to an admin
// @Summary Assign a ticket to an admin (Admin)
// @Description Assign a ticket to a specific admin user
// @Tags Admin-Ticket-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Ticket ID"
// @Param request body service.AssignTicketRequest true "Assignment request"
// @Success 200 {object} response.Response{data=model.TicketResponse} "Ticket assigned successfully"
// @Failure 400 {object} response.Response "Invalid request"
// @Failure 401 {object} response.Response "Unauthorized"
// @Failure 403 {object} response.Response "Access denied"
// @Failure 404 {object} response.Response "Ticket not found"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /admin/tickets/{id}/assign [post]
func (h *TicketStatusHandler) AssignTicket(c *gin.Context) {
	// Validate ticket ID
	ticketID, err := h.Validator.ValidateTicketID(c)
	if err != nil {
		return // Response already sent by validator
	}

	var req service.AssignTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Failed to bind assign ticket request", logger.Error2("error", err))
		response.BadRequest(c, "Invalid request format", err.Error())
		return
	}

	// Validate assignment request
	if err := h.Validator.ValidateAssignmentRequest(req.AssignedToID); err != nil {
		response.BadRequest(c, "Invalid assignment request", err.Error())
		return
	}

	// Assign ticket
	ticket, err := h.TicketService.AssignTicket(c.Request.Context(), ticketID, &req)
	if err != nil {
		logger.Error("Failed to assign ticket", logger.Error2("error", err))
		if err.Error() == "ticket not found" {
			response.NotFound(c, "Ticket not found")
			return
		}
		if err.Error() == "assigned user not found" {
			response.BadRequest(c, "Assigned user not found")
			return
		}
		if err.Error() == "can only assign tickets to admin users" {
			response.BadRequest(c, "Can only assign tickets to admin users")
			return
		}
		response.InternalServerError(c, "Failed to assign ticket", err.Error())
		return
	}

	// Convert to admin response format
	adminResponse := ticket.ToResponse()
	response.OK(c, "Ticket assigned successfully", adminResponse)
}

// ResolveTicket resolves a ticket
// @Summary Resolve a ticket (Admin)
// @Description Resolve a ticket with a resolution message
// @Tags Admin-Ticket-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Ticket ID"
// @Param request body service.ResolveTicketRequest true "Resolution request"
// @Success 200 {object} response.Response{data=model.TicketResponse} "Ticket resolved successfully"
// @Failure 400 {object} response.Response "Invalid request"
// @Failure 401 {object} response.Response "Unauthorized"
// @Failure 403 {object} response.Response "Access denied"
// @Failure 404 {object} response.Response "Ticket not found"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /admin/tickets/{id}/resolve [post]
func (h *TicketStatusHandler) ResolveTicket(c *gin.Context) {
	// Validate ticket ID
	ticketID, err := h.Validator.ValidateTicketID(c)
	if err != nil {
		return // Response already sent by validator
	}

	var req service.ResolveTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Failed to bind resolve ticket request", logger.Error2("error", err))
		response.BadRequest(c, "Invalid request format", err.Error())
		return
	}

	// Validate resolution message
	if err := h.Validator.ValidateResolutionMessage(req.Resolution); err != nil {
		response.BadRequest(c, "Invalid resolution message", err.Error())
		return
	}

	// Get current user from context
	userValue, exists := c.Get(middleware.AuthContextKey)
	if !exists {
		logger.Error("User not found in context")
		response.Unauthorized(c, "User not found")
		return
	}

	currentUser := userValue.(*model.User)

	// Resolve ticket
	ticket, err := h.TicketService.ResolveTicket(c.Request.Context(), ticketID, currentUser.ID, &req)
	if err != nil {
		logger.Error("Failed to resolve ticket", logger.Error2("error", err))
		if err.Error() == "ticket not found" {
			response.NotFound(c, "Ticket not found")
			return
		}
		response.InternalServerError(c, "Failed to resolve ticket", err.Error())
		return
	}

	// Convert to admin response format
	adminResponse := ticket.ToResponse()
	response.OK(c, "Ticket resolved successfully", adminResponse)
}

// CloseTicket closes a ticket
// @Summary Close a ticket (Admin)
// @Description Close a ticket (typically after resolution)
// @Tags Admin-Ticket-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Ticket ID"
// @Success 200 {object} response.Response{data=model.TicketResponse} "Ticket closed successfully"
// @Failure 400 {object} response.Response "Invalid ticket ID"
// @Failure 401 {object} response.Response "Unauthorized"
// @Failure 403 {object} response.Response "Access denied"
// @Failure 404 {object} response.Response "Ticket not found"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /admin/tickets/{id}/close [post]
func (h *TicketStatusHandler) CloseTicket(c *gin.Context) {
	// Validate ticket ID
	ticketID, err := h.Validator.ValidateTicketID(c)
	if err != nil {
		return // Response already sent by validator
	}

	// Close ticket
	ticket, err := h.TicketService.CloseTicket(c.Request.Context(), ticketID)
	if err != nil {
		logger.Error("Failed to close ticket", logger.Error2("error", err))
		if err.Error() == "ticket not found" {
			response.NotFound(c, "Ticket not found")
			return
		}
		response.InternalServerError(c, "Failed to close ticket", err.Error())
		return
	}

	// Convert to admin response format
	adminResponse := ticket.ToResponse()
	response.OK(c, "Ticket closed successfully", adminResponse)
}