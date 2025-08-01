package dto

import (
	"time"

	"linke/internal/server/domain/model"
)

// ServerGroup DTOs

// ServerGroupResponse represents the server group response DTO
type ServerGroupResponse struct {
	ID        uint   `json:"id" example:"1"`
	Name      string `json:"name" example:"Asia Pacific"`
	CreatedAt string `json:"created_at" example:"2024-01-01T00:00:00Z"`
	UpdatedAt string `json:"updated_at" example:"2024-01-01T00:00:00Z"`
}

// CreateServerGroupRequest represents the request to create a server group
type CreateServerGroupRequest struct {
	Name string `json:"name" binding:"required,max=255" example:"Asia Pacific"`
}

// UpdateServerGroupRequest represents the request to update a server group
type UpdateServerGroupRequest struct {
	Name *string `json:"name,omitempty" binding:"omitempty,max=255" example:"Europe"`
}

// ServerGroupListResponse represents the server group list response
type ServerGroupListResponse struct {
	Groups []ServerGroupResponse `json:"groups"`
	Total  int64                 `json:"total"`
}

// Shadowsocks Server DTOs

// ShadowsocksServerResponse represents the shadowsocks server response DTO
type ShadowsocksServerResponse struct {
	ID           int     `json:"id" example:"1"`
	GroupID      uint    `json:"group_id" example:"1"`
	RouteID      string  `json:"route_id,omitempty" example:"route-1"`
	ParentID     *int    `json:"parent_id,omitempty" example:"1"`
	Name         string  `json:"name" example:"US-01"`
	Tags         string  `json:"tags,omitempty" example:"premium,fast"`
	Host         string  `json:"host" example:"us01.example.com"`
	Port         int     `json:"port" example:"443"`
	ServerPort   int     `json:"server_port" example:"8388"`
	Cipher       string  `json:"cipher" example:"aes-256-gcm"`
	Obfs         string  `json:"obfs,omitempty" example:"tls"`
	ObfsSettings string  `json:"obfs_settings,omitempty" example:"obfs=tls"`
	Excludes     string  `json:"excludes,omitempty" example:"192.168.0.0/16"`
	IPs          string  `json:"ips,omitempty" example:"0.0.0.0/0"`
	Rate         float64 `json:"rate" example:"1.0"`
	Show         int     `json:"show" example:"1"`
	Sort         *int    `json:"sort,omitempty" example:"1"`
	CreatedAt    string  `json:"created_at" example:"2024-01-01T00:00:00Z"`
	UpdatedAt    string  `json:"updated_at" example:"2024-01-01T00:00:00Z"`
	
	// Relationship data
	ServerGroup *ServerGroupResponse `json:"server_group,omitempty"`
}

// CreateShadowsocksServerRequest represents the request to create a shadowsocks server
type CreateShadowsocksServerRequest struct {
	GroupID      uint    `json:"group_id" binding:"required"`
	RouteID      string  `json:"route_id,omitempty" binding:"max=255"`
	ParentID     *int    `json:"parent_id,omitempty"`
	Name         string  `json:"name" binding:"required,max=255"`
	Tags         string  `json:"tags,omitempty" binding:"max=255"`
	Host         string  `json:"host" binding:"required,max=255"`
	Port         int     `json:"port" binding:"required,min=1,max=65535"`
	ServerPort   int     `json:"server_port" binding:"required,min=1,max=65535"`
	Cipher       string  `json:"cipher" binding:"required,max=255"`
	Obfs         string  `json:"obfs,omitempty" binding:"max=11"`
	ObfsSettings string  `json:"obfs_settings,omitempty" binding:"max=255"`
	Excludes     string  `json:"excludes,omitempty"`
	IPs          string  `json:"ips,omitempty" binding:"max=255"`
	Rate         float64 `json:"rate" binding:"required,min=0.1"`
	Show         int     `json:"show" binding:"min=0,max=1"`
	Sort         *int    `json:"sort,omitempty"`
}

