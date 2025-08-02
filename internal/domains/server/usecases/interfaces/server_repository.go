package interfaces

import (
	"context"
	"linke/internal/domains/server/entities"
)

// ServerGroupRepository defines the interface for server group data access operations
type ServerGroupRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, group *entities.ServerGroup) error
	GetByID(ctx context.Context, id uint) (*entities.ServerGroup, error)
	Update(ctx context.Context, group *entities.ServerGroup) error
	Delete(ctx context.Context, id uint) error

	// Status operations
	ListActive(ctx context.Context, limit, offset int) ([]*entities.ServerGroup, int64, error)
	UpdateStatus(ctx context.Context, id uint, isActive bool) error

	// List operations
	List(ctx context.Context, limit, offset int) ([]*entities.ServerGroup, int64, error)
	CountTotal(ctx context.Context) (int64, error)
}

// ShadowsocksServerRepository defines the interface for shadowsocks server data access operations
type ShadowsocksServerRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, server *entities.ShadowsocksServer) error
	GetByID(ctx context.Context, id uint) (*entities.ShadowsocksServer, error)
	Update(ctx context.Context, server *entities.ShadowsocksServer) error
	Delete(ctx context.Context, id uint) error

	// Group operations
	ListByGroup(ctx context.Context, groupID uint, limit, offset int) ([]*entities.ShadowsocksServer, int64, error)
	CountByGroup(ctx context.Context, groupID uint) (int64, error)

	// Status operations
	ListActive(ctx context.Context, limit, offset int) ([]*entities.ShadowsocksServer, int64, error)
	ListByStatus(ctx context.Context, status string, limit, offset int) ([]*entities.ShadowsocksServer, int64, error)
	UpdateStatus(ctx context.Context, id uint, status string) error

	// Location operations
	ListByLocation(ctx context.Context, country, region string, limit, offset int) ([]*entities.ShadowsocksServer, int64, error)

	// List operations
	List(ctx context.Context, limit, offset int) ([]*entities.ShadowsocksServer, int64, error)
	CountTotal(ctx context.Context) (int64, error)

	// Statistics
	GetLocationStats(ctx context.Context) (map[string]int64, error)
	GetStatusStats(ctx context.Context) (map[string]int64, error)
}