package repository

import (
	"context"
	"errors"

	"linke/internal/server/domain/model"
	"linke/internal/server/domain/valueobject"
)

// Repository errors
var (
	ErrServerGroupNotFound = errors.New("server group not found")
	ErrShadowsocksServerNotFound = errors.New("shadowsocks server not found")
)

// ServerGroupRepository defines the repository interface for server groups
type ServerGroupRepository interface {
	// Save saves a server group
	Save(ctx context.Context, group *model.ServerGroup) error
	
	// FindByID finds a server group by ID
	FindByID(ctx context.Context, id valueobject.ServerGroupID) (*model.ServerGroup, error)
	
	// FindByName finds a server group by name
	FindByName(ctx context.Context, name valueobject.ServerGroupName) (*model.ServerGroup, error)
	
	// FindAll finds all server groups with pagination
	FindAll(ctx context.Context, limit, offset int) ([]*model.ServerGroup, int64, error)
	
	// ExistsByName checks if a server group with the given name exists
	ExistsByName(ctx context.Context, name valueobject.ServerGroupName) (bool, error)
	
	// ExistsByNameExcludingID checks if a server group with the given name exists, excluding a specific ID
	ExistsByNameExcludingID(ctx context.Context, name valueobject.ServerGroupName, excludeID valueobject.ServerGroupID) (bool, error)
	
	// Delete deletes a server group
	Delete(ctx context.Context, id valueobject.ServerGroupID) error
	
	// Count returns the total count of server groups
	Count(ctx context.Context) (int64, error)
}