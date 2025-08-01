package admin

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"linke/internal/response"
	"linke/internal/user/domain/valueobject"
	"linke/internal/user/handler/dto"
	"linke/internal/user/service"
	"linke/internal/user/service/command"
	"linke/internal/user/service/query"
)

// UserAdminHandler handles user administration HTTP requests
type UserAdminHandler struct {
	userAppService   *service.UserApplicationService
	userQueryHandler *query.UserQueryHandler
}

// NewUserAdminHandler creates a new UserAdminHandler
func NewUserAdminHandler(
	userAppService *service.UserApplicationService,
	userQueryHandler *query.UserQueryHandler,
) *UserAdminHandler {
	return &UserAdminHandler{
		userAppService:   userAppService,
		userQueryHandler: userQueryHandler,
	}
}

// GetUser godoc
// @Summary Get user by ID (Admin)
// @Description Get detailed user information by ID (includes all fields)
// @Tags admin/users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID"
// @Success 200 {object} response.StandardResponse{data=dto.UserResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Router /api/v1/admin/users/{id} [get]
func (h *UserAdminHandler) GetUser(c *gin.Context) {
	idParam := c.Param("id")
	userID, err := valueobject.NewUserIDFromString(idParam)
	if err != nil {
		response.BadRequest(c, "Invalid user ID", err.Error())
		return
	}

	// Get user by ID
	user, err := h.userQueryHandler.GetUserByID(c.Request.Context(), query.GetUserByIDQuery{UserID: userID})
	if err != nil {
		response.NotFound(c, "User not found")
		return
	}

	// Return full user information (admin view)
	response.OK(c, "User retrieved successfully", dto.FromUser(user))
}

// ListUsers godoc
// @Summary List users (Admin)
// @Description Get a paginated list of users with full details
// @Tags admin/users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param size query int false "Page size" default(10)
// @Param status query string false "User status filter" Enums(active, inactive, banned)
// @Param role query string false "User role filter" Enums(user, admin)
// @Param provider query string false "Provider filter" Enums(local, google, github, telegram)
// @Success 200 {object} response.StandardResponse{data=dto.UserListResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Router /api/v1/admin/users [get]
func (h *UserAdminHandler) ListUsers(c *gin.Context) {
	var queryParams dto.ListUsersQuery
	if err := c.ShouldBindQuery(&queryParams); err != nil {
		response.BadRequest(c, "Invalid query parameters", err.Error())
		return
	}

	// Validate and set defaults
	queryParams.Validate()

	// Convert to service query
	serviceQuery := query.ListUsersQuery{
		Page:     queryParams.Page,
		Size:     queryParams.Size,
		Status:   queryParams.Status,
		Role:     queryParams.Role,
		Provider: queryParams.Provider,
	}

	// List users
	result, err := h.userQueryHandler.ListUsers(c.Request.Context(), serviceQuery)
	if err != nil {
		response.InternalServerError(c, "Failed to list users", err.Error())
		return
	}

	// Convert to DTO (admin view includes all fields)
	users := dto.FromUsers(result.Users)

	// Create response
	resp := dto.UserListResponse{
		Users:      users,
		Total:      result.Total,
		Page:       result.Page,
		Size:       result.Size,
		TotalPages: result.TotalPages,
		HasNext:    result.HasNext,
		HasPrev:    result.HasPrev,
	}

	response.OK(c, "Users retrieved successfully", resp)
}

// SearchUsers godoc
// @Summary Search users (Admin)
// @Description Search users by email or username with full details
// @Tags admin/users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param q query string true "Search query"
// @Param page query int false "Page number" default(1)
// @Param size query int false "Page size" default(10)
// @Param status query string false "User status filter" Enums(active, inactive, banned)
// @Param role query string false "User role filter" Enums(user, admin)
// @Param provider query string false "Provider filter" Enums(local, google, github, telegram)
// @Success 200 {object} response.StandardResponse{data=dto.UserListResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Router /api/v1/admin/users/search [get]
func (h *UserAdminHandler) SearchUsers(c *gin.Context) {
	var queryParams dto.SearchUsersQuery
	if err := c.ShouldBindQuery(&queryParams); err != nil {
		response.BadRequest(c, "Invalid query parameters", err.Error())
		return
	}

	// Validate and set defaults
	queryParams.Validate()

	// Convert to service query
	serviceQuery := query.SearchUsersQuery{
		Query:    queryParams.Query,
		Page:     queryParams.Page,
		Size:     queryParams.Size,
		Status:   queryParams.Status,
		Role:     queryParams.Role,
		Provider: queryParams.Provider,
	}

	// Search users
	result, err := h.userQueryHandler.SearchUsers(c.Request.Context(), serviceQuery)
	if err != nil {
		response.InternalServerError(c, "Failed to search users", err.Error())
		return
	}

	// Convert to DTO (admin view includes all fields)
	users := dto.FromUsers(result.Users)

	// Create response
	resp := dto.UserListResponse{
		Users:      users,
		Total:      result.Total,
		Page:       result.Page,
		Size:       result.Size,
		TotalPages: result.TotalPages,
		HasNext:    result.HasNext,
		HasPrev:    result.HasPrev,
	}

	response.OK(c, "Users retrieved successfully", resp)
}

