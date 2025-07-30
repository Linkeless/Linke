package status

import (
	"strings"

	"linke/internal/handler/admin/user/shared"
	"linke/internal/logger"
	"linke/internal/middleware"
	"linke/internal/model"
	"linke/internal/response"
	"linke/internal/service"

	"github.com/gin-gonic/gin"
)

// UserStatusHandler handles user status and role management
type UserStatusHandler struct {
	*shared.BaseHandler
}

// NewUserStatusHandler creates a new user status handler
func NewUserStatusHandler(userService *service.UserService, authService *service.AuthService) *UserStatusHandler {
	return &UserStatusHandler{
		BaseHandler: shared.NewBaseHandler(userService, authService),
	}
}

// UpdateUserRole godoc
// @Summary [Admin] Update user role
// @Description Update user role (admin only)
// @Tags Admin-User-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Param role body map[string]string true "Role data"
// @Success 200 {object} response.StandardResponse{data=model.UserResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Router /admin/users/{id}/role [put]
func (h *UserStatusHandler) UpdateUserRole(c *gin.Context) {
	id, err := h.Validator.ValidateUserID(c)
	if err != nil {
		return // Response already handled by validator
	}

	var roleData struct {
		Role string `json:"role" binding:"required,oneof=user admin"`
	}

	if err := c.ShouldBindJSON(&roleData); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	user, err := h.UserService.UpdateUserRole(c.Request.Context(), id, roleData.Role)
	if err != nil {
		logger.Error("Admin failed to update user role",
			logger.Uint("user_id", id),
			logger.String("role", roleData.Role),
			logger.Error2("error", err),
		)
		response.NotFound(c, "User not found")
		return
	}

	logger.Info("User role updated successfully",
		logger.Uint("user_id", id),
		logger.String("new_role", roleData.Role),
	)

	response.Success(c, user)
}

// UpdateUserStatus godoc
// @Summary [Admin] Update user status
// @Description Update user status (admin only)
// @Tags Admin-User-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Param status body map[string]string true "Status data"
// @Success 200 {object} response.StandardResponse{data=model.UserResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Router /admin/users/{id}/status [put]
func (h *UserStatusHandler) UpdateUserStatus(c *gin.Context) {
	id, err := h.Validator.ValidateUserID(c)
	if err != nil {
		return // Response already handled by validator
	}

	var statusData struct {
		Status string `json:"status" binding:"required,oneof=active inactive banned"`
	}

	if err := c.ShouldBindJSON(&statusData); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	user, err := h.UserService.UpdateUserStatus(c.Request.Context(), id, statusData.Status)
	if err != nil {
		logger.Error("Admin failed to update user status",
			logger.Uint("user_id", id),
			logger.String("status", statusData.Status),
			logger.Error2("error", err),
		)
		response.NotFound(c, "User not found")
		return
	}

	logger.Info("User status updated successfully",
		logger.Uint("user_id", id),
		logger.String("new_status", statusData.Status),
	)

	response.Success(c, user)
}

// ResetUserPassword godoc
// @Summary Reset user password
// @Description Reset a user's password (Admin only). Only works for local accounts.
// @Tags Admin-User-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path uint true "User ID"
// @Param password body ResetPasswordRequest true "New password data"
// @Success 200 {object} response.MessageOnlyResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/users/{id}/reset-password [post]
func (h *UserStatusHandler) ResetUserPassword(c *gin.Context) {
	// Get admin user from context
	adminUser, exists := c.Get(middleware.AuthContextKey)
	if !exists {
		response.Unauthorized(c, "Admin authentication required")
		return
	}

	admin, ok := adminUser.(*model.User)
	if !ok {
		response.InternalServerError(c, "Invalid admin user context")
		return
	}

	// Parse target user ID
	userID, err := h.Validator.ValidateUserID(c)
	if err != nil {
		return // Response already handled by validator
	}

	// Parse request body
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Reset password using auth service
	if err := h.AuthService.AdminResetPassword(c.Request.Context(), admin.ID, userID, req.NewPassword); err != nil {
		logger.Error("Admin password reset failed",
			logger.Uint("admin_id", admin.ID),
			logger.String("admin_email", admin.Email),
			logger.Uint("target_user_id", userID),
			logger.Error2("error", err),
		)
		
		// Check specific error types for appropriate HTTP responses
		if strings.Contains(err.Error(), "user not found") {
			response.NotFound(c, err.Error())
		} else if strings.Contains(err.Error(), "insufficient permissions") {
			response.Forbidden(c, err.Error())
		} else if strings.Contains(err.Error(), "OAuth") || strings.Contains(err.Error(), "authentication") {
			response.BadRequest(c, err.Error())
		} else {
			response.InternalServerError(c, "Failed to reset password")
		}
		return
	}

	response.SuccessWithMessage(c, "User password reset successfully. All existing tokens have been revoked.", nil)
}

// ResetPasswordRequest represents the request structure for admin password reset
type ResetPasswordRequest struct {
	NewPassword string `json:"new_password" binding:"required,min=6,max=255" example:"newSecurePassword123"`
}