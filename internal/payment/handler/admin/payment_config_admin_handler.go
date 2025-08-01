package admin

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"linke/internal/payment/service/command"
	"linke/internal/payment/service/query"
	"linke/internal/response"
)

// PaymentConfigAdminHandler handles admin payment config operations
type PaymentConfigAdminHandler struct {
	commandHandler *command.PaymentConfigCommandHandler
	queryHandler   *query.PaymentConfigQueryHandler
}

// NewPaymentConfigAdminHandler creates a new PaymentConfigAdminHandler
func NewPaymentConfigAdminHandler(
	commandHandler *command.PaymentConfigCommandHandler,
	queryHandler *query.PaymentConfigQueryHandler,
) *PaymentConfigAdminHandler {
	return &PaymentConfigAdminHandler{
		commandHandler: commandHandler,
		queryHandler:   queryHandler,
	}
}

// CreatePaymentConfig godoc
// @Summary Create a new payment config
// @Description Create a new payment gateway configuration (Admin)
// @Tags Admin - Payment Configs
// @Accept json
// @Produce json
// @Param request body command.CreatePaymentConfigCommand true "Create payment config request"
// @Success 201 {object} response.StandardResponse{data=command.CreatePaymentConfigResult}
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /admin/payment-configs [post]
// @Security BearerAuth
func (h *PaymentConfigAdminHandler) CreatePaymentConfig(c *gin.Context) {
	var cmd command.CreatePaymentConfigCommand
	if err := c.ShouldBindJSON(&cmd); err != nil {
		response.BadRequest(c, "Invalid request body", err.Error())
		return
	}

	result, err := h.commandHandler.CreatePaymentConfig(c.Request.Context(), cmd)
	if err != nil {
		response.InternalServerError(c, "Failed to create payment config", err.Error())
		return
	}

	response.CreatedWithMessage(c, "Payment config created successfully", result)
}

// GetPaymentConfig godoc
// @Summary Get payment config by ID
// @Description Get detailed payment config information by ID (Admin)
// @Tags Admin - Payment Configs
// @Accept json
// @Produce json
// @Param id path int true "Payment Config ID"
// @Success 200 {object} response.StandardResponse{data=query.PaymentConfigDTO}
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /admin/payment-configs/{id} [get]
// @Security BearerAuth
func (h *PaymentConfigAdminHandler) GetPaymentConfig(c *gin.Context) {
	configID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid payment config ID")
		return
	}

	queryReq := query.GetPaymentConfigQuery{
		ConfigID: uint(configID),
	}

	config, err := h.queryHandler.GetPaymentConfig(c.Request.Context(), queryReq)
	if err != nil {
		if err.Error() == "payment config not found" {
			response.NotFound(c, "Payment config not found")
			return
		}
		response.InternalServerError(c, "Failed to get payment config")
		return
	}

	response.OK(c, "Payment config retrieved successfully", config)
}

// ListPaymentConfigs godoc
// @Summary List payment configs with filters
// @Description List all payment configs with filtering and pagination (Admin)
// @Tags Admin - Payment Configs
// @Accept json
// @Produce json
// @Param gateway query string false "Payment gateway"
// @Param is_enabled query bool false "Is enabled"
// @Param currency query string false "Supported currency"
// @Param method query string false "Supported method"
// @Param sort_by query string false "Sort field" Enums(sort_order,created_at,updated_at,gateway,name)
// @Param sort_order query string false "Sort order" Enums(asc,desc)
// @Param limit query int false "Limit" default(20)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} response.StandardListResponse{data=query.PaymentConfigListResult}
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /admin/payment-configs [get]
// @Security BearerAuth
func (h *PaymentConfigAdminHandler) ListPaymentConfigs(c *gin.Context) {
	var queryReq query.ListPaymentConfigsQuery
	if err := c.ShouldBindQuery(&queryReq); err != nil {
		response.BadRequest(c, "Invalid query parameters", err.Error())
		return
	}

	// Set default values
	if queryReq.Limit == 0 {
		queryReq.Limit = 20
	}

	result, err := h.queryHandler.ListPaymentConfigs(c.Request.Context(), queryReq)
	if err != nil {
		response.InternalServerError(c, "Failed to list payment configs")
		return
	}

	response.OKPaginated(c, "Payment configs retrieved successfully", result.Configs, result.TotalCount, result.Limit, result.Offset)
}

// UpdatePaymentConfig godoc
// @Summary Update a payment config
// @Description Update payment config details (Admin)
// @Tags Admin - Payment Configs
// @Accept json
// @Produce json
// @Param id path int true "Payment Config ID"
// @Param request body command.UpdatePaymentConfigCommand true "Update payment config request"
// @Success 200 {object} response.StandardResponse{data=command.PaymentConfigCommandResult}
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /admin/payment-configs/{id} [put]
// @Security BearerAuth
func (h *PaymentConfigAdminHandler) UpdatePaymentConfig(c *gin.Context) {
	configID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid payment config ID")
		return
	}

	var cmd command.UpdatePaymentConfigCommand
	if err := c.ShouldBindJSON(&cmd); err != nil {
		response.BadRequest(c, "Invalid request body", err.Error())
		return
	}

	cmd.ConfigID = uint(configID)

	result, err := h.commandHandler.UpdatePaymentConfig(c.Request.Context(), cmd)
	if err != nil {
		response.InternalServerError(c, "Failed to update payment config", err.Error())
		return
	}

	response.OK(c, "Payment config updated successfully", result)
}

