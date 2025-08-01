package api

import (
	"strings"

	"github.com/gin-gonic/gin"

	"linke/internal/response"
	"linke/internal/user/domain/valueobject"
	"linke/internal/user/handler/dto"
	"linke/internal/user/service/query"
)

// UserHandler handles user-related HTTP requests for regular users
type UserHandler struct {
	userQueryHandler *query.UserQueryHandler
}

// NewUserHandler creates a new UserHandler
func NewUserHandler(userQueryHandler *query.UserQueryHandler) *UserHandler {
	return &UserHandler{
		userQueryHandler: userQueryHandler,
	}
}

// GetUser godoc
// @Summary Get user by ID
// @Description Get user information by ID (public profile)
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} response.StandardResponse{data=dto.UserResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 404 {object} response.NotFoundResponse
// @Router /api/v1/user/users/{id} [get]
func (h *UserHandler) GetUser(c *gin.Context) {
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

	// Return public profile (excluding sensitive information)
	userResp := dto.FromUser(user)
	// Clear sensitive fields for public view
	userResp.ProviderData = nil
	userResp.InviteCodeID = nil
	userResp.InviteCodeUsed = nil

	response.OK(c, "User retrieved successfully", userResp)
}

// GetUserByUsername godoc
// @Summary Get user by username
// @Description Get user information by username (public profile)
// @Tags users
// @Accept json
// @Produce json
// @Param username path string true "Username"
// @Success 200 {object} response.StandardResponse{data=dto.UserResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 404 {object} response.NotFoundResponse
// @Router /api/v1/user/users/username/{username} [get]
func (h *UserHandler) GetUserByUsername(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		response.BadRequest(c, "Username is required")
		return
	}

	// Get user by username
	user, err := h.userQueryHandler.GetUserByUsername(c.Request.Context(), query.GetUserByUsernameQuery{Username: username})
	if err != nil {
		response.NotFound(c, "User not found")
		return
	}

	// Return public profile
	userResp := dto.FromUser(user)
	// Clear sensitive fields for public view
	userResp.ProviderData = nil
	userResp.InviteCodeID = nil
	userResp.InviteCodeUsed = nil

	response.OK(c, "User retrieved successfully", userResp)
}

