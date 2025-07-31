package profile

import (
	profileshared "linke/internal/handler/user/profile/shared"
	"linke/internal/logger"
	"linke/internal/response"
	"linke/internal/service"

	"github.com/gin-gonic/gin"
)

// ProfileManagementHandler handles profile CRUD operations
type ProfileManagementHandler struct {
	*profileshared.BaseProfileHandler
	validator *profileshared.ProfileValidator
}

// NewProfileManagementHandler creates a new profile management handler
func NewProfileManagementHandler(userService *service.UserService) *ProfileManagementHandler {
	return &ProfileManagementHandler{
		BaseProfileHandler: profileshared.NewBaseProfileHandler(userService),
		validator:          profileshared.NewProfileValidator(),
	}
}

// GetProfile godoc
// @Summary Get user profile
// @Description Get current user's profile information
// @Tags User-Profile
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.StandardResponse{data=model.UserResponse}
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /user/profile [get]
func (h *ProfileManagementHandler) GetProfile(c *gin.Context) {
	// Get current user from context (set by auth middleware)
	user, valid := h.validator.GetUserFromContext(c)
	if !valid {
		return
	}

	// Fetch fresh user data from database (only active users)
	currentUser, err := h.UserService.GetActiveUserByID(c.Request.Context(), user.ID)
	if err != nil {
		logger.Error("Failed to get active user profile",
			logger.Uint("user_id", user.ID),
			logger.Error2("error", err),
		)
		response.Unauthorized(c, "User account is not active")
		return
	}

	response.Success(c, currentUser.ToResponse())
}

// UpdateProfile godoc
// @Summary [User] Update own profile
// @Description Update current user's profile information (limited fields)
// @Tags User-Profile
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param user body profileshared.UserProfileUpdateRequest true "User profile data"
// @Success 200 {object} response.StandardResponse{data=model.UserResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /user/profile [put]
func (h *ProfileManagementHandler) UpdateProfile(c *gin.Context) {
	// Get current user from context
	currentUser, valid := h.validator.GetUserFromContext(c)
	if !valid {
		return
	}

	// Bind and validate update request
	updateData, valid := h.validator.BindAndValidateProfileUpdateRequest(c)
	if !valid {
		return
	}

	// Get current user data from database (only active users)
	user, err := h.UserService.GetActiveUserByID(c.Request.Context(), currentUser.ID)
	if err != nil {
		logger.Error("Failed to get active user for profile update",
			logger.Uint("user_id", currentUser.ID),
			logger.Error2("error", err),
		)
		response.Unauthorized(c, "User account is not active")
		return
	}

	// Update only allowed fields
	if updateData.Username != "" {
		user.Username = updateData.Username
	}
	if updateData.Name != "" {
		user.Name = updateData.Name
	}
	if updateData.Avatar != "" {
		user.Avatar = updateData.Avatar
	}

	// Save the updated user
	if err := h.UserService.UpdateUser(c.Request.Context(), user); err != nil {
		logger.Error("Failed to update user profile",
			logger.Uint("user_id", currentUser.ID),
			logger.Error2("error", err),
		)
		response.InternalServerError(c, "Failed to update profile")
		return
	}

	response.Success(c, user.ToResponse())
}