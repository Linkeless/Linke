package handlers

import (
	"context"
	"encoding/json"
	"io"
	"strconv"
	"time"

	serverEntities "linke/internal/domains/server/entities"
	serverInterfaces "linke/internal/domains/server/usecases/interfaces"
	subscriptionEntities "linke/internal/domains/subscription/entities"
	subscriptionInterfaces "linke/internal/domains/subscription/usecases/interfaces"
	userEntities "linke/internal/domains/user/entities"
	userInterfaces "linke/internal/domains/user/usecases/interfaces"
	"linke/internal/shared/config"
	"linke/internal/shared/database"
	"linke/internal/shared/logger"
	"linke/internal/shared/response"

	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
)

type ServerAPIHandler struct {
	shadowsocksService      serverInterfaces.ShadowsocksServerService
	userService             userInterfaces.UserService
	userSubscriptionService subscriptionInterfaces.UserSubscriptionService
	db                      *database.Database
	asynqClient             *asynq.Client
	config                  *config.Config
}

func NewServerAPIHandler(shadowsocksService serverInterfaces.ShadowsocksServerService, userService userInterfaces.UserService, userSubscriptionService subscriptionInterfaces.UserSubscriptionService, db *database.Database, asynqClient *asynq.Client, cfg *config.Config) *ServerAPIHandler {
	return &ServerAPIHandler{
		shadowsocksService:      shadowsocksService,
		userService:             userService,
		userSubscriptionService: userSubscriptionService,
		db:                      db,
		asynqClient:             asynqClient,
		config:                  cfg,
	}
}

// Health check endpoint for server API
// @Summary Server API Health Check
// @Description Health check endpoint for server API
// @Tags Server-API
// @Accept json
// @Produce json
// @Success 200 {object} response.StandardResponse{data=dto.ServerAPIHealthResponse}
// @Router /server/UniProxy/health [get]
func (h *ServerAPIHandler) Health(c *gin.Context) {
	response.Success(c, gin.H{
		"status":  "ok",
		"service": "server-api",
	})
}

// UniProxyConfigResponse represents the config response for UniProxy
type UniProxyConfigResponse struct {
	ServerPort   int                `json:"server_port"`
	Cipher       string             `json:"cipher"`
	Obfs         interface{}        `json:"obfs" swaggertype:"string" example:"tls1.2_ticket_auth"`
	ObfsSettings interface{}        `json:"obfs_settings" swaggertype:"string" example:"cloudflare.com"`
	BaseConfig   UniProxyBaseConfig `json:"base_config"`
}

// UniProxyBaseConfig represents the base configuration for UniProxy
type UniProxyBaseConfig struct {
	PushInterval int `json:"push_interval"`
	PullInterval int `json:"pull_interval"`
}

