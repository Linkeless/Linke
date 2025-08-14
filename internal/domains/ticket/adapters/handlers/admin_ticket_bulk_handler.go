package handlers

import (
	"linke/internal/domains/ticket/constants"
	"linke/internal/domains/ticket/dto"
	"linke/internal/shared/logger"
	"linke/internal/shared/response"

	"github.com/gin-gonic/gin"
)

// AdminTicketBulkHandler handles bulk ticket operations
type AdminTicketBulkHandler struct {
	*AdminTicketHandlerBase
}

// NewAdminTicketBulkHandler creates a new admin ticket bulk handler
func NewAdminTicketBulkHandler(base *AdminTicketHandlerBase) *AdminTicketBulkHandler {
	return &AdminTicketBulkHandler{
		AdminTicketHandlerBase: base,
	}
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
func (h *AdminTicketBulkHandler) BulkAssignTickets(c *gin.Context) {
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
			logger.ErrorField(err))
		response.InternalServerError(c, "Failed to bulk assign tickets")
		return
	}

	logger.Info("Admin bulk assigned tickets successfully",
		logger.Any("ticket_ids", req.TicketIDs),
		logger.Uint("assigned_to_id", *req.AssignedToID))

	response.OK(c, gin.H{
		"message":      "Tickets assigned successfully",
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
func (h *AdminTicketBulkHandler) BulkUpdateStatus(c *gin.Context) {
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
			logger.ErrorField(err))
		response.InternalServerError(c, "Failed to bulk update status")
		return
	}

	logger.Info("Admin bulk updated ticket status successfully",
		logger.Any("ticket_ids", req.TicketIDs),
		logger.String("status", *req.Status))

	response.OK(c, gin.H{
		"message":      "Ticket status updated successfully",
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
func (h *AdminTicketBulkHandler) BulkCloseTickets(c *gin.Context) {
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
			logger.ErrorField(err))
		response.InternalServerError(c, "Failed to bulk close tickets")
		return
	}

	logger.Info("Admin bulk closed tickets successfully",
		logger.Any("ticket_ids", req.TicketIDs))

	response.OK(c, gin.H{
		"message":      "Tickets closed successfully",
		"ticket_count": len(req.TicketIDs),
		"reason":       req.Reason,
	})
}
