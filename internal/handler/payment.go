package handler

import (
	"strconv"

	"linke/internal/logger"
	"linke/internal/middleware"
	"linke/internal/model"
	"linke/internal/response"
	"linke/internal/service"

	"github.com/gin-gonic/gin"
)

type PaymentHandler struct {
	paymentService       *service.PaymentService
	paymentConfigService *service.PaymentConfigService
}

func NewPaymentHandler(paymentService *service.PaymentService, paymentConfigService *service.PaymentConfigService) *PaymentHandler {
	return &PaymentHandler{
		paymentService:       paymentService,
		paymentConfigService: paymentConfigService,
	}
}

// CreatePaymentRequest represents the request to create a payment
type CreatePaymentRequest struct {
	InvoiceID       uint    `json:"invoice_id" binding:"required" example:"1"`
	PaymentGateway  string  `json:"payment_gateway" binding:"required" example:"stripe"`
	PaymentMethod   string  `json:"payment_method" binding:"required" example:"credit_card"`
	Amount          float64 `json:"amount" binding:"required,gt=0" example:"29.99"`
	Currency        string  `json:"currency" binding:"required" example:"USD"`
	Description     string  `json:"description,omitempty" example:"Invoice payment"`
	ReturnURL       string  `json:"return_url,omitempty" example:"https://example.com/payment/return"`
	Reference       string  `json:"reference,omitempty" example:"custom-ref-123"`
}

// CreatePayment godoc
// @Summary [User] Create payment
// @Description Create a new payment for an invoice
// @Tags User-Payment
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param payment body CreatePaymentRequest true "Payment data"
// @Success 201 {object} response.StandardResponse{data=model.PaymentResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /payments [post]
func (h *PaymentHandler) CreatePayment(c *gin.Context) {
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

	// Bind request
	var req CreatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request data", err.Error())
		return
	}

	// Create payment service request
	paymentReq := &service.CreatePaymentRequest{
		InvoiceID:      req.InvoiceID,
		UserID:         user.ID,
		Amount:         req.Amount,
		Currency:       req.Currency,
		PaymentGateway: req.PaymentGateway,
		PaymentMethod:  req.PaymentMethod,
		Description:    req.Description,
		Reference:      req.Reference,
	}

	// Create payment
	payment, err := h.paymentService.CreatePayment(c.Request.Context(), paymentReq)
	if err != nil {
		logger.Error("Failed to create payment", logger.Error2("error", err), logger.Uint("user_id", user.ID))
		response.InternalServerError(c, "Failed to create payment", err.Error())
		return
	}

	response.CreatedWithMessage(c, "Payment created successfully", payment.ToResponse())
}

// GetPayment godoc
// @Summary [User] Get payment
// @Description Get payment details
// @Tags User-Payment
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param payment_id path string true "Payment ID"
// @Success 200 {object} response.StandardResponse{data=model.PaymentResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /payments/{payment_id} [get]
func (h *PaymentHandler) GetPayment(c *gin.Context) {
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

	// Get payment ID from path
	paymentIDStr := c.Param("payment_id")
	if paymentIDStr == "" {
		response.BadRequest(c, "Payment ID is required")
		return
	}

	paymentID, err := strconv.ParseUint(paymentIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid payment ID")
		return
	}

	// Get payment record
	payment, err := h.paymentService.GetPayment(c.Request.Context(), uint(paymentID))
	if err != nil {
		if err.Error() == "payment not found" {
			response.NotFound(c, "Payment not found")
			return
		}
		logger.Error("Failed to get payment", logger.Error2("error", err), logger.Uint("payment_id", uint(paymentID)))
		response.InternalServerError(c, "Failed to get payment", err.Error())
		return
	}

	// Check if user has access to this payment
	if !user.IsAdmin() && payment.UserID != user.ID {
		response.Forbidden(c, "You can only access your own payments")
		return
	}

	response.OK(c, "Payment retrieved successfully", payment.ToResponse())
}

