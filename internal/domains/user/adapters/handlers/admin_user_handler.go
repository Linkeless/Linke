package handlers

import (
	"strconv"
	"strings"

	authInterfaces "linke/internal/domains/auth/usecases/interfaces"
	"linke/internal/domains/user/constants"
	"linke/internal/domains/user/dto"
	"linke/internal/domains/user/entities"
	userInterfaces "linke/internal/domains/user/usecases/interfaces"
	"linke/internal/shared/handlers"
	"linke/internal/shared/logger"
	"linke/internal/shared/middleware"
	"linke/internal/shared/response"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type AdminUserHandler struct {
	userService userInterfaces.UserService
	authService authInterfaces.AuthService
}

func NewAdminUserHandler(userService userInterfaces.UserService, authService authInterfaces.AuthService) *AdminUserHandler {
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
// @Param user body dto.CreateUserRequest true "User creation data"
// @Success 201 {object} dto.UserResponse
// @Failure 400 {object} response.ProblemJSONResponse
// @Failure 401 {object} response.ProblemJSONResponse
// @Failure 403 {object} response.ProblemJSONResponse
// @Failure 409 {object} response.ProblemJSONResponse
// @Failure 500 {object} response.ProblemJSONResponse
// @Router /admin/users [post]
func (h *AdminUserHandler) CreateUser(c *gin.Context) {
	createReq, err := handlers.BindAndValidate[dto.CreateUserRequest](c)
	if err != nil {
		return
	}

	// Create user model from request
	user := &entities.User{
		Email:    createReq.Email,
		Username: createReq.Username,
		Name:     createReq.Name,
		Role:     createReq.Role,
		Status:   createReq.Status,
		Provider: constants.ProviderLocal,
	}

	// Set password if provided (hash the password before storing)
	if createReq.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(createReq.Password), bcrypt.DefaultCost)
		if err != nil {
			logger.Error("Failed to hash password during user creation",
				logger.String("email", createReq.Email),
				logger.ErrorField(err),
			)
			response.InternalServerError(c, "Failed to process password")
			return
		}
		user.Password = string(hashedPassword)
	}

	// Set default values if not provided
	if user.Role == "" {
		user.Role = constants.UserRoleUser
	}
	if user.Status == "" {
		user.Status = constants.UserStatusActive
	}

	// Create the user
	if err := h.userService.CreateUser(c.Request.Context(), user); err != nil {
		logger.Error("Admin failed to create user",
			logger.String("email", createReq.Email),
			logger.ErrorField(err),
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

	// Convert to response format
	userResponse := dto.ToUserResponse(user)
	response.Created(c, userResponse)
	// Return the response to pool after use
	dto.PutUserResponse(userResponse)
}

// GetUser godoc
// @Summary Get user information
// @Description Get user details by user ID (Admin only)
// @Tags Admin-User-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Success 200 {object} dto.UserResponse
// @Failure 400 {object} response.ProblemJSONResponse
// @Failure 401 {object} response.ProblemJSONResponse
// @Failure 403 {object} response.ProblemJSONResponse
// @Failure 404 {object} response.ProblemJSONResponse
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
			logger.ErrorField(err),
		)
		response.NotFound(c, "User not found")
		return
	}

	response.OK(c, user)
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
// @Param q query string false "Search by email/username/name (substring match)"
// @Param status query string false "Filter by status" Enums(active,inactive,banned)
// @Param role query string false "Filter by role" Enums(user,admin)
// @Param provider query string false "Filter by provider" Enums(local,google,github,telegram)
// @Success 200 {object} response.HALCollectionResponse
// @Failure 401 {object} response.ProblemJSONResponse
// @Failure 403 {object} response.ProblemJSONResponse
// @Failure 500 {object} response.ProblemJSONResponse
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

	// Optional filters
	q := strings.TrimSpace(c.Query("q"))
	status := strings.TrimSpace(c.Query("status"))
	role := strings.TrimSpace(c.Query("role"))
	provider := strings.TrimSpace(c.Query("provider"))

	req := &dto.AdvancedUserSearchRequest{
		Query:    q,
		Status:   status,
		Provider: provider,
		Role:     role,
		Limit:    limit,
		Offset:   offset,
	}

	users, total, err := h.userService.ListUsersFiltered(c.Request.Context(), req)
	if err != nil {
		logger.Error("Admin failed to list users", logger.ErrorField(err))
		response.InternalServerError(c, "Failed to list users")
		return
	}

	response.SendPaginatedResponse(c, users, total)
}

