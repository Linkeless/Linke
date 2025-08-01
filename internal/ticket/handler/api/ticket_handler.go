package api

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"linke/internal/response"
	"linke/internal/ticket/domain/valueobject"
	"linke/internal/ticket/handler/dto"
	"linke/internal/ticket/service"
	"linke/internal/ticket/service/command"
	"linke/internal/ticket/service/query"
	sharedvo "linke/internal/shared/valueobject"
)

// TicketHandler handles ticket-related HTTP requests for regular users
type TicketHandler struct {
	ticketService *service.TicketApplicationService
}

// NewTicketHandler creates a new ticket handler
func NewTicketHandler(ticketService *service.TicketApplicationService) *TicketHandler {
	return &TicketHandler{
		ticketService: ticketService,
	}
}

// CreateTicket creates a new ticket
// @Summary Create a new ticket
// @Description Create a new support ticket
// @Tags User-Tickets
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateTicketRequest true "Ticket creation request"
// @Success 201 {object} response.Response{data=dto.TicketResponse} "Ticket created successfully"
// @Failure 400 {object} response.Response "Invalid request"
// @Failure 401 {object} response.Response "Unauthorized"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /api/v1/user/tickets [post]
func (h *TicketHandler) CreateTicket(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	
	userIDUint, ok := userID.(uint)
	if !ok {
		response.InternalServerError(c, "Invalid user ID")
		return
	}
	
	var req dto.CreateTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request format", err.Error())
		return
	}
	
	// Create command
	userIDVO, err := sharedvo.NewUserIDFromUint(userIDUint)
	if err != nil {
		response.BadRequest(c, "Invalid user ID", err.Error())
		return
	}
	
	cmd := command.CreateTicketCommand{
		UserID:      userIDVO,
		Title:       req.Title,
		Description: req.Description,
		Category:    req.Category,
		Priority:    req.Priority,
		Tags:        req.Tags,
		Metadata:    req.Metadata,
	}
	
	// Execute command
	ticket, err := h.ticketService.CreateTicket(c.Request.Context(), cmd)
	if err != nil {
		response.InternalServerError(c, "Failed to create ticket", err.Error())
		return
	}
	
	// Convert to response DTO
	ticketResponse := dto.FromDomainTicket(ticket)
	
	response.CreatedWithMessage(c, "Ticket created successfully", ticketResponse)
}

// ListMyTickets lists tickets for the authenticated user
// @Summary List my tickets
// @Description Get a paginated list of tickets for the authenticated user
// @Tags User-Tickets
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param status query string false "Filter by status" Enums(open,in_progress,pending,resolved,closed)
// @Param priority query string false "Filter by priority" Enums(low,normal,high,urgent,critical)
// @Param category query string false "Filter by category" Enums(general,technical,billing,account,feature,bug,subscription,payment)
// @Param search query string false "Search in title, description, or ticket number"
// @Param limit query int false "Number of items per page" default(10)
// @Param offset query int false "Number of items to skip" default(0)
// @Param sort_by query string false "Sort by field" default(created_at)
// @Param sort_order query string false "Sort order" Enums(asc,desc) default(desc)
// @Success 200 {object} response.Response{data=dto.TicketListResponse} "Tickets retrieved successfully"
// @Failure 400 {object} response.Response "Invalid request"
// @Failure 401 {object} response.Response "Unauthorized"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /api/v1/user/tickets [get]
func (h *TicketHandler) ListMyTickets(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	
	userIDUint, ok := userID.(uint)
	if !ok {
		response.InternalServerError(c, "Invalid user ID")
		return
	}
	
	userIDValue, err := sharedvo.NewUserIDFromUint(userIDUint)
	if err != nil {
		response.BadRequest(c, "Invalid user ID", err.Error())
		return
	}
	
	// Parse query parameters
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	
	// Validate limit
	if limit > 100 {
		limit = 100
	}
	if limit <= 0 {
		limit = 10
	}
	
	// Create query
	query := query.ListTicketsQuery{
		UserID:      &userIDValue, // User can only see their own tickets
		Status:      getQueryString(c, "status"),
		Priority:    getQueryString(c, "priority"),
		Category:    getQueryString(c, "category"),
		SearchTerm:  c.Query("search"),
		Limit:       limit,
		Offset:      offset,
		SortBy:      c.DefaultQuery("sort_by", "created_at"),
		SortOrder:   c.DefaultQuery("sort_order", "desc"),
		RequestedBy: userIDValue,
		IsAdmin:     false,
	}
	
	// Execute query
	tickets, total, err := h.ticketService.ListTickets(c.Request.Context(), query)
	if err != nil {
		response.InternalServerError(c, "Failed to list tickets", err.Error())
		return
	}
	
	// Convert to response DTOs
	ticketResponses := make([]dto.TicketResponse, len(tickets))
	for i, ticket := range tickets {
		ticketResponses[i] = dto.FromDomainTicket(ticket)
	}
	
	listResponse := dto.TicketListResponse{
		Tickets: ticketResponses,
		Total:   total,
		Limit:   limit,
		Offset:  offset,
	}
	
	response.OK(c, "Tickets retrieved successfully", listResponse)
}

