package handlers

import (
	"strconv"
	"strings"
	"time"

	"linke/internal/domains/ticket/constants"
	"linke/internal/domains/ticket/dto"
	"linke/internal/domains/ticket/entities"
	ticketInterfaces "linke/internal/domains/ticket/usecases/interfaces"
	userInterfaces "linke/internal/domains/user/usecases/interfaces"
	sharedDTO "linke/internal/shared/dto"
	"linke/internal/shared/logger"
	"linke/internal/shared/response"

	"github.com/gin-gonic/gin"
)

// AdminTicketHandler provides comprehensive admin ticket management functionality
type AdminTicketHandler struct {
	ticketService        ticketInterfaces.TicketService
	ticketMessageService ticketInterfaces.TicketMessageService
	userService          userInterfaces.UserService
}

// NewAdminTicketHandler creates a new admin ticket handler
func NewAdminTicketHandler(
	ticketService ticketInterfaces.TicketService,
	ticketMessageService ticketInterfaces.TicketMessageService,
	userService userInterfaces.UserService,
) *AdminTicketHandler {
	return &AdminTicketHandler{
		ticketService:        ticketService,
		ticketMessageService: ticketMessageService,
		userService:          userService,
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
// @Success 201 {object} response.StandardResponse{data=entities.TicketResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/tickets [post]
func (h *AdminTicketHandler) CreateTicket(c *gin.Context) {
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
			logger.Error2("error", err))
		response.NotFound(c, "User not found")
		return
	}

	// Verify assigned user if specified
	if req.AssignedToID != nil {
		assignedUser, err := h.userService.GetUserByID(c.Request.Context(), *req.AssignedToID)
		if err != nil {
			logger.Error("Failed to verify assigned user for ticket creation",
				logger.Uint("assigned_to_id", *req.AssignedToID),
				logger.Error2("error", err))
			response.NotFound(c, "Assigned user not found")
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
			logger.Error2("error", err))
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
				logger.Error2("error", err))
			// Continue with unassigned ticket rather than failing
		}
	}

	// Populate user data in response
	ticketResponse := ticket.ToResponse()
	ticketResponse.User = &sharedDTO.UserBasicDTO{
		ID:       user.ID,
		Email:    user.Email,
		Username: user.Username,
		Name:     user.Name,
		Avatar:   user.Avatar,
		Provider: user.Provider,
		Status:   user.Status,
		Role:     user.Role,
	}

	logger.Info("Admin created ticket successfully",
		logger.Uint("ticket_id", ticket.ID),
		logger.String("ticket_no", ticket.TicketNo),
		logger.Uint("user_id", req.UserID))

	response.Created(c, ticketResponse)
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
// @Success 200 {object} response.StandardListResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/tickets [get]
func (h *AdminTicketHandler) ListTickets(c *gin.Context) {
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
		logger.Error("Admin failed to list tickets", logger.Error2("error", err))
		response.InternalServerError(c, "Failed to list tickets")
		return
	}

	// Convert to response format and populate user data
	responses := make([]*entities.TicketResponse, len(tickets))
	for i, ticket := range tickets {
		responses[i] = ticket.ToResponse()

		// Populate user data if available
		if user, err := h.userService.GetUserByID(c.Request.Context(), ticket.UserID); err == nil {
			responses[i].User = &sharedDTO.UserBasicDTO{
				ID:       user.ID,
				Email:    user.Email,
				Username: user.Username,
				Name:     user.Name,
				Avatar:   user.Avatar,
				Provider: user.Provider,
				Status:   user.Status,
				Role:     user.Role,
			}
		}

		// Populate assigned user data if available
		if ticket.AssignedToID != nil {
			if assignedUser, err := h.userService.GetUserByID(c.Request.Context(), *ticket.AssignedToID); err == nil {
				responses[i].AssignedTo = &sharedDTO.UserBasicDTO{
					ID:       assignedUser.ID,
					Email:    assignedUser.Email,
					Username: assignedUser.Username,
					Name:     assignedUser.Name,
					Avatar:   assignedUser.Avatar,
					Provider: assignedUser.Provider,
					Status:   assignedUser.Status,
					Role:     assignedUser.Role,
				}
			}
		}

		// Populate resolved by user data if available
		if ticket.ResolvedByID != nil {
			if resolvedUser, err := h.userService.GetUserByID(c.Request.Context(), *ticket.ResolvedByID); err == nil {
				responses[i].ResolvedBy = &sharedDTO.UserBasicDTO{
					ID:       resolvedUser.ID,
					Email:    resolvedUser.Email,
					Username: resolvedUser.Username,
					Name:     resolvedUser.Name,
					Avatar:   resolvedUser.Avatar,
					Provider: resolvedUser.Provider,
					Status:   resolvedUser.Status,
					Role:     resolvedUser.Role,
				}
			}
		}
	}

	response.SuccessList(c, responses, page, limit, total)
}

