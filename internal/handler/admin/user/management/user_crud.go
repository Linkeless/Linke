package management

import (
	"strings"

	"linke/internal/handler/admin/user/shared"
	"linke/internal/logger"
	"linke/internal/model"
	"linke/internal/response"
	"linke/internal/service"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// UserCRUDHandler handles user CRUD operations
type UserCRUDHandler struct {
	*shared.BaseHandler
}

// NewUserCRUDHandler creates a new user CRUD handler
func NewUserCRUDHandler(userService *service.UserService, authService *service.AuthService) *UserCRUDHandler {
	return &UserCRUDHandler{
		BaseHandler: shared.NewBaseHandler(userService, authService),
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
func (h *UserCRUDHandler) CreateUser(c *gin.Context) {
	var createReq model.CreateUserRequest
	if err := c.ShouldBindJSON(&createReq); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Validate the request
	if err := h.Validator.ValidateCreateUserRequest(&createReq); err != nil {
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

	// Set password if provided
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
	if err := h.UserService.CreateUser(c.Request.Context(), user); err != nil {
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
func (h *UserCRUDHandler) GetUser(c *gin.Context) {
	id, err := h.Validator.ValidateUserID(c)
	if err != nil {
		return // Response already handled by validator
	}

	user, err := h.UserService.GetUserByID(c.Request.Context(), id)
	if err != nil {
		logger.Error("Admin failed to get user",
			logger.Uint("user_id", id),
			logger.Error2("error", err),
		)
		response.NotFound(c, "User not found")
		return
	}

	response.Success(c, user)
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
func (h *UserCRUDHandler) UpdateUser(c *gin.Context) {
	id, err := h.Validator.ValidateUserID(c)
	if err != nil {
		return // Response already handled by validator
	}

	var user model.User
	if err := c.ShouldBindJSON(&user); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	user.ID = id
	if err := h.UserService.UpdateUser(c.Request.Context(), &user); err != nil {
		logger.Error("Admin failed to update user",
			logger.Uint("user_id", id),
			logger.Error2("error", err),
		)
		response.InternalServerError(c, "Failed to update user")
		return
	}

	response.Success(c, user)
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
func (h *UserCRUDHandler) PatchUser(c *gin.Context) {
	id, err := h.Validator.ValidateUserID(c)
	if err != nil {
		return // Response already handled by validator
	}

	// Get current user
	currentUser, err := h.UserService.GetUserByID(c.Request.Context(), id)
	if err != nil {
		logger.Error("Admin failed to get user for patch",
			logger.Uint("user_id", id),
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

	// Validate partial update data
	if err := h.Validator.ValidatePartialUserUpdate(updateData); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Apply partial updates to the current user
	user := *currentUser
	
	// Update only the fields present in the request
	if name, exists := updateData["name"]; exists {
		user.Name = name.(string)
	}
	
	if email, exists := updateData["email"]; exists {
		user.Email = email.(string)
	}
	
	if username, exists := updateData["username"]; exists {
		user.Username = username.(string)
	}
	
	if role, exists := updateData["role"]; exists {
		user.Role = role.(string)
	}
	
	if status, exists := updateData["status"]; exists {
		user.Status = status.(string)
	}

	// Update the user
	if err := h.UserService.UpdateUser(c.Request.Context(), &user); err != nil {
		logger.Error("Admin failed to patch user",
			logger.Uint("user_id", id),
			logger.Any("update_data", updateData),
			logger.Error2("error", err),
		)
		response.InternalServerError(c, "Failed to update user")
		return
	}

	response.Success(c, user)
}