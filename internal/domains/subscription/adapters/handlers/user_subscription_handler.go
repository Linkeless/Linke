package handlers

import (
	"net/http"
	"strconv"
	"strings"

	authInterfaces "linke/internal/domains/auth/usecases/interfaces"
	serverEntities "linke/internal/domains/server/entities"
	"linke/internal/domains/subscription/entities"
	"linke/internal/domains/subscription/usecases/interfaces"
	userEntities "linke/internal/domains/user/entities"
	"linke/internal/shared/logger"
	"linke/internal/shared/middleware"
	"linke/internal/shared/response"

	"gopkg.in/yaml.v3"

	"github.com/gin-gonic/gin"
)

type UserSubscriptionHandler struct {
	userSubscriptionService interfaces.UserSubscriptionService
	authService             authInterfaces.AuthService
}

func NewUserSubscriptionHandler(userSubscriptionService interfaces.UserSubscriptionService, authService authInterfaces.AuthService) *UserSubscriptionHandler {
	return &UserSubscriptionHandler{
		userSubscriptionService: userSubscriptionService,
		authService:             authService,
	}
}

// GetMySubscriptions godoc
// @Summary [User] Get my subscriptions
// @Description Get current user's subscriptions with optional filtering
// @Tags User-Subscription
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param status query string false "Filter by status" Enums(active, paused, cancelled, expired, trial)
// @Param limit query int false "Limit results" minimum(1) maximum(100) example(10)
// @Param offset query int false "Offset results" minimum(0) example(0)
// @Success 200 {object} response.HALCollectionResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /subscriptions/my [get]
func (h *UserSubscriptionHandler) GetMySubscriptions(c *gin.Context) {
	// Get current user from context
	userValue, exists := c.Get(middleware.AuthContextKey)
	if !exists {
		response.Unauthorized(c, "Authentication required")
		return
	}

	user, ok := userValue.(*userEntities.User)
	if !ok {
		response.Unauthorized(c, "Invalid user context")
		return
	}

	// Parse query parameters
	var req interfaces.GetUserSubscriptionsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "Invalid query parameters")
		return
	}

	// Force user ID to current user
	req.UserID = user.ID

	// Set defaults
	if req.Limit == 0 {
		req.Limit = 10
	}

	// Get user subscriptions
	subscriptions, totalCount, err := h.userSubscriptionService.GetUserSubscriptions(c.Request.Context(), &req)
	if err != nil {
		logger.Error("Failed to get user subscriptions", logger.ErrorField(err), logger.Uint("user_id", user.ID))
		response.InternalServerError(c, "Failed to get subscriptions")
		return
	}

	// Convert to response format
	var subscriptionResponses []*entities.UserSubscriptionResponse
	for _, sub := range subscriptions {
		subscriptionResponses = append(subscriptionResponses, sub.ToResponse())
	}

	// Convert offset to page number
	_ = (req.Offset / req.Limit) + 1 // page calculation
	response.SendPaginatedResponse(c, subscriptionResponses, totalCount)
}

// GetMyActiveSubscriptions godoc
// @Summary [User] Get my active subscriptions
// @Description Get current user's active subscriptions
// @Tags User-Subscription
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} []entities.UserSubscriptionResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /subscriptions/my/active [get]
func (h *UserSubscriptionHandler) GetMyActiveSubscriptions(c *gin.Context) {
	// Get current user from context
	userValue, exists := c.Get(middleware.AuthContextKey)
	if !exists {
		response.Unauthorized(c, "Authentication required")
		return
	}

	user, ok := userValue.(*userEntities.User)
	if !ok {
		response.Unauthorized(c, "Invalid user context")
		return
	}

	// Get active subscriptions
	subscriptions, err := h.userSubscriptionService.GetUserActiveSubscriptions(c.Request.Context(), user.ID)
	if err != nil {
		logger.Error("Failed to get active subscriptions", logger.ErrorField(err), logger.Uint("user_id", user.ID))
		response.InternalServerError(c, "Failed to get active subscriptions")
		return
	}

	// Convert to response format
	var subscriptionResponses []*entities.UserSubscriptionResponse
	for _, sub := range subscriptions {
		subscriptionResponses = append(subscriptionResponses, sub.ToResponse())
	}

	response.OK(c, subscriptionResponses)
}

