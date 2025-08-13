package handlers

import (
	"strconv"

	"linke/internal/domains/subscription/dto"
	"linke/internal/domains/subscription/entities"
	"linke/internal/domains/subscription/usecases/interfaces"
	userEntities "linke/internal/domains/user/entities"
	"linke/internal/shared/logger"
	"linke/internal/shared/middleware"
	"linke/internal/shared/response"

	"github.com/gin-gonic/gin"
)

type SubscriptionOrderHandler struct {
	subscriptionOrderService interfaces.SubscriptionOrderService
}

func NewSubscriptionOrderHandler(subscriptionOrderService interfaces.SubscriptionOrderService) *SubscriptionOrderHandler {
	return &SubscriptionOrderHandler{
		subscriptionOrderService: subscriptionOrderService,
	}
}

// GetSubscriptionOrderSummary godoc
// @Summary [User] Get order summary (order + latest payment + latest invoice)
// @Description Aggregate order with latest payment and invoice
// @Tags User-Subscription
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Subscription order ID"
// @Success 200 {object} map[string]any
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /orders/{id}/summary [get]
func (h *SubscriptionOrderHandler) GetSubscriptionOrderSummary(c *gin.Context) {
	// auth
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

	// parse id
	orderIDStr := c.Param("id")
	orderID64, err := strconv.ParseUint(orderIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid order ID")
		return
	}
	orderID := uint(orderID64)

	// enforce ownership in service aggregation result by reusing GetSubscriptionOrder first
	order, err := h.subscriptionOrderService.GetSubscriptionOrder(c.Request.Context(), orderID)
	if err != nil {
		response.NotFound(c, "Subscription order not found")
		return
	}
	if !user.IsAdmin() && order.UserID != user.ID {
		response.Forbidden(c, "You can only access your own orders")
		return
	}

	summary, err := h.subscriptionOrderService.GetSubscriptionOrderSummary(c.Request.Context(), orderID)
	if err != nil {
		logger.Error("Failed to get order summary", logger.Uint("orderID", uint(orderID)))
		response.InternalServerError(c, "Failed to get order summary")
		return
	}
	response.OK(c, summary)
}

// ==============================================================================
// NEW STANDARDIZED ORDER CREATION FLOW
// ==============================================================================

// CreateOrderWithInvoice godoc
// @Summary [User] Create order with complete invoice and payment flow
// @Description Create a new subscription order with invoice and payment record
// @Tags User-Subscription
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateOrderRequest true "Order creation request"
// @Success 201 {object} object
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /orders/create [post]
func (h *SubscriptionOrderHandler) CreateOrderWithInvoice(c *gin.Context) {
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
	var req dto.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request data")
		return
	}

	// Only allow users to create orders for themselves (unless admin)
	if !user.IsAdmin() && req.UserID != user.ID {
		response.Forbidden(c, "You can only create orders for yourself")
		return
	}

	// Create order with complete flow
	orderResponse, err := h.subscriptionOrderService.CreateOrderWithInvoice(c.Request.Context(), &req)
	if err != nil {
		logger.Error("Failed to create order with invoice", logger.ErrorField(err), logger.Uint("user_id", user.ID))
		response.InternalServerError(c, "Failed to create order")
		return
	}

	response.Created(c, orderResponse)
}

// GenerateInvoiceForOrder godoc
// @Summary [User] Generate invoice for existing order
// @Description Generate an invoice for an existing subscription order
// @Tags User-Subscription
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Subscription order ID"
// @Success 200 {object} any
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /orders/{id}/invoice [post]
func (h *SubscriptionOrderHandler) GenerateInvoiceForOrder(c *gin.Context) {
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
		response.BadRequest(c, "Order ID must be a valid number")
		return
	}

	// Check order ownership
	order, err := h.subscriptionOrderService.GetSubscriptionOrder(c.Request.Context(), uint(orderID))
	if err != nil {
		response.NotFound(c, "Subscription order not found")
		return
	}

	if !user.IsAdmin() && order.UserID != user.ID {
		response.Forbidden(c, "You can only generate invoices for your own orders")
		return
	}

	// Generate invoice
	invoice, err := h.subscriptionOrderService.GenerateInvoiceForOrder(c.Request.Context(), uint(orderID))
	if err != nil {
		logger.Error("Failed to generate invoice for order", logger.ErrorField(err), logger.Uint("order_id", uint(orderID)))
		response.InternalServerError(c, "Failed to generate invoice")
		return
	}

	response.OK(c, invoice)
}

