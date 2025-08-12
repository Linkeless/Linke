package dto

import "linke/internal/domains/server/constants"

// ServerGroup相关请求结构体

// CreateServerGroupRequest 创建服务器组请求
type CreateServerGroupRequest struct {
	Name string `json:"name" binding:"required" validate:"max=255" example:"Asia Pacific"`
}

// UpdateServerGroupRequest 更新服务器组请求
type UpdateServerGroupRequest struct {
	Name *string `json:"name,omitempty" binding:"omitempty" validate:"omitempty,max=255" example:"Europe"`
}

// PatchServerGroupRequest 部分更新服务器组请求
type PatchServerGroupRequest struct {
	Name *string `json:"name,omitempty" example:"Europe"`
}

// GetServerGroupsRequest 获取服务器组列表请求
type GetServerGroupsRequest struct {
	Limit  int `json:"limit,omitempty" binding:"omitempty,min=1,max=100" example:"10"`
	Offset int `json:"offset,omitempty" binding:"omitempty,min=0" example:"0"`
}

// BatchServerGroupIDsRequest 批量操作服务器组ID请求
type BatchServerGroupIDsRequest struct {
	IDs []uint `json:"ids" binding:"required,min=1,max=100" example:"[1,2,3]"`
}

// ShadowsocksServer相关请求结构体

// CreateShadowsocksServerRequest 创建Shadowsocks服务器请求
type CreateShadowsocksServerRequest struct {
	GroupID      uint    `json:"group_id" binding:"required" example:"1"`
	RouteID      string  `json:"route_id,omitempty" validate:"max=255" example:"route-1"`
	ParentID     *int    `json:"parent_id,omitempty" example:"1"`
	Name         string  `json:"name" binding:"required" validate:"max=255" example:"US-01"`
	Tags         string  `json:"tags,omitempty" validate:"max=255" example:"premium,fast"`
	Host         string  `json:"host" binding:"required" validate:"max=255" example:"us01.example.com"`
	Port         int     `json:"port" binding:"required" validate:"min=1,max=65535" example:"443"`
	ServerPort   int     `json:"server_port" binding:"required" validate:"min=1,max=65535" example:"8388"`
	Cipher       string  `json:"cipher" binding:"required" validate:"max=255" example:"aes-256-gcm"`
	Obfs         string  `json:"obfs,omitempty" validate:"max=11" example:"tls"`
	ObfsSettings string  `json:"obfs_settings,omitempty" validate:"max=255" example:"obfs=tls"`
	Excludes     string  `json:"excludes,omitempty" validate:"max=500" example:"192.168.0.0/16"`
	IPs          string  `json:"ips,omitempty" validate:"max=255" example:"0.0.0.0/0"`
	Rate         float64 `json:"rate" binding:"required,min=0.1" example:"1.0"`
	Show         int     `json:"show" binding:"min=0,max=1" example:"1"`
	Sort         *int    `json:"sort,omitempty" example:"1"`
}