// GetSubscription godoc
// @Summary [User] Get subscription details
// @Description Get details of a specific subscription
// @Tags User-Subscription
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Subscription ID"
// @Success 200 {object} entities.UserSubscriptionResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /subscriptions/{id} [get]
func (h *UserSubscriptionHandler) GetSubscription(c *gin.Context) {
	// Get current user from context
	userValue, exists := c.Get(middleware.AuthContextKey)
	if !exists {
		response.Unauthorized(c, "Authentication required")
		return
	}

	user, ok := userValue.(*userEntities.User)
	if !ok {
		response.Unauthorized(c, "Invalid user context")
		return
	}

	// Parse subscription ID
	subscriptionIDStr := c.Param("id")
	subscriptionID, err := strconv.ParseUint(subscriptionIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Subscription ID must be a valid number")
		return
	}

	// Get subscription
	subscription, err := h.userSubscriptionService.GetUserSubscription(c.Request.Context(), uint(subscriptionID))
	if err != nil {
		if err.Error() == "subscription not found" {
			response.NotFound(c, "Subscription not found")
			return
		}
		logger.Error("Failed to get subscription", logger.ErrorField(err), logger.Uint("subscription_id", uint(subscriptionID)))
		response.InternalServerError(c, "Failed to get subscription")
		return
	}

	// Check if user has access to this subscription
	if !user.IsAdmin() && subscription.UserID != user.ID {
		response.Forbidden(c, "You can only access your own subscriptions")
		return
	}

	response.OK(c, subscription.ToResponse())
}

// CancelSubscription godoc
// @Summary [User] Cancel subscription
// @Description Cancel a subscription
// @Tags User-Subscription
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Subscription ID"
// @Param request body CancelSubscriptionRequest true "Cancel request"
// @Success 200 {object} entities.UserSubscriptionResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /subscriptions/{id}/cancel [post]
func (h *UserSubscriptionHandler) CancelSubscription(c *gin.Context) {
	// Get current user from context
	userValue, exists := c.Get(middleware.AuthContextKey)
	if !exists {
		response.Unauthorized(c, "Authentication required")
		return
	}

	user, ok := userValue.(*userEntities.User)
	if !ok {
		response.Unauthorized(c, "Invalid user context")
		return
	}

	// Parse subscription ID
	subscriptionIDStr := c.Param("id")
	subscriptionID, err := strconv.ParseUint(subscriptionIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Subscription ID must be a valid number")
		return
	}

	// Bind request
	var req CancelSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request data")
		return
	}

	// Get subscription to check ownership
	subscription, err := h.userSubscriptionService.GetUserSubscription(c.Request.Context(), uint(subscriptionID))
	if err != nil {
		if err.Error() == "subscription not found" {
			response.NotFound(c, "Subscription not found")
			return
		}
		logger.Error("Failed to get subscription", logger.ErrorField(err), logger.Uint("subscription_id", uint(subscriptionID)))
		response.InternalServerError(c, "Failed to get subscription")
		return
	}

	// Check if user has access to cancel this subscription
	if !user.IsAdmin() && subscription.UserID != user.ID {
		response.Forbidden(c, "You can only cancel your own subscriptions")
		return
	}

	// Cancel subscription
	if err := h.userSubscriptionService.CancelUserSubscription(c.Request.Context(), uint(subscriptionID), req.Reason, req.CancelAtPeriodEnd); err != nil {
		logger.Error("Failed to cancel subscription", logger.ErrorField(err), logger.Uint("subscription_id", uint(subscriptionID)))
		response.InternalServerError(c, "Failed to cancel subscription")
		return
	}

	// Get updated subscription
	updatedSubscription, err := h.userSubscriptionService.GetUserSubscription(c.Request.Context(), uint(subscriptionID))
	if err != nil {
		logger.Error("Failed to get updated subscription", logger.ErrorField(err), logger.Uint("subscription_id", uint(subscriptionID)))
		response.InternalServerError(c, "Subscription cancelled but failed to get updated details")
		return
	}

	response.OK(c, updatedSubscription.ToResponse())
}

