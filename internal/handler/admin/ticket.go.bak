package admin

import (
	"strconv"

	"linke/internal/logger"
	"linke/internal/middleware"
	"linke/internal/model"
	"linke/internal/response"
	"linke/internal/service"

	"github.com/gin-gonic/gin"
)

// AdminTicketHandler handles admin-facing ticket operations
type AdminTicketHandler struct {
	ticketService        *service.TicketService
	ticketMessageService *service.TicketMessageService
}

// NewAdminTicketHandler creates a new AdminTicketHandler
func NewAdminTicketHandler(ticketService *service.TicketService, ticketMessageService *service.TicketMessageService) *AdminTicketHandler {
	return &AdminTicketHandler{
		ticketService:        ticketService,
		ticketMessageService: ticketMessageService,
	}
}

// ListTickets gets all tickets with admin filtering
// @Summary List all tickets (Admin)
// @Description Get all tickets in the system with advanced filtering options
// @Tags Admin-Ticket-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param user_id query int false "Filter by user ID"
// @Param assigned_to_id query int false "Filter by assigned admin ID"
// @Param status query string false "Filter by status" Enums(open,in_progress,pending,resolved,closed)
// @Param priority query string false "Filter by priority" Enums(low,normal,high,urgent,critical)
// @Param category query string false "Filter by category" Enums(general,technical,billing,account,feature,bug,subscription,payment)
// @Param search query string false "Search in title, description, or ticket number"
// @Param limit query int false "Limit number of results" minimum(1) maximum(100) default(10)
// @Param offset query int false "Offset for pagination" minimum(0) default(0)
// @Success 200 {object} response.Response{data=response.PaginatedResponse{items=[]model.TicketResponse}} "Tickets retrieved successfully"
// @Failure 400 {object} response.Response "Invalid request"
// @Failure 401 {object} response.Response "Unauthorized"
// @Failure 403 {object} response.Response "Access denied"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /admin/tickets [get]
func (h *AdminTicketHandler) ListTickets(c *gin.Context) {
	var req service.GetTicketsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		logger.Error("Failed to bind get tickets request", logger.Error2("error", err))
		response.BadRequest(c, "Invalid request format", err.Error())
		return
	}

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

	// Convert to admin response format
	adminTickets := make([]model.TicketResponse, len(tickets))
	for i, ticket := range tickets {
		adminTickets[i] = *ticket.ToResponse()
	}

	// Send paginated response
	response.OKPaginated(c, "Tickets retrieved successfully", adminTickets, total, req.Limit, req.Offset)
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
func (h *AdminTicketHandler) GetTicket(c *gin.Context) {
	// Parse ticket ID
	ticketIDStr := c.Param("id")
	ticketID, err := strconv.ParseUint(ticketIDStr, 10, 32)
	if err != nil {
		logger.Error("Invalid ticket ID", logger.Error2("error", err))
		response.BadRequest(c, "Invalid ticket ID", err.Error())
		return
	}

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

	// Convert to admin response format
	adminResponse := ticket.ToResponse()
	response.OK(c, "Ticket retrieved successfully", adminResponse)
}