// EnablePaymentConfig godoc
// @Summary Enable a payment config
// @Description Enable a payment gateway configuration (Admin)
// @Tags Admin - Payment Configs
// @Accept json
// @Produce json
// @Param id path int true "Payment Config ID"
// @Success 200 {object} response.StandardResponse{data=command.PaymentConfigCommandResult}
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /admin/payment-configs/{id}/enable [post]
// @Security BearerAuth
func (h *PaymentConfigAdminHandler) EnablePaymentConfig(c *gin.Context) {
	configID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid payment config ID")
		return
	}

	cmd := command.EnablePaymentConfigCommand{
		ConfigID: uint(configID),
	}

	result, err := h.commandHandler.EnablePaymentConfig(c.Request.Context(), cmd)
	if err != nil {
		response.InternalServerError(c, "Failed to enable payment config", err.Error())
		return
	}

	response.OK(c, "Payment config enabled successfully", result)
}

// DisablePaymentConfig godoc
// @Summary Disable a payment config
// @Description Disable a payment gateway configuration (Admin)
// @Tags Admin - Payment Configs
// @Accept json
// @Produce json
// @Param id path int true "Payment Config ID"
// @Success 200 {object} response.StandardResponse{data=command.PaymentConfigCommandResult}
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /admin/payment-configs/{id}/disable [post]
// @Security BearerAuth
func (h *PaymentConfigAdminHandler) DisablePaymentConfig(c *gin.Context) {
	configID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid payment config ID")
		return
	}

	cmd := command.DisablePaymentConfigCommand{
		ConfigID: uint(configID),
	}

	result, err := h.commandHandler.DisablePaymentConfig(c.Request.Context(), cmd)
	if err != nil {
		response.InternalServerError(c, "Failed to disable payment config", err.Error())
		return
	}

	response.OK(c, "Payment config disabled successfully", result)
}

// DeletePaymentConfig godoc
// @Summary Delete a payment config
// @Description Soft delete a payment gateway configuration (Admin)
// @Tags Admin - Payment Configs
// @Accept json
// @Produce json
// @Param id path int true "Payment Config ID"
// @Success 200 {object} response.StandardResponse{data=command.PaymentConfigCommandResult}
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /admin/payment-configs/{id} [delete]
// @Security BearerAuth
func (h *PaymentConfigAdminHandler) DeletePaymentConfig(c *gin.Context) {
	configID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid payment config ID")
		return
	}

	cmd := command.DeletePaymentConfigCommand{
		ConfigID: uint(configID),
	}

	result, err := h.commandHandler.DeletePaymentConfig(c.Request.Context(), cmd)
	if err != nil {
		response.InternalServerError(c, "Failed to delete payment config", err.Error())
		return
	}

	response.OK(c, "Payment config deleted successfully", result)
}

// GetPaymentConfigStats godoc
// @Summary Get payment config statistics
// @Description Get payment configuration statistics (Admin)
// @Tags Admin - Payment Configs
// @Accept json
// @Produce json
// @Success 200 {object} response.StandardResponse{data=query.PaymentConfigStatsResult}
// @Failure 500 {object} response.ErrorResponse
// @Router /admin/payment-configs/stats [get]
// @Security BearerAuth
func (h *PaymentConfigAdminHandler) GetPaymentConfigStats(c *gin.Context) {
	stats, err := h.queryHandler.GetPaymentConfigStats(c.Request.Context())
	if err != nil {
		response.InternalServerError(c, "Failed to get payment config stats")
		return
	}

	response.OK(c, "Payment config statistics retrieved successfully", stats)
}

// AddSupportedCurrency godoc
// @Summary Add supported currency to payment config
// @Description Add a new supported currency to a payment config (Admin)
// @Tags Admin - Payment Configs
// @Accept json
// @Produce json
// @Param id path int true "Payment Config ID"
// @Param request body command.AddSupportedCurrencyCommand true "Add currency request"
// @Success 200 {object} response.StandardResponse{data=command.PaymentConfigCommandResult}
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /admin/payment-configs/{id}/currencies [post]
// @Security BearerAuth
func (h *PaymentConfigAdminHandler) AddSupportedCurrency(c *gin.Context) {
	configID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid payment config ID")
		return
	}

	var cmd command.AddSupportedCurrencyCommand
	if err := c.ShouldBindJSON(&cmd); err != nil {
		response.BadRequest(c, "Invalid request body", err.Error())
		return
	}

	cmd.ConfigID = uint(configID)

	result, err := h.commandHandler.AddSupportedCurrency(c.Request.Context(), cmd)
	if err != nil {
		response.InternalServerError(c, "Failed to add supported currency", err.Error())
		return
	}

	response.OK(c, "Supported currency added successfully", result)
}

