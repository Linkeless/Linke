package api

import (
	"strconv"

	"linke/internal/response"
	"linke/internal/server/handler/dto"
	"linke/internal/server/service/query"

	"github.com/gin-gonic/gin"
)

// ServerHandler handles user server requests
type ServerHandler struct {
	serverGroupQueryHandler     *query.ServerGroupQueryHandler
	shadowsocksServerQueryHandler *query.ShadowsocksServerQueryHandler
}

// NewServerHandler creates a new server handler
func NewServerHandler(
	serverGroupQueryHandler *query.ServerGroupQueryHandler,
	shadowsocksServerQueryHandler *query.ShadowsocksServerQueryHandler,
) *ServerHandler {
	return &ServerHandler{
		serverGroupQueryHandler:     serverGroupQueryHandler,
		shadowsocksServerQueryHandler: shadowsocksServerQueryHandler,
	}
}

// GetServerGroups gets all server groups for users
// @Summary Get server groups
// @Description Get all server groups available to users
// @Tags User Servers
// @Accept json
// @Produce json
// @Success 200 {object} response.StandardResponse{data=[]dto.ServerGroupResponse}
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /api/v1/servers/groups [get]
func (h *ServerHandler) GetServerGroups(c *gin.Context) {
	query := query.GetAllServerGroupsQuery{}
	
	groups, err := h.serverGroupQueryHandler.HandleGetAllServerGroups(c.Request.Context(), query)
	if err != nil {
		response.InternalServerError(c, "Failed to get server groups", err.Error())
		return
	}
	
	resp := dto.FromServerGroupDomainList(groups)
	response.OK(c, "Server groups retrieved successfully", resp)
}

// GetVisibleServers gets all visible shadowsocks servers
// @Summary Get visible servers
// @Description Get all visible shadowsocks servers available to users
// @Tags User Servers
// @Accept json
// @Produce json
// @Success 200 {object} response.StandardResponse{data=[]dto.ShadowsocksServerResponse}
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /api/v1/servers/shadowsocks [get]
func (h *ServerHandler) GetVisibleServers(c *gin.Context) {
	query := query.GetVisibleShadowsocksServersQuery{}
	
	servers, err := h.shadowsocksServerQueryHandler.HandleGetVisibleShadowsocksServers(c.Request.Context(), query)
	if err != nil {
		response.InternalServerError(c, "Failed to get visible servers", err.Error())
		return
	}
	
	resp := dto.FromShadowsocksServerDomainList(servers)
	response.OK(c, "Visible servers retrieved successfully", resp)
}

// GetServersByGroup gets visible servers by group
// @Summary Get servers by group
// @Description Get visible shadowsocks servers by server group
// @Tags User Servers
// @Accept json
// @Produce json
// @Param group_id path int true "Server group ID"
// @Success 200 {object} response.StandardResponse{data=[]dto.ShadowsocksServerResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /api/v1/servers/groups/{group_id}/shadowsocks [get]
func (h *ServerHandler) GetServersByGroup(c *gin.Context) {
	groupIDStr := c.Param("group_id")
	groupID, err := strconv.ParseUint(groupIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid group ID", err.Error())
		return
	}
	
	query := query.GetShadowsocksServersByGroupQuery{
		GroupID:     uint(groupID),
		VisibleOnly: true, // Only show visible servers to users
	}
	
	servers, err := h.shadowsocksServerQueryHandler.HandleGetShadowsocksServersByGroup(c.Request.Context(), query)
	if err != nil {
		response.InternalServerError(c, "Failed to get servers by group", err.Error())
		return
	}
	
	resp := dto.FromShadowsocksServerDomainList(servers)
	response.OK(c, "Servers retrieved successfully", resp)
}