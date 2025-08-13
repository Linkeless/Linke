package handlers

import (
	"context"
	"fmt"

	"linke/internal/domains/payment/constants"
	"linke/internal/domains/payment/dto"
	"linke/internal/domains/payment/usecases/interfaces"
	userEntities "linke/internal/domains/user/entities"
	"linke/internal/shared/logger"
	"linke/internal/shared/middleware"
	"linke/internal/shared/response"

	"github.com/gin-gonic/gin"
)

// PaymentOrderHandler handles HTTP requests for payment order management
type PaymentOrderHandler struct {
	paymentService interfaces.PaymentService
}

// NewPaymentOrderHandler creates a new payment order handler
func NewPaymentOrderHandler(paymentService interfaces.PaymentService) *PaymentOrderHandler {
	return &PaymentOrderHandler{
		paymentService: paymentService,
	}
}

// CreatePaymentOrder godoc
// @Summary [User] Create payment order
// @Description Create a new payment order for invoice or subscription. Supports epay gateway with Alipay, WeChat Pay, and QQ Pay. Includes anti-replay protection and secure payment URL generation. Either invoice_id or subscription_order_id must be provided.
// @Tags User-Payment
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param payment_order body dto.CreatePaymentOrderRequest true "Payment order data with gateway (epay), method (alipay/wechat/qqpay), amount, and either invoice_id or subscription_order_id"
// @Success 201 {object} dto.PaymentRecordResponse "Returns payment record with payment URL, QR code, and expiration time"
// @Failure 400 {object} response.BadRequestResponse "Invalid request data, unsupported payment method, or missing required fields"
// @Failure 401 {object} response.UnauthorizedResponse "Authentication required"
// @Failure 422 {object} response.BadRequestResponse "Amount outside valid range or gateway configuration error"
// @Failure 500 {object} response.InternalServerErrorResponse "Payment gateway error or database failure"
// @Router /payment/orders [post]
func (h *PaymentOrderHandler) CreatePaymentOrder(c *gin.Context) {
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
	var req dto.CreatePaymentOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request")
		return
	}

	// Validate that either invoice_id or subscription_order_id is provided
	if req.InvoiceID == nil && req.SubscriptionOrderID == nil {
		response.BadRequest(c, "Either invoice_id or subscription_order_id is required")
		return
	}

	// Get client IP
	clientIP := c.ClientIP()

	// Build notify URL
	notifyURL := "https://" + c.Request.Host + "/api/v1/payment/notify/" + req.Gateway

	// Create payment service request
	paymentReq := &dto.CreatePaymentOrderRequest{
		UserID:              user.ID,
		SubscriptionOrderID: req.SubscriptionOrderID,
		InvoiceID:           req.InvoiceID,
		Gateway:             req.Gateway,
		PaymentMethod:       req.PaymentMethod,
		Amount:              req.Amount,
		Currency:            req.Currency,
		Subject:             req.Subject,
		Body:                req.Body,
		ClientIP:            clientIP,
		NotifyURL:           notifyURL,
		ReturnURL:           req.ReturnURL,
		ExpiredMinutes:      req.ExpiredMinutes,
	}

	// Create payment order
	paymentRecord, err := h.paymentService.CreatePaymentOrder(c.Request.Context(), paymentReq)
	if err != nil {
		logger.Error("Failed to create payment order", logger.ErrorField(err), logger.Uint("user_id", user.ID))
		response.InternalServerError(c, "Failed to create payment order")
		return
	}

	response.Created(c, dto.ToPaymentRecordUserResponse(paymentRecord))
}

// GetPaymentOrder godoc
// @Summary [User] Get payment order
// @Description Get payment order details
// @Tags User-Payment
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param payment_no path string true "Payment number"
// @Success 200 {object} dto.PaymentRecordResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /payment/orders/{payment_no} [get]
func (h *PaymentOrderHandler) GetPaymentOrder(c *gin.Context) {
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

	// Get payment number from path
	paymentNo := c.Param("payment_no")
	if paymentNo == "" {
		response.BadRequest(c, "Payment number is required")
		return
	}

	// Get payment record
	paymentRecord, err := h.paymentService.GetPaymentRecord(c.Request.Context(), paymentNo)
	if err != nil {
		if err.Error() == "payment record not found" {
			response.NotFound(c, "Payment order not found")
			return
		}
		logger.Error("Failed to get payment order", logger.String("payment_no", paymentNo), logger.ErrorField(err))
		response.InternalServerError(c, "Failed to get payment order")
		return
	}

	// Check if user has access to this payment order
	if !user.IsAdmin() && paymentRecord.UserID != user.ID {
		response.Forbidden(c, "You can only access your own payment orders")
		return
	}

	response.OK(c, dto.ToPaymentRecordUserResponse(paymentRecord))
}

