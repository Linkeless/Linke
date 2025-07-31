package invitecode

import (
	invitecodeshared "linke/internal/handler/user/invite_code/shared"
	"linke/internal/logger"
	"linke/internal/response"
	"linke/internal/service"

	"github.com/gin-gonic/gin"
)

// InviteCodeStatusHandler handles invite code status operations
type InviteCodeStatusHandler struct {
	*invitecodeshared.BaseInviteCodeHandler
	validator *invitecodeshared.InviteCodeValidator
}

// NewInviteCodeStatusHandler creates a new invite code status handler
func NewInviteCodeStatusHandler(inviteCodeService *service.InviteCodeService, inviteCodeUsageService *service.InviteCodeUsageService) *InviteCodeStatusHandler {
	return &InviteCodeStatusHandler{
		BaseInviteCodeHandler: invitecodeshared.NewBaseInviteCodeHandler(inviteCodeService, inviteCodeUsageService),
		validator:             invitecodeshared.NewInviteCodeValidator(),
	}
}

// UpdateInviteCodeStatusRequest represents the status update request
type UpdateInviteCodeStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=active disabled"`
}

// UpdateInviteCodeStatus godoc
// @Summary [User] Update invite code status
// @Description Update the status of an invite code (only creator or admin can update)
// @Tags User-Invitation
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Invite code ID"
// @Param status body UpdateInviteCodeStatusRequest true "New status"
// @Success 200 {object} response.StandardResponse{data=model.InviteCodeResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /invite-codes/{id}/status [put]
func (h *InviteCodeStatusHandler) UpdateInviteCodeStatus(c *gin.Context) {
	// Get current user from context
	user, valid := h.validator.GetUserFromContext(c)
	if !valid {
		return
	}

	// Validate ID parameter
	id, valid := h.validator.ValidateIDParam(c, "id")
	if !valid {
		return
	}

	var req UpdateInviteCodeStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Check if user owns the invite code
	inviteCode, err := h.InviteCodeService.GetInviteCodeByID(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, "Invite code not found")
		return
	}

	if inviteCode.CreatedByID != user.ID && !user.IsAdmin() {
		response.Forbidden(c, "You can only update your own invite codes")
		return
	}

	updatedCode, err := h.InviteCodeService.UpdateInviteCodeStatus(c.Request.Context(), id, req.Status)
	if err != nil {
		logger.Error("Failed to update invite code status",
			logger.Uint("invite_code_id", id),
			logger.String("status", req.Status),
			logger.Error2("error", err),
		)
		response.InternalServerError(c, "Failed to update invite code status")
		return
	}

	response.Success(c, updatedCode.ToResponse())
}