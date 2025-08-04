package handlers

import (
	"net"

	paymententities "linke/internal/domains/payment/entities"
	"linke/internal/domains/subscription/usecases/interfaces"
	userEntities "linke/internal/domains/user/entities"
	"linke/internal/shared/logger"
	"linke/internal/shared/middleware"
	"linke/internal/shared/response"

	"github.com/gin-gonic/gin"
)

type QuickPurchaseHandler struct {
	subscriptionOrderService interfaces.SubscriptionOrderService
}

func NewQuickPurchaseHandler(subscriptionOrderService interfaces.SubscriptionOrderService) *QuickPurchaseHandler {
	return &QuickPurchaseHandler{
		subscriptionOrderService: subscriptionOrderService,
	}
}

// QuickPurchase godoc
// @Summary [User] Quick purchase subscription
// @Description Create a payment directly for subscription without creating order/invoice first. Order and invoice are created asynchronously after payment success.
// @Tags User-Subscription
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param purchase body interfaces.QuickPurchaseRequest true "Quick purchase data"
// @Success 201 {object} response.StandardResponse{data=interfaces.QuickPurchaseResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /subscription/quick-purchase [post]
func (h *QuickPurchaseHandler) QuickPurchase(c *gin.Context) {
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
	var req interfaces.QuickPurchaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request data", err.Error())
		return
	}

	// Only allow users to create purchases for themselves (unless admin)
	if !user.IsAdmin() && req.UserID != user.ID {
		response.Forbidden(c, "You can only create purchases for yourself")
		return
	}

	// Set user ID if not provided or if user is not admin
	if !user.IsAdmin() {
		req.UserID = user.ID
	}

	// Get client IP for payment processing
	if req.ClientIP == "" {
		req.ClientIP = getClientIP(c)
	}

	// Validate payment method and gateway
	if req.PaymentGateway == "" {
		response.BadRequest(c, "Payment gateway is required")
		return
	}

	if req.PaymentMethod == "" {
		response.BadRequest(c, "Payment method is required")
		return
	}

	// Create quick purchase
	purchaseResponse, err := h.subscriptionOrderService.QuickPurchase(c.Request.Context(), &req)
	if err != nil {
		logger.Error("Failed to create quick purchase",
			logger.Error2("error", err),
			logger.Uint("user_id", user.ID),
			logger.Uint("plan_id", req.PlanID))
		response.InternalServerError(c, "Failed to create quick purchase", err.Error())
		return
	}

	// Type assert PaymentRecord to get the payment number for logging
	var paymentNo string
	if paymentRecord, ok := purchaseResponse.PaymentRecord.(*paymententities.PaymentRecordResponse); ok {
		paymentNo = paymentRecord.PaymentNo
	}
	
	logger.Info("Quick purchase created successfully",
		logger.Uint("user_id", user.ID),
		logger.Uint("plan_id", req.PlanID),
		logger.String("payment_no", paymentNo))

	response.CreatedWithMessage(c, "Quick purchase created successfully", purchaseResponse)
}

// getClientIP extracts the real client IP from the request
func getClientIP(c *gin.Context) string {
	// Check if behind proxy (X-Forwarded-For header)
	if xff := c.GetHeader("X-Forwarded-For"); xff != "" {
		// X-Forwarded-For may contain multiple IPs, get the first one
		if ip := parseFirstIP(xff); ip != "" {
			return ip
		}
	}

	// Check X-Real-IP header
	if xri := c.GetHeader("X-Real-IP"); xri != "" {
		if net.ParseIP(xri) != nil {
			return xri
		}
	}

	// Fallback to RemoteAddr
	ip, _, _ := net.SplitHostPort(c.Request.RemoteAddr)
	return ip
}

// parseFirstIP parses the first valid IP from a comma-separated list
func parseFirstIP(ips string) string {
	for _, ip := range []string{ips} {
		if trimmed := net.ParseIP(ip); trimmed != nil {
			return ip
		}
	}
	return ""
}
