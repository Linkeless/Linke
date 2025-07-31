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

// OrderBulkHandler handles bulk order operations
type OrderBulkHandler struct {
	*shared.BaseHandler
}

// NewOrderBulkHandler creates a new order bulk handler
func NewOrderBulkHandler(
	subscriptionOrderService *service.SubscriptionOrderService,
	paymentService *service.PaymentService,
	userService *service.UserService,
) *OrderBulkHandler {
	return &OrderBulkHandler{
		BaseHandler: shared.NewBaseHandler(subscriptionOrderService, paymentService, userService),
	}
}

// BulkUpdate godoc
// @Summary [Admin] Bulk update orders
// @Description Perform bulk operations on multiple orders (Admin only)
// @Tags Admin-Order-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body shared.BulkUpdateRequest true "Bulk update data"
// @Success 200 {object} response.StandardResponse{data=map[string]interface{}}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/orders/bulk [post]
func (h *OrderBulkHandler) BulkUpdate(c *gin.Context) {
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
	var req shared.BulkUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request data", err.Error())
		return
	}

	// Validate bulk operation
	if err := h.Validator.ValidateBulkOperation(req.Operation); err != nil {
		response.BadRequest(c, "Invalid bulk operation", err.Error())
		return
	}

	// Validate order IDs
	if err := h.Validator.ValidateBulkOrderIDs(req.OrderIDs); err != nil {
		response.BadRequest(c, "Invalid order IDs", err.Error())
		return
	}

	// Validate admin confirmation for bulk operations
	if err := h.Validator.ValidateAdminConfirmation(req.AdminConfirmed, "bulk"); err != nil {
		response.BadRequest(c, "Admin confirmation required", err.Error())
		return
	}

	// Process bulk operation
	result, err := h.SubscriptionOrderService.BulkUpdateOrders(c.Request.Context(), &service.BulkUpdateOrdersRequest{
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