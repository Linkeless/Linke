package operation

import (
	"linke/internal/handler/admin/order/shared"
	"linke/internal/logger"
	"linke/internal/middleware"
	"linke/internal/model"
	"linke/internal/response"
	"linke/internal/service"

	"github.com/gin-gonic/gin"
)

// OrderRefundHandler handles order refund operations
type OrderRefundHandler struct {
	*shared.BaseHandler
}

// NewOrderRefundHandler creates a new order refund handler
func NewOrderRefundHandler(
	subscriptionOrderService *service.SubscriptionOrderService,
	paymentService *service.PaymentService,
	userService *service.UserService,
) *OrderRefundHandler {
	return &OrderRefundHandler{
		BaseHandler: shared.NewBaseHandler(subscriptionOrderService, paymentService, userService),
	}
}

// ProcessRefund godoc
// @Summary [Admin] Process order refund
// @Description Process a full or partial refund for an order (Admin only)
// @Tags Admin-Order-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Order ID"
// @Param request body shared.ProcessRefundRequest true "Refund data"
// @Success 200 {object} response.StandardResponse{data=model.SubscriptionOrderResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/orders/{id}/refund [post]
func (h *OrderRefundHandler) ProcessRefund(c *gin.Context) {
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

	// Bind request
	var req shared.ProcessRefundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request data", err.Error())
		return
	}

	// Validate refund amount
	if err := h.Validator.ValidateRefundAmount(req.Amount); err != nil {
		response.BadRequest(c, "Invalid refund amount", err.Error())
		return
	}

	// Validate refund reason
	if err := h.Validator.ValidateRefundReason(req.Reason); err != nil {
		response.BadRequest(c, "Invalid refund reason", err.Error())
		return
	}

	// Validate admin confirmation for refund processing
	if err := h.Validator.ValidateAdminConfirmation(req.AdminConfirmed, "refund"); err != nil {
		response.BadRequest(c, "Admin confirmation required", err.Error())
		return
	}

	// Process refund
	order, err := h.SubscriptionOrderService.ProcessRefund(c.Request.Context(), &service.ProcessRefundRequest{
		OrderID:      orderID,
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
			logger.Uint("order_id", orderID),
			logger.Uint("admin_id", user.ID),
			logger.Any("amount", req.Amount))
		response.InternalServerError(c, "Failed to process refund", err.Error())
		return
	}

	// Log admin action
	logger.Info("Admin processed refund",
		logger.Uint("admin_id", user.ID),
		logger.Uint("order_id", orderID),
		logger.Any("amount", req.Amount),
		logger.String("reason", req.Reason))

	response.OK(c, "Refund processed successfully", order.ToResponse())
}