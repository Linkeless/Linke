package admin

import (
	"strconv"

	"linke/internal/logger"
	"linke/internal/response"
	"linke/internal/service"

	"github.com/gin-gonic/gin"
)

type ServerGroupHandler struct {
	serverGroupService *service.ServerGroupService
}

func NewServerGroupHandler(serverGroupService *service.ServerGroupService) *ServerGroupHandler {
	return &ServerGroupHandler{
		serverGroupService: serverGroupService,
	}
}

// CreateServerGroup godoc
// @Summary [Admin] Create server group
// @Description Create a new server group
// @Tags Admin-ServerGroup-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param group body service.CreateServerGroupRequest true "Server group data"
// @Success 201 {object} response.StandardResponse{data=model.ServerGroupResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/server-groups [post]
func (h *ServerGroupHandler) CreateServerGroup(c *gin.Context) {
	var req service.CreateServerGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request", err.Error())
		return
	}

	group, err := h.serverGroupService.CreateServerGroup(c.Request.Context(), &req)
	if err != nil {
		logger.Error("Failed to create server group",
			logger.String("name", req.Name),
			logger.Error2("error", err),
		)
		response.BadRequest(c, "Failed to create server group", err.Error())
		return
	}

	response.CreatedWithMessage(c, "Server group created successfully", group.ToResponse())
}

// ListServerGroups godoc
// @Summary [Admin] List server groups
// @Description Get paginated list of server groups
// @Tags Admin-ServerGroup-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} response.StandardListResponse{data=[]model.ServerGroupResponse}
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/server-groups [get]
func (h *ServerGroupHandler) ListServerGroups(c *gin.Context) {
	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit

	req := &service.GetServerGroupsRequest{
		Limit:  limit,
		Offset: offset,
	}

	groups, total, err := h.serverGroupService.GetServerGroups(c.Request.Context(), req)
	if err != nil {
		logger.Error("Failed to get server groups",
			logger.Error2("error", err),
		)
		response.InternalServerError(c, "Failed to get server groups", err.Error())
		return
	}

	// Convert to response format
	var groupResponses []*response.ServerGroupResponseData
	for _, group := range groups {
		groupResponses = append(groupResponses, &response.ServerGroupResponseData{
			ID:        group.ID,
			Name:      group.Name,
			CreatedAt: group.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt: group.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	response.SuccessList(c, groupResponses, page, limit, total)
}

// GetServerGroup godoc
// @Summary [Admin] Get server group by ID
// @Description Get server group details by ID
// @Tags Admin-ServerGroup-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Server group ID"
// @Success 200 {object} response.StandardResponse{data=model.ServerGroupResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Router /admin/server-groups/{id} [get]
func (h *ServerGroupHandler) GetServerGroup(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid server group ID", "Server group ID must be a valid number")
		return
	}

	group, err := h.serverGroupService.GetServerGroup(c.Request.Context(), uint(id))
	if err != nil {
		logger.Error("Failed to get server group",
			logger.Uint("group_id", uint(id)),
			logger.Error2("error", err),
		)
		if err.Error() == "server group not found" {
			response.NotFound(c, "Server group not found")
			return
		}
		response.InternalServerError(c, "Failed to get server group", err.Error())
		return
	}

	response.SuccessWithMessage(c, "Server group retrieved successfully", group.ToResponse())
}

// UpdateServerGroup godoc
// @Summary [Admin] Update server group
// @Description Update server group by ID
// @Tags Admin-ServerGroup-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Server group ID"
// @Param group body service.UpdateServerGroupRequest true "Updated server group data"
// @Success 200 {object} response.StandardResponse{data=model.ServerGroupResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Router /admin/server-groups/{id} [put]
func (h *ServerGroupHandler) UpdateServerGroup(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid server group ID", "Server group ID must be a valid number")
		return
	}

	var req service.UpdateServerGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request", err.Error())
		return
	}

	group, err := h.serverGroupService.UpdateServerGroup(c.Request.Context(), uint(id), &req)
	if err != nil {
		logger.Error("Failed to update server group",
			logger.Uint("group_id", uint(id)),
			logger.Error2("error", err),
		)
		if err.Error() == "server group not found" {
			response.NotFound(c, "Server group not found")
			return
		}
		response.BadRequest(c, "Failed to update server group", err.Error())
		return
	}

	response.SuccessWithMessage(c, "Server group updated successfully", group.ToResponse())
}

// DeleteServerGroup godoc
// @Summary [Admin] Delete server group
// @Description Delete server group by ID
// @Tags Admin-ServerGroup-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Server group ID"
// @Success 200 {object} response.MessageOnlyResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 404 {object} response.NotFoundResponse
// @Router /admin/server-groups/{id} [delete]
func (h *ServerGroupHandler) DeleteServerGroup(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid server group ID", "Server group ID must be a valid number")
		return
	}

	if err := h.serverGroupService.DeleteServerGroup(c.Request.Context(), uint(id)); err != nil {
		logger.Error("Failed to delete server group",
			logger.Uint("group_id", uint(id)),
			logger.Error2("error", err),
		)
		if err.Error() == "server group not found" {
			response.NotFound(c, "Server group not found")
			return
		}
		response.InternalServerError(c, "Failed to delete server group", err.Error())
		return
	}

	response.SuccessWithMessage(c, "Server group deleted successfully", nil)
}

// GetAllServerGroups godoc
// @Summary [Admin] Get all server groups
// @Description Get all server groups for dropdown/selection purposes
// @Tags Admin-ServerGroup-Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.StandardResponse{data=[]model.ServerGroupResponse}
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 403 {object} response.ForbiddenResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/server-groups/all [get]
func (h *ServerGroupHandler) GetAllServerGroups(c *gin.Context) {
	groups, err := h.serverGroupService.GetAllServerGroups(c.Request.Context())
	if err != nil {
		logger.Error("Failed to get all server groups",
			logger.Error2("error", err),
		)
		response.InternalServerError(c, "Failed to get server groups", err.Error())
		return
	}

	// Convert to response format
	var groupResponses []*response.ServerGroupResponseData
	for _, group := range groups {
		groupResponses = append(groupResponses, &response.ServerGroupResponseData{
			ID:        group.ID,
			Name:      group.Name,
			CreatedAt: group.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt: group.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	response.SuccessWithMessage(c, "Server groups retrieved successfully", groupResponses)
}