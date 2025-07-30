package operation

import (
	"linke/internal/handler/admin/user/shared"
	"linke/internal/logger"
	"linke/internal/response"
	"linke/internal/service"

	"github.com/gin-gonic/gin"
)

// UserBatchHandler handles batch user operations
type UserBatchHandler struct {
	*shared.BaseHandler
}

// NewUserBatchHandler creates a new user batch handler
func NewUserBatchHandler(userService *service.UserService, authService *service.AuthService) *UserBatchHandler {
	return &UserBatchHandler{
		BaseHandler: shared.NewBaseHandler(userService, authService),
	}
}

// BatchDeleteUsers godoc
// @Summary [Admin] Batch delete users
// @Description Soft delete multiple users (admin only)
// @Tags Admin-User-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param ids body map[string][]uint true "User IDs"
// @Success 200 {object} response.StandardResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/users/batch/delete [post]
func (h *UserBatchHandler) BatchDeleteUsers(c *gin.Context) {
	ids, err := h.Validator.ValidateBatchIDs(c)
	if err != nil {
		return // Response already handled by validator
	}

	result, err := h.UserService.BatchDeleteUsers(c.Request.Context(), ids)
	if err != nil {
		logger.Error("Admin failed to batch delete users",
			logger.Any("user_ids", ids),
			logger.Error2("error", err),
		)
		response.InternalServerError(c, "Failed to delete users")
		return
	}

	logger.Info("Batch delete completed",
		logger.Int("deleted_count", result.DeletedCount),
		logger.Int("failed_count", len(result.FailedIDs)),
	)

	response.SuccessWithMessage(c, "Users deleted successfully", map[string]interface{}{
		"deleted_count": result.DeletedCount,
		"failed_ids": result.FailedIDs,
	})
}

// BatchRestoreUsers godoc
// @Summary [Admin] Batch restore users
// @Description Restore multiple soft deleted users (admin only)
// @Tags Admin-User-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param ids body map[string][]uint true "User IDs"
// @Success 200 {object} response.StandardResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/users/batch/restore [post]
func (h *UserBatchHandler) BatchRestoreUsers(c *gin.Context) {
	ids, err := h.Validator.ValidateBatchIDs(c)
	if err != nil {
		return // Response already handled by validator
	}

	result, err := h.UserService.BatchRestoreUsers(c.Request.Context(), ids)
	if err != nil {
		logger.Error("Admin failed to batch restore users",
			logger.Any("user_ids", ids),
			logger.Error2("error", err),
		)
		response.InternalServerError(c, "Failed to restore users")
		return
	}

	logger.Info("Batch restore completed",
		logger.Int("restored_count", result.RestoredCount),
		logger.Int("failed_count", len(result.FailedIDs)),
	)

	response.SuccessWithMessage(c, "Users restored successfully", map[string]interface{}{
		"restored_count": result.RestoredCount,
		"failed_ids": result.FailedIDs,
	})
}