// GetMyPayments godoc
// @Summary [User] Get my payments
// @Description Get current user's payments
// @Tags User-Payment
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Limit results" minimum(1) maximum(100) example(10)
// @Param offset query int false "Offset results" minimum(0) example(0)
// @Success 200 {object} response.PaginatedResponse{data=[]model.PaymentResponse}
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /payments/my [get]
func (h *PaymentHandler) GetMyPayments(c *gin.Context) {
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

	// Get user payments
	userID := user.ID
	filters := &service.PaymentFilters{
		UserID: &userID,
		Limit:  limit,
		Offset: offset,
	}

	payments, totalCount, err := h.paymentService.ListPayments(c.Request.Context(), filters)
	if err != nil {
		logger.Error("Failed to get user payments", logger.Error2("error", err), logger.Uint("user_id", user.ID))
		response.InternalServerError(c, "Failed to get payments", err.Error())
		return
	}

	// Convert to response format
	var paymentResponses []interface{}
	for _, payment := range payments {
		paymentResponses = append(paymentResponses, payment.ToResponse())
	}

	response.OKPaginated(c, "My payments retrieved successfully", paymentResponses, totalCount, limit, offset)
}

// GetAvailablePaymentMethods godoc
// @Summary [Public] Get available payment methods
// @Description Get available payment methods grouped by gateway
// @Tags User-Payment
// @Accept json
// @Produce json
// @Success 200 {object} response.StandardResponse{data=map[string][]string}
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /payments/methods [get]
func (h *PaymentHandler) GetAvailablePaymentMethods(c *gin.Context) {
	// TODO: Implement payment methods discovery for new business flow
	// For now, return a static list
	methods := map[string][]string{
		"stripe": {"credit_card", "bank_transfer"},
		"paypal": {"paypal_account"},
	}
	response.OK(c, "Available payment methods retrieved successfully", methods)
}

// PaymentNotify godoc
// @Summary [Webhook] Payment notification
// @Description Handle payment notification from gateway
// @Tags User-Payment
// @Accept json
// @Produce json
// @Param gateway path string true "Payment gateway"
// @Success 200 {string} string "success"
// @Failure 400 {string} string "fail"
// @Failure 500 {string} string "fail"
// @Router /payments/notify/{gateway} [post]
func (h *PaymentHandler) PaymentNotify(c *gin.Context) {
	gateway := c.Param("gateway")
	if gateway == "" {
		c.String(400, "fail")
		return
	}

	// TODO: Implement payment gateway notification handling for new business flow
	// This is a simplified placeholder implementation
	logger.Info("Payment notification received", 
		logger.String("gateway", gateway),
		logger.String("client_ip", c.ClientIP()))
	
	// For now, always return success
	c.String(200, "success")
}

// CompletePayment godoc
// @Summary [Admin] Complete payment
// @Description Mark a payment as completed
// @Tags Admin-Payment
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param payment_id path string true "Payment ID"
// @Success 200 {object} response.StandardResponse{data=model.PaymentResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/payments/{payment_id}/complete [patch]
func (h *PaymentHandler) CompletePayment(c *gin.Context) {
	// Get payment ID from path
	paymentIDStr := c.Param("payment_id")
	if paymentIDStr == "" {
		response.BadRequest(c, "Payment ID is required")
		return
	}

	paymentID, err := strconv.ParseUint(paymentIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid payment ID")
		return
	}

	// Complete payment
	err = h.paymentService.CompletePayment(c.Request.Context(), uint(paymentID), "Admin completed")
	if err != nil {
		logger.Error("Failed to complete payment", logger.Error2("error", err), logger.Uint("payment_id", uint(paymentID)))
		response.InternalServerError(c, "Failed to complete payment", err.Error())
		return
	}

	// Get updated payment
	payment, err := h.paymentService.GetPayment(c.Request.Context(), uint(paymentID))
	if err != nil {
		logger.Error("Failed to get completed payment", logger.Error2("error", err), logger.Uint("payment_id", uint(paymentID)))
		response.InternalServerError(c, "Failed to get updated payment", err.Error())
		return
	}

	response.OK(c, "Payment completed successfully", payment.ToResponse())
}