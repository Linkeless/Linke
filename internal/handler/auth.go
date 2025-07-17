package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"linke/config"
	"linke/internal/logger"
	"linke/internal/middleware"
	"linke/internal/model"
	"linke/internal/repository"
	"linke/internal/response"
	"linke/internal/service"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	cfg          *config.Config
	db           *repository.Database
	oauthService *service.OAuthService
	authService  *service.AuthService
	jwtService   *service.JWTService
}

func NewAuthHandler(cfg *config.Config, db *repository.Database, authService *service.AuthService, jwtService *service.JWTService) *AuthHandler {
	return &AuthHandler{
		cfg:          cfg,
		db:           db,
		oauthService: service.NewOAuthService(cfg),
		authService:  authService,
		jwtService:   jwtService,
	}
}

// @Summary OAuth login
// @Description Initiate OAuth login for various providers
// @Tags auth
// @Param provider path string true "OAuth provider (google, github, telegram)"
// @Success 302 {string} string "redirect"
// @Failure 400 {object} response.BadRequestResponse
// @Router /auth/{provider} [get]
func (h *AuthHandler) Login(c *gin.Context) {
	provider := c.Param("provider")

	if provider == "telegram" {
		url := h.oauthService.GetTelegramLoginURL()
		if url == "" {
			response.BadRequest(c, "Telegram bot token not configured")
			return
		}
		c.Redirect(http.StatusFound, url)
		return
	}

	state := "oauth-state-" + provider
	url, err := h.oauthService.GetAuthURL(provider, state)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	c.Redirect(http.StatusFound, url)
}

// @Summary OAuth callback
// @Description Handle OAuth callback from providers
// @Tags auth
// @Param provider path string true "OAuth provider"
// @Param code query string false "Authorization code (for OAuth2)"
// @Param state query string false "State parameter (for OAuth2)"
// @Success 200 {object} response.StandardResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /auth/{provider}/callback [get]
func (h *AuthHandler) Callback(c *gin.Context) {
	provider := c.Param("provider")

	if provider == "telegram" {
		h.handleTelegramCallback(c)
		return
	}

	code := c.Query("code")
	state := c.Query("state")

	if code == "" {
		response.BadRequest(c, "Authorization code is required")
		return
	}

	expectedState := "oauth-state-" + provider
	if state != expectedState {
		response.BadRequest(c, "Invalid state parameter")
		return
	}

	token, err := h.oauthService.ExchangeCodeForToken(c.Request.Context(), provider, code)
	if err != nil {
		response.InternalServerError(c, "Failed to exchange code for token: " + err.Error())
		return
	}

	userInfo, err := h.oauthService.GetUserInfo(c.Request.Context(), provider, token)
	if err != nil {
		response.InternalServerError(c, "Failed to get user info: " + err.Error())
		return
	}

	user, err := h.createOrUpdateUser(userInfo)
	if err != nil {
		response.InternalServerError(c, "Failed to create or update user: " + err.Error())
		return
	}

	// Generate JWT token for the user
	jwtToken, err := h.jwtService.GenerateToken(user)
	if err != nil {
		response.InternalServerError(c, "Failed to generate JWT token: " + err.Error())
		return
	}

	response.SuccessWithMessage(c, "Authentication successful", gin.H{
		"user":  user,
		"token": jwtToken,
	})
}

// @Summary Get supported OAuth providers
// @Description Get list of supported OAuth providers
// @Tags auth
// @Success 200 {object} response.StandardResponse
// @Router /auth/providers [get]
func (h *AuthHandler) GetProviders(c *gin.Context) {
	providers := []map[string]interface{}{
		{
			"name":         "Google",
			"key":          "google",
			"login_url":    "/api/v1/auth/google",
			"callback_url": "/api/v1/auth/google/callback",
			"enabled":      h.cfg.OAuth2.GoogleClientID != "",
		},
		{
			"name":         "GitHub",
			"key":          "github",
			"login_url":    "/api/v1/auth/github",
			"callback_url": "/api/v1/auth/github/callback",
			"enabled":      h.cfg.OAuth2.GitHubClientID != "",
		},
		{
			"name":         "Telegram",
			"key":          "telegram",
			"login_url":    "/api/v1/auth/telegram",
			"callback_url": "/api/v1/auth/telegram/callback",
			"enabled":      h.cfg.OAuth2.TelegramBotToken != "",
		},
	}

	response.Success(c, gin.H{
		"providers": providers,
	})
}