// GetSubscriptionTrafficStats godoc
// @Summary [User] Get subscription traffic statistics
// @Description Get traffic statistics for a subscription
// @Tags User-Subscription
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Subscription ID"
// @Success 200 {object} dto.TrafficStatsResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /subscriptions/{id}/traffic-stats [get]
func (h *UserSubscriptionHandler) GetSubscriptionTrafficStats(c *gin.Context) {
	// Get current user from context
	userValue, exists := c.Get(middleware.AuthContextKey)
	if !exists {
		response.Unauthorized(c, "Authentication required")
		return
	}

	user, ok := userValue.(*userEntities.User)
	if !ok {
		response.Unauthorized(c, "Invalid user context")
		return
	}

	// Parse subscription ID
	subscriptionIDStr := c.Param("id")
	subscriptionID, err := strconv.ParseUint(subscriptionIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Subscription ID must be a valid number")
		return
	}

	// Get subscription to check ownership
	subscription, err := h.userSubscriptionService.GetUserSubscription(c.Request.Context(), uint(subscriptionID))
	if err != nil {
		if err.Error() == "subscription not found" {
			response.NotFound(c, "Subscription not found")
			return
		}
		logger.Error("Failed to get subscription", logger.ErrorField(err), logger.Uint("subscription_id", uint(subscriptionID)))
		response.InternalServerError(c, "Failed to get subscription")
		return
	}

	// Check if user has access to this subscription
	if !user.IsAdmin() && subscription.UserID != user.ID {
		response.Forbidden(c, "You can only access your own subscription traffic stats")
		return
	}

	// Get traffic statistics
	stats, err := h.userSubscriptionService.GetSubscriptionTrafficStats(c.Request.Context(), uint(subscriptionID))
	if err != nil {
		logger.Error("Failed to get traffic stats", logger.ErrorField(err), logger.Uint("subscription_id", uint(subscriptionID)))
		response.InternalServerError(c, "Failed to get traffic statistics")
		return
	}

	response.OK(c, stats)
}

// CancelSubscriptionRequest represents the request to cancel a subscription
type CancelSubscriptionRequest struct {
	Reason            string `json:"reason" binding:"required,min=1,max=255" example:"No longer needed"`
	CancelAtPeriodEnd bool   `json:"cancel_at_period_end" example:"true"`
}


// authServiceAdapter adapts the domain AuthService to middleware AuthService interface
type authServiceAdapter struct {
	authService authInterfaces.AuthService
}

