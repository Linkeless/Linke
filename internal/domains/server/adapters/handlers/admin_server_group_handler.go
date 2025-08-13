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


type AdminServerGroupHandler struct {
	serverGroupService     interfaces.ServerGroupService
	shadowsocksService     interfaces.ShadowsocksServerService
}

func NewAdminServerGroupHandler(
	serverGroupService interfaces.ServerGroupService,
	shadowsocksService interfaces.ShadowsocksServerService,
) *AdminServerGroupHandler {
	return &AdminServerGroupHandler{
		serverGroupService: serverGroupService,
		shadowsocksService: shadowsocksService,
	}
}

// CreateGroup godoc
// @Summary Create new server group
// @Description Create a new server group (Admin only)
// @Tags Admin-Server-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param group body dto.CreateServerGroupRequest true "Server group creation data"
// @Success 201 {object} entities.ServerGroupResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 409 {object} response.ConflictResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/server-groups [post]
func (h *AdminServerGroupHandler) CreateGroup(c *gin.Context) {
	var createReq dto.CreateServerGroupRequest
	if err := c.ShouldBindJSON(&createReq); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Convert to service request
	serviceReq := &dto.CreateServerGroupRequest{
		Name: createReq.Name,
	}

	group, err := h.serverGroupService.CreateServerGroup(c.Request.Context(), serviceReq)
	if err != nil {
		logger.Error("Admin failed to create server group",
			logger.String("name", createReq.Name),
			logger.ErrorField(err),
		)

		// Check if it's a duplicate key error
		if strings.Contains(err.Error(), "Duplicate entry") || strings.Contains(err.Error(), "UNIQUE constraint") {
			response.Conflict(c, "Server group with this name already exists")
			return
		}

		response.InternalServerError(c, "Failed to create server group")
		return
	}

	logger.Info("Admin created new server group",
		logger.Uint("group_id", group.ID),
		logger.String("name", group.Name),
		logger.String("admin_action", "create_server_group"),
	)

	response.Created(c, group.ToResponse())
}

// ListGroups godoc
// @Summary List all server groups
// @Description Get paginated list of all server groups (Admin only)
// @Tags Admin-Server-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} response.HALCollectionResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/server-groups [get]
func (h *AdminServerGroupHandler) ListGroups(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit

	// Create service request
	serviceReq := &dto.GetServerGroupsRequest{
		Limit:  limit,
		Offset: offset,
	}

	groups, total, err := h.serverGroupService.GetServerGroups(c.Request.Context(), serviceReq)
	if err != nil {
		logger.Error("Admin failed to list server groups", logger.ErrorField(err))
		response.InternalServerError(c, "Failed to list server groups")
		return
	}

	// Convert to response format
	groupResponses := make([]*entities.ServerGroupResponse, len(groups))
	for i, group := range groups {
		groupResponses[i] = group.ToResponse()
	}

	response.Paginated(c, "Server groups retrieved successfully", groupResponses, page, limit, total, "/api/v1/admin/server-groups")
}

// GetGroup godoc
// @Summary Get server group information
// @Description Get server group details by group ID (Admin only)
// @Tags Admin-Server-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Group ID"
// @Success 200 {object} entities.ServerGroupResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Router /admin/server-groups/{id} [get]
func (h *AdminServerGroupHandler) GetGroup(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid group ID")
		return
	}

	group, err := h.serverGroupService.GetServerGroup(c.Request.Context(), uint(id))
	if err != nil {
		logger.Error("Admin failed to get server group",
			logger.Uint("group_id", uint(id)),
			logger.ErrorField(err),
		)
		response.NotFound(c, "Server group not found")
		return
	}

	response.OK(c, group.ToResponse())
}

// UpdateGroup godoc
// @Summary [Admin] Update server group
// @Description Update server group information (admin only)
// @Tags Admin-Server-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Group ID"
// @Param group body dto.UpdateServerGroupRequest true "Server group data"
// @Success 200 {object} entities.ServerGroupResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/server-groups/{id} [put]
func (h *AdminServerGroupHandler) UpdateGroup(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid group ID")
		return
	}

	var updateReq dto.UpdateServerGroupRequest
	if err := c.ShouldBindJSON(&updateReq); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Convert to service request
	serviceReq := &dto.UpdateServerGroupRequest{
		Name: updateReq.Name,
	}

	group, err := h.serverGroupService.UpdateServerGroup(c.Request.Context(), uint(id), serviceReq)
	if err != nil {
		logger.Error("Admin failed to update server group",
			logger.Uint("group_id", uint(id)),
			logger.ErrorField(err),
		)

		if strings.Contains(err.Error(), "not found") {
			response.NotFound(c, "Server group not found")
			return
		}

		response.InternalServerError(c, "Failed to update server group")
		return
	}

	response.OK(c, group.ToResponse())
}