// RemoveSupportedCurrency godoc
// @Summary Remove supported currency from payment config
// @Description Remove a supported currency from a payment config (Admin)
// @Tags Admin - Payment Configs
// @Accept json
// @Produce json
// @Param id path int true "Payment Config ID"
// @Param currency path string true "Currency Code"
// @Success 200 {object} response.StandardResponse{data=command.PaymentConfigCommandResult}
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /admin/payment-configs/{id}/currencies/{currency} [delete]
// @Security BearerAuth
func (h *PaymentConfigAdminHandler) RemoveSupportedCurrency(c *gin.Context) {
	configID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid payment config ID")
		return
	}

	currency := c.Param("currency")
	if currency == "" {
		response.BadRequest(c, "Currency code is required")
		return
	}

	cmd := command.RemoveSupportedCurrencyCommand{
		ConfigID: uint(configID),
		Currency: currency,
	}

	result, err := h.commandHandler.RemoveSupportedCurrency(c.Request.Context(), cmd)
	if err != nil {
		response.InternalServerError(c, "Failed to remove supported currency", err.Error())
		return
	}

	response.OK(c, "Supported currency removed successfully", result)
}

// AddSupportedMethod godoc
// @Summary Add supported payment method to payment config
// @Description Add a new supported payment method to a payment config (Admin)
// @Tags Admin - Payment Configs
// @Accept json
// @Produce json
// @Param id path int true "Payment Config ID"
// @Param request body command.AddSupportedMethodCommand true "Add method request"
// @Success 200 {object} response.StandardResponse{data=command.PaymentConfigCommandResult}
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /admin/payment-configs/{id}/methods [post]
// @Security BearerAuth
func (h *PaymentConfigAdminHandler) AddSupportedMethod(c *gin.Context) {
	configID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid payment config ID")
		return
	}

	var cmd command.AddSupportedMethodCommand
	if err := c.ShouldBindJSON(&cmd); err != nil {
		response.BadRequest(c, "Invalid request body", err.Error())
		return
	}

	cmd.ConfigID = uint(configID)

	result, err := h.commandHandler.AddSupportedMethod(c.Request.Context(), cmd)
	if err != nil {
		response.InternalServerError(c, "Failed to add supported method", err.Error())
		return
	}

	response.OK(c, "Supported method added successfully", result)
}

// UpdateSupportedMethod godoc
// @Summary Update supported payment method in payment config
// @Description Update a supported payment method configuration (Admin)
// @Tags Admin - Payment Configs
// @Accept json
// @Produce json
// @Param id path int true "Payment Config ID"
// @Param method path string true "Payment Method"
// @Param request body command.UpdateSupportedMethodCommand true "Update method request"
// @Success 200 {object} response.StandardResponse{data=command.PaymentConfigCommandResult}
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /admin/payment-configs/{id}/methods/{method} [put]
// @Security BearerAuth
func (h *PaymentConfigAdminHandler) UpdateSupportedMethod(c *gin.Context) {
	configID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid payment config ID")
		return
	}

	method := c.Param("method")
	if method == "" {
		response.BadRequest(c, "Payment method is required")
		return
	}

	var cmd command.UpdateSupportedMethodCommand
	if err := c.ShouldBindJSON(&cmd); err != nil {
		response.BadRequest(c, "Invalid request body", err.Error())
		return
	}

	cmd.ConfigID = uint(configID)
	cmd.Method = method

	result, err := h.commandHandler.UpdateSupportedMethod(c.Request.Context(), cmd)
	if err != nil {
		response.InternalServerError(c, "Failed to update supported method", err.Error())
		return
	}

	response.OK(c, "Supported method updated successfully", result)
}

// RemoveSupportedMethod godoc
// @Summary Remove supported payment method from payment config
// @Description Remove a supported payment method from a payment config (Admin)
// @Tags Admin - Payment Configs
// @Accept json
// @Produce json
// @Param id path int true "Payment Config ID"
// @Param method path string true "Payment Method"
// @Success 200 {object} response.StandardResponse{data=command.PaymentConfigCommandResult}
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /admin/payment-configs/{id}/methods/{method} [delete]
// @Security BearerAuth
func (h *PaymentConfigAdminHandler) RemoveSupportedMethod(c *gin.Context) {
	configID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid payment config ID")
		return
	}

	method := c.Param("method")
	if method == "" {
		response.BadRequest(c, "Payment method is required")
		return
	}

	cmd := command.RemoveSupportedMethodCommand{
		ConfigID: uint(configID),
		Method:   method,
	}

	result, err := h.commandHandler.RemoveSupportedMethod(c.Request.Context(), cmd)
	if err != nil {
		response.InternalServerError(c, "Failed to remove supported method", err.Error())
		return
	}

	response.OK(c, "Supported method removed successfully", result)
}
