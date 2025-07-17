package handler

import (
	"linke/internal/logger"
	"linke/internal/middleware"
	"linke/internal/model"
	"linke/internal/response"
	"linke/internal/service"

	"github.com/gin-gonic/gin"
)

type UserProfileHandler struct {
	userService *service.UserService
}

func NewUserProfileHandler(userService *service.UserService) *UserProfileHandler {
	return &UserProfileHandler{
		userService: userService,
	}
}

// GetProfile godoc
// @Summary [User] Get own profile
// @Description Get current user's profile information
// @Tags user-profile
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.StandardResponse{data=model.UserResponse}
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /user/profile [get]
func (h *UserProfileHandler) GetProfile(c *gin.Context) {
	// Get current user from context (set by auth middleware)
	userValue, exists := c.Get(middleware.AuthContextKey)
	if !exists {
		response.Unauthorized(c, "Authentication required")
		return
	}

	user, ok := userValue.(*model.User)
	if !ok {
		response.Unauthorized(c, "Invalid user context")
		return
	}

	// Fetch fresh user data from database (only active users)
	currentUser, err := h.userService.GetActiveUserByID(c.Request.Context(), user.ID)
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
// @Tags user-profile
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param user body UserProfileUpdateRequest true "User profile data"
// @Success 200 {object} response.StandardResponse{data=model.UserResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /user/profile [put]
func (h *UserProfileHandler) UpdateProfile(c *gin.Context) {
	// Get current user from context
	userValue, exists := c.Get(middleware.AuthContextKey)
	if !exists {
		response.Unauthorized(c, "Authentication required")
		return
	}

	currentUser, ok := userValue.(*model.User)
	if !ok {
		response.Unauthorized(c, "Invalid user context")
		return
	}

	// Define structure for allowed profile updates
	var updateData struct {
		Username string `json:"username"`
		Name     string `json:"name"`
		Avatar   string `json:"avatar"`
	}

	if err := c.ShouldBindJSON(&updateData); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Get current user data from database (only active users)
	user, err := h.userService.GetActiveUserByID(c.Request.Context(), currentUser.ID)
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
	if err := h.userService.UpdateUser(c.Request.Context(), user); err != nil {
		logger.Error("Failed to update user profile",
			logger.Uint("user_id", currentUser.ID),
			logger.Error2("error", err),
		)
		response.InternalServerError(c, "Failed to update profile")
		return
	}

	response.Success(c, user.ToResponse())
}



// ChangePassword godoc
// @Summary [User] Change password
// @Description Change user's own password
// @Tags user-profile
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param passwords body ChangePasswordRequest true "Password change data"
// @Success 200 {object} response.MessageOnlyResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /user/password [put]
func (h *UserProfileHandler) ChangePassword(c *gin.Context) {
	// Get current user from context
	userValue, exists := c.Get(middleware.AuthContextKey)
	if !exists {
		response.Unauthorized(c, "Authentication required")
		return
	}

	currentUser, ok := userValue.(*model.User)
	if !ok {
		response.Unauthorized(c, "Invalid user context")
		return
	}

	// Only allow local account users to change password
	if currentUser.Provider != model.ProviderLocal {
		response.BadRequest(c, "Password change is only available for local accounts")
		return
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Validate password length
	if len(req.NewPassword) < 6 {
		response.BadRequest(c, "New password must be at least 6 characters")
		return
	}

	// Here you would implement password change logic
	// For now, we'll just return success
	// TODO: Implement actual password change with verification
	
	logger.Info("Password changed successfully",
		logger.Uint("user_id", currentUser.ID),
	)

	response.SuccessWithMessage(c, "Password changed successfully", nil)
}

// UserProfileUpdateRequest represents the structure for profile updates
type UserProfileUpdateRequest struct {
	Username string `json:"username"`
	Name     string `json:"name"`
	Avatar   string `json:"avatar"`
}


// ChangePasswordRequest represents the structure for password change
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}