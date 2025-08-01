package auth

import (
	"strings"

	"github.com/gin-gonic/gin"

	"linke/internal/response"
	"linke/internal/user/domain/valueobject"
	"linke/internal/user/handler/dto"
	"linke/internal/user/infra/external"
	"linke/internal/user/service"
	"linke/internal/user/service/command"
	"linke/internal/user/service/query"
)

// AuthHandler handles authentication-related HTTP requests
type AuthHandler struct {
	userAppService   *service.UserApplicationService
	userQueryHandler *query.UserQueryHandler
	oauthService     *external.MultiProviderOAuthService
	jwtService       JWTService // This should be implemented elsewhere
}

// JWTService defines the interface for JWT operations
type JWTService interface {
	GenerateToken(user interface{}) (*dto.TokenResponse, error)
	ValidateToken(token string) (map[string]interface{}, error)
	RefreshToken(token string) (*dto.TokenResponse, error)
	RevokeToken(token string) error
}

// NewAuthHandler creates a new AuthHandler
func NewAuthHandler(
	userAppService *service.UserApplicationService,
	userQueryHandler *query.UserQueryHandler,
	oauthService *external.MultiProviderOAuthService,
	jwtService JWTService,
) *AuthHandler {
	return &AuthHandler{
		userAppService:   userAppService,
		userQueryHandler: userQueryHandler,
		oauthService:     oauthService,
		jwtService:       jwtService,
	}
}

// Register godoc
// @Summary User registration
// @Description Register a new user with email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param user body dto.CreateUserRequest true "Registration data"
// @Success 201 {object} response.StandardResponse{data=dto.AuthResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 409 {object} response.ConflictResponse
// @Router /api/v1/user/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
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
		response.Conflict(c, "Registration failed: "+err.Error())
		return
	}

	// Generate JWT token
	tokenResp, err := h.jwtService.GenerateToken(user)
	if err != nil {
		response.InternalServerError(c, "Failed to generate authentication token", err.Error())
		return
	}

	// Create response
	authResp := dto.AuthResponse{
		User:  dto.FromUser(user),
		Token: *tokenResp,
	}

	response.CreatedWithMessage(c, "User registered successfully", authResp)
}

// Login godoc
// @Summary User login
// @Description Login with email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param credentials body dto.LoginRequest true "Login credentials"
// @Success 200 {object} response.StandardResponse{data=dto.AuthResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Router /api/v1/user/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request format", err.Error())
		return
	}

	// Create command with client context
	cmd := command.AuthenticateUserCommand{
		Email:     req.Email,
		Password:  req.Password,
		IPAddress: c.ClientIP(),
		UserAgent: c.GetHeader("User-Agent"),
	}

	// Authenticate user
	user, err := h.userAppService.AuthenticateUser(c.Request.Context(), cmd)
	if err != nil {
		response.Unauthorized(c, "Authentication failed: "+err.Error())
		return
	}

	// Generate JWT token
	tokenResp, err := h.jwtService.GenerateToken(user)
	if err != nil {
		response.InternalServerError(c, "Failed to generate authentication token", err.Error())
		return
	}

	// Create response
	authResp := dto.AuthResponse{
		User:  dto.FromUser(user),
		Token: *tokenResp,
	}

	response.OK(c, "Login successful", authResp)
}

// Logout godoc
// @Summary User logout
// @Description Logout user (revoke token)
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.MessageOnlyResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Router /api/v1/user/auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	// Get the token from Authorization header
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		response.BadRequest(c, "Authorization header is required")
		return
	}

	tokenParts := strings.SplitN(authHeader, " ", 2)
	if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
		response.BadRequest(c, "Invalid authorization header format")
		return
	}

	token := tokenParts[1]

	// Revoke token
	if err := h.jwtService.RevokeToken(token); err != nil {
		response.InternalServerError(c, "Failed to logout securely", err.Error())
		return
	}

	response.OK(c, "Logged out successfully", nil)
}

// RefreshToken godoc
// @Summary Refresh JWT token
// @Description Refresh an existing JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.StandardResponse{data=dto.TokenResponse}
// @Failure 401 {object} response.UnauthorizedResponse
// @Router /api/v1/user/auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		response.Unauthorized(c, "Authorization header is required")
		return
	}

	tokenParts := strings.SplitN(authHeader, " ", 2)
	if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
		response.Unauthorized(c, "Invalid authorization header format")
		return
	}

	token := tokenParts[1]
	newToken, err := h.jwtService.RefreshToken(token)
	if err != nil {
		response.Unauthorized(c, "Token refresh failed: "+err.Error())
		return
	}

	response.OK(c, "Token refreshed successfully", newToken)
}

