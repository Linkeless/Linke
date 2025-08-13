package interfaces

import (
	"context"
	"linke/internal/domains/server/dto"
	"linke/internal/domains/server/entities"
)

// ShadowsocksServerService defines the interface for shadowsocks server operations
type ShadowsocksServerService interface {
	// Server CRUD operations
	CreateShadowsocksServer(ctx context.Context, req *dto.CreateShadowsocksServerRequest) (*entities.ShadowsocksServer, error)
	GetShadowsocksServer(ctx context.Context, serverID uint) (*entities.ShadowsocksServer, error)
	GetShadowsocksServerByID(ctx context.Context, serverID int) (*entities.ShadowsocksServer, error)
	UpdateShadowsocksServer(ctx context.Context, serverID uint, req *dto.UpdateShadowsocksServerRequest) (*entities.ShadowsocksServer, error)
	DeleteShadowsocksServer(ctx context.Context, serverID uint) error

	// Server listing and filtering
	GetShadowsocksServers(ctx context.Context, req *dto.GetShadowsocksServersRequest) ([]*entities.ShadowsocksServer, int64, error)
	GetServersByGroup(ctx context.Context, groupID uint) ([]*entities.ShadowsocksServer, error)
	GetVisibleServers(ctx context.Context, groupID *uint) ([]*entities.ShadowsocksServer, error)

	// Server management
	UpdateServerStatus(ctx context.Context, serverID uint, status string) error
	BulkUpdateServers(ctx context.Context, serverIDs []uint, updates map[string]any) error

	// Server health and monitoring
	CheckServerHealth(ctx context.Context, serverID uint) (map[string]any, error)
	GetServerStatistics(ctx context.Context, serverID uint) (map[string]any, error)
}