// PatchGroup godoc
// @Summary [Admin] Partially update server group
// @Description Partially update server group information using PATCH method (admin only)
// @Tags Admin-Server-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Group ID"
// @Param group body dto.PatchServerGroupRequest true "Partial server group data"
// @Success 200 {object} entities.ServerGroupResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/server-groups/{id} [patch]
func (h *AdminServerGroupHandler) PatchGroup(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid group ID")
		return
	}

	var patchReq dto.PatchServerGroupRequest
	if err := c.ShouldBindJSON(&patchReq); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Convert to service request (only non-nil fields)
	serviceReq := &dto.UpdateServerGroupRequest{
		Name: patchReq.Name,
	}

	group, err := h.serverGroupService.UpdateServerGroup(c.Request.Context(), uint(id), serviceReq)
	if err != nil {
		logger.Error("Admin failed to patch server group",
			logger.Uint("group_id", uint(id)),
			logger.Any("patch_request", patchReq),
			logger.ErrorField(err),
		)

		if strings.Contains(err.Error(), "not found") {
			response.NotFound(c, "Server group not found")
			return
		}

		response.InternalServerError(c, "Failed to update server group")
		return
	}

	response.OK(c, group.ToResponse())
}

// DeleteGroup godoc
// @Summary [Admin] Delete server group
// @Description Delete a server group (admin only). Note: This will fail if group has servers.
// @Tags Admin-Server-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Group ID"
// @Success 200 {object} string
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 409 {object} response.ConflictResponse
// @Router /admin/server-groups/{id} [delete]
func (h *AdminServerGroupHandler) DeleteGroup(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid group ID")
		return
	}

	// Check if group has servers
	serverCount, err := h.serverGroupService.GetGroupServerCount(c.Request.Context(), uint(id))
	if err != nil {
		logger.Error("Admin failed to check server count for group deletion",
			logger.Uint("group_id", uint(id)),
			logger.ErrorField(err),
		)
		response.InternalServerError(c, "Failed to verify group status")
		return
	}

	if serverCount > 0 {
		response.Conflict(c, "Cannot delete group that contains servers. Move or delete servers first.")
		return
	}

	if err := h.serverGroupService.DeleteServerGroup(c.Request.Context(), uint(id)); err != nil {
		logger.Error("Admin failed to delete server group",
			logger.Uint("group_id", uint(id)),
			logger.ErrorField(err),
		)

		if strings.Contains(err.Error(), "not found") {
			response.NotFound(c, "Server group not found")
			return
		}

		response.InternalServerError(c, "Failed to delete server group")
		return
	}

	response.OK(c, nil)
}

// GetGroupServers godoc
// @Summary [Admin] Get servers in a group
// @Description Get all servers belonging to a specific group (admin only)
// @Tags Admin-Server-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Group ID"
// @Success 200 {object} response.StandardResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/server-groups/{id}/servers [get]
func (h *AdminServerGroupHandler) GetGroupServers(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid group ID")
		return
	}

	servers, err := h.serverGroupService.GetGroupServers(c.Request.Context(), uint(id))
	if err != nil {
		logger.Error("Admin failed to get group servers",
			logger.Uint("group_id", uint(id)),
			logger.ErrorField(err),
		)

		if strings.Contains(err.Error(), "not found") {
			response.NotFound(c, "Server group not found")
			return
		}

		response.InternalServerError(c, "Failed to get group servers")
		return
	}

	// Convert to response format
	serverResponses := make([]*entities.ShadowsocksServerResponse, len(servers))
	for i, server := range servers {
		serverResponses[i] = server.ToResponse()
	}

	response.OK(c, serverResponses)
}

// GetGroupStatistics godoc
// @Summary [Admin] Get server group statistics
// @Description Get server group statistics and metrics (admin only)
// @Tags Admin-Server-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Group ID"
// @Success 200 {object} response.StandardResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/server-groups/{id}/statistics [get]
func (h *AdminServerGroupHandler) GetGroupStatistics(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid group ID")
		return
	}

	stats, err := h.serverGroupService.GetGroupStatistics(c.Request.Context(), uint(id))
	if err != nil {
		logger.Error("Admin failed to get group statistics",
			logger.Uint("group_id", uint(id)),
			logger.ErrorField(err),
		)

		if strings.Contains(err.Error(), "not found") {
			response.NotFound(c, "Server group not found")
			return
		}

		response.InternalServerError(c, "Failed to get group statistics")
		return
	}

	response.OK(c, stats)
}

// GetAllGroupStatistics godoc
// @Summary [Admin] Get all groups statistics
// @Description Get statistics for all server groups (admin only)
// @Tags Admin-Server-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.StandardResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/server-groups/statistics [get]
func (h *AdminServerGroupHandler) GetAllGroupStatistics(c *gin.Context) {
	groups, err := h.serverGroupService.GetAllServerGroups(c.Request.Context())
	if err != nil {
		logger.Error("Admin failed to get all server groups for statistics", logger.ErrorField(err))
		response.InternalServerError(c, "Failed to get group statistics")
		return
	}

	statistics := make(map[string]any)
	for _, group := range groups {
		stats, err := h.serverGroupService.GetGroupStatistics(c.Request.Context(), group.ID)
		if err != nil {
			logger.Warn("Failed to get statistics for group",
				logger.Uint("group_id", group.ID),
				logger.String("group_name", group.Name),
				logger.ErrorField(err),
			)
			continue
		}
		statistics[group.Name] = stats
	}

	response.OK(c, gin.H{
		"total_groups": len(groups),
		"statistics":   statistics,
	})
}