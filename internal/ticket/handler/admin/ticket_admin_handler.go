package admin

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

// TicketAdminHandler handles ticket-related HTTP requests for admins
type TicketAdminHandler struct {
	ticketService *service.TicketApplicationService
}

// NewTicketAdminHandler creates a new ticket admin handler
func NewTicketAdminHandler(ticketService *service.TicketApplicationService) *TicketAdminHandler {
	return &TicketAdminHandler{
		ticketService: ticketService,
	}
}

// ListAllTickets lists all tickets (admin only)
// @Summary List all tickets (Admin)
// @Description Get a paginated list of all tickets (admin only)
// @Tags Admin-Tickets
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param user_id query int false "Filter by user ID"
// @Param assigned_to_id query int false "Filter by assigned admin ID"
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
// @Failure 403 {object} response.Response "Access denied"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /api/v1/admin/tickets [get]
func (h *TicketAdminHandler) ListAllTickets(c *gin.Context) {
	// Get admin user ID from context
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
	
	// Convert to shared UserID
	requestedBy, err := sharedvo.NewUserIDFromUint(userIDUint)
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
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
	
	// Parse optional user filter
	var filterUserID *sharedvo.UserID
	if userIDStr := c.Query("user_id"); userIDStr != "" {
		if filterUserIDUint, err := strconv.ParseUint(userIDStr, 10, 32); err == nil {
			if filterUserIDValue, err := sharedvo.NewUserIDFromUint(uint(filterUserIDUint)); err == nil {
				filterUserID = &filterUserIDValue
			}
		}
	}
	
	// Parse optional assigned filter
	var assignedToID *sharedvo.UserID
	if assignedIDStr := c.Query("assigned_to_id"); assignedIDStr != "" {
		if assignedIDUint, err := strconv.ParseUint(assignedIDStr, 10, 32); err == nil {
			if assignedIDValue, err := sharedvo.NewUserIDFromUint(uint(assignedIDUint)); err == nil {
				assignedToID = &assignedIDValue
			}
		}
	}
	
	// Create query
	query := query.ListTicketsQuery{
		UserID:       filterUserID, // Admin can filter by any user or see all
		AssignedToID: assignedToID,
		Status:       getQueryString(c, "status"),
		Priority:     getQueryString(c, "priority"),
		Category:     getQueryString(c, "category"),
		SearchTerm:   c.Query("search"),
		Limit:        limit,
		Offset:       offset,
		SortBy:       c.DefaultQuery("sort_by", "created_at"),
		SortOrder:    c.DefaultQuery("sort_order", "desc"),
		RequestedBy:  requestedBy,
		IsAdmin:      true,
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

// AssignTicket assigns a ticket to an admin
// @Summary Assign ticket (Admin)
// @Description Assign a ticket to an admin user
// @Tags Admin-Tickets
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Ticket ID"
// @Param request body dto.AssignTicketRequest true "Assignment request"
// @Success 200 {object} response.Response{data=dto.TicketResponse} "Ticket assigned successfully"
// @Failure 400 {object} response.Response "Invalid request"
// @Failure 401 {object} response.Response "Unauthorized"
// @Failure 403 {object} response.Response "Access denied"
// @Failure 404 {object} response.Response "Ticket not found"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /api/v1/admin/tickets/{id}/assign [post]
func (h *TicketAdminHandler) AssignTicket(c *gin.Context) {
	// Get admin user ID from context
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
	
	var req dto.AssignTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request format", err.Error())
		return
	}
	
	// Create user IDs
	assignedToID, err := sharedvo.NewUserIDFromUint(req.AssignedToID)
	if err != nil {
		response.BadRequest(c, "Invalid assigned to user ID")
		return
	}
	
	assignedBy, err := sharedvo.NewUserIDFromUint(userIDUint)
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}
	
	// Create command
	cmd := command.AssignTicketCommand{
		TicketID:     valueobject.NewTicketID(uint(ticketIDUint)),
		AssignedToID: assignedToID,
		AssignedBy:   assignedBy,
	}
	
	// Execute command
	ticket, err := h.ticketService.AssignTicket(c.Request.Context(), cmd)
	if err != nil {
		response.InternalServerError(c, "Failed to assign ticket", err.Error())
		return
	}
	
	// Convert to response DTO
	ticketResponse := dto.FromDomainTicket(ticket)
	
	response.OK(c, "Ticket assigned successfully", ticketResponse)
}

// ChangeTicketStatus changes the status of a ticket
// @Summary Change ticket status (Admin)
// @Description Change the status of a ticket
// @Tags Admin-Tickets
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Ticket ID"
// @Param request body dto.ChangeTicketStatusRequest true "Status change request"
// @Success 200 {object} response.Response{data=dto.TicketResponse} "Ticket status changed successfully"
// @Failure 400 {object} response.Response "Invalid request"
// @Failure 401 {object} response.Response "Unauthorized"
// @Failure 403 {object} response.Response "Access denied"
// @Failure 404 {object} response.Response "Ticket not found"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /api/v1/admin/tickets/{id}/status [put]
func (h *TicketAdminHandler) ChangeTicketStatus(c *gin.Context) {
	// Get admin user ID from context
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
	
	var req dto.ChangeTicketStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request format", err.Error())
		return
	}
	
	// Create user ID
	changedBy, err := sharedvo.NewUserIDFromUint(userIDUint)
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}
	
	// Create command
	cmd := command.ChangeTicketStatusCommand{
		TicketID:  valueobject.NewTicketID(uint(ticketIDUint)),
		NewStatus: req.Status,
		ChangedBy: changedBy,
		IsAdmin:   true,
	}
	
	// Execute command
	ticket, err := h.ticketService.ChangeTicketStatus(c.Request.Context(), cmd)
	if err != nil {
		response.InternalServerError(c, "Failed to change ticket status", err.Error())
		return
	}
	
	// Convert to response DTO
	ticketResponse := dto.FromDomainTicket(ticket)
	
	response.OK(c, "Ticket status changed successfully", ticketResponse)
}

