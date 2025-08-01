package api

import (
	"github.com/gin-gonic/gin"

	"linke/internal/payment/service/query"
	"linke/internal/response"
)

// PaymentConfigHandler handles public payment config operations
type PaymentConfigHandler struct {
	queryHandler *query.PaymentConfigQueryHandler
}

// NewPaymentConfigHandler creates a new PaymentConfigHandler
func NewPaymentConfigHandler(queryHandler *query.PaymentConfigQueryHandler) *PaymentConfigHandler {
	return &PaymentConfigHandler{
		queryHandler: queryHandler,
	}
}

// GetActivePaymentConfigs godoc
// @Summary Get active payment configurations
// @Description Get all active payment gateway configurations available for users
// @Tags Public - Payment Configs
// @Accept json
// @Produce json
// @Param currency query string false "Filter by supported currency"
// @Success 200 {object} response.StandardResponse{data=query.PublicPaymentConfigListResult}
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/payment-configs [get]
func (h *PaymentConfigHandler) GetActivePaymentConfigs(c *gin.Context) {
	var queryReq query.GetActivePaymentConfigsQuery
	if err := c.ShouldBindQuery(&queryReq); err != nil {
		response.BadRequest(c, "Invalid query parameters", err.Error())
		return
	}

	result, err := h.queryHandler.GetActivePaymentConfigs(c.Request.Context(), queryReq)
	if err != nil {
		response.InternalServerError(c, "Failed to get active payment configs")
		return
	}

	response.OK(c, "Active payment configurations retrieved successfully", result)
}

// GetPaymentConfigsByCurrency godoc
// @Summary Get payment configs by currency
// @Description Get payment configurations that support a specific currency
// @Tags Public - Payment Configs
// @Accept json
// @Produce json
// @Param currency path string true "Currency code (e.g., USD, CNY)"
// @Success 200 {object} response.StandardResponse{data=[]query.PublicPaymentConfigDTO}
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/payment-configs/currency/{currency} [get]
func (h *PaymentConfigHandler) GetPaymentConfigsByCurrency(c *gin.Context) {
	currency := c.Param("currency")
	if currency == "" {
		response.BadRequest(c, "Currency code is required")
		return
	}

	queryReq := query.GetPaymentConfigsByCurrencyQuery{
		Currency: currency,
	}

	configs, err := h.queryHandler.GetPaymentConfigsByCurrency(c.Request.Context(), queryReq)
	if err != nil {
		response.InternalServerError(c, "Failed to get payment configs by currency")
		return
	}

	// Convert to public DTOs
	publicConfigs := make([]query.PublicPaymentConfigDTO, 0, len(configs))
	for _, config := range configs {
		if config.IsActive {
			publicConfig := query.PublicPaymentConfigDTO{
				ID:                  config.ID,
				Gateway:             config.Gateway,
				Name:                config.Name,
				SortOrder:           config.SortOrder,
				SupportedCurrencies: config.SupportedCurrencies,
				SupportedMethods:    filterActiveMethods(config.SupportedMethods),
				MinAmount:           config.MinAmount,
				MaxAmount:           config.MaxAmount,
				GatewayDisplay:      config.GatewayDisplay,
				GatewayType:         config.GatewayType,
				MethodCount:         config.ActiveMethodCount,
			}
			publicConfigs = append(publicConfigs, publicConfig)
		}
	}

	response.OK(c, "Payment configurations retrieved successfully", publicConfigs)
}