// UniProxy config endpoint
// @Summary Get UniProxy Server Config
// @Description Get configuration for UniProxy server based on node_id and node_type
// @Tags Server-API
// @Accept json
// @Produce json
// @Param node_id query int true "Node ID"
// @Param node_type query string true "Node Type (shadowsocks)"
// @Param token query string true "Authentication Token"
// @Success 200 {object} response.StandardResponse{data=UniProxyConfigResponse}
// @Failure 400 {object} response.StandardResponse
// @Failure 401 {object} response.StandardResponse
// @Failure 404 {object} response.StandardResponse
// @Router /server/UniProxy/config [get]
func (h *ServerAPIHandler) UniProxyConfig(c *gin.Context) {
	// Get query parameters
	nodeIDStr := c.Query("node_id")
	nodeType := c.Query("node_type")
	token := c.Query("token")

	// Validate required parameters
	if nodeIDStr == "" {
		logger.Warn("Missing node_id parameter", logger.String("remote_addr", c.ClientIP()))
		response.BadRequest(c, "node_id parameter is required")
		return
	}
	if nodeType == "" {
		logger.Warn("Missing node_type parameter", logger.String("remote_addr", c.ClientIP()))
		response.BadRequest(c, "node_type parameter is required")
		return
	}
	if token == "" {
		logger.Warn("Missing token parameter", logger.String("remote_addr", c.ClientIP()))
		response.Unauthorized(c, "token parameter is required")
		return
	}

	// Parse node_id
	nodeID, err := strconv.Atoi(nodeIDStr)
	if err != nil {
		logger.Warn("Invalid node_id parameter",
			logger.String("node_id", nodeIDStr),
			logger.String("remote_addr", c.ClientIP()),
		)
		response.BadRequest(c, "node_id must be a valid integer")
		return
	}

	// Validate node_type (for now only support shadowsocks)
	if nodeType != "shadowsocks" {
		logger.Warn("Unsupported node_type",
			logger.String("node_type", nodeType),
			logger.String("remote_addr", c.ClientIP()),
		)
		response.BadRequest(c, "only shadowsocks node_type is supported")
		return
	}

	// Simple token validation - you may want to implement more sophisticated validation
	if !h.validateServerToken(token) {
		logger.Warn("Invalid server token",
			logger.String("token", token),
			logger.String("remote_addr", c.ClientIP()),
		)
		response.Unauthorized(c, "invalid authentication token")
		return
	}

	// Get shadowsocks server by ID
	server, err := h.shadowsocksService.GetShadowsocksServerByID(context.Background(), nodeID)
	if err != nil {
		logger.Error("Failed to get shadowsocks server",
			logger.Int("node_id", nodeID),
			logger.String("remote_addr", c.ClientIP()),
			logger.Error2("error", err),
		)
		response.NotFound(c, "shadowsocks server not found")
		return
	}

	// Check if server is visible/active
	if !server.IsVisible() {
		logger.Warn("Server not visible",
			logger.Int("node_id", nodeID),
			logger.String("remote_addr", c.ClientIP()),
		)
		response.NotFound(c, "shadowsocks server not available")
		return
	}

	// Build config response
	configResponse := &UniProxyConfigResponse{
		ServerPort:   server.ServerPort,
		Cipher:       server.Cipher,
		Obfs:         getObfsValue(server.Obfs),
		ObfsSettings: getObfsSettingsValue(server.ObfsSettings),
		BaseConfig: UniProxyBaseConfig{
			PushInterval: 60,
			PullInterval: 60,
		},
	}

	logger.Info("UniProxy config retrieved successfully",
		logger.Int("node_id", nodeID),
		logger.String("node_type", nodeType),
		logger.String("server_name", server.Name),
		logger.String("remote_addr", c.ClientIP()),
	)

	c.JSON(200, configResponse)
}

// UniProxyUserItem represents a single user item for UniProxy
type UniProxyUserItem struct {
	ID         uint        `json:"id"`          // Subscription ID
	UUID       string      `json:"uuid"`        // Subscription UUID
	SpeedLimit interface{} `json:"speed_limit" swaggertype:"integer" example:"100"` // Speed limit (null for unlimited)
}

// UniProxyUsersResponse represents the users response for UniProxy
type UniProxyUsersResponse struct {
	Users []UniProxyUserItem `json:"users"`
}

