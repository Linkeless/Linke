package ticket

import (
	ticketshared "linke/internal/handler/user/ticket/shared"
	"linke/internal/logger"
	"linke/internal/model"
	"linke/internal/response"
	"linke/internal/service"

	"github.com/gin-gonic/gin"
)

// TicketQueryHandler handles ticket query operations
type TicketQueryHandler struct {
	*ticketshared.BaseTicketHandler
	validator *ticketshared.TicketValidator
}

// NewTicketQueryHandler creates a new ticket query handler
func NewTicketQueryHandler(ticketService *service.TicketService, ticketMessageService *service.TicketMessageService) *TicketQueryHandler {
	return &TicketQueryHandler{
		BaseTicketHandler: ticketshared.NewBaseTicketHandler(ticketService, ticketMessageService),
		validator:         ticketshared.NewTicketValidator(),
	}
}

// GetMyTickets gets tickets for the current user
// @Summary Get user's tickets
// @Description Get all tickets created by the authenticated user
// @Tags User-Ticket
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param status query string false "Filter by status" Enums(open,in_progress,pending,resolved,closed)
// @Param category query string false "Filter by category" Enums(general,technical,billing,account,feature,bug,subscription,payment)
// @Param priority query string false "Filter by priority" Enums(low,normal,high,urgent,critical)
// @Param search query string false "Search in title, description, or ticket number"
// @Param limit query int false "Limit number of results" minimum(1) maximum(100) default(10)
// @Param offset query int false "Offset for pagination" minimum(0) default(0)
// @Success 200 {object} response.Response{data=response.PaginatedResponse{items=[]model.TicketUserResponse}} "Tickets retrieved successfully"
// @Failure 401 {object} response.Response "Unauthorized"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /user/tickets [get]
func (h *TicketQueryHandler) GetMyTickets(c *gin.Context) {
	var req service.GetTicketsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		logger.Error("Failed to bind get tickets request", logger.Error2("error", err))
		response.BadRequest(c, "Invalid request format", err.Error())
		return
	}

	// Get current user from context
	currentUser, valid := h.validator.GetUserFromContext(c)
	if !valid {
		return
	}

	// Set user ID filter to current user
	req.UserID = currentUser.ID

	// Set default limit
	if req.Limit == 0 {
		req.Limit = 10
	}

	// Get tickets
	tickets, total, err := h.TicketService.GetTickets(c.Request.Context(), &req)
	if err != nil {
		logger.Error("Failed to get tickets", logger.Error2("error", err))
		response.InternalServerError(c, "Failed to get tickets", err.Error())
		return
	}

	// Convert to user response format
	userTickets := make([]model.TicketUserResponse, len(tickets))
	for i, ticket := range tickets {
		userTickets[i] = *ticket.ToUserResponse()
	}

	// Send paginated response
	response.OKPaginated(c, "Tickets retrieved successfully", userTickets, total, req.Limit, req.Offset)
}

// GetTicketByNumber gets a specific ticket by ticket number
// @Summary Get a ticket by ticket number
// @Description Get details of a specific ticket by ticket number (only accessible to the ticket owner)
// @Tags User-Ticket
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param ticket_no path string true "Ticket Number"
// @Success 200 {object} response.Response{data=model.TicketUserResponse} "Ticket retrieved successfully"
// @Failure 400 {object} response.Response "Invalid ticket number"
// @Failure 401 {object} response.Response "Unauthorized"
// @Failure 403 {object} response.Response "Access denied"
// @Failure 404 {object} response.Response "Ticket not found"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /user/tickets/number/{ticket_no} [get]
func (h *TicketQueryHandler) GetTicketByNumber(c *gin.Context) {
	// Get ticket number from path
	ticketNo, valid := h.validator.ValidateTicketNumberParam(c)
	if !valid {
		return
	}

	// Get current user from context
	currentUser, valid := h.validator.GetUserFromContext(c)
	if !valid {
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

	// Check if user owns the ticket
	if !h.validator.CheckTicketOwnership(currentUser, ticket, c) {
		return
	}

	// Convert to user response format
	userResponse := ticket.ToUserResponse()
	response.OK(c, "Ticket retrieved successfully", userResponse)
}