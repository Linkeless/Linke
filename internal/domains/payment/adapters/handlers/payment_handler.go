package handlers

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"linke/internal/domains/payment/constants"
	"linke/internal/domains/payment/dto"
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
	paymentRetryService  interfaces.PaymentRetryService
}

func NewPaymentHandler(paymentService interfaces.PaymentService, paymentConfigService interfaces.PaymentConfigService, paymentRetryService interfaces.PaymentRetryService) *PaymentHandler {
	return &PaymentHandler{
		paymentService:       paymentService,
		paymentConfigService: paymentConfigService,
		paymentRetryService:  paymentRetryService,
	}
}


// CreatePaymentOrder godoc
// @Summary [User] Create payment order
// @Description Create a new payment order. Supports method-based payments (epay, crypto, etc.)
// @Tags User-Payment
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param payment_order body dto.CreatePaymentOrderRequest true "Payment order data with method field"
// @Success 201 {object} response.StandardResponse{data=dto.PaymentRecordResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /payment/orders [post]
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
	var req dto.CreatePaymentOrderRequest
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
		response.InternalServerError(c, "Failed to create payment order", err.Error())
		return
	}

	response.CreatedWithMessage(c, "Payment order created successfully", dto.ToPaymentRecordUserResponse(paymentRecord))
}

// GetPaymentOrder godoc
// @Summary [User] Get payment order
// @Description Get payment order details
// @Tags User-Payment
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param payment_no path string true "Payment number"
// @Success 200 {object} response.StandardResponse{data=dto.PaymentRecordResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /payment/orders/{payment_no} [get]
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
		logger.Error("Failed to get payment order", logger.String("payment_no", paymentNo), logger.ErrorField(err))
		response.InternalServerError(c, "Failed to get payment order", err.Error())
		return
	}

	// Check if user has access to this payment order
	if !user.IsAdmin() && paymentRecord.UserID != user.ID {
		response.Forbidden(c, "You can only access your own payment orders")
		return
	}

	response.OK(c, "Payment order retrieved successfully", dto.ToPaymentRecordUserResponse(paymentRecord))
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
// @Success 200 {object} response.PaginatedResponse{data=[]dto.PaymentRecordResponse}
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /payment/orders/my [get]
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
		logger.Error("Failed to get user payment orders", logger.ErrorField(err), logger.Uint("user_id", user.ID))
		response.InternalServerError(c, "Failed to get payment orders", err.Error())
		return
	}

	// Convert to response format
	var recordResponses []*dto.PaymentRecordResponse
	for _, record := range records {
		recordResponses = append(recordResponses, dto.ToPaymentRecordUserResponse(record))
	}

	response.OKPaginated(c, "My payment orders retrieved successfully", recordResponses, totalCount, limit, offset)
}

// GetAvailablePaymentMethods godoc
// @Summary [Public] Get available payment methods
// @Description Get available payment methods organized by payment method type (epay, crypto, etc.)
// @Tags User-Payment
// @Accept json
// @Produce json
// @Success 200 {object} response.StandardResponse{data=map[string][]string}
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /payment/methods [get]
func (h *PaymentHandler) GetAvailablePaymentMethods(c *gin.Context) {
	// Get available payment methods
	methods, err := h.paymentService.GetAvailablePaymentMethods(c.Request.Context())
	if err != nil {
		logger.Error("Failed to get available payment methods", logger.ErrorField(err))
		response.InternalServerError(c, "Failed to get payment methods", err.Error())
		return
	}

	response.OK(c, "Available payment methods retrieved successfully", methods)
}

// GetActivePaymentConfigs godoc
// @Summary [Public] Get active payment configs
// @Description Get active payment configurations for public display. Returns simplified method-based configs
// @Tags User-Payment
// @Accept json
// @Produce json
// @Param currency query string false "Filter by currency" example("CNY")
// @Success 200 {object} response.StandardResponse{data=[]dto.PaymentConfigResponse}
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /payment/configs [get]
func (h *PaymentHandler) GetActivePaymentConfigs(c *gin.Context) {
	currency := c.Query("currency")

	// Get active payment configs
	configs, err := h.paymentConfigService.GetActivePaymentConfigs(c.Request.Context(), currency)
	if err != nil {
		logger.Error("Failed to get active payment configs", logger.ErrorField(err))
		response.InternalServerError(c, "Failed to get payment configs", err.Error())
		return
	}

	// Convert to public response format
	var configResponses []*dto.PaymentConfigResponse
	for _, config := range configs {
		configResponses = append(configResponses, dto.ToPaymentConfigPublicResponse(config))
	}

	response.OK(c, "Active payment configs retrieved successfully", configResponses)
}