// UniProxy user endpoint
// @Summary Get UniProxy Users
// @Description Get users with active subscriptions that have access to the specified shadowsocks server. Returns all valid subscriptions, including multiple subscriptions for the same user.
// @Tags Server-API
// @Accept json
// @Produce json
// @Param node_id query int true "Node ID"
// @Param node_type query string true "Node Type (shadowsocks)"
// @Param token query string true "Authentication Token"
// @Success 200 {object} UniProxyUsersResponse
// @Failure 400 {object} response.StandardResponse
// @Failure 401 {object} response.StandardResponse
// @Failure 404 {object} response.StandardResponse
// @Router /server/UniProxy/user [get]
func (h *ServerAPIHandler) UniProxyUsers(c *gin.Context) {
	// Get query parameters
	nodeIDStr := c.Query("node_id")
	nodeType := c.Query("node_type")
	token := c.Query("token")

	// Validate required parameters
	if nodeIDStr == "" {
		logger.Warn("Missing node_id parameter", logger.String("remote_addr", c.ClientIP()))
		response.BadRequest(c, "node_id parameter is required")
		return
	}
	if nodeType == "" {
		logger.Warn("Missing node_type parameter", logger.String("remote_addr", c.ClientIP()))
		response.BadRequest(c, "node_type parameter is required")
		return
	}
	if token == "" {
		logger.Warn("Missing token parameter", logger.String("remote_addr", c.ClientIP()))
		response.Unauthorized(c, "token parameter is required")
		return
	}

	// Parse node_id
	nodeID, err := strconv.Atoi(nodeIDStr)
	if err != nil {
		logger.Warn("Invalid node_id parameter",
			logger.String("node_id", nodeIDStr),
			logger.String("remote_addr", c.ClientIP()),
		)
		response.BadRequest(c, "node_id must be a valid integer")
		return
	}

	// Validate node_type (for now only support shadowsocks)
	if nodeType != "shadowsocks" {
		logger.Warn("Unsupported node_type",
			logger.String("node_type", nodeType),
			logger.String("remote_addr", c.ClientIP()),
		)
		response.BadRequest(c, "only shadowsocks node_type is supported")
		return
	}

	// Simple token validation
	if !h.validateServerToken(token) {
		logger.Warn("Invalid server token",
			logger.String("token", token),
			logger.String("remote_addr", c.ClientIP()),
		)
		response.Unauthorized(c, "invalid authentication token")
		return
	}

	// Get shadowsocks server by ID to verify it exists and is active
	server, err := h.shadowsocksService.GetShadowsocksServerByID(context.Background(), nodeID)
	if err != nil {
		logger.Error("Failed to get shadowsocks server",
			logger.Int("node_id", nodeID),
			logger.String("remote_addr", c.ClientIP()),
			logger.Error2("error", err),
		)
		response.NotFound(c, "shadowsocks server not found")
		return
	}

	// Check if server is visible/active
	if !server.IsVisible() {
		logger.Warn("Server not visible",
			logger.Int("node_id", nodeID),
			logger.String("remote_addr", c.ClientIP()),
		)
		response.NotFound(c, "shadowsocks server not available")
		return
	}

	// Get users with active subscriptions that have access to this server
	users, err := h.getUsersForServer(context.Background(), server)
	if err != nil {
		logger.Error("Failed to get users for server",
			logger.Int("node_id", nodeID),
			logger.String("remote_addr", c.ClientIP()),
			logger.Error2("error", err),
		)
		response.InternalServerError(c, "failed to get users")
		return
	}

	logger.Info("UniProxy users retrieved successfully",
		logger.Int("node_id", nodeID),
		logger.String("node_type", nodeType),
		logger.Int("user_count", len(users.Users)),
		logger.String("remote_addr", c.ClientIP()),
	)

	c.JSON(200, users)
}

