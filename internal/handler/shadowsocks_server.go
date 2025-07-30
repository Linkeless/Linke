package handler

import (
	"strconv"

	"linke/internal/response"
	"linke/internal/service"

	"github.com/gin-gonic/gin"
)

type ShadowsocksServerHandler struct {
	shadowsocksServerService *service.ShadowsocksServerService
	userService              *service.UserService
	subscriptionService      *service.UserSubscriptionService
}

func NewShadowsocksServerHandler(
	shadowsocksServerService *service.ShadowsocksServerService,
	_ interface{}, // Placeholder for removed ServerGroupService
	userService *service.UserService,
	subscriptionService *service.UserSubscriptionService,
) *ShadowsocksServerHandler {
	return &ShadowsocksServerHandler{
		shadowsocksServerService: shadowsocksServerService,
		userService:              userService,
		subscriptionService:      subscriptionService,
	}
}

// CreateShadowsocksServer creates a new shadowsocks server (Admin only)
// @Summary Create Shadowsocks Server
// @Description Create a new shadowsocks server (Admin only)
// @Tags Admin-ShadowsocksServers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body service.CreateShadowsocksServerRequest true "Create shadowsocks server request"
// @Success 201 {object} response.StandardResponse{data=model.ShadowsocksServerResponse}
// @Failure 400 {object} response.StandardResponse
// @Failure 401 {object} response.StandardResponse
// @Failure 403 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /admin/shadowsocks-servers [post]
func (h *ShadowsocksServerHandler) CreateShadowsocksServer(c *gin.Context) {
	var req service.CreateShadowsocksServerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request", err.Error())
		return
	}

	server, err := h.shadowsocksServerService.CreateShadowsocksServer(c.Request.Context(), &req)
	if err != nil {
		response.InternalServerError(c, "Failed to create shadowsocks server", err.Error())
		return
	}

	response.CreatedWithMessage(c, "Shadowsocks server created successfully", server.ToResponse())
}

// GetShadowsocksServers retrieves shadowsocks servers with optional filters (Admin only)
// @Summary Get Shadowsocks Servers
// @Description Get shadowsocks servers with optional filters (Admin only)
// @Tags Admin-ShadowsocksServers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param group_id query string false "Group ID filter"
// @Param status query string false "Status filter"
// @Param is_show query bool false "Show filter"
// @Param is_online query bool false "Online filter"
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Success 200 {object} response.StandardResponse{data=response.PaginatedResponse}
// @Failure 400 {object} response.StandardResponse
// @Failure 401 {object} response.StandardResponse
// @Failure 403 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /admin/shadowsocks-servers [get]
func (h *ShadowsocksServerHandler) GetShadowsocksServers(c *gin.Context) {
	req := &service.GetShadowsocksServersRequest{}
	
	// Parse group_id parameter
	if groupIDStr := c.Query("group_id"); groupIDStr != "" {
		if groupID, err := strconv.ParseUint(groupIDStr, 10, 32); err == nil {
			groupIDUint := uint(groupID)
			req.GroupID = &groupIDUint
		}
	}

	// Parse show parameter
	if showStr := c.Query("show"); showStr != "" {
		if show, err := strconv.Atoi(showStr); err == nil && (show == 0 || show == 1) {
			req.Show = &show
		}
	}

	// Parse pagination parameters
	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			req.Limit = limit
		}
	}
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil {
			req.Offset = offset
		}
	}

	servers, total, err := h.shadowsocksServerService.GetShadowsocksServers(c.Request.Context(), req)
	if err != nil {
		response.InternalServerError(c, "Failed to get shadowsocks servers", err.Error())
		return
	}

	// Convert to response format
	serverResponses := make([]interface{}, 0, len(servers))
	for _, server := range servers {
		serverResponses = append(serverResponses, server.ToResponse())
	}

	response.OKPaginated(c, "Shadowsocks servers retrieved successfully", serverResponses, total, req.Limit, req.Offset)
}