// ResolveTicket resolves a ticket
// @Summary Resolve ticket (Admin)
// @Description Resolve a ticket with a solution
// @Tags Admin-Tickets
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Ticket ID"
// @Param request body dto.ResolveTicketRequest true "Resolution request"
// @Success 200 {object} response.Response{data=dto.TicketResponse} "Ticket resolved successfully"
// @Failure 400 {object} response.Response "Invalid request"
// @Failure 401 {object} response.Response "Unauthorized"
// @Failure 403 {object} response.Response "Access denied"
// @Failure 404 {object} response.Response "Ticket not found"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /api/v1/admin/tickets/{id}/resolve [post]
func (h *TicketAdminHandler) ResolveTicket(c *gin.Context) {
	// Get admin user ID from context
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
	
	var req dto.ResolveTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request format", err.Error())
		return
	}
	
	// Create user ID
	resolvedBy, err := sharedvo.NewUserIDFromUint(userIDUint)
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}
	
	// Create command
	cmd := command.ResolveTicketCommand{
		TicketID:   valueobject.NewTicketID(uint(ticketIDUint)),
		ResolvedBy: resolvedBy,
		Resolution: req.Resolution,
	}
	
	// Execute command
	ticket, err := h.ticketService.ResolveTicket(c.Request.Context(), cmd)
	if err != nil {
		response.InternalServerError(c, "Failed to resolve ticket", err.Error())
		return
	}
	
	// Convert to response DTO
	ticketResponse := dto.FromDomainTicket(ticket)
	
	response.OK(c, "Ticket resolved successfully", ticketResponse)
}

// AddAdminMessage adds an admin message to a ticket
// @Summary Add admin message (Admin)
// @Description Add an admin message to a ticket (can be internal)
// @Tags Admin-Tickets
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
// @Router /api/v1/admin/tickets/{id}/messages [post]
func (h *TicketAdminHandler) AddAdminMessage(c *gin.Context) {
	// Get admin user ID from context
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
	
	// Default to admin message type for admin users
	messageType := req.MessageType
	if messageType == "" {
		messageType = "admin"
	}
	
	// Create user ID
	userIDVO, err := sharedvo.NewUserIDFromUint(userIDUint)
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}
	
	// Create command
	cmd := command.AddTicketMessageCommand{
		TicketID:    valueobject.NewTicketID(uint(ticketIDUint)),
		UserID:      userIDVO,
		Content:     req.Content,
		MessageType: messageType,
		Attachments: attachments,
		IsInternal:  req.IsInternal, // Admins can create internal messages
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

// GetAllMessages gets all messages for a ticket (including internal)
// @Summary Get all ticket messages (Admin)
// @Description Get all messages for a ticket including internal messages (admin only)
// @Tags Admin-Tickets
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
// @Router /api/v1/admin/tickets/{id}/messages [get]
func (h *TicketAdminHandler) GetAllMessages(c *gin.Context) {
	// Get admin user ID from context
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
	
	// Convert to shared UserID
	requestedBy, err := sharedvo.NewUserIDFromUint(userIDUint)
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
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
	query := query.GetTicketMessagesQuery{
		TicketID:        valueobject.NewTicketID(uint(ticketIDUint)),
		IncludeInternal: true, // Admins can see internal messages
		Limit:           limit,
		Offset:          offset,
		RequestedBy:     requestedBy,
		IsAdmin:         true,
	}
	
	// Execute query
	messages, total, err := h.ticketService.GetTicketMessages(c.Request.Context(), query)
	if err != nil {
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

// GetTicketStatistics gets ticket statistics
// @Summary Get ticket statistics (Admin)
// @Description Get comprehensive ticket statistics (admin only)
// @Tags Admin-Tickets
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=dto.TicketStatisticsResponse} "Statistics retrieved successfully"
// @Failure 401 {object} response.Response "Unauthorized"
// @Failure 403 {object} response.Response "Access denied"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /api/v1/admin/tickets/statistics [get]
func (h *TicketAdminHandler) GetTicketStatistics(c *gin.Context) {
	// Get admin user ID from context
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
	
	// Convert to shared UserID
	requestedBy, err := sharedvo.NewUserIDFromUint(userIDUint)
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}
	
	// Create query
	query := query.GetTicketStatisticsQuery{
		RequestedBy: requestedBy,
		IsAdmin:     true,
	}
	
	// Execute query
	stats, err := h.ticketService.GetTicketStatistics(c.Request.Context(), query)
	if err != nil {
		response.InternalServerError(c, "Failed to get statistics", err.Error())
		return
	}
	
	// Convert to response DTO
	statsResponse := dto.TicketStatisticsResponse{
		Total:         stats.Total,
		ByStatus:      stats.ByStatus,
		ByPriority:    stats.ByPriority,
		ByCategory:    stats.ByCategory,
		Unassigned:    stats.Unassigned,
		OverdueCount:  stats.OverdueCount,
		ResolvedToday: stats.ResolvedToday,
		CreatedToday:  stats.CreatedToday,
	}
	
	response.OK(c, "Statistics retrieved successfully", statsResponse)
}

// Helper function to get string pointer from query
func getQueryString(c *gin.Context, key string) *string {
	value := c.Query(key)
	if value == "" {
		return nil
	}
	return &value
}