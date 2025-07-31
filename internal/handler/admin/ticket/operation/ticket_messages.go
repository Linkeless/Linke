package operation

import (
	"linke/internal/handler/admin/ticket/shared"
	"linke/internal/logger"
	"linke/internal/middleware"
	"linke/internal/model"
	"linke/internal/response"
	"linke/internal/service"

	"github.com/gin-gonic/gin"
)

// TicketMessageHandler handles ticket message operations
type TicketMessageHandler struct {
	*shared.BaseHandler
}

// NewTicketMessageHandler creates a new ticket message handler
func NewTicketMessageHandler(ticketService *service.TicketService, ticketMessageService *service.TicketMessageService) *TicketMessageHandler {
	return &TicketMessageHandler{
		BaseHandler: shared.NewBaseHandler(ticketService, ticketMessageService),
	}
}

// AddTicketMessage adds a message to a ticket (admin)
// @Summary Add a message to a ticket (Admin)
// @Description Add a message to a ticket with admin privileges (can create internal messages)
// @Tags Admin-Ticket-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Ticket ID"
// @Param request body service.CreateTicketMessageRequest true "Message creation request"
// @Success 201 {object} response.Response{data=model.TicketMessageResponse} "Message added successfully"
// @Failure 400 {object} response.Response "Invalid request"
// @Failure 401 {object} response.Response "Unauthorized"
// @Failure 403 {object} response.Response "Access denied"
// @Failure 404 {object} response.Response "Ticket not found"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /admin/tickets/{id}/messages [post]
func (h *TicketMessageHandler) AddTicketMessage(c *gin.Context) {
	// Validate ticket ID
	ticketID, err := h.Validator.ValidateTicketID(c)
	if err != nil {
		return // Response already sent by validator
	}

	var req service.CreateTicketMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Failed to bind create ticket message request", logger.Error2("error", err))
		response.BadRequest(c, "Invalid request format", err.Error())
		return
	}

	// Validate message content
	if err := h.Validator.ValidateTicketMessage(req.Content); err != nil {
		response.BadRequest(c, "Invalid message content", err.Error())
		return
	}

	// Validate message type if provided
	if req.MessageType != "" {
		if err := h.Validator.ValidateMessageType(req.MessageType); err != nil {
			response.BadRequest(c, "Invalid message type", err.Error())
			return
		}
	}

	// Get current user from context
	userValue, exists := c.Get(middleware.AuthContextKey)
	if !exists {
		logger.Error("User not found in context")
		response.Unauthorized(c, "User not found")
		return
	}

	currentUser := userValue.(*model.User)

	// Set message type to 'admin' if not specified
	if req.MessageType == "" {
		req.MessageType = "admin"
	}

	// Create message
	message, err := h.TicketMessageService.CreateTicketMessage(c.Request.Context(), ticketID, currentUser.ID, &req)
	if err != nil {
		logger.Error("Failed to create ticket message", logger.Error2("error", err))
		if err.Error() == "ticket not found" {
			response.NotFound(c, "Ticket not found")
			return
		}
		response.InternalServerError(c, "Failed to create message", err.Error())
		return
	}

	// Convert to admin response format
	adminResponse := message.ToResponse()
	response.CreatedWithMessage(c, "Message added successfully", adminResponse)
}

// GetTicketMessages gets all messages for a ticket (admin)
// @Summary Get messages for a ticket (Admin)
// @Description Get all messages for a specific ticket including internal messages
// @Tags Admin-Ticket-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Ticket ID"
// @Param message_type query string false "Filter by message type" Enums(user,admin,system)
// @Param include_internal query bool false "Include internal messages" default(true)
// @Param limit query int false "Limit number of results" minimum(1) maximum(100) default(10)
// @Param offset query int false "Offset for pagination" minimum(0) default(0)
// @Success 200 {object} response.Response{data=response.PaginatedResponse{items=[]model.TicketMessageResponse}} "Messages retrieved successfully"
// @Failure 400 {object} response.Response "Invalid request"
// @Failure 401 {object} response.Response "Unauthorized"
// @Failure 403 {object} response.Response "Access denied"
// @Failure 404 {object} response.Response "Ticket not found"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /admin/tickets/{id}/messages [get]
func (h *TicketMessageHandler) GetTicketMessages(c *gin.Context) {
	// Validate ticket ID
	ticketID, err := h.Validator.ValidateTicketID(c)
	if err != nil {
		return // Response already sent by validator
	}

	var req service.GetTicketMessagesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		logger.Error("Failed to bind get ticket messages request", logger.Error2("error", err))
		response.BadRequest(c, "Invalid request format", err.Error())
		return
	}

	// Set ticket ID from path
	req.TicketID = ticketID

	// Set defaults
	if req.Limit == 0 {
		req.Limit = 10
	}
	
	// Admin users can see internal messages by default
	if c.Query("include_internal") == "" {
		req.IncludeInternal = true
	}

	// Validate message type if provided
	if req.MessageType != "" {
		if err := h.Validator.ValidateMessageType(req.MessageType); err != nil {
			response.BadRequest(c, "Invalid message type parameter", err.Error())
			return
		}
	}

	// Get messages
	messages, total, err := h.TicketMessageService.GetTicketMessages(c.Request.Context(), &req)
	if err != nil {
		logger.Error("Failed to get ticket messages", logger.Error2("error", err))
		response.InternalServerError(c, "Failed to get messages", err.Error())
		return
	}

	// Convert to admin response format
	adminMessages := make([]model.TicketMessageResponse, len(messages))
	for i, message := range messages {
		adminMessages[i] = *message.ToResponse()
	}

	// Send paginated response
	response.OKPaginated(c, "Messages retrieved successfully", adminMessages, total, req.Limit, req.Offset)
}