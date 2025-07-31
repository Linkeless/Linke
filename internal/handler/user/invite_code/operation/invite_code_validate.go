package invitecode

import (
	invitecodeshared "linke/internal/handler/user/invite_code/shared"
	"linke/internal/logger"
	"linke/internal/response"
	"linke/internal/service"

	"github.com/gin-gonic/gin"
)

// InviteCodeValidationHandler handles invite code validation operations
type InviteCodeValidationHandler struct {
	*invitecodeshared.BaseInviteCodeHandler
	validator *invitecodeshared.InviteCodeValidator
}

// NewInviteCodeValidationHandler creates a new invite code validation handler
func NewInviteCodeValidationHandler(inviteCodeService *service.InviteCodeService, inviteCodeUsageService *service.InviteCodeUsageService) *InviteCodeValidationHandler {
	return &InviteCodeValidationHandler{
		BaseInviteCodeHandler: invitecodeshared.NewBaseInviteCodeHandler(inviteCodeService, inviteCodeUsageService),
		validator:             invitecodeshared.NewInviteCodeValidator(),
	}
}

// ValidateInviteCode godoc
// @Summary [Public] Validate invite code
// @Description Validate if an invite code can be used
// @Tags User-Invitation
// @Accept json
// @Produce json
// @Param code path string true "Invite code"
// @Success 200 {object} response.StandardResponse{data=model.InviteCodeResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 404 {object} response.NotFoundResponse
// @Router /invite-codes/validate/{code} [get]
func (h *InviteCodeValidationHandler) ValidateInviteCode(c *gin.Context) {
	// Validate code parameter
	code, valid := h.validator.ValidateCodeParam(c)
	if !valid {
		return
	}

	inviteCode, err := h.InviteCodeService.ValidateInviteCode(c.Request.Context(), code)
	if err != nil {
		logger.Warn("Invite code validation failed",
			logger.String("code", code),
			logger.Error2("error", err),
		)
		response.NotFound(c, err.Error())
		return
	}

	response.Success(c, inviteCode.ToPublicResponse())
}