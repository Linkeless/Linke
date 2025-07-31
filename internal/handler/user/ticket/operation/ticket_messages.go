package ticket

import (
	ticketshared "linke/internal/handler/user/ticket/shared"
	"linke/internal/logger"
	"linke/internal/model"
	"linke/internal/response"
	"linke/internal/service"

	"github.com/gin-gonic/gin"
)

// TicketMessageHandler handles ticket message operations
type TicketMessageHandler struct {
	*ticketshared.BaseTicketHandler
	validator *ticketshared.TicketValidator
}

// NewTicketMessageHandler creates a new ticket message handler
func NewTicketMessageHandler(ticketService *service.TicketService, ticketMessageService *service.TicketMessageService) *TicketMessageHandler {
	return &TicketMessageHandler{
		BaseTicketHandler: ticketshared.NewBaseTicketHandler(ticketService, ticketMessageService),
		validator:         ticketshared.NewTicketValidator(),
	}
}

// AddTicketMessage adds a message to a ticket
// @Summary Add a message to a ticket
// @Description Add a message to a ticket (only accessible to the ticket owner)
// @Tags User-Ticket
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Ticket ID"
// @Param request body service.CreateTicketMessageRequest true "Message creation request"
// @Success 201 {object} response.Response{data=model.TicketMessageUserResponse} "Message added successfully"
// @Failure 400 {object} response.Response "Invalid request"
// @Failure 401 {object} response.Response "Unauthorized"
// @Failure 403 {object} response.Response "Access denied"
// @Failure 404 {object} response.Response "Ticket not found"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /user/tickets/{id}/messages [post]
func (h *TicketMessageHandler) AddTicketMessage(c *gin.Context) {
	// Validate ticket ID parameter
	ticketID, valid := h.validator.ValidateTicketIDParam(c)
	if !valid {
		return
	}

	var req service.CreateTicketMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Failed to bind create ticket message request", logger.Error2("error", err))
		response.BadRequest(c, "Invalid request format", err.Error())
		return
	}

	// Get current user from context
	currentUser, valid := h.validator.GetUserFromContext(c)
	if !valid {
		return
	}

	// First, verify the ticket exists and user owns it
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

	// Force message type to 'user' for user-submitted messages
	req.MessageType = "user"
	req.IsInternal = false // Users cannot create internal messages

	// Create message
	message, err := h.TicketMessageService.CreateTicketMessage(c.Request.Context(), ticketID, currentUser.ID, &req)
	if err != nil {
		logger.Error("Failed to create ticket message", logger.Error2("error", err))
		response.InternalServerError(c, "Failed to create message", err.Error())
		return
	}

	// Convert to user response format
	userResponse := message.ToUserResponse()
	response.CreatedWithMessage(c, "Message added successfully", userResponse)
}

// GetTicketMessages gets messages for a specific ticket
// @Summary Get messages for a ticket
// @Description Get all messages for a specific ticket (only accessible to the ticket owner)
// @Tags User-Ticket
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Ticket ID"
// @Param limit query int false "Limit number of results" minimum(1) maximum(100) default(10)
// @Param offset query int false "Offset for pagination" minimum(0) default(0)
// @Success 200 {object} response.Response{data=response.PaginatedResponse{items=[]model.TicketMessageUserResponse}} "Messages retrieved successfully"
// @Failure 400 {object} response.Response "Invalid request"
// @Failure 401 {object} response.Response "Unauthorized"
// @Failure 403 {object} response.Response "Access denied"
// @Failure 404 {object} response.Response "Ticket not found"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /user/tickets/{id}/messages [get]
func (h *TicketMessageHandler) GetTicketMessages(c *gin.Context) {
	// Validate ticket ID parameter
	ticketID, valid := h.validator.ValidateTicketIDParam(c)
	if !valid {
		return
	}

	// Validate pagination parameters
	limit, offset, valid := h.validator.ValidatePaginationParams(c)
	if !valid {
		return
	}

	// Get current user from context
	currentUser, valid := h.validator.GetUserFromContext(c)
	if !valid {
		return
	}

	// Get messages (this method includes access control)
	messages, total, err := h.TicketMessageService.GetTicketMessagesForUser(c.Request.Context(), ticketID, currentUser.ID, limit, offset)
	if err != nil {
		logger.Error("Failed to get ticket messages", logger.Error2("error", err))
		if err.Error() == "ticket not found" {
			response.NotFound(c, "Ticket not found")
			return
		}
		if err.Error() == "access denied: you can only view messages for your own tickets" {
			response.Forbidden(c, "Access denied: you can only view messages for your own tickets")
			return
		}
		response.InternalServerError(c, "Failed to get messages", err.Error())
		return
	}

	// Convert to user response format
	userMessages := make([]model.TicketMessageUserResponse, len(messages))
	for i, message := range messages {
		userMessages[i] = *message.ToUserResponse()
	}

	// Send paginated response
	response.OKPaginated(c, "Messages retrieved successfully", userMessages, total, limit, offset)
}