package admin

import (
	"strconv"
	"strings"

	"linke/internal/logger"
	"linke/internal/middleware"
	"linke/internal/model"
	"linke/internal/response"
	"linke/internal/service"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type AdminUserHandler struct {
	userService *service.UserService
	authService *service.AuthService
}

func NewAdminUserHandler(userService *service.UserService, authService *service.AuthService) *AdminUserHandler {
	return &AdminUserHandler{
		userService: userService,
		authService: authService,
	}
}

// CreateUser godoc
// @Summary Create new user
// @Description Create a new user account (Admin only)
// @Tags Admin-User-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param user body model.CreateUserRequest true "User creation data"
// @Success 201 {object} response.StandardResponse{data=model.UserResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 409 {object} response.ConflictResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/users [post]
func (h *AdminUserHandler) CreateUser(c *gin.Context) {
	var createReq model.CreateUserRequest
	if err := c.ShouldBindJSON(&createReq); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Create user model from request
	user := &model.User{
		Email:    createReq.Email,
		Username: createReq.Username,
		Name:     createReq.Name,
		Role:     createReq.Role,
		Status:   createReq.Status,
		Provider: model.ProviderLocal,
	}

	// Set password if provided (hash the password before storing)
	if createReq.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(createReq.Password), bcrypt.DefaultCost)
		if err != nil {
			logger.Error("Failed to hash password during user creation",
				logger.String("email", createReq.Email),
				logger.Error2("error", err),
			)
			response.InternalServerError(c, "Failed to process password")
			return
		}
		user.Password = string(hashedPassword)
	}

	// Set default values if not provided
	if user.Role == "" {
		user.Role = model.UserRoleUser
	}
	if user.Status == "" {
		user.Status = model.UserStatusActive
	}

	// Create the user
	if err := h.userService.CreateUser(c.Request.Context(), user); err != nil {
		logger.Error("Admin failed to create user",
			logger.String("email", createReq.Email),
			logger.Error2("error", err),
		)
		
		// Check if it's a duplicate key error
		if strings.Contains(err.Error(), "Duplicate entry") || strings.Contains(err.Error(), "UNIQUE constraint") {
			response.Conflict(c, "User with this email already exists")
			return
		}
		
		response.InternalServerError(c, "Failed to create user")
		return
	}

	logger.Info("Admin created new user",
		logger.Uint("user_id", user.ID),
		logger.String("email", user.Email),
		logger.String("admin_action", "create_user"),
	)

	response.Created(c, user.ToResponse())
}

// GetUser godoc
// @Summary Get user information
// @Description Get user details by user ID (Admin only)
// @Tags Admin-User-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Success 200 {object} response.StandardResponse{data=model.UserResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Router /admin/users/{id} [get]
func (h *AdminUserHandler) GetUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	user, err := h.userService.GetUserByID(c.Request.Context(), uint(id))
	if err != nil {
		logger.Error("Admin failed to get user",
			logger.Uint("user_id", uint(id)),
			logger.Error2("error", err),
		)
		response.NotFound(c, "User not found")
		return
	}

	response.Success(c, user)
}

// ListUsers godoc
// @Summary List all users
// @Description Get paginated list of all users (Admin only)
// @Tags Admin-User-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} response.StandardListResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/users [get]
func (h *AdminUserHandler) ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit

	users, total, err := h.userService.ListUsers(c.Request.Context(), limit, offset)
	if err != nil {
		logger.Error("Admin failed to list users", logger.Error2("error", err))
		response.InternalServerError(c, "Failed to list users")
		return
	}

	response.SuccessList(c, users, page, limit, total)
}

// UpdateUser godoc
// @Summary [Admin] Update any user
// @Description Update any user information (admin only)
// @Tags Admin-User-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Param user body model.UserResponse true "User data"
// @Success 200 {object} response.StandardResponse{data=model.UserResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/users/{id} [put]
func (h *AdminUserHandler) UpdateUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	var user model.User
	if err := c.ShouldBindJSON(&user); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	user.ID = uint(id)
	if err := h.userService.UpdateUser(c.Request.Context(), &user); err != nil {
		logger.Error("Admin failed to update user",
			logger.Uint("user_id", uint(id)),
			logger.Error2("error", err),
		)
		response.InternalServerError(c, "Failed to update user")
		return
	}

	response.Success(c, user)
}

