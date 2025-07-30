package admin

import (
	"strconv"

	"linke/internal/logger"
	"linke/internal/middleware"
	"linke/internal/model"
	"linke/internal/response"
	"linke/internal/service"

	"github.com/gin-gonic/gin"
)

type AdminOrderHandler struct {
	subscriptionOrderService *service.SubscriptionOrderService
	paymentService           *service.PaymentService
	userService              *service.UserService
}

func NewAdminOrderHandler(
	subscriptionOrderService *service.SubscriptionOrderService,
	paymentService *service.PaymentService,
	userService *service.UserService,
) *AdminOrderHandler {
	return &AdminOrderHandler{
		subscriptionOrderService: subscriptionOrderService,
		paymentService:           paymentService,
		userService:              userService,
	}
}

// GetOrdersRequest represents the request to get orders with filtering
type GetOrdersRequest struct {
	UserID         *uint   `form:"user_id" example:"1"`
	Status         string  `form:"status" example:"paid"`
	OrderType      string  `form:"order_type" example:"new"`
	PaymentMethod  string  `form:"payment_method" example:"alipay"`
	PaymentGateway string  `form:"payment_gateway" example:"epay"`
	MinAmount      float64 `form:"min_amount" example:"0"`
	MaxAmount      float64 `form:"max_amount" example:"1000"`
	StartDate      string  `form:"start_date" example:"2024-01-01"`
	EndDate        string  `form:"end_date" example:"2024-12-31"`
	CouponCode     string  `form:"coupon_code" example:"SAVE20"`
	Search         string  `form:"search" example:"ORD2024001"`
	SortBy         string  `form:"sort_by" example:"created_at"`
	SortOrder      string  `form:"sort_order" example:"desc"`
	Limit          int     `form:"limit" example:"10"`
	Offset         int     `form:"offset" example:"0"`
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
func (h *AdminOrderHandler) ListOrders(c *gin.Context) {
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
	var req GetOrdersRequest
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

	// Get orders with filtering
	orders, totalCount, err := h.subscriptionOrderService.GetOrdersWithFiltering(c.Request.Context(), &service.GetOrdersRequest{
		UserID:         req.UserID,
		Status:         req.Status,
		OrderType:      req.OrderType,
		PaymentMethod:  req.PaymentMethod,
		PaymentGateway: req.PaymentGateway,
		MinAmount:      req.MinAmount,
		MaxAmount:      req.MaxAmount,
		StartDate:      req.StartDate,
		EndDate:        req.EndDate,
		CouponCode:     req.CouponCode,
		Search:         req.Search,
		SortBy:         req.SortBy,
		SortOrder:      req.SortOrder,
		Limit:          req.Limit,
		Offset:         req.Offset,
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
func (h *AdminOrderHandler) GetOrder(c *gin.Context) {
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

	// Parse order ID
	orderIDStr := c.Param("id")
	orderID, err := strconv.ParseUint(orderIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid order ID", "Order ID must be a valid number")
		return
	}

	// Get order with relations
	order, err := h.subscriptionOrderService.GetOrderWithRelations(c.Request.Context(), uint(orderID))
	if err != nil {
		if err.Error() == "subscription order not found" {
			response.NotFound(c, "Order not found")
			return
		}
		logger.Error("Failed to get order", logger.Error2("error", err), logger.Uint("order_id", uint(orderID)))
		response.InternalServerError(c, "Failed to get order", err.Error())
		return
	}

	response.OK(c, "Order retrieved successfully", order.ToResponse())
}

// UpdateOrderStatusRequest represents the request to update order status
type UpdateOrderStatusRequest struct {
	Status         string  `json:"status" binding:"required,oneof=pending paid failed cancelled refunded" example:"paid"`
	Notes          string  `json:"notes,omitempty" example:"Manual status update by admin"`
	Reason         string  `json:"reason,omitempty" example:"Payment verification completed"`
	NotifyUser     bool    `json:"notify_user,omitempty" example:"true"`
	// Payment evidence required when marking as paid
	PaymentEvidence *PaymentEvidenceRequest `json:"payment_evidence,omitempty"`
	// Admin confirmation for critical operations
	AdminConfirmed bool `json:"admin_confirmed,omitempty" example:"true"`
}

// PaymentEvidenceRequest represents payment verification data
type PaymentEvidenceRequest struct {
	TransactionID     string  `json:"transaction_id" binding:"required" example:"txn_123456789"`
	PaymentGateway    string  `json:"payment_gateway" binding:"required" example:"alipay"`
	PaymentMethod     string  `json:"payment_method" binding:"required" example:"alipay"`
	AmountReceived    float64 `json:"amount_received" binding:"required,min=0.01" example:"29.99"`
	Currency          string  `json:"currency" binding:"required" example:"USD"`
	PaymentTime       string  `json:"payment_time" binding:"required" example:"2024-01-01T10:30:00Z"`
	GatewayResponse   string  `json:"gateway_response,omitempty" example:"SUCCESS"`
	VerificationNotes string  `json:"verification_notes,omitempty" example:"Payment verified manually"`
}

// UpdateOrderStatus godoc
// @Summary [Admin] Update order status
// @Description Manually update order status with notes (Admin only)
// @Tags Admin-Order-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Order ID"
// @Param request body UpdateOrderStatusRequest true "Status update data"
// @Success 200 {object} response.StandardResponse{data=model.SubscriptionOrderResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/orders/{id}/status [patch]
func (h *AdminOrderHandler) UpdateOrderStatus(c *gin.Context) {
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

	// Parse order ID
	orderIDStr := c.Param("id")
	orderID, err := strconv.ParseUint(orderIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid order ID", "Order ID must be a valid number")
		return
	}

	// Bind request
	var req UpdateOrderStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request data", err.Error())
		return
	}

	// Validate payment evidence when marking as paid
	if req.Status == "paid" {
		if req.PaymentEvidence == nil {
			response.BadRequest(c, "Payment evidence required", "Payment evidence is required when marking order as paid")
			return
		}
		
		// Validate admin confirmation for critical operations
		if !req.AdminConfirmed {
			response.BadRequest(c, "Admin confirmation required", "Admin confirmation is required for marking orders as paid")
			return
		}
	}

	// Check if this is a critical operation requiring confirmation
	if (req.Status == "refunded" || req.Status == "cancelled") && !req.AdminConfirmed {
		response.BadRequest(c, "Admin confirmation required", "Admin confirmation is required for this operation")
		return
	}

	var order *model.SubscriptionOrder
	var updateErr error

	// Use enhanced method when marking as paid
	if req.Status == "paid" && req.PaymentEvidence != nil {
		// Convert request to service struct
		evidence := &service.PaymentEvidence{
			TransactionID:    req.PaymentEvidence.TransactionID,
			PaymentMethod:    req.PaymentEvidence.PaymentMethod,
			PaymentReference: req.PaymentEvidence.GatewayResponse,
			PaymentProof:     req.PaymentEvidence.VerificationNotes,
			VerifiedAmount:   req.PaymentEvidence.AmountReceived,
			Notes:            req.PaymentEvidence.VerificationNotes,
		}
		
		updateReq := &service.UpdateOrderStatusRequest{
			Status:          req.Status,
			Notes:           req.Notes,
			Reason:          req.Reason,
			PaymentEvidence: evidence,
			AdminConfirm:    req.AdminConfirmed,
		}
		
		order, updateErr = h.subscriptionOrderService.UpdateOrderStatusWithEvidence(c.Request.Context(), uint(orderID), updateReq, user.ID)
	} else {
		order, updateErr = h.subscriptionOrderService.UpdateOrderStatus(c.Request.Context(), uint(orderID), req.Status, user.ID, req.Notes, req.Reason)
	}

	if updateErr != nil {
		if updateErr.Error() == "subscription order not found" {
			response.NotFound(c, "Order not found")
			return
		}
		logger.Error("Failed to update order status", 
			logger.Error2("error", updateErr), 
			logger.Uint("order_id", uint(orderID)),
			logger.Uint("admin_id", user.ID),
			logger.String("new_status", req.Status))
		response.InternalServerError(c, "Failed to update order status", updateErr.Error())
		return
	}

	// Log admin action
	logger.Info("Admin updated order status",
		logger.Uint("admin_id", user.ID),
		logger.Uint("order_id", uint(orderID)),
		logger.String("new_status", req.Status),
		logger.String("notes", req.Notes))

	response.OK(c, "Order status updated successfully", order.ToResponse())
}

// ProcessRefundRequest represents the request to process a refund
type ProcessRefundRequest struct {
	Amount         float64 `json:"amount" binding:"required,min=0.01" example:"29.99"`
	Reason         string  `json:"reason" binding:"required" example:"Customer request"`
	RefundMethod   string  `json:"refund_method,omitempty" example:"original"`
	Notes          string  `json:"notes,omitempty" example:"Refund processed by admin"`
	NotifyUser     bool    `json:"notify_user,omitempty" example:"true"`
	AdminConfirmed bool    `json:"admin_confirmed" binding:"required" example:"true"`
}

// ProcessRefund godoc
// @Summary [Admin] Process order refund
// @Description Process a full or partial refund for an order (Admin only)
// @Tags Admin-Order-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Order ID"
// @Param request body ProcessRefundRequest true "Refund data"
// @Success 200 {object} response.StandardResponse{data=model.SubscriptionOrderResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/orders/{id}/refund [post]
func (h *AdminOrderHandler) ProcessRefund(c *gin.Context) {
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

	// Parse order ID
	orderIDStr := c.Param("id")
	orderID, err := strconv.ParseUint(orderIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid order ID", "Order ID must be a valid number")
		return
	}

	// Bind request
	var req ProcessRefundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request data", err.Error())
		return
	}

	// Validate admin confirmation for refund processing
	if !req.AdminConfirmed {
		response.BadRequest(c, "Admin confirmation required", "Admin confirmation is required for processing refunds")
		return
	}

	// Process refund
	order, err := h.subscriptionOrderService.ProcessRefund(c.Request.Context(), &service.ProcessRefundRequest{
		OrderID:      uint(orderID),
		Amount:       req.Amount,
		Reason:       req.Reason,
		RefundMethod: req.RefundMethod,
		Notes:        req.Notes,
		ProcessedBy:  user.ID,
	})

	if err != nil {
		if err.Error() == "subscription order not found" {
			response.NotFound(c, "Order not found")
			return
		}
		logger.Error("Failed to process refund", 
			logger.Error2("error", err), 
			logger.Uint("order_id", uint(orderID)),
			logger.Uint("admin_id", user.ID),
			logger.Any("amount", req.Amount))
		response.InternalServerError(c, "Failed to process refund", err.Error())
		return
	}

	// Log admin action
	logger.Info("Admin processed refund",
		logger.Uint("admin_id", user.ID),
		logger.Uint("order_id", uint(orderID)),
		logger.Any("amount", req.Amount),
		logger.String("reason", req.Reason))

	response.OK(c, "Refund processed successfully", order.ToResponse())
}

// GetOrderStatsResponse represents order statistics
type GetOrderStatsResponse struct {
	TotalOrders     int64   `json:"total_orders"`
	PendingOrders   int64   `json:"pending_orders"`
	PaidOrders      int64   `json:"paid_orders"`
	FailedOrders    int64   `json:"failed_orders"`
	CancelledOrders int64   `json:"cancelled_orders"`
	RefundedOrders  int64   `json:"refunded_orders"`
	TotalRevenue    float64 `json:"total_revenue"`
	TotalRefunded   float64 `json:"total_refunded"`
	AvgOrderValue   float64 `json:"avg_order_value"`
	ConversionRate  float64 `json:"conversion_rate"`
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
// @Success 200 {object} response.StandardResponse{data=GetOrderStatsResponse}
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/orders/stats [get]
func (h *AdminOrderHandler) GetOrderStats(c *gin.Context) {
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

	// Get order statistics
	stats, err := h.subscriptionOrderService.GetOrderStatistics(c.Request.Context(), &service.GetOrderStatsRequest{
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

// BulkUpdateRequest represents bulk operation request
type BulkUpdateRequest struct {
	OrderIDs       []uint `json:"order_ids" binding:"required,min=1" example:"1,2,3"`
	Operation      string `json:"operation" binding:"required,oneof=cancel refund export" example:"cancel"`
	Reason         string `json:"reason,omitempty" example:"Bulk cancellation"`
	Notes          string `json:"notes,omitempty" example:"Bulk operation by admin"`
	AdminConfirmed bool   `json:"admin_confirmed" binding:"required" example:"true"`
}

// BulkUpdate godoc
// @Summary [Admin] Bulk update orders
// @Description Perform bulk operations on multiple orders (Admin only)
// @Tags Admin-Order-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body BulkUpdateRequest true "Bulk update data"
// @Success 200 {object} response.StandardResponse{data=map[string]interface{}}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/orders/bulk [post]
func (h *AdminOrderHandler) BulkUpdate(c *gin.Context) {
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

	// Bind request
	var req BulkUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request data", err.Error())
		return
	}

	// Validate admin confirmation for bulk operations
	if !req.AdminConfirmed {
		response.BadRequest(c, "Admin confirmation required", "Admin confirmation is required for bulk operations")
		return
	}

	// Validate order IDs limit
	if len(req.OrderIDs) > 100 {
		response.BadRequest(c, "Too many orders", "Maximum 100 orders can be processed at once")
		return
	}

	// Process bulk operation
	result, err := h.subscriptionOrderService.BulkUpdateOrders(c.Request.Context(), &service.BulkUpdateOrdersRequest{
		OrderIDs:    req.OrderIDs,
		Operation:   req.Operation,
		Reason:      req.Reason,
		Notes:       req.Notes,
		ProcessedBy: user.ID,
	})

	if err != nil {
		logger.Error("Failed to process bulk update", 
			logger.Error2("error", err), 
			logger.Uint("admin_id", user.ID),
			logger.String("operation", req.Operation),
			logger.Any("order_count", len(req.OrderIDs)))
		response.InternalServerError(c, "Failed to process bulk update", err.Error())
		return
	}

	// Log admin action
	logger.Info("Admin performed bulk operation",
		logger.Uint("admin_id", user.ID),
		logger.String("operation", req.Operation),
		logger.Any("order_count", len(req.OrderIDs)),
		logger.Any("successful", result.Successful),
		logger.Any("failed", result.Failed))

	response.OK(c, "Bulk operation completed", result)
}