// CreateUser godoc
// @Summary Create user (Admin)
// @Description Create a new user account (admin operation)
// @Tags admin/users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param user body dto.CreateUserRequest true "User creation data"
// @Success 201 {object} response.StandardResponse{data=dto.UserResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 409 {object} response.ConflictResponse
// @Router /api/v1/admin/users [post]
func (h *UserAdminHandler) CreateUser(c *gin.Context) {
	var req dto.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request format", err.Error())
		return
	}

	// Create command
	cmd := command.CreateUserCommand{
		Email:      req.Email,
		Password:   req.Password,
		Username:   req.Username,
		Name:       req.Name,
		InviteCode: req.InviteCode,
	}

	// Create user
	user, err := h.userAppService.CreateUser(c.Request.Context(), cmd)
	if err != nil {
		response.Conflict(c, "User creation failed: "+err.Error())
		return
	}

	response.CreatedWithMessage(c, "User created successfully", dto.FromUser(user))
}

// UpdateUserStatus godoc
// @Summary Update user status (Admin)
// @Description Change a user's status (active, inactive, banned)
// @Tags admin/users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID"
// @Param status body dto.ChangeUserStatusRequest true "Status change data"
// @Success 200 {object} response.StandardResponse{data=dto.UserResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Router /api/v1/admin/users/{id}/status [put]
func (h *UserAdminHandler) UpdateUserStatus(c *gin.Context) {
	idParam := c.Param("id")
	userID, err := valueobject.NewUserIDFromString(idParam)
	if err != nil {
		response.BadRequest(c, "Invalid user ID", err.Error())
		return
	}

	var req dto.ChangeUserStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request format", err.Error())
		return
	}

	// Create command
	cmd := command.ChangeUserStatusCommand{
		UserID:    userID,
		NewStatus: req.Status,
		Reason:    req.Reason,
	}

	// Change user status
	if err := h.userAppService.ChangeUserStatus(c.Request.Context(), cmd); err != nil {
		response.BadRequest(c, "Status change failed", err.Error())
		return
	}

	// Get updated user
	user, err := h.userQueryHandler.GetUserByID(c.Request.Context(), query.GetUserByIDQuery{UserID: userID})
	if err != nil {
		response.InternalServerError(c, "Failed to fetch updated user")
		return
	}

	response.OK(c, "User status updated successfully", dto.FromUser(user))
}

// UpdateUserRole godoc
// @Summary Update user role (Admin)
// @Description Change a user's role (user, admin)
// @Tags admin/users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID"
// @Param role body dto.ChangeUserRoleRequest true "Role change data"
// @Success 200 {object} response.StandardResponse{data=dto.UserResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Router /api/v1/admin/users/{id}/role [put]
func (h *UserAdminHandler) UpdateUserRole(c *gin.Context) {
	idParam := c.Param("id")
	userID, err := valueobject.NewUserIDFromString(idParam)
	if err != nil {
		response.BadRequest(c, "Invalid user ID", err.Error())
		return
	}

	// Get admin user ID from context (set by auth middleware)
	adminUserIDStr, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "Admin user not authenticated")
		return
	}

	adminUserID, err := valueobject.NewUserIDFromString(adminUserIDStr.(string))
	if err != nil {
		response.Unauthorized(c, "Invalid admin user ID")
		return
	}

	var req dto.ChangeUserRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request format", err.Error())
		return
	}

	// Create command
	cmd := command.ChangeUserRoleCommand{
		UserID:    userID,
		NewRole:   req.Role,
		ChangedBy: adminUserID,
		Reason:    req.Reason,
	}

	// Change user role
	if err := h.userAppService.ChangeUserRole(c.Request.Context(), cmd); err != nil {
		response.BadRequest(c, "Role change failed", err.Error())
		return
	}

	// Get updated user
	user, err := h.userQueryHandler.GetUserByID(c.Request.Context(), query.GetUserByIDQuery{UserID: userID})
	if err != nil {
		response.InternalServerError(c, "Failed to fetch updated user")
		return
	}

	response.OK(c, "User role updated successfully", dto.FromUser(user))
}

