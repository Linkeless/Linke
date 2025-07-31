package invitecode

import (
	invitecodeshared "linke/internal/handler/user/invite_code/shared"
	"linke/internal/logger"
	"linke/internal/response"
	"linke/internal/service"

	"github.com/gin-gonic/gin"
)

// InviteCodeManagementHandler handles invite code CRUD operations
type InviteCodeManagementHandler struct {
	*invitecodeshared.BaseInviteCodeHandler
	validator *invitecodeshared.InviteCodeValidator
}

// NewInviteCodeManagementHandler creates a new invite code management handler
func NewInviteCodeManagementHandler(inviteCodeService *service.InviteCodeService, inviteCodeUsageService *service.InviteCodeUsageService) *InviteCodeManagementHandler {
	return &InviteCodeManagementHandler{
		BaseInviteCodeHandler: invitecodeshared.NewBaseInviteCodeHandler(inviteCodeService, inviteCodeUsageService),
		validator:             invitecodeshared.NewInviteCodeValidator(),
	}
}

// CreateInviteCode godoc
// @Summary [User] Create invite code
// @Description Create a new invite code
// @Tags User-Invitation
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param invite_code body service.CreateInviteCodeRequest true "Invite code data"
// @Success 201 {object} response.StandardResponse{data=model.InviteCodeResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /invite-codes [post]
func (h *InviteCodeManagementHandler) CreateInviteCode(c *gin.Context) {
	// Get current user from context
	user, valid := h.validator.GetUserFromContext(c)
	if !valid {
		return
	}

	var req service.CreateInviteCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	inviteCode, err := h.InviteCodeService.CreateInviteCode(c.Request.Context(), user.ID, &req)
	if err != nil {
		logger.Error("Failed to create invite code",
			logger.Uint("user_id", user.ID),
			logger.Error2("error", err),
		)
		response.BadRequest(c, err.Error())
		return
	}

	response.Created(c, inviteCode.ToResponse())
}

// GetInviteCode godoc
// @Summary [User] Get invite code by ID
// @Description Get invite code details by ID (only creator or admin can access)
// @Tags User-Invitation
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Invite code ID"
// @Success 200 {object} response.StandardResponse{data=model.InviteCodeResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Router /invite-codes/{id} [get]
func (h *InviteCodeManagementHandler) GetInviteCode(c *gin.Context) {
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

	inviteCode, err := h.InviteCodeService.GetInviteCodeByIDWithRelations(c.Request.Context(), id)
	if err != nil {
		logger.Error("Failed to get invite code",
			logger.Uint("invite_code_id", id),
			logger.Error2("error", err),
		)
		response.NotFound(c, "Invite code not found")
		return
	}

	// Check if user is the creator or admin
	if inviteCode.CreatedByID != user.ID && !user.IsAdmin() {
		response.Forbidden(c, "You can only access your own invite codes")
		return
	}

	response.Success(c, inviteCode.ToResponse())
}

// DeleteInviteCode godoc
// @Summary [User] Delete invite code
// @Description Delete an invite code (only creator or admin can delete)
// @Tags User-Invitation
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Invite code ID"
// @Success 200 {object} response.MessageOnlyResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /invite-codes/{id} [delete]
func (h *InviteCodeManagementHandler) DeleteInviteCode(c *gin.Context) {
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

	// Check if user owns the invite code
	inviteCode, err := h.InviteCodeService.GetInviteCodeByID(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, "Invite code not found")
		return
	}

	if inviteCode.CreatedByID != user.ID && !user.IsAdmin() {
		response.Forbidden(c, "You can only delete your own invite codes")
		return
	}

	if err := h.InviteCodeService.DeleteInviteCode(c.Request.Context(), id); err != nil {
		logger.Error("Failed to delete invite code",
			logger.Uint("invite_code_id", id),
			logger.Error2("error", err),
		)
		response.InternalServerError(c, "Failed to delete invite code")
		return
	}

	response.SuccessWithMessage(c, "Invite code deleted successfully", nil)
}