// PaymentNotify godoc
// @Summary [Webhook] Payment notification
// @Description Handle payment notification from payment method. Supports EPay and crypto payment methods
// @Tags User-Payment
// @Accept json
// @Produce json
// @Param method path string true "Payment method" Enums(epay, epusdt, crypto_btc, crypto_usdt)
// @Success 200 {string} string "success"
// @Failure 400 {string} string "fail"
// @Failure 500 {string} string "fail"
// @Router /payment/notify/{method} [post]
func (h *PaymentHandler) PaymentNotify(c *gin.Context) {
	gateway := c.Param("gateway")
	if gateway == "" {
		c.String(400, "fail")
		return
	}

	// SECURITY: Validate gateway parameter
	validGateways := []string{constants.PaymentGatewayEpay, constants.PaymentGatewayEPUSDT}
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
	case constants.PaymentGatewayEPUSDT:
		c.JSON(200, gin.H{"code": 1, "message": "success"})
	default:
		c.String(200, "success")
	}
}

// parseNotificationData parses notification data for backward compatibility
func (h *PaymentHandler) parseNotificationData(c *gin.Context, gateway string) (map[string]any, error) {
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

// CreatePaymentConfig godoc
// @Summary [Admin] Create payment config
// @Description Create a new payment configuration with simplified structure (method + url + pid + key)
// @Tags Admin-Payment-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param config body dto.CreatePaymentConfigRequest true "Payment config data with method (epay/crypto), URL, PID, and Key"
// @Success 201 {object} response.StandardResponse{data=dto.PaymentConfigResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/payment/configs [post]
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
	var req dto.CreatePaymentConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request data", err.Error())
		return
	}

	// Create payment config
	config, err := h.paymentConfigService.CreatePaymentConfig(c.Request.Context(), &req)
	if err != nil {
		logger.Error("Failed to create payment config", logger.ErrorField(err), logger.Uint("admin_id", user.ID))
		response.InternalServerError(c, "Failed to create payment config", err.Error())
		return
	}

	response.CreatedWithMessage(c, "Payment config created successfully", dto.ToPaymentConfigResponse(config))
}

// GetPaymentConfigs godoc
// @Summary [Admin] Get payment configs
// @Description Get payment configurations with filtering and pagination. Configs use method-based structure (epay, crypto_btc, etc.)
// @Tags Admin-Payment-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param method query string false "Filter by payment method" example("epay")
// @Param is_enabled query bool false "Filter by enabled status" example(true)
// @Param limit query int false "Limit results" minimum(1) maximum(100) example(10)
// @Param offset query int false "Offset results" minimum(0) example(0)
// @Success 200 {object} response.PaginatedResponse{data=[]dto.PaymentConfigResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/payment/configs [get]
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
	var req dto.GetPaymentConfigsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "Invalid query parameters", err.Error())
		return
	}

	// Get payment configs
	configs, totalCount, err := h.paymentConfigService.GetPaymentConfigs(c.Request.Context(), &req)
	if err != nil {
		logger.Error("Failed to get payment configs", logger.ErrorField(err))
		response.InternalServerError(c, "Failed to get payment configs", err.Error())
		return
	}

	// Convert to response format
	var configResponses []*dto.PaymentConfigResponse
	for _, config := range configs {
		configResponses = append(configResponses, dto.ToPaymentConfigResponse(config))
	}

	response.OKPaginated(c, "Payment configs retrieved successfully", configResponses, totalCount, req.Limit, req.Offset)
}

// UpdatePaymentConfig godoc
// @Summary [Admin] Update payment config
// @Description Update a payment configuration. Supports method-based structure with URL, PID, Key fields
// @Tags Admin-Payment-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Payment config ID"
// @Param config body dto.UpdatePaymentConfigRequest true "Updated payment config data (URL, PID, Key, etc.)"
// @Success 200 {object} response.StandardResponse{data=dto.PaymentConfigResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/payment/configs/{id} [put]
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
	var req dto.UpdatePaymentConfigRequest
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
		logger.Error("Failed to update payment config", logger.ErrorField(err), logger.Uint("config_id", uint(configID)), logger.Uint("admin_id", user.ID))
		response.InternalServerError(c, "Failed to update payment config", err.Error())
		return
	}

	response.OK(c, "Payment config updated successfully", dto.ToPaymentConfigResponse(config))
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
// @Router /admin/payment/configs/{id} [delete]
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
		logger.Error("Failed to delete payment config", logger.ErrorField(err), logger.Uint("config_id", uint(configID)), logger.Uint("admin_id", user.ID))
		response.InternalServerError(c, "Failed to delete payment config", err.Error())
		return
	}

	response.OK(c, "Payment config deleted successfully", nil)
}