// GetTicket gets a specific ticket by ID
// @Summary Get ticket by ID
// @Description Get a specific ticket by ID (user can only access their own tickets)
// @Tags User-Tickets
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Ticket ID"
// @Success 200 {object} response.Response{data=dto.TicketResponse} "Ticket retrieved successfully"
// @Failure 400 {object} response.Response "Invalid request"
// @Failure 401 {object} response.Response "Unauthorized"
// @Failure 403 {object} response.Response "Access denied"
// @Failure 404 {object} response.Response "Ticket not found"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /api/v1/user/tickets/{id} [get]
func (h *TicketHandler) GetTicket(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	
	userIDUint, ok := userID.(uint)
	if !ok {
		response.InternalServerError(c, "Invalid user ID")
		return
	}
	
	// Parse ticket ID
	ticketIDStr := c.Param("id")
	ticketIDUint, err := strconv.ParseUint(ticketIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid ticket ID")
		return
	}
	
	// Create query
	userIDVO, err := sharedvo.NewUserIDFromUint(userIDUint)
	if err != nil {
		response.BadRequest(c, "Invalid user ID", err.Error())
		return
	}
	
	query := query.GetTicketQuery{
		TicketID: valueobject.NewTicketID(uint(ticketIDUint)),
		UserID:   userIDVO,
		IsAdmin:  false,
	}
	
	// Execute query
	ticket, err := h.ticketService.GetTicket(c.Request.Context(), query)
	if err != nil {
		if err.Error() == "access denied: cannot access this ticket" {
			response.Forbidden(c, "Access denied")
			return
		}
		if err.Error() == "ticket not found" {
			response.NotFound(c, "Ticket not found")
			return
		}
		response.InternalServerError(c, "Failed to get ticket", err.Error())
		return
	}
	
	// Convert to response DTO
	ticketResponse := dto.FromDomainTicket(ticket)
	
	response.OK(c, "Ticket retrieved successfully", ticketResponse)
}

