package handlers

import (
	"linke/internal/domains/user/dto"
	"linke/internal/domains/user/entities"
	userEntities "linke/internal/domains/user/entities"
	"linke/internal/domains/user/usecases/interfaces"
	"linke/internal/shared/logger"
	"linke/internal/shared/middleware"
	"linke/internal/shared/response"

	"github.com/gin-gonic/gin"
)

// MessageResponse represents a simple message response
type MessageResponse struct {
	Message string `json:"message" example:"Operation completed successfully"`
}

// UserAccountBindingHandler handles user account binding related requests
type UserAccountBindingHandler struct {
	bindingService interfaces.UserAccountBindingService
}

// NewUserAccountBindingHandler creates a new user account binding handler
func NewUserAccountBindingHandler(bindingService interfaces.UserAccountBindingService) *UserAccountBindingHandler {
	return &UserAccountBindingHandler{
		bindingService: bindingService,
	}
}

// CreateBinding godoc
// @Summary Bind a third-party account
// @Description Bind a third-party account (Google, GitHub, Telegram) to the current user
// @Tags user-bindings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param provider path string true "Provider name (google, github, telegram)"
// @Param binding body dto.CreateBindingRequest true "Binding data"
// @Success 201 {object} dto.UserAccountBindingResponse
// @Failure 400 {object} response.ProblemJSONResponse
// @Failure 401 {object} response.ProblemJSONResponse
// @Failure 409 {object} response.ProblemJSONResponse
// @Failure 500 {object} response.ProblemJSONResponse
// @Router /user/bindings/{provider} [post]
func (h *UserAccountBindingHandler) CreateBinding(c *gin.Context) {
	user, exists := c.Get(middleware.AuthContextKey)
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	u, ok := user.(*userEntities.User)
	if !ok {
		response.InternalServerError(c, "Invalid user context")
		return
	}

	provider := c.Param("provider")
	if !dto.ValidateProvider(provider) {
		response.BadRequest(c, "Invalid provider. Supported providers: google, github, telegram")
		return
	}

	var req dto.CreateBindingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request format: "+err.Error())
		return
	}

	// Ensure provider matches the URL parameter
	req.Provider = provider

	binding, err := h.bindingService.CreateBinding(c.Request.Context(), u.ID, &req)
	if err != nil {
		logger.Error("Failed to create account binding",
			logger.Uint("user_id", u.ID),
			logger.String("provider", provider),
			logger.ErrorField(err))

		// Check if it's a conflict error (already bound)
		if err.Error() == "user already has a binding for provider: "+provider ||
			err.Error() == "provider account "+provider+":"+req.ProviderUserID+" is already bound to another user" {
			response.Conflict(c, err.Error())
			return
		}

		response.InternalServerError(c, "Failed to create account binding: "+err.Error())
		return
	}

	logger.Info("Account binding created successfully",
		logger.Uint("user_id", u.ID),
		logger.String("provider", provider),
		logger.Uint("binding_id", binding.ID))

	bindingResponse := dto.ToUserAccountBindingResponse(binding)
	response.Created(c, bindingResponse)
	// Return the response to pool after use
	dto.PutUserAccountBindingResponse(bindingResponse)
}

// GetBindings godoc
// @Summary Get user's account bindings
// @Description Get all third-party account bindings for the current user
// @Tags user-bindings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.BindingListResponse
// @Failure 401 {object} response.ProblemJSONResponse
// @Failure 500 {object} response.ProblemJSONResponse
// @Router /user/bindings [get]
func (h *UserAccountBindingHandler) GetBindings(c *gin.Context) {
	user, exists := c.Get(middleware.AuthContextKey)
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	u, ok := user.(*userEntities.User)
	if !ok {
		response.InternalServerError(c, "Invalid user context")
		return
	}

	bindings, err := h.bindingService.GetUserBindings(c.Request.Context(), u.ID)
	if err != nil {
		logger.Error("Failed to get user bindings",
			logger.Uint("user_id", u.ID),
			logger.ErrorField(err))
		response.InternalServerError(c, "Failed to get account bindings: "+err.Error())
		return
	}

	// Convert to response format
	bindingResponses := dto.ToUserAccountBindingResponseSlice(bindings)

	bindingList := &dto.BindingListResponse{
		Bindings: bindingResponses,
		Total:    int64(len(bindingResponses)),
	}

	response.OK(c, bindingList)
	// Return the responses to pool after use
	dto.PutUserAccountBindingResponseSlice(bindingResponses)
}