// Dynamic Configuration Endpoints

// GetPaymentMethodSchemas godoc
// @Summary [Admin] Get payment method schemas
// @Description Get configuration schemas for all supported payment methods
// @Tags Admin-Payment-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.StandardResponse{data=map[string]dto.PaymentMethodConfigSchema}
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/payment/schemas [get]
func (h *PaymentHandler) GetPaymentMethodSchemas(c *gin.Context) {
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

	// Get all payment method schemas
	schemas := dto.GetAllPaymentMethodSchemas()
	
	response.OK(c, "Payment method schemas retrieved successfully", schemas)
}

// GetPaymentMethodSchema godoc
// @Summary [Admin] Get payment method schema
// @Description Get configuration schema for a specific payment method
// @Tags Admin-Payment-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param method path string true "Payment method" example("epay")
// @Success 200 {object} response.StandardResponse{data=dto.PaymentMethodConfigSchema}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/payment/schemas/{method} [get]
func (h *PaymentHandler) GetPaymentMethodSchema(c *gin.Context) {
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

	// Get method from path
	method := c.Param("method")
	if method == "" {
		response.BadRequest(c, "Payment method is required")
		return
	}

	// Get payment method schema
	schema, exists := dto.GetPaymentMethodSchema(method)
	if !exists {
		response.NotFound(c, fmt.Sprintf("Payment method schema not found for: %s", method))
		return
	}
	
	response.OK(c, "Payment method schema retrieved successfully", schema)
}

// CreateDynamicPaymentConfig godoc
// @Summary [Admin] Create dynamic payment config
// @Description Create a new payment configuration using dynamic field structure based on payment method
// @Tags Admin-Payment-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param config body dto.DynamicCreatePaymentConfigRequest true "Dynamic payment config data"
// @Success 201 {object} response.StandardResponse{data=dto.DynamicPaymentConfigResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/payment/configs/dynamic [post]
func (h *PaymentHandler) CreateDynamicPaymentConfig(c *gin.Context) {
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
	var req dto.DynamicCreatePaymentConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request data", err.Error())
		return
	}

	// Validate payment method
	schema, exists := dto.GetPaymentMethodSchema(req.Method)
	if !exists {
		response.BadRequest(c, fmt.Sprintf("Unsupported payment method: %s", req.Method))
		return
	}

	// Validate configuration against schema
	if validationErrors, err := dto.ValidatePaymentMethodConfig(req.Method, req.Config); err != nil {
		response.InternalServerError(c, "Validation error", err.Error())
		return
	} else if len(validationErrors) > 0 {
		response.BadRequest(c, "Configuration validation failed", strings.Join(validationErrors, "; "))
		return
	}

	// Convert to standard create request
	standardReq := h.convertDynamicToStandard(req)

	// Create payment config
	config, err := h.paymentConfigService.CreatePaymentConfig(c.Request.Context(), standardReq)
	if err != nil {
		logger.Error("Failed to create dynamic payment config", logger.ErrorField(err), logger.Uint("admin_id", user.ID))
		response.InternalServerError(c, "Failed to create payment config", err.Error())
		return
	}

	// Convert to dynamic response
	dynamicResp := h.convertToDynamicResponse(config, schema)

	response.CreatedWithMessage(c, "Dynamic payment config created successfully", dynamicResp)
}

