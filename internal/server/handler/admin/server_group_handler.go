package admin

import (
	"strconv"

	"linke/internal/response"
	"linke/internal/server/handler/dto"
	"linke/internal/server/service/command"
	"linke/internal/server/service/query"

	"github.com/gin-gonic/gin"
)

// ServerGroupHandler handles admin server group requests
type ServerGroupHandler struct {
	commandHandler *command.ServerGroupCommandHandler
	queryHandler   *query.ServerGroupQueryHandler
}

// NewServerGroupHandler creates a new server group handler
func NewServerGroupHandler(
	commandHandler *command.ServerGroupCommandHandler,
	queryHandler *query.ServerGroupQueryHandler,
) *ServerGroupHandler {
	return &ServerGroupHandler{
		commandHandler: commandHandler,
		queryHandler:   queryHandler,
	}
}

// CreateServerGroup creates a new server group
// @Summary Create a new server group
// @Description Create a new server group with the provided information
// @Tags Server Groups
// @Accept json
// @Produce json
// @Param request body dto.CreateServerGroupRequest true "Server group creation request"
// @Success 201 {object} response.StandardResponse{data=dto.ServerGroupResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 409 {object} response.ConflictResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/server-groups [post]
func (h *ServerGroupHandler) CreateServerGroup(c *gin.Context) {
	var req dto.CreateServerGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request", err.Error())
		return
	}

	cmd := command.CreateServerGroupCommand{
		Name: req.Name,
	}

	group, err := h.commandHandler.HandleCreateServerGroup(c.Request.Context(), cmd)
	if err != nil {
		if err == command.ErrServerGroupAlreadyExists {
			response.Conflict(c, "Server group already exists")
			return
		}
		response.InternalServerError(c, "Failed to create server group", err.Error())
		return
	}

	resp := dto.FromServerGroupDomain(group)
	response.CreatedWithMessage(c, "Server group created successfully", resp)
}

// GetServerGroup gets a server group by ID
// @Summary Get a server group
// @Description Get a server group by its ID
// @Tags Server Groups
// @Accept json
// @Produce json
// @Param id path int true "Server group ID"
// @Success 200 {object} response.StandardResponse{data=dto.ServerGroupResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/server-groups/{id} [get]
func (h *ServerGroupHandler) GetServerGroup(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid server group ID", err.Error())
		return
	}

	query := query.GetServerGroupQuery{
		ID: uint(id),
	}

	group, err := h.queryHandler.HandleGetServerGroup(c.Request.Context(), query)
	if err != nil {
		// Check for not found error using string comparison
		if err.Error() == "server group not found" {
			response.NotFound(c, "Server group not found")
			return
		}
		response.InternalServerError(c, "Failed to get server group", err.Error())
		return
	}

	resp := dto.FromServerGroupDomain(group)
	response.OK(c, "Server group retrieved successfully", resp)
}

// GetServerGroups gets server groups with pagination
// @Summary Get server groups
// @Description Get server groups with pagination
// @Tags Server Groups
// @Accept json
// @Produce json
// @Param limit query int false "Limit" default(10)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} response.StandardResponse{data=dto.ServerGroupListResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/server-groups [get]
func (h *ServerGroupHandler) GetServerGroups(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 0 || limit > 100 {
		response.BadRequest(c, "Invalid limit parameter", "Limit must be between 0 and 100")
		return
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		response.BadRequest(c, "Invalid offset parameter", "Offset must be >= 0")
		return
	}

	query := query.GetServerGroupsQuery{
		Limit:  limit,
		Offset: offset,
	}

	groups, total, err := h.queryHandler.HandleGetServerGroups(c.Request.Context(), query)
	if err != nil {
		response.InternalServerError(c, "Failed to get server groups", err.Error())
		return
	}

	resp := dto.ServerGroupListResponse{
		Groups: dto.FromServerGroupDomainList(groups),
		Total:  total,
	}

	response.OK(c, "Server groups retrieved successfully", resp)
}

// GetAllServerGroups gets all server groups
// @Summary Get all server groups
// @Description Get all server groups without pagination
// @Tags Server Groups
// @Accept json
// @Produce json
// @Success 200 {object} response.StandardResponse{data=[]dto.ServerGroupResponse}
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/server-groups/all [get]
func (h *ServerGroupHandler) GetAllServerGroups(c *gin.Context) {
	query := query.GetAllServerGroupsQuery{}

	groups, err := h.queryHandler.HandleGetAllServerGroups(c.Request.Context(), query)
	if err != nil {
		response.InternalServerError(c, "Failed to get all server groups", err.Error())
		return
	}

	resp := dto.FromServerGroupDomainList(groups)
	response.OK(c, "All server groups retrieved successfully", resp)
}

// UpdateServerGroup updates a server group
// @Summary Update a server group
// @Description Update a server group with the provided information
// @Tags Server Groups
// @Accept json
// @Produce json
// @Param id path int true "Server group ID"
// @Param request body dto.UpdateServerGroupRequest true "Server group update request"
// @Success 200 {object} response.StandardResponse{data=dto.ServerGroupResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 409 {object} response.ConflictResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/server-groups/{id} [put]
func (h *ServerGroupHandler) UpdateServerGroup(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid server group ID", err.Error())
		return
	}

	var req dto.UpdateServerGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request", err.Error())
		return
	}

	cmd := command.UpdateServerGroupCommand{
		ID:   uint(id),
		Name: req.Name,
	}

	group, err := h.commandHandler.HandleUpdateServerGroup(c.Request.Context(), cmd)
	if err != nil {
		if err == command.ErrServerGroupNotFound {
			response.NotFound(c, "Server group not found")
			return
		}
		if err == command.ErrServerGroupAlreadyExists {
			response.Conflict(c, "Server group name already exists")
			return
		}
		response.InternalServerError(c, "Failed to update server group", err.Error())
		return
	}

	resp := dto.FromServerGroupDomain(group)
	response.OK(c, "Server group updated successfully", resp)
}

// DeleteServerGroup deletes a server group
// @Summary Delete a server group
// @Description Delete a server group by its ID
// @Tags Server Groups
// @Accept json
// @Produce json
// @Param id path int true "Server group ID"
// @Success 200 {object} response.StandardResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 409 {object} response.ConflictResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/server-groups/{id} [delete]
func (h *ServerGroupHandler) DeleteServerGroup(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid server group ID", err.Error())
		return
	}

	cmd := command.DeleteServerGroupCommand{
		ID: uint(id),
	}

	err = h.commandHandler.HandleDeleteServerGroup(c.Request.Context(), cmd)
	if err != nil {
		if err == command.ErrServerGroupNotFound {
			response.NotFound(c, "Server group not found")
			return
		}
		if err == command.ErrCannotDeleteServerGroupWithServers {
			response.Conflict(c, "Cannot delete server group with servers")
			return
		}
		response.InternalServerError(c, "Failed to delete server group", err.Error())
		return
	}

	response.OK(c, "Server group deleted successfully", nil)
}
