package statistics

import (
	"linke/internal/handler/admin/ticket/shared"
	"linke/internal/logger"
	"linke/internal/response"
	"linke/internal/service"

	"github.com/gin-gonic/gin"
)

// TicketStatsHandler handles ticket statistics operations
type TicketStatsHandler struct {
	*shared.BaseHandler
}

// NewTicketStatsHandler creates a new ticket statistics handler
func NewTicketStatsHandler(ticketService *service.TicketService, ticketMessageService *service.TicketMessageService) *TicketStatsHandler {
	return &TicketStatsHandler{
		BaseHandler: shared.NewBaseHandler(ticketService, ticketMessageService),
	}
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
func (h *TicketStatsHandler) GetTicketStats(c *gin.Context) {
	// Get ticket statistics
	stats, err := h.TicketService.GetTicketStats(c.Request.Context())
	if err != nil {
		logger.Error("Failed to get ticket stats", logger.Error2("error", err))
		response.InternalServerError(c, "Failed to get statistics", err.Error())
		return
	}

	response.OK(c, "Statistics retrieved successfully", stats)
}