// GetTicket godoc
// @Summary Get ticket details
// @Description Get detailed ticket information including messages (Admin only)
// @Tags Admin-Ticket-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path uint true "Ticket ID"
// @Success 200 {object} response.StandardResponse{data=entities.TicketResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Router /admin/tickets/{id} [get]
func (h *AdminTicketHandler) GetTicket(c *gin.Context) {
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
			logger.Error2("error", err))
		response.NotFound(c, "Ticket not found")
		return
	}

	// Convert to response format and populate user data
	ticketResponse := ticket.ToResponse()

	// Populate user data
	if user, err := h.userService.GetUserByID(c.Request.Context(), ticket.UserID); err == nil {
		ticketResponse.User = &sharedDTO.UserBasicDTO{
			ID:       user.ID,
			Email:    user.Email,
			Username: user.Username,
			Name:     user.Name,
			Avatar:   user.Avatar,
			Provider: user.Provider,
			Status:   user.Status,
			Role:     user.Role,
		}
	}

	// Populate assigned user data if available
	if ticket.AssignedToID != nil {
		if assignedUser, err := h.userService.GetUserByID(c.Request.Context(), *ticket.AssignedToID); err == nil {
			ticketResponse.AssignedTo = &sharedDTO.UserBasicDTO{
				ID:       assignedUser.ID,
				Email:    assignedUser.Email,
				Username: assignedUser.Username,
				Name:     assignedUser.Name,
				Avatar:   assignedUser.Avatar,
				Provider: assignedUser.Provider,
				Status:   assignedUser.Status,
				Role:     assignedUser.Role,
			}
		}
	}

	// Populate resolved by user data if available
	if ticket.ResolvedByID != nil {
		if resolvedUser, err := h.userService.GetUserByID(c.Request.Context(), *ticket.ResolvedByID); err == nil {
			ticketResponse.ResolvedBy = &sharedDTO.UserBasicDTO{
				ID:       resolvedUser.ID,
				Email:    resolvedUser.Email,
				Username: resolvedUser.Username,
				Name:     resolvedUser.Name,
				Avatar:   resolvedUser.Avatar,
				Provider: resolvedUser.Provider,
				Status:   resolvedUser.Status,
				Role:     resolvedUser.Role,
			}
		}
	}

	// Populate message user data
	for i := range ticketResponse.Messages {
		if user, err := h.userService.GetUserByID(c.Request.Context(), ticketResponse.Messages[i].UserID); err == nil {
			ticketResponse.Messages[i].User = &sharedDTO.UserBasicDTO{
				ID:       user.ID,
				Email:    user.Email,
				Username: user.Username,
				Name:     user.Name,
				Avatar:   user.Avatar,
				Provider: user.Provider,
				Status:   user.Status,
				Role:     user.Role,
			}
		}
	}

	response.Success(c, ticketResponse)
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
// @Success 200 {object} response.StandardResponse{data=entities.TicketResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/tickets/{id} [put]
func (h *AdminTicketHandler) UpdateTicket(c *gin.Context) {
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
			logger.Error2("error", err))

		if strings.Contains(err.Error(), "not found") {
			response.NotFound(c, "Ticket not found")
		} else {
			response.InternalServerError(c, "Failed to update ticket")
		}
		return
	}

	// Convert to response format and populate user data
	ticketResponse := ticket.ToResponse()
	if user, err := h.userService.GetUserByID(c.Request.Context(), ticket.UserID); err == nil {
		ticketResponse.User = &sharedDTO.UserBasicDTO{
			ID:       user.ID,
			Email:    user.Email,
			Username: user.Username,
			Name:     user.Name,
			Avatar:   user.Avatar,
			Provider: user.Provider,
			Status:   user.Status,
			Role:     user.Role,
		}
	}

	logger.Info("Admin updated ticket successfully",
		logger.Uint("ticket_id", uint(id)))

	response.Success(c, ticketResponse)
}

// AssignTicket godoc
// @Summary Assign ticket
// @Description Assign ticket to an agent (Admin only)
// @Tags Admin-Ticket-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path uint true "Ticket ID"
// @Param assignment body dto.AssignTicketRequest true "Assignment data"
// @Success 200 {object} response.StandardResponse{data=entities.TicketResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/tickets/{id}/assign [post]
func (h *AdminTicketHandler) AssignTicket(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid ticket ID")
		return
	}

	var req dto.AssignTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Verify assigned user exists and is admin
	assignedUser, err := h.userService.GetUserByID(c.Request.Context(), req.AssignedToID)
	if err != nil {
		logger.Error("Failed to verify assigned user",
			logger.Uint("assigned_to_id", req.AssignedToID),
			logger.Error2("error", err))
		response.NotFound(c, "Assigned user not found")
		return
	}

	if assignedUser.Role != "admin" {
		response.BadRequest(c, "Assigned user must be an admin")
		return
	}

	// Assign ticket
	assignReq := &dto.AssignTicketRequest{
		AssignedToID: req.AssignedToID,
	}

	ticket, err := h.ticketService.AssignTicket(c.Request.Context(), uint(id), assignReq)
	if err != nil {
		logger.Error("Admin failed to assign ticket",
			logger.Uint("ticket_id", uint(id)),
			logger.Uint("assigned_to_id", req.AssignedToID),
			logger.Error2("error", err))

		if strings.Contains(err.Error(), "not found") {
			response.NotFound(c, "Ticket not found")
		} else {
			response.InternalServerError(c, "Failed to assign ticket")
		}
		return
	}

	// Add internal note if provided
	if req.Note != "" {
		noteReq := &dto.CreateTicketMessageRequest{
			Content:     req.Note,
			MessageType: constants.MessageTypeSystem,
			IsInternal:  true,
		}

		// Note: We're passing 0 as userID for system messages - this might need adjustment
		_, err = h.ticketMessageService.CreateTicketMessage(c.Request.Context(), ticket.ID, 0, noteReq)
		if err != nil {
			logger.Error("Failed to create assignment note",
				logger.Uint("ticket_id", ticket.ID),
				logger.Error2("error", err))
			// Continue without failing the assignment
		}
	}

	// Convert to response format and populate user data
	ticketResponse := ticket.ToResponse()
	if user, err := h.userService.GetUserByID(c.Request.Context(), ticket.UserID); err == nil {
		ticketResponse.User = &sharedDTO.UserBasicDTO{
			ID:       user.ID,
			Email:    user.Email,
			Username: user.Username,
			Name:     user.Name,
			Avatar:   user.Avatar,
			Provider: user.Provider,
			Status:   user.Status,
			Role:     user.Role,
		}
	}

	ticketResponse.AssignedTo = &sharedDTO.UserBasicDTO{
		ID:       assignedUser.ID,
		Email:    assignedUser.Email,
		Username: assignedUser.Username,
		Name:     assignedUser.Name,
		Avatar:   assignedUser.Avatar,
		Provider: assignedUser.Provider,
		Status:   assignedUser.Status,
		Role:     assignedUser.Role,
	}

	logger.Info("Admin assigned ticket successfully",
		logger.Uint("ticket_id", uint(id)),
		logger.Uint("assigned_to_id", req.AssignedToID))

	response.Success(c, ticketResponse)
}

