package admin

import (
	"strconv"

	"linke/internal/response"
	"linke/internal/server/handler/dto"
	"linke/internal/server/service/command"
	"linke/internal/server/service/query"

	"github.com/gin-gonic/gin"
)

// ShadowsocksServerHandler handles admin shadowsocks server requests
type ShadowsocksServerHandler struct {
	commandHandler *command.ShadowsocksServerCommandHandler
	queryHandler   *query.ShadowsocksServerQueryHandler
}

// NewShadowsocksServerHandler creates a new shadowsocks server handler
func NewShadowsocksServerHandler(
	commandHandler *command.ShadowsocksServerCommandHandler,
	queryHandler *query.ShadowsocksServerQueryHandler,
) *ShadowsocksServerHandler {
	return &ShadowsocksServerHandler{
		commandHandler: commandHandler,
		queryHandler:   queryHandler,
	}
}

// CreateShadowsocksServer creates a new shadowsocks server
// @Summary Create a new shadowsocks server
// @Description Create a new shadowsocks server with the provided information
// @Tags Shadowsocks Servers
// @Accept json
// @Produce json
// @Param request body dto.CreateShadowsocksServerRequest true "Shadowsocks server creation request"
// @Success 201 {object} response.StandardResponse{data=dto.ShadowsocksServerResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 422 {object} response.UnprocessableEntityResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/shadowsocks-servers [post]
func (h *ShadowsocksServerHandler) CreateShadowsocksServer(c *gin.Context) {
	var req dto.CreateShadowsocksServerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request", err.Error())
		return
	}
	
	cmd := command.CreateShadowsocksServerCommand{
		GroupID:      req.GroupID,
		RouteID:      req.RouteID,
		ParentID:     req.ParentID,
		Name:         req.Name,
		Tags:         req.Tags,
		Host:         req.Host,
		Port:         req.Port,
		ServerPort:   req.ServerPort,
		Cipher:       req.Cipher,
		Obfs:         req.Obfs,
		ObfsSettings: req.ObfsSettings,
		Excludes:     req.Excludes,
		IPs:          req.IPs,
		Rate:         req.Rate,
		Show:         req.Show,
		Sort:         req.Sort,
	}
	
	server, err := h.commandHandler.HandleCreateShadowsocksServer(c.Request.Context(), cmd)
	if err != nil {
		if err == command.ErrInvalidServerGroup {
			response.UnprocessableEntity(c, "Invalid server group", err.Error())
			return
		}
		response.InternalServerError(c, "Failed to create shadowsocks server", err.Error())
		return
	}
	
	resp := dto.FromShadowsocksServerDomain(server)
	response.CreatedWithMessage(c, "Shadowsocks server created successfully", resp)
}

// GetShadowsocksServer gets a shadowsocks server by ID
// @Summary Get a shadowsocks server
// @Description Get a shadowsocks server by its ID
// @Tags Shadowsocks Servers
// @Accept json
// @Produce json
// @Param id path int true "Shadowsocks server ID"
// @Success 200 {object} response.StandardResponse{data=dto.ShadowsocksServerResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/shadowsocks-servers/{id} [get]
func (h *ShadowsocksServerHandler) GetShadowsocksServer(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.BadRequest(c, "Invalid shadowsocks server ID", err.Error())
		return
	}
	
	query := query.GetShadowsocksServerQuery{
		ID: id,
	}
	
	server, err := h.queryHandler.HandleGetShadowsocksServer(c.Request.Context(), query)
	if err != nil {
		if err.Error() == "shadowsocks server not found" {
			response.NotFound(c, "Shadowsocks server not found")
			return
		}
		response.InternalServerError(c, "Failed to get shadowsocks server", err.Error())
		return
	}
	
	resp := dto.FromShadowsocksServerDomain(server)
	response.OK(c, "Shadowsocks server retrieved successfully", resp)
}

