package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	authImplementations "linke/internal/domains/auth/usecases/implementations"
	authInterfaces "linke/internal/domains/auth/usecases/interfaces"
	userEntities "linke/internal/domains/user/entities"
	"linke/internal/shared/config"
	"linke/internal/shared/database"
	"linke/internal/shared/logger"
	"linke/internal/shared/middleware"
	"linke/internal/shared/response"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	cfg          *config.Config
	db           *database.Database
	oauthService authInterfaces.OAuthService
	authService  authInterfaces.AuthService
	jwtService   authInterfaces.JWTService
	stateStore   authInterfaces.OAuthStateService
}

func NewAuthHandler(cfg *config.Config, db *database.Database, authService authInterfaces.AuthService, jwtService authInterfaces.JWTService) *AuthHandler {
	return &AuthHandler{
		cfg:          cfg,
		db:           db,
		oauthService: authImplementations.NewOAuthService(cfg),
		authService:  authService,
		jwtService:   jwtService,
		stateStore:   authImplementations.NewOAuthStateStore(),
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
		response.InternalServerError(c, "Failed to exchange code for token: "+err.Error())
		return
	}

	logger.Debug("Getting user info from provider",
		logger.String("provider", provider))

	userInfo, err := h.oauthService.GetUserInfo(c.Request.Context(), provider, token)
	if err != nil {
		logger.Error("Failed to get user info",
			logger.String("provider", provider),
			logger.Error2("error", err))
		response.InternalServerError(c, "Failed to get user info: "+err.Error())
		return
	}

	logger.Debug("Creating or updating user",
		logger.String("provider", provider),
		logger.String("user_id", userInfo.ID),
		logger.String("email", userInfo.Email))

	user, err := h.createOrUpdateUser(userInfo)
	if err != nil {
		logger.Error("Failed to create or update user",
			logger.String("provider", provider),
			logger.String("user_id", userInfo.ID),
			logger.Error2("error", err))
		response.InternalServerError(c, "Failed to create or update user: "+err.Error())
		return
	}

	// Generate JWT token for the user
	jwtToken, err := h.jwtService.GenerateToken(user)
	if err != nil {
		response.InternalServerError(c, "Failed to generate JWT token: "+err.Error())
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
		response.Unauthorized(c, "Invalid Telegram authentication: "+err.Error())
		return
	}

	user, err := h.createOrUpdateUser(userInfo)
	if err != nil {
		response.InternalServerError(c, "Failed to create or update user: "+err.Error())
		return
	}

	// Generate JWT token for the user
	jwtToken, err := h.jwtService.GenerateToken(user)
	if err != nil {
		response.InternalServerError(c, "Failed to generate JWT token: "+err.Error())
		return
	}

	response.SuccessWithMessage(c, "Telegram authentication successful", gin.H{
		"user":  user,
		"token": jwtToken,
	})
}

