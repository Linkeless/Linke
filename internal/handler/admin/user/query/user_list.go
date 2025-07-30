package query

import (
	"linke/internal/handler/admin/user/shared"
	"linke/internal/logger"
	"linke/internal/response"
	"linke/internal/service"

	"github.com/gin-gonic/gin"
)

// UserListHandler handles user listing and filtering operations
type UserListHandler struct {
	*shared.BaseHandler
}

// NewUserListHandler creates a new user list handler
func NewUserListHandler(userService *service.UserService, authService *service.AuthService) *UserListHandler {
	return &UserListHandler{
		BaseHandler: shared.NewBaseHandler(userService, authService),
	}
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
func (h *UserListHandler) ListUsers(c *gin.Context) {
	page, limit, offset := h.Validator.ValidatePaginationParams(c)

	users, total, err := h.UserService.ListUsers(c.Request.Context(), limit, offset)
	if err != nil {
		logger.Error("Admin failed to list users", logger.Error2("error", err))
		response.InternalServerError(c, "Failed to list users")
		return
	}

	response.SuccessList(c, users, page, limit, total)
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
func (h *UserListHandler) ListDeletedUsers(c *gin.Context) {
	page, limit, offset := h.Validator.ValidatePaginationParams(c)

	users, total, err := h.UserService.ListDeletedUsers(c.Request.Context(), limit, offset)
	if err != nil {
		logger.Error("Admin failed to list deleted users", logger.Error2("error", err))
		response.InternalServerError(c, "Failed to list deleted users")
		return
	}

	response.SuccessList(c, users, page, limit, total)
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
func (h *UserListHandler) ListUsersByProvider(c *gin.Context) {
	provider := c.Query("provider")
	if provider == "" {
		response.BadRequest(c, "Provider parameter is required")
		return
	}

	if err := h.Validator.ValidateProvider(provider); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	page, limit, offset := h.Validator.ValidatePaginationParams(c)

	users, total, err := h.UserService.ListUsersByProvider(c.Request.Context(), provider, limit, offset)
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