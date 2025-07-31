package invitecode

import (
	"strconv"

	invitecodeshared "linke/internal/handler/user/invite_code/shared"
	"linke/internal/logger"
	"linke/internal/model"
	"linke/internal/response"
	"linke/internal/service"

	"github.com/gin-gonic/gin"
)

// InviteCodeListHandler handles invite code listing operations
type InviteCodeListHandler struct {
	*invitecodeshared.BaseInviteCodeHandler
	validator *invitecodeshared.InviteCodeValidator
}

// NewInviteCodeListHandler creates a new invite code list handler
func NewInviteCodeListHandler(inviteCodeService *service.InviteCodeService, inviteCodeUsageService *service.InviteCodeUsageService) *InviteCodeListHandler {
	return &InviteCodeListHandler{
		BaseInviteCodeHandler: invitecodeshared.NewBaseInviteCodeHandler(inviteCodeService, inviteCodeUsageService),
		validator:             invitecodeshared.NewInviteCodeValidator(),
	}
}

// ListAllInviteCodes godoc
// @Summary [Admin] List all invite codes
// @Description List all invite codes with pagination (Admin only)
// @Tags Admin-Invitation-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} response.StandardListResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/invite-codes [get]
func (h *InviteCodeListHandler) ListAllInviteCodes(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit

	codes, total, err := h.InviteCodeService.ListAllInviteCodes(c.Request.Context(), limit, offset)
	if err != nil {
		logger.Error("Failed to list all invite codes",
			logger.Error2("error", err),
		)
		response.InternalServerError(c, "Failed to list invite codes")
		return
	}

	// Convert to response
	var responseData []*model.InviteCodeResponse
	for _, code := range codes {
		responseData = append(responseData, code.ToResponse())
	}

	response.SuccessList(c, responseData, page, limit, total)
}

// GetMyInviteCodes godoc
// @Summary [User] Get my invite codes
// @Description Get invite codes created by current user
// @Tags User-Invitation
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} response.StandardListResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /invite-codes/my [get]
func (h *InviteCodeListHandler) GetMyInviteCodes(c *gin.Context) {
	// Get current user from context
	user, valid := h.validator.GetUserFromContext(c)
	if !valid {
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit

	codes, total, err := h.InviteCodeService.ListInviteCodesByCreator(c.Request.Context(), user.ID, limit, offset)
	if err != nil {
		logger.Error("Failed to get user invite codes",
			logger.Uint("user_id", user.ID),
			logger.Error2("error", err),
		)
		response.InternalServerError(c, "Failed to get invite codes")
		return
	}

	// Convert to response
	var responseData []*model.InviteCodeResponse
	for _, code := range codes {
		responseData = append(responseData, code.ToResponse())
	}

	response.SuccessList(c, responseData, page, limit, total)
}