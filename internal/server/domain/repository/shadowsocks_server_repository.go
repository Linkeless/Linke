package repository

import (
	"context"

	"linke/internal/server/domain/model"
	"linke/internal/server/domain/valueobject"
)

// ShadowsocksServerRepository defines the repository interface for shadowsocks servers
type ShadowsocksServerRepository interface {
	// Save saves a shadowsocks server
	Save(ctx context.Context, server *model.ShadowsocksServer) error
	
	// FindByID finds a shadowsocks server by ID
	FindByID(ctx context.Context, id valueobject.ServerID) (*model.ShadowsocksServer, error)
	
	// FindByGroupID finds shadowsocks servers by group ID
	FindByGroupID(ctx context.Context, groupID valueobject.ServerGroupID) ([]*model.ShadowsocksServer, error)
	
	// FindVisibleByGroupID finds visible shadowsocks servers by group ID
	FindVisibleByGroupID(ctx context.Context, groupID valueobject.ServerGroupID) ([]*model.ShadowsocksServer, error)
	
	// FindAll finds all shadowsocks servers with filters and pagination
	FindAll(ctx context.Context, filters ShadowsocksServerFilters, limit, offset int) ([]*model.ShadowsocksServer, int64, error)
	
	// FindVisible finds all visible shadowsocks servers
	FindVisible(ctx context.Context) ([]*model.ShadowsocksServer, error)
	
	// Delete deletes a shadowsocks server
	Delete(ctx context.Context, id valueobject.ServerID) error
	
	// Count returns the total count of shadowsocks servers
	Count(ctx context.Context) (int64, error)
	
	// CountByGroupID returns the count of shadowsocks servers in a group
	CountByGroupID(ctx context.Context, groupID valueobject.ServerGroupID) (int64, error)
}

// ShadowsocksServerFilters represents filters for shadowsocks server queries
type ShadowsocksServerFilters struct {
	GroupID   *valueobject.ServerGroupID
	IsVisible *bool
	Tags      string
	Cipher    *valueobject.Cipher
}