// EscalateTicket godoc
// @Summary Escalate ticket
// @Description Escalate ticket to higher priority or different agent (Admin only)
// @Tags Admin-Ticket-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path uint true "Ticket ID"
// @Param escalation body dto.EscalateTicketRequest true "Escalation data"
// @Success 200 {object} response.StandardResponse{data=entities.TicketResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/tickets/{id}/escalate [post]
func (h *AdminTicketHandler) EscalateTicket(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid ticket ID")
		return
	}

	var req dto.EscalateTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Verify escalated user exists and is admin
	escalatedUser, err := h.userService.GetUserByID(c.Request.Context(), req.EscalatedToID)
	if err != nil {
		logger.Error("Failed to verify escalated user",
			logger.Uint("escalated_to_id", req.EscalatedToID),
			logger.Error2("error", err))
		response.NotFound(c, "Escalated user not found")
		return
	}

	if escalatedUser.Role != "admin" {
		response.BadRequest(c, "Escalated user must be an admin")
		return
	}

	// Get current ticket
	ticket, err := h.ticketService.GetTicket(c.Request.Context(), uint(id))
	if err != nil {
		logger.Error("Failed to get ticket for escalation",
			logger.Uint("ticket_id", uint(id)),
			logger.Error2("error", err))
		response.NotFound(c, "Ticket not found")
		return
	}

	// Update priority if specified
	updateReq := &dto.UpdateTicketRequest{}
	if req.Priority != nil {
		updateReq.Priority = req.Priority
	}

	// Update ticket if needed
	if req.Priority != nil {
		ticket, err = h.ticketService.UpdateTicket(c.Request.Context(), uint(id), updateReq)
		if err != nil {
			logger.Error("Failed to update ticket priority during escalation",
				logger.Uint("ticket_id", uint(id)),
				logger.Error2("error", err))
			response.InternalServerError(c, "Failed to escalate ticket")
			return
		}
	}

	// Reassign ticket
	assignReq := &dto.AssignTicketRequest{
		AssignedToID: req.EscalatedToID,
	}

	ticket, err = h.ticketService.AssignTicket(c.Request.Context(), uint(id), assignReq)
	if err != nil {
		logger.Error("Failed to reassign ticket during escalation",
			logger.Uint("ticket_id", uint(id)),
			logger.Uint("escalated_to_id", req.EscalatedToID),
			logger.Error2("error", err))
		response.InternalServerError(c, "Failed to escalate ticket")
		return
	}

	// Add escalation note
	escalationNote := "Ticket escalated. Reason: " + req.EscalationReason
	noteReq := &dto.CreateTicketMessageRequest{
		Content:     escalationNote,
		MessageType: constants.MessageTypeSystem,
		IsInternal:  true,
	}

	_, err = h.ticketMessageService.CreateTicketMessage(c.Request.Context(), ticket.ID, 0, noteReq)
	if err != nil {
		logger.Error("Failed to create escalation note",
			logger.Uint("ticket_id", ticket.ID),
			logger.Error2("error", err))
		// Continue without failing the escalation
	}

	// Convert to response format and populate user data
	ticketResponse := ticket.ToResponse()
	if user, err := h.userService.GetUserByID(c.Request.Context(), ticket.UserID); err == nil {
		ticketResponse.User = &sharedDTO.UserBasicDTO{
			ID:       user.ID,
			Email:    user.Email,
			Username: user.Username,
			Name:     user.Name,
			Avatar:   user.Avatar,
			Provider: user.Provider,
			Status:   user.Status,
			Role:     user.Role,
		}
	}

	ticketResponse.AssignedTo = &sharedDTO.UserBasicDTO{
		ID:       escalatedUser.ID,
		Email:    escalatedUser.Email,
		Username: escalatedUser.Username,
		Name:     escalatedUser.Name,
		Avatar:   escalatedUser.Avatar,
		Provider: escalatedUser.Provider,
		Status:   escalatedUser.Status,
		Role:     escalatedUser.Role,
	}

	logger.Info("Admin escalated ticket successfully",
		logger.Uint("ticket_id", uint(id)),
		logger.Uint("escalated_to_id", req.EscalatedToID))

	response.Success(c, ticketResponse)
}