// DeleteUser godoc
// @Summary Delete user (Admin)
// @Description Soft delete a user account
// @Tags admin/users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID"
// @Param delete_data body dto.DeleteUserRequest true "Deletion data"
// @Success 200 {object} response.MessageOnlyResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Router /api/v1/admin/users/{id} [delete]
func (h *UserAdminHandler) DeleteUser(c *gin.Context) {
	idParam := c.Param("id")
	userID, err := valueobject.NewUserIDFromString(idParam)
	if err != nil {
		response.BadRequest(c, "Invalid user ID", err.Error())
		return
	}

	// Get admin user ID from context
	adminUserIDStr, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "Admin user not authenticated")
		return
	}

	adminUserID, err := valueobject.NewUserIDFromString(adminUserIDStr.(string))
	if err != nil {
		response.Unauthorized(c, "Invalid admin user ID")
		return
	}

	var req dto.DeleteUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request format", err.Error())
		return
	}

	// Create command
	cmd := command.DeleteUserCommand{
		UserID:    userID,
		DeletedBy: adminUserID,
		Reason:    req.Reason,
	}

	// Delete user
	if err := h.userAppService.DeleteUser(c.Request.Context(), cmd); err != nil {
		response.BadRequest(c, "User deletion failed", err.Error())
		return
	}

	response.OK(c, "User deleted successfully", nil)
}

// RestoreUser godoc
// @Summary Restore deleted user (Admin)
// @Description Restore a soft deleted user account
// @Tags admin/users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID"
// @Param restore_data body dto.RestoreUserRequest true "Restoration data"
// @Success 200 {object} response.StandardResponse{data=dto.UserResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Router /api/v1/admin/users/{id}/restore [post]
func (h *UserAdminHandler) RestoreUser(c *gin.Context) {
	idParam := c.Param("id")
	userID, err := valueobject.NewUserIDFromString(idParam)
	if err != nil {
		response.BadRequest(c, "Invalid user ID", err.Error())
		return
	}

	// Get admin user ID from context
	adminUserIDStr, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "Admin user not authenticated")
		return
	}

	adminUserID, err := valueobject.NewUserIDFromString(adminUserIDStr.(string))
	if err != nil {
		response.Unauthorized(c, "Invalid admin user ID")
		return
	}

	var req dto.RestoreUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request format", err.Error())
		return
	}

	// Create command
	cmd := command.RestoreUserCommand{
		UserID:     userID,
		RestoredBy: adminUserID,
		Reason:     req.Reason,
	}

	// Restore user
	if err := h.userAppService.RestoreUser(c.Request.Context(), cmd); err != nil {
		response.BadRequest(c, "User restoration failed", err.Error())
		return
	}

	// Get restored user
	user, err := h.userQueryHandler.GetUserByID(c.Request.Context(), query.GetUserByIDQuery{UserID: userID})
	if err != nil {
		response.InternalServerError(c, "Failed to fetch restored user")
		return
	}

	response.OK(c, "User restored successfully", dto.FromUser(user))
}

// GetDeletedUsers godoc
// @Summary Get deleted users (Admin)
// @Description Get a list of soft deleted users
// @Tags admin/users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param size query int false "Page size" default(10)
// @Success 200 {object} response.StandardResponse{data=dto.UserListResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Router /api/v1/admin/users/deleted [get]
func (h *UserAdminHandler) GetDeletedUsers(c *gin.Context) {
	page := 1
	size := 10

	if pageParam := c.Query("page"); pageParam != "" {
		if p, err := strconv.Atoi(pageParam); err == nil && p > 0 {
			page = p
		}
	}

	if sizeParam := c.Query("size"); sizeParam != "" {
		if s, err := strconv.Atoi(sizeParam); err == nil && s > 0 && s <= 100 {
			size = s
		}
	}

	// This would require extending the query service to support deleted users
	// For now, return an empty list as a placeholder
	resp := dto.UserListResponse{
		Users:      []dto.UserResponse{},
		Total:      0,
		Page:       page,
		Size:       size,
		TotalPages: 0,
		HasNext:    false,
		HasPrev:    false,
	}

	response.OK(c, "Deleted users retrieved successfully", resp)
}

// GetUserStats godoc
// @Summary Get comprehensive user statistics (Admin)
// @Description Get detailed user statistics with all data
// @Tags admin/users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param group_by query string false "Group statistics by" Enums(status, role, provider)
// @Success 200 {object} response.StandardResponse{data=dto.UserStatsResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Router /api/v1/admin/users/stats [get]
func (h *UserAdminHandler) GetUserStats(c *gin.Context) {
	var queryParams dto.UserStatsQuery
	if err := c.ShouldBindQuery(&queryParams); err != nil {
		response.BadRequest(c, "Invalid query parameters", err.Error())
		return
	}

	// Convert to service query
	serviceQuery := query.GetUserStatsQuery{
		GroupBy: queryParams.GroupBy,
	}

	// Get user stats
	stats, err := h.userQueryHandler.GetUserStats(c.Request.Context(), serviceQuery)
	if err != nil {
		response.InternalServerError(c, "Failed to get user statistics", err.Error())
		return
	}

	// Convert to DTO
	resp := dto.UserStatsResponse{
		Total:      stats.Total,
		ByStatus:   stats.ByStatus,
		ByRole:     stats.ByRole,
		ByProvider: stats.ByProvider,
	}

	response.OK(c, "User statistics retrieved successfully", resp)
}