// UpdateUser godoc
// @Summary [Admin] Update any user
// @Description Update any user information (admin only)
// @Tags Admin-User-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Param user body dto.UpdateUserRequest true "User data"
// @Success 200 {object} dto.UserResponse
// @Failure 400 {object} response.ProblemJSONResponse
// @Failure 401 {object} response.ProblemJSONResponse
// @Failure 403 {object} response.ProblemJSONResponse
// @Failure 404 {object} response.ProblemJSONResponse
// @Failure 500 {object} response.ProblemJSONResponse
// @Router /admin/users/{id} [put]
func (h *AdminUserHandler) UpdateUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	// Get the existing user first
	existingUser, err := h.userService.GetUserByID(c.Request.Context(), uint(id))
	if err != nil {
		logger.Error("Failed to get user for update",
			logger.Uint("user_id", uint(id)),
			logger.ErrorField(err),
		)
		response.NotFound(c, "User not found")
		return
	}

	// Bind update data to UserResponse (for safe API binding)
	var updateData dto.UserResponse
	if err := c.ShouldBindJSON(&updateData); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Update the existing user with new data
	existingUser.Email = updateData.Email
	existingUser.Username = updateData.Username
	existingUser.Name = updateData.Name
	existingUser.Avatar = updateData.Avatar
	existingUser.Status = updateData.Status
	existingUser.Role = updateData.Role
	existingUser.Provider = updateData.Provider

	if err := h.userService.UpdateUser(c.Request.Context(), existingUser); err != nil {
		logger.Error("Admin failed to update user",
			logger.Uint("user_id", uint(id)),
			logger.ErrorField(err),
		)
		response.InternalServerError(c, "Failed to update user")
		return
	}

	// Convert to response format
	userResponse := dto.ToUserResponse(existingUser)
	response.OK(c, userResponse)
	// Return the response to pool after use
	dto.PutUserResponse(userResponse)
}

// UpdateUserRole godoc
// @Summary [Admin] Update user role
// @Description Update user role (admin only)
// @Tags Admin-User-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Param role body dto.UpdateUserRoleRequest true "Role data"
// @Success 200 {object} dto.UserResponse
// @Failure 400 {object} response.ProblemJSONResponse
// @Failure 401 {object} response.ProblemJSONResponse
// @Failure 403 {object} response.ProblemJSONResponse
// @Failure 404 {object} response.ProblemJSONResponse
// @Router /admin/users/{id}/role [put]
func (h *AdminUserHandler) UpdateUserRole(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	var roleData dto.UpdateUserRoleRequest

	if err := c.ShouldBindJSON(&roleData); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	user, err := h.userService.UpdateUserRole(c.Request.Context(), uint(id), roleData.Role)
	if err != nil {
		logger.Error("Admin failed to update user role",
			logger.Uint("user_id", uint(id)),
			logger.String("role", roleData.Role),
			logger.ErrorField(err),
		)
		response.NotFound(c, "User not found")
		return
	}

	response.OK(c, user)
}

// UpdateUserStatus godoc
// @Summary [Admin] Update user status
// @Description Update user status (admin only)
// @Tags Admin-User-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Param status body dto.UpdateUserStatusRequest true "Status data"
// @Success 200 {object} dto.UserResponse
// @Failure 400 {object} response.ProblemJSONResponse
// @Failure 401 {object} response.ProblemJSONResponse
// @Failure 403 {object} response.ProblemJSONResponse
// @Failure 404 {object} response.ProblemJSONResponse
// @Router /admin/users/{id}/status [put]
func (h *AdminUserHandler) UpdateUserStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	var statusData dto.UpdateUserStatusRequest

	if err := c.ShouldBindJSON(&statusData); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	user, err := h.userService.UpdateUserStatus(c.Request.Context(), uint(id), statusData.Status)
	if err != nil {
		logger.Error("Admin failed to update user status",
			logger.Uint("user_id", uint(id)),
			logger.String("status", statusData.Status),
			logger.ErrorField(err),
		)
		response.NotFound(c, "User not found")
		return
	}

	response.OK(c, user)
}

// SoftDeleteUser godoc
// @Summary [Admin] Soft delete user
// @Description Soft delete any user (admin only)
// @Tags Admin-User-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Success 200 {string} string "message"
// @Failure 400 {object} response.ProblemJSONResponse
// @Failure 401 {object} response.ProblemJSONResponse
// @Failure 403 {object} response.ProblemJSONResponse
// @Failure 404 {object} response.ProblemJSONResponse
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
			logger.ErrorField(err),
		)
		response.NotFound(c, "User not found")
		return
	}

	response.OK(c, gin.H{"message": "User deleted successfully"})
}