// CloseTicket godoc
// @Summary Close ticket
// @Description Close a support ticket (Admin only)
// @Tags Admin-Ticket-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path uint true "Ticket ID"
// @Success 200 {object} response.StandardResponse{data=entities.TicketResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/tickets/{id}/close [post]
func (h *AdminTicketHandler) CloseTicket(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid ticket ID")
		return
	}

	ticket, err := h.ticketService.CloseTicket(c.Request.Context(), uint(id), "Ticket closed by admin")
	if err != nil {
		logger.Error("Admin failed to close ticket",
			logger.Uint("ticket_id", uint(id)),
			logger.Error2("error", err))

		if strings.Contains(err.Error(), "not found") {
			response.NotFound(c, "Ticket not found")
		} else {
			response.InternalServerError(c, "Failed to close ticket")
		}
		return
	}

	// Convert to response format and populate user data
	ticketResponse := ticket.ToResponse()
	if user, err := h.userService.GetUserByID(c.Request.Context(), ticket.UserID); err == nil {
		ticketResponse.User = &sharedDTO.UserBasicDTO{
			ID:       user.ID,
			Email:    user.Email,
			Username: user.Username,
			Name:     user.Name,
			Avatar:   user.Avatar,
			Provider: user.Provider,
			Status:   user.Status,
			Role:     user.Role,
		}
	}

	logger.Info("Admin closed ticket successfully",
		logger.Uint("ticket_id", uint(id)))

	response.Success(c, ticketResponse)
}

// ReopenTicket godoc
// @Summary Reopen ticket
// @Description Reopen a closed support ticket (Admin only)
// @Tags Admin-Ticket-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path uint true "Ticket ID"
// @Success 200 {object} response.StandardResponse{data=entities.TicketResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/tickets/{id}/reopen [post]
func (h *AdminTicketHandler) ReopenTicket(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid ticket ID")
		return
	}

	ticket, err := h.ticketService.ReopenTicket(c.Request.Context(), uint(id), "Ticket reopened by admin")
	if err != nil {
		logger.Error("Admin failed to reopen ticket",
			logger.Uint("ticket_id", uint(id)),
			logger.Error2("error", err))

		if strings.Contains(err.Error(), "not found") {
			response.NotFound(c, "Ticket not found")
		} else {
			response.InternalServerError(c, "Failed to reopen ticket")
		}
		return
	}

	// Convert to response format and populate user data
	ticketResponse := ticket.ToResponse()
	if user, err := h.userService.GetUserByID(c.Request.Context(), ticket.UserID); err == nil {
		ticketResponse.User = &sharedDTO.UserBasicDTO{
			ID:       user.ID,
			Email:    user.Email,
			Username: user.Username,
			Name:     user.Name,
			Avatar:   user.Avatar,
			Provider: user.Provider,
			Status:   user.Status,
			Role:     user.Role,
		}
	}

	logger.Info("Admin reopened ticket successfully",
		logger.Uint("ticket_id", uint(id)))

	response.Success(c, ticketResponse)
}

// GetTicketMessages godoc
// @Summary Get ticket messages
// @Description Get all messages for a ticket including internal notes (Admin only)
// @Tags Admin-Ticket-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path uint true "Ticket ID"
// @Param include_internal query bool false "Include internal notes" default(true)
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} response.StandardListResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Router /admin/tickets/{id}/messages [get]
func (h *AdminTicketHandler) GetTicketMessages(c *gin.Context) {
	idStr := c.Param("id")
	ticketID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid ticket ID")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	includeInternal := c.DefaultQuery("include_internal", "true") == "true"

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	// Verify ticket exists
	_, err = h.ticketService.GetTicket(c.Request.Context(), uint(ticketID))
	if err != nil {
		logger.Error("Failed to verify ticket for messages",
			logger.Uint("ticket_id", uint(ticketID)),
			logger.Error2("error", err))
		response.NotFound(c, "Ticket not found")
		return
	}

	req := &dto.GetTicketMessagesRequest{
		TicketID:        uint(ticketID),
		IncludeInternal: includeInternal,
		Limit:           limit,
		Offset:          offset,
	}

	messages, total, err := h.ticketMessageService.GetTicketMessages(c.Request.Context(), req)
	if err != nil {
		logger.Error("Admin failed to get ticket messages",
			logger.Uint("ticket_id", uint(ticketID)),
			logger.Error2("error", err))
		response.InternalServerError(c, "Failed to get messages")
		return
	}

	// Convert to response format and populate user data
	responses := make([]*entities.TicketMessageResponse, len(messages))
	for i, message := range messages {
		responses[i] = message.ToResponse()

		// Populate user data
		if user, err := h.userService.GetUserByID(c.Request.Context(), message.UserID); err == nil {
			responses[i].User = &sharedDTO.UserBasicDTO{
				ID:       user.ID,
				Email:    user.Email,
				Username: user.Username,
				Name:     user.Name,
				Avatar:   user.Avatar,
				Provider: user.Provider,
				Status:   user.Status,
				Role:     user.Role,
			}
		}
	}

	response.SuccessList(c, responses, page, limit, total)
}

