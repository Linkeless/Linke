package handlers

import (
	"strconv"
	"strings"

	"linke/internal/domains/payment/dto"
	"linke/internal/domains/payment/usecases/interfaces"
	userEntities "linke/internal/domains/user/entities"
	"linke/internal/shared/logger"
	"linke/internal/shared/middleware"
	"linke/internal/shared/response"

	"github.com/gin-gonic/gin"
)

// PaymentConfigHandler handles HTTP requests for payment configuration management
type PaymentConfigHandler struct {
	paymentConfigService interfaces.PaymentConfigService
}

// NewPaymentConfigHandler creates a new payment config handler
func NewPaymentConfigHandler(paymentConfigService interfaces.PaymentConfigService) *PaymentConfigHandler {
	return &PaymentConfigHandler{
		paymentConfigService: paymentConfigService,
	}
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
func (h *PaymentConfigHandler) CreatePaymentConfig(c *gin.Context) {
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

	// Validate epay configuration
	if validationErrors := dto.ValidateEpayConfig(&req); len(validationErrors) > 0 {
		response.BadRequest(c, "Configuration validation failed", strings.Join(validationErrors, "; "))
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
func (h *PaymentConfigHandler) GetPaymentConfigs(c *gin.Context) {
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
func (h *PaymentConfigHandler) UpdatePaymentConfig(c *gin.Context) {
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

	// Get existing config to determine payment method for validation
	existingConfig, err := h.paymentConfigService.GetPaymentConfig(c.Request.Context(), uint(configID))
	if err != nil {
		if err.Error() == "payment config not found" {
			response.NotFound(c, "Payment config not found")
			return
		}
		logger.Error("Failed to get payment config for validation", logger.ErrorField(err))
		response.InternalServerError(c, "Failed to get payment config", err.Error())
		return
	}

	// Validate epay configuration if it's an epay config
	if validationErrors := dto.ValidateEpayUpdateConfig(&req, existingConfig.Method); len(validationErrors) > 0 {
		response.BadRequest(c, "Configuration validation failed", strings.Join(validationErrors, "; "))
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
func (h *PaymentConfigHandler) DeletePaymentConfig(c *gin.Context) {
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
func (h *PaymentConfigHandler) GetActivePaymentConfigs(c *gin.Context) {
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