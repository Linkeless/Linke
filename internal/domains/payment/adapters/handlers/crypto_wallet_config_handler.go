package handlers

import (
	"strconv"

	"linke/internal/domains/payment/dto"
	"linke/internal/domains/payment/entities"
	"linke/internal/domains/payment/usecases/interfaces"
	"linke/internal/shared/framework"
	"linke/internal/shared/response"

	"github.com/gin-gonic/gin"
)

// CryptoWalletConfigHandler handles HTTP requests for crypto wallet configuration management
type CryptoWalletConfigHandler struct {
	service interfaces.CryptoWalletConfigService
	logger  framework.Logger
}

// NewCryptoWalletConfigHandler creates a new CryptoWalletConfigHandler
func NewCryptoWalletConfigHandler(
	service interfaces.CryptoWalletConfigService,
	logger framework.Logger,
) *CryptoWalletConfigHandler {
	return &CryptoWalletConfigHandler{
		service: service,
		logger:  logger,
	}
}

// parseIDParam parses ID parameter from URL path
func (h *CryptoWalletConfigHandler) parseIDParam(c *gin.Context, paramName string) (uint, error) {
	idStr := c.Param(paramName)
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}

// CreateCryptoWalletConfig creates a new crypto wallet config
// @Summary [Admin] Create crypto wallet config
// @Description Create a new cryptocurrency wallet configuration for receiving payments
// @Tags Admin-Payment-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param config body dto.CreateCryptoWalletConfigRequest true "Crypto wallet config creation request"
// @Success 201 {object} response.StandardResponse{data=dto.CryptoWalletConfigResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/payment/crypto-wallets [post]
func (h *CryptoWalletConfigHandler) CreateCryptoWalletConfig(c *gin.Context) {
	var req dto.CreateCryptoWalletConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body", err.Error())
		return
	}

	// Create the crypto wallet config
	config, err := h.service.CreateCryptoWalletConfig(c.Request.Context(), &req)
	if err != nil {
		response.InternalServerError(c, "Failed to create crypto wallet config", err.Error())
		return
	}

	// Convert to response DTO
	responseData := dto.ToCryptoWalletConfigResponse(config)
	response.CreatedWithMessage(c, "Crypto wallet config created successfully", responseData)
}

// GetCryptoWalletConfig gets a crypto wallet config by ID
// @Summary [Admin] Get crypto wallet config
// @Description Get cryptocurrency wallet configuration by ID
// @Tags Admin-Payment-Management
// @Produce json
// @Security BearerAuth
// @Param id path int true "Crypto wallet config ID"
// @Success 200 {object} response.StandardResponse{data=dto.CryptoWalletConfigResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/payment/crypto-wallets/{id} [get]
func (h *CryptoWalletConfigHandler) GetCryptoWalletConfig(c *gin.Context) {
	configID, err := h.parseIDParam(c, "id")
	if err != nil {
		response.BadRequest(c, "Invalid config ID", err.Error())
		return
	}

	config, err := h.service.GetCryptoWalletConfig(c.Request.Context(), configID)
	if err != nil {
		response.InternalServerError(c, "Failed to get crypto wallet config", err.Error())
		return
	}

	// Convert to response DTO
	responseData := dto.ToCryptoWalletConfigResponse(config)
	response.SuccessWithMessage(c, "Crypto wallet config retrieved successfully", responseData)
}

// GetCryptoWalletConfigs gets crypto wallet configs with filtering
// @Summary [Admin] List crypto wallet configs
// @Description Get list of cryptocurrency wallet configurations with optional filtering
// @Tags Admin-Payment-Management
// @Produce json
// @Security BearerAuth
// @Param network query string false "Filter by network" example("trc")
// @Param currency query string false "Filter by currency" example("USDT")
// @Param is_enabled query bool false "Filter by enabled status" example(true)
// @Param is_active query bool false "Filter by active status" example(true)
// @Param limit query int false "Limit results" minimum(1) maximum(100) example(10)
// @Param offset query int false "Offset results" minimum(0) example(0)
// @Success 200 {object} response.PaginatedResponse{data=dto.CryptoWalletConfigListResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/payment/crypto-wallets [get]
func (h *CryptoWalletConfigHandler) GetCryptoWalletConfigs(c *gin.Context) {
	var req dto.GetCryptoWalletConfigsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "Invalid query parameters", err.Error())
		return
	}

	configs, total, err := h.service.GetCryptoWalletConfigs(c.Request.Context(), &req)
	if err != nil {
		response.InternalServerError(c, "Failed to get crypto wallet configs", err.Error())
		return
	}

	// Convert to response DTOs
	var responses []dto.CryptoWalletConfigResponse
	for _, config := range configs {
		responses = append(responses, *dto.ToCryptoWalletConfigResponse(config))
	}

	responseData := dto.CryptoWalletConfigListResponse{
		Configs: responses,
		Total:   int(total),
	}

	response.SuccessWithMessage(c, "Crypto wallet configs retrieved successfully", responseData)
}