// UpdateTicket updates a ticket
// @Summary Update ticket
// @Description Update ticket details (user can only update their own tickets)
// @Tags User-Tickets
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Ticket ID"
// @Param request body dto.UpdateTicketRequest true "Ticket update request"
// @Success 200 {object} response.Response{data=dto.TicketResponse} "Ticket updated successfully"
// @Failure 400 {object} response.Response "Invalid request"
// @Failure 401 {object} response.Response "Unauthorized"
// @Failure 403 {object} response.Response "Access denied"
// @Failure 404 {object} response.Response "Ticket not found"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /api/v1/user/tickets/{id} [put]
func (h *TicketHandler) UpdateTicket(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	
	userIDUint, ok := userID.(uint)
	if !ok {
		response.InternalServerError(c, "Invalid user ID")
		return
	}
	
	// Parse ticket ID
	ticketIDStr := c.Param("id")
	ticketIDUint, err := strconv.ParseUint(ticketIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid ticket ID")
		return
	}
	
	var req dto.UpdateTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request format", err.Error())
		return
	}
	
	// Create command
	userIDVO, err := sharedvo.NewUserIDFromUint(userIDUint)
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}
	
	cmd := command.UpdateTicketCommand{
		TicketID:    valueobject.NewTicketID(uint(ticketIDUint)),
		UserID:      userIDVO,
		IsAdmin:     false,
		Title:       req.Title,
		Description: req.Description,
		Category:    req.Category,
		Priority:    req.Priority,
		Tags:        req.Tags,
	}
	
	// Execute command
	ticket, err := h.ticketService.UpdateTicket(c.Request.Context(), cmd)
	if err != nil {
		if err.Error() == "access denied: can only update your own tickets" {
			response.Forbidden(c, "Access denied")
			return
		}
		response.InternalServerError(c, "Failed to update ticket", err.Error())
		return
	}
	
	// Convert to response DTO
	ticketResponse := dto.FromDomainTicket(ticket)
	
	response.OK(c, "Ticket updated successfully", ticketResponse)
}

// CloseTicket closes a ticket
// @Summary Close ticket
// @Description Close a ticket (user can only close their own tickets)
// @Tags User-Tickets
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Ticket ID"
// @Success 200 {object} response.Response{data=dto.TicketResponse} "Ticket closed successfully"
// @Failure 400 {object} response.Response "Invalid request"
// @Failure 401 {object} response.Response "Unauthorized"
// @Failure 403 {object} response.Response "Access denied"
// @Failure 404 {object} response.Response "Ticket not found"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /api/v1/user/tickets/{id}/close [post]
func (h *TicketHandler) CloseTicket(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	
	userIDUint, ok := userID.(uint)
	if !ok {
		response.InternalServerError(c, "Invalid user ID")
		return
	}
	
	// Parse ticket ID
	ticketIDStr := c.Param("id")
	ticketIDUint, err := strconv.ParseUint(ticketIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid ticket ID")
		return
	}
	
	// Create command
	userIDVO, err := sharedvo.NewUserIDFromUint(userIDUint)
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}
	
	cmd := command.CloseTicketCommand{
		TicketID: valueobject.NewTicketID(uint(ticketIDUint)),
		ClosedBy: userIDVO,
		IsAdmin:  false,
	}
	
	// Execute command
	ticket, err := h.ticketService.CloseTicket(c.Request.Context(), cmd)
	if err != nil {
		if err.Error() == "access denied: only ticket owner or admins can close tickets" {
			response.Forbidden(c, "Access denied")
			return
		}
		response.InternalServerError(c, "Failed to close ticket", err.Error())
		return
	}
	
	// Convert to response DTO
	ticketResponse := dto.FromDomainTicket(ticket)
	
	response.OK(c, "Ticket closed successfully", ticketResponse)
}

