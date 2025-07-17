package handler

import (
	"strconv"
	"strings"

	"linke/internal/logger"
	"linke/internal/model"
	"linke/internal/response"
	"linke/internal/service"

	"github.com/gin-gonic/gin"
)

type AdminUserHandler struct {
	userService *service.UserService
}

func NewAdminUserHandler(userService *service.UserService) *AdminUserHandler {
	return &AdminUserHandler{
		userService: userService,
	}
}

// GetUser godoc
// @Summary [Admin] Get user by ID
// @Description Get any user information by user ID (admin only)
// @Tags admin-users
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
// @Summary [Admin] List all users
// @Description Get list of all users with pagination (admin only)
// @Tags admin-users
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
// @Tags admin-users
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
// @Tags admin-users
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
// @Tags admin-users
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
// @Tags admin-users
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
// @Tags admin-users
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
// @Tags admin-users
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
// @Tags admin-users
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
// @Tags admin-users
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
// @Tags admin-users
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
// @Tags admin-users
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
// @Tags admin-users
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

// BatchRestoreUsers godoc
// @Summary [Admin] Batch restore users
// @Description Restore multiple soft deleted users (admin only)
// @Tags admin-users
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