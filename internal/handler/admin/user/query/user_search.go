package query

import (
	"linke/internal/handler/admin/user/shared"
	"linke/internal/logger"
	"linke/internal/response"
	"linke/internal/service"

	"github.com/gin-gonic/gin"
)

// UserSearchHandler handles user search operations
type UserSearchHandler struct {
	*shared.BaseHandler
}

// NewUserSearchHandler creates a new user search handler
func NewUserSearchHandler(userService *service.UserService, authService *service.AuthService) *UserSearchHandler {
	return &UserSearchHandler{
		BaseHandler: shared.NewBaseHandler(userService, authService),
	}
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
func (h *UserSearchHandler) SearchUsers(c *gin.Context) {
	query, err := h.Validator.ValidateSearchQuery(c)
	if err != nil {
		return // Response already handled by validator
	}

	page, limit, offset := h.Validator.ValidatePaginationParams(c)

	users, total, err := h.UserService.SearchUsers(c.Request.Context(), query, limit, offset)
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