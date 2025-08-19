package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"linke/internal/domains/ticket/constants"
	"linke/internal/domains/ticket/dto"
	ticketInterfaces "linke/internal/domains/ticket/usecases/interfaces"
	sharedErrors "linke/internal/shared/errors"
	"linke/internal/shared/handlers"
	"linke/internal/shared/logger"
	"linke/internal/shared/response"
)

// UserTicketHandler provides user ticket management functionality
type UserTicketHandler struct {
	ticketService        ticketInterfaces.TicketService
	ticketMessageService ticketInterfaces.TicketMessageService
}

// NewUserTicketHandler creates a new user ticket handler
func NewUserTicketHandler(
	ticketService ticketInterfaces.TicketService,
	ticketMessageService ticketInterfaces.TicketMessageService,
) *UserTicketHandler {
	return &UserTicketHandler{
		ticketService:        ticketService,
		ticketMessageService: ticketMessageService,
	}
}

// CreateTicket godoc
// @Summary Create new ticket
// @Description Create a new support ticket for the authenticated user
// @Tags User-Tickets
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param ticket body dto.UserCreateTicketRequest true "Ticket creation data"
// @Success 201 {object} dto.TicketUserResponse
// @Failure 400 {object} response.ProblemJSONResponse
// @Failure 401 {object} response.ProblemJSONResponse
// @Failure 500 {object} response.ProblemJSONResponse
// @Router /tickets [post]
func (h *UserTicketHandler) CreateTicket(c *gin.Context) {
	// Get current user from context
	user, ok := handlers.GetCurrentUser(c)
	if !ok {
		return // Error response already handled by GetCurrentUser
	}

	var req dto.UserCreateTicketRequest
	if !handlers.BindJSONRequest(c, &req) {
		return // Error response already handled by BindJSONRequest
	}

	// Create ticket request for service layer
	createReq := &dto.CreateTicketRequest{
		Title:       req.Title,
		Description: req.Description,
		Category:    req.Category,
		Priority:    req.Priority,
		Tags:        req.Tags,
		Metadata:    req.Metadata,
	}

	// Set default priority if not provided
	if createReq.Priority == "" {
		createReq.Priority = constants.TicketPriorityNormal
	}

	// Create ticket
	ticket, err := h.ticketService.CreateTicket(c.Request.Context(), user.ID, createReq)
	if err != nil {
		logger.Error("User failed to create ticket",
			logger.Uint("user_id", user.ID),
			logger.String("category", req.Category),
			logger.ErrorField(err))
		response.InternalServerError(c, "Failed to create ticket")
		return
	}

	logger.Info("User created ticket successfully",
		logger.Uint("ticket_id", ticket.ID),
		logger.String("ticket_no", ticket.TicketNo),
		logger.Uint("user_id", user.ID))

	// Convert to DTO and return user-appropriate response
	ticketResponse := dto.ToTicketUserResponse(ticket)
	response.Created(c, ticketResponse)
	// Return DTO to pool after use
	dto.PutTicketUserResponse(ticketResponse)
}

// GetMyTickets godoc
// @Summary Get user's tickets
// @Description Get paginated list of tickets created by the authenticated user
// @Tags User-Tickets
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param status query string false "Filter by status" Enums(open,in_progress,pending,resolved,closed) example(open)
// @Param priority query string false "Filter by priority" Enums(low,normal,high,urgent,critical) example(high)
// @Param category query string false "Filter by category" Enums(general,technical,billing,account,feature,bug,subscription,payment) example(subscription)
// @Param search query string false "Search in title or description" example("subscription issue")
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} response.HALCollectionResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /tickets/my [get]
func (h *UserTicketHandler) GetMyTickets(c *gin.Context) {
	// Get current user from context
	user, ok := handlers.GetCurrentUser(c)
	if !ok {
		return // Error response already handled by GetCurrentUser
	}

	// Parse pagination
	pagination := handlers.ParsePagination(c)

	// For page-based pagination, convert to offset
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * pagination.Limit

	// Create request for service layer with user-specific filtering
	req := &dto.GetTicketsRequest{
		UserID:   user.ID, // Only get tickets for this user
		Status:   c.Query("status"),
		Priority: c.Query("priority"),
		Category: c.Query("category"),
		Search:   c.Query("search"),
		Limit:    pagination.Limit,
		Offset:   offset,
	}

	tickets, total, err := h.ticketService.GetTickets(c.Request.Context(), req)
	if err != nil {
		logger.Error("User failed to get tickets",
			logger.Uint("user_id", user.ID),
			logger.ErrorField(err))
		response.InternalServerError(c, "Failed to get tickets")
		return
	}

	// Convert to user response format using DTO functions
	responses := make([]*dto.TicketUserResponse, len(tickets))
	for i, ticket := range tickets {
		responses[i] = dto.ToTicketUserResponse(ticket)
	}

	response.SendPaginatedResponse(c, responses, total)

	// Return DTOs to pool after use
	for _, resp := range responses {
		dto.PutTicketUserResponse(resp)
	}
}