// UpdateDynamicPaymentConfig godoc
// @Summary [Admin] Update dynamic payment config
// @Description Update a payment configuration using dynamic field structure
// @Tags Admin-Payment-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path uint true "Config ID"
// @Param config body dto.DynamicUpdatePaymentConfigRequest true "Dynamic payment config update data"
// @Success 200 {object} response.StandardResponse{data=dto.DynamicPaymentConfigResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/payment/configs/dynamic/{id} [put]
func (h *PaymentHandler) UpdateDynamicPaymentConfig(c *gin.Context) {
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
	var req dto.DynamicUpdatePaymentConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request data", err.Error())
		return
	}

	// Get existing config to determine payment method
	existingConfig, err := h.paymentConfigService.GetPaymentConfig(c.Request.Context(), uint(configID))
	if err != nil {
		if err.Error() == "payment config not found" {
			response.NotFound(c, "Payment config not found")
			return
		}
		logger.Error("Failed to get payment config", logger.ErrorField(err))
		response.InternalServerError(c, "Failed to get payment config", err.Error())
		return
	}

	// Get schema for validation
	schema, exists := dto.GetPaymentMethodSchema(existingConfig.Method)
	if !exists {
		response.InternalServerError(c, "Payment method schema not found", fmt.Sprintf("Schema not found for method: %s", existingConfig.Method))
		return
	}

	// Validate configuration if provided
	if req.Config != nil && len(req.Config) > 0 {
		if validationErrors, err := dto.ValidatePaymentMethodConfig(existingConfig.Method, req.Config); err != nil {
			response.InternalServerError(c, "Validation error", err.Error())
			return
		} else if len(validationErrors) > 0 {
			response.BadRequest(c, "Configuration validation failed", strings.Join(validationErrors, "; "))
			return
		}
	}

	// Convert to standard update request
	standardReq := h.convertDynamicUpdateToStandard(req)

	// Update payment config
	updatedConfig, err := h.paymentConfigService.UpdatePaymentConfig(c.Request.Context(), uint(configID), standardReq)
	if err != nil {
		if err.Error() == "payment config not found" {
			response.NotFound(c, "Payment config not found")
			return
		}
		logger.Error("Failed to update dynamic payment config", logger.ErrorField(err), logger.Uint("config_id", uint(configID)), logger.Uint("admin_id", user.ID))
		response.InternalServerError(c, "Failed to update payment config", err.Error())
		return
	}

	// Convert to dynamic response
	dynamicResp := h.convertToDynamicResponse(updatedConfig, schema)

	response.OK(c, "Dynamic payment config updated successfully", dynamicResp)
}

// GetDynamicPaymentConfigs godoc
// @Summary [Admin] Get dynamic payment configs
// @Description Get payment configurations with dynamic field structure
// @Tags Admin-Payment-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param method query string false "Filter by payment method" example("epay")
// @Param is_enabled query bool false "Filter by enabled status" example(true)
// @Param limit query int false "Limit results" minimum(1) maximum(100) example(10)
// @Param offset query int false "Offset results" minimum(0) example(0)
// @Success 200 {object} response.PaginatedResponse{data=[]dto.DynamicPaymentConfigResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/payment/configs/dynamic [get]
func (h *PaymentHandler) GetDynamicPaymentConfigs(c *gin.Context) {
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

	// Parse query parameters
	var req dto.GetPaymentConfigsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "Invalid query parameters", err.Error())
		return
	}

	// Set default limit if not provided
	if req.Limit == 0 {
		req.Limit = 10
	}
	if req.Limit > 100 {
		req.Limit = 100
	}

	// Get payment configs
	configs, total, err := h.paymentConfigService.GetPaymentConfigs(c.Request.Context(), &req)
	if err != nil {
		logger.Error("Failed to get dynamic payment configs", logger.ErrorField(err))
		response.InternalServerError(c, "Failed to get payment configs", err.Error())
		return
	}

	// Convert to dynamic responses
	var dynamicConfigs []*dto.DynamicPaymentConfigResponse
	for _, config := range configs {
		if schema, exists := dto.GetPaymentMethodSchema(config.Method); exists {
			dynamicResp := h.convertToDynamicResponse(config, schema)
			dynamicConfigs = append(dynamicConfigs, dynamicResp)
		}
	}

	// Calculate pagination info
	currentPage := (req.Offset / req.Limit) + 1

	response.SuccessListWithMessage(c, "Dynamic payment configs retrieved successfully", dynamicConfigs, currentPage, req.Limit, total)
}

// Helper functions for dynamic configuration conversion