// AddMessage godoc
// @Summary Add message to ticket
// @Description Add admin reply or internal note to ticket (Admin only)
// @Tags Admin-Ticket-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path uint true "Ticket ID"
// @Param message body dto.AdminTicketMessageRequest true "Message data"
// @Success 201 {object} response.StandardResponse{data=entities.TicketMessageResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/tickets/{id}/messages [post]
func (h *AdminTicketHandler) AddMessage(c *gin.Context) {
	idStr := c.Param("id")
	ticketID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid ticket ID")
		return
	}

	var req dto.AdminTicketMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Verify ticket exists
	_, err = h.ticketService.GetTicket(c.Request.Context(), uint(ticketID))
	if err != nil {
		logger.Error("Failed to verify ticket for message creation",
			logger.Uint("ticket_id", uint(ticketID)),
			logger.Error2("error", err))
		response.NotFound(c, "Ticket not found")
		return
	}

	// TODO: Get admin user ID from context
	// For now using 0, but this should be extracted from the authentication context
	adminUserID := uint(0) // This needs to be properly implemented

	// Set default message type
	messageType := req.MessageType
	if messageType == "" {
		messageType = "admin"
	}

	createReq := &dto.CreateTicketMessageRequest{
		Content:     req.Content,
		MessageType: messageType,
		IsInternal:  req.IsInternal,
		Attachments: req.Attachments,
		Metadata:    req.Metadata,
	}

	message, err := h.ticketMessageService.CreateTicketMessage(c.Request.Context(), uint(ticketID), adminUserID, createReq)
	if err != nil {
		logger.Error("Admin failed to create ticket message",
			logger.Uint("ticket_id", uint(ticketID)),
			logger.Error2("error", err))
		response.InternalServerError(c, "Failed to create message")
		return
	}

	// Convert to response format and populate user data
	messageResponse := message.ToResponse()
	if user, err := h.userService.GetUserByID(c.Request.Context(), message.UserID); err == nil {
		messageResponse.User = &sharedDTO.UserBasicDTO{
			ID:       user.ID,
			Email:    user.Email,
			Username: user.Username,
			Name:     user.Name,
			Avatar:   user.Avatar,
			Provider: user.Provider,
			Status:   user.Status,
			Role:     user.Role,
		}
	}

	logger.Info("Admin added message to ticket successfully",
		logger.Uint("ticket_id", uint(ticketID)),
		logger.Uint("message_id", message.ID))

	response.Created(c, messageResponse)
}

// GetMessage godoc
// @Summary Get message details
// @Description Get detailed message information (Admin only)
// @Tags Admin-Ticket-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path uint true "Ticket ID"
// @Param msg_id path uint true "Message ID"
// @Success 200 {object} response.StandardResponse{data=entities.TicketMessageResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Router /admin/tickets/{id}/messages/{msg_id} [get]
func (h *AdminTicketHandler) GetMessage(c *gin.Context) {
	msgIDStr := c.Param("msg_id")
	msgID, err := strconv.ParseUint(msgIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid message ID")
		return
	}

	message, err := h.ticketMessageService.GetTicketMessage(c.Request.Context(), uint(msgID))
	if err != nil {
		logger.Error("Admin failed to get message",
			logger.Uint("message_id", uint(msgID)),
			logger.Error2("error", err))
		response.NotFound(c, "Message not found")
		return
	}

	// Convert to response format and populate user data
	messageResponse := message.ToResponse()
	if user, err := h.userService.GetUserByID(c.Request.Context(), message.UserID); err == nil {
		messageResponse.User = &sharedDTO.UserBasicDTO{
			ID:       user.ID,
			Email:    user.Email,
			Username: user.Username,
			Name:     user.Name,
			Avatar:   user.Avatar,
			Provider: user.Provider,
			Status:   user.Status,
			Role:     user.Role,
		}
	}

	response.Success(c, messageResponse)
}

// UpdateMessage godoc
// @Summary Update message
// @Description Update message content or metadata (Admin only)
// @Tags Admin-Ticket-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path uint true "Ticket ID"
// @Param msg_id path uint true "Message ID"
// @Param message body dto.UpdateTicketMessageRequest true "Message update data"
// @Success 200 {object} response.StandardResponse{data=entities.TicketMessageResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/tickets/{id}/messages/{msg_id} [put]
func (h *AdminTicketHandler) UpdateMessage(c *gin.Context) {
	msgIDStr := c.Param("msg_id")
	msgID, err := strconv.ParseUint(msgIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid message ID")
		return
	}

	var req dto.UpdateTicketMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	message, err := h.ticketMessageService.UpdateTicketMessage(c.Request.Context(), uint(msgID), &req)
	if err != nil {
		logger.Error("Admin failed to update message",
			logger.Uint("message_id", uint(msgID)),
			logger.Error2("error", err))

		if strings.Contains(err.Error(), "not found") {
			response.NotFound(c, "Message not found")
		} else {
			response.InternalServerError(c, "Failed to update message")
		}
		return
	}

	// Convert to response format and populate user data
	messageResponse := message.ToResponse()
	if user, err := h.userService.GetUserByID(c.Request.Context(), message.UserID); err == nil {
		messageResponse.User = &sharedDTO.UserBasicDTO{
			ID:       user.ID,
			Email:    user.Email,
			Username: user.Username,
			Name:     user.Name,
			Avatar:   user.Avatar,
			Provider: user.Provider,
			Status:   user.Status,
			Role:     user.Role,
		}
	}

	logger.Info("Admin updated message successfully",
		logger.Uint("message_id", uint(msgID)))

	response.Success(c, messageResponse)
}

