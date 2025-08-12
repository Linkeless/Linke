package handlers

import (
	"strconv"
	"strings"

	"linke/internal/domains/server/dto"
	"linke/internal/domains/server/entities"
	"linke/internal/domains/server/usecases/interfaces"
	"linke/internal/shared/logger"
	"linke/internal/shared/response"

	"github.com/gin-gonic/gin"
)

// UpdateServerStatusRequest represents the request body for updating server status
type UpdateServerStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=active inactive maintenance" example:"active"`
}

// BulkUpdateServersRequest represents the request body for bulk server operations
type BulkUpdateServersRequest struct {
	IDs     []uint         `json:"ids" binding:"required,min=1,max=100"`
	Updates map[string]any `json:"updates" binding:"required"`
}

// BatchServerIDsRequest represents the request body for batch operations on servers
type BatchServerIDsRequest struct {
	IDs []uint `json:"ids" binding:"required,min=1,max=100"`
}

type AdminServerHandler struct {
	shadowsocksService interfaces.ShadowsocksServerService
	serverGroupService interfaces.ServerGroupService
}

func NewAdminServerHandler(
	shadowsocksService interfaces.ShadowsocksServerService,
	serverGroupService interfaces.ServerGroupService,
) *AdminServerHandler {
	return &AdminServerHandler{
		shadowsocksService: shadowsocksService,
		serverGroupService: serverGroupService,
	}
}

// CreateServer godoc
// @Summary Create new shadowsocks server
// @Description Create a new shadowsocks server (Admin only)
// @Tags Admin-Server-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param server body dto.CreateShadowsocksServerRequest true "Server creation data"
// @Success 201 {object} response.StandardResponse{data=entities.ShadowsocksServerResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 409 {object} response.ConflictResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/servers [post]
func (h *AdminServerHandler) CreateServer(c *gin.Context) {
	var createReq dto.CreateShadowsocksServerRequest
	if err := c.ShouldBindJSON(&createReq); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Convert to service request
	serviceReq := &dto.CreateShadowsocksServerRequest{
		GroupID:      createReq.GroupID,
		RouteID:      createReq.RouteID,
		ParentID:     createReq.ParentID,
		Name:         createReq.Name,
		Tags:         createReq.Tags,
		Host:         createReq.Host,
		Port:         createReq.Port,
		ServerPort:   createReq.ServerPort,
		Cipher:       createReq.Cipher,
		Obfs:         createReq.Obfs,
		ObfsSettings: createReq.ObfsSettings,
		Excludes:     createReq.Excludes,
		IPs:          createReq.IPs,
		Rate:         createReq.Rate,
		Show:         createReq.Show,
		Sort:         createReq.Sort,
	}

	server, err := h.shadowsocksService.CreateShadowsocksServer(c.Request.Context(), serviceReq)
	if err != nil {
		logger.Error("Admin failed to create server",
			logger.String("name", createReq.Name),
			logger.String("host", createReq.Host),
			logger.ErrorField(err),
		)

		// Check if it's a duplicate key error
		if strings.Contains(err.Error(), "Duplicate entry") || strings.Contains(err.Error(), "UNIQUE constraint") {
			response.Conflict(c, "Server with this name or host already exists")
			return
		}

		response.InternalServerError(c, "Failed to create server")
		return
	}

	logger.Info("Admin created new server",
		logger.Int("server_id", server.ID),
		logger.String("name", server.Name),
		logger.String("host", server.Host),
		logger.String("admin_action", "create_server"),
	)

	response.Created(c, server.ToResponse())
}

// ListServers godoc
// @Summary List all shadowsocks servers
// @Description Get paginated list of all shadowsocks servers (Admin only)
// @Tags Admin-Server-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param group_id query int false "Filter by server group ID"
// @Param show query int false "Filter by visibility (0 or 1)"
// @Param name query string false "Filter by server name (substring match)"
// @Param sort_by query string false "Sort by field" Enums(sort,created_at,updated_at,name,rate) default(sort)
// @Param sort_order query string false "Sort order" Enums(asc,desc) default(asc)
// @Success 200 {object} response.StandardListResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/servers [get]
func (h *AdminServerHandler) ListServers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit

	// Parse filters
	var groupID *uint
	if groupIDStr := c.Query("group_id"); groupIDStr != "" {
		if gid, err := strconv.ParseUint(groupIDStr, 10, 32); err == nil {
			id := uint(gid)
			groupID = &id
		}
	}

	var show *int
	if showStr := c.Query("show"); showStr != "" {
		if s, err := strconv.Atoi(showStr); err == nil {
			show = &s
		}
	}

	name := c.Query("name")
	sortBy := c.DefaultQuery("sort_by", "sort")
	sortOrder := strings.ToLower(c.DefaultQuery("sort_order", "asc"))
	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "asc"
	}

	// Create service request
	serviceReq := &dto.GetShadowsocksServersRequest{
		GroupID:   groupID,
		Show:      show,
		Name:      name,
		SortBy:    sortBy,
		SortOrder: sortOrder,
		Limit:     limit,
		Offset:    offset,
	}

	servers, total, err := h.shadowsocksService.GetShadowsocksServers(c.Request.Context(), serviceReq)
	if err != nil {
		logger.Error("Admin failed to list servers", logger.ErrorField(err))
		response.InternalServerError(c, "Failed to list servers")
		return
	}

	// Convert to response format
	serverResponses := make([]*entities.ShadowsocksServerResponse, len(servers))
	for i, server := range servers {
		serverResponses[i] = server.ToResponse()
	}

	response.SuccessList(c, serverResponses, page, limit, total)
}