// getUsersForServer gets all users with active subscriptions that have access to the specified server
func (h *ServerAPIHandler) getUsersForServer(ctx context.Context, server *serverEntities.ShadowsocksServer) (*UniProxyUsersResponse, error) {
	var users []UniProxyUserItem

	// Query to get active users with active subscriptions that have access to this server's group
	// SECURITY: Filter out traffic-suspended subscriptions to enforce traffic limits
	var userSubscriptions []subscriptionEntities.UserSubscription
	if err := h.db.DB.WithContext(ctx).
		Table("user_subscriptions").
		Select("id, user_id, uuid, server_group_ids, status, traffic_limit, traffic_used, traffic_suspended, trial_end_date").
		Where("status IN ?", []string{subscriptionEntities.UserSubscriptionStatusActive, subscriptionEntities.UserSubscriptionStatusTrial}).
		Where("deleted_at IS NULL").
		Where("end_date IS NULL OR end_date > NOW()"). // Use database NOW() instead of Go time.Now()
		Where("traffic_suspended = ?", false).
		Where("trial_end_date IS NULL OR trial_end_date > NOW()"). // Check trial period
		Find(&userSubscriptions).Error; err != nil {
		logger.Error("Failed to query user subscriptions", logger.Error2("error", err))
		return nil, err
	}

	// Filter subscriptions that have access to the server's group and have sufficient traffic
	var validSubscriptions []subscriptionEntities.UserSubscription
	var validUserIDs []uint
	userIDSet := make(map[uint]bool) // To avoid duplicate user IDs

	for _, subscription := range userSubscriptions {
		// Check traffic limits (additional safety check beyond traffic_suspended)
		if subscription.TrafficLimit > 0 && subscription.TrafficUsed >= subscription.TrafficLimit {
			logger.Debug("Subscription over traffic limit",
				logger.Uint("subscription_id", subscription.ID),
				logger.Int64("limit", subscription.TrafficLimit),
				logger.Int64("used", subscription.TrafficUsed))
			continue
		}

		// Check if subscription has access to the server's group
		if h.hasAccessToServerGroup(&subscription, server.GroupID) {
			validSubscriptions = append(validSubscriptions, subscription)
			// Only add user ID to slice if not already present
			if !userIDSet[subscription.UserID] {
				validUserIDs = append(validUserIDs, subscription.UserID)
				userIDSet[subscription.UserID] = true
			}
		}
	}

	if len(validSubscriptions) == 0 {
		return &UniProxyUsersResponse{Users: users}, nil
	}

	// Get active users
	activeUserSet := make(map[uint]bool)
	var activeUsers []userEntities.User
	if err := h.db.DB.WithContext(ctx).
		Select("id").
		Where("id IN ? AND status = ? AND deleted_at IS NULL", validUserIDs, userEntities.UserStatusActive).
		Find(&activeUsers).Error; err != nil {
		logger.Error("Failed to query active users", logger.Error2("error", err))
		return nil, err
	}

	// Create a set of active user IDs for quick lookup
	for _, user := range activeUsers {
		activeUserSet[user.ID] = true
	}

	// Convert to response format - include all valid subscriptions for active users
	for _, subscription := range validSubscriptions {
		// Only include subscriptions for active users
		if activeUserSet[subscription.UserID] {
			userItem := UniProxyUserItem{
				ID:         subscription.ID, // Use subscription ID
				UUID:       subscription.UUID,
				SpeedLimit: nil, // Set to null as specified
			}
			users = append(users, userItem)
		}
	}

	return &UniProxyUsersResponse{Users: users}, nil
}

// validateServerToken performs secure token validation
func (h *ServerAPIHandler) validateServerToken(token string) bool {
	// SECURITY: Check if server token is configured
	if h.config.API.ServerToken == "" {
		logger.Error("Server API token not configured, rejecting all requests")
		return false
	}

	// SECURITY: Check minimum token length for security
	if len(h.config.API.ServerToken) < 20 {
		logger.Error("Server API token too short, must be at least 20 characters")
		return false
	}

	// SECURITY: Use constant-time comparison to prevent timing attacks
	if len(token) != len(h.config.API.ServerToken) {
		return false
	}

	// Constant-time comparison
	var result byte
	for i := 0; i < len(token); i++ {
		result |= token[i] ^ h.config.API.ServerToken[i]
	}

	if result != 0 {
		logger.Warn("Invalid server API token attempted",
			logger.String("client_ip", "unknown"), // Add IP logging in middleware
			logger.Int("token_length", len(token)))
		return false
	}

	return true
}