// DeleteMessage godoc
// @Summary Delete message
// @Description Soft delete a message (Admin only)
// @Tags Admin-Ticket-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path uint true "Ticket ID"
// @Param msg_id path uint true "Message ID"
// @Success 200 {object} response.MessageOnlyResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/tickets/{id}/messages/{msg_id} [delete]
func (h *AdminTicketHandler) DeleteMessage(c *gin.Context) {
	msgIDStr := c.Param("msg_id")
	msgID, err := strconv.ParseUint(msgIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid message ID")
		return
	}

	err = h.ticketMessageService.DeleteTicketMessage(c.Request.Context(), uint(msgID))
	if err != nil {
		logger.Error("Admin failed to delete message",
			logger.Uint("message_id", uint(msgID)),
			logger.Error2("error", err))

		if strings.Contains(err.Error(), "not found") {
			response.NotFound(c, "Message not found")
		} else {
			response.InternalServerError(c, "Failed to delete message")
		}
		return
	}

	logger.Info("Admin deleted message successfully",
		logger.Uint("message_id", uint(msgID)))

	response.SuccessWithMessage(c, "Message deleted successfully", nil)
}

// AddInternalNote godoc
// @Summary Add internal note
// @Description Add internal note to ticket (Admin only)
// @Tags Admin-Ticket-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path uint true "Ticket ID"
// @Param note body dto.AdminTicketMessageRequest true "Internal note data"
// @Success 201 {object} response.StandardResponse{data=entities.TicketMessageResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/tickets/{id}/notes [post]
func (h *AdminTicketHandler) AddInternalNote(c *gin.Context) {
	idStr := c.Param("id")
	ticketID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid ticket ID")
		return
	}

	var req dto.AdminTicketMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Verify ticket exists
	_, err = h.ticketService.GetTicket(c.Request.Context(), uint(ticketID))
	if err != nil {
		logger.Error("Failed to verify ticket for internal note",
			logger.Uint("ticket_id", uint(ticketID)),
			logger.Error2("error", err))
		response.NotFound(c, "Ticket not found")
		return
	}

	// TODO: Get admin user ID from context
	adminUserID := uint(0)

	// Use internal message service method if available, otherwise create regular message marked as internal
	message, err := h.ticketMessageService.CreateInternalMessage(c.Request.Context(), uint(ticketID), adminUserID, req.Content)
	if err != nil {
		logger.Error("Admin failed to create internal note",
			logger.Uint("ticket_id", uint(ticketID)),
			logger.Error2("error", err))
		response.InternalServerError(c, "Failed to create internal note")
		return
	}

	// Convert to response format and populate user data
	messageResponse := message.ToResponse()
	if user, err := h.userService.GetUserByID(c.Request.Context(), message.UserID); err == nil {
		messageResponse.User = &sharedDTO.UserBasicDTO{
			ID:       user.ID,
			Email:    user.Email,
			Username: user.Username,
			Name:     user.Name,
			Avatar:   user.Avatar,
			Provider: user.Provider,
			Status:   user.Status,
			Role:     user.Role,
		}
	}

	logger.Info("Admin added internal note successfully",
		logger.Uint("ticket_id", uint(ticketID)),
		logger.Uint("message_id", message.ID))

	response.Created(c, messageResponse)
}

// SearchTickets godoc
// @Summary Search tickets
// @Description Advanced ticket search with multiple filters (Admin only)
// @Tags Admin-Ticket-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param query query string false "Search query" example("login issue")
// @Param user_id query uint false "Filter by user ID" example(123)
// @Param assigned_to_id query uint false "Filter by assigned agent ID" example(456)
// @Param status query string false "Filter by status" Enums(open,in_progress,pending,resolved,closed) example(open)
// @Param priority query string false "Filter by priority" Enums(low,normal,high,urgent,critical) example(high)
// @Param category query string false "Filter by category" Enums(general,technical,billing,account,feature,bug,subscription,payment) example(technical)
// @Param created_after query string false "Created after date" format(date) example("2024-01-01")
// @Param created_before query string false "Created before date" format(date) example("2024-12-31")
// @Param tags query string false "Filter by tags" example("urgent,billing")
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} response.SearchResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/tickets/search [get]
func (h *AdminTicketHandler) SearchTickets(c *gin.Context) {
	var req dto.TicketSearchRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Set defaults
	if req.Limit <= 0 || req.Limit > 100 {
		req.Limit = 20
	}
	if req.Offset < 0 {
		req.Offset = 0
	}

	// Convert to service request format
	serviceReq := &dto.GetTicketsRequest{
		UserID:       req.UserID,
		AssignedToID: req.AssignedToID,
		Status:       req.Status,
		Priority:     req.Priority,
		Category:     req.Category,
		Search:       req.Query,
		Limit:        req.Limit,
		Offset:       req.Offset,
	}

	tickets, total, err := h.ticketService.GetTickets(c.Request.Context(), serviceReq)
	if err != nil {
		logger.Error("Admin failed to search tickets",
			logger.String("query", req.Query),
			logger.Error2("error", err))
		response.InternalServerError(c, "Failed to search tickets")
		return
	}

	// Convert to response format and populate user data
	responses := make([]*entities.TicketResponse, len(tickets))
	for i, ticket := range tickets {
		responses[i] = ticket.ToResponse()

		// Populate user data
		if user, err := h.userService.GetUserByID(c.Request.Context(), ticket.UserID); err == nil {
			responses[i].User = &sharedDTO.UserBasicDTO{
				ID:       user.ID,
				Email:    user.Email,
				Username: user.Username,
				Name:     user.Name,
				Avatar:   user.Avatar,
				Provider: user.Provider,
				Status:   user.Status,
				Role:     user.Role,
			}
		}

		// Populate assigned user data if available
		if ticket.AssignedToID != nil {
			if assignedUser, err := h.userService.GetUserByID(c.Request.Context(), *ticket.AssignedToID); err == nil {
				responses[i].AssignedTo = &sharedDTO.UserBasicDTO{
					ID:       assignedUser.ID,
					Email:    assignedUser.Email,
					Username: assignedUser.Username,
					Name:     assignedUser.Name,
					Avatar:   assignedUser.Avatar,
					Provider: assignedUser.Provider,
					Status:   assignedUser.Status,
					Role:     assignedUser.Role,
				}
			}
		}
	}

	page := (req.Offset / req.Limit) + 1
	response.SuccessListWithExtra(c, "Search completed", responses, page, req.Limit, total, gin.H{
		"query": req.Query,
		"filters": gin.H{
			"status":   req.Status,
			"priority": req.Priority,
			"category": req.Category,
			"tags":     req.Tags,
		},
	})
}

