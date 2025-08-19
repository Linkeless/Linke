package handlers

import (
	"strconv"

	"linke/internal/domains/ticket/dto"
	sharedErrors "linke/internal/shared/errors"
	"linke/internal/shared/handlers"
	"linke/internal/shared/logger"
	"linke/internal/shared/response"

	"github.com/gin-gonic/gin"
)

// AdminTicketBaseHandler handles basic ticket CRUD operations
type AdminTicketBaseHandler struct {
	*AdminTicketHandlerBase
}

// NewAdminTicketBaseHandler creates a new admin ticket base handler
func NewAdminTicketBaseHandler(base *AdminTicketHandlerBase) *AdminTicketBaseHandler {
	return &AdminTicketBaseHandler{
		AdminTicketHandlerBase: base,
	}
}

// CreateTicket godoc
// @Summary Create new ticket
// @Description Create a new support ticket (Admin only)
// @Tags Admin-Ticket-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param ticket body dto.AdminCreateTicketRequest true "Ticket creation data"
// @Success 201 {object} dto.TicketResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/tickets [post]
func (h *AdminTicketBaseHandler) CreateTicket(c *gin.Context) {
	var req dto.AdminCreateTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Verify user exists
	user, err := h.userService.GetUserByID(c.Request.Context(), req.UserID)
	if err != nil {
		logger.Error("Failed to verify user for ticket creation",
			logger.Uint("user_id", req.UserID),
			logger.ErrorField(err))

		convertedErr := sharedErrors.ConvertUserErrorUint(err, req.UserID)
		if sharedErrors.IsUserNotFound(convertedErr) {
			response.NotFound(c, "User not found")
		} else {
			response.InternalServerError(c, "Failed to verify user")
		}
		return
	}

	// Verify assigned user if specified
	if req.AssignedToID != nil {
		assignedUser, err := h.userService.GetUserByID(c.Request.Context(), *req.AssignedToID)
		if err != nil {
			logger.Error("Failed to verify assigned user for ticket creation",
				logger.Uint("assigned_to_id", *req.AssignedToID),
				logger.ErrorField(err))

			convertedErr := sharedErrors.ConvertUserErrorUint(err, *req.AssignedToID)
			if sharedErrors.IsUserNotFound(convertedErr) {
				response.NotFound(c, "Assigned user not found")
			} else {
				response.InternalServerError(c, "Failed to verify assigned user")
			}
			return
		}

		// Verify assigned user is admin
		if assignedUser.Role != "admin" {
			response.BadRequest(c, "Assigned user must be an admin")
			return
		}
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

	// Create ticket
	ticket, err := h.ticketService.CreateTicket(c.Request.Context(), req.UserID, createReq)
	if err != nil {
		logger.Error("Admin failed to create ticket",
			logger.Uint("user_id", req.UserID),
			logger.String("category", req.Category),
			logger.ErrorField(err))
		response.InternalServerError(c, "Failed to create ticket")
		return
	}

	// Assign ticket if specified
	if req.AssignedToID != nil {
		assignReq := &dto.AssignTicketRequest{
			AssignedToID: *req.AssignedToID,
		}
		ticket, err = h.ticketService.AssignTicket(c.Request.Context(), ticket.ID, assignReq)
		if err != nil {
			logger.Error("Failed to assign ticket during creation",
				logger.Uint("ticket_id", ticket.ID),
				logger.Uint("assigned_to_id", *req.AssignedToID),
				logger.ErrorField(err))
			// Continue with unassigned ticket rather than failing
		}
	}

	// Populate user data in response using DTO functions
	ticketResponse := dto.ToTicketResponse(ticket)
	ticketResponse.User = handlers.ConvertUserToBasicDTO(user)

	logger.Info("Admin created ticket successfully",
		logger.Uint("ticket_id", ticket.ID),
		logger.String("ticket_no", ticket.TicketNo),
		logger.Uint("user_id", req.UserID))

	response.Created(c, ticketResponse)
	// Return DTO to pool after use
	dto.PutTicketResponse(ticketResponse)
}

// ListTickets godoc
// @Summary List all tickets
// @Description Get paginated list of all tickets with filtering options (Admin only)
// @Tags Admin-Ticket-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param user_id query uint false "Filter by user ID" example(123)
// @Param assigned_to_id query uint false "Filter by assigned agent ID" example(456)
// @Param status query string false "Filter by status" Enums(open,in_progress,pending,resolved,closed) example(open)
// @Param priority query string false "Filter by priority" Enums(low,normal,high,urgent,critical) example(high)
// @Param category query string false "Filter by category" Enums(general,technical,billing,account,feature,bug,subscription,payment) example(technical)
// @Param search query string false "Search in title, description, or ticket number" example("login issue")
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} response.HALCollectionResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/tickets [get]
func (h *AdminTicketBaseHandler) ListTickets(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	// Parse assigned_to_id if provided
	var assignedToID *uint
	if assignedToIDStr := c.Query("assigned_to_id"); assignedToIDStr != "" {
		if parsedID, err := strconv.ParseUint(assignedToIDStr, 10, 32); err == nil {
			id := uint(parsedID)
			assignedToID = &id
		}
	}

	// Parse user_id if provided
	var userID uint
	if userIDStr := c.Query("user_id"); userIDStr != "" {
		if parsedID, err := strconv.ParseUint(userIDStr, 10, 32); err == nil {
			userID = uint(parsedID)
		}
	}

	// Create request for service layer
	req := &dto.GetTicketsRequest{
		UserID:       userID,
		AssignedToID: assignedToID,
		Status:       c.Query("status"),
		Priority:     c.Query("priority"),
		Category:     c.Query("category"),
		Search:       c.Query("search"),
		Limit:        limit,
		Offset:       offset,
	}

	tickets, total, err := h.ticketService.GetTickets(c.Request.Context(), req)
	if err != nil {
		logger.Error("Admin failed to list tickets", logger.ErrorField(err))
		response.InternalServerError(c, "Failed to list tickets")
		return
	}

	// Convert to response format and batch load user data
	responses := make([]*dto.TicketResponse, len(tickets))
	userIDCollector := handlers.NewUserIDCollector()

	// First pass: convert tickets and collect user IDs
	for i, ticket := range tickets {
		responses[i] = dto.ToTicketResponse(ticket)

		// Collect all user IDs that need to be loaded
		userIDCollector.Add(ticket.UserID)
		userIDCollector.AddPtr(ticket.AssignedToID)
		userIDCollector.AddPtr(ticket.ResolvedByID)
	}

	// Batch load all users
	userLoader := handlers.NewBatchUserLoader(h.userService)
	if err := userLoader.LoadUsers(c.Request.Context(), userIDCollector.ToSlice()); err != nil {
		logger.Error("Failed to batch load users for tickets",
			logger.Int("user_count", userIDCollector.Count()),
			logger.ErrorField(err))
	}

	// Second pass: populate user data from cache
	for _, response := range responses {
		// Populate user data
		response.User = userLoader.GetUser(response.UserID)

		// Populate assigned user data if available
		if response.AssignedToID != nil {
			response.AssignedTo = userLoader.GetUser(*response.AssignedToID)
		}

		// Populate resolved by user data if available
		if response.ResolvedByID != nil {
			response.ResolvedBy = userLoader.GetUser(*response.ResolvedByID)
		}
	}

	logger.Debug("Batch loaded user data for tickets",
		logger.Int("ticket_count", len(tickets)),
		logger.Int("unique_users", userIDCollector.Count()),
		logger.Int("cached_users", userLoader.CacheSize()))

	response.Paginated(c, "Tickets retrieved successfully", responses, page, limit, total, "/api/v1/admin/tickets")
}

