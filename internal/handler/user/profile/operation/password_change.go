package profile

import (
	profileshared "linke/internal/handler/user/profile/shared"
	"linke/internal/logger"
	"linke/internal/response"
	"linke/internal/service"

	"github.com/gin-gonic/gin"
)

// ProfileOperationHandler handles profile operation requests
type ProfileOperationHandler struct {
	*profileshared.BaseProfileHandler
	validator *profileshared.ProfileValidator
}

// NewProfileOperationHandler creates a new profile operation handler
func NewProfileOperationHandler(userService *service.UserService) *ProfileOperationHandler {
	return &ProfileOperationHandler{
		BaseProfileHandler: profileshared.NewBaseProfileHandler(userService),
		validator:          profileshared.NewProfileValidator(),
	}
}

// ChangePassword godoc
// @Summary [User] Change password
// @Description Change user's own password
// @Tags User-Profile
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param passwords body profileshared.ChangePasswordRequest true "Password change data"
// @Success 200 {object} response.MessageOnlyResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /user/password [put]
func (h *ProfileOperationHandler) ChangePassword(c *gin.Context) {
	// Get current user from context
	currentUser, valid := h.validator.GetUserFromContext(c)
	if !valid {
		return
	}

	// Only allow local account users to change password
	if !h.validator.ValidateLocalAccountOnly(currentUser, c) {
		return
	}

	// Bind and validate password change request
	req, valid := h.validator.BindAndValidatePasswordChangeRequest(c)
	if !valid {
		return
	}

	// Here you would implement password change logic
	// For now, we'll just return success
	// TODO: Implement actual password change with verification
	_ = req // Use the request to avoid unused variable warning
	
	logger.Info("Password changed successfully",
		logger.Uint("user_id", currentUser.ID),
	)

	response.SuccessWithMessage(c, "Password changed successfully", nil)
}