// GetShadowsocksServerByID retrieves a shadowsocks server by ID (Admin only)
// @Summary Get Shadowsocks Server by ID
// @Description Get a shadowsocks server by ID (Admin only)
// @Tags Admin-ShadowsocksServers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Shadowsocks Server ID"
// @Success 200 {object} response.StandardResponse{data=model.ShadowsocksServerResponse}
// @Failure 400 {object} response.StandardResponse
// @Failure 401 {object} response.StandardResponse
// @Failure 403 {object} response.StandardResponse
// @Failure 404 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /admin/shadowsocks-servers/{id} [get]
func (h *ShadowsocksServerHandler) GetShadowsocksServerByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.BadRequest(c, "Invalid server ID", "Server ID must be a valid number")
		return
	}

	server, err := h.shadowsocksServerService.GetShadowsocksServerByID(c.Request.Context(), id)
	if err != nil {
		if err.Error() == "shadowsocks server not found" {
			response.NotFound(c, "Shadowsocks server not found")
			return
		}
		response.InternalServerError(c, "Failed to get shadowsocks server", err.Error())
		return
	}

	response.SuccessWithMessage(c, "Shadowsocks server retrieved successfully", server.ToResponse())
}

// UpdateShadowsocksServer updates a shadowsocks server (Admin only)
// @Summary Update Shadowsocks Server
// @Description Update a shadowsocks server (Admin only)
// @Tags Admin-ShadowsocksServers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Shadowsocks Server ID"
// @Param request body service.UpdateShadowsocksServerRequest true "Update shadowsocks server request"
// @Success 200 {object} response.StandardResponse{data=model.ShadowsocksServerResponse}
// @Failure 400 {object} response.StandardResponse
// @Failure 401 {object} response.StandardResponse
// @Failure 403 {object} response.StandardResponse
// @Failure 404 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /admin/shadowsocks-servers/{id} [put]
func (h *ShadowsocksServerHandler) UpdateShadowsocksServer(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.BadRequest(c, "Invalid server ID", "Server ID must be a valid number")
		return
	}

	var req service.UpdateShadowsocksServerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request", err.Error())
		return
	}

	server, err := h.shadowsocksServerService.UpdateShadowsocksServer(c.Request.Context(), id, &req)
	if err != nil {
		if err.Error() == "shadowsocks server not found" {
			response.NotFound(c, "Shadowsocks server not found")
			return
		}
		response.InternalServerError(c, "Failed to update shadowsocks server", err.Error())
		return
	}

	response.SuccessWithMessage(c, "Shadowsocks server updated successfully", server.ToResponse())
}