// UpdateShadowsocksServerRequest represents the request to update a shadowsocks server
type UpdateShadowsocksServerRequest struct {
	GroupID      *uint    `json:"group_id,omitempty"`
	RouteID      *string  `json:"route_id,omitempty" binding:"omitempty,max=255"`
	ParentID     *int     `json:"parent_id,omitempty"`
	Name         *string  `json:"name,omitempty" binding:"omitempty,max=255"`
	Tags         *string  `json:"tags,omitempty" binding:"omitempty,max=255"`
	Host         *string  `json:"host,omitempty" binding:"omitempty,max=255"`
	Port         *int     `json:"port,omitempty" binding:"omitempty,min=1,max=65535"`
	ServerPort   *int     `json:"server_port,omitempty" binding:"omitempty,min=1,max=65535"`
	Cipher       *string  `json:"cipher,omitempty" binding:"omitempty,max=255"`
	Obfs         *string  `json:"obfs,omitempty" binding:"omitempty,max=11"`
	ObfsSettings *string  `json:"obfs_settings,omitempty" binding:"omitempty,max=255"`
	Excludes     *string  `json:"excludes,omitempty"`
	IPs          *string  `json:"ips,omitempty" binding:"omitempty,max=255"`
	Rate         *float64 `json:"rate,omitempty" binding:"omitempty,min=0.1"`
	Show         *int     `json:"show,omitempty" binding:"omitempty,min=0,max=1"`
	Sort         *int     `json:"sort,omitempty"`
}

// ShadowsocksServerListResponse represents the shadowsocks server list response
type ShadowsocksServerListResponse struct {
	Servers []ShadowsocksServerResponse `json:"servers"`
	Total   int64                       `json:"total"`
}

// ChangeVisibilityRequest represents the request to change server visibility
type ChangeVisibilityRequest struct {
	IsVisible bool `json:"is_visible"`
}

// MoveToGroupRequest represents the request to move server to a group
type MoveToGroupRequest struct {
	GroupID uint `json:"group_id" binding:"required"`
}

// Conversion methods

// FromServerGroupDomain converts domain model to DTO
func FromServerGroupDomain(group *model.ServerGroup) ServerGroupResponse {
	return ServerGroupResponse{
		ID:        group.ID().Value(),
		Name:      group.Name().Value(),
		CreatedAt: group.CreatedAt().Format(time.RFC3339),
		UpdatedAt: group.UpdatedAt().Format(time.RFC3339),
	}
}

// FromShadowsocksServerDomain converts domain model to DTO
func FromShadowsocksServerDomain(server *model.ShadowsocksServer) ShadowsocksServerResponse {
	resp := ShadowsocksServerResponse{
		ID:           server.ID().Value(),
		GroupID:      server.GroupID().Value(),
		RouteID:      server.RouteID(),
		ParentID:     server.ParentID(),
		Name:         server.Name().Value(),
		Tags:         server.Tags(),
		Host:         server.Host().Value(),
		Port:         server.Port().Value(),
		ServerPort:   server.ServerPort().Value(),
		Cipher:       server.Cipher().Value(),
		Obfs:         server.Obfs(),
		ObfsSettings: server.ObfsSettings(),
		Excludes:     server.Excludes(),
		IPs:          server.IPs(),
		Rate:         server.Rate().Value(),
		Sort:         server.Sort(),
		CreatedAt:    server.CreatedAt().Format(time.RFC3339),
		UpdatedAt:    server.UpdatedAt().Format(time.RFC3339),
	}
	
	if server.IsVisible() {
		resp.Show = 1
	} else {
		resp.Show = 0
	}
	
	return resp
}

// FromShadowsocksServerDomainList converts domain model list to DTO list
func FromShadowsocksServerDomainList(servers []*model.ShadowsocksServer) []ShadowsocksServerResponse {
	responses := make([]ShadowsocksServerResponse, len(servers))
	for i, server := range servers {
		responses[i] = FromShadowsocksServerDomain(server)
	}
	return responses
}

// FromServerGroupDomainList converts domain model list to DTO list
func FromServerGroupDomainList(groups []*model.ServerGroup) []ServerGroupResponse {
	responses := make([]ServerGroupResponse, len(groups))
	for i, group := range groups {
		responses[i] = FromServerGroupDomain(group)
	}
	return responses
}