// UpdateUserRole godoc
// @Summary [Admin] Update user role
// @Description Update user role (admin only)
// @Tags Admin-User-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Param role body map[string]string true "Role data"
// @Success 200 {object} response.StandardResponse{data=model.UserResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Router /admin/users/{id}/role [put]
func (h *AdminUserHandler) UpdateUserRole(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	var roleData struct {
		Role string `json:"role" binding:"required,oneof=user admin"`
	}

	if err := c.ShouldBindJSON(&roleData); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	user, err := h.userService.UpdateUserRole(c.Request.Context(), uint(id), roleData.Role)
	if err != nil {
		logger.Error("Admin failed to update user role",
			logger.Uint("user_id", uint(id)),
			logger.String("role", roleData.Role),
			logger.Error2("error", err),
		)
		response.NotFound(c, "User not found")
		return
	}

	response.Success(c, user)
}

// UpdateUserStatus godoc
// @Summary [Admin] Update user status
// @Description Update user status (admin only)
// @Tags Admin-User-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Param status body map[string]string true "Status data"
// @Success 200 {object} response.StandardResponse{data=model.UserResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Router /admin/users/{id}/status [put]
func (h *AdminUserHandler) UpdateUserStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	var statusData struct {
		Status string `json:"status" binding:"required,oneof=active inactive banned"`
	}

	if err := c.ShouldBindJSON(&statusData); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	user, err := h.userService.UpdateUserStatus(c.Request.Context(), uint(id), statusData.Status)
	if err != nil {
		logger.Error("Admin failed to update user status",
			logger.Uint("user_id", uint(id)),
			logger.String("status", statusData.Status),
			logger.Error2("error", err),
		)
		response.NotFound(c, "User not found")
		return
	}

	response.Success(c, user)
}

// SoftDeleteUser godoc
// @Summary [Admin] Soft delete user
// @Description Soft delete any user (admin only)
// @Tags Admin-User-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Success 200 {object} response.MessageOnlyResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Router /admin/users/{id} [delete]
func (h *AdminUserHandler) SoftDeleteUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	if err := h.userService.SoftDeleteUser(c.Request.Context(), uint(id)); err != nil {
		logger.Error("Admin failed to soft delete user",
			logger.Uint("user_id", uint(id)),
			logger.Error2("error", err),
		)
		response.NotFound(c, "User not found")
		return
	}

	response.SuccessWithMessage(c, "User deleted successfully", nil)
}

// RestoreUser godoc
// @Summary [Admin] Restore user
// @Description Restore a soft deleted user (admin only)
// @Tags Admin-User-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Success 200 {object} response.MessageOnlyResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Router /admin/users/{id}/restore [post]
func (h *AdminUserHandler) RestoreUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	if err := h.userService.RestoreUser(c.Request.Context(), uint(id)); err != nil {
		logger.Error("Admin failed to restore user",
			logger.Uint("user_id", uint(id)),
			logger.Error2("error", err),
		)
		response.NotFound(c, "User not found")
		return
	}

	response.SuccessWithMessage(c, "User restored successfully", nil)
}

// HardDeleteUser godoc
// @Summary [Admin] Hard delete user
// @Description Permanently delete a user from database (admin only)
// @Tags Admin-User-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Success 200 {object} response.MessageOnlyResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Router /admin/users/{id}/hard-delete [delete]
func (h *AdminUserHandler) HardDeleteUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	if err := h.userService.HardDeleteUser(c.Request.Context(), uint(id)); err != nil {
		logger.Error("Admin failed to hard delete user",
			logger.Uint("user_id", uint(id)),
			logger.Error2("error", err),
		)
		response.NotFound(c, "User not found")
		return
	}

	response.SuccessWithMessage(c, "User permanently deleted", nil)
}

// ListDeletedUsers godoc
// @Summary [Admin] List deleted users
// @Description Get list of soft deleted users with pagination (admin only)
// @Tags Admin-User-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} response.StandardListResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/users/deleted [get]
func (h *AdminUserHandler) ListDeletedUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit

	users, total, err := h.userService.ListDeletedUsers(c.Request.Context(), limit, offset)
	if err != nil {
		logger.Error("Admin failed to list deleted users", logger.Error2("error", err))
		response.InternalServerError(c, "Failed to list deleted users")
		return
	}

	response.SuccessList(c, users, page, limit, total)
}

