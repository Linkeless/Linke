package shared

import (
	"errors"
	"net/mail"
	"strconv"
	"strings"

	"linke/internal/model"
	"linke/internal/response"

	"github.com/gin-gonic/gin"
)

// UserValidator provides validation utilities for user operations
type UserValidator struct{}

// NewUserValidator creates a new user validator
func NewUserValidator() *UserValidator {
	return &UserValidator{}
}

// ValidateUserID validates and parses user ID from URL parameter
func (v *UserValidator) ValidateUserID(c *gin.Context) (uint, error) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return 0, errors.New("invalid user ID")
	}
	return uint(id), nil
}

// ValidateUserRole validates user role value
func (v *UserValidator) ValidateUserRole(role string) error {
	if role != model.UserRoleUser && role != model.UserRoleAdmin {
		return errors.New("invalid role value, must be 'user' or 'admin'")
	}
	return nil
}

// ValidateUserStatus validates user status value
func (v *UserValidator) ValidateUserStatus(status string) error {
	if status != model.UserStatusActive && status != model.UserStatusInactive && status != model.UserStatusBanned {
		return errors.New("invalid status value, must be 'active', 'inactive', or 'banned'")
	}
	return nil
}

// ValidateEmail validates email format
func (v *UserValidator) ValidateEmail(email string) error {
	if email == "" {
		return errors.New("email is required")
	}
	
	_, err := mail.ParseAddress(email)
	if err != nil {
		return errors.New("invalid email format")
	}
	
	return nil
}

// ValidatePassword validates password requirements
func (v *UserValidator) ValidatePassword(password string) error {
	if len(password) < 6 {
		return errors.New("password must be at least 6 characters long")
	}
	
	if len(password) > 255 {
		return errors.New("password must be less than 255 characters")
	}
	
	return nil
}

// ValidatePaginationParams validates and returns pagination parameters
func (v *UserValidator) ValidatePaginationParams(c *gin.Context) (int, int, int) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit
	return page, limit, offset
}

// ValidateSearchQuery validates search query parameter
func (v *UserValidator) ValidateSearchQuery(c *gin.Context) (string, error) {
	query := strings.TrimSpace(c.Query("q"))
	if query == "" {
		response.BadRequest(c, "Search query is required")
		return "", errors.New("search query is required")
	}
	return query, nil
}

// ValidateProvider validates OAuth provider parameter
func (v *UserValidator) ValidateProvider(provider string) error {
	validProviders := map[string]bool{
		model.ProviderGoogle:   true,
		model.ProviderGitHub:   true,
		model.ProviderTelegram: true,
		model.ProviderLocal:    true,
	}

	if !validProviders[provider] {
		return errors.New("invalid provider")
	}
	
	return nil
}

// ValidateBatchIDs validates batch operation IDs
func (v *UserValidator) ValidateBatchIDs(c *gin.Context) ([]uint, error) {
	var requestData struct {
		IDs []uint `json:"ids" binding:"required,min=1,max=100"`
	}

	if err := c.ShouldBindJSON(&requestData); err != nil {
		response.BadRequest(c, err.Error())
		return nil, err
	}

	return requestData.IDs, nil
}

// ValidateCreateUserRequest validates user creation request
func (v *UserValidator) ValidateCreateUserRequest(req *model.CreateUserRequest) error {
	if err := v.ValidateEmail(req.Email); err != nil {
		return err
	}

	if req.Password != "" {
		if err := v.ValidatePassword(req.Password); err != nil {
			return err
		}
	}

	if req.Role != "" {
		if err := v.ValidateUserRole(req.Role); err != nil {
			return err
		}
	}

	if req.Status != "" {
		if err := v.ValidateUserStatus(req.Status); err != nil {
			return err
		}
	}

	return nil
}

// ValidatePartialUserUpdate validates partial user update data
func (v *UserValidator) ValidatePartialUserUpdate(updateData map[string]interface{}) error {
	if name, exists := updateData["name"]; exists {
		if _, ok := name.(string); !ok {
			return errors.New("invalid name field type")
		}
	}

	if email, exists := updateData["email"]; exists {
		if emailStr, ok := email.(string); ok {
			if err := v.ValidateEmail(emailStr); err != nil {
				return err
			}
		} else {
			return errors.New("invalid email field type")
		}
	}

	if username, exists := updateData["username"]; exists {
		if _, ok := username.(string); !ok {
			return errors.New("invalid username field type")
		}
	}

	if role, exists := updateData["role"]; exists {
		if roleStr, ok := role.(string); ok {
			if err := v.ValidateUserRole(roleStr); err != nil {
				return err
			}
		} else {
			return errors.New("invalid role field type")
		}
	}

	if status, exists := updateData["status"]; exists {
		if statusStr, ok := status.(string); ok {
			if err := v.ValidateUserStatus(statusStr); err != nil {
				return err
			}
		} else {
			return errors.New("invalid status field type")
		}
	}

	return nil
}