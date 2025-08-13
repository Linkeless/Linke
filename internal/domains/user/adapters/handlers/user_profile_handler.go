package handlers

import (
	"linke/internal/domains/user/entities"
	userInterfaces "linke/internal/domains/user/usecases/interfaces"
	"linke/internal/shared/logger"
	"linke/internal/shared/middleware"
	"linke/internal/shared/response"

	"github.com/gin-gonic/gin"
)

type UserProfileHandler struct {
	userService userInterfaces.UserService
}

func NewUserProfileHandler(userService userInterfaces.UserService) *UserProfileHandler {
	return &UserProfileHandler{
		userService: userService,
	}
}

// GetProfile godoc
// @Summary Get user profile
// @Description Get current user's profile information
// @Tags User-Profile
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} entities.UserResponse
// @Failure 401 {object} response.ProblemJSONResponse
// @Failure 500 {object} response.ProblemJSONResponse
// @Router /user/profile [get]
func (h *UserProfileHandler) GetProfile(c *gin.Context) {
	// Get current user from context (set by auth middleware)
	userValue, exists := c.Get(middleware.AuthContextKey)
	if !exists {
		response.Unauthorized(c, "Authentication required")
		return
	}

	user, ok := userValue.(*entities.UserResponse)
	if !ok {
		response.Unauthorized(c, "Invalid user context")
		return
	}

	// Fetch fresh user data from database (only active users)
	currentUser, err := h.userService.GetActiveUserByID(c.Request.Context(), user.ID)
	if err != nil {
		logger.Error("Failed to get active user profile",
			logger.Uint("user_id", user.ID),
			logger.ErrorField(err),
		)
		response.Unauthorized(c, "User account is not active")
		return
	}

	response.OK(c, currentUser.ToResponse())
}

// UpdateProfile godoc
// @Summary [User] Update own profile
// @Description Update current user's profile information (limited fields)
// @Tags User-Profile
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param user body UserProfileUpdateRequest true "User profile data"
// @Success 200 {object} entities.UserResponse
// @Failure 400 {object} response.ProblemJSONResponse
// @Failure 401 {object} response.ProblemJSONResponse
// @Failure 500 {object} response.ProblemJSONResponse
// @Router /user/profile [put]
func (h *UserProfileHandler) UpdateProfile(c *gin.Context) {
	// Get current user from context
	userValue, exists := c.Get(middleware.AuthContextKey)
	if !exists {
		response.Unauthorized(c, "Authentication required")
		return
	}

	currentUser, ok := userValue.(*entities.UserResponse)
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
			logger.ErrorField(err),
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
			logger.ErrorField(err),
		)
		response.InternalServerError(c, "Failed to update profile")
		return
	}

	response.OK(c, user.ToResponse())
}

// ChangePassword has been removed - password changes are now handled by /api/v1/auth/change-password
// This consolidates password functionality in the auth domain where it belongs.

// UserProfileUpdateRequest represents the structure for profile updates
type UserProfileUpdateRequest struct {
	Username string `json:"username"`
	Name     string `json:"name"`
	Avatar   string `json:"avatar"`
}