// UpdateCryptoWalletConfig updates a crypto wallet config
// @Summary [Admin] Update crypto wallet config
// @Description Update cryptocurrency wallet configuration
// @Tags Admin-Payment-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Crypto wallet config ID"
// @Param config body dto.UpdateCryptoWalletConfigRequest true "Crypto wallet config update request"
// @Success 200 {object} response.StandardResponse{data=dto.CryptoWalletConfigResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/payment/crypto-wallets/{id} [put]
func (h *CryptoWalletConfigHandler) UpdateCryptoWalletConfig(c *gin.Context) {
	configID, err := h.parseIDParam(c, "id")
	if err != nil {
		response.BadRequest(c, "Invalid config ID", err.Error())
		return
	}

	var req dto.UpdateCryptoWalletConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body", err.Error())
		return
	}

	config, err := h.service.UpdateCryptoWalletConfig(c.Request.Context(), configID, &req)
	if err != nil {
		response.InternalServerError(c, "Failed to update crypto wallet config", err.Error())
		return
	}

	// Convert to response DTO
	respData := dto.ToCryptoWalletConfigResponse(config)
	response.SuccessWithMessage(c, "Crypto wallet config updated successfully", respData)
}

// DeleteCryptoWalletConfig deletes a crypto wallet config
// @Summary [Admin] Delete crypto wallet config
// @Description Soft delete cryptocurrency wallet configuration
// @Tags Admin-Payment-Management
// @Produce json
// @Security BearerAuth
// @Param id path int true "Crypto wallet config ID"
// @Success 200 {object} response.StandardResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/payment/crypto-wallets/{id} [delete]
func (h *CryptoWalletConfigHandler) DeleteCryptoWalletConfig(c *gin.Context) {
	configID, err := h.parseIDParam(c, "id")
	if err != nil {
		response.BadRequest(c, "Invalid config ID", err.Error())
		return
	}

	err = h.service.DeleteCryptoWalletConfig(c.Request.Context(), configID)
	if err != nil {
		response.InternalServerError(c, "Failed to delete crypto wallet config", err.Error())
		return
	}

	response.SuccessWithMessage(c, "Crypto wallet config deleted successfully", nil)
}

// ToggleCryptoWalletConfig toggles the enabled status of a crypto wallet config
// @Summary [Admin] Toggle crypto wallet config
// @Description Toggle enabled status of cryptocurrency wallet configuration
// @Tags Admin-Payment-Management
// @Produce json
// @Security BearerAuth
// @Param id path int true "Crypto wallet config ID"
// @Success 200 {object} response.StandardResponse{data=dto.CryptoWalletConfigResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/payment/crypto-wallets/{id}/toggle [post]
func (h *CryptoWalletConfigHandler) ToggleCryptoWalletConfig(c *gin.Context) {
	configID, err := h.parseIDParam(c, "id")
	if err != nil {
		response.BadRequest(c, "Invalid config ID", err.Error())
		return
	}

	config, err := h.service.ToggleCryptoWalletConfig(c.Request.Context(), configID)
	if err != nil {
		response.InternalServerError(c, "Failed to toggle crypto wallet config", err.Error())
		return
	}

	// Convert to response DTO
	respData := dto.ToCryptoWalletConfigResponse(config)
	response.SuccessWithMessage(c, "Crypto wallet config toggled successfully", respData)
}

// ValidateWalletAddress validates a wallet address
// @Summary [Admin] Validate wallet address
// @Description Validate cryptocurrency wallet address format
// @Tags Admin-Payment-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.ValidateCryptoWalletAddressRequest true "Wallet address validation request"
// @Success 200 {object} response.StandardResponse{data=dto.ValidateCryptoWalletAddressResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/payment/crypto-wallets/validate-address [post]
func (h *CryptoWalletConfigHandler) ValidateWalletAddress(c *gin.Context) {
	var req dto.ValidateCryptoWalletAddressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body", err.Error())
		return
	}

	result, err := h.service.ValidateWalletAddress(c.Request.Context(), &req)
	if err != nil {
		response.InternalServerError(c, "Failed to validate wallet address", err.Error())
		return
	}

	response.SuccessWithMessage(c, "Address validation completed", result)
}

// GetPublicCryptoWalletConfigs gets public crypto wallet configs for payment selection
// @Summary [Public] Get available crypto wallet configs
// @Description Get enabled cryptocurrency wallet configurations for public payment selection
// @Tags User-Payment
// @Produce json
// @Param network query string false "Filter by network" example("trc")
// @Param currency query string false "Filter by currency" example("USDT")
// @Param amount query number false "Filter by supported amount" example(100.50)
// @Success 200 {object} response.StandardResponse{data=dto.CryptoWalletConfigListResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /payment/crypto-wallets [get]
func (h *CryptoWalletConfigHandler) GetPublicCryptoWalletConfigs(c *gin.Context) {
	network := c.Query("network")
	currency := c.Query("currency")
	amountStr := c.Query("amount")

	var configs []*entities.CryptoWalletConfig
	var err error

	if amountStr != "" {
		amount, parseErr := strconv.ParseFloat(amountStr, 64)
		if parseErr != nil {
			response.BadRequest(c, "Invalid amount parameter", parseErr.Error())
			return
		}
		configs, err = h.service.GetAvailableConfigsForPayment(c.Request.Context(), network, currency, amount)
	} else {
		configs, err = h.service.GetActiveCryptoWalletConfigs(c.Request.Context(), network, currency)
	}

	if err != nil {
		response.InternalServerError(c, "Failed to get public crypto wallet configs", err.Error())
		return
	}

	// Convert to public response DTOs (hide sensitive data)
	var responses []dto.CryptoWalletConfigResponse
	for _, config := range configs {
		responses = append(responses, *dto.ToCryptoWalletConfigPublicResponse(config))
	}

	responseData := dto.CryptoWalletConfigListResponse{
		Configs: responses,
		Total:   len(responses),
	}

	response.SuccessWithMessage(c, "Public crypto wallet configs retrieved successfully", responseData)
}