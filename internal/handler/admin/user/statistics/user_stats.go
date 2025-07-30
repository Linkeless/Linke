package statistics

import (
	"linke/internal/handler/admin/user/shared"
	"linke/internal/logger"
	"linke/internal/response"
	"linke/internal/service"

	"github.com/gin-gonic/gin"
)

// UserStatsHandler handles user statistics operations
type UserStatsHandler struct {
	*shared.BaseHandler
}

// NewUserStatsHandler creates a new user stats handler
func NewUserStatsHandler(userService *service.UserService, authService *service.AuthService) *UserStatsHandler {
	return &UserStatsHandler{
		BaseHandler: shared.NewBaseHandler(userService, authService),
	}
}

// GetUserStats godoc
// @Summary [Admin] Get user statistics
// @Description Get overall user statistics (admin only)
// @Tags Admin-User-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.StandardResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/users/stats [get]
func (h *UserStatsHandler) GetUserStats(c *gin.Context) {
	stats, err := h.UserService.GetUserStats(c.Request.Context())
	if err != nil {
		logger.Error("Admin failed to get user stats", logger.Error2("error", err))
		response.InternalServerError(c, "Failed to get user statistics")
		return
	}

	response.Success(c, stats)
}