// PatchShadowsocksServer partially updates a shadowsocks server (Admin only)
// @Summary Partially Update Shadowsocks Server
// @Description Partially update a shadowsocks server using PATCH method (Admin only)
// @Tags Admin-ShadowsocksServers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Shadowsocks Server ID"
// @Param request body map[string]interface{} true "Partial shadowsocks server update data"
// @Success 200 {object} response.StandardResponse{data=model.ShadowsocksServerResponse}
// @Failure 400 {object} response.StandardResponse
// @Failure 401 {object} response.StandardResponse
// @Failure 403 {object} response.StandardResponse
// @Failure 404 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /admin/shadowsocks-servers/{id} [patch]
func (h *ShadowsocksServerHandler) PatchShadowsocksServer(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.BadRequest(c, "Invalid server ID", "Server ID must be a valid number")
		return
	}

	// Get current server to verify it exists
	_, err = h.shadowsocksServerService.GetShadowsocksServerByID(c.Request.Context(), id)
	if err != nil {
		if err.Error() == "shadowsocks server not found" {
			response.NotFound(c, "Shadowsocks server not found")
			return
		}
		response.InternalServerError(c, "Failed to get shadowsocks server", err.Error())
		return
	}

	// Parse partial update data
	var updateData map[string]interface{}
	if err := c.ShouldBindJSON(&updateData); err != nil {
		response.BadRequest(c, "Invalid request", err.Error())
		return
	}

	// Create update request with only provided fields
	req := &service.UpdateShadowsocksServerRequest{}

	// Apply partial updates
	if groupID, exists := updateData["group_id"]; exists {
		if groupIDFloat, ok := groupID.(float64); ok {
			groupIDUint := uint(groupIDFloat)
			req.GroupID = &groupIDUint
		} else {
			response.BadRequest(c, "Invalid group_id field type")
			return
		}
	}

	if routeID, exists := updateData["route_id"]; exists {
		if routeIDStr, ok := routeID.(string); ok {
			req.RouteID = &routeIDStr
		} else {
			response.BadRequest(c, "Invalid route_id field type")
			return
		}
	}

	if parentID, exists := updateData["parent_id"]; exists {
		if parentIDFloat, ok := parentID.(float64); ok {
			parentIDInt := int(parentIDFloat)
			req.ParentID = &parentIDInt
		} else {
			response.BadRequest(c, "Invalid parent_id field type")
			return
		}
	}

	if name, exists := updateData["name"]; exists {
		if nameStr, ok := name.(string); ok {
			req.Name = &nameStr
		} else {
			response.BadRequest(c, "Invalid name field type")
			return
		}
	}

	if tags, exists := updateData["tags"]; exists {
		if tagsStr, ok := tags.(string); ok {
			req.Tags = &tagsStr
		} else {
			response.BadRequest(c, "Invalid tags field type")
			return
		}
	}

	if host, exists := updateData["host"]; exists {
		if hostStr, ok := host.(string); ok {
			req.Host = &hostStr
		} else {
			response.BadRequest(c, "Invalid host field type")
			return
		}
	}

	if port, exists := updateData["port"]; exists {
		if portFloat, ok := port.(float64); ok {
			portInt := int(portFloat)
			req.Port = &portInt
		} else {
			response.BadRequest(c, "Invalid port field type")
			return
		}
	}

	if serverPort, exists := updateData["server_port"]; exists {
		if serverPortFloat, ok := serverPort.(float64); ok {
			serverPortInt := int(serverPortFloat)
			if serverPortInt < 1 || serverPortInt > 65535 {
				response.BadRequest(c, "Invalid server_port value, must be between 1 and 65535")
				return
			}
			req.ServerPort = &serverPortInt
		} else {
			response.BadRequest(c, "Invalid server_port field type")
			return
		}
	}

	if cipher, exists := updateData["cipher"]; exists {
		if cipherStr, ok := cipher.(string); ok {
			req.Cipher = &cipherStr
		} else {
			response.BadRequest(c, "Invalid cipher field type")
			return
		}
	}

	if obfs, exists := updateData["obfs"]; exists {
		if obfsStr, ok := obfs.(string); ok {
			req.Obfs = &obfsStr
		} else {
			response.BadRequest(c, "Invalid obfs field type")
			return
		}
	}

	if obfsSettings, exists := updateData["obfs_settings"]; exists {
		if obfsSettingsStr, ok := obfsSettings.(string); ok {
			req.ObfsSettings = &obfsSettingsStr
		} else {
			response.BadRequest(c, "Invalid obfs_settings field type")
			return
		}
	}

	if excludes, exists := updateData["excludes"]; exists {
		if excludesStr, ok := excludes.(string); ok {
			req.Excludes = &excludesStr
		} else {
			response.BadRequest(c, "Invalid excludes field type")
			return
		}
	}

	if ips, exists := updateData["ips"]; exists {
		if ipsStr, ok := ips.(string); ok {
			req.IPs = &ipsStr
		} else {
			response.BadRequest(c, "Invalid ips field type")
			return
		}
	}

	if rate, exists := updateData["rate"]; exists {
		if rateFloat, ok := rate.(float64); ok {
			req.Rate = &rateFloat
		} else {
			response.BadRequest(c, "Invalid rate field type")
			return
		}
	}

	if show, exists := updateData["show"]; exists {
		if showFloat, ok := show.(float64); ok {
			showInt := int(showFloat)
			if showInt != 0 && showInt != 1 {
				response.BadRequest(c, "Invalid show value, must be 0 or 1")
				return
			}
			req.Show = &showInt
		} else {
			response.BadRequest(c, "Invalid show field type")
			return
		}
	}

	if sort, exists := updateData["sort"]; exists {
		if sortFloat, ok := sort.(float64); ok {
			sortInt := int(sortFloat)
			req.Sort = &sortInt
		} else {
			response.BadRequest(c, "Invalid sort field type")
			return
		}
	}

	// Update the server
	server, err := h.shadowsocksServerService.UpdateShadowsocksServer(c.Request.Context(), id, req)
	if err != nil {
		if err.Error() == "shadowsocks server not found" {
			response.NotFound(c, "Shadowsocks server not found")
			return
		}
		response.InternalServerError(c, "Failed to update shadowsocks server", err.Error())
		return
	}

	response.SuccessWithMessage(c, "Shadowsocks server updated successfully", server.ToResponse())
}

