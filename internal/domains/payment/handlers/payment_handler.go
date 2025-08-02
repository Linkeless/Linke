package handlers

import (
	"strconv"

	"linke/internal/domains/payment/entities"
	"linke/internal/domains/payment/usecases/interfaces"
	userEntities "linke/internal/domains/user/entities"
	"linke/internal/shared/logger"
	"linke/internal/shared/middleware"
	"linke/internal/shared/response"

	"github.com/gin-gonic/gin"
)

type PaymentHandler struct {
	paymentService       interfaces.PaymentService
	paymentConfigService interfaces.PaymentConfigService
}

func NewPaymentHandler(paymentService interfaces.PaymentService, paymentConfigService interfaces.PaymentConfigService) *PaymentHandler {
	return &PaymentHandler{
		paymentService:       paymentService,
		paymentConfigService: paymentConfigService,
	}
}

// CreatePaymentOrderRequest represents the request to create a payment order
type CreatePaymentOrderRequest struct {
	SubscriptionOrderID *uint   `json:"subscription_order_id,omitempty" example:"1"`
	InvoiceID           *uint   `json:"invoice_id,omitempty" example:"1"`
	Gateway             string  `json:"gateway" binding:"required" example:"epay"`
	PaymentMethod       string  `json:"payment_method" binding:"required" example:"alipay"`
	Amount              float64 `json:"amount" binding:"required,gt=0" example:"29.99"`
	Currency            string  `json:"currency" binding:"required" example:"CNY"`
	Subject             string  `json:"subject" binding:"required" example:"Premium Subscription"`
	Body                string  `json:"body,omitempty" example:"Monthly premium subscription payment"`
	ReturnURL           string  `json:"return_url,omitempty" example:"https://example.com/payment/return"`
	ExpiredMinutes      int     `json:"expired_minutes,omitempty" example:"30"`
}

// CreatePaymentOrder godoc
// @Summary [User] Create payment order
// @Description Create a new payment order
// @Tags User-Payment
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param payment_order body CreatePaymentOrderRequest true "Payment order data"
// @Success 201 {object} response.StandardResponse{data=entities.PaymentRecordResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /payments/orders [post]
func (h *PaymentHandler) CreatePaymentOrder(c *gin.Context) {
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
	var req CreatePaymentOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request data", err.Error())
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
	notifyURL := "https://" + c.Request.Host + "/api/v1/payments/notify/" + req.Gateway

	// Create payment service request
	paymentReq := &interfaces.CreatePaymentOrderRequest{
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
		logger.Error("Failed to create payment order", logger.Error2("error", err), logger.Uint("user_id", user.ID))
		response.InternalServerError(c, "Failed to create payment order", err.Error())
		return
	}

	response.CreatedWithMessage(c, "Payment order created successfully", paymentRecord.ToUserResponse())
}

// GetPaymentOrder godoc
// @Summary [User] Get payment order
// @Description Get payment order details
// @Tags User-Payment
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param payment_no path string true "Payment number"
// @Success 200 {object} response.StandardResponse{data=entities.PaymentRecordResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /payments/orders/{payment_no} [get]
func (h *PaymentHandler) GetPaymentOrder(c *gin.Context) {
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
		logger.Error("Failed to get payment order", logger.Error2("error", err), logger.String("payment_no", paymentNo))
		response.InternalServerError(c, "Failed to get payment order", err.Error())
		return
	}

	// Check if user has access to this payment order
	if !user.IsAdmin() && paymentRecord.UserID != user.ID {
		response.Forbidden(c, "You can only access your own payment orders")
		return
	}

	response.OK(c, "Payment order retrieved successfully", paymentRecord.ToUserResponse())
}

// GetMyPaymentOrders godoc
// @Summary [User] Get my payment orders
// @Description Get current user's payment orders
// @Tags User-Payment
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Limit results" minimum(1) maximum(100) example(10)
// @Param offset query int false "Offset results" minimum(0) example(0)
// @Success 200 {object} response.PaginatedResponse{data=[]entities.PaymentRecordResponse}
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /payments/orders/my [get]
func (h *PaymentHandler) GetMyPaymentOrders(c *gin.Context) {
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

	// Get user payment records
	records, totalCount, err := h.paymentService.GetUserPaymentRecords(c.Request.Context(), user.ID, limit, offset)
	if err != nil {
		logger.Error("Failed to get user payment orders", logger.Error2("error", err), logger.Uint("user_id", user.ID))
		response.InternalServerError(c, "Failed to get payment orders", err.Error())
		return
	}

	// Convert to response format
	var recordResponses []*entities.PaymentRecordResponse
	for _, record := range records {
		recordResponses = append(recordResponses, record.ToUserResponse())
	}

	response.OKPaginated(c, "My payment orders retrieved successfully", recordResponses, totalCount, limit, offset)
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
	// Get available payment methods
	methods, err := h.paymentService.GetAvailablePaymentMethods(c.Request.Context())
	if err != nil {
		logger.Error("Failed to get available payment methods", logger.Error2("error", err))
		response.InternalServerError(c, "Failed to get payment methods", err.Error())
		return
	}

	response.OK(c, "Available payment methods retrieved successfully", methods)
}