// GetTicket godoc
// @Summary Get ticket details
// @Description Get detailed ticket information including messages (user can only access own tickets)
// @Tags User-Tickets
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path uint true "Ticket ID"
// @Success 200 {object} dto.TicketUserResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Router /tickets/{id} [get]
func (h *UserTicketHandler) GetTicket(c *gin.Context) {
	// Get current user from context
	user, ok := handlers.GetCurrentUser(c)
	if !ok {
		return // Error response already handled by GetCurrentUser
	}

	// Parse ticket ID
	ticketID, ok := handlers.ParseIDParam(c, "id")
	if !ok {
		return // Error response already handled by ParseIDParam
	}

	// Get ticket
	ticket, err := h.ticketService.GetTicket(c.Request.Context(), ticketID)
	if err != nil {
		logger.Error("User failed to get ticket",
			logger.Uint("ticket_id", ticketID),
			logger.Uint("user_id", user.ID),
			logger.ErrorField(err))
		response.NotFound(c, "Ticket not found")
		return
	}

	// Check ownership - users can only access their own tickets
	if ticket.UserID != user.ID {
		logger.Warn("User attempted to access ticket they don't own",
			logger.Uint("ticket_id", ticketID),
			logger.Uint("user_id", user.ID),
			logger.Uint("ticket_owner_id", ticket.UserID))
		response.Forbidden(c, "You can only access your own tickets")
		return
	}

	// Convert to DTO and return user-appropriate response (filters out internal messages and admin-only fields)
	ticketResponse := dto.ToTicketUserResponse(ticket)
	response.OK(c, ticketResponse)
	// Return DTO to pool after use
	dto.PutTicketUserResponse(ticketResponse)
}

// CloseTicket godoc
// @Summary Close ticket
// @Description Close a support ticket (user can only close own tickets)
// @Tags User-Tickets
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path uint true "Ticket ID"
// @Param closure body dto.CloseTicketRequest false "Closure reason"
// @Success 200 {object} dto.TicketUserResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /tickets/{id}/close [put]
func (h *UserTicketHandler) CloseTicket(c *gin.Context) {
	// Get current user from context
	user, ok := handlers.GetCurrentUser(c)
	if !ok {
		return // Error response already handled by GetCurrentUser
	}

	// Parse ticket ID
	ticketID, ok := handlers.ParseIDParam(c, "id")
	if !ok {
		return // Error response already handled by ParseIDParam
	}

	// Parse request body (optional)
	var req dto.CloseTicketRequest
	c.ShouldBindJSON(&req) // Don't fail if body is empty

	// Get ticket to verify ownership
	ticket, err := h.ticketService.GetTicket(c.Request.Context(), ticketID)
	if err != nil {
		logger.Error("User failed to get ticket for closure",
			logger.Uint("ticket_id", ticketID),
			logger.Uint("user_id", user.ID),
			logger.ErrorField(err))
		response.NotFound(c, "Ticket not found")
		return
	}

	// Check ownership - users can only close their own tickets
	if ticket.UserID != user.ID {
		logger.Warn("User attempted to close ticket they don't own",
			logger.Uint("ticket_id", ticketID),
			logger.Uint("user_id", user.ID),
			logger.Uint("ticket_owner_id", ticket.UserID))
		response.Forbidden(c, "You can only close your own tickets")
		return
	}

	// Check if ticket is already closed
	if ticket.Status == constants.TicketStatusClosed {
		response.BadRequest(c, "Ticket is already closed")
		return
	}

	// Set default reason if not provided
	reason := req.Reason
	if reason == "" {
		reason = "Ticket closed by user"
	}

	// Close ticket
	closedTicket, err := h.ticketService.CloseTicket(c.Request.Context(), ticketID, reason)
	if err != nil {
		logger.Error("User failed to close ticket",
			logger.Uint("ticket_id", ticketID),
			logger.Uint("user_id", user.ID),
			logger.ErrorField(err))

		convertedErr := sharedErrors.ConvertTicketErrorUint(err, ticketID)
		if sharedErrors.IsTicketNotFound(convertedErr) {
			response.NotFound(c, "Ticket not found")
		} else {
			response.InternalServerError(c, "Failed to close ticket")
		}
		return
	}

	logger.Info("User closed ticket successfully",
		logger.Uint("ticket_id", ticketID),
		logger.Uint("user_id", user.ID))

	// Return user-appropriate response
	// Convert to DTO and return response
	ticketResponse := dto.ToTicketUserResponse(closedTicket)
	response.OK(c, ticketResponse)
	// Return DTO to pool after use
	dto.PutTicketUserResponse(ticketResponse)
}

