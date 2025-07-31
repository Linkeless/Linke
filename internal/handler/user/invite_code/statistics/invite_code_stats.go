package invitecode

import (
	invitecodeshared "linke/internal/handler/user/invite_code/shared"
	"linke/internal/logger"
	"linke/internal/response"
	"linke/internal/service"

	"github.com/gin-gonic/gin"
)

// InviteCodeStatisticsHandler handles invite code statistics operations
type InviteCodeStatisticsHandler struct {
	*invitecodeshared.BaseInviteCodeHandler
	validator *invitecodeshared.InviteCodeValidator
}

// NewInviteCodeStatisticsHandler creates a new invite code statistics handler
func NewInviteCodeStatisticsHandler(inviteCodeService *service.InviteCodeService, inviteCodeUsageService *service.InviteCodeUsageService) *InviteCodeStatisticsHandler {
	return &InviteCodeStatisticsHandler{
		BaseInviteCodeHandler: invitecodeshared.NewBaseInviteCodeHandler(inviteCodeService, inviteCodeUsageService),
		validator:             invitecodeshared.NewInviteCodeValidator(),
	}
}

// GetInviteCodeStats godoc
// @Summary Get invitation code statistics
// @Description Get invitation code statistics (Admin only)
// @Tags Admin-Invitation-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.StandardResponse{data=map[string]interface{}}
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/invite-codes/stats [get]
func (h *InviteCodeStatisticsHandler) GetInviteCodeStats(c *gin.Context) {
	stats, err := h.InviteCodeService.GetInviteCodeStats(c.Request.Context())
	if err != nil {
		logger.Error("Failed to get invite code stats",
			logger.Error2("error", err),
		)
		response.InternalServerError(c, "Failed to get invite code statistics")
		return
	}

	response.Success(c, stats)
}