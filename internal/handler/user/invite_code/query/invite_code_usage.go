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

// InviteCodeUsageHandler handles invite code usage query operations
type InviteCodeUsageHandler struct {
	*invitecodeshared.BaseInviteCodeHandler
	validator *invitecodeshared.InviteCodeValidator
}

// NewInviteCodeUsageHandler creates a new invite code usage handler
func NewInviteCodeUsageHandler(inviteCodeService *service.InviteCodeService, inviteCodeUsageService *service.InviteCodeUsageService) *InviteCodeUsageHandler {
	return &InviteCodeUsageHandler{
		BaseInviteCodeHandler: invitecodeshared.NewBaseInviteCodeHandler(inviteCodeService, inviteCodeUsageService),
		validator:             invitecodeshared.NewInviteCodeValidator(),
	}
}

// GetInviteCodeUsages godoc
// @Summary [User] Get invite code usages
// @Description Get usage records for a specific invite code (only creator or admin can access)
// @Tags User-Invitation
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Invite code ID"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} response.StandardListResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Router /invite-codes/{id}/usages [get]
func (h *InviteCodeUsageHandler) GetInviteCodeUsages(c *gin.Context) {
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
		response.Forbidden(c, "You can only access your own invite codes")
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

	usages, total, err := h.InviteCodeUsageService.GetUsagesByInviteCode(c.Request.Context(), id, limit, offset)
	if err != nil {
		logger.Error("Failed to get invite code usages",
			logger.Uint("invite_code_id", id),
			logger.Error2("error", err),
		)
		response.InternalServerError(c, "Failed to get invite code usages")
		return
	}

	// Load related data
	if err := h.InviteCodeUsageService.LoadRelatedData(c.Request.Context(), usages); err != nil {
		logger.Error("Failed to load related data for usages",
			logger.Uint("invite_code_id", id),
			logger.Error2("error", err),
		)
	}

	// Convert to response
	var responseData []*model.InviteCodeUsageResponse
	for _, usage := range usages {
		responseData = append(responseData, usage.ToResponse())
	}

	response.SuccessList(c, responseData, page, limit, total)
}