// DeleteShadowsocksServer deletes a shadowsocks server (Admin only)
// @Summary Delete Shadowsocks Server
// @Description Delete a shadowsocks server (Admin only)
// @Tags Admin-ShadowsocksServers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Shadowsocks Server ID"
// @Success 200 {object} response.StandardResponse
// @Failure 400 {object} response.StandardResponse
// @Failure 401 {object} response.StandardResponse
// @Failure 403 {object} response.StandardResponse
// @Failure 404 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /admin/shadowsocks-servers/{id} [delete]
func (h *ShadowsocksServerHandler) DeleteShadowsocksServer(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.BadRequest(c, "Invalid server ID", "Server ID must be a valid number")
		return
	}

	err = h.shadowsocksServerService.DeleteShadowsocksServer(c.Request.Context(), id)
	if err != nil {
		if err.Error() == "shadowsocks server not found" {
			response.NotFound(c, "Shadowsocks server not found")
			return
		}
		response.InternalServerError(c, "Failed to delete shadowsocks server", err.Error())
		return
	}

	response.SuccessWithMessage(c, "Shadowsocks server deleted successfully", nil)
}

// GetAvailableShadowsocksServers retrieves available shadowsocks servers for current user
// @Summary Get Available Shadowsocks Servers
// @Description Get available shadowsocks servers for current user based on subscription
// @Tags User-ShadowsocksServers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.StandardResponse{data=[]model.ShadowsocksServerResponse}
// @Failure 401 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /user/shadowsocks-servers [get]
func (h *ShadowsocksServerHandler) GetAvailableShadowsocksServers(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	// Get user's active subscriptions
	subscriptions, _, err := h.subscriptionService.GetUserSubscriptions(c.Request.Context(), &service.GetUserSubscriptionsRequest{
		UserID: userID.(uint),
		Status: "active",
		Limit:  1, // We only need one active subscription
	})
	if err != nil {
		response.InternalServerError(c, "Failed to get user subscriptions", err.Error())
		return
	}

	if len(subscriptions) == 0 {
		// User has no active subscription, return empty list
		response.SuccessWithMessage(c, "No available shadowsocks servers", []interface{}{})
		return
	}

	// Use the first active subscription
	subscription := subscriptions[0]

	// Get server group IDs that user can access
	serverGroupIDs := subscription.GetServerGroupIDs()
	if len(serverGroupIDs) == 0 {
		// SECURITY: Follow "deny by default" principle
		// Empty server_group_ids means no access to any servers
		// Only [0] explicitly grants access to all servers
		response.SuccessWithMessage(c, "No available shadowsocks servers - subscription has no server group access", []interface{}{})
		return
	}

	// Check if user has access to all server groups (indicated by group ID 0)
	hasAllAccess := false
	var filteredGroupIDs []uint
	for _, groupID := range serverGroupIDs {
		if groupID == 0 {
			hasAllAccess = true
			break
		}
		filteredGroupIDs = append(filteredGroupIDs, groupID)
	}

	if hasAllAccess {
		// User has explicit access to all server groups
		servers, err := h.shadowsocksServerService.GetActiveShadowsocksServers(c.Request.Context())
		if err != nil {
			response.InternalServerError(c, "Failed to get shadowsocks servers", err.Error())
			return
		}

		// Convert to public response format (hide sensitive info)
		serverResponses := make([]interface{}, 0, len(servers))
		for _, server := range servers {
			serverResponses = append(serverResponses, server.ToPublicResponse())
		}

		response.SuccessWithMessage(c, "Available shadowsocks servers retrieved successfully", serverResponses)
		return
	}

	// Get servers for specific groups (user has limited access)
	var allServers []interface{}
	for _, groupID := range filteredGroupIDs {
		// Skip group ID 0 as it's handled separately above
		if groupID == 0 {
			continue
		}
		
		servers, err := h.shadowsocksServerService.GetShadowsocksServersByGroupID(c.Request.Context(), groupID)
		if err != nil {
			// Log error but continue with other groups
			continue
		}

		for _, server := range servers {
			allServers = append(allServers, server.ToPublicResponse())
		}
	}

	response.SuccessWithMessage(c, "Available shadowsocks servers retrieved successfully", allServers)
}

