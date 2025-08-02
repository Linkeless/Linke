package handlers

import (
	"strconv"

	subscriptionEntities "linke/internal/domains/subscription/entities"
	subscriptionInterfaces "linke/internal/domains/subscription/usecases/interfaces"
	userEntities "linke/internal/domains/user/entities"
	"linke/internal/shared/logger"
	"linke/internal/shared/middleware"
	"linke/internal/shared/response"

	"github.com/gin-gonic/gin"
)

type SubscriptionOrderHandler struct {
	subscriptionOrderService subscriptionInterfaces.SubscriptionOrderService
}

func NewSubscriptionOrderHandler(subscriptionOrderService subscriptionInterfaces.SubscriptionOrderService) *SubscriptionOrderHandler {
	return &SubscriptionOrderHandler{
		subscriptionOrderService: subscriptionOrderService,
	}
}

// CreateSubscriptionOrder godoc
// @Summary [User] Create subscription order
// @Description Create a new subscription order with payment
// @Tags User-Subscription
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param order body service.CreateSubscriptionOrderRequest true "Subscription order data"
// @Success 201 {object} response.StandardResponse{data=service.CreateSubscriptionOrderResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /subscription-orders [post]
func (h *SubscriptionOrderHandler) CreateSubscriptionOrder(c *gin.Context) {
	// Get current user from context
	userValue, exists := c.Get(middleware.AuthContextKey)
	if !exists {
		response.Unauthorized(c, "Authentication required")
		return
	}

	user, ok := userValue.(*userEntities.User)
	if !ok {
		response.Unauthorized(c, "Invalid user context")
		return
	}

	// Bind request
	var req subscriptionInterfaces.CreateSubscriptionOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request data", err.Error())
		return
	}

	// Only allow users to create orders for themselves (unless admin)
	if !user.IsAdmin() && req.UserID != user.ID {
		response.Forbidden(c, "You can only create orders for yourself")
		return
	}

	// Create subscription order
	orderResponse, err := h.subscriptionOrderService.CreateSubscriptionOrder(c.Request.Context(), &req)
	if err != nil {
		logger.Error("Failed to create subscription order", logger.Error2("error", err), logger.Uint("user_id", user.ID))
		response.InternalServerError(c, "Failed to create subscription order", err.Error())
		return
	}

	response.CreatedWithMessage(c, "Subscription order created successfully", orderResponse)
}

// GetSubscriptionOrder godoc
// @Summary [User] Get subscription order
// @Description Get subscription order details
// @Tags User-Subscription
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Subscription order ID"
// @Success 200 {object} response.StandardResponse{data=model.SubscriptionOrderResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /subscription-orders/{id} [get]
func (h *SubscriptionOrderHandler) GetSubscriptionOrder(c *gin.Context) {
	// Get current user from context
	userValue, exists := c.Get(middleware.AuthContextKey)
	if !exists {
		response.Unauthorized(c, "Authentication required")
		return
	}

	user, ok := userValue.(*userEntities.User)
	if !ok {
		response.Unauthorized(c, "Invalid user context")
		return
	}

	// Parse order ID
	orderIDStr := c.Param("id")
	orderID, err := strconv.ParseUint(orderIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid order ID", "Order ID must be a valid number")
		return
	}

	// Get subscription order
	order, err := h.subscriptionOrderService.GetSubscriptionOrder(c.Request.Context(), uint(orderID))
	if err != nil {
		if err.Error() == "subscription order not found" {
			response.NotFound(c, "Subscription order not found")
			return
		}
		logger.Error("Failed to get subscription order", logger.Error2("error", err), logger.Uint("order_id", uint(orderID)))
		response.InternalServerError(c, "Failed to get subscription order", err.Error())
		return
	}

	// Check if user has access to this order
	if !user.IsAdmin() && order.UserID != user.ID {
		response.Forbidden(c, "You can only access your own orders")
		return
	}

	response.OK(c, "Subscription order retrieved successfully", order.ToResponse())
}

// GetMySubscriptionOrders godoc
// @Summary [User] Get my subscription orders
// @Description Get current user's subscription orders
// @Tags User-Subscription
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Limit results" minimum(1) maximum(100) example(10)
// @Param offset query int false "Offset results" minimum(0) example(0)
// @Success 200 {object} response.PaginatedResponse{data=[]model.SubscriptionOrderResponse}
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /subscription-orders/my [get]
func (h *SubscriptionOrderHandler) GetMySubscriptionOrders(c *gin.Context) {
	// Get current user from context
	userValue, exists := c.Get(middleware.AuthContextKey)
	if !exists {
		response.Unauthorized(c, "Authentication required")
		return
	}

	user, ok := userValue.(*userEntities.User)
	if !ok {
		response.Unauthorized(c, "Invalid user context")
		return
	}

	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		limit = 10
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	// Get user subscription orders
	orders, totalCount, err := h.subscriptionOrderService.GetUserSubscriptionOrders(c.Request.Context(), user.ID, limit, offset)
	if err != nil {
		logger.Error("Failed to get user subscription orders", logger.Error2("error", err), logger.Uint("user_id", user.ID))
		response.InternalServerError(c, "Failed to get subscription orders", err.Error())
		return
	}

	// Convert to response format
	var orderResponses []*subscriptionEntities.SubscriptionOrderResponse
	for _, order := range orders {
		orderResponses = append(orderResponses, order.ToResponse())
	}

	response.OKPaginated(c, "My subscription orders retrieved successfully", orderResponses, totalCount, limit, offset)
}