// hasAccessToServerGroup checks if a user subscription has access to a specific server group
func (h *ServerAPIHandler) hasAccessToServerGroup(subscription *subscriptionEntities.UserSubscription, groupID uint) bool {
	// SECURITY: If no server group IDs are specified, deny access by default
	// This prevents accidental access to all groups due to misconfiguration
	if subscription.ServerGroupIDs == "" {
		logger.Debug("Subscription has no server group access configured",
			logger.Uint("subscription_id", subscription.ID))
		return false
	}

	// Parse the server group IDs JSON
	var groupIDs []uint
	if err := json.Unmarshal([]byte(subscription.ServerGroupIDs), &groupIDs); err != nil {
		logger.Error("Failed to parse server group IDs",
			logger.Uint("subscription_id", subscription.ID),
			logger.String("server_group_ids", subscription.ServerGroupIDs),
			logger.Error2("error", err),
		)
		return false
	}

	// Special case: if groupIDs contains 0, it means access to all groups
	// This must be explicitly configured, not just an empty field
	for _, id := range groupIDs {
		if id == 0 { // 0 represents "all groups" access
			logger.Debug("Subscription has access to all server groups",
				logger.Uint("subscription_id", subscription.ID))
			return true
		}
	}

	// Check if the specific group ID is in the list
	for _, id := range groupIDs {
		if id == groupID {
			return true
		}
	}

	return false
}

// getObfsValue returns null if obfs is empty, otherwise returns the obfs value
func getObfsValue(obfs string) any {
	if obfs == "" {
		return nil
	}
	return obfs
}

// getObfsSettingsValue returns null if obfs_settings is empty, otherwise returns the obfs_settings value
func getObfsSettingsValue(obfsSettings string) any {
	if obfsSettings == "" {
		return nil
	}
	return obfsSettings
}

// UniProxyPushRequest represents the push data from UniProxy nodes
type UniProxyPushRequest struct {
	NodeID   uint   `form:"node_id" binding:"required"`
	NodeType string `form:"node_type" binding:"required"`
	Token    string `form:"token" binding:"required"`
}

// UniProxyPush handles data push from UniProxy nodes
// @Summary UniProxy Node Data Push
// @Description Receive and log data push from UniProxy nodes (shadowsocks, etc.)
// @Tags Server-API
// @Accept json
// @Produce json
// @Param node_id query int true "Node ID"
// @Param node_type query string true "Node Type" Enums(shadowsocks)
// @Param token query string true "Authentication Token"
// @Success 200 {object} response.StandardResponse{data=dto.UniProxyPushResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /server/UniProxy/push [post]
func (h *ServerAPIHandler) UniProxyPush(c *gin.Context) {
	// Parse query parameters
	var req UniProxyPushRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		logger.Warn("Invalid UniProxy push parameters",
			logger.String("client_ip", c.ClientIP()),
			logger.Error2("error", err),
		)
		response.BadRequest(c, "Invalid parameters", err.Error())
		return
	}

	// Read request body
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		logger.Error("Failed to read UniProxy push request body",
			logger.String("client_ip", c.ClientIP()),
			logger.Uint("node_id", req.NodeID),
			logger.String("node_type", req.NodeType),
			logger.Error2("error", err),
		)
		response.InternalServerError(c, "Failed to read request body")
		return
	}

	requestBody := string(bodyBytes)

	// Log the push request with all details
	logger.Info("UniProxy push request received",
		logger.String("client_ip", c.ClientIP()),
		logger.Uint("node_id", req.NodeID),
		logger.String("node_type", req.NodeType),
		logger.String("token", req.Token),
		logger.String("request_body", requestBody),
		logger.String("content_type", c.GetHeader("Content-Type")),
		logger.String("user_agent", c.GetHeader("User-Agent")),
	)

	// TODO: Add token validation logic here
	// For now, we accept all requests for logging purposes

	// Process the node data (simplified implementation)
	logger.Info("UniProxy push request processed",
		logger.Uint("node_id", req.NodeID),
		logger.String("node_type", req.NodeType),
	)

	response.OK(c, "Node data received and processed", gin.H{
		"status":    "processed",
		"node_id":   req.NodeID,
		"node_type": req.NodeType,
		"timestamp": time.Now().Unix(),
	})
}
