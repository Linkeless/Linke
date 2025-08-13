package interfaces

import (
	"context"
	"linke/internal/domains/server/dto"
	"linke/internal/domains/server/entities"
)

// ServerGroupService defines the interface for server group operations
type ServerGroupService interface {
	// Group CRUD operations
	CreateServerGroup(ctx context.Context, req *dto.CreateServerGroupRequest) (*entities.ServerGroup, error)
	GetServerGroup(ctx context.Context, groupID uint) (*entities.ServerGroup, error)
	UpdateServerGroup(ctx context.Context, groupID uint, req *dto.UpdateServerGroupRequest) (*entities.ServerGroup, error)
	DeleteServerGroup(ctx context.Context, groupID uint) error

	// Group listing and management
	GetServerGroups(ctx context.Context, req *dto.GetServerGroupsRequest) ([]*entities.ServerGroup, int64, error)
	GetAllServerGroups(ctx context.Context) ([]*entities.ServerGroup, error)

	// Group-server relationship management
	GetGroupServers(ctx context.Context, groupID uint) ([]*entities.ShadowsocksServer, error)
	GetGroupServerCount(ctx context.Context, groupID uint) (int64, error)

	// Group statistics
	GetGroupStatistics(ctx context.Context, groupID uint) (map[string]any, error)
}
