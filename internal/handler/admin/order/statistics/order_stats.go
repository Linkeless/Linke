package statistics

import (
	"linke/internal/handler/admin/order/shared"
	"linke/internal/logger"
	"linke/internal/middleware"
	"linke/internal/model"
	"linke/internal/response"
	"linke/internal/service"

	"github.com/gin-gonic/gin"
)

// OrderStatsHandler handles order statistics operations
type OrderStatsHandler struct {
	*shared.BaseHandler
}

// NewOrderStatsHandler creates a new order statistics handler
func NewOrderStatsHandler(
	subscriptionOrderService *service.SubscriptionOrderService,
	paymentService *service.PaymentService,
	userService *service.UserService,
) *OrderStatsHandler {
	return &OrderStatsHandler{
		BaseHandler: shared.NewBaseHandler(subscriptionOrderService, paymentService, userService),
	}
}

// GetOrderStats godoc
// @Summary [Admin] Get order statistics
// @Description Get comprehensive order statistics and analytics (Admin only)
// @Tags Admin-Order-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param period query string false "Statistics period" Enums(today,week,month,quarter,year,all) default(month)
// @Param start_date query string false "Custom start date (YYYY-MM-DD)"
// @Param end_date query string false "Custom end date (YYYY-MM-DD)"
// @Success 200 {object} response.StandardResponse{data=shared.GetOrderStatsResponse}
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/orders/stats [get]
func (h *OrderStatsHandler) GetOrderStats(c *gin.Context) {
	// Get current user from context
	userValue, exists := c.Get(middleware.AuthContextKey)
	if !exists {
		response.Unauthorized(c, "Authentication required")
		return
	}

	user, ok := userValue.(*model.User)
	if !ok {
		response.Unauthorized(c, "Invalid user context")
		return
	}

	// Check if user is admin
	if !user.IsAdmin() {
		response.Forbidden(c, "Admin access required")
		return
	}

	// Parse query parameters
	period := c.DefaultQuery("period", "month")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	// Validate statistics period
	if err := h.Validator.ValidateStatsPeriod(period); err != nil {
		response.BadRequest(c, "Invalid statistics period", err.Error())
		return
	}

	// Validate date range if provided
	if err := h.Validator.ValidateDateRange(startDate, endDate); err != nil {
		response.BadRequest(c, "Invalid date range", err.Error())
		return
	}

	// Get order statistics
	stats, err := h.SubscriptionOrderService.GetOrderStatistics(c.Request.Context(), &service.GetOrderStatsRequest{
		Period:    period,
		StartDate: startDate,
		EndDate:   endDate,
	})

	if err != nil {
		logger.Error("Failed to get order statistics", logger.Error2("error", err))
		response.InternalServerError(c, "Failed to get order statistics", err.Error())
		return
	}

	response.OK(c, "Order statistics retrieved successfully", stats)
}