func (h *AuthHandler) createOrUpdateUser(userInfo *authInterfaces.UserInfo) (*userEntities.User, error) {
	var user userEntities.User
	var userExists bool

	// Find user by provider-specific ID
	switch userInfo.Provider {
	case "google":
		result := h.db.DB.Where("google_id = ? AND status = ?", userInfo.ID, userEntities.UserStatusActive).First(&user)
		userExists = result.Error == nil
		if !userExists {
			user = userEntities.User{
				Email:    userInfo.Email,
				Name:     userInfo.Name,
				Avatar:   userInfo.Avatar,
				GoogleID: &userInfo.ID,
				Username: userInfo.Username,
				Provider: "google",
				Status:   userEntities.UserStatusActive,
				Role:     userEntities.UserRoleUser,
			}
		}

	case "github":
		result := h.db.DB.Where("github_id = ? AND status = ?", userInfo.ID, userEntities.UserStatusActive).First(&user)
		userExists = result.Error == nil
		if !userExists {
			user = userEntities.User{
				Email:    userInfo.Email,
				Name:     userInfo.Name,
				Avatar:   userInfo.Avatar,
				GitHubID: &userInfo.ID,
				Username: userInfo.Username,
				Provider: "github",
				Status:   userEntities.UserStatusActive,
				Role:     userEntities.UserRoleUser,
			}
		}

	case "telegram":
		result := h.db.DB.Where("telegram_id = ? AND status = ?", userInfo.ID, userEntities.UserStatusActive).First(&user)
		userExists = result.Error == nil
		if !userExists {
			user = userEntities.User{
				Email:      userInfo.Email,
				Name:       userInfo.Name,
				Avatar:     userInfo.Avatar,
				TelegramID: &userInfo.ID,
				Username:   userInfo.Username,
				Provider:   "telegram",
				Status:     userEntities.UserStatusActive,
				Role:       userEntities.UserRoleUser,
			}
		}

	default:
		return nil, gin.Error{Err: nil, Type: gin.ErrorTypePrivate}
	}

	// Handle user creation or update
	if !userExists {
		// Create new user
		providerDataBytes, _ := json.Marshal(userInfo)
		providerDataStr := string(providerDataBytes)
		user.ProviderData = &providerDataStr

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
			providerDataStr := string(providerDataBytes)
			user.ProviderData = &providerDataStr

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
func (h *AuthHandler) hasUserDataChanged(user *userEntities.User, userInfo *authInterfaces.UserInfo) bool {
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
	var req authInterfaces.RegisterRequest
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
	var req authInterfaces.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Add client context information
	ctx := context.WithValue(c.Request.Context(), "client_ip", c.ClientIP())
	ctx = context.WithValue(ctx, "user_agent", c.GetHeader("User-Agent"))

	authResponse, err := h.authService.Login(ctx, &req)
	if err != nil {
		logger.Warn("Login failed",
			logger.String("email", req.Email),
			logger.String("ip", c.ClientIP()),
			logger.Error2("error", err),
		)
		response.Unauthorized(c, err.Error())
		return
	}

	response.Success(c, authResponse)
}

// Logout godoc
// @Summary User logout
// @Description Logout user (server-side token revocation)
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.MessageOnlyResponse
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	user, exists := c.Get(middleware.AuthContextKey)
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	u, ok := user.(*userEntities.User)
	if !ok {
		response.InternalServerError(c, "Invalid user context")
		return
	}

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

	// Add client context information
	ctx := context.WithValue(c.Request.Context(), "client_ip", c.ClientIP())
	ctx = context.WithValue(ctx, "user_agent", c.GetHeader("User-Agent"))

	// Perform secure logout
	if err := h.authService.Logout(ctx, token, u.ID); err != nil {
		logger.Error("Logout failed",
			logger.Uint("user_id", u.ID),
			logger.Error2("error", err))
		response.InternalServerError(c, "Failed to logout securely")
		return
	}

	response.SuccessWithMessage(c, "Logged out successfully. Token has been revoked.", nil)
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

	u, ok := user.(*userEntities.User)
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

	u, ok := user.(*userEntities.User)
	if !ok {
		response.InternalServerError(c, "Invalid user context")
		return
	}

	response.Success(c, u.ToResponse())
}

// GetAuthURL godoc
// @Summary Get OAuth authorization URL
// @Description Generate OAuth authorization URL for frontend applications
// @Tags auth
// @Accept json
// @Produce json
// @Param request body service.AuthorizeURLRequest true "Authorization URL request"
// @Success 200 {object} response.StandardResponse{data=service.AuthorizeURLResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /auth/url [post]
func (h *AuthHandler) GetAuthURL(c *gin.Context) {
	var req authInterfaces.AuthorizeURLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request format", err.Error())
		return
	}

	// Validate provider
	if req.Provider != "google" && req.Provider != "github" && req.Provider != "telegram" {
		response.BadRequest(c, "Unsupported OAuth provider", "Supported providers: google, github, telegram")
		return
	}

	// Telegram uses different authentication flow, handle separately
	if req.Provider == "telegram" {
		url := h.oauthService.GetTelegramLoginURL()
		if url == "" {
			response.BadRequest(c, "Telegram bot token not configured")
			return
		}

		telegramResp := &authInterfaces.AuthorizeURLResponse{
			AuthURL: url,
			State:   "telegram-auth",
		}
		response.OK(c, "Authorization URL generated successfully", telegramResp)
		return
	}

	// Generate authorization URL
	state := req.State
	if state == "" {
		state = h.oauthService.GenerateState()
	}

	url, err := h.oauthService.GetAuthURL(req.Provider, state)
	if err != nil {
		response.InternalServerError(c, "Failed to generate authorization URL", err.Error())
		return
	}

	// Store state information for later validation
	stateInfo := &authInterfaces.OAuthStateInfo{
		Provider:    req.Provider,
		RedirectURI: req.RedirectURI,
		CreatedAt:   time.Now(),
		ExpiresAt:   time.Now().Add(10 * time.Minute),
	}
	h.stateStore.StoreState(state, stateInfo)

	logger.Info("Authorization URL generated",
		logger.String("provider", req.Provider),
		logger.String("state", state))

	response.OK(c, "Authorization URL generated successfully", &authInterfaces.AuthorizeURLResponse{
		AuthURL: url,
		State:   state,
	})
}

