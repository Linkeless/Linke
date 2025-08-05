package interfaces

import (
	"context"
	"linke/internal/domains/server/entities"
	"linke/internal/shared/framework"
)

// ServerGroupRepository defines the interface for server group data access operations
type ServerGroupRepository interface {
	framework.GenericRepository[entities.ServerGroup, uint]

	// Status operations
	ListActive(ctx context.Context, limit, offset int) ([]*entities.ServerGroup, int64, error)
	UpdateActiveStatus(ctx context.Context, id uint, isActive bool) error
}

// ShadowsocksServerRepository defines the interface for shadowsocks server data access operations
type ShadowsocksServerRepository interface {
	framework.GenericRepository[entities.ShadowsocksServer, int]

	// Group operations
	ListByGroup(ctx context.Context, groupID uint, limit, offset int) ([]*entities.ShadowsocksServer, int64, error)
	CountByGroup(ctx context.Context, groupID uint) (int64, error)

	// Status operations  
	ListActive(ctx context.Context, limit, offset int) ([]*entities.ShadowsocksServer, int64, error)

	// Location operations
	ListByLocation(ctx context.Context, country, region string, limit, offset int) ([]*entities.ShadowsocksServer, int64, error)

	// Statistics
	GetLocationStats(ctx context.Context) (map[string]int64, error)
	GetStatusStats(ctx context.Context) (map[string]int64, error)
}