// convertDynamicToStandard converts a dynamic create request to standard format
func (h *PaymentHandler) convertDynamicToStandard(req dto.DynamicCreatePaymentConfigRequest) *dto.CreatePaymentConfigRequest {
	// Extract standard fields from config
	url := h.getStringFromConfig(req.Config, "url")
	pid := h.getStringFromConfig(req.Config, "pid")
	key := h.getStringFromConfig(req.Config, "key")
	notifyURL := h.getStringFromConfig(req.Config, "notify_url")
	returnURL := h.getStringFromConfig(req.Config, "return_url")
	
	// Build supported currencies from config if available
	supportedCurrencies := "CNY" // default
	if schema, exists := dto.GetPaymentMethodSchema(req.Method); exists {
		if len(schema.SupportedCurrencies) > 0 {
			supportedCurrencies = strings.Join(schema.SupportedCurrencies, ",")
		}
	}

	return &dto.CreatePaymentConfigRequest{
		Method:              req.Method,
		Name:                req.Name,
		URL:                 url,
		PID:                 pid,
		Key:                 key,
		NotifyURL:           notifyURL,
		ReturnURL:           returnURL,
		IsEnabled:           req.IsEnabled,
		SortOrder:           req.SortOrder,
		SupportedCurrencies: supportedCurrencies,
		MinAmount:           req.MinAmount,
		MaxAmount:           req.MaxAmount,
		FixedFee:            req.FixedFee,
		PercentageFee:       req.PercentageFee,
	}
}

// convertDynamicUpdateToStandard converts a dynamic update request to standard format
func (h *PaymentHandler) convertDynamicUpdateToStandard(req dto.DynamicUpdatePaymentConfigRequest) *dto.UpdatePaymentConfigRequest {
	standardReq := &dto.UpdatePaymentConfigRequest{
		Name:          req.Name,
		IsEnabled:     req.IsEnabled,
		SortOrder:     req.SortOrder,
		MinAmount:     req.MinAmount,
		MaxAmount:     req.MaxAmount,
		FixedFee:      req.FixedFee,
		PercentageFee: req.PercentageFee,
	}

	// Extract standard fields from config if provided
	if req.Config != nil {
		if url := h.getStringFromConfig(req.Config, "url"); url != "" {
			standardReq.URL = &url
		}
		if pid := h.getStringFromConfig(req.Config, "pid"); pid != "" {
			standardReq.PID = &pid
		}
		if key := h.getStringFromConfig(req.Config, "key"); key != "" {
			standardReq.Key = &key
		}
		if notifyURL := h.getStringFromConfig(req.Config, "notify_url"); notifyURL != "" {
			standardReq.NotifyURL = &notifyURL
		}
		if returnURL := h.getStringFromConfig(req.Config, "return_url"); returnURL != "" {
			standardReq.ReturnURL = &returnURL
		}
	}

	return standardReq
}

// convertToDynamicResponse converts a standard config to dynamic response
func (h *PaymentHandler) convertToDynamicResponse(config *entities.PaymentConfig, schema dto.PaymentMethodConfigSchema) *dto.DynamicPaymentConfigResponse {
	// Build config map from standard fields
	configMap := make(map[string]interface{})
	
	// Add standard fields to config map
	configMap["url"] = config.URL
	configMap["pid"] = config.PID
	configMap["key"] = config.Key
	if config.NotifyURL != "" {
		configMap["notify_url"] = config.NotifyURL
	}
	if config.ReturnURL != "" {
		configMap["return_url"] = config.ReturnURL
	}

	// Build required and optional field lists
	var requiredFields []string
	var optionalFields []string
	fieldDescriptions := make(map[string]string)
	
	for _, field := range schema.RequiredFields {
		requiredFields = append(requiredFields, field.Name)
		fieldDescriptions[field.Name] = field.Description
	}
	
	for _, field := range schema.OptionalFields {
		optionalFields = append(optionalFields, field.Name)
		fieldDescriptions[field.Name] = field.Description
	}

	return &dto.DynamicPaymentConfigResponse{
		ID:               config.ID,
		Method:           config.Method,
		Name:             config.Name,
		IsEnabled:        config.IsEnabled,
		SortOrder:        config.SortOrder,
		MinAmount:        config.MinAmount,
		MaxAmount:        config.MaxAmount,
		FixedFee:         config.FixedFee,
		PercentageFee:    config.PercentageFee,
		CreatedAt:        config.CreatedAt,
		UpdatedAt:        config.UpdatedAt,
		Config:           configMap,
		RequiredFields:   requiredFields,
		OptionalFields:   optionalFields,
		FieldDescriptions: fieldDescriptions,
	}
}

