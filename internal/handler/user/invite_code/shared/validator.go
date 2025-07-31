package invitecode

import (
	"strconv"

	"linke/internal/logger"
	"linke/internal/middleware"
	"linke/internal/model"
	"linke/internal/response"

	"github.com/gin-gonic/gin"
)

// InviteCodeValidator provides validation utilities for invite code handlers
type InviteCodeValidator struct{}

// NewInviteCodeValidator creates a new invite code validator
func NewInviteCodeValidator() *InviteCodeValidator {
	return &InviteCodeValidator{}
}

// ValidatePaginationParams validates and parses pagination parameters
func (v *InviteCodeValidator) ValidatePaginationParams(c *gin.Context) (limit, offset int, valid bool) {
	// Default values
	limit = 20
	offset = 0

	// Parse limit
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 100 {
			limit = parsedLimit
		}
	}

	// Parse offset
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	return limit, offset, true
}

// GetUserFromContext extracts and validates user from context
func (v *InviteCodeValidator) GetUserFromContext(c *gin.Context) (*model.User, bool) {
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

// ValidateIDParam validates and parses ID parameter from path
func (v *InviteCodeValidator) ValidateIDParam(c *gin.Context, paramName string) (uint, bool) {
	idStr := c.Param(paramName)
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		logger.Error("Invalid ID parameter", logger.String("param", paramName), logger.Error2("error", err))
		response.BadRequest(c, "Invalid ID parameter")
		return 0, false
	}
	return uint(id), true
}

// ValidateCodeParam validates code parameter from path
func (v *InviteCodeValidator) ValidateCodeParam(c *gin.Context) (string, bool) {
	code := c.Param("code")
	if code == "" {
		logger.Error("Empty invite code parameter")
		response.BadRequest(c, "Invite code is required")
		return "", false
	}
	return code, true
}

// CheckInviteCodeOwnership checks if user owns the invite code
func (v *InviteCodeValidator) CheckInviteCodeOwnership(user *model.User, inviteCode *model.InviteCode, c *gin.Context) bool {
	if inviteCode.CreatedByID != user.ID {
		logger.Error("Access denied to invite code", 
			logger.Uint("invite_code_id", inviteCode.ID),
			logger.Uint("user_id", user.ID),
			logger.Uint("owner_id", inviteCode.CreatedByID))
		response.Forbidden(c, "Access denied: you can only access your own invite codes")
		return false
	}
	return true
}