// GetShadowsocksServers gets shadowsocks servers with filters and pagination
// @Summary Get shadowsocks servers
// @Description Get shadowsocks servers with filters and pagination
// @Tags Shadowsocks Servers
// @Accept json
// @Produce json
// @Param group_id query int false "Server group ID filter"
// @Param show query int false "Visibility filter (0 or 1)"
// @Param tags query string false "Tags filter"
// @Param cipher query string false "Cipher filter"
// @Param limit query int false "Limit" default(10)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} response.StandardResponse{data=dto.ShadowsocksServerListResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/shadowsocks-servers [get]
func (h *ShadowsocksServerHandler) GetShadowsocksServers(c *gin.Context) {
	// Parse query parameters
	var groupID *uint
	if groupIDStr := c.Query("group_id"); groupIDStr != "" {
		id, err := strconv.ParseUint(groupIDStr, 10, 32)
		if err != nil {
			response.BadRequest(c, "Invalid group_id parameter", err.Error())
			return
		}
		gid := uint(id)
		groupID = &gid
	}
	
	var show *int
	if showStr := c.Query("show"); showStr != "" {
		s, err := strconv.Atoi(showStr)
		if err != nil || (s != 0 && s != 1) {
			response.BadRequest(c, "Invalid show parameter", "Show must be 0 or 1")
			return
		}
		show = &s
	}
	
	tags := c.Query("tags")
	cipher := c.Query("cipher")
	
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
	
	query := query.GetShadowsocksServersQuery{
		GroupID: groupID,
		Show:    show,
		Tags:    tags,
		Cipher:  cipher,
		Limit:   limit,
		Offset:  offset,
	}
	
	servers, total, err := h.queryHandler.HandleGetShadowsocksServers(c.Request.Context(), query)
	if err != nil {
		response.InternalServerError(c, "Failed to get shadowsocks servers", err.Error())
		return
	}
	
	resp := dto.ShadowsocksServerListResponse{
		Servers: dto.FromShadowsocksServerDomainList(servers),
		Total:   total,
	}
	
	response.OK(c, "Shadowsocks servers retrieved successfully", resp)
}

// UpdateShadowsocksServer updates a shadowsocks server
// @Summary Update a shadowsocks server
// @Description Update a shadowsocks server with the provided information
// @Tags Shadowsocks Servers
// @Accept json
// @Produce json
// @Param id path int true "Shadowsocks server ID"
// @Param request body dto.UpdateShadowsocksServerRequest true "Shadowsocks server update request"
// @Success 200 {object} response.StandardResponse{data=dto.ShadowsocksServerResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 422 {object} response.UnprocessableEntityResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/shadowsocks-servers/{id} [put]
func (h *ShadowsocksServerHandler) UpdateShadowsocksServer(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.BadRequest(c, "Invalid shadowsocks server ID", err.Error())
		return
	}
	
	var req dto.UpdateShadowsocksServerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request", err.Error())
		return
	}
	
	cmd := command.UpdateShadowsocksServerCommand{
		ID:           id,
		GroupID:      req.GroupID,
		RouteID:      req.RouteID,
		ParentID:     req.ParentID,
		Name:         req.Name,
		Tags:         req.Tags,
		Host:         req.Host,
		Port:         req.Port,
		ServerPort:   req.ServerPort,
		Cipher:       req.Cipher,
		Obfs:         req.Obfs,
		ObfsSettings: req.ObfsSettings,
		Excludes:     req.Excludes,
		IPs:          req.IPs,
		Rate:         req.Rate,
		Show:         req.Show,
		Sort:         req.Sort,
	}
	
	server, err := h.commandHandler.HandleUpdateShadowsocksServer(c.Request.Context(), cmd)
	if err != nil {
		if err == command.ErrShadowsocksServerNotFound {
			response.NotFound(c, "Shadowsocks server not found")
			return
		}
		if err == command.ErrInvalidServerGroup {
			response.UnprocessableEntity(c, "Invalid server group", err.Error())
			return
		}
		response.InternalServerError(c, "Failed to update shadowsocks server", err.Error())
		return
	}
	
	resp := dto.FromShadowsocksServerDomain(server)
	response.OK(c, "Shadowsocks server updated successfully", resp)
}