// SearchUsers godoc
// @Summary Search users
// @Description Search users by email or username
// @Tags users
// @Accept json
// @Produce json
// @Param q query string true "Search query"
// @Param page query int false "Page number" default(1)
// @Param size query int false "Page size" default(10)
// @Param status query string false "User status filter" Enums(active, inactive, banned)
// @Param role query string false "User role filter" Enums(user, admin)
// @Param provider query string false "Provider filter" Enums(local, google, github, telegram)
// @Success 200 {object} response.StandardResponse{data=dto.UserListResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Router /api/v1/user/users/search [get]
func (h *UserHandler) SearchUsers(c *gin.Context) {
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

	// Convert to DTO and clear sensitive fields
	users := dto.FromUsers(result.Users)
	for i := range users {
		users[i].ProviderData = nil
		users[i].InviteCodeID = nil
		users[i].InviteCodeUsed = nil
	}

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

// CheckEmailExists godoc
// @Summary Check if email exists
// @Description Check if an email address is already registered
// @Tags users
// @Accept json
// @Produce json
// @Param email query string true "Email address to check"
// @Success 200 {object} response.StandardResponse{data=map[string]bool}
// @Failure 400 {object} response.BadRequestResponse
// @Router /api/v1/user/users/check/email [get]
func (h *UserHandler) CheckEmailExists(c *gin.Context) {
	email := c.Query("email")
	if email == "" {
		response.BadRequest(c, "Email parameter is required")
		return
	}

	// Check if email exists
	exists, err := h.userQueryHandler.CheckEmailExists(c.Request.Context(), query.CheckEmailExistsQuery{Email: email})
	if err != nil {
		response.InternalServerError(c, "Failed to check email", err.Error())
		return
	}

	response.OK(c, "Email check completed", map[string]bool{"exists": exists})
}

// CheckUsernameExists godoc
// @Summary Check if username exists
// @Description Check if a username is already taken
// @Tags users
// @Accept json
// @Produce json
// @Param username query string true "Username to check"
// @Success 200 {object} response.StandardResponse{data=map[string]bool}
// @Failure 400 {object} response.BadRequestResponse
// @Router /api/v1/user/users/check/username [get]
func (h *UserHandler) CheckUsernameExists(c *gin.Context) {
	username := c.Query("username")
	if username == "" {
		response.BadRequest(c, "Username parameter is required")
		return
	}

	// Check if username exists
	exists, err := h.userQueryHandler.CheckUsernameExists(c.Request.Context(), query.CheckUsernameExistsQuery{Username: username})
	if err != nil {
		response.InternalServerError(c, "Failed to check username", err.Error())
		return
	}

	response.OK(c, "Username check completed", map[string]bool{"exists": exists})
}

// GetUserStats godoc
// @Summary Get user statistics
// @Description Get public user statistics
// @Tags users
// @Accept json
// @Produce json
// @Param group_by query string false "Group statistics by" Enums(status, role, provider)
// @Success 200 {object} response.StandardResponse{data=dto.UserStatsResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Router /api/v1/user/users/stats [get]
func (h *UserHandler) GetUserStats(c *gin.Context) {
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

// GetUsers godoc
// @Summary List users
// @Description Get a paginated list of users (public profiles only)
// @Tags users
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param size query int false "Page size" default(10)
// @Param status query string false "User status filter" Enums(active, inactive, banned)
// @Param role query string false "User role filter" Enums(user, admin)
// @Param provider query string false "Provider filter" Enums(local, google, github, telegram)
// @Success 200 {object} response.StandardResponse{data=dto.UserListResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Router /api/v1/user/users [get]
func (h *UserHandler) GetUsers(c *gin.Context) {
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

	// Convert to DTO and clear sensitive fields for public view
	users := dto.FromUsers(result.Users)
	for i := range users {
		users[i].ProviderData = nil
		users[i].InviteCodeID = nil
		users[i].InviteCodeUsed = nil
	}

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

// GetMultipleUsers godoc
// @Summary Get multiple users by IDs
// @Description Get multiple users by their IDs (public profiles only)
// @Tags users
// @Accept json
// @Produce json
// @Param ids query string true "Comma-separated user IDs"
// @Success 200 {object} response.StandardResponse{data=[]dto.UserResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Router /api/v1/user/users/batch [get]
func (h *UserHandler) GetMultipleUsers(c *gin.Context) {
	idsParam := c.Query("ids")
	if idsParam == "" {
		response.BadRequest(c, "IDs parameter is required")
		return
	}

	// Parse comma-separated IDs
	idStrings := strings.Split(idsParam, ",")
	if len(idStrings) > 100 {
		response.BadRequest(c, "Maximum 100 IDs allowed")
		return
	}

	// Convert string IDs to UserIDs
	userIDs := make([]valueobject.UserID, 0, len(idStrings))
	for _, idStr := range idStrings {
		idStr = strings.TrimSpace(idStr)
		if idStr == "" {
			continue
		}

		userID, err := valueobject.NewUserIDFromString(idStr)
		if err != nil {
			response.BadRequest(c, "Invalid user ID: "+idStr, err.Error())
			return
		}
		userIDs = append(userIDs, userID)
	}

	if len(userIDs) == 0 {
		response.BadRequest(c, "No valid user IDs provided")
		return
	}

	// Get users by IDs
	users, err := h.userQueryHandler.GetUsersByIDs(c.Request.Context(), query.GetUsersByIDsQuery{UserIDs: userIDs})
	if err != nil {
		response.InternalServerError(c, "Failed to get users", err.Error())
		return
	}

	// Convert to DTO and clear sensitive fields
	userResponses := dto.FromUsers(users)
	for i := range userResponses {
		userResponses[i].ProviderData = nil
		userResponses[i].InviteCodeID = nil
		userResponses[i].InviteCodeUsed = nil
	}

	response.OK(c, "Users retrieved successfully", userResponses)
}