// CreatePaymentForOrder godoc
// @Summary [User] Create payment for existing order
// @Description Create a payment record for an existing subscription order
// @Tags User-Subscription
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.PayOrderRequest true "Payment creation request"
// @Success 201 {object} object
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /orders/pay [post]
func (h *SubscriptionOrderHandler) CreatePaymentForOrder(c *gin.Context) {
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
	var req dto.PayOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request data")
		return
	}

	// Check order ownership
	order, err := h.subscriptionOrderService.GetSubscriptionOrder(c.Request.Context(), req.OrderID)
	if err != nil {
		response.NotFound(c, "Subscription order not found")
		return
	}

	if !user.IsAdmin() && order.UserID != user.ID {
		response.Forbidden(c, "You can only create payments for your own orders")
		return
	}

	// Create payment
	paymentResponse, err := h.subscriptionOrderService.CreatePaymentForOrder(c.Request.Context(), &req)
	if err != nil {
		logger.Error("Failed to create payment for order", logger.ErrorField(err), logger.Uint("order_id", req.OrderID))
		response.InternalServerError(c, "Failed to create payment")
		return
	}

	response.Created(c, paymentResponse)
}

// ==============================================================================
// LEGACY ENDPOINTS (for backward compatibility)
// ==============================================================================

// Deprecated (user): Use POST /api/v1/orders/create instead. Handler remains for internal/admin flows if needed.
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
	var req dto.CreateSubscriptionOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request data")
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
		logger.Error("Failed to create subscription order", logger.ErrorField(err), logger.Uint("user_id", user.ID))
		response.InternalServerError(c, "Failed to create subscription order")
		return
	}

	response.Created(c, orderResponse)
}

// GetSubscriptionOrder godoc
// @Summary [User] Get order
// @Description Get order details
// @Tags User-Subscription
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Subscription order ID"
// @Success 200 {object} entities.SubscriptionOrderResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /orders/{id} [get]
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
		response.BadRequest(c, "Order ID must be a valid number")
		return
	}

	// Get subscription order
	order, err := h.subscriptionOrderService.GetSubscriptionOrder(c.Request.Context(), uint(orderID))
	if err != nil {
		if err.Error() == "subscription order not found" {
			response.NotFound(c, "Subscription order not found")
			return
		}
		logger.Error("Failed to get subscription order", logger.ErrorField(err), logger.Uint("order_id", uint(orderID)))
		response.InternalServerError(c, "Failed to get subscription order")
		return
	}

	// Check if user has access to this order
	if !user.IsAdmin() && order.UserID != user.ID {
		response.Forbidden(c, "You can only access your own orders")
		return
	}

	response.OK(c, order.ToResponse())
}

// GetMySubscriptionOrders godoc
// @Summary [User] Get my orders
// @Description Get current user's orders
// @Tags User-Subscription
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Limit results" minimum(1) maximum(100) example(10)
// @Param offset query int false "Offset results" minimum(0) example(0)
// @Success 200 {object} response.HALCollectionResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /orders [get]
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
		logger.Error("Failed to get user subscription orders", logger.ErrorField(err), logger.Uint("user_id", user.ID))
		response.InternalServerError(c, "Failed to get subscription orders")
		return
	}

	// Convert to response format
	var orderResponses []*entities.SubscriptionOrderResponse
	for _, order := range orders {
		orderResponses = append(orderResponses, order.ToResponse())
	}

	// Convert offset to page number
	_ = (offset / limit) + 1 // page calculation
	response.SendPaginatedResponse(c, orderResponses, totalCount)
}
