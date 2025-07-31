package status

import (
	"linke/internal/handler/admin/order/shared"
	"linke/internal/logger"
	"linke/internal/middleware"
	"linke/internal/model"
	"linke/internal/response"
	"linke/internal/service"

	"github.com/gin-gonic/gin"
)

// OrderStatusHandler handles order status management operations
type OrderStatusHandler struct {
	*shared.BaseHandler
}

// NewOrderStatusHandler creates a new order status handler
func NewOrderStatusHandler(
	subscriptionOrderService *service.SubscriptionOrderService,
	paymentService *service.PaymentService,
	userService *service.UserService,
) *OrderStatusHandler {
	return &OrderStatusHandler{
		BaseHandler: shared.NewBaseHandler(subscriptionOrderService, paymentService, userService),
	}
}

// UpdateOrderStatus godoc
// @Summary [Admin] Update order status
// @Description Manually update order status with notes (Admin only)
// @Tags Admin-Order-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Order ID"
// @Param request body shared.UpdateOrderStatusRequest true "Status update data"
// @Success 200 {object} response.StandardResponse{data=model.SubscriptionOrderResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/orders/{id}/status [patch]
func (h *OrderStatusHandler) UpdateOrderStatus(c *gin.Context) {
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
	var req shared.UpdateOrderStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request data", err.Error())
		return
	}

	// Validate order status
	if err := h.Validator.ValidateOrderStatus(req.Status); err != nil {
		response.BadRequest(c, "Invalid status", err.Error())
		return
	}

	// Validate admin confirmation
	if err := h.Validator.ValidateAdminConfirmation(req.AdminConfirmed, req.Status); err != nil {
		response.BadRequest(c, "Admin confirmation required", err.Error())
		return
	}

	// Validate payment evidence when marking as paid
	if req.Status == "paid" {
		if req.PaymentEvidence == nil {
			response.BadRequest(c, "Payment evidence required", "Payment evidence is required when marking order as paid")
			return
		}
		
		if err := h.Validator.ValidatePaymentEvidence(req.PaymentEvidence); err != nil {
			response.BadRequest(c, "Invalid payment evidence", err.Error())
			return
		}
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
		
		order, updateErr = h.SubscriptionOrderService.UpdateOrderStatusWithEvidence(c.Request.Context(), orderID, updateReq, user.ID)
	} else {
		order, updateErr = h.SubscriptionOrderService.UpdateOrderStatus(c.Request.Context(), orderID, req.Status, user.ID, req.Notes, req.Reason)
	}

	if updateErr != nil {
		if updateErr.Error() == "subscription order not found" {
			response.NotFound(c, "Order not found")
			return
		}
		logger.Error("Failed to update order status", 
			logger.Error2("error", updateErr), 
			logger.Uint("order_id", orderID),
			logger.Uint("admin_id", user.ID),
			logger.String("new_status", req.Status))
		response.InternalServerError(c, "Failed to update order status", updateErr.Error())
		return
	}

	// Log admin action
	logger.Info("Admin updated order status",
		logger.Uint("admin_id", user.ID),
		logger.Uint("order_id", orderID),
		logger.String("new_status", req.Status),
		logger.String("notes", req.Notes))

	response.OK(c, "Order status updated successfully", order.ToResponse())
}