// UpdateTicket updates a ticket
// @Summary Update a ticket (Admin)
// @Description Update ticket details (admin only)
// @Tags Admin-Ticket-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Ticket ID"
// @Param request body service.UpdateTicketRequest true "Ticket update request"
// @Success 200 {object} response.Response{data=model.TicketResponse} "Ticket updated successfully"
// @Failure 400 {object} response.Response "Invalid request"
// @Failure 401 {object} response.Response "Unauthorized"
// @Failure 403 {object} response.Response "Access denied"
// @Failure 404 {object} response.Response "Ticket not found"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /admin/tickets/{id} [put]
func (h *AdminTicketHandler) UpdateTicket(c *gin.Context) {
	// Parse ticket ID
	ticketIDStr := c.Param("id")
	ticketID, err := strconv.ParseUint(ticketIDStr, 10, 32)
	if err != nil {
		logger.Error("Invalid ticket ID", logger.Error2("error", err))
		response.BadRequest(c, "Invalid ticket ID", err.Error())
		return
	}

	var req service.UpdateTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Failed to bind update ticket request", logger.Error2("error", err))
		response.BadRequest(c, "Invalid request format", err.Error())
		return
	}

	// Update ticket
	ticket, err := h.ticketService.UpdateTicket(c.Request.Context(), uint(ticketID), &req)
	if err != nil {
		logger.Error("Failed to update ticket", logger.Error2("error", err))
		if err.Error() == "ticket not found" {
			response.NotFound(c, "Ticket not found")
			return
		}
		response.InternalServerError(c, "Failed to update ticket", err.Error())
		return
	}

	// Convert to admin response format
	adminResponse := ticket.ToResponse()
	response.OK(c, "Ticket updated successfully", adminResponse)
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
func (h *AdminTicketHandler) AssignTicket(c *gin.Context) {
	// Parse ticket ID
	ticketIDStr := c.Param("id")
	ticketID, err := strconv.ParseUint(ticketIDStr, 10, 32)
	if err != nil {
		logger.Error("Invalid ticket ID", logger.Error2("error", err))
		response.BadRequest(c, "Invalid ticket ID", err.Error())
		return
	}

	var req service.AssignTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Failed to bind assign ticket request", logger.Error2("error", err))
		response.BadRequest(c, "Invalid request format", err.Error())
		return
	}

	// Assign ticket
	ticket, err := h.ticketService.AssignTicket(c.Request.Context(), uint(ticketID), &req)
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
func (h *AdminTicketHandler) ResolveTicket(c *gin.Context) {
	// Parse ticket ID
	ticketIDStr := c.Param("id")
	ticketID, err := strconv.ParseUint(ticketIDStr, 10, 32)
	if err != nil {
		logger.Error("Invalid ticket ID", logger.Error2("error", err))
		response.BadRequest(c, "Invalid ticket ID", err.Error())
		return
	}

	var req service.ResolveTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Failed to bind resolve ticket request", logger.Error2("error", err))
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

	// Resolve ticket
	ticket, err := h.ticketService.ResolveTicket(c.Request.Context(), uint(ticketID), currentUser.ID, &req)
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
func (h *AdminTicketHandler) CloseTicket(c *gin.Context) {
	// Parse ticket ID
	ticketIDStr := c.Param("id")
	ticketID, err := strconv.ParseUint(ticketIDStr, 10, 32)
	if err != nil {
		logger.Error("Invalid ticket ID", logger.Error2("error", err))
		response.BadRequest(c, "Invalid ticket ID", err.Error())
		return
	}

	// Close ticket
	ticket, err := h.ticketService.CloseTicket(c.Request.Context(), uint(ticketID))
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

// DeleteTicket deletes a ticket
// @Summary Delete a ticket (Admin)
// @Description Soft delete a ticket (admin only)
// @Tags Admin-Ticket-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Ticket ID"
// @Success 200 {object} response.Response "Ticket deleted successfully"
// @Failure 400 {object} response.Response "Invalid ticket ID"
// @Failure 401 {object} response.Response "Unauthorized"
// @Failure 403 {object} response.Response "Access denied"
// @Failure 404 {object} response.Response "Ticket not found"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /admin/tickets/{id} [delete]
func (h *AdminTicketHandler) DeleteTicket(c *gin.Context) {
	// Parse ticket ID
	ticketIDStr := c.Param("id")
	ticketID, err := strconv.ParseUint(ticketIDStr, 10, 32)
	if err != nil {
		logger.Error("Invalid ticket ID", logger.Error2("error", err))
		response.BadRequest(c, "Invalid ticket ID", err.Error())
		return
	}

	// Delete ticket
	if err := h.ticketService.DeleteTicket(c.Request.Context(), uint(ticketID)); err != nil {
		logger.Error("Failed to delete ticket", logger.Error2("error", err))
		if err.Error() == "ticket not found" {
			response.NotFound(c, "Ticket not found")
			return
		}
		response.InternalServerError(c, "Failed to delete ticket", err.Error())
		return
	}

	response.OK(c, "Ticket deleted successfully", nil)
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
func (h *AdminTicketHandler) AddTicketMessage(c *gin.Context) {
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

	// Set message type to 'admin' if not specified
	if req.MessageType == "" {
		req.MessageType = "admin"
	}

	// Create message
	message, err := h.ticketMessageService.CreateTicketMessage(c.Request.Context(), uint(ticketID), currentUser.ID, &req)
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
func (h *AdminTicketHandler) GetTicketMessages(c *gin.Context) {
	// Parse ticket ID
	ticketIDStr := c.Param("id")
	ticketID, err := strconv.ParseUint(ticketIDStr, 10, 32)
	if err != nil {
		logger.Error("Invalid ticket ID", logger.Error2("error", err))
		response.BadRequest(c, "Invalid ticket ID", err.Error())
		return
	}

	var req service.GetTicketMessagesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		logger.Error("Failed to bind get ticket messages request", logger.Error2("error", err))
		response.BadRequest(c, "Invalid request format", err.Error())
		return
	}

	// Set ticket ID from path
	req.TicketID = uint(ticketID)

	// Set defaults
	if req.Limit == 0 {
		req.Limit = 10
	}
	
	// Admin users can see internal messages by default
	if c.Query("include_internal") == "" {
		req.IncludeInternal = true
	}

	// Get messages
	messages, total, err := h.ticketMessageService.GetTicketMessages(c.Request.Context(), &req)
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

// GetTicketStats gets ticket statistics
// @Summary Get ticket statistics (Admin)
// @Description Get comprehensive statistics about tickets
// @Tags Admin-Ticket-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=map[string]interface{}} "Statistics retrieved successfully"
// @Failure 401 {object} response.Response "Unauthorized"
// @Failure 403 {object} response.Response "Access denied"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /admin/tickets/stats [get]
func (h *AdminTicketHandler) GetTicketStats(c *gin.Context) {
	// Get ticket statistics
	stats, err := h.ticketService.GetTicketStats(c.Request.Context())
	if err != nil {
		logger.Error("Failed to get ticket stats", logger.Error2("error", err))
		response.InternalServerError(c, "Failed to get statistics", err.Error())
		return
	}

	response.OK(c, "Statistics retrieved successfully", stats)
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
func (h *AdminTicketHandler) GetTicketByNumber(c *gin.Context) {
	// Get ticket number from path
	ticketNo := c.Param("ticket_no")
	if ticketNo == "" {
		logger.Error("Empty ticket number")
		response.BadRequest(c, "Ticket number is required")
		return
	}

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

	// Convert to admin response format
	adminResponse := ticket.ToResponse()
	response.OK(c, "Ticket retrieved successfully", adminResponse)
}

// SetupAdminTicketRoutes sets up routes for admin ticket operations
func (h *AdminTicketHandler) SetupAdminTicketRoutes(adminGroup *gin.RouterGroup) {
	// Note: Authentication and admin middleware are already applied by parent group
	// adminGroup.Use(middleware.AuthMiddleware(authService))
	// adminGroup.Use(middleware.RequireAdmin())

	// Ticket routes
	ticketGroup := adminGroup.Group("/tickets")
	{
		// Basic CRUD operations
		ticketGroup.GET("", h.ListTickets)
		ticketGroup.GET("/:id", h.GetTicket)
		ticketGroup.PUT("/:id", h.UpdateTicket)
		ticketGroup.DELETE("/:id", h.DeleteTicket)
		ticketGroup.GET("/number/:ticket_no", h.GetTicketByNumber)
		
		// Ticket management operations
		ticketGroup.POST("/:id/assign", h.AssignTicket)
		ticketGroup.POST("/:id/resolve", h.ResolveTicket)
		ticketGroup.POST("/:id/close", h.CloseTicket)
		
		// Message operations
		ticketGroup.POST("/:id/messages", h.AddTicketMessage)
		ticketGroup.GET("/:id/messages", h.GetTicketMessages)
		
		// Statistics
		ticketGroup.GET("/stats", h.GetTicketStats)
	}
}