// GetStatistics godoc
// @Summary Get ticket statistics
// @Description Get comprehensive ticket statistics (Admin only)
// @Tags Admin-Ticket-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.StandardResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/tickets/statistics [get]
func (h *AdminTicketHandler) GetStatistics(c *gin.Context) {
	stats, err := h.ticketService.GetTicketStatistics(c.Request.Context(), "", "")
	if err != nil {
		logger.Error("Admin failed to get ticket statistics", logger.Error2("error", err))
		response.InternalServerError(c, "Failed to get ticket statistics")
		return
	}

	response.Success(c, stats)
}

// GetAnalytics godoc
// @Summary Get ticket analytics
// @Description Get detailed ticket analytics with time-based grouping (Admin only)
// @Tags Admin-Ticket-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param start_date query string false "Start date" format(date) example("2024-01-01")
// @Param end_date query string false "End date" format(date) example("2024-12-31")
// @Param group_by query string false "Group by period" Enums(day,week,month,agent,category,priority) example("day")
// @Param agent_id query uint false "Filter by agent ID" example(456)
// @Param category query string false "Filter by category" Enums(general,technical,billing,account,feature,bug,subscription,payment) example(technical)
// @Param priority query string false "Filter by priority" Enums(low,normal,high,urgent,critical) example(high)
// @Success 200 {object} response.StandardResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/tickets/analytics [get]
func (h *AdminTicketHandler) GetAnalytics(c *gin.Context) {
	var req dto.TicketAnalyticsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Set default date range if not provided
	if req.StartDate == "" {
		req.StartDate = time.Now().AddDate(0, -1, 0).Format("2006-01-02") // Last month
	}
	if req.EndDate == "" {
		req.EndDate = time.Now().Format("2006-01-02") // Today
	}

	// Get basic statistics for the period
	stats, err := h.ticketService.GetTicketStatistics(c.Request.Context(), req.StartDate, req.EndDate)
	if err != nil {
		logger.Error("Admin failed to get ticket analytics", logger.Error2("error", err))
		response.InternalServerError(c, "Failed to get ticket analytics")
		return
	}

	// Add time-based analytics
	analytics := gin.H{
		"basic_stats": stats,
		"period": gin.H{
			"start_date": req.StartDate,
			"end_date":   req.EndDate,
			"group_by":   req.GroupBy,
		},
		"filters": gin.H{
			"agent_id": req.AgentID,
			"category": req.Category,
			"priority": req.Priority,
		},
	}

	// Add agent-specific stats if requested
	if req.AgentID != nil {
		agentStats, err := h.ticketService.GetAgentTicketStatistics(c.Request.Context(), *req.AgentID)
		if err == nil {
			analytics["agent_stats"] = agentStats
		}
	}

	response.Success(c, analytics)
}

// GetAgents godoc
// @Summary List available agents
// @Description Get list of admin users who can be assigned tickets (Admin only)
// @Tags Admin-Ticket-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} response.StandardListResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/tickets/agents [get]
func (h *AdminTicketHandler) GetAgents(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	// Get admin users from user service
	// TODO: Add a method to get users by role to the user service interface
	users, _, err := h.userService.ListUsers(c.Request.Context(), limit, offset)
	if err != nil {
		logger.Error("Admin failed to get agent list", logger.Error2("error", err))
		response.InternalServerError(c, "Failed to get agents")
		return
	}

	// Filter only admin users
	agents := make([]*sharedDTO.UserBasicDTO, 0)
	for _, user := range users {
		if user.Role == "admin" {
			agents = append(agents, &sharedDTO.UserBasicDTO{
				ID:       user.ID,
				Email:    user.Email,
				Username: user.Username,
				Name:     user.Name,
				Avatar:   user.Avatar,
				Provider: user.Provider,
				Status:   user.Status,
				Role:     user.Role,
			})
		}
	}

	response.SuccessList(c, agents, page, limit, int64(len(agents)))
}