// GetTicket godoc
// @Summary Get ticket details
// @Description Get detailed ticket information including messages (Admin only)
// @Tags Admin-Ticket-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path uint true "Ticket ID"
// @Success 200 {object} dto.TicketResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Router /admin/tickets/{id} [get]
func (h *AdminTicketBaseHandler) GetTicket(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid ticket ID")
		return
	}

	ticket, err := h.ticketService.GetTicket(c.Request.Context(), uint(id))
	if err != nil {
		logger.Error("Admin failed to get ticket",
			logger.Uint("ticket_id", uint(id)),
			logger.ErrorField(err))
		response.NotFound(c, "Ticket not found")
		return
	}

	// Convert to response format and populate user data
	ticketResponse := dto.ToTicketResponse(ticket)

	// Collect all user IDs that need to be loaded
	userIDCollector := handlers.NewUserIDCollector()
	userIDCollector.Add(ticket.UserID)
	userIDCollector.AddPtr(ticket.AssignedToID)
	userIDCollector.AddPtr(ticket.ResolvedByID)

	// Collect user IDs from messages
	for i := range ticketResponse.Messages {
		userIDCollector.Add(ticketResponse.Messages[i].UserID)
	}

	// Batch load all users
	userLoader := handlers.NewBatchUserLoader(h.userService)
	if err := userLoader.LoadUsers(c.Request.Context(), userIDCollector.ToSlice()); err != nil {
		logger.Error("Failed to batch load users for ticket details",
			logger.Int("user_count", userIDCollector.Count()),
			logger.ErrorField(err))
	}

	// Populate user data from cache
	ticketResponse.User = userLoader.GetUser(ticket.UserID)

	// Populate assigned user data if available
	if ticket.AssignedToID != nil {
		ticketResponse.AssignedTo = userLoader.GetUser(*ticket.AssignedToID)
	}

	// Populate resolved by user data if available
	if ticket.ResolvedByID != nil {
		ticketResponse.ResolvedBy = userLoader.GetUser(*ticket.ResolvedByID)
	}

	// Populate message user data
	for i := range ticketResponse.Messages {
		ticketResponse.Messages[i].User = userLoader.GetUser(ticketResponse.Messages[i].UserID)
	}

	logger.Debug("Batch loaded user data for ticket details",
		logger.Int("unique_users", userIDCollector.Count()),
		logger.Int("cached_users", userLoader.CacheSize()))

	response.OK(c, ticketResponse)
}