// GetServer godoc
// @Summary Get server information
// @Description Get server details by server ID (Admin only)
// @Tags Admin-Server-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Server ID"
// @Success 200 {object} response.StandardResponse{data=entities.ShadowsocksServerResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Router /admin/servers/{id} [get]
func (h *AdminServerHandler) GetServer(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.BadRequest(c, "Invalid server ID")
		return
	}

	server, err := h.shadowsocksService.GetShadowsocksServerByID(c.Request.Context(), id)
	if err != nil {
		logger.Error("Admin failed to get server",
			logger.Int("server_id", id),
			logger.ErrorField(err),
		)
		response.NotFound(c, "Server not found")
		return
	}

	response.Success(c, server.ToResponse())
}

// UpdateServer godoc
// @Summary [Admin] Update shadowsocks server
// @Description Update shadowsocks server information (admin only)
// @Tags Admin-Server-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Server ID"
// @Param server body dto.UpdateShadowsocksServerRequest true "Server data"
// @Success 200 {object} response.StandardResponse{data=entities.ShadowsocksServerResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/servers/{id} [put]
func (h *AdminServerHandler) UpdateServer(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid server ID")
		return
	}

	var updateReq dto.UpdateShadowsocksServerRequest
	if err := c.ShouldBindJSON(&updateReq); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Convert to service request
	serviceReq := &dto.UpdateShadowsocksServerRequest{
		GroupID:      updateReq.GroupID,
		RouteID:      updateReq.RouteID,
		ParentID:     updateReq.ParentID,
		Name:         updateReq.Name,
		Tags:         updateReq.Tags,
		Host:         updateReq.Host,
		Port:         updateReq.Port,
		ServerPort:   updateReq.ServerPort,
		Cipher:       updateReq.Cipher,
		Obfs:         updateReq.Obfs,
		ObfsSettings: updateReq.ObfsSettings,
		Excludes:     updateReq.Excludes,
		IPs:          updateReq.IPs,
		Rate:         updateReq.Rate,
		Show:         updateReq.Show,
		Sort:         updateReq.Sort,
	}

	server, err := h.shadowsocksService.UpdateShadowsocksServer(c.Request.Context(), uint(id), serviceReq)
	if err != nil {
		logger.Error("Admin failed to update server",
			logger.Uint("server_id", uint(id)),
			logger.ErrorField(err),
		)
		response.InternalServerError(c, "Failed to update server")
		return
	}

	response.Success(c, server.ToResponse())
}

// PatchServer godoc
// @Summary [Admin] Partially update server
// @Description Partially update server information using PATCH method (admin only)
// @Tags Admin-Server-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Server ID"
// @Param server body dto.PatchShadowsocksServerRequest true "Partial server data"
// @Success 200 {object} response.StandardResponse{data=entities.ShadowsocksServerResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/servers/{id} [patch]
func (h *AdminServerHandler) PatchServer(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid server ID")
		return
	}

	var patchReq dto.PatchShadowsocksServerRequest
	if err := c.ShouldBindJSON(&patchReq); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Convert to service request (only non-nil fields)
	serviceReq := &dto.UpdateShadowsocksServerRequest{
		GroupID:      patchReq.GroupID,
		RouteID:      patchReq.RouteID,
		ParentID:     patchReq.ParentID,
		Name:         patchReq.Name,
		Tags:         patchReq.Tags,
		Host:         patchReq.Host,
		Port:         patchReq.Port,
		ServerPort:   patchReq.ServerPort,
		Cipher:       patchReq.Cipher,
		Obfs:         patchReq.Obfs,
		ObfsSettings: patchReq.ObfsSettings,
		Excludes:     patchReq.Excludes,
		IPs:          patchReq.IPs,
		Rate:         patchReq.Rate,
		Show:         patchReq.Show,
		Sort:         patchReq.Sort,
	}

	server, err := h.shadowsocksService.UpdateShadowsocksServer(c.Request.Context(), uint(id), serviceReq)
	if err != nil {
		logger.Error("Admin failed to patch server",
			logger.Uint("server_id", uint(id)),
			logger.Any("patch_request", patchReq),
			logger.ErrorField(err),
		)
		response.InternalServerError(c, "Failed to update server")
		return
	}

	response.Success(c, server.ToResponse())
}