// PaymentNotify godoc
// @Summary [Webhook] Payment notification callback
// @Description Handle payment notification callbacks from payment gateways. Supports epay gateway with MD5 signature verification and anti-replay protection. Updates payment status and triggers subscription activation on successful payment.
// @Tags User-Payment
// @Accept application/x-www-form-urlencoded
// @Accept json
// @Produce plain
// @Param gateway path string true "Payment gateway name" Enums(epay)
// @Success 200 {string} string "success" "Payment processed successfully"
// @Failure 400 {string} string "fail" "Invalid gateway, signature verification failed, or malformed request"
// @Failure 500 {string} string "fail" "Internal processing error or database failure"
// @Router /payment/notify/{gateway} [post]
func (h *PaymentOrderHandler) PaymentNotify(c *gin.Context) {
	gateway := c.Param("gateway")
	if gateway == "" {
		c.String(400, "fail")
		return
	}

	// SECURITY: Validate gateway parameter
	validGateways := []string{constants.PaymentGatewayEpay}
	isValidGateway := false
	for _, validGateway := range validGateways {
		if gateway == validGateway {
			isValidGateway = true
			break
		}
	}
	if !isValidGateway {
		logger.Warn("Invalid payment gateway in notification",
			logger.String("gateway", gateway),
			logger.String("client_ip", c.ClientIP()))
		c.String(400, "fail")
		return
	}

	// Get notification data from context (set by security middleware)
	var notifyData map[string]any
	if data, exists := c.Get("payment_request_data"); exists {
		if requestData, ok := data.(map[string]any); ok {
			notifyData = requestData
		}
	}

	// Fallback: parse data if not provided by middleware (backward compatibility)
	if notifyData == nil {
		var err error
		notifyData, err = h.parseNotificationData(c, gateway)
		if err != nil {
			logger.Error("Failed to parse notification data",
				logger.ErrorField(err),
				logger.String("gateway", gateway),
				logger.String("client_ip", c.ClientIP()))
			c.String(400, "fail")
			return
		}
	}

	// Set client IP in context for payment service
	ctx := context.WithValue(c.Request.Context(), "client_ip", c.ClientIP())

	// Process notification
	if err := h.paymentService.ProcessNotification(ctx, gateway, notifyData); err != nil {
		logger.Error("Failed to process payment notification",
			logger.ErrorField(err),
			logger.String("gateway", gateway))
		c.String(500, "fail")
		return
	}

	// Return success response based on gateway
	switch gateway {
	case constants.PaymentGatewayEpay:
		c.String(200, "success")
	default:
		c.String(200, "success")
	}
}

// parseNotificationData parses notification data for backward compatibility
func (h *PaymentOrderHandler) parseNotificationData(c *gin.Context, gateway string) (map[string]any, error) {
	// SECURITY: Validate request size to prevent DoS
	const maxRequestSize = 1024 * 1024 // 1MB
	if c.Request.ContentLength > maxRequestSize {
		logger.Warn("Payment notification request too large",
			logger.Int64("content_length", c.Request.ContentLength),
			logger.String("gateway", gateway),
			logger.String("client_ip", c.ClientIP()))
		return nil, fmt.Errorf("request too large")
	}

	var notifyData map[string]any

	contentType := c.GetHeader("Content-Type")
	if contentType == "application/json" {
		// JSON format (EPUSDT)
		if err := c.ShouldBindJSON(&notifyData); err != nil {
			return nil, fmt.Errorf("failed to parse JSON: %w", err)
		}

		// SECURITY: Validate JSON data structure
		if len(notifyData) == 0 {
			return nil, fmt.Errorf("empty notification data")
		}

	} else {
		// Form format (Epay)
		if err := c.Request.ParseForm(); err != nil {
			return nil, fmt.Errorf("failed to parse form: %w", err)
		}

		notifyData = make(map[string]any)
		for key, values := range c.Request.PostForm {
			if len(values) > 0 {
				// SECURITY: Limit form field length to prevent DoS
				if len(values[0]) > 1000 {
					logger.Warn("Form field too long in payment notification",
						logger.String("field", key),
						logger.Int("length", len(values[0])),
						logger.String("gateway", gateway))
					return nil, fmt.Errorf("form field too long")
				}
				notifyData[key] = values[0]
			}
		}

		// SECURITY: Validate form data structure
		if len(notifyData) == 0 {
			return nil, fmt.Errorf("empty form data")
		}
	}

	return notifyData, nil
}