// getStringFromConfig safely extracts a string value from config map
func (h *PaymentHandler) getStringFromConfig(config map[string]interface{}, key string) string {
	if config == nil {
		return ""
	}
	if value, exists := config[key]; exists {
		if strValue, ok := value.(string); ok {
			return strValue
		}
	}
	return ""
}

// Retry Management Endpoints

// GetPaymentRetries godoc
// @Summary [Admin] Get payment retries
// @Description Get payment retries with filtering and pagination
// @Tags Admin-Payment-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param user_id query int false "Filter by user ID" example(1)
// @Param method query string false "Filter by payment method" example("epay")
// @Param payment_method query string false "Filter by specific payment method" example("alipay")
// @Param status query string false "Filter by retry status" Enums(pending, in_progress, completed, failed, cancelled) example("pending")
// @Param failure_type query string false "Filter by failure type" Enums(temporary, permanent, network, gateway, business) example("temporary")
// @Param from_date query string false "Filter from date (RFC3339)" example("2024-01-01T00:00:00Z")
// @Param to_date query string false "Filter to date (RFC3339)" example("2024-12-31T23:59:59Z")
// @Param min_attempts query int false "Minimum attempts" example(1)
// @Param max_attempts query int false "Maximum attempts" example(5)
// @Param include_history query bool false "Include retry history" example(false)
// @Param sort_by query string false "Sort by field" Enums(created_at, next_retry_at, attempt_number) example("created_at")
// @Param sort_order query string false "Sort order" Enums(asc, desc) example("desc")
// @Param limit query int false "Limit results" minimum(1) maximum(100) example(20)
// @Param offset query int false "Offset results" minimum(0) example(0)
// @Success 200 {object} response.PaginatedResponse{data=dto.AdminRetryResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/payment/retries [get]
func (h *PaymentHandler) GetPaymentRetries(c *gin.Context) {
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

	// Parse filters
	filters := &dto.AdminRetryFilters{
		RetryFilters: &dto.RetryFilters{},
		Limit:        20,
		Offset:       0,
		SortBy:       "created_at",
		SortOrder:    "desc",
	}

	// Parse query parameters
	if userIDStr := c.Query("user_id"); userIDStr != "" {
		if userID, err := strconv.ParseUint(userIDStr, 10, 32); err == nil {
			userIDUint := uint(userID)
			filters.RetryFilters.UserID = &userIDUint
		}
	}

	if gateway := c.Query("gateway"); gateway != "" {
		filters.RetryFilters.Gateway = &gateway
	}

	if paymentMethod := c.Query("payment_method"); paymentMethod != "" {
		filters.RetryFilters.PaymentMethod = &paymentMethod
	}

	if status := c.Query("status"); status != "" {
		filters.RetryFilters.Status = &status
	}

	if failureType := c.Query("failure_type"); failureType != "" {
		filters.RetryFilters.FailureType = &failureType
	}

	if fromDateStr := c.Query("from_date"); fromDateStr != "" {
		if fromDate, err := time.Parse(time.RFC3339, fromDateStr); err == nil {
			filters.RetryFilters.FromDate = &fromDate
		}
	}

	if toDateStr := c.Query("to_date"); toDateStr != "" {
		if toDate, err := time.Parse(time.RFC3339, toDateStr); err == nil {
			filters.RetryFilters.ToDate = &toDate
		}
	}

	if minAttemptsStr := c.Query("min_attempts"); minAttemptsStr != "" {
		if minAttempts, err := strconv.Atoi(minAttemptsStr); err == nil {
			filters.RetryFilters.MinAttempts = &minAttempts
		}
	}

	if maxAttemptsStr := c.Query("max_attempts"); maxAttemptsStr != "" {
		if maxAttempts, err := strconv.Atoi(maxAttemptsStr); err == nil {
			filters.RetryFilters.MaxAttempts = &maxAttempts
		}
	}

	if includeHistoryStr := c.Query("include_history"); includeHistoryStr == "true" {
		filters.IncludeHistory = true
	}

	if sortBy := c.Query("sort_by"); sortBy != "" {
		filters.SortBy = sortBy
	}

	if sortOrder := c.Query("sort_order"); sortOrder != "" {
		filters.SortOrder = sortOrder
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 && limit <= 100 {
			filters.Limit = limit
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			filters.Offset = offset
		}
	}

	// Get retries
	result, err := h.paymentRetryService.GetRetriesForAdmin(c.Request.Context(), filters)
	if err != nil {
		logger.Error("Failed to get payment retries", logger.ErrorField(err))
		response.InternalServerError(c, "Failed to get payment retries", err.Error())
		return
	}

	response.OKPaginated(c, "Payment retries retrieved successfully", result, result.TotalCount, filters.Limit, filters.Offset)
}

