package ticket

import (
	ticketshared "linke/internal/handler/user/ticket/shared"
	"linke/internal/logger"
	"linke/internal/response"
	"linke/internal/service"

	"github.com/gin-gonic/gin"
)

// TicketManagementHandler handles ticket CRUD operations
type TicketManagementHandler struct {
	*ticketshared.BaseTicketHandler
	validator *ticketshared.TicketValidator
}

// NewTicketManagementHandler creates a new ticket management handler
func NewTicketManagementHandler(ticketService *service.TicketService, ticketMessageService *service.TicketMessageService) *TicketManagementHandler {
	return &TicketManagementHandler{
		BaseTicketHandler: ticketshared.NewBaseTicketHandler(ticketService, ticketMessageService),
		validator:         ticketshared.NewTicketValidator(),
	}
}

// CreateTicket creates a new ticket
// @Summary Create a new support ticket
// @Description Creates a new support ticket for the authenticated user
// @Tags User-Ticket
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body service.CreateTicketRequest true "Ticket creation request"
// @Success 201 {object} response.Response{data=model.TicketUserResponse} "Ticket created successfully"
// @Failure 400 {object} response.Response "Invalid request"
// @Failure 401 {object} response.Response "Unauthorized"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /user/tickets [post]
func (h *TicketManagementHandler) CreateTicket(c *gin.Context) {
	var req service.CreateTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Failed to bind create ticket request", logger.Error2("error", err))
		response.BadRequest(c, "Invalid request format", err.Error())
		return
	}

	// Get current user from context
	currentUser, valid := h.validator.GetUserFromContext(c)
	if !valid {
		return
	}

	// Create ticket
	ticket, err := h.TicketService.CreateTicket(c.Request.Context(), currentUser.ID, &req)
	if err != nil {
		logger.Error("Failed to create ticket", logger.Error2("error", err))
		response.InternalServerError(c, "Failed to create ticket", err.Error())
		return
	}

	// Convert to user response format
	userResponse := ticket.ToUserResponse()
	response.CreatedWithMessage(c, "Ticket created successfully", userResponse)
}

// GetTicket gets a specific ticket by ID
// @Summary Get a specific ticket
// @Description Get details of a specific ticket by ID (only accessible to the ticket owner)
// @Tags User-Ticket
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Ticket ID"
// @Success 200 {object} response.Response{data=model.TicketUserResponse} "Ticket retrieved successfully"
// @Failure 400 {object} response.Response "Invalid ticket ID"
// @Failure 401 {object} response.Response "Unauthorized"
// @Failure 403 {object} response.Response "Access denied"
// @Failure 404 {object} response.Response "Ticket not found"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /user/tickets/{id} [get]
func (h *TicketManagementHandler) GetTicket(c *gin.Context) {
	// Validate ticket ID parameter
	ticketID, valid := h.validator.ValidateTicketIDParam(c)
	if !valid {
		return
	}

	// Get current user from context
	currentUser, valid := h.validator.GetUserFromContext(c)
	if !valid {
		return
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

	// Check if user owns the ticket
	if !h.validator.CheckTicketOwnership(currentUser, ticket, c) {
		return
	}

	// Convert to user response format
	userResponse := ticket.ToUserResponse()
	response.OK(c, "Ticket retrieved successfully", userResponse)
}