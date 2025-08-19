package handlers

import (
	"strconv"

	"linke/internal/domains/ticket/constants"
	"linke/internal/domains/ticket/dto"
	sharedDTO "linke/internal/shared/dto"
	sharedErrors "linke/internal/shared/errors"
	"linke/internal/shared/handlers"
	"linke/internal/shared/logger"
	"linke/internal/shared/response"

	"github.com/gin-gonic/gin"
)

// AdminTicketStatusHandler handles ticket status and assignment operations
type AdminTicketStatusHandler struct {
	*AdminTicketHandlerBase
}

// NewAdminTicketStatusHandler creates a new admin ticket status handler
func NewAdminTicketStatusHandler(base *AdminTicketHandlerBase) *AdminTicketStatusHandler {
	return &AdminTicketStatusHandler{
		AdminTicketHandlerBase: base,
	}
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
// @Success 200 {object} dto.TicketResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/tickets/{id}/assign [post]
func (h *AdminTicketStatusHandler) AssignTicket(c *gin.Context) {
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
			logger.ErrorField(err))
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
			logger.ErrorField(err))

		convertedErr := sharedErrors.ConvertTicketErrorUint(err, uint(id))
		if sharedErrors.IsTicketNotFound(convertedErr) {
			response.NotFound(c, "Ticket not found")
		} else {
			response.InternalServerError(c, "Failed to assign ticket")
		}
		return
	}

	// Add internal note if provided
	if req.Note != "" {
		// Get admin user ID from context
		adminUserID, err := h.getAdminUserFromContext(c)
		if err != nil {
			logger.Error("Failed to get admin user from context for assignment note",
				logger.Uint("ticket_id", ticket.ID),
				logger.ErrorField(err))
			// Continue without failing the assignment since the main operation succeeded
		} else {
			noteReq := &dto.CreateTicketMessageRequest{
				Content:     req.Note,
				MessageType: constants.MessageTypeSystem,
				IsInternal:  true,
			}

			_, err = h.ticketMessageService.CreateTicketMessage(c.Request.Context(), ticket.ID, adminUserID, noteReq)
			if err != nil {
				logger.Error("Failed to create assignment note",
					logger.Uint("ticket_id", ticket.ID),
					logger.Uint("admin_user_id", adminUserID),
					logger.ErrorField(err))
				// Continue without failing the assignment
			}
		}
	}

	// Convert to response format and populate user data
	ticketResponse := dto.ToTicketResponse(ticket)
	if user, err := h.userService.GetUserByID(c.Request.Context(), ticket.UserID); err == nil {
		ticketResponse.User = handlers.ConvertUserToBasicDTO(user)
	}

	ticketResponse.AssignedTo = handlers.ConvertUserToBasicDTO(assignedUser)

	logger.Info("Admin assigned ticket successfully",
		logger.Uint("ticket_id", uint(id)),
		logger.Uint("assigned_to_id", req.AssignedToID))

	response.OK(c, ticketResponse)
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
// @Success 200 {object} dto.TicketResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/tickets/{id}/escalate [post]
func (h *AdminTicketStatusHandler) EscalateTicket(c *gin.Context) {
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
			logger.ErrorField(err))
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
			logger.ErrorField(err))
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
				logger.ErrorField(err))
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
			logger.ErrorField(err))
		response.InternalServerError(c, "Failed to escalate ticket")
		return
	}

	// Add escalation note
	escalationNote := "Ticket escalated. Reason: " + req.EscalationReason

	// Get admin user ID from context for the escalation note
	adminUserID, err := h.getAdminUserFromContext(c)
	if err != nil {
		logger.Error("Failed to get admin user from context for escalation note",
			logger.Uint("ticket_id", ticket.ID),
			logger.ErrorField(err))
		// Continue without failing the escalation since the main operation succeeded
	} else {
		noteReq := &dto.CreateTicketMessageRequest{
			Content:     escalationNote,
			MessageType: constants.MessageTypeSystem,
			IsInternal:  true,
		}

		_, err = h.ticketMessageService.CreateTicketMessage(c.Request.Context(), ticket.ID, adminUserID, noteReq)
		if err != nil {
			logger.Error("Failed to create escalation note",
				logger.Uint("ticket_id", ticket.ID),
				logger.Uint("admin_user_id", adminUserID),
				logger.ErrorField(err))
			// Continue without failing the escalation
		}
	}

	// Convert to response format and populate user data
	ticketResponse := dto.ToTicketResponse(ticket)
	if user, err := h.userService.GetUserByID(c.Request.Context(), ticket.UserID); err == nil {
		ticketResponse.User = handlers.ConvertUserToBasicDTO(user)
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

	response.OK(c, ticketResponse)
}

// CloseTicket godoc
// @Summary Close ticket
// @Description Close a support ticket (Admin only)
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
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/tickets/{id}/close [post]
func (h *AdminTicketStatusHandler) CloseTicket(c *gin.Context) {
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
			logger.ErrorField(err))

		convertedErr := sharedErrors.ConvertTicketErrorUint(err, uint(id))
		if sharedErrors.IsTicketNotFound(convertedErr) {
			response.NotFound(c, "Ticket not found")
		} else {
			response.InternalServerError(c, "Failed to close ticket")
		}
		return
	}

	// Convert to response format and populate user data
	ticketResponse := dto.ToTicketResponse(ticket)
	if user, err := h.userService.GetUserByID(c.Request.Context(), ticket.UserID); err == nil {
		ticketResponse.User = handlers.ConvertUserToBasicDTO(user)
	}

	logger.Info("Admin closed ticket successfully",
		logger.Uint("ticket_id", uint(id)))

	response.OK(c, ticketResponse)
}

// ReopenTicket godoc
// @Summary Reopen ticket
// @Description Reopen a closed support ticket (Admin only)
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
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/tickets/{id}/reopen [post]
func (h *AdminTicketStatusHandler) ReopenTicket(c *gin.Context) {
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
			logger.ErrorField(err))

		convertedErr := sharedErrors.ConvertTicketErrorUint(err, uint(id))
		if sharedErrors.IsTicketNotFound(convertedErr) {
			response.NotFound(c, "Ticket not found")
		} else {
			response.InternalServerError(c, "Failed to reopen ticket")
		}
		return
	}

	// Convert to response format and populate user data
	ticketResponse := dto.ToTicketResponse(ticket)
	if user, err := h.userService.GetUserByID(c.Request.Context(), ticket.UserID); err == nil {
		ticketResponse.User = handlers.ConvertUserToBasicDTO(user)
	}

	logger.Info("Admin reopened ticket successfully",
		logger.Uint("ticket_id", uint(id)))

	response.OK(c, ticketResponse)
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
// @Success 200 {object} response.HALCollectionResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/tickets/agents [get]
func (h *AdminTicketStatusHandler) GetAgents(c *gin.Context) {
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
		logger.Error("Admin failed to get agent list", logger.ErrorField(err))
		response.InternalServerError(c, "Failed to get agents")
		return
	}

	// Filter only admin users
	agents := make([]*sharedDTO.UserBasicDTO, 0)
	for _, user := range users {
		if user.Role == "admin" {
			agents = append(agents, handlers.ConvertUserToBasicDTO(user))
		}
	}

	response.Paginated(c, "Agents retrieved successfully", agents, page, limit, int64(len(agents)), "/api/v1/admin/tickets/agents")
}
