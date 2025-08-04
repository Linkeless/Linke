package handlers

import (
	"context"
	"fmt"
	"strconv"
	"time"

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
	notifyURL := "https://" + c.Request.Host + "/api/v1/payment/notify/" + req.Gateway

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
// @Router /payment/methods [get]
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
// @Router /payment/configs [get]
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
// @Router /payment/notify/{gateway} [post]
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

	// Get notification data from context (set by security middleware)
	var notifyData map[string]interface{}
	if data, exists := c.Get("payment_request_data"); exists {
		if requestData, ok := data.(map[string]interface{}); ok {
			notifyData = requestData
		}
	}

	// Fallback: parse data if not provided by middleware (backward compatibility)
	if notifyData == nil {
		var err error
		notifyData, err = h.parseNotificationData(c, gateway)
		if err != nil {
			logger.Error("Failed to parse notification data",
				logger.Error2("error", err),
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

// parseNotificationData parses notification data for backward compatibility
func (h *PaymentHandler) parseNotificationData(c *gin.Context, gateway string) (map[string]interface{}, error) {
	// SECURITY: Validate request size to prevent DoS
	const maxRequestSize = 1024 * 1024 // 1MB
	if c.Request.ContentLength > maxRequestSize {
		logger.Warn("Payment notification request too large",
			logger.Int64("content_length", c.Request.ContentLength),
			logger.String("gateway", gateway),
			logger.String("client_ip", c.ClientIP()))
		return nil, fmt.Errorf("request too large")
	}

	var notifyData map[string]interface{}

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

		notifyData = make(map[string]interface{})
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
		logger.Error("Failed to delete payment config", logger.Error2("error", err), logger.Uint("config_id", uint(configID)), logger.Uint("admin_id", user.ID))
		response.InternalServerError(c, "Failed to delete payment config", err.Error())
		return
	}

	response.OK(c, "Payment config deleted successfully", nil)
}

// Retry Management Endpoints

// GetPaymentRetries godoc
// @Summary [Admin] Get payment retries
// @Description Get payment retries with filtering and pagination
// @Tags Admin-Payment-Retry
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param user_id query int false "Filter by user ID" example(1)
// @Param gateway query string false "Filter by gateway" example("epay")
// @Param payment_method query string false "Filter by payment method" example("alipay")
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
// @Success 200 {object} response.PaginatedResponse{data=interfaces.AdminRetryResponse}
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
	filters := &interfaces.AdminRetryFilters{
		RetryFilters: &interfaces.RetryFilters{},
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
		logger.Error("Failed to get payment retries", logger.Error2("error", err))
		response.InternalServerError(c, "Failed to get payment retries", err.Error())
		return
	}

	response.OKPaginated(c, "Payment retries retrieved successfully", result, result.TotalCount, filters.Limit, filters.Offset)
}

// GetPaymentRetry godoc
// @Summary [Admin] Get payment retry details
// @Description Get detailed payment retry information with history
// @Tags Admin-Payment-Retry
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Retry ID"
// @Success 200 {object} response.StandardResponse{data=interfaces.RetryWithHistory}
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
		logger.Error("Failed to get payment retry", logger.Error2("error", err), logger.Uint("retry_id", uint(retryID)))
		response.InternalServerError(c, "Failed to get payment retry", err.Error())
		return
	}

	response.OK(c, "Payment retry retrieved successfully", retryWithHistory)
}

// CancelPaymentRetry godoc
// @Summary [Admin] Cancel payment retry
// @Description Cancel a payment retry sequence
// @Tags Admin-Payment-Retry
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
		logger.Error("Failed to cancel payment retry", logger.Error2("error", err), logger.Uint("retry_id", uint(retryID)), logger.Uint("admin_id", user.ID))
		response.InternalServerError(c, "Failed to cancel payment retry", err.Error())
		return
	}

	response.OK(c, "Payment retry cancelled successfully", nil)
}

// ResetPaymentRetry godoc
// @Summary [Admin] Reset payment retry
// @Description Reset a payment retry sequence to start over
// @Tags Admin-Payment-Retry
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
		logger.Error("Failed to reset payment retry", logger.Error2("error", err), logger.Uint("retry_id", uint(retryID)), logger.Uint("admin_id", user.ID))
		response.InternalServerError(c, "Failed to reset payment retry", err.Error())
		return
	}

	response.OK(c, "Payment retry reset successfully", nil)
}

// BulkCancelPaymentRetries godoc
// @Summary [Admin] Bulk cancel payment retries
// @Description Cancel multiple payment retry sequences
// @Tags Admin-Payment-Retry
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
		logger.Error("Failed to bulk cancel payment retries", logger.Error2("error", err), logger.Int("count", len(req.RetryIDs)), logger.Uint("admin_id", user.ID))
		response.InternalServerError(c, "Failed to bulk cancel payment retries", err.Error())
		return
	}

	response.OK(c, fmt.Sprintf("Successfully cancelled %d payment retries", len(req.RetryIDs)), nil)
}

// BulkResetPaymentRetries godoc
// @Summary [Admin] Bulk reset payment retries
// @Description Reset multiple payment retry sequences
// @Tags Admin-Payment-Retry
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
		logger.Error("Failed to bulk reset payment retries", logger.Error2("error", err), logger.Int("count", len(req.RetryIDs)), logger.Uint("admin_id", user.ID))
		response.InternalServerError(c, "Failed to bulk reset payment retries", err.Error())
		return
	}

	response.OK(c, fmt.Sprintf("Successfully reset %d payment retries", len(req.RetryIDs)), nil)
}

// GetRetryStatistics godoc
// @Summary [Admin] Get retry statistics
// @Description Get payment retry statistics for a specific gateway
// @Tags Admin-Payment-Retry
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param gateway query string true "Gateway name" example("epay")
// @Param days query int false "Number of days to analyze" minimum(1) maximum(365) example(30)
// @Success 200 {object} response.StandardResponse{data=interfaces.RetryStatistics}
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
		logger.Error("Failed to get retry statistics", logger.Error2("error", err), logger.String("gateway", gateway))
		response.InternalServerError(c, "Failed to get retry statistics", err.Error())
		return
	}

	response.OK(c, "Retry statistics retrieved successfully", stats)
}

// GetRetryHealthMetrics godoc
// @Summary [Admin] Get retry system health metrics
// @Description Get overall health metrics for the retry system
// @Tags Admin-Payment-Retry
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.StandardResponse{data=interfaces.RetryHealthMetrics}
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
		logger.Error("Failed to get retry health metrics", logger.Error2("error", err))
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