// UpdateTicket godoc
// @Summary Update ticket
// @Description Update ticket information (Admin only)
// @Tags Admin-Ticket-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path uint true "Ticket ID"
// @Param ticket body dto.AdminUpdateTicketRequest true "Ticket update data"
// @Success 200 {object} dto.TicketResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/tickets/{id} [put]
func (h *AdminTicketBaseHandler) UpdateTicket(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid ticket ID")
		return
	}

	var req dto.AdminUpdateTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Create update request for service layer
	updateReq := &dto.UpdateTicketRequest{
		Title:       req.Title,
		Description: req.Description,
		Category:    req.Category,
		Priority:    req.Priority,
		Status:      req.Status,
		Tags:        req.Tags,
		Metadata:    req.Metadata,
	}

	ticket, err := h.ticketService.UpdateTicket(c.Request.Context(), uint(id), updateReq)
	if err != nil {
		logger.Error("Admin failed to update ticket",
			logger.Uint("ticket_id", uint(id)),
			logger.ErrorField(err))

		convertedErr := sharedErrors.ConvertTicketErrorUint(err, uint(id))
		if sharedErrors.IsTicketNotFound(convertedErr) {
			response.NotFound(c, "Ticket not found")
		} else {
			response.InternalServerError(c, "Failed to update ticket")
		}
		return
	}

	// Convert to response format and populate user data
	ticketResponse := dto.ToTicketResponse(ticket)
	if user, err := h.userService.GetUserByID(c.Request.Context(), ticket.UserID); err == nil {
		ticketResponse.User = handlers.ConvertUserToBasicDTO(user)
	}

	logger.Info("Admin updated ticket successfully",
		logger.Uint("ticket_id", uint(id)))

	response.OK(c, ticketResponse)
}

// DeleteTicket godoc
// @Summary Delete ticket
// @Description Soft delete a support ticket (Admin only)
// @Tags Admin-Ticket-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path uint true "Ticket ID"
// @Success 200 {object} string
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/tickets/{id} [delete]
func (h *AdminTicketBaseHandler) DeleteTicket(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid ticket ID")
		return
	}

	err = h.ticketService.DeleteTicket(c.Request.Context(), uint(id))
	if err != nil {
		logger.Error("Admin failed to delete ticket",
			logger.Uint("ticket_id", uint(id)),
			logger.ErrorField(err))

		convertedErr := sharedErrors.ConvertTicketErrorUint(err, uint(id))
		if sharedErrors.IsTicketNotFound(convertedErr) {
			response.NotFound(c, "Ticket not found")
		} else {
			response.InternalServerError(c, "Failed to delete ticket")
		}
		return
	}

	logger.Info("Admin deleted ticket successfully",
		logger.Uint("ticket_id", uint(id)))

	response.OK(c, nil)
}