// ChangePassword godoc
// @Summary Change user password
// @Description Change password for local account users
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param passwords body dto.ChangePasswordRequest true "Password change data"
// @Success 200 {object} response.MessageOnlyResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Router /api/v1/user/auth/change-password [post]
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	// Get user ID from JWT token (this should be set by auth middleware)
	userIDStr, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	userID, err := valueobject.NewUserIDFromString(userIDStr.(string))
	if err != nil {
		response.Unauthorized(c, "Invalid user ID")
		return
	}

	var req dto.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request format", err.Error())
		return
	}

	// Create command
	cmd := command.ChangePasswordCommand{
		UserID:      userID,
		OldPassword: req.OldPassword,
		NewPassword: req.NewPassword,
	}

	// Change password
	if err := h.userAppService.ChangePassword(c.Request.Context(), cmd); err != nil {
		response.BadRequest(c, "Password change failed", err.Error())
		return
	}

	response.OK(c, "Password changed successfully", nil)
}

// GetProfile godoc
// @Summary Get user profile
// @Description Get current user's profile information
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.StandardResponse{data=dto.UserResponse}
// @Failure 401 {object} response.UnauthorizedResponse
// @Router /api/v1/user/auth/profile [get]
func (h *AuthHandler) GetProfile(c *gin.Context) {
	// Get user ID from JWT token
	userIDStr, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	userID, err := valueobject.NewUserIDFromString(userIDStr.(string))
	if err != nil {
		response.Unauthorized(c, "Invalid user ID")
		return
	}

	// Get user by ID
	user, err := h.userQueryHandler.GetUserByID(c.Request.Context(), query.GetUserByIDQuery{UserID: userID})
	if err != nil {
		response.NotFound(c, "User not found")
		return
	}

	response.OK(c, "Profile retrieved successfully", dto.FromUser(user))
}

// UpdateProfile godoc
// @Summary Update user profile
// @Description Update current user's profile information
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param profile body dto.UpdateProfileRequest true "Profile update data"
// @Success 200 {object} response.StandardResponse{data=dto.UserResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Router /api/v1/user/auth/profile [put]
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	// Get user ID from JWT token
	userIDStr, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	userID, err := valueobject.NewUserIDFromString(userIDStr.(string))
	if err != nil {
		response.Unauthorized(c, "Invalid user ID")
		return
	}

	var req dto.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request format", err.Error())
		return
	}

	// Create command
	cmd := command.UpdateUserProfileCommand{
		UserID:   userID,
		Name:     req.Name,
		Username: req.Username,
		Avatar:   req.Avatar,
	}

	// Update profile
	if err := h.userAppService.UpdateUserProfile(c.Request.Context(), cmd); err != nil {
		response.BadRequest(c, "Profile update failed", err.Error())
		return
	}

	// Get updated user
	user, err := h.userQueryHandler.GetUserByID(c.Request.Context(), query.GetUserByIDQuery{UserID: userID})
	if err != nil {
		response.InternalServerError(c, "Failed to fetch updated profile")
		return
	}

	response.OK(c, "Profile updated successfully", dto.FromUser(user))
}

// GetOAuthURL godoc
// @Summary Get OAuth authorization URL
// @Description Generate OAuth authorization URL for frontend applications
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.OAuthAuthURLRequest true "Authorization URL request"
// @Success 200 {object} response.StandardResponse{data=dto.OAuthAuthURLResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Router /api/v1/user/auth/oauth/url [post]
func (h *AuthHandler) GetOAuthURL(c *gin.Context) {
	var req dto.OAuthAuthURLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request format", err.Error())
		return
	}

	// Generate state if not provided
	state := "oauth-" + req.Provider
	if req.State != nil {
		state = *req.State
	}

	// Get authorization URL
	authURL, err := h.oauthService.GetAuthURL(req.Provider, state)
	if err != nil {
		response.BadRequest(c, "Failed to generate authorization URL", err.Error())
		return
	}

	resp := dto.OAuthAuthURLResponse{
		AuthURL: authURL,
		State:   state,
	}

	response.OK(c, "Authorization URL generated successfully", resp)
}