// DeleteServer godoc
// @Summary [Admin] Delete server
// @Description Delete a shadowsocks server (admin only)
// @Tags Admin-Server-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Server ID"
// @Success 200 {object} response.MessageOnlyResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Router /admin/servers/{id} [delete]
func (h *AdminServerHandler) DeleteServer(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid server ID")
		return
	}

	if err := h.shadowsocksService.DeleteShadowsocksServer(c.Request.Context(), uint(id)); err != nil {
		logger.Error("Admin failed to delete server",
			logger.Uint("server_id", uint(id)),
			logger.ErrorField(err),
		)
		response.NotFound(c, "Server not found")
		return
	}

	response.SuccessWithMessage(c, "Server deleted successfully", nil)
}

// UpdateServerStatus godoc
// @Summary [Admin] Update server status
// @Description Update server status (admin only)
// @Tags Admin-Server-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Server ID"
// @Param status body UpdateServerStatusRequest true "Status data"
// @Success 200 {object} response.MessageOnlyResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Router /admin/servers/{id}/status [put]
func (h *AdminServerHandler) UpdateServerStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid server ID")
		return
	}

	var statusData UpdateServerStatusRequest
	if err := c.ShouldBindJSON(&statusData); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.shadowsocksService.UpdateServerStatus(c.Request.Context(), uint(id), statusData.Status); err != nil {
		logger.Error("Admin failed to update server status",
			logger.Uint("server_id", uint(id)),
			logger.String("status", statusData.Status),
			logger.ErrorField(err),
		)
		response.NotFound(c, "Server not found")
		return
	}

	response.SuccessWithMessage(c, "Server status updated successfully", nil)
}

// GetServerStatistics godoc
// @Summary [Admin] Get server statistics
// @Description Get server statistics and metrics (admin only)
// @Tags Admin-Server-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Server ID"
// @Success 200 {object} response.StandardResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/servers/{id}/statistics [get]
func (h *AdminServerHandler) GetServerStatistics(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid server ID")
		return
	}

	stats, err := h.shadowsocksService.GetServerStatistics(c.Request.Context(), uint(id))
	if err != nil {
		logger.Error("Admin failed to get server statistics",
			logger.Uint("server_id", uint(id)),
			logger.ErrorField(err),
		)
		response.InternalServerError(c, "Failed to get server statistics")
		return
	}

	response.Success(c, stats)
}

// BulkUpdateServers godoc
// @Summary [Admin] Bulk update servers
// @Description Update multiple servers at once (admin only)
// @Tags Admin-Server-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body BulkUpdateServersRequest true "Bulk update data"
// @Success 200 {object} response.StandardResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/servers/bulk/update [post]
func (h *AdminServerHandler) BulkUpdateServers(c *gin.Context) {
	var requestData BulkUpdateServersRequest
	if err := c.ShouldBindJSON(&requestData); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.shadowsocksService.BulkUpdateServers(c.Request.Context(), requestData.IDs, requestData.Updates); err != nil {
		logger.Error("Admin failed to bulk update servers",
			logger.Any("server_ids", requestData.IDs),
			logger.Any("updates", requestData.Updates),
			logger.ErrorField(err),
		)
		response.InternalServerError(c, "Failed to update servers")
		return
	}

	response.SuccessWithMessage(c, "Servers updated successfully", gin.H{
		"updated_count": len(requestData.IDs),
		"server_ids":    requestData.IDs,
	})
}

// CheckServerHealth godoc
// @Summary [Admin] Check server health
// @Description Check server health status (admin only)
// @Tags Admin-Server-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Server ID"
// @Success 200 {object} response.StandardResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/servers/{id}/health [get]
func (h *AdminServerHandler) CheckServerHealth(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid server ID")
		return
	}

	health, err := h.shadowsocksService.CheckServerHealth(c.Request.Context(), uint(id))
	if err != nil {
		logger.Error("Admin failed to check server health",
			logger.Uint("server_id", uint(id)),
			logger.ErrorField(err),
		)
		response.InternalServerError(c, "Failed to check server health")
		return
	}

	response.Success(c, health)
}

// GetServersByGroup godoc
// @Summary [Admin] Get servers by group
// @Description Get all servers in a specific group (admin only)
// @Tags Admin-Server-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param group_id path int true "Group ID"
// @Success 200 {object} response.StandardResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/servers/group/{group_id} [get]
func (h *AdminServerHandler) GetServersByGroup(c *gin.Context) {
	groupIDStr := c.Param("group_id")
	groupID, err := strconv.ParseUint(groupIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid group ID")
		return
	}

	servers, err := h.shadowsocksService.GetServersByGroup(c.Request.Context(), uint(groupID))
	if err != nil {
		logger.Error("Admin failed to get servers by group",
			logger.Uint("group_id", uint(groupID)),
			logger.ErrorField(err),
		)
		response.InternalServerError(c, "Failed to get servers")
		return
	}

	// Convert to response format
	serverResponses := make([]*entities.ShadowsocksServerResponse, len(servers))
	for i, server := range servers {
		serverResponses[i] = server.ToResponse()
	}

	response.Success(c, serverResponses)
}