// GetActivePaymentConfigs godoc
// @Summary [Public] Get active payment configs
// @Description Get active payment configurations for public display
// @Tags User-Payment
// @Accept json
// @Produce json
// @Param currency query string false "Filter by currency" example("CNY")
// @Success 200 {object} response.StandardResponse{data=[]entities.PaymentConfigResponse}
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /payments/configs [get]
func (h *PaymentHandler) GetActivePaymentConfigs(c *gin.Context) {
	currency := c.Query("currency")

	// Get active payment configs
	configs, err := h.paymentConfigService.GetActivePaymentConfigs(c.Request.Context(), currency)
	if err != nil {
		logger.Error("Failed to get active payment configs", logger.Error2("error", err))
		response.InternalServerError(c, "Failed to get payment configs", err.Error())
		return
	}

	// Convert to public response format
	var configResponses []*entities.PaymentConfigResponse
	for _, config := range configs {
		configResponses = append(configResponses, config.ToPublicResponse())
	}

	response.OK(c, "Active payment configs retrieved successfully", configResponses)
}

// PaymentNotify godoc
// @Summary [Webhook] Payment notification
// @Description Handle payment notification from gateway
// @Tags User-Payment
// @Accept json
// @Produce json
// @Param gateway path string true "Payment gateway" Enums(epay, epusdt)
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

	// SECURITY: Validate gateway parameter
	validGateways := []string{entities.PaymentGatewayEpay, entities.PaymentGatewayEPUSDT}
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

	// SECURITY: Validate request size to prevent DoS
	const maxRequestSize = 1024 * 1024 // 1MB
	if c.Request.ContentLength > maxRequestSize {
		logger.Warn("Payment notification request too large",
			logger.Int64("content_length", c.Request.ContentLength),
			logger.String("gateway", gateway),
			logger.String("client_ip", c.ClientIP()))
		c.String(400, "fail")
		return
	}

	// Parse notification data based on content type
	var notifyData map[string]interface{}

	contentType := c.GetHeader("Content-Type")
	if contentType == "application/json" {
		// JSON format (EPUSDT)
		if err := c.ShouldBindJSON(&notifyData); err != nil {
			logger.Error("Failed to parse JSON notification",
				logger.Error2("error", err),
				logger.String("gateway", gateway),
				logger.String("client_ip", c.ClientIP()))
			c.String(400, "fail")
			return
		}

		// SECURITY: Validate JSON data structure
		if len(notifyData) == 0 {
			logger.Warn("Empty payment notification data",
				logger.String("gateway", gateway),
				logger.String("client_ip", c.ClientIP()))
			c.String(400, "fail")
			return
		}

	} else {
		// Form format (Epay)
		if err := c.Request.ParseForm(); err != nil {
			logger.Error("Failed to parse form data",
				logger.Error2("error", err),
				logger.String("gateway", gateway),
				logger.String("client_ip", c.ClientIP()))
			c.String(400, "fail")
			return
		}

		notifyData = make(map[string]interface{})
		for key, values := range c.Request.PostForm {
			if len(values) > 0 {
				// SECURITY: Limit form field length to prevent DoS
				if len(values[0]) > 1000 {
					logger.Warn("Form field too long in payment notification",
						logger.String("field", key),
						logger.Int("length", len(values[0])),
						logger.String("gateway", gateway))
					c.String(400, "fail")
					return
				}
				notifyData[key] = values[0]
			}
		}

		// SECURITY: Validate form data structure
		if len(notifyData) == 0 {
			logger.Warn("Empty payment notification form data",
				logger.String("gateway", gateway),
				logger.String("client_ip", c.ClientIP()))
			c.String(400, "fail")
			return
		}
	}

	// Process notification
	if err := h.paymentService.ProcessNotification(c.Request.Context(), gateway, notifyData); err != nil {
		logger.Error("Failed to process payment notification",
			logger.Error2("error", err),
			logger.String("gateway", gateway))
		c.String(500, "fail")
		return
	}

	// Return success response based on gateway
	switch gateway {
	case entities.PaymentGatewayEpay:
		c.String(200, "success")
	case entities.PaymentGatewayEPUSDT:
		c.JSON(200, gin.H{"code": 1, "message": "success"})
	default:
		c.String(200, "success")
	}
}