// DeleteShadowsocksServer deletes a shadowsocks server
// @Summary Delete a shadowsocks server
// @Description Delete a shadowsocks server by its ID
// @Tags Shadowsocks Servers
// @Accept json
// @Produce json
// @Param id path int true "Shadowsocks server ID"
// @Success 200 {object} response.StandardResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/shadowsocks-servers/{id} [delete]
func (h *ShadowsocksServerHandler) DeleteShadowsocksServer(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.BadRequest(c, "Invalid shadowsocks server ID", err.Error())
		return
	}
	
	cmd := command.DeleteShadowsocksServerCommand{
		ID: id,
	}
	
	err = h.commandHandler.HandleDeleteShadowsocksServer(c.Request.Context(), cmd)
	if err != nil {
		if err == command.ErrShadowsocksServerNotFound {
			response.NotFound(c, "Shadowsocks server not found")
			return
		}
		response.InternalServerError(c, "Failed to delete shadowsocks server", err.Error())
		return
	}
	
	response.OK(c, "Shadowsocks server deleted successfully", nil)
}

// ChangeVisibility changes the visibility of a shadowsocks server
// @Summary Change server visibility
// @Description Change the visibility status of a shadowsocks server
// @Tags Shadowsocks Servers
// @Accept json
// @Produce json
// @Param id path int true "Shadowsocks server ID"
// @Param request body dto.ChangeVisibilityRequest true "Visibility change request"
// @Success 200 {object} response.StandardResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/shadowsocks-servers/{id}/visibility [patch]
func (h *ShadowsocksServerHandler) ChangeVisibility(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.BadRequest(c, "Invalid shadowsocks server ID", err.Error())
		return
	}
	
	var req dto.ChangeVisibilityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request", err.Error())
		return
	}
	
	cmd := command.ChangeShadowsocksServerVisibilityCommand{
		ID:        id,
		IsVisible: req.IsVisible,
	}
	
	err = h.commandHandler.HandleChangeShadowsocksServerVisibility(c.Request.Context(), cmd)
	if err != nil {
		if err == command.ErrShadowsocksServerNotFound {
			response.NotFound(c, "Shadowsocks server not found")
			return
		}
		response.InternalServerError(c, "Failed to change server visibility", err.Error())
		return
	}
	
	response.OK(c, "Server visibility changed successfully", nil)
}

// MoveToGroup moves a shadowsocks server to a different group
// @Summary Move server to group
// @Description Move a shadowsocks server to a different server group
// @Tags Shadowsocks Servers
// @Accept json
// @Produce json
// @Param id path int true "Shadowsocks server ID"
// @Param request body dto.MoveToGroupRequest true "Move to group request"
// @Success 200 {object} response.StandardResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 422 {object} response.UnprocessableEntityResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /admin/shadowsocks-servers/{id}/move [patch]
func (h *ShadowsocksServerHandler) MoveToGroup(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.BadRequest(c, "Invalid shadowsocks server ID", err.Error())
		return
	}
	
	var req dto.MoveToGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request", err.Error())
		return
	}
	
	cmd := command.MoveShadowsocksServerToGroupCommand{
		ID:      id,
		GroupID: req.GroupID,
	}
	
	err = h.commandHandler.HandleMoveShadowsocksServerToGroup(c.Request.Context(), cmd)
	if err != nil {
		if err == command.ErrShadowsocksServerNotFound {
			response.NotFound(c, "Shadowsocks server not found")
			return
		}
		if err == command.ErrInvalidServerGroup {
			response.UnprocessableEntity(c, "Invalid server group", err.Error())
			return
		}
		response.InternalServerError(c, "Failed to move server to group", err.Error())
		return
	}
	
	response.OK(c, "Server moved to group successfully", nil)
}