// SearchUsers godoc
// @Summary [Admin] Search users
// @Description Search users by name, email, or username with pagination (admin only)
// @Tags Admin-User-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param q query string true "Search query"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} response.SearchResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/users/search [get]
func (h *AdminUserHandler) SearchUsers(c *gin.Context) {
	query := strings.TrimSpace(c.Query("q"))
	if query == "" {
		response.BadRequest(c, "Search query is required")
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

	users, total, err := h.userService.SearchUsers(c.Request.Context(), query, limit, offset)
	if err != nil {
		logger.Error("Admin failed to search users", 
			logger.String("query", query),
			logger.Error2("error", err),
		)
		response.InternalServerError(c, "Failed to search users")
		return
	}

	response.SuccessListWithExtra(c, "Search completed", users, page, limit, total, map[string]interface{}{
		"query": query,
	})
}

// ListUsersByProvider godoc
// @Summary [Admin] List users by provider
// @Description Get users filtered by OAuth provider with pagination (admin only)
// @Tags Admin-User-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param provider query string true "Provider (google, github, telegram)"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} response.ProviderFilterResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/users/provider [get]
func (h *AdminUserHandler) ListUsersByProvider(c *gin.Context) {
	provider := c.Query("provider")
	if provider == "" {
		response.BadRequest(c, "Provider parameter is required")
		return
	}

	validProviders := map[string]bool{
		model.ProviderGoogle:   true,
		model.ProviderGitHub:   true,
		model.ProviderTelegram: true,
	}

	if !validProviders[provider] {
		response.BadRequest(c, "Invalid provider")
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

	users, total, err := h.userService.ListUsersByProvider(c.Request.Context(), provider, limit, offset)
	if err != nil {
		logger.Error("Admin failed to list users by provider",
			logger.String("provider", provider),
			logger.Error2("error", err),
		)
		response.InternalServerError(c, "Failed to list users")
		return
	}

	response.SuccessListWithExtra(c, "Users retrieved successfully", users, page, limit, total, map[string]interface{}{
		"provider": provider,
	})
}

// GetUserStats godoc
// @Summary [Admin] Get user statistics
// @Description Get overall user statistics (admin only)
// @Tags Admin-User-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.StandardResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/users/stats [get]
func (h *AdminUserHandler) GetUserStats(c *gin.Context) {
	stats, err := h.userService.GetUserStats(c.Request.Context())
	if err != nil {
		logger.Error("Admin failed to get user stats", logger.Error2("error", err))
		response.InternalServerError(c, "Failed to get user statistics")
		return
	}

	response.Success(c, stats)
}

// BatchDeleteUsers godoc
// @Summary [Admin] Batch delete users
// @Description Soft delete multiple users (admin only)
// @Tags Admin-User-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param ids body map[string][]uint true "User IDs"
// @Success 200 {object} response.StandardResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/users/batch/delete [post]
func (h *AdminUserHandler) BatchDeleteUsers(c *gin.Context) {
	var requestData struct {
		IDs []uint `json:"ids" binding:"required,min=1,max=100"`
	}

	if err := c.ShouldBindJSON(&requestData); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.userService.BatchDeleteUsers(c.Request.Context(), requestData.IDs)
	if err != nil {
		logger.Error("Admin failed to batch delete users",
			logger.Any("user_ids", requestData.IDs),
			logger.Error2("error", err),
		)
		response.InternalServerError(c, "Failed to delete users")
		return
	}

	response.SuccessWithMessage(c, "Users deleted successfully", map[string]interface{}{
		"deleted_count": result.DeletedCount,
		"failed_ids": result.FailedIDs,
	})
}

// PatchUser godoc
// @Summary [Admin] Partially update user
// @Description Partially update user information using PATCH method (admin only)
// @Tags Admin-User-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Param user body map[string]interface{} true "Partial user data"
// @Success 200 {object} response.StandardResponse{data=model.UserResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/users/{id} [patch]
func (h *AdminUserHandler) PatchUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	// Get current user
	currentUser, err := h.userService.GetUserByID(c.Request.Context(), uint(id))
	if err != nil {
		logger.Error("Admin failed to get user for patch",
			logger.Uint("user_id", uint(id)),
			logger.Error2("error", err),
		)
		response.NotFound(c, "User not found")
		return
	}

	// Parse partial update data
	var updateData map[string]interface{}
	if err := c.ShouldBindJSON(&updateData); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Apply partial updates to the current user
	user := *currentUser
	
	// Update only the fields present in the request
	if name, exists := updateData["name"]; exists {
		if nameStr, ok := name.(string); ok {
			user.Name = nameStr
		} else {
			response.BadRequest(c, "Invalid name field type")
			return
		}
	}
	
	if email, exists := updateData["email"]; exists {
		if emailStr, ok := email.(string); ok {
			user.Email = emailStr
		} else {
			response.BadRequest(c, "Invalid email field type")
			return
		}
	}
	
	if username, exists := updateData["username"]; exists {
		if usernameStr, ok := username.(string); ok {
			user.Username = usernameStr
		} else {
			response.BadRequest(c, "Invalid username field type")
			return
		}
	}
	
	if role, exists := updateData["role"]; exists {
		if roleStr, ok := role.(string); ok {
			if roleStr != "user" && roleStr != "admin" {
				response.BadRequest(c, "Invalid role value, must be 'user' or 'admin'")
				return
			}
			user.Role = roleStr
		} else {
			response.BadRequest(c, "Invalid role field type")
			return
		}
	}
	
	if status, exists := updateData["status"]; exists {
		if statusStr, ok := status.(string); ok {
			if statusStr != "active" && statusStr != "inactive" && statusStr != "banned" {
				response.BadRequest(c, "Invalid status value, must be 'active', 'inactive', or 'banned'")
				return
			}
			user.Status = statusStr
		} else {
			response.BadRequest(c, "Invalid status field type")
			return
		}
	}

	// Update the user
	if err := h.userService.UpdateUser(c.Request.Context(), &user); err != nil {
		logger.Error("Admin failed to patch user",
			logger.Uint("user_id", uint(id)),
			logger.Any("update_data", updateData),
			logger.Error2("error", err),
		)
		response.InternalServerError(c, "Failed to update user")
		return
	}

	response.Success(c, user)
}

// BatchRestoreUsers godoc
// @Summary [Admin] Batch restore users
// @Description Restore multiple soft deleted users (admin only)
// @Tags Admin-User-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param ids body map[string][]uint true "User IDs"
// @Success 200 {object} response.StandardResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/users/batch/restore [post]
func (h *AdminUserHandler) BatchRestoreUsers(c *gin.Context) {
	var requestData struct {
		IDs []uint `json:"ids" binding:"required,min=1,max=100"`
	}

	if err := c.ShouldBindJSON(&requestData); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.userService.BatchRestoreUsers(c.Request.Context(), requestData.IDs)
	if err != nil {
		logger.Error("Admin failed to batch restore users",
			logger.Any("user_ids", requestData.IDs),
			logger.Error2("error", err),
		)
		response.InternalServerError(c, "Failed to restore users")
		return
	}

	response.SuccessWithMessage(c, "Users restored successfully", map[string]interface{}{
		"restored_count": result.RestoredCount,
		"failed_ids": result.FailedIDs,
	})
}

// ResetUserPassword godoc
// @Summary Reset user password
// @Description Reset a user's password (Admin only). Only works for local accounts.
// @Tags Admin-User-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path uint true "User ID"
// @Param password body ResetPasswordRequest true "New password data"
// @Success 200 {object} response.MessageOnlyResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/users/{id}/reset-password [post]
func (h *AdminUserHandler) ResetUserPassword(c *gin.Context) {
	// Get admin user from context
	adminUser, exists := c.Get(middleware.AuthContextKey)
	if !exists {
		response.Unauthorized(c, "Admin authentication required")
		return
	}

	admin, ok := adminUser.(*model.User)
	if !ok {
		response.InternalServerError(c, "Invalid admin user context")
		return
	}

	// Parse target user ID
	userIDStr := c.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	// Parse request body
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Reset password using auth service
	if err := h.authService.AdminResetPassword(c.Request.Context(), admin.ID, uint(userID), req.NewPassword); err != nil {
		logger.Error("Admin password reset failed",
			logger.Uint("admin_id", admin.ID),
			logger.String("admin_email", admin.Email),
			logger.Uint("target_user_id", uint(userID)),
			logger.Error2("error", err),
		)
		
		// Check specific error types for appropriate HTTP responses
		if strings.Contains(err.Error(), "user not found") {
			response.NotFound(c, err.Error())
		} else if strings.Contains(err.Error(), "insufficient permissions") {
			response.Forbidden(c, err.Error())
		} else if strings.Contains(err.Error(), "OAuth") || strings.Contains(err.Error(), "authentication") {
			response.BadRequest(c, err.Error())
		} else {
			response.InternalServerError(c, "Failed to reset password")
		}
		return
	}

	response.SuccessWithMessage(c, "User password reset successfully. All existing tokens have been revoked.", nil)
}

// ResetPasswordRequest represents the request structure for admin password reset
type ResetPasswordRequest struct {
	NewPassword string `json:"new_password" binding:"required,min=6,max=255" example:"newSecurePassword123"`
}