// CreatePaymentConfig godoc
// @Summary [Admin] Create payment config
// @Description Create a new payment configuration
// @Tags Admin-Payment-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param config body interfaces.CreatePaymentConfigRequest true "Payment config data"
// @Success 201 {object} response.StandardResponse{data=entities.PaymentConfigResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/payments/configs [post]
func (h *PaymentHandler) CreatePaymentConfig(c *gin.Context) {
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

	// Check if user is admin
	if !user.IsAdmin() {
		response.Forbidden(c, "Admin access required")
		return
	}

	// Bind request
	var req interfaces.CreatePaymentConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request data", err.Error())
		return
	}

	// Create payment config
	config, err := h.paymentConfigService.CreatePaymentConfig(c.Request.Context(), &req)
	if err != nil {
		logger.Error("Failed to create payment config", logger.Error2("error", err), logger.Uint("admin_id", user.ID))
		response.InternalServerError(c, "Failed to create payment config", err.Error())
		return
	}

	response.CreatedWithMessage(c, "Payment config created successfully", config.ToResponse())
}

// GetPaymentConfigs godoc
// @Summary [Admin] Get payment configs
// @Description Get payment configurations with filtering and pagination
// @Tags Admin-Payment-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param gateway query string false "Filter by gateway" example("epay")
// @Param method query string false "Filter by method" example("alipay")
// @Param is_enabled query bool false "Filter by enabled status" example(true)
// @Param environment query string false "Filter by environment" example("production")
// @Param limit query int false "Limit results" minimum(1) maximum(100) example(10)
// @Param offset query int false "Offset results" minimum(0) example(0)
// @Success 200 {object} response.PaginatedResponse{data=[]entities.PaymentConfigResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/payments/configs [get]
func (h *PaymentHandler) GetPaymentConfigs(c *gin.Context) {
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

	// Check if user is admin
	if !user.IsAdmin() {
		response.Forbidden(c, "Admin access required")
		return
	}

	// Bind query parameters
	var req interfaces.GetPaymentConfigsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "Invalid query parameters", err.Error())
		return
	}

	// Get payment configs
	configs, totalCount, err := h.paymentConfigService.GetPaymentConfigs(c.Request.Context(), &req)
	if err != nil {
		logger.Error("Failed to get payment configs", logger.Error2("error", err))
		response.InternalServerError(c, "Failed to get payment configs", err.Error())
		return
	}

	// Convert to response format
	var configResponses []*entities.PaymentConfigResponse
	for _, config := range configs {
		configResponses = append(configResponses, config.ToResponse())
	}

	response.OKPaginated(c, "Payment configs retrieved successfully", configResponses, totalCount, req.Limit, req.Offset)
}

// UpdatePaymentConfig godoc
// @Summary [Admin] Update payment config
// @Description Update a payment configuration
// @Tags Admin-Payment-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Payment config ID"
// @Param config body interfaces.UpdatePaymentConfigRequest true "Updated payment config data"
// @Success 200 {object} response.StandardResponse{data=entities.PaymentConfigResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/payments/configs/{id} [put]
func (h *PaymentHandler) UpdatePaymentConfig(c *gin.Context) {
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

	// Check if user is admin
	if !user.IsAdmin() {
		response.Forbidden(c, "Admin access required")
		return
	}

	// Parse config ID
	configIDStr := c.Param("id")
	configID, err := strconv.ParseUint(configIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid config ID", "Config ID must be a valid number")
		return
	}

	// Bind request
	var req interfaces.UpdatePaymentConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request data", err.Error())
		return
	}

	// Update payment config
	config, err := h.paymentConfigService.UpdatePaymentConfig(c.Request.Context(), uint(configID), &req)
	if err != nil {
		if err.Error() == "payment config not found" {
			response.NotFound(c, "Payment config not found")
			return
		}
		logger.Error("Failed to update payment config", logger.Error2("error", err), logger.Uint("config_id", uint(configID)), logger.Uint("admin_id", user.ID))
		response.InternalServerError(c, "Failed to update payment config", err.Error())
		return
	}

	response.OK(c, "Payment config updated successfully", config.ToResponse())
}

// DeletePaymentConfig godoc
// @Summary [Admin] Delete payment config
// @Description Soft delete a payment configuration
// @Tags Admin-Payment-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Payment config ID"
// @Success 200 {object} response.StandardResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/payments/configs/{id} [delete]
func (h *PaymentHandler) DeletePaymentConfig(c *gin.Context) {
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

	// Check if user is admin
	if !user.IsAdmin() {
		response.Forbidden(c, "Admin access required")
		return
	}

	// Parse config ID
	configIDStr := c.Param("id")
	configID, err := strconv.ParseUint(configIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid config ID", "Config ID must be a valid number")
		return
	}

	// Delete payment config
	if err := h.paymentConfigService.DeletePaymentConfig(c.Request.Context(), uint(configID)); err != nil {
		if err.Error() == "payment config not found" {
			response.NotFound(c, "Payment config not found")
			return
		}
		logger.Error("Failed to delete payment config", logger.Error2("error", err), logger.Uint("config_id", uint(configID)), logger.Uint("admin_id", user.ID))
		response.InternalServerError(c, "Failed to delete payment config", err.Error())
		return
	}

	response.OK(c, "Payment config deleted successfully", nil)
}
