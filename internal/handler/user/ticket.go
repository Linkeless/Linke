package user

import (
	"strconv"

	"linke/internal/logger"
	"linke/internal/middleware"
	"linke/internal/model"
	"linke/internal/response"
	"linke/internal/service"

	"github.com/gin-gonic/gin"
)

// UserTicketHandler handles user-facing ticket operations
type UserTicketHandler struct {
	ticketService        *service.TicketService
	ticketMessageService *service.TicketMessageService
}

// NewUserTicketHandler creates a new UserTicketHandler
func NewUserTicketHandler(ticketService *service.TicketService, ticketMessageService *service.TicketMessageService) *UserTicketHandler {
	return &UserTicketHandler{
		ticketService:        ticketService,
		ticketMessageService: ticketMessageService,
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
func (h *UserTicketHandler) CreateTicket(c *gin.Context) {
	var req service.CreateTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Failed to bind create ticket request", logger.Error2("error", err))
		response.BadRequest(c, "Invalid request format", err.Error())
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

	// Create ticket
	ticket, err := h.ticketService.CreateTicket(c.Request.Context(), currentUser.ID, &req)
	if err != nil {
		logger.Error("Failed to create ticket", logger.Error2("error", err))
		response.InternalServerError(c, "Failed to create ticket", err.Error())
		return
	}

	// Convert to user response format
	userResponse := ticket.ToUserResponse()
	response.CreatedWithMessage(c, "Ticket created successfully", userResponse)
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
func (h *UserTicketHandler) GetMyTickets(c *gin.Context) {
	var req service.GetTicketsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		logger.Error("Failed to bind get tickets request", logger.Error2("error", err))
		response.BadRequest(c, "Invalid request format", err.Error())
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

	// Set user ID filter to current user
	req.UserID = currentUser.ID

	// Set default limit
	if req.Limit == 0 {
		req.Limit = 10
	}

	// Get tickets
	tickets, total, err := h.ticketService.GetTickets(c.Request.Context(), &req)
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
func (h *UserTicketHandler) GetTicket(c *gin.Context) {
	// Parse ticket ID
	ticketIDStr := c.Param("id")
	ticketID, err := strconv.ParseUint(ticketIDStr, 10, 32)
	if err != nil {
		logger.Error("Invalid ticket ID", logger.Error2("error", err))
		response.BadRequest(c, "Invalid ticket ID", err.Error())
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

	// Get ticket
	ticket, err := h.ticketService.GetTicket(c.Request.Context(), uint(ticketID))
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
	if ticket.UserID != currentUser.ID {
		logger.Error("Access denied to ticket", logger.Uint("ticket_id", uint(ticketID)), logger.Uint("user_id", currentUser.ID))
		response.Forbidden(c, "Access denied: you can only view your own tickets")
		return
	}

	// Convert to user response format
	userResponse := ticket.ToUserResponse()
	response.OK(c, "Ticket retrieved successfully", userResponse)
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
func (h *UserTicketHandler) AddTicketMessage(c *gin.Context) {
	// Parse ticket ID
	ticketIDStr := c.Param("id")
	ticketID, err := strconv.ParseUint(ticketIDStr, 10, 32)
	if err != nil {
		logger.Error("Invalid ticket ID", logger.Error2("error", err))
		response.BadRequest(c, "Invalid ticket ID", err.Error())
		return
	}

	var req service.CreateTicketMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Failed to bind create ticket message request", logger.Error2("error", err))
		response.BadRequest(c, "Invalid request format", err.Error())
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

	// First, verify the ticket exists and user owns it
	ticket, err := h.ticketService.GetTicket(c.Request.Context(), uint(ticketID))
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
	if ticket.UserID != currentUser.ID {
		logger.Error("Access denied to ticket", logger.Uint("ticket_id", uint(ticketID)), logger.Uint("user_id", currentUser.ID))
		response.Forbidden(c, "Access denied: you can only add messages to your own tickets")
		return
	}

	// Force message type to 'user' for user-submitted messages
	req.MessageType = "user"
	req.IsInternal = false // Users cannot create internal messages

	// Create message
	message, err := h.ticketMessageService.CreateTicketMessage(c.Request.Context(), uint(ticketID), currentUser.ID, &req)
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
func (h *UserTicketHandler) GetTicketMessages(c *gin.Context) {
	// Parse ticket ID
	ticketIDStr := c.Param("id")
	ticketID, err := strconv.ParseUint(ticketIDStr, 10, 32)
	if err != nil {
		logger.Error("Invalid ticket ID", logger.Error2("error", err))
		response.BadRequest(c, "Invalid ticket ID", err.Error())
		return
	}

	// Get query parameters
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 10
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	// Get current user from context
	userValue, exists := c.Get(middleware.AuthContextKey)
	if !exists {
		logger.Error("User not found in context")
		response.Unauthorized(c, "User not found")
		return
	}

	currentUser := userValue.(*model.User)

	// Get messages (this method includes access control)
	messages, total, err := h.ticketMessageService.GetTicketMessagesForUser(c.Request.Context(), uint(ticketID), currentUser.ID, limit, offset)
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
func (h *UserTicketHandler) GetTicketByNumber(c *gin.Context) {
	// Get ticket number from path
	ticketNo := c.Param("ticket_no")
	if ticketNo == "" {
		logger.Error("Empty ticket number")
		response.BadRequest(c, "Ticket number is required")
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

	// Get ticket
	ticket, err := h.ticketService.GetTicketByNumber(c.Request.Context(), ticketNo)
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
	if ticket.UserID != currentUser.ID {
		logger.Error("Access denied to ticket", logger.String("ticket_no", ticketNo), logger.Uint("user_id", currentUser.ID))
		response.Forbidden(c, "Access denied: you can only view your own tickets")
		return
	}

	// Convert to user response format
	userResponse := ticket.ToUserResponse()
	response.OK(c, "Ticket retrieved successfully", userResponse)
}

// SetupUserTicketRoutes sets up routes for user ticket operations
func (h *UserTicketHandler) SetupUserTicketRoutes(userGroup *gin.RouterGroup) {
	// Note: Authentication middleware is already applied by parent group
	// userGroup.Use(middleware.AuthMiddleware(authService))

	// Ticket routes
	ticketGroup := userGroup.Group("/tickets")
	{
		ticketGroup.POST("", h.CreateTicket)
		ticketGroup.GET("", h.GetMyTickets)
		ticketGroup.GET("/:id", h.GetTicket)
		ticketGroup.GET("/number/:ticket_no", h.GetTicketByNumber)
		ticketGroup.POST("/:id/messages", h.AddTicketMessage)
		ticketGroup.GET("/:id/messages", h.GetTicketMessages)
	}
}