// UpdateShadowsocksServerRequest 更新Shadowsocks服务器请求
type UpdateShadowsocksServerRequest struct {
	GroupID      *uint    `json:"group_id,omitempty" example:"1"`
	RouteID      *string  `json:"route_id,omitempty" binding:"omitempty" validate:"omitempty,max=255" example:"route-1"`
	ParentID     *int     `json:"parent_id,omitempty" example:"1"`
	Name         *string  `json:"name,omitempty" binding:"omitempty" validate:"omitempty,max=255" example:"US-01"`
	Tags         *string  `json:"tags,omitempty" binding:"omitempty" validate:"omitempty,max=255" example:"premium,fast"`
	Host         *string  `json:"host,omitempty" binding:"omitempty" validate:"omitempty,max=255" example:"us01.example.com"`
	Port         *int     `json:"port,omitempty" binding:"omitempty" validate:"omitempty,min=1,max=65535" example:"443"`
	ServerPort   *int     `json:"server_port,omitempty" binding:"omitempty" validate:"omitempty,min=1,max=65535" example:"8388"`
	Cipher       *string  `json:"cipher,omitempty" binding:"omitempty" validate:"omitempty,max=255" example:"aes-256-gcm"`
	Obfs         *string  `json:"obfs,omitempty" binding:"omitempty" validate:"omitempty,max=11" example:"tls"`
	ObfsSettings *string  `json:"obfs_settings,omitempty" binding:"omitempty" validate:"omitempty,max=255" example:"obfs=tls"`
	Excludes     *string  `json:"excludes,omitempty" validate:"omitempty,max=500" example:"192.168.0.0/16"`
	IPs          *string  `json:"ips,omitempty" binding:"omitempty" validate:"omitempty,max=255" example:"0.0.0.0/0"`
	Rate         *float64 `json:"rate,omitempty" binding:"omitempty" validate:"omitempty,min=0.1" example:"1.0"`
	Show         *int     `json:"show,omitempty" binding:"omitempty" validate:"omitempty,min=0,max=1" example:"1"`
	Sort         *int     `json:"sort,omitempty" example:"1"`
}

// PatchShadowsocksServerRequest 部分更新Shadowsocks服务器请求
type PatchShadowsocksServerRequest struct {
	GroupID      *uint    `json:"group_id,omitempty" example:"1"`
	RouteID      *string  `json:"route_id,omitempty" example:"route-1"`
	ParentID     *int     `json:"parent_id,omitempty" example:"1"`
	Name         *string  `json:"name,omitempty" example:"US-01"`
	Tags         *string  `json:"tags,omitempty" example:"premium,fast"`
	Host         *string  `json:"host,omitempty" example:"us01.example.com"`
	Port         *int     `json:"port,omitempty" example:"443"`
	ServerPort   *int     `json:"server_port,omitempty" example:"8388"`
	Cipher       *string  `json:"cipher,omitempty" example:"aes-256-gcm"`
	Obfs         *string  `json:"obfs,omitempty" example:"tls"`
	ObfsSettings *string  `json:"obfs_settings,omitempty" example:"obfs=tls"`
	Excludes     *string  `json:"excludes,omitempty" example:"192.168.0.0/16"`
	IPs          *string  `json:"ips,omitempty" example:"0.0.0.0/0"`
	Rate         *float64 `json:"rate,omitempty" example:"1.0"`
	Show         *int     `json:"show,omitempty" example:"1"`
	Sort         *int     `json:"sort,omitempty" example:"1"`
}

// GetShadowsocksServersRequest 获取Shadowsocks服务器列表请求
type GetShadowsocksServersRequest struct {
	GroupID   *uint  `json:"group_id,omitempty" form:"group_id" example:"1"`
	Show      *int   `json:"show,omitempty" form:"show" binding:"omitempty,min=0,max=1" example:"1"`
	Name      string `json:"name,omitempty" form:"name" example:"US"`
	Tags      string `json:"tags,omitempty" form:"tags" example:"premium"`
	SortBy    string `json:"sort_by,omitempty" form:"sort_by" example:"sort"`
	SortOrder string `json:"sort_order,omitempty" form:"sort_order" example:"asc"`
	Limit     int    `json:"limit,omitempty" form:"limit" binding:"omitempty,min=1,max=100" example:"10"`
	Offset    int    `json:"offset,omitempty" form:"offset" binding:"omitempty,min=0" example:"0"`
}

// BatchShadowsocksServerIDsRequest 批量操作Shadowsocks服务器ID请求
type BatchShadowsocksServerIDsRequest struct {
	IDs []uint `json:"ids" binding:"required,min=1,max=100" example:"[1,2,3]"`
}

// 验证辅助方法

// ValidateShow 验证Show字段值
func ValidateShow(show int) bool {
	return show == constants.ServerShowVisible || show == constants.ServerShowHidden
}

// IsVisible 检查Show值是否表示可见
func IsVisible(show int) bool {
	return show == constants.ServerShowVisible
}