// GetBinding godoc
// @Summary Get a specific account binding
// @Description Get a specific third-party account binding for the current user
// @Tags user-bindings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param provider path string true "Provider name (google, github, telegram)"
// @Success 200 {object} dto.UserAccountBindingResponse
// @Failure 400 {object} response.ProblemJSONResponse
// @Failure 401 {object} response.ProblemJSONResponse
// @Failure 404 {object} response.ProblemJSONResponse
// @Failure 500 {object} response.ProblemJSONResponse
// @Router /user/bindings/{provider} [get]
func (h *UserAccountBindingHandler) GetBinding(c *gin.Context) {
	user, exists := c.Get(middleware.AuthContextKey)
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	u, ok := user.(*userEntities.User)
	if !ok {
		response.InternalServerError(c, "Invalid user context")
		return
	}

	provider := c.Param("provider")
	if !dto.ValidateProvider(provider) {
		response.BadRequest(c, "Invalid provider. Supported providers: google, github, telegram")
		return
	}

	bindings, err := h.bindingService.GetUserBindings(c.Request.Context(), u.ID)
	if err != nil {
		logger.Error("Failed to get user bindings",
			logger.Uint("user_id", u.ID),
			logger.String("provider", provider),
			logger.ErrorField(err))
		response.NotFound(c, "Account binding not found")
		return
	}

	var binding *entities.UserAccountBinding
	for _, b := range bindings {
		if b.Provider == provider {
			binding = b
			break
		}
	}

	if binding == nil {
		response.NotFound(c, "Account binding not found")
		return
	}

	bindingResponse := dto.ToUserAccountBindingResponse(binding)
	response.OK(c, bindingResponse)
	// Return the response to pool after use
	dto.PutUserAccountBindingResponse(bindingResponse)
}

// UpdateBinding godoc
// @Summary Update an account binding
// @Description Update a third-party account binding for the current user
// @Tags user-bindings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param provider path string true "Provider name (google, github, telegram)"
// @Param binding body dto.UpdateBindingRequest true "Updated binding data"
// @Success 200 {object} dto.UserAccountBindingResponse
// @Failure 400 {object} response.ProblemJSONResponse
// @Failure 401 {object} response.ProblemJSONResponse
// @Failure 404 {object} response.ProblemJSONResponse
// @Failure 500 {object} response.ProblemJSONResponse
// @Router /user/bindings/{provider} [put]
func (h *UserAccountBindingHandler) UpdateBinding(c *gin.Context) {
	user, exists := c.Get(middleware.AuthContextKey)
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	u, ok := user.(*userEntities.User)
	if !ok {
		response.InternalServerError(c, "Invalid user context")
		return
	}

	provider := c.Param("provider")
	if !dto.ValidateProvider(provider) {
		response.BadRequest(c, "Invalid provider. Supported providers: google, github, telegram")
		return
	}

	var req dto.UpdateBindingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request format: "+err.Error())
		return
	}

	binding, err := h.bindingService.UpdateBinding(c.Request.Context(), u.ID, provider, &req)
	if err != nil {
		logger.Error("Failed to update account binding",
			logger.Uint("user_id", u.ID),
			logger.String("provider", provider),
			logger.ErrorField(err))

		if err.Error() == "failed to get user binding: failed to get user binding by user ID and provider" ||
			err.Error() == "binding not found" {
			response.NotFound(c, "Account binding not found")
			return
		}

		response.InternalServerError(c, "Failed to update account binding: "+err.Error())
		return
	}

	logger.Info("Account binding updated successfully",
		logger.Uint("user_id", u.ID),
		logger.String("provider", provider))

	bindingResponse := dto.ToUserAccountBindingResponse(binding)
	response.OK(c, bindingResponse)
	// Return the response to pool after use
	dto.PutUserAccountBindingResponse(bindingResponse)
}

// DeleteBinding godoc
// @Summary Unbind a third-party account
// @Description Remove a third-party account binding from the current user
// @Tags user-bindings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param provider path string true "Provider name (google, github, telegram)"
// @Success 200 {object} MessageResponse
// @Failure 400 {object} response.ProblemJSONResponse
// @Failure 401 {object} response.ProblemJSONResponse
// @Failure 404 {object} response.ProblemJSONResponse
// @Failure 500 {object} response.ProblemJSONResponse
// @Router /user/bindings/{provider} [delete]
func (h *UserAccountBindingHandler) DeleteBinding(c *gin.Context) {
	user, exists := c.Get(middleware.AuthContextKey)
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	u, ok := user.(*userEntities.User)
	if !ok {
		response.InternalServerError(c, "Invalid user context")
		return
	}

	provider := c.Param("provider")
	if !dto.ValidateProvider(provider) {
		response.BadRequest(c, "Invalid provider. Supported providers: google, github, telegram")
		return
	}

	err := h.bindingService.DeleteBinding(c.Request.Context(), u.ID, provider)
	if err != nil {
		logger.Error("Failed to delete account binding",
			logger.Uint("user_id", u.ID),
			logger.String("provider", provider),
			logger.ErrorField(err))

		if err.Error() == "failed to get user binding: failed to get user binding by user ID and provider" ||
			err.Error() == "binding not found" {
			response.NotFound(c, "Account binding not found")
			return
		}

		response.InternalServerError(c, "Failed to delete account binding: "+err.Error())
		return
	}

	logger.Info("Account binding deleted successfully",
		logger.Uint("user_id", u.ID),
		logger.String("provider", provider))

	response.OK(c, gin.H{"message": "Account binding removed successfully"})
}

