package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"linke/internal/domains/payment/entities"
	"linke/internal/domains/payment/usecases/interfaces"
	userEntities "linke/internal/domains/user/entities"
	"linke/internal/shared/logger"
	"linke/internal/shared/response"
)

// PaymentMethodHandler handles HTTP requests for payment method management
type PaymentMethodHandler struct {
	paymentMethodService interfaces.PaymentMethodService
	logger               logger.Logger
}

// NewPaymentMethodHandler creates a new payment method handler
func NewPaymentMethodHandler(
	paymentMethodService interfaces.PaymentMethodService,
	logger logger.Logger,
) *PaymentMethodHandler {
	return &PaymentMethodHandler{
		paymentMethodService: paymentMethodService,
		logger:               logger,
	}
}

// CreatePaymentMethod godoc
// @Summary Create a new payment method
// @Description Add a new payment method for the authenticated user
// @Tags payment-methods
// @Accept json
// @Produce json
// @Param request body entities.CreatePaymentMethodRequest true "Payment method creation request"
// @Success 201 {object} response.APIResponse{data=entities.PaymentMethodResponse}
// @Failure 400 {object} response.APIResponse
// @Failure 401 {object} response.APIResponse
// @Failure 409 {object} response.APIResponse
// @Failure 422 {object} response.APIResponse
// @Failure 500 {object} response.APIResponse
// @Security BearerAuth
// @Router /payment-methods [post]
func (h *PaymentMethodHandler) CreatePaymentMethod(c *gin.Context) {
	userID := h.getUserIDFromContext(c)
	if userID == 0 {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	var req entities.CreatePaymentMethodRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		response.BadRequest(c, "Invalid request body", err.Error())
		return
	}

	// Rate limiting - check if user is creating too many payment methods too quickly
	if err := h.checkRateLimit(c, userID, "create_payment_method"); err != nil {
		response.Error(c, http.StatusTooManyRequests, 4029, "Rate limit exceeded")
		return
	}

	result, err := h.paymentMethodService.CreatePaymentMethod(c.Request.Context(), userID, &req)
	if err != nil {
		h.logger.Error("Failed to create payment method", zap.Error(err), zap.Uint("user_id", userID))

		// Handle specific error cases
		if isLimitReachedError(err) {
			response.Conflict(c, "Payment method limit reached")
			return
		}
		if isValidationError(err) {
			response.Error(c, http.StatusUnprocessableEntity, 4022, "Payment method validation failed")
			return
		}
		if isDuplicateError(err) {
			response.Conflict(c, "Payment method already exists")
			return
		}

		response.InternalServerError(c, "Failed to create payment method")
		return
	}

	h.logger.Info("Payment method created successfully", zap.Uint("user_id", userID), zap.Uint("payment_method_id", result.ID))
	response.CreatedWithMessage(c, "Payment method created successfully", result)
}