// ExchangeOAuthToken godoc
// @Summary Exchange OAuth authorization code for tokens
// @Description Exchange OAuth authorization code for JWT tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.OAuthTokenExchangeRequest true "Token exchange request"
// @Success 200 {object} response.StandardResponse{data=dto.AuthResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Router /api/v1/user/auth/oauth/token [post]
func (h *AuthHandler) ExchangeOAuthToken(c *gin.Context) {
	var req dto.OAuthTokenExchangeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request format", err.Error())
		return
	}

	// Exchange code for access token
	accessToken, err := h.oauthService.ExchangeCodeForToken(req.Provider, req.Code)
	if err != nil {
		response.Unauthorized(c, "Failed to exchange authorization code: "+err.Error())
		return
	}

	// Get user info from OAuth provider
	userInfo, err := h.oauthService.GetUserInfo(req.Provider, accessToken)
	if err != nil {
		response.InternalServerError(c, "Failed to get user information", err.Error())
		return
	}

	// Create OAuth authentication command
	cmd := command.AuthenticateOAuthUserCommand{
		Provider:     req.Provider,
		ProviderID:   userInfo.ID,
		Email:        userInfo.Email,
		Name:         userInfo.Name,
		Username:     &userInfo.Username,
		Avatar:       &userInfo.Avatar,
		IPAddress:    c.ClientIP(),
		UserAgent:    c.GetHeader("User-Agent"),
	}

	// Authenticate or create OAuth user
	user, err := h.userAppService.AuthenticateOAuthUser(c.Request.Context(), cmd)
	if err != nil {
		response.InternalServerError(c, "OAuth authentication failed", err.Error())
		return
	}

	// Generate JWT token
	tokenResp, err := h.jwtService.GenerateToken(user)
	if err != nil {
		response.InternalServerError(c, "Failed to generate authentication token", err.Error())
		return
	}

	// Create response
	authResp := dto.AuthResponse{
		User:  dto.FromUser(user),
		Token: *tokenResp,
	}

	response.OK(c, "OAuth authentication successful", authResp)
}

// AuthenticateTelegram godoc
// @Summary Authenticate with Telegram
// @Description Authenticate user using Telegram widget data
// @Tags auth
// @Accept json
// @Produce json
// @Param telegram_data body dto.TelegramAuthRequest true "Telegram authentication data"
// @Success 200 {object} response.StandardResponse{data=dto.AuthResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Router /api/v1/user/auth/telegram [post]
func (h *AuthHandler) AuthenticateTelegram(c *gin.Context) {
	var req dto.TelegramAuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request format", err.Error())
		return
	}

	// Convert request to map for validation
	data := map[string]string{
		"id":         req.ID,
		"first_name": req.FirstName,
		"auth_date":  req.AuthDate,
		"hash":       req.Hash,
	}

	if req.LastName != nil {
		data["last_name"] = *req.LastName
	}
	if req.Username != nil {
		data["username"] = *req.Username
	}
	if req.PhotoURL != nil {
		data["photo_url"] = *req.PhotoURL
	}

	// Validate Telegram authentication data
	userInfo, err := h.oauthService.ValidateTelegramAuth(data)
	if err != nil {
		response.Unauthorized(c, "Invalid Telegram authentication: "+err.Error())
		return
	}

	// Create OAuth authentication command
	cmd := command.AuthenticateOAuthUserCommand{
		Provider:     "telegram",
		ProviderID:   userInfo.ID,
		Email:        userInfo.Email,
		Name:         userInfo.Name,
		Username:     &userInfo.Username,
		Avatar:       &userInfo.Avatar,
		IPAddress:    c.ClientIP(),
		UserAgent:    c.GetHeader("User-Agent"),
	}

	// Authenticate or create OAuth user
	user, err := h.userAppService.AuthenticateOAuthUser(c.Request.Context(), cmd)
	if err != nil {
		response.InternalServerError(c, "Telegram authentication failed", err.Error())
		return
	}

	// Generate JWT token
	tokenResp, err := h.jwtService.GenerateToken(user)
	if err != nil {
		response.InternalServerError(c, "Failed to generate authentication token", err.Error())
		return
	}

	// Create response
	authResp := dto.AuthResponse{
		User:  dto.FromUser(user),
		Token: *tokenResp,
	}

	response.OK(c, "Telegram authentication successful", authResp)
}