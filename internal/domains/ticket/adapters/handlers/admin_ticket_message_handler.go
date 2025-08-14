package handlers

import (
	"fmt"
	"strconv"

	"linke/internal/domains/ticket/dto"
	sharedErrors "linke/internal/shared/errors"
	"linke/internal/shared/handlers"
	"linke/internal/shared/logger"
	"linke/internal/shared/response"

	"github.com/gin-gonic/gin"
)

// AdminTicketMessageHandler handles ticket message operations
type AdminTicketMessageHandler struct {
	*AdminTicketHandlerBase
}

// NewAdminTicketMessageHandler creates a new admin ticket message handler
func NewAdminTicketMessageHandler(base *AdminTicketHandlerBase) *AdminTicketMessageHandler {
	return &AdminTicketMessageHandler{
		AdminTicketHandlerBase: base,
	}
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
// @Success 200 {object} response.HALCollectionResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Router /admin/tickets/{id}/messages [get]
func (h *AdminTicketMessageHandler) GetTicketMessages(c *gin.Context) {
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
			logger.ErrorField(err))
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
			logger.ErrorField(err))
		response.InternalServerError(c, "Failed to get messages")
		return
	}

	// Convert to response format and batch load user data
	responses := make([]*dto.TicketMessageResponse, len(messages))
	userIDCollector := handlers.NewUserIDCollector()

	// First pass: convert messages and collect user IDs
	for i, message := range messages {
		responses[i] = dto.ToTicketMessageResponse(message)
		userIDCollector.Add(message.UserID)
	}

	// Batch load all users
	userLoader := handlers.NewBatchUserLoader(h.userService)
	if err := userLoader.LoadUsers(c.Request.Context(), userIDCollector.ToSlice()); err != nil {
		logger.Error("Failed to batch load users for ticket messages",
			logger.Int("user_count", userIDCollector.Count()),
			logger.ErrorField(err))
	}

	// Second pass: populate user data from cache
	for _, response := range responses {
		response.User = userLoader.GetUser(response.UserID)
	}

	response.Paginated(c, "Messages retrieved successfully", responses, page, limit, total, fmt.Sprintf("/api/v1/admin/tickets/%s/messages", idStr))
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
// @Success 201 {object} dto.TicketMessageResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/tickets/{id}/messages [post]
func (h *AdminTicketMessageHandler) AddMessage(c *gin.Context) {
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
			logger.ErrorField(err))
		response.NotFound(c, "Ticket not found")
		return
	}

	// Get admin user ID from context
	adminUserID, err := h.getAdminUserFromContext(c)
	if err != nil {
		logger.Error("Failed to get admin user from context for message creation",
			logger.Uint("ticket_id", uint(ticketID)),
			logger.ErrorField(err))
		response.Unauthorized(c, "Admin authentication required")
		return
	}

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
			logger.ErrorField(err))
		response.InternalServerError(c, "Failed to create message")
		return
	}

	// Convert to response format and populate user data
	messageResponse := dto.ToTicketMessageResponse(message)
	if user, err := h.userService.GetUserByID(c.Request.Context(), message.UserID); err == nil {
		messageResponse.User = handlers.ConvertUserToBasicDTO(user)
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
// @Success 200 {object} dto.TicketMessageResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Router /admin/tickets/{id}/messages/{msg_id} [get]
func (h *AdminTicketMessageHandler) GetMessage(c *gin.Context) {
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
			logger.ErrorField(err))
		response.NotFound(c, "Message not found")
		return
	}

	// Convert to response format and populate user data
	messageResponse := dto.ToTicketMessageResponse(message)
	if user, err := h.userService.GetUserByID(c.Request.Context(), message.UserID); err == nil {
		messageResponse.User = handlers.ConvertUserToBasicDTO(user)
	}

	response.OK(c, messageResponse)
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
// @Success 200 {object} dto.TicketMessageResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/tickets/{id}/messages/{msg_id} [put]
func (h *AdminTicketMessageHandler) UpdateMessage(c *gin.Context) {
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
			logger.ErrorField(err))

		convertedErr := sharedErrors.ConvertServiceError(err, "message", uint(msgID))
		if sharedErrors.IsTicketMessageNotFound(convertedErr) {
			response.NotFound(c, "Message not found")
		} else {
			response.InternalServerError(c, "Failed to update message")
		}
		return
	}

	// Convert to response format and populate user data
	messageResponse := dto.ToTicketMessageResponse(message)
	if user, err := h.userService.GetUserByID(c.Request.Context(), message.UserID); err == nil {
		messageResponse.User = handlers.ConvertUserToBasicDTO(user)
	}

	logger.Info("Admin updated message successfully",
		logger.Uint("message_id", uint(msgID)))

	response.OK(c, messageResponse)
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
// @Success 200 {object} string
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/tickets/{id}/messages/{msg_id} [delete]
func (h *AdminTicketMessageHandler) DeleteMessage(c *gin.Context) {
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
			logger.ErrorField(err))

		convertedErr := sharedErrors.ConvertServiceError(err, "message", uint(msgID))
		if sharedErrors.IsTicketMessageNotFound(convertedErr) {
			response.NotFound(c, "Message not found")
		} else {
			response.InternalServerError(c, "Failed to delete message")
		}
		return
	}

	logger.Info("Admin deleted message successfully",
		logger.Uint("message_id", uint(msgID)))

	response.OK(c, nil)
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
// @Success 201 {object} dto.TicketMessageResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/tickets/{id}/notes [post]
func (h *AdminTicketMessageHandler) AddInternalNote(c *gin.Context) {
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
			logger.ErrorField(err))
		response.NotFound(c, "Ticket not found")
		return
	}

	// Get admin user ID from context
	adminUserID, err := h.getAdminUserFromContext(c)
	if err != nil {
		logger.Error("Failed to get admin user from context for internal note creation",
			logger.Uint("ticket_id", uint(ticketID)),
			logger.ErrorField(err))
		response.Unauthorized(c, "Admin authentication required")
		return
	}

	// Use internal message service method if available, otherwise create regular message marked as internal
	message, err := h.ticketMessageService.CreateInternalMessage(c.Request.Context(), uint(ticketID), adminUserID, req.Content)
	if err != nil {
		logger.Error("Admin failed to create internal note",
			logger.Uint("ticket_id", uint(ticketID)),
			logger.ErrorField(err))
		response.InternalServerError(c, "Failed to create internal note")
		return
	}

	// Convert to response format and populate user data
	messageResponse := dto.ToTicketMessageResponse(message)
	if user, err := h.userService.GetUserByID(c.Request.Context(), message.UserID); err == nil {
		messageResponse.User = handlers.ConvertUserToBasicDTO(user)
	}

	logger.Info("Admin added internal note successfully",
		logger.Uint("ticket_id", uint(ticketID)),
		logger.Uint("message_id", message.ID))

	response.Created(c, messageResponse)
}