// GetPaymentMethod godoc
// @Summary Get a payment method by ID
// @Description Retrieve a specific payment method for the authenticated user
// @Tags payment-methods
// @Produce json
// @Param id path int true "Payment method ID"
// @Success 200 {object} response.APIResponse{data=entities.PaymentMethodResponse}
// @Failure 400 {object} response.APIResponse
// @Failure 401 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Failure 500 {object} response.APIResponse
// @Security BearerAuth
// @Router /payment-methods/{id} [get]
func (h *PaymentMethodHandler) GetPaymentMethod(c *gin.Context) {
	userID := h.getUserIDFromContext(c)
	if userID == 0 {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	paymentMethodID, err := h.getPaymentMethodIDFromPath(c)
	if err != nil {
		response.BadRequest(c, "Invalid payment method ID", err.Error())
		return
	}

	result, err := h.paymentMethodService.GetPaymentMethod(c.Request.Context(), userID, paymentMethodID)
	if err != nil {
		h.logger.Error("Failed to get payment method", zap.Error(err), zap.Uint("user_id", userID), zap.Uint("payment_method_id", paymentMethodID))

		if isNotFoundError(err) {
			response.NotFound(c, "Payment method not found")
			return
		}

		response.InternalServerError(c, "Failed to get payment method")
		return
	}

	response.SuccessWithMessage(c, "Payment method retrieved successfully", result)
}

// ListPaymentMethods godoc
// @Summary List all payment methods
// @Description Retrieve all payment methods for the authenticated user
// @Tags payment-methods
// @Produce json
// @Param gateway query string false "Filter by payment gateway"
// @Param active_only query bool false "Show only active payment methods"
// @Success 200 {object} response.APIResponse{data=entities.PaymentMethodListResponse}
// @Failure 401 {object} response.APIResponse
// @Failure 500 {object} response.APIResponse
// @Security BearerAuth
// @Router /payment-methods [get]
func (h *PaymentMethodHandler) ListPaymentMethods(c *gin.Context) {
	userID := h.getUserIDFromContext(c)
	if userID == 0 {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	gateway := c.Query("gateway")
	activeOnly := c.Query("active_only") == "true"

	var result *entities.PaymentMethodListResponse
	var err error

	switch {
	case gateway != "" && activeOnly:
		// Get active methods for specific gateway
		result, err = h.paymentMethodService.ListPaymentMethodsByGateway(c.Request.Context(), userID, gateway)
		if err == nil && result != nil {
			// Filter for active only
			activeResults := make([]entities.PaymentMethodResponse, 0)
			for _, pm := range result.PaymentMethods {
				if pm.CanBeUsed {
					activeResults = append(activeResults, pm)
				}
			}
			result.PaymentMethods = activeResults
			result.Total = len(activeResults)
		}
	case gateway != "":
		// Get all methods for specific gateway
		result, err = h.paymentMethodService.ListPaymentMethodsByGateway(c.Request.Context(), userID, gateway)
	case activeOnly:
		// Get all active methods
		result, err = h.paymentMethodService.ListActivePaymentMethods(c.Request.Context(), userID)
	default:
		// Get all methods
		result, err = h.paymentMethodService.ListPaymentMethods(c.Request.Context(), userID)
	}

	if err != nil {
		h.logger.Error("Failed to list payment methods", zap.Error(err), zap.Uint("user_id", userID))
		response.InternalServerError(c, "Failed to list payment methods")
		return
	}

	response.SuccessWithMessage(c, "Payment methods retrieved successfully", result)
}

// UpdatePaymentMethod godoc
// @Summary Update a payment method
// @Description Update details of an existing payment method
// @Tags payment-methods
// @Accept json
// @Produce json
// @Param id path int true "Payment method ID"
// @Param request body entities.UpdatePaymentMethodRequest true "Payment method update request"
// @Success 200 {object} response.APIResponse{data=entities.PaymentMethodResponse}
// @Failure 400 {object} response.APIResponse
// @Failure 401 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Failure 500 {object} response.APIResponse
// @Security BearerAuth
// @Router /payment-methods/{id} [put]
func (h *PaymentMethodHandler) UpdatePaymentMethod(c *gin.Context) {
	userID := h.getUserIDFromContext(c)
	if userID == 0 {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	paymentMethodID, err := h.getPaymentMethodIDFromPath(c)
	if err != nil {
		response.BadRequest(c, "Invalid payment method ID", err.Error())
		return
	}

	var req entities.UpdatePaymentMethodRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		response.BadRequest(c, "Invalid request body", err.Error())
		return
	}

	result, err := h.paymentMethodService.UpdatePaymentMethod(c.Request.Context(), userID, paymentMethodID, &req)
	if err != nil {
		h.logger.Error("Failed to update payment method", zap.Error(err), zap.Uint("user_id", userID), zap.Uint("payment_method_id", paymentMethodID))

		if isNotFoundError(err) {
			response.NotFound(c, "Payment method not found")
			return
		}

		response.InternalServerError(c, "Failed to update payment method")
		return
	}

	h.logger.Info("Payment method updated successfully", zap.Uint("user_id", userID), zap.Uint("payment_method_id", paymentMethodID))
	response.SuccessWithMessage(c, "Payment method updated successfully", result)
}

// SetDefaultPaymentMethod godoc
// @Summary Set a payment method as default
// @Description Set a payment method as the default for the user
// @Tags payment-methods
// @Produce json
// @Param id path int true "Payment method ID"
// @Success 200 {object} response.APIResponse{data=entities.PaymentMethodResponse}
// @Failure 400 {object} response.APIResponse
// @Failure 401 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Failure 422 {object} response.APIResponse
// @Failure 500 {object} response.APIResponse
// @Security BearerAuth
// @Router /payment-methods/{id}/default [put]
func (h *PaymentMethodHandler) SetDefaultPaymentMethod(c *gin.Context) {
	userID := h.getUserIDFromContext(c)
	if userID == 0 {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	paymentMethodID, err := h.getPaymentMethodIDFromPath(c)
	if err != nil {
		response.BadRequest(c, "Invalid payment method ID", err.Error())
		return
	}

	result, err := h.paymentMethodService.SetDefaultPaymentMethod(c.Request.Context(), userID, paymentMethodID)
	if err != nil {
		h.logger.Error("Failed to set default payment method", zap.Error(err), zap.Uint("user_id", userID), zap.Uint("payment_method_id", paymentMethodID))

		if isNotFoundError(err) {
			response.NotFound(c, "Payment method not found")
			return
		}
		if isValidationError(err) {
			response.Error(c, http.StatusUnprocessableEntity, 4022, "Cannot set payment method as default")
			return
		}

		response.InternalServerError(c, "Failed to set default payment method")
		return
	}

	h.logger.Info("Payment method set as default", zap.Uint("user_id", userID), zap.Uint("payment_method_id", paymentMethodID))
	response.SuccessWithMessage(c, "Payment method set as default successfully", result)
}

// DeletePaymentMethod godoc
// @Summary Delete a payment method
// @Description Soft delete a payment method for the authenticated user
// @Tags payment-methods
// @Produce json
// @Param id path int true "Payment method ID"
// @Success 200 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Failure 401 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Failure 500 {object} response.APIResponse
// @Security BearerAuth
// @Router /payment-methods/{id} [delete]
func (h *PaymentMethodHandler) DeletePaymentMethod(c *gin.Context) {
	userID := h.getUserIDFromContext(c)
	if userID == 0 {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	paymentMethodID, err := h.getPaymentMethodIDFromPath(c)
	if err != nil {
		response.BadRequest(c, "Invalid payment method ID", err.Error())
		return
	}

	if err := h.paymentMethodService.DeletePaymentMethod(c.Request.Context(), userID, paymentMethodID); err != nil {
		h.logger.Error("Failed to delete payment method", zap.Error(err), zap.Uint("user_id", userID), zap.Uint("payment_method_id", paymentMethodID))

		if isNotFoundError(err) {
			response.NotFound(c, "Payment method not found")
			return
		}

		response.InternalServerError(c, "Failed to delete payment method")
		return
	}

	h.logger.Info("Payment method deleted successfully", zap.Uint("user_id", userID), zap.Uint("payment_method_id", paymentMethodID))
	response.SuccessWithMessage(c, "Payment method deleted successfully", nil)
}

// ValidatePaymentMethod godoc
// @Summary Validate a payment method
// @Description Validate a payment method with the payment gateway
// @Tags payment-methods
// @Produce json
// @Param id path int true "Payment method ID"
// @Success 200 {object} response.APIResponse{data=entities.PaymentMethodResponse}
// @Failure 400 {object} response.APIResponse
// @Failure 401 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Failure 422 {object} response.APIResponse
// @Failure 500 {object} response.APIResponse
// @Security BearerAuth
// @Router /payment-methods/{id}/validate [post]
func (h *PaymentMethodHandler) ValidatePaymentMethod(c *gin.Context) {
	userID := h.getUserIDFromContext(c)
	if userID == 0 {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	paymentMethodID, err := h.getPaymentMethodIDFromPath(c)
	if err != nil {
		response.BadRequest(c, "Invalid payment method ID", err.Error())
		return
	}

	// Rate limiting for validation requests
	if err := h.checkRateLimit(c, userID, "validate_payment_method"); err != nil {
		response.Error(c, http.StatusTooManyRequests, 4029, "Rate limit exceeded")
		return
	}

	result, err := h.paymentMethodService.ValidatePaymentMethod(c.Request.Context(), userID, paymentMethodID)
	if err != nil {
		h.logger.Error("Failed to validate payment method", zap.Error(err), zap.Uint("user_id", userID), zap.Uint("payment_method_id", paymentMethodID))

		if isNotFoundError(err) {
			response.NotFound(c, "Payment method not found")
			return
		}
		if isValidationError(err) {
			response.Error(c, http.StatusUnprocessableEntity, 4022, "Payment method validation failed")
			return
		}

		response.InternalServerError(c, "Failed to validate payment method")
		return
	}

	h.logger.Info("Payment method validated successfully", zap.Uint("user_id", userID), zap.Uint("payment_method_id", paymentMethodID))
	response.SuccessWithMessage(c, "Payment method validated successfully", result)
}

// GetDefaultPaymentMethod godoc
// @Summary Get default payment method
// @Description Retrieve the default payment method for the authenticated user
// @Tags payment-methods
// @Produce json
// @Param gateway query string false "Filter by payment gateway"
// @Success 200 {object} response.APIResponse{data=entities.PaymentMethodResponse}
// @Failure 401 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Failure 500 {object} response.APIResponse
// @Security BearerAuth
// @Router /payment-methods/default [get]
func (h *PaymentMethodHandler) GetDefaultPaymentMethod(c *gin.Context) {
	userID := h.getUserIDFromContext(c)
	if userID == 0 {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	gateway := c.Query("gateway")

	var result *entities.PaymentMethodResponse
	var err error

	if gateway != "" {
		result, err = h.paymentMethodService.GetDefaultPaymentMethodByGateway(c.Request.Context(), userID, gateway)
	} else {
		result, err = h.paymentMethodService.GetDefaultPaymentMethod(c.Request.Context(), userID)
	}

	if err != nil {
		h.logger.Error("Failed to get default payment method", zap.Error(err), zap.Uint("user_id", userID), zap.String("gateway", gateway))

		if isNotFoundError(err) {
			response.NotFound(c, "No default payment method found")
			return
		}

		response.InternalServerError(c, "Failed to get default payment method")
		return
	}

	response.SuccessWithMessage(c, "Default payment method retrieved successfully", result)
}

// GetPaymentMethodUsageStats godoc
// @Summary Get payment method usage statistics
// @Description Retrieve usage statistics for a specific payment method
// @Tags payment-methods
// @Produce json
// @Param id path int true "Payment method ID"
// @Success 200 {object} response.APIResponse{data=interfaces.PaymentMethodUsageStats}
// @Failure 400 {object} response.APIResponse
// @Failure 401 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Failure 500 {object} response.APIResponse
// @Security BearerAuth
// @Router /payment-methods/{id}/stats [get]
func (h *PaymentMethodHandler) GetPaymentMethodUsageStats(c *gin.Context) {
	userID := h.getUserIDFromContext(c)
	if userID == 0 {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	paymentMethodID, err := h.getPaymentMethodIDFromPath(c)
	if err != nil {
		response.BadRequest(c, "Invalid payment method ID", err.Error())
		return
	}

	result, err := h.paymentMethodService.GetPaymentMethodUsageStats(c.Request.Context(), userID, paymentMethodID)
	if err != nil {
		h.logger.Error("Failed to get payment method usage stats", zap.Error(err), zap.Uint("user_id", userID), zap.Uint("payment_method_id", paymentMethodID))

		if isNotFoundError(err) {
			response.NotFound(c, "Payment method not found")
			return
		}

		response.InternalServerError(c, "Failed to get payment method usage stats")
		return
	}

	response.SuccessWithMessage(c, "Payment method usage stats retrieved successfully", result)
}

// Helper methods

// getUserIDFromContext extracts user ID from the gin context
// This assumes the authentication middleware sets the user object in the context
func (h *PaymentMethodHandler) getUserIDFromContext(c *gin.Context) uint {
	// Try to get user from context using the standard middleware key
	if userValue, exists := c.Get("auth_user"); exists {
		if user, ok := userValue.(*userEntities.User); ok {
			return user.ID
		}
	}
	
	// Fallback: try the old key for backward compatibility
	if userID, exists := c.Get("user_id"); exists {
		if id, ok := userID.(uint); ok {
			return id
		}
	}
	return 0
}

// getPaymentMethodIDFromPath extracts payment method ID from the URL path
func (h *PaymentMethodHandler) getPaymentMethodIDFromPath(c *gin.Context) (uint, error) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}

// checkRateLimit implements rate limiting for sensitive operations
// This is a placeholder - implement actual rate limiting logic
func (h *PaymentMethodHandler) checkRateLimit(c *gin.Context, userID uint, operation string) error {
	// TODO: Implement actual rate limiting
	// This could check Redis or another cache for recent operations
	return nil
}

// Error type checking helpers

func isNotFoundError(err error) bool {
	return err != nil && (containsString(err.Error(), "not found") ||
		containsString(err.Error(), "access denied"))
}

func isValidationError(err error) bool {
	return err != nil && (containsString(err.Error(), "validation failed") ||
		containsString(err.Error(), "cannot be used") ||
		containsString(err.Error(), "inactive or expired"))
}

func isLimitReachedError(err error) bool {
	return err != nil && containsString(err.Error(), "maximum number")
}

func isDuplicateError(err error) bool {
	return err != nil && containsString(err.Error(), "already exists")
}

func containsString(str, substr string) bool {
	return len(str) >= len(substr) &&
		(str == substr ||
			(len(str) > len(substr) &&
				(str[:len(substr)] == substr ||
					str[len(str)-len(substr):] == substr ||
					containsSubstring(str, substr))))
}

func containsSubstring(str, substr string) bool {
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