// GetPaymentConfigsByMethod godoc
// @Summary Get payment configs by method
// @Description Get payment configurations that support a specific payment method
// @Tags Public - Payment Configs
// @Accept json
// @Produce json
// @Param method path string true "Payment method (e.g., alipay, wechat, credit_card)"
// @Success 200 {object} response.StandardResponse{data=[]query.PublicPaymentConfigDTO}
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/payment-configs/method/{method} [get]
func (h *PaymentConfigHandler) GetPaymentConfigsByMethod(c *gin.Context) {
	method := c.Param("method")
	if method == "" {
		response.BadRequest(c, "Payment method is required")
		return
	}

	queryReq := query.GetPaymentConfigsByMethodQuery{
		Method: method,
	}

	configs, err := h.queryHandler.GetPaymentConfigsByMethod(c.Request.Context(), queryReq)
	if err != nil {
		response.InternalServerError(c, "Failed to get payment configs by method")
		return
	}

	// Convert to public DTOs
	publicConfigs := make([]query.PublicPaymentConfigDTO, 0, len(configs))
	for _, config := range configs {
		if config.IsActive {
			publicConfig := query.PublicPaymentConfigDTO{
				ID:                  config.ID,
				Gateway:             config.Gateway,
				Name:                config.Name,
				SortOrder:           config.SortOrder,
				SupportedCurrencies: config.SupportedCurrencies,
				SupportedMethods:    filterActiveMethods(config.SupportedMethods),
				MinAmount:           config.MinAmount,
				MaxAmount:           config.MaxAmount,
				GatewayDisplay:      config.GatewayDisplay,
				GatewayType:         config.GatewayType,
				MethodCount:         config.ActiveMethodCount,
			}
			publicConfigs = append(publicConfigs, publicConfig)
		}
	}

	response.OK(c, "Payment configurations retrieved successfully", publicConfigs)
}

// GetPaymentMethods godoc
// @Summary Get available payment methods
// @Description Get all available payment methods from active configurations
// @Tags Public - Payment Configs
// @Accept json
// @Produce json
// @Param currency query string false "Filter by currency support"
// @Success 200 {object} response.StandardResponse{data=[]query.PaymentMethodConfigDTO}
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /api/payment-methods [get]
func (h *PaymentConfigHandler) GetPaymentMethods(c *gin.Context) {
	currency := c.Query("currency")
	
	var queryReq query.GetActivePaymentConfigsQuery
	if currency != "" {
		queryReq.Currency = currency
	}

	result, err := h.queryHandler.GetActivePaymentConfigs(c.Request.Context(), queryReq)
	if err != nil {
		response.InternalServerError(c, "Failed to get payment methods")
		return
	}

	// Extract unique methods from all configs
	methodMap := make(map[string]query.PaymentMethodConfigDTO)
	for _, config := range result.Configs {
		for _, method := range config.SupportedMethods {
			if method.IsEnabled {
				// Use method code as key to avoid duplicates
				methodMap[method.Method] = method
			}
		}
	}

	// Convert map to slice
	methods := make([]query.PaymentMethodConfigDTO, 0, len(methodMap))
	for _, method := range methodMap {
		methods = append(methods, method)
	}

	response.OK(c, "Payment methods retrieved successfully", methods)
}

// GetPaymentGateways godoc
// @Summary Get available payment gateways
// @Description Get all active payment gateways with their basic information
// @Tags Public - Payment Configs
// @Accept json
// @Produce json
// @Success 200 {object} response.StandardResponse{data=[]map[string]interface{}}
// @Failure 500 {object} response.ErrorResponse
// @Router /api/payment-gateways [get]
func (h *PaymentConfigHandler) GetPaymentGateways(c *gin.Context) {
	queryReq := query.GetActivePaymentConfigsQuery{}

	result, err := h.queryHandler.GetActivePaymentConfigs(c.Request.Context(), queryReq)
	if err != nil {
		response.InternalServerError(c, "Failed to get payment gateways")
		return
	}

	// Extract gateway information
	gateways := make([]map[string]interface{}, 0, len(result.Configs))
	for _, config := range result.Configs {
		gateway := map[string]interface{}{
			"id":           config.ID,
			"gateway":      config.Gateway,
			"name":         config.Name,
			"display_name": config.GatewayDisplay,
			"type":         config.GatewayType,
			"sort_order":   config.SortOrder,
			"currencies":   config.SupportedCurrencies,
			"method_count": config.MethodCount,
			"min_amount":   config.MinAmount,
			"max_amount":   config.MaxAmount,
		}
		gateways = append(gateways, gateway)
	}

	response.OK(c, "Payment gateways retrieved successfully", gateways)
}

// Helper functions

// filterActiveMethods filters only enabled payment methods
func filterActiveMethods(methods []query.PaymentMethodConfigDTO) []query.PaymentMethodConfigDTO {
	activeMethods := make([]query.PaymentMethodConfigDTO, 0)
	for _, method := range methods {
		if method.IsEnabled {
			activeMethods = append(activeMethods, method)
		}
	}
	return activeMethods
}