// @Summary Get Telegram Login Widget
// @Description Get Telegram Login Widget HTML for frontend integration
// @Tags auth
// @Param bot_username query string false "Bot username (optional)"
// @Success 200 {object} response.StandardResponse
// @Failure 400 {object} response.BadRequestResponse
// @Router /auth/telegram/widget [get]
func (h *AuthHandler) GetTelegramWidget(c *gin.Context) {
	if h.cfg.OAuth2.TelegramBotToken == "" {
		response.BadRequest(c, "Telegram bot token not configured")
		return
	}

	botUsername := c.Query("bot_username")
	if botUsername == "" {
		botUsername = "YourBot"
	}

	redirectURL := h.cfg.OAuth2.TelegramRedirectURL

	widgetHTML := `<script async src="https://telegram.org/js/telegram-widget.js?22" 
		data-telegram-login="` + botUsername + `" 
		data-size="large" 
		data-auth-url="` + redirectURL + `"
		data-request-access="write"></script>`

	response.Success(c, gin.H{
		"widget_html":  widgetHTML,
		"redirect_url": redirectURL,
		"instructions": "1. Create a bot via @BotFather on Telegram\n2. Set domain with /setdomain command\n3. Replace 'YourBot' with your bot username\n4. Use the widget HTML in your frontend",
	})
}

func (h *AuthHandler) handleTelegramCallback(c *gin.Context) {
	data := make(map[string]string)

	for key, values := range c.Request.URL.Query() {
		if len(values) > 0 {
			data[key] = values[0]
		}
	}

	if len(data) == 0 {
		response.BadRequest(c, "No authentication data received")
		return
	}

	userInfo, err := h.oauthService.VerifyTelegramAuth(data)
	if err != nil {
		response.Unauthorized(c, "Invalid Telegram authentication: " + err.Error())
		return
	}

	user, err := h.createOrUpdateUser(userInfo)
	if err != nil {
		response.InternalServerError(c, "Failed to create or update user: " + err.Error())
		return
	}

	// Generate JWT token for the user
	jwtToken, err := h.jwtService.GenerateToken(user)
	if err != nil {
		response.InternalServerError(c, "Failed to generate JWT token: " + err.Error())
		return
	}

	response.SuccessWithMessage(c, "Telegram authentication successful", gin.H{
		"user":  user,
		"token": jwtToken,
	})
}

func (h *AuthHandler) createOrUpdateUser(userInfo *service.UserInfo) (*model.User, error) {
	var user model.User
	var userExists bool

	// Find user by provider-specific ID
	switch userInfo.Provider {
	case "google":
		result := h.db.DB.Where("google_id = ? AND status = ?", userInfo.ID, model.UserStatusActive).First(&user)
		userExists = result.Error == nil
		if !userExists {
			user = model.User{
				Email:    userInfo.Email,
				Name:     userInfo.Name,
				Avatar:   userInfo.Avatar,
				GoogleID: &userInfo.ID,
				Username: userInfo.Username,
				Provider: "google",
				Status:   model.UserStatusActive,
				Role:     model.UserRoleUser,
			}
		}

	case "github":
		result := h.db.DB.Where("github_id = ? AND status = ?", userInfo.ID, model.UserStatusActive).First(&user)
		userExists = result.Error == nil
		if !userExists {
			user = model.User{
				Email:    userInfo.Email,
				Name:     userInfo.Name,
				Avatar:   userInfo.Avatar,
				GitHubID: &userInfo.ID,
				Username: userInfo.Username,
				Provider: "github",
				Status:   model.UserStatusActive,
				Role:     model.UserRoleUser,
			}
		}

	case "telegram":
		result := h.db.DB.Where("telegram_id = ? AND status = ?", userInfo.ID, model.UserStatusActive).First(&user)
		userExists = result.Error == nil
		if !userExists {
			user = model.User{
				Email:      userInfo.Email,
				Name:       userInfo.Name,
				Avatar:     userInfo.Avatar,
				TelegramID: &userInfo.ID,
				Username:   userInfo.Username,
				Provider:   "telegram",
				Status:     model.UserStatusActive,
				Role:       model.UserRoleUser,
			}
		}

	default:
		return nil, gin.Error{Err: nil, Type: gin.ErrorTypePrivate}
	}

	// Handle user creation or update
	if !userExists {
		// Create new user
		providerDataBytes, _ := json.Marshal(userInfo)
		user.ProviderData = string(providerDataBytes)
		
		if err := h.db.DB.Create(&user).Error; err != nil {
			return nil, err
		}
		
		logger.Info("New OAuth user created",
			logger.String("provider", userInfo.Provider),
			logger.String("provider_id", userInfo.ID),
			logger.Uint("user_id", user.ID),
		)
	} else {
		// Check if user data has changed (only name and avatar)
		if h.hasUserDataChanged(&user, userInfo) {
			// Update only name and avatar fields
			user.Name = userInfo.Name
			user.Avatar = userInfo.Avatar
			
			// Update provider data to keep it current
			providerDataBytes, _ := json.Marshal(userInfo)
			user.ProviderData = string(providerDataBytes)
			
			if err := h.db.DB.Save(&user).Error; err != nil {
				return nil, err
			}
			
			logger.Info("OAuth user profile updated",
				logger.String("provider", userInfo.Provider),
				logger.String("provider_id", userInfo.ID),
				logger.Uint("user_id", user.ID),
				logger.String("updated_fields", "name,avatar"),
			)
		} else {
			logger.Debug("OAuth user profile unchanged, skipping update",
				logger.String("provider", userInfo.Provider),
				logger.String("provider_id", userInfo.ID),
				logger.Uint("user_id", user.ID),
			)
		}
	}

	return &user, nil
}