// RestoreUser godoc
// @Summary [Admin] Restore user
// @Description Restore a soft deleted user (admin only)
// @Tags Admin-User-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Success 200 {string} string "message"
// @Failure 400 {object} response.ProblemJSONResponse
// @Failure 401 {object} response.ProblemJSONResponse
// @Failure 403 {object} response.ProblemJSONResponse
// @Failure 404 {object} response.ProblemJSONResponse
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
			logger.ErrorField(err),
		)
		response.NotFound(c, "User not found")
		return
	}

	response.OK(c, gin.H{"message": "User restored successfully"})
}

// HardDeleteUser godoc
// @Summary [Admin] Hard delete user
// @Description Permanently delete a user from database (admin only)
// @Tags Admin-User-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Success 200 {string} string "message"
// @Failure 400 {object} response.ProblemJSONResponse
// @Failure 401 {object} response.ProblemJSONResponse
// @Failure 403 {object} response.ProblemJSONResponse
// @Failure 404 {object} response.ProblemJSONResponse
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
			logger.ErrorField(err),
		)
		response.NotFound(c, "User not found")
		return
	}

	response.OK(c, gin.H{"message": "User permanently deleted"})
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
// @Success 200 {object} response.HALCollectionResponse
// @Failure 401 {object} response.ProblemJSONResponse
// @Failure 403 {object} response.ProblemJSONResponse
// @Failure 500 {object} response.ProblemJSONResponse
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
		logger.Error("Admin failed to list deleted users", logger.ErrorField(err))
		response.InternalServerError(c, "Failed to list deleted users")
		return
	}

	response.SendPaginatedResponse(c, users, total)
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
// @Success 200 {object} response.HALCollectionResponse
// @Failure 400 {object} response.ProblemJSONResponse
// @Failure 401 {object} response.ProblemJSONResponse
// @Failure 403 {object} response.ProblemJSONResponse
// @Failure 500 {object} response.ProblemJSONResponse
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
			logger.ErrorField(err),
		)
		response.InternalServerError(c, "Failed to search users")
		return
	}

	response.SendFilteredCollection(c, users, total, map[string]string{"query": query})
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
// @Success 200 {object} response.HALCollectionResponse
// @Failure 400 {object} response.ProblemJSONResponse
// @Failure 401 {object} response.ProblemJSONResponse
// @Failure 403 {object} response.ProblemJSONResponse
// @Failure 500 {object} response.ProblemJSONResponse
// @Router /admin/users/provider [get]
func (h *AdminUserHandler) ListUsersByProvider(c *gin.Context) {
	provider := c.Query("provider")
	if provider == "" {
		response.BadRequest(c, "Provider parameter is required")
		return
	}

	validProviders := map[string]bool{
		constants.ProviderGoogle:   true,
		constants.ProviderGitHub:   true,
		constants.ProviderTelegram: true,
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
			logger.ErrorField(err),
		)
		response.InternalServerError(c, "Failed to list users")
		return
	}

	response.SendFilteredCollection(c, users, total, map[string]string{"provider": provider})
}

// GetUserStats godoc
// @Summary [Admin] Get user statistics
// @Description Get overall user statistics (admin only)
// @Tags Admin-User-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} response.ProblemJSONResponse
// @Failure 403 {object} response.ProblemJSONResponse
// @Failure 500 {object} response.ProblemJSONResponse
// @Router /admin/users/statistics [get]
func (h *AdminUserHandler) GetUserStats(c *gin.Context) {
	stats, err := h.userService.GetUserStats(c.Request.Context())
	if err != nil {
		logger.Error("Admin failed to get user stats", logger.ErrorField(err))
		response.InternalServerError(c, "Failed to get user statistics")
		return
	}

	response.OK(c, stats)
}