// AddMessage adds a message to a ticket
// @Summary Add message to ticket
// @Description Add a message to a ticket
// @Tags User-Tickets
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Ticket ID"
// @Param request body dto.AddTicketMessageRequest true "Message request"
// @Success 201 {object} response.Response{data=dto.TicketResponse} "Message added successfully"
// @Failure 400 {object} response.Response "Invalid request"
// @Failure 401 {object} response.Response "Unauthorized"
// @Failure 403 {object} response.Response "Access denied"
// @Failure 404 {object} response.Response "Ticket not found"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /api/v1/user/tickets/{id}/messages [post]
func (h *TicketHandler) AddMessage(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	
	userIDUint, ok := userID.(uint)
	if !ok {
		response.InternalServerError(c, "Invalid user ID")
		return
	}
	
	// Parse ticket ID
	ticketIDStr := c.Param("id")
	ticketIDUint, err := strconv.ParseUint(ticketIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid ticket ID")
		return
	}
	
	var req dto.AddTicketMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request format", err.Error())
		return
	}
	
	// Convert attachments
	attachments := make([]command.AttachmentData, len(req.Attachments))
	for i, att := range req.Attachments {
		attachments[i] = command.AttachmentData{
			Name: att.Name,
			URL:  att.URL,
			Size: att.Size,
			Type: att.Type,
		}
	}
	
	// Users can only create user messages, not admin or system messages
	messageType := req.MessageType
	if messageType == "" || messageType == "admin" || messageType == "system" {
		messageType = "user"
	}
	
	// Create command
	userIDVO, err := sharedvo.NewUserIDFromUint(userIDUint)
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}
	
	cmd := command.AddTicketMessageCommand{
		TicketID:    valueobject.NewTicketID(uint(ticketIDUint)),
		UserID:      userIDVO,
		Content:     req.Content,
		MessageType: messageType,
		Attachments: attachments,
		IsInternal:  false, // Users cannot create internal messages
		Metadata:    req.Metadata,
	}
	
	// Execute command
	ticket, err := h.ticketService.AddTicketMessage(c.Request.Context(), cmd)
	if err != nil {
		response.InternalServerError(c, "Failed to add message", err.Error())
		return
	}
	
	// Convert to response DTO
	ticketResponse := dto.FromDomainTicket(ticket)
	
	response.CreatedWithMessage(c, "Message added successfully", ticketResponse)
}

// GetMessages gets messages for a ticket
// @Summary Get ticket messages
// @Description Get messages for a ticket (user can only access their own ticket messages)
// @Tags User-Tickets
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Ticket ID"
// @Param limit query int false "Number of items per page" default(50)
// @Param offset query int false "Number of items to skip" default(0)
// @Success 200 {object} response.Response{data=dto.TicketMessageListResponse} "Messages retrieved successfully"
// @Failure 400 {object} response.Response "Invalid request"
// @Failure 401 {object} response.Response "Unauthorized"
// @Failure 403 {object} response.Response "Access denied"
// @Failure 404 {object} response.Response "Ticket not found"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /api/v1/user/tickets/{id}/messages [get]
func (h *TicketHandler) GetMessages(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	
	userIDUint, ok := userID.(uint)
	if !ok {
		response.InternalServerError(c, "Invalid user ID")
		return
	}
	
	// Parse ticket ID
	ticketIDStr := c.Param("id")
	ticketIDUint, err := strconv.ParseUint(ticketIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid ticket ID")
		return
	}
	
	// Parse query parameters
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	
	// Validate limit
	if limit > 100 {
		limit = 100
	}
	if limit <= 0 {
		limit = 50
	}
	
	// Create query
	userIDVO, err := sharedvo.NewUserIDFromUint(userIDUint)
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}
	
	query := query.GetTicketMessagesQuery{
		TicketID:        valueobject.NewTicketID(uint(ticketIDUint)),
		IncludeInternal: false, // Users cannot see internal messages
		Limit:           limit,
		Offset:          offset,
		RequestedBy:     userIDVO,
		IsAdmin:         false,
	}
	
	// Execute query
	messages, total, err := h.ticketService.GetTicketMessages(c.Request.Context(), query)
	if err != nil {
		if err.Error() == "access denied: cannot access this ticket" {
			response.Forbidden(c, "Access denied")
			return
		}
		response.InternalServerError(c, "Failed to get messages", err.Error())
		return
	}
	
	// Convert to response DTOs
	messageResponses := make([]dto.TicketMessageResponse, len(messages))
	for i, message := range messages {
		messageResponses[i] = dto.FromDomainTicketMessage(message)
	}
	
	listResponse := dto.TicketMessageListResponse{
		Messages: messageResponses,
		Total:    total,
		Limit:    limit,
		Offset:   offset,
	}
	
	response.OK(c, "Messages retrieved successfully", listResponse)
}

// Helper function to get string pointer from query
func getQueryString(c *gin.Context, key string) *string {
	value := c.Query(key)
	if value == "" {
		return nil
	}
	return &value
}