// hasUserDataChanged checks if user data has changed compared to OAuth provider data
// Only compares name and avatar fields as these are the main changeable fields from OAuth providers
func (h *AuthHandler) hasUserDataChanged(user *model.User, userInfo *service.UserInfo) bool {
	// Check only name and avatar fields that can be updated from OAuth provider
	if user.Name != userInfo.Name {
		return true
	}
	if user.Avatar != userInfo.Avatar {
		return true
	}
	
	return false
}

// Register godoc
// @Summary User registration
// @Description Register a new user with email and password. Username and name are auto-generated from email. Optional invite code can be provided.
// @Tags auth
// @Accept json
// @Produce json
// @Param user body service.RegisterRequest true "Registration data (email, password, and optional invite_code)"
// @Success 201 {object} response.StandardResponse{data=service.AuthResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 409 {object} response.ConflictResponse
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req service.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	authResponse, err := h.authService.Register(c.Request.Context(), &req)
	if err != nil {
		logger.Error("Registration failed",
			logger.String("email", req.Email),
			logger.Error2("error", err),
		)
		response.Conflict(c, err.Error())
		return
	}

	response.Created(c, authResponse)
}

// LoginLocal godoc
// @Summary User login with email/password
// @Description Login with email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param credentials body service.LoginRequest true "Login credentials"
// @Success 200 {object} response.StandardResponse{data=service.AuthResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Router /auth/login [post]
func (h *AuthHandler) LoginLocal(c *gin.Context) {
	var req service.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	authResponse, err := h.authService.Login(c.Request.Context(), &req)
	if err != nil {
		logger.Warn("Login failed",
			logger.String("email", req.Email),
			logger.Error2("error", err),
		)
		response.Unauthorized(c, err.Error())
		return
	}

	response.Success(c, authResponse)
}

// Logout godoc
// @Summary User logout
// @Description Logout user (client-side token invalidation)
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.MessageOnlyResponse
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	// For JWT tokens, logout is typically handled client-side by removing the token
	// Server-side logout would require a token blacklist, which can be implemented later
	user, exists := c.Get(middleware.AuthContextKey)
	if exists {
		if u, ok := user.(*model.User); ok {
			logger.Info("User logged out",
				logger.Uint("user_id", u.ID),
				logger.String("email", u.Email),
			)
		}
	}

	response.SuccessWithMessage(c, "Logged out successfully. Please remove the token from client storage.", nil)
}

// RefreshToken godoc
// @Summary Refresh JWT token
// @Description Refresh an existing JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.StandardResponse{data=service.TokenResponse}
// @Failure 401 {object} response.UnauthorizedResponse
// @Router /auth/refresh [post]
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
		logger.Warn("Token refresh failed",
			logger.Error2("error", err),
		)
		response.Unauthorized(c, err.Error())
		return
	}

	response.Success(c, newToken)
}

// ChangePassword godoc
// @Summary Change user password
// @Description Change password for local account users
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param passwords body map[string]string true "Password change data"
// @Success 200 {object} response.MessageOnlyResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Router /auth/change-password [post]
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	user, exists := c.Get(middleware.AuthContextKey)
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	u, ok := user.(*model.User)
	if !ok {
		response.InternalServerError(c, "Invalid user context")
		return
	}

	var req struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=6"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.authService.ChangePassword(c.Request.Context(), u.ID, req.OldPassword, req.NewPassword); err != nil {
		logger.Error("Password change failed",
			logger.Uint("user_id", u.ID),
			logger.Error2("error", err),
		)
		response.BadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "Password changed successfully", nil)
}

// GetProfile godoc
// @Summary Get user profile
// @Description Get current user's profile information
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.StandardResponse{data=model.UserResponse}
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /auth/profile [get]
func (h *AuthHandler) GetProfile(c *gin.Context) {
	user, exists := c.Get(middleware.AuthContextKey)
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	u, ok := user.(*model.User)
	if !ok {
		response.InternalServerError(c, "Invalid user context")
		return
	}

	response.Success(c, u.ToResponse())
}
