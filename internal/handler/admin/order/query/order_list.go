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

// OrderListHandler handles order listing operations
type OrderListHandler struct {
	*shared.BaseHandler
}

// NewOrderListHandler creates a new order list handler
func NewOrderListHandler(
	subscriptionOrderService *service.SubscriptionOrderService,
	paymentService *service.PaymentService,
	userService *service.UserService,
) *OrderListHandler {
	return &OrderListHandler{
		BaseHandler: shared.NewBaseHandler(subscriptionOrderService, paymentService, userService),
	}
}

// ListOrders godoc
// @Summary [Admin] List all subscription orders
// @Description Get all subscription orders with advanced filtering and search (Admin only)
// @Tags Admin-Order-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param user_id query int false "Filter by user ID"
// @Param status query string false "Filter by status" Enums(pending,paid,failed,cancelled,refunded)
// @Param order_type query string false "Filter by order type" Enums(new,renewal,upgrade,downgrade)
// @Param payment_method query string false "Filter by payment method"
// @Param payment_gateway query string false "Filter by payment gateway"
// @Param min_amount query number false "Minimum amount filter"
// @Param max_amount query number false "Maximum amount filter"
// @Param start_date query string false "Start date filter (YYYY-MM-DD)"
// @Param end_date query string false "End date filter (YYYY-MM-DD)"
// @Param coupon_code query string false "Filter by coupon code"
// @Param search query string false "Search in order number, transaction ID, user email"
// @Param sort_by query string false "Sort by field" Enums(created_at,paid_at,amount,total_amount) default(created_at)
// @Param sort_order query string false "Sort order" Enums(asc,desc) default(desc)
// @Param limit query int false "Limit results" minimum(1) maximum(100) example(10)
// @Param offset query int false "Offset results" minimum(0) example(0)
// @Success 200 {object} response.PaginatedResponse{data=[]model.SubscriptionOrderResponse}
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/orders [get]
func (h *OrderListHandler) ListOrders(c *gin.Context) {
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

	// Bind query parameters
	var req shared.GetOrdersRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "Invalid query parameters", err.Error())
		return
	}

	// Set defaults
	if req.Limit <= 0 || req.Limit > 100 {
		req.Limit = 10
	}
	if req.Offset < 0 {
		req.Offset = 0
	}
	if req.SortBy == "" {
		req.SortBy = "created_at"
	}
	if req.SortOrder == "" {
		req.SortOrder = "desc"
	}

	// Validate parameters
	if req.Status != "" {
		if err := h.Validator.ValidateOrderStatus(req.Status); err != nil {
			response.BadRequest(c, "Invalid status parameter", err.Error())
			return
		}
	}

	if req.OrderType != "" {
		if err := h.Validator.ValidateOrderType(req.OrderType); err != nil {
			response.BadRequest(c, "Invalid order type parameter", err.Error())
			return
		}
	}

	if req.PaymentMethod != "" {
		if err := h.Validator.ValidatePaymentMethod(req.PaymentMethod); err != nil {
			response.BadRequest(c, "Invalid payment method parameter", err.Error())
			return
		}
	}

	if req.PaymentGateway != "" {
		if err := h.Validator.ValidatePaymentGateway(req.PaymentGateway); err != nil {
			response.BadRequest(c, "Invalid payment gateway parameter", err.Error())
			return
		}
	}

	if err := h.Validator.ValidateAmount(req.MinAmount); err != nil {
		response.BadRequest(c, "Invalid minimum amount", err.Error())
		return
	}

	if err := h.Validator.ValidateAmount(req.MaxAmount); err != nil {
		response.BadRequest(c, "Invalid maximum amount", err.Error())
		return
	}

	if err := h.Validator.ValidateDateRange(req.StartDate, req.EndDate); err != nil {
		response.BadRequest(c, "Invalid date range", err.Error())
		return
	}

	if err := h.Validator.ValidateSearchQuery(req.Search); err != nil {
		response.BadRequest(c, "Invalid search query", err.Error())
		return
	}

	// Get orders with filtering
	orders, totalCount, err := h.SubscriptionOrderService.GetOrdersWithFiltering(c.Request.Context(), &service.GetOrdersRequest{
		UserID:           req.UserID,
		Status:           req.Status,
		OrderType:        req.OrderType,
		PaymentMethod:    req.PaymentMethod,
		PaymentGateway:   req.PaymentGateway,
		MinAmount:        req.MinAmount,
		MaxAmount:        req.MaxAmount,
		StartDate:        req.StartDate,
		EndDate:          req.EndDate,
		CouponCode:       req.CouponCode,
		Search:           req.Search,
		SortBy:           req.SortBy,
		SortOrder:        req.SortOrder,
		Limit:            req.Limit,
		Offset:           req.Offset,
		IncludeRelations: true, // Admin needs full details
	})

	if err != nil {
		logger.Error("Failed to get orders", logger.Error2("error", err))
		response.InternalServerError(c, "Failed to get orders", err.Error())
		return
	}

	// Convert to response format
	var orderResponses []*model.SubscriptionOrderResponse
	for _, order := range orders {
		orderResponses = append(orderResponses, order.ToResponse())
	}

	response.OKPaginated(c, "Orders retrieved successfully", orderResponses, totalCount, req.Limit, req.Offset)
}