// ExchangeToken godoc
// @Summary Exchange authorization code for tokens
// @Description Exchange OAuth authorization code for JWT tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param request body TokenExchangeRequest true "Token exchange request"
// @Success 200 {object} response.StandardResponse{data=service.AuthResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /auth/token [post]
func (h *AuthHandler) ExchangeToken(c *gin.Context) {
	var req TokenExchangeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request format", err.Error())
		return
	}

	// Validate provider
	if req.Provider != "google" && req.Provider != "github" {
		response.BadRequest(c, "Unsupported OAuth provider for token exchange", "Supported providers: google, github")
		return
	}

	// Validate state parameter if provided
	if req.State != "" {
		stateInfo, err := h.stateStore.GetState(req.State)
		if err != nil {
			logger.Error("Invalid state parameter",
				logger.String("provider", req.Provider),
				logger.String("state", req.State),
				logger.Error2("error", err))
			response.Unauthorized(c, "Invalid or expired state parameter")
			return
		}

		// Verify provider matches
		if stateInfo.Provider != req.Provider {
			logger.Error("Provider mismatch in state",
				logger.String("expected_provider", stateInfo.Provider),
				logger.String("actual_provider", req.Provider))
			response.Unauthorized(c, "Provider mismatch")
			return
		}
	}

	// Exchange authorization code for token (simple OAuth flow)
	token, err := h.oauthService.ExchangeCodeForToken(c.Request.Context(), req.Provider, req.Code)
	if err != nil {
		logger.Error("Failed to exchange code for token",
			logger.String("provider", req.Provider),
			logger.String("code", req.Code[:10]+"..."),
			logger.Error2("error", err))
		response.Unauthorized(c, "Failed to exchange authorization code")
		return
	}

	// Get user info from OAuth provider
	userInfo, err := h.oauthService.GetUserInfo(c.Request.Context(), req.Provider, token)
	if err != nil {
		logger.Error("Failed to get user info",
			logger.String("provider", req.Provider),
			logger.Error2("error", err))
		response.InternalServerError(c, "Failed to get user information", err.Error())
		return
	}

	// Create or update user
	user, err := h.createOrUpdateUser(userInfo)
	if err != nil {
		logger.Error("Failed to create or update user",
			logger.String("provider", req.Provider),
			logger.String("provider_user_id", userInfo.ID),
			logger.Error2("error", err))
		response.InternalServerError(c, "Failed to process user information", err.Error())
		return
	}

	// Generate JWT token
	jwtToken, err := h.jwtService.GenerateToken(user)
	if err != nil {
		logger.Error("Failed to generate JWT token",
			logger.Uint("user_id", user.ID),
			logger.Error2("error", err))
		response.InternalServerError(c, "Failed to generate authentication tokens", err.Error())
		return
	}

	// Prepare response
	authResponse := &authInterfaces.AuthResponse{
		User:  user.ToResponse(),
		Token: jwtToken,
	}

	logger.Info("OAuth token exchange successful",
		logger.String("provider", req.Provider),
		logger.Uint("user_id", user.ID))

	response.OK(c, "Authentication successful", authResponse)
}

// TokenExchangeRequest represents token exchange request
type TokenExchangeRequest struct {
	Provider string `json:"provider" binding:"required"`
	Code     string `json:"code" binding:"required"`
	State    string `json:"state,omitempty"`
}
