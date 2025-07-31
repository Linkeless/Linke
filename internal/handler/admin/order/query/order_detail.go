package query

import (
	"linke/internal/handler/admin/order/shared"
	"linke/internal/logger"
	"linke/internal/middleware"
	"linke/internal/model"
	"linke/internal/response"
	"linke/internal/service"

	"github.com/gin-gonic/gin"
)

// OrderDetailHandler handles order detail operations
type OrderDetailHandler struct {
	*shared.BaseHandler
}

// NewOrderDetailHandler creates a new order detail handler
func NewOrderDetailHandler(
	subscriptionOrderService *service.SubscriptionOrderService,
	paymentService *service.PaymentService,
	userService *service.UserService,
) *OrderDetailHandler {
	return &OrderDetailHandler{
		BaseHandler: shared.NewBaseHandler(subscriptionOrderService, paymentService, userService),
	}
}

// GetOrder godoc
// @Summary [Admin] Get subscription order by ID
// @Description Get a subscription order by ID with full details (Admin only)
// @Tags Admin-Order-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Order ID"
// @Success 200 {object} response.StandardResponse{data=model.SubscriptionOrderResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/orders/{id} [get]
func (h *OrderDetailHandler) GetOrder(c *gin.Context) {
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

	// Validate order ID
	orderID, err := h.Validator.ValidateOrderID(c)
	if err != nil {
		return // Response already sent by validator
	}

	// Get order with relations
	order, err := h.SubscriptionOrderService.GetOrderWithRelations(c.Request.Context(), orderID)
	if err != nil {
		if err.Error() == "subscription order not found" {
			response.NotFound(c, "Order not found")
			return
		}
		logger.Error("Failed to get order", logger.Error2("error", err), logger.Uint("order_id", orderID))
		response.InternalServerError(c, "Failed to get order", err.Error())
		return
	}

	response.OK(c, "Order retrieved successfully", order.ToResponse())
}