// SetPrimaryBinding godoc
// @Summary Set primary account binding
// @Description Set a third-party account binding as the primary one
// @Tags user-bindings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param provider path string true "Provider name (google, github, telegram)"
// @Success 200 {object} MessageResponse
// @Failure 400 {object} response.ProblemJSONResponse
// @Failure 401 {object} response.ProblemJSONResponse
// @Failure 404 {object} response.ProblemJSONResponse
// @Failure 500 {object} response.ProblemJSONResponse
// @Router /user/bindings/{provider}/primary [put]
func (h *UserAccountBindingHandler) SetPrimaryBinding(c *gin.Context) {
	user, exists := c.Get(middleware.AuthContextKey)
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	u, ok := user.(*userEntities.User)
	if !ok {
		response.InternalServerError(c, "Invalid user context")
		return
	}

	provider := c.Param("provider")
	if !dto.ValidateProvider(provider) {
		response.BadRequest(c, "Invalid provider. Supported providers: google, github, telegram")
		return
	}

	err := h.bindingService.SetPrimaryBinding(c.Request.Context(), u.ID, provider)
	if err != nil {
		logger.Error("Failed to set primary binding",
			logger.Uint("user_id", u.ID),
			logger.String("provider", provider),
			logger.ErrorField(err))

		if err.Error() == "failed to get binding: failed to get user binding by user ID and provider" ||
			err.Error() == "binding not found" {
			response.NotFound(c, "Account binding not found")
			return
		}

		response.InternalServerError(c, "Failed to set primary binding: "+err.Error())
		return
	}

	logger.Info("Primary binding set successfully",
		logger.Uint("user_id", u.ID),
		logger.String("provider", provider))

	response.OK(c, gin.H{"message": "Primary binding set successfully"})
}

/*
// GetBindingStats godoc
// @Summary Get binding statistics
// @Description Get statistics about account bindings (admin only)
// @Tags user-bindings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} interfaces.BindingStats
// @Failure 401 {object} response.ProblemJSONResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 500 {object} response.ProblemJSONResponse
// @Router /admin/bindings/stats [get]
func (h *UserAccountBindingHandler) GetBindingStats(c *gin.Context) {
	user, exists := c.Get(middleware.AuthContextKey)
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	u, ok := user.(*userEntities.User)
	if !ok {
		response.InternalServerError(c, "Invalid user context")
		return
	}

	if !u.IsAdmin() {
		response.Forbidden(c, "Admin access required")
		return
	}

	stats, err := h.bindingService.GetBindingStats(c.Request.Context())
	if err != nil {
		logger.Error("Failed to get binding stats",
			logger.ErrorField(err))
		response.InternalServerError(c, "Failed to get binding statistics", err.Error())
		return
	}

	response.Success(c, stats)
}
*/

/*
// CleanupInactiveBindings godoc
// @Summary Cleanup inactive bindings
// @Description Remove inactive bindings older than specified days (admin only)
// @Tags user-bindings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param days query int false "Days of inactivity threshold (default: 90)"
// @Success 200 {object} MessageResponse
// @Failure 400 {object} response.ProblemJSONResponse
// @Failure 401 {object} response.ProblemJSONResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 500 {object} response.ProblemJSONResponse
// @Router /admin/bindings/cleanup [post]
func (h *UserAccountBindingHandler) CleanupInactiveBindings(c *gin.Context) {
	user, exists := c.Get(middleware.AuthContextKey)
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	u, ok := user.(*userEntities.User)
	if !ok {
		response.InternalServerError(c, "Invalid user context")
		return
	}

	if !u.IsAdmin() {
		response.Forbidden(c, "Admin access required")
		return
	}

	// Parse days parameter, default to 90 days
	days := 90
	if daysParam := c.Query("days"); daysParam != "" {
		if parsedDays, err := strconv.Atoi(daysParam); err == nil && parsedDays > 0 {
			days = parsedDays
		} else {
			response.BadRequest(c, "Invalid days parameter")
			return
		}
	}

	err := h.bindingService.CleanupInactiveBindings(c.Request.Context(), days)
	if err != nil {
		logger.Error("Failed to cleanup inactive bindings",
			logger.Int("days", days),
			logger.ErrorField(err))
		response.InternalServerError(c, "Failed to cleanup inactive bindings", err.Error())
		return
	}

	logger.Info("Inactive bindings cleanup completed",
		logger.Int("days", days))

	response.SuccessWithMessage(c, "Inactive bindings cleanup completed", gin.H{
		"days_threshold": days,
	})
}
*/
