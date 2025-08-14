package handlers

import (
	"time"

	"linke/internal/domains/ticket/dto"
	"linke/internal/shared/handlers"
	"linke/internal/shared/logger"
	"linke/internal/shared/response"

	"github.com/gin-gonic/gin"
)

// AdminTicketSearchHandler handles ticket search and analytics operations
type AdminTicketSearchHandler struct {
	*AdminTicketHandlerBase
}

// NewAdminTicketSearchHandler creates a new admin ticket search handler
func NewAdminTicketSearchHandler(base *AdminTicketHandlerBase) *AdminTicketSearchHandler {
	return &AdminTicketSearchHandler{
		AdminTicketHandlerBase: base,
	}
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
// @Success 200 {object} response.HALCollectionResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/tickets/search [get]
func (h *AdminTicketSearchHandler) SearchTickets(c *gin.Context) {
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
			logger.ErrorField(err))
		response.InternalServerError(c, "Failed to search tickets")
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
		logger.Error("Failed to batch load users for search results",
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
	}

	page := (req.Offset / req.Limit) + 1
	response.PaginatedWithQuery(c, "Search completed", responses, page, req.Limit, total, "/api/v1/admin/tickets/search", map[string]any{
		"query": req.Query,
		"filters": map[string]any{
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
func (h *AdminTicketSearchHandler) GetStatistics(c *gin.Context) {
	stats, err := h.ticketService.GetTicketStatistics(c.Request.Context(), "", "")
	if err != nil {
		logger.Error("Admin failed to get ticket statistics", logger.ErrorField(err))
		response.InternalServerError(c, "Failed to get ticket statistics")
		return
	}

	response.OK(c, stats)
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
func (h *AdminTicketSearchHandler) GetAnalytics(c *gin.Context) {
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
		logger.Error("Admin failed to get ticket analytics", logger.ErrorField(err))
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

	response.OK(c, analytics)
}