// BatchDeleteUsers godoc
// @Summary [Admin] Batch delete users
// @Description Soft delete multiple users (admin only)
// @Tags Admin-User-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param ids body dto.BatchUserIDsRequest true "User IDs"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} response.ProblemJSONResponse
// @Failure 401 {object} response.ProblemJSONResponse
// @Failure 403 {object} response.ProblemJSONResponse
// @Failure 500 {object} response.ProblemJSONResponse
// @Router /admin/users/bulk/delete [post]
func (h *AdminUserHandler) BatchDeleteUsers(c *gin.Context) {
	var requestData dto.BatchUserIDsRequest

	if err := c.ShouldBindJSON(&requestData); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.userService.BatchDeleteUsers(c.Request.Context(), requestData.IDs)
	if err != nil {
		logger.Error("Admin failed to batch delete users",
			logger.Any("user_ids", requestData.IDs),
			logger.ErrorField(err),
		)
		response.InternalServerError(c, "Failed to delete users")
		return
	}

	response.OK(c, gin.H{
		"deleted_count": result.DeletedCount,
		"failed_ids":    result.FailedIDs,
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
// @Param user body dto.PatchUserRequest true "Partial user data"
// @Success 200 {object} dto.UserResponse
// @Failure 400 {object} response.ProblemJSONResponse
// @Failure 401 {object} response.ProblemJSONResponse
// @Failure 403 {object} response.ProblemJSONResponse
// @Failure 404 {object} response.ProblemJSONResponse
// @Failure 500 {object} response.ProblemJSONResponse
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
			logger.ErrorField(err),
		)
		response.NotFound(c, "User not found")
		return
	}

	// Parse partial update data
	var patchReq dto.PatchUserRequest
	if err := c.ShouldBindJSON(&patchReq); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Apply partial updates to the current user
	user := *currentUser

	// Update only the fields present in the request
	if patchReq.Name != nil {
		user.Name = *patchReq.Name
	}

	if patchReq.Email != nil {
		user.Email = *patchReq.Email
	}

	if patchReq.Username != nil {
		user.Username = *patchReq.Username
	}

	if patchReq.Role != nil {
		if *patchReq.Role != "user" && *patchReq.Role != "admin" {
			response.BadRequest(c, "Invalid role value, must be 'user' or 'admin'")
			return
		}
		user.Role = *patchReq.Role
	}

	if patchReq.Status != nil {
		if *patchReq.Status != "active" && *patchReq.Status != "inactive" && *patchReq.Status != "banned" {
			response.BadRequest(c, "Invalid status value, must be 'active', 'inactive', or 'banned'")
			return
		}
		user.Status = *patchReq.Status
	}

	// Update the user
	if err := h.userService.UpdateUser(c.Request.Context(), &user); err != nil {
		logger.Error("Admin failed to patch user",
			logger.Uint("user_id", uint(id)),
			logger.Any("patch_request", patchReq),
			logger.ErrorField(err),
		)
		response.InternalServerError(c, "Failed to update user")
		return
	}

	response.OK(c, user)
}

// BatchRestoreUsers godoc
// @Summary [Admin] Batch restore users
// @Description Restore multiple soft deleted users (admin only)
// @Tags Admin-User-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param ids body dto.BatchUserIDsRequest true "User IDs"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} response.ProblemJSONResponse
// @Failure 401 {object} response.ProblemJSONResponse
// @Failure 403 {object} response.ProblemJSONResponse
// @Failure 500 {object} response.ProblemJSONResponse
// @Router /admin/users/bulk/restore [post]
func (h *AdminUserHandler) BatchRestoreUsers(c *gin.Context) {
	var requestData dto.BatchUserIDsRequest

	if err := c.ShouldBindJSON(&requestData); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.userService.BatchRestoreUsers(c.Request.Context(), requestData.IDs)
	if err != nil {
		logger.Error("Admin failed to batch restore users",
			logger.Any("user_ids", requestData.IDs),
			logger.ErrorField(err),
		)
		response.InternalServerError(c, "Failed to restore users")
		return
	}

	response.OK(c, gin.H{
		"restored_count": result.RestoredCount,
		"failed_ids":     result.FailedIDs,
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
// @Param password body dto.ResetPasswordRequest true "New password data"
// @Success 200 {string} string "message"
// @Failure 400 {object} response.ProblemJSONResponse
// @Failure 401 {object} response.ProblemJSONResponse
// @Failure 403 {object} response.ProblemJSONResponse
// @Failure 404 {object} response.ProblemJSONResponse
// @Failure 500 {object} response.ProblemJSONResponse
// @Router /admin/users/{id}/reset-password [post]
func (h *AdminUserHandler) ResetUserPassword(c *gin.Context) {
	// Get admin user from context
	adminUser, exists := c.Get(middleware.AuthContextKey)
	if !exists {
		response.Unauthorized(c, "Admin authentication required")
		return
	}

	admin, ok := adminUser.(*entities.User)
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
	var req dto.ResetPasswordRequest
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
			logger.ErrorField(err),
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

	response.OK(c, gin.H{"message": "User password reset successfully. All existing tokens have been revoked."})
}
