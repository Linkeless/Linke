package interfaces

import (
	"context"
	"linke/internal/domains/server/entities"
)

// ShadowsocksServerService defines the interface for shadowsocks server operations
type ShadowsocksServerService interface {
	// Server CRUD operations
	CreateShadowsocksServer(ctx context.Context, req *CreateShadowsocksServerRequest) (*entities.ShadowsocksServer, error)
	GetShadowsocksServer(ctx context.Context, serverID uint) (*entities.ShadowsocksServer, error)
	GetShadowsocksServerByID(ctx context.Context, serverID int) (*entities.ShadowsocksServer, error)
	UpdateShadowsocksServer(ctx context.Context, serverID uint, req *UpdateShadowsocksServerRequest) (*entities.ShadowsocksServer, error)
	DeleteShadowsocksServer(ctx context.Context, serverID uint) error

	// Server listing and filtering
	GetShadowsocksServers(ctx context.Context, req *GetShadowsocksServersRequest) ([]*entities.ShadowsocksServer, int64, error)
	GetServersByGroup(ctx context.Context, groupID uint) ([]*entities.ShadowsocksServer, error)
	GetVisibleServers(ctx context.Context, groupID *uint) ([]*entities.ShadowsocksServer, error)

	// Server management
	UpdateServerStatus(ctx context.Context, serverID uint, status string) error
	BulkUpdateServers(ctx context.Context, serverIDs []uint, updates map[string]any) error

	// Server health and monitoring
	CheckServerHealth(ctx context.Context, serverID uint) (map[string]any, error)
	GetServerStatistics(ctx context.Context, serverID uint) (map[string]any, error)
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

// GetShadowsocksServersRequest represents the request to get shadowsocks servers with filters
type GetShadowsocksServersRequest struct {
	GroupID *uint `json:"group_id,omitempty"`
	Show    *int  `json:"show,omitempty"`
	Limit   int   `json:"limit,omitempty"`
	Offset  int   `json:"offset,omitempty"`
}