func (a *authServiceAdapter) ValidateToken(token string) (any, error) {
	user, err := a.authService.ValidateToken(token)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// RegisterRoutes registers all user subscription routes
func (h *UserSubscriptionHandler) RegisterRoutes(router *gin.RouterGroup) {
	// Create auth service adapter
	authAdapter := &authServiceAdapter{authService: h.authService}

	// User subscription routes - accessible to authenticated users
	subscriptionGroup := router.Group("/subscriptions")
	{
		// Public endpoint for testing (should be protected in production)
		subscriptionGroup.GET("", func(c *gin.Context) {
			response.Error(c, http.StatusUnauthorized, 4001, "User not authenticated - please login to access subscriptions")
		})

		// Protected routes requiring authentication
		subscriptionGroup.GET("/my", middleware.AuthMiddleware(authAdapter), h.GetMySubscriptions)
		subscriptionGroup.GET("/my/active", middleware.AuthMiddleware(authAdapter), h.GetMyActiveSubscriptions)
		subscriptionGroup.GET("/:id", middleware.AuthMiddleware(authAdapter), h.GetSubscription)
		subscriptionGroup.POST("/:id/cancel", middleware.AuthMiddleware(authAdapter), h.CancelSubscription)
		subscriptionGroup.GET("/:id/traffic-stats", middleware.AuthMiddleware(authAdapter), h.GetSubscriptionTrafficStats)

		// Subscription config export (Clash-compatible YAML)
		subscriptionGroup.GET("/clash", h.GetClashConfig)
	}

}

// GetClashConfig godoc
// @Summary Get Clash subscription YAML
// @Description Return Clash-compatible YAML built from user's accessible servers
// @Tags User-Subscription
// @Produce text/plain
// @Security BearerAuth
// @Param token query string false "Subscription token (Bearer token also supported)"
// @Success 200 {string} string "YAML"
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /subscriptions/clash [get]
func (h *UserSubscriptionHandler) GetClashConfig(c *gin.Context) {
	// Auth: prefer Bearer; fallback to token query
	var user *userEntities.User
	if authz := c.GetHeader("Authorization"); strings.HasPrefix(strings.ToLower(authz), "bearer ") {
		// Validate token via auth service
		u, err := h.authService.ValidateToken(strings.TrimSpace(authz[7:]))
		if err == nil {
			user = u
		}
	}
	// Fallback: token query param as JWT
	if user == nil {
		if qToken := strings.TrimSpace(c.Query("token")); qToken != "" {
			if u, err := h.authService.ValidateToken(qToken); err == nil {
				user = u
			}
		}
	}
	if user == nil {
		response.Unauthorized(c, "Authentication required")
		return
	}

	// Fetch user's active subscriptions to get subscription UUID
	activeSubs, err := h.userSubscriptionService.GetUserActiveSubscriptions(c.Request.Context(), user.ID)
	if err != nil || len(activeSubs) == 0 {
		response.Unauthorized(c, "No active subscription found")
		return
	}
	sub := activeSubs[0]

	// Fetch accessible servers by user's active subscriptions
	servers, err := h.userSubscriptionService.GetUserAccessibleServers(c.Request.Context(), user.ID)
	if err != nil {
		response.InternalServerError(c, "Failed to fetch accessible servers")
		return
	}

	// Build minimal Clash YAML
	yamlStr, err := BuildMinimalClashYAML(user, sub.UUID, servers)
	if err != nil {
		response.InternalServerError(c, "Failed to build YAML")
		return
	}

	c.Header("Content-Type", "text/yaml")
	// subscription-userinfo header (upload, download, total, expire)
	var expire int64 = 0
	if sub.EndDate != nil {
		expire = sub.EndDate.Unix()
	}
	// Note: upload/download values are placeholders as we track aggregate usage
	c.Header("subscription-userinfo",
		"upload="+strconv.FormatInt(sub.TrafficUsed, 10)+"; download="+strconv.FormatInt(0, 10)+"; total="+strconv.FormatInt(sub.TrafficLimit, 10)+"; expire="+strconv.FormatInt(expire, 10))
	c.Header("profile-update-interval", "24")
	c.String(200, yamlStr)
}

// BuildMinimalClashYAML builds a minimal Clash YAML from servers
func BuildMinimalClashYAML(user *userEntities.User, subscriptionUUID string, servers []*serverEntities.ShadowsocksServer) (string, error) {
	type proxy struct {
		Name     string `yaml:"name"`
		Type     string `yaml:"type"`
		Server   string `yaml:"server"`
		Port     int    `yaml:"port"`
		Cipher   string `yaml:"cipher"`
		Password string `yaml:"password"`
		UDP      bool   `yaml:"udp"`
	}
	cfg := map[string]any{
		"port":       7890,
		"socks-port": 7891,
		"allow-lan":  true,
		"mode":       "Rule",
		"log-level":  "info",
		"proxies":    []proxy{},
		"proxy-groups": []map[string]any{
			{
				"name":     "Auto",
				"type":     "url-test",
				"proxies":  []string{},
				"url":      "http://www.gstatic.com/generate_204",
				"interval": 300,
			},
		},
		"rules": []string{"MATCH,Auto"},
	}
	var proxies []proxy
	var names []string
	for _, s := range servers {
		// Only shadowsocks supported here
		p := proxy{
			Name:     s.Name,
			Type:     "ss",
			Server:   s.Host,
			Port:     s.Port,
			Cipher:   s.Cipher,
			Password: subscriptionUUID,
			UDP:      true,
		}
		proxies = append(proxies, p)
		names = append(names, s.Name)
	}
	cfg["proxies"] = proxies
	// fill group proxies
	groups := cfg["proxy-groups"].([]map[string]any)
	groups[0]["proxies"] = names
	cfg["proxy-groups"] = groups
	out, err := yaml.Marshal(cfg)
	if err != nil {
		return "", err
	}
	return string(out), nil
}