// BulkAssignTickets godoc
// @Summary Bulk assign tickets
// @Description Assign multiple tickets to an agent (Admin only)
// @Tags Admin-Ticket-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param assignment body dto.BulkTicketActionRequest true "Bulk assignment data"
// @Success 200 {object} response.StandardResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/tickets/bulk/assign [post]
func (h *AdminTicketHandler) BulkAssignTickets(c *gin.Context) {
	var req dto.BulkTicketActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if req.Action != "assign" {
		response.BadRequest(c, "Invalid action for bulk assign endpoint")
		return
	}

	if req.AssignedToID == nil {
		response.BadRequest(c, "assigned_to_id is required for bulk assign")
		return
	}

	// Verify assigned user exists and is admin
	assignedUser, err := h.userService.GetUserByID(c.Request.Context(), *req.AssignedToID)
	if err != nil {
		response.NotFound(c, "Assigned user not found")
		return
	}

	if assignedUser.Role != "admin" {
		response.BadRequest(c, "Assigned user must be an admin")
		return
	}

	// Bulk assign tickets
	err = h.ticketService.BulkAssignTickets(c.Request.Context(), req.TicketIDs, *req.AssignedToID)
	if err != nil {
		logger.Error("Admin failed to bulk assign tickets",
			logger.Any("ticket_ids", req.TicketIDs),
			logger.Uint("assigned_to_id", *req.AssignedToID),
			logger.Error2("error", err))
		response.InternalServerError(c, "Failed to bulk assign tickets")
		return
	}

	logger.Info("Admin bulk assigned tickets successfully",
		logger.Any("ticket_ids", req.TicketIDs),
		logger.Uint("assigned_to_id", *req.AssignedToID))

	response.SuccessWithMessage(c, "Tickets assigned successfully", gin.H{
		"ticket_count": len(req.TicketIDs),
		"assigned_to":  assignedUser.Name,
		"reason":       req.Reason,
	})
}

// BulkUpdateStatus godoc
// @Summary Bulk update ticket status
// @Description Update status of multiple tickets (Admin only)
// @Tags Admin-Ticket-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param update body dto.BulkTicketActionRequest true "Bulk status update data"
// @Success 200 {object} response.StandardResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/tickets/bulk/status [post]
func (h *AdminTicketHandler) BulkUpdateStatus(c *gin.Context) {
	var req dto.BulkTicketActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if req.Action != "update_status" {
		response.BadRequest(c, "Invalid action for bulk status update endpoint")
		return
	}

	if req.Status == nil {
		response.BadRequest(c, "status is required for bulk status update")
		return
	}

	// Bulk update ticket status
	err := h.ticketService.BulkUpdateTicketStatus(c.Request.Context(), req.TicketIDs, *req.Status)
	if err != nil {
		logger.Error("Admin failed to bulk update ticket status",
			logger.Any("ticket_ids", req.TicketIDs),
			logger.String("status", *req.Status),
			logger.Error2("error", err))
		response.InternalServerError(c, "Failed to bulk update status")
		return
	}

	logger.Info("Admin bulk updated ticket status successfully",
		logger.Any("ticket_ids", req.TicketIDs),
		logger.String("status", *req.Status))

	response.SuccessWithMessage(c, "Ticket status updated successfully", gin.H{
		"ticket_count": len(req.TicketIDs),
		"new_status":   *req.Status,
		"reason":       req.Reason,
	})
}

// BulkCloseTickets godoc
// @Summary Bulk close tickets
// @Description Close multiple tickets (Admin only)
// @Tags Admin-Ticket-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param closure body dto.BulkTicketActionRequest true "Bulk closure data"
// @Success 200 {object} response.StandardResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/tickets/bulk/close [post]
func (h *AdminTicketHandler) BulkCloseTickets(c *gin.Context) {
	var req dto.BulkTicketActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if req.Action != "close" {
		response.BadRequest(c, "Invalid action for bulk close endpoint")
		return
	}

	// Bulk close tickets by updating status to closed
	err := h.ticketService.BulkUpdateTicketStatus(c.Request.Context(), req.TicketIDs, constants.TicketStatusClosed)
	if err != nil {
		logger.Error("Admin failed to bulk close tickets",
			logger.Any("ticket_ids", req.TicketIDs),
			logger.Error2("error", err))
		response.InternalServerError(c, "Failed to bulk close tickets")
		return
	}

	logger.Info("Admin bulk closed tickets successfully",
		logger.Any("ticket_ids", req.TicketIDs))

	response.SuccessWithMessage(c, "Tickets closed successfully", gin.H{
		"ticket_count": len(req.TicketIDs),
		"reason":       req.Reason,
	})
}

// DeleteTicket godoc
// @Summary Delete ticket
// @Description Soft delete a support ticket (Admin only)
// @Tags Admin-Ticket-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path uint true "Ticket ID"
// @Success 200 {object} response.MessageOnlyResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/tickets/{id} [delete]
func (h *AdminTicketHandler) DeleteTicket(c *gin.Context) {
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
			logger.Error2("error", err))

		if strings.Contains(err.Error(), "not found") {
			response.NotFound(c, "Ticket not found")
		} else {
			response.InternalServerError(c, "Failed to delete ticket")
		}
		return
	}

	logger.Info("Admin deleted ticket successfully",
		logger.Uint("ticket_id", uint(id)))

	response.SuccessWithMessage(c, "Ticket deleted successfully", nil)
}
