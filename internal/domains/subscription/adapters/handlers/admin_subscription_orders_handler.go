package handlers

import (
	"strconv"
	"strings"
	"time"

	"linke/internal/domains/subscription/dto"
	"linke/internal/domains/subscription/entities"
	"linke/internal/shared/logger"
	"linke/internal/shared/response"

	"github.com/gin-gonic/gin"
)

// AdminSubscriptionOrdersHandler handles subscription order management operations
type AdminSubscriptionOrdersHandler struct {
	*AdminSubscriptionHandlerBase
}

// NewAdminSubscriptionOrdersHandler creates a new admin subscription orders handler
func NewAdminSubscriptionOrdersHandler(base *AdminSubscriptionHandlerBase) *AdminSubscriptionOrdersHandler {
	return &AdminSubscriptionOrdersHandler{
		AdminSubscriptionHandlerBase: base,
	}
}

// GetSubscriptionOrder godoc
// @Summary Get subscription order
// @Description Get subscription order details by ID (Admin only)
// @Tags Admin-Subscription-Orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Order ID"
// @Success 200 {object} entities.SubscriptionOrderResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Router /admin/subscriptions/orders/{id} [get]
func (h *AdminSubscriptionOrdersHandler) GetSubscriptionOrder(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid order ID")
		return
	}

	order, err := h.subscriptionOrderService.GetSubscriptionOrder(c.Request.Context(), uint(id))
	if err != nil {
		logger.Error("Admin failed to get subscription order",
			logger.Uint("order_id", uint(id)),
			logger.ErrorField(err),
		)
		response.NotFound(c, "Subscription order not found")
		return
	}

	response.OK(c, order.ToResponse())
}

// ListSubscriptionOrders godoc
// @Summary List subscription orders
// @Description Get paginated list of all subscription orders (Admin only)
// @Tags Admin-Subscription-Orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param user_id query int false "Filter by user ID"
// @Param status query string false "Filter by status"
// @Param order_type query string false "Filter by order type"
// @Param date_from query string false "Filter from date (YYYY-MM-DD)"
// @Param date_to query string false "Filter to date (YYYY-MM-DD)"
// @Param limit query int false "Items per page" default(10)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} response.HALCollectionResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/subscriptions/orders [get]
func (h *AdminSubscriptionOrdersHandler) ListSubscriptionOrders(c *gin.Context) {
	// Parse query parameters
	req := &dto.GetSubscriptionOrdersRequest{}

	if err := c.ShouldBindQuery(req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Set defaults
	if req.Limit == 0 {
		req.Limit = 10
	}
	if req.Limit > 100 {
		req.Limit = 100
	}

	orders, total, err := h.subscriptionOrderService.GetSubscriptionOrders(c.Request.Context(), req)
	if err != nil {
		logger.Error("Admin failed to list subscription orders", logger.ErrorField(err))
		response.InternalServerError(c, "Failed to list subscription orders")
		return
	}

	// Convert to response format
	orderResponses := make([]*entities.SubscriptionOrderResponse, len(orders))
	for i, order := range orders {
		orderResponses[i] = order.ToResponse()
	}

	_ = (req.Offset / req.Limit) + 1 // page calculation for future use
	response.SendPaginatedResponse(c, orderResponses, total)
}

// CancelSubscriptionOrder godoc
// @Summary Cancel subscription order
// @Description Cancel a pending subscription order (Admin only)
// @Tags Admin-Subscription-Orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Order ID"
// @Param cancel body object{reason=string} true "Cancel data"
// @Success 204 "No Content"
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/subscriptions/orders/{id}/cancel [post]
func (h *AdminSubscriptionOrdersHandler) CancelSubscriptionOrder(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid order ID")
		return
	}

	var cancelData struct {
		Reason string `json:"reason" binding:"required,max=255"`
	}
	if err := c.ShouldBindJSON(&cancelData); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.subscriptionOrderService.CancelSubscriptionOrder(c.Request.Context(), uint(id), cancelData.Reason); err != nil {
		logger.Error("Admin failed to cancel subscription order",
			logger.Uint("order_id", uint(id)),
			logger.ErrorField(err),
		)

		if strings.Contains(err.Error(), "not found") {
			response.NotFound(c, "Subscription order not found")
		} else {
			response.InternalServerError(c, "Failed to cancel order")
		}
		return
	}

	logger.Info("Admin cancelled subscription order",
		logger.Uint("order_id", uint(id)),
		logger.String("reason", cancelData.Reason),
		logger.String("admin_action", "cancel_order"),
	)

	response.OK(c, gin.H{"message": "Order cancelled successfully"})
}

// GetOrderStatistics godoc
// @Summary Get order statistics
// @Description Get subscription order statistics (Admin only)
// @Tags Admin-Subscription-Analytics
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param from_date query string false "Start date (YYYY-MM-DD)"
// @Param to_date query string false "End date (YYYY-MM-DD)"
// @Success 200 {object} response.StandardResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/subscriptions/analytics/orders [get]
func (h *AdminSubscriptionOrdersHandler) GetOrderStatistics(c *gin.Context) {
	fromDateStr := c.Query("from_date")
	toDateStr := c.Query("to_date")

	// Default to last 30 days if not specified
	var fromDate, toDate time.Time
	var err error

	if fromDateStr == "" {
		fromDate = time.Now().AddDate(0, 0, -30)
	} else {
		fromDate, err = time.Parse("2006-01-02", fromDateStr)
		if err != nil {
			response.BadRequest(c, "Invalid from_date format, use YYYY-MM-DD")
			return
		}
	}

	if toDateStr == "" {
		toDate = time.Now()
	} else {
		toDate, err = time.Parse("2006-01-02", toDateStr)
		if err != nil {
			response.BadRequest(c, "Invalid to_date format, use YYYY-MM-DD")
			return
		}
	}

	stats, err := h.subscriptionOrderService.GetOrderStatistics(c.Request.Context(), fromDate, toDate)
	if err != nil {
		logger.Error("Admin failed to get order statistics", logger.ErrorField(err))
		response.InternalServerError(c, "Failed to get order statistics")
		return
	}

	response.OK(c, stats)
}