package interfaces

import (
	"context"
	"linke/internal/domains/server/entities"
)

// ServerGroupService defines the interface for server group operations
type ServerGroupService interface {
	// Group CRUD operations
	CreateServerGroup(ctx context.Context, req *CreateServerGroupRequest) (*entities.ServerGroup, error)
	GetServerGroup(ctx context.Context, groupID uint) (*entities.ServerGroup, error)
	UpdateServerGroup(ctx context.Context, groupID uint, req *UpdateServerGroupRequest) (*entities.ServerGroup, error)
	DeleteServerGroup(ctx context.Context, groupID uint) error

	// Group listing and management
	GetServerGroups(ctx context.Context, req *GetServerGroupsRequest) ([]*entities.ServerGroup, int64, error)
	GetAllServerGroups(ctx context.Context) ([]*entities.ServerGroup, error)

	// Group-server relationship management
	GetGroupServers(ctx context.Context, groupID uint) ([]*entities.ShadowsocksServer, error)
	GetGroupServerCount(ctx context.Context, groupID uint) (int64, error)

	// Group statistics
	GetGroupStatistics(ctx context.Context, groupID uint) (map[string]interface{}, error)
}

// CreateServerGroupRequest represents the request to create a server group
type CreateServerGroupRequest struct {
	Name string `json:"name" binding:"required,max=255" example:"Asia Pacific"`
}

// UpdateServerGroupRequest represents the request to update a server group
type UpdateServerGroupRequest struct {
	Name *string `json:"name,omitempty" binding:"omitempty,max=255" example:"Europe"`
}

// GetServerGroupsRequest represents the request to get server groups with filters
type GetServerGroupsRequest struct {
	Limit  int `json:"limit,omitempty" binding:"omitempty,min=1,max=100" example:"10"`
	Offset int `json:"offset,omitempty" binding:"omitempty,min=0" example:"0"`
}