// GetTicketMessages godoc
// @Summary Get ticket messages
// @Description Get messages for a ticket (user can only access messages for own tickets, excluding internal notes)
// @Tags User-Tickets
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path uint true "Ticket ID"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} response.HALCollectionResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Router /tickets/{id}/messages [get]
func (h *UserTicketHandler) GetTicketMessages(c *gin.Context) {
	// Get current user from context
	user, ok := handlers.GetCurrentUser(c)
	if !ok {
		return // Error response already handled by GetCurrentUser
	}

	// Parse ticket ID
	ticketID, ok := handlers.ParseIDParam(c, "id")
	if !ok {
		return // Error response already handled by ParseIDParam
	}

	// Parse pagination
	pagination := handlers.ParsePagination(c)

	// For page-based pagination, convert to offset
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * pagination.Limit

	// Verify ticket exists and user owns it
	ticket, err := h.ticketService.GetTicket(c.Request.Context(), ticketID)
	if err != nil {
		logger.Error("User failed to verify ticket for messages",
			logger.Uint("ticket_id", ticketID),
			logger.Uint("user_id", user.ID),
			logger.ErrorField(err))
		response.NotFound(c, "Ticket not found")
		return
	}

	// Check ownership
	if ticket.UserID != user.ID {
		logger.Warn("User attempted to access messages for ticket they don't own",
			logger.Uint("ticket_id", ticketID),
			logger.Uint("user_id", user.ID),
			logger.Uint("ticket_owner_id", ticket.UserID))
		response.Forbidden(c, "You can only access messages for your own tickets")
		return
	}

	// Get messages (excluding internal notes)
	req := &dto.GetTicketMessagesRequest{
		TicketID:        ticketID,
		IncludeInternal: false, // Users should not see internal admin notes
		Limit:           pagination.Limit,
		Offset:          offset,
	}

	messages, total, err := h.ticketMessageService.GetTicketMessages(c.Request.Context(), req)
	if err != nil {
		logger.Error("User failed to get ticket messages",
			logger.Uint("ticket_id", ticketID),
			logger.Uint("user_id", user.ID),
			logger.ErrorField(err))
		response.InternalServerError(c, "Failed to get messages")
		return
	}

	// Convert to user response format using DTO functions
	responses := make([]*dto.TicketMessageUserResponse, len(messages))
	for i, message := range messages {
		responses[i] = dto.ToTicketMessageUserResponse(message)
	}

	response.SendPaginatedResponse(c, responses, total)

	// Return DTOs to pool after use
	for _, resp := range responses {
		dto.PutTicketMessageUserResponse(resp)
	}
}

// AddMessage godoc
// @Summary Add message to ticket
// @Description Add a message to a ticket (user can only add messages to own tickets)
// @Tags User-Tickets
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path uint true "Ticket ID"
// @Param message body dto.UserTicketMessageRequest true "Message data"
// @Success 201 {object} dto.TicketMessageUserResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /tickets/{id}/messages [post]
func (h *UserTicketHandler) AddMessage(c *gin.Context) {
	// Get current user from context
	user, ok := handlers.GetCurrentUser(c)
	if !ok {
		return // Error response already handled by GetCurrentUser
	}

	// Parse ticket ID
	ticketID, ok := handlers.ParseIDParam(c, "id")
	if !ok {
		return // Error response already handled by ParseIDParam
	}

	var req dto.UserTicketMessageRequest
	if !handlers.BindJSONRequest(c, &req) {
		return // Error response already handled by BindJSONRequest
	}

	// Verify ticket exists and user owns it
	ticket, err := h.ticketService.GetTicket(c.Request.Context(), ticketID)
	if err != nil {
		logger.Error("User failed to verify ticket for message creation",
			logger.Uint("ticket_id", ticketID),
			logger.Uint("user_id", user.ID),
			logger.ErrorField(err))
		response.NotFound(c, "Ticket not found")
		return
	}

	// Check ownership
	if ticket.UserID != user.ID {
		logger.Warn("User attempted to add message to ticket they don't own",
			logger.Uint("ticket_id", ticketID),
			logger.Uint("user_id", user.ID),
			logger.Uint("ticket_owner_id", ticket.UserID))
		response.Forbidden(c, "You can only add messages to your own tickets")
		return
	}

	// Check if ticket is closed
	if ticket.Status == constants.TicketStatusClosed {
		response.BadRequest(c, "Cannot add messages to closed tickets")
		return
	}

	// Create message request for service layer
	createReq := &dto.CreateTicketMessageRequest{
		Content:     req.Content,
		MessageType: constants.MessageTypeUser, // User messages are always type "user"
		Attachments: req.Attachments,
		IsInternal:  false, // User messages are never internal
		Metadata:    req.Metadata,
	}

	// Create message
	message, err := h.ticketMessageService.CreateTicketMessage(c.Request.Context(), ticketID, user.ID, createReq)
	if err != nil {
		logger.Error("User failed to create ticket message",
			logger.Uint("ticket_id", ticketID),
			logger.Uint("user_id", user.ID),
			logger.ErrorField(err))
		response.InternalServerError(c, "Failed to create message")
		return
	}

	logger.Info("User added message to ticket successfully",
		logger.Uint("ticket_id", ticketID),
		logger.Uint("message_id", message.ID),
		logger.Uint("user_id", user.ID))

	// Return user-appropriate response
	// Convert to DTO and return response
	messageResponse := dto.ToTicketMessageUserResponse(message)
	response.Created(c, messageResponse)
	// Return DTO to pool after use
	dto.PutTicketMessageUserResponse(messageResponse)
}
