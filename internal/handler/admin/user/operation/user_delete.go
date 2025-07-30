package operation

import (
	"linke/internal/handler/admin/user/shared"
	"linke/internal/logger"
	"linke/internal/response"
	"linke/internal/service"

	"github.com/gin-gonic/gin"
)

// UserDeleteHandler handles user deletion and restoration operations
type UserDeleteHandler struct {
	*shared.BaseHandler
}

// NewUserDeleteHandler creates a new user delete handler
func NewUserDeleteHandler(userService *service.UserService, authService *service.AuthService) *UserDeleteHandler {
	return &UserDeleteHandler{
		BaseHandler: shared.NewBaseHandler(userService, authService),
	}
}

// SoftDeleteUser godoc
// @Summary [Admin] Soft delete user
// @Description Soft delete any user (admin only)
// @Tags Admin-User-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Success 200 {object} response.MessageOnlyResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Router /admin/users/{id} [delete]
func (h *UserDeleteHandler) SoftDeleteUser(c *gin.Context) {
	id, err := h.Validator.ValidateUserID(c)
	if err != nil {
		return // Response already handled by validator
	}

	if err := h.UserService.SoftDeleteUser(c.Request.Context(), id); err != nil {
		logger.Error("Admin failed to soft delete user",
			logger.Uint("user_id", id),
			logger.Error2("error", err),
		)
		response.NotFound(c, "User not found")
		return
	}

	logger.Info("User soft deleted successfully",
		logger.Uint("user_id", id),
	)

	response.SuccessWithMessage(c, "User deleted successfully", nil)
}

// RestoreUser godoc
// @Summary [Admin] Restore user
// @Description Restore a soft deleted user (admin only)
// @Tags Admin-User-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Success 200 {object} response.MessageOnlyResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Router /admin/users/{id}/restore [post]
func (h *UserDeleteHandler) RestoreUser(c *gin.Context) {
	id, err := h.Validator.ValidateUserID(c)
	if err != nil {
		return // Response already handled by validator
	}

	if err := h.UserService.RestoreUser(c.Request.Context(), id); err != nil {
		logger.Error("Admin failed to restore user",
			logger.Uint("user_id", id),
			logger.Error2("error", err),
		)
		response.NotFound(c, "User not found")
		return
	}

	logger.Info("User restored successfully",
		logger.Uint("user_id", id),
	)

	response.SuccessWithMessage(c, "User restored successfully", nil)
}

// HardDeleteUser godoc
// @Summary [Admin] Hard delete user
// @Description Permanently delete a user from database (admin only)
// @Tags Admin-User-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Success 200 {object} response.MessageOnlyResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Router /admin/users/{id}/hard-delete [delete]
func (h *UserDeleteHandler) HardDeleteUser(c *gin.Context) {
	id, err := h.Validator.ValidateUserID(c)
	if err != nil {
		return // Response already handled by validator
	}

	if err := h.UserService.HardDeleteUser(c.Request.Context(), id); err != nil {
		logger.Error("Admin failed to hard delete user",
			logger.Uint("user_id", id),
			logger.Error2("error", err),
		)
		response.NotFound(c, "User not found")
		return
	}

	logger.Warn("User permanently deleted",
		logger.Uint("user_id", id),
	)

	response.SuccessWithMessage(c, "User permanently deleted", nil)
}