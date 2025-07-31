package profile

import (
	"linke/internal/middleware"
	"linke/internal/model"
	"linke/internal/response"

	"github.com/gin-gonic/gin"
)

// ProfileValidator provides validation utilities for profile handlers
type ProfileValidator struct{}

// NewProfileValidator creates a new profile validator
func NewProfileValidator() *ProfileValidator {
	return &ProfileValidator{}
}

// GetUserFromContext extracts and validates user from context
func (v *ProfileValidator) GetUserFromContext(c *gin.Context) (*model.User, bool) {
	userValue, exists := c.Get(middleware.AuthContextKey)
	if !exists {
		response.Unauthorized(c, "Authentication required")
		return nil, false
	}

	user, ok := userValue.(*model.User)
	if !ok {
		response.Unauthorized(c, "Invalid user context")
		return nil, false
	}

	return user, true
}

// UserProfileUpdateRequest represents the structure for profile updates
type UserProfileUpdateRequest struct {
	Username string `json:"username"`
	Name     string `json:"name"`
	Avatar   string `json:"avatar"`
}

// BindAndValidateProfileUpdateRequest binds and validates profile update request
func (v *ProfileValidator) BindAndValidateProfileUpdateRequest(c *gin.Context) (*UserProfileUpdateRequest, bool) {
	var updateData UserProfileUpdateRequest
	
	if err := c.ShouldBindJSON(&updateData); err != nil {
		response.BadRequest(c, err.Error())
		return nil, false
	}

	return &updateData, true
}

// ChangePasswordRequest represents the structure for password change
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

// BindAndValidatePasswordChangeRequest binds and validates password change request
func (v *ProfileValidator) BindAndValidatePasswordChangeRequest(c *gin.Context) (*ChangePasswordRequest, bool) {
	var req ChangePasswordRequest
	
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return nil, false
	}

	// Validate password length
	if len(req.NewPassword) < 6 {
		response.BadRequest(c, "New password must be at least 6 characters")
		return nil, false
	}

	return &req, true
}

// ValidateLocalAccountOnly validates that user has a local account (not OAuth)
func (v *ProfileValidator) ValidateLocalAccountOnly(user *model.User, c *gin.Context) bool {
	if user.Provider != model.ProviderLocal {
		response.BadRequest(c, "Password change is only available for local accounts")
		return false
	}
	return true
}