package query

import (
	"linke/internal/handler/admin/ticket/shared"
	"linke/internal/logger"
	"linke/internal/model"
	"linke/internal/response"
	"linke/internal/service"

	"github.com/gin-gonic/gin"
)

// TicketListHandler handles ticket listing operations
type TicketListHandler struct {
	*shared.BaseHandler
}

// NewTicketListHandler creates a new ticket list handler
func NewTicketListHandler(ticketService *service.TicketService, ticketMessageService *service.TicketMessageService) *TicketListHandler {
	return &TicketListHandler{
		BaseHandler: shared.NewBaseHandler(ticketService, ticketMessageService),
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
func (h *TicketListHandler) ListTickets(c *gin.Context) {
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

	// Validate optional parameters
	if req.Status != "" {
		if err := h.Validator.ValidateTicketStatus(req.Status); err != nil {
			response.BadRequest(c, "Invalid status parameter", err.Error())
			return
		}
	}

	if req.Priority != "" {
		if err := h.Validator.ValidateTicketPriority(req.Priority); err != nil {
			response.BadRequest(c, "Invalid priority parameter", err.Error())
			return
		}
	}

	if req.Category != "" {
		if err := h.Validator.ValidateTicketCategory(req.Category); err != nil {
			response.BadRequest(c, "Invalid category parameter", err.Error())
			return
		}
	}

	if req.Search != "" {
		if err := h.Validator.ValidateSearchQuery(req.Search); err != nil {
			response.BadRequest(c, "Invalid search query", err.Error())
			return
		}
	}

	// Get tickets
	tickets, total, err := h.TicketService.GetTickets(c.Request.Context(), &req)
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