// GetPaymentRetry godoc
// @Summary [Admin] Get payment retry details
// @Description Get detailed payment retry information with history
// @Tags Admin-Payment-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Retry ID"
// @Success 200 {object} response.StandardResponse{data=dto.RetryWithHistory}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/payment/retries/{id} [get]
func (h *PaymentHandler) GetPaymentRetry(c *gin.Context) {
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

	// Parse retry ID
	retryIDStr := c.Param("id")
	retryID, err := strconv.ParseUint(retryIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid retry ID", "Retry ID must be a valid number")
		return
	}

	// Get retry with history
	retryWithHistory, err := h.paymentRetryService.GetRetryWithHistory(c.Request.Context(), uint(retryID))
	if err != nil {
		if err.Error() == "payment retry not found" {
			response.NotFound(c, "Payment retry not found")
			return
		}
		logger.Error("Failed to get payment retry", logger.ErrorField(err), logger.Uint("retry_id", uint(retryID)))
		response.InternalServerError(c, "Failed to get payment retry", err.Error())
		return
	}

	response.OK(c, "Payment retry retrieved successfully", retryWithHistory)
}

// CancelPaymentRetry godoc
// @Summary [Admin] Cancel payment retry
// @Description Cancel a payment retry sequence
// @Tags Admin-Payment-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Retry ID"
// @Param request body CancelRetryRequest true "Cancel reason"
// @Success 200 {object} response.StandardResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/payment/retries/{id}/cancel [post]
func (h *PaymentHandler) CancelPaymentRetry(c *gin.Context) {
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

	// Parse retry ID
	retryIDStr := c.Param("id")
	retryID, err := strconv.ParseUint(retryIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid retry ID", "Retry ID must be a valid number")
		return
	}

	// Bind request
	var req CancelRetryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request data", err.Error())
		return
	}

	// Cancel retry
	if err := h.paymentRetryService.CancelRetry(c.Request.Context(), uint(retryID), req.Reason); err != nil {
		logger.Error("Failed to cancel payment retry", logger.ErrorField(err), logger.Uint("retry_id", uint(retryID)), logger.Uint("admin_id", user.ID))
		response.InternalServerError(c, "Failed to cancel payment retry", err.Error())
		return
	}

	response.OK(c, "Payment retry cancelled successfully", nil)
}

// ResetPaymentRetry godoc
// @Summary [Admin] Reset payment retry
// @Description Reset a payment retry sequence to start over
// @Tags Admin-Payment-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Retry ID"
// @Success 200 {object} response.StandardResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/payment/retries/{id}/reset [post]
func (h *PaymentHandler) ResetPaymentRetry(c *gin.Context) {
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

	// Parse retry ID
	retryIDStr := c.Param("id")
	retryID, err := strconv.ParseUint(retryIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid retry ID", "Retry ID must be a valid number")
		return
	}

	// Reset retry
	if err := h.paymentRetryService.ResetRetry(c.Request.Context(), uint(retryID)); err != nil {
		logger.Error("Failed to reset payment retry", logger.ErrorField(err), logger.Uint("retry_id", uint(retryID)), logger.Uint("admin_id", user.ID))
		response.InternalServerError(c, "Failed to reset payment retry", err.Error())
		return
	}

	response.OK(c, "Payment retry reset successfully", nil)
}

// BulkCancelPaymentRetries godoc
// @Summary [Admin] Bulk cancel payment retries
// @Description Cancel multiple payment retry sequences
// @Tags Admin-Payment-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body BulkRetryActionRequest true "Retry IDs and reason"
// @Success 200 {object} response.StandardResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/payment/retries/bulk/cancel [post]
func (h *PaymentHandler) BulkCancelPaymentRetries(c *gin.Context) {
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
	var req BulkRetryActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request data", err.Error())
		return
	}

	// Validate request
	if len(req.RetryIDs) == 0 {
		response.BadRequest(c, "No retry IDs provided", "At least one retry ID is required")
		return
	}

	if len(req.RetryIDs) > 100 {
		response.BadRequest(c, "Too many retry IDs", "Maximum 100 retries can be processed at once")
		return
	}

	// Bulk cancel retries
	if err := h.paymentRetryService.BulkCancelRetries(c.Request.Context(), req.RetryIDs, req.Reason); err != nil {
		logger.Error("Failed to bulk cancel payment retries", logger.ErrorField(err), logger.Int("count", len(req.RetryIDs)), logger.Uint("admin_id", user.ID))
		response.InternalServerError(c, "Failed to bulk cancel payment retries", err.Error())
		return
	}

	response.OK(c, fmt.Sprintf("Successfully cancelled %d payment retries", len(req.RetryIDs)), nil)
}

// BulkResetPaymentRetries godoc
// @Summary [Admin] Bulk reset payment retries
// @Description Reset multiple payment retry sequences
// @Tags Admin-Payment-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body BulkRetryActionRequest true "Retry IDs"
// @Success 200 {object} response.StandardResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/payment/retries/bulk/reset [post]
func (h *PaymentHandler) BulkResetPaymentRetries(c *gin.Context) {
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
	var req BulkRetryActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request data", err.Error())
		return
	}

	// Validate request
	if len(req.RetryIDs) == 0 {
		response.BadRequest(c, "No retry IDs provided", "At least one retry ID is required")
		return
	}

	if len(req.RetryIDs) > 100 {
		response.BadRequest(c, "Too many retry IDs", "Maximum 100 retries can be processed at once")
		return
	}

	// Bulk reset retries
	if err := h.paymentRetryService.BulkResetRetries(c.Request.Context(), req.RetryIDs); err != nil {
		logger.Error("Failed to bulk reset payment retries", logger.ErrorField(err), logger.Int("count", len(req.RetryIDs)), logger.Uint("admin_id", user.ID))
		response.InternalServerError(c, "Failed to bulk reset payment retries", err.Error())
		return
	}

	response.OK(c, fmt.Sprintf("Successfully reset %d payment retries", len(req.RetryIDs)), nil)
}

// GetRetryStatistics godoc
// @Summary [Admin] Get retry statistics
// @Description Get payment retry statistics for a specific payment method
// @Tags Admin-Payment-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param method query string true "Payment method name" example("epay")
// @Param days query int false "Number of days to analyze" minimum(1) maximum(365) example(30)
// @Success 200 {object} response.StandardResponse{data=dto.RetryStatistics}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/payment/retries/statistics [get]
func (h *PaymentHandler) GetRetryStatistics(c *gin.Context) {
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

	// Get parameters
	gateway := c.Query("gateway")
	if gateway == "" {
		response.BadRequest(c, "Gateway parameter is required", "Please specify a gateway")
		return
	}

	daysStr := c.DefaultQuery("days", "30")
	days, err := strconv.Atoi(daysStr)
	if err != nil || days < 1 || days > 365 {
		days = 30
	}

	// Get statistics
	stats, err := h.paymentRetryService.GetRetryStatistics(c.Request.Context(), gateway, days)
	if err != nil {
		logger.Error("Failed to get retry statistics", logger.String("gateway", gateway), logger.ErrorField(err))
		response.InternalServerError(c, "Failed to get retry statistics", err.Error())
		return
	}

	response.OK(c, "Retry statistics retrieved successfully", stats)
}

// GetRetryHealthMetrics godoc
// @Summary [Admin] Get retry system health metrics
// @Description Get overall health metrics for the retry system
// @Tags Admin-Payment-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.StandardResponse{data=dto.RetryHealthMetrics}
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/payment/retries/health [get]
func (h *PaymentHandler) GetRetryHealthMetrics(c *gin.Context) {
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

	// Get health metrics
	metrics, err := h.paymentRetryService.GetRetryHealthMetrics(c.Request.Context())
	if err != nil {
		logger.Error("Failed to get retry health metrics", logger.ErrorField(err))
		response.InternalServerError(c, "Failed to get retry health metrics", err.Error())
		return
	}

	response.OK(c, "Retry health metrics retrieved successfully", metrics)
}

// Request/Response structures for retry endpoints

// CancelRetryRequest represents the request to cancel a retry
type CancelRetryRequest struct {
	Reason string `json:"reason" binding:"required" example:"Manual cancellation by admin"`
}

// BulkRetryActionRequest represents the request for bulk retry actions
type BulkRetryActionRequest struct {
	RetryIDs []uint `json:"retry_ids" binding:"required"`
	Reason   string `json:"reason,omitempty" example:"Bulk operation by admin"`
}
