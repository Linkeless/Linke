package service

import (
	"context"
	"errors"

	"linke/internal/server/domain/repository"
	"linke/internal/server/domain/valueobject"
)

var (
	ErrServerGroupHasServers = errors.New("server group cannot be deleted because it contains servers")
)

// ServerGroupDomainService provides domain services for server groups
type ServerGroupDomainService struct {
	serverGroupRepo        repository.ServerGroupRepository
	shadowsocksServerRepo  repository.ShadowsocksServerRepository
}

// NewServerGroupDomainService creates a new server group domain service
func NewServerGroupDomainService(
	serverGroupRepo repository.ServerGroupRepository,
	shadowsocksServerRepo repository.ShadowsocksServerRepository,
) *ServerGroupDomainService {
	return &ServerGroupDomainService{
		serverGroupRepo:       serverGroupRepo,
		shadowsocksServerRepo: shadowsocksServerRepo,
	}
}

// CanDeleteServerGroup checks if a server group can be deleted
func (s *ServerGroupDomainService) CanDeleteServerGroup(ctx context.Context, groupID valueobject.ServerGroupID) error {
	count, err := s.shadowsocksServerRepo.CountByGroupID(ctx, groupID)
	if err != nil {
		return err
	}
	
	if count > 0 {
		return ErrServerGroupHasServers
	}
	
	return nil
}

// IsServerGroupNameUnique checks if a server group name is unique
func (s *ServerGroupDomainService) IsServerGroupNameUnique(ctx context.Context, name valueobject.ServerGroupName) (bool, error) {
	exists, err := s.serverGroupRepo.ExistsByName(ctx, name)
	if err != nil {
		return false, err
	}
	
	return !exists, nil
}

// IsServerGroupNameUniqueForUpdate checks if a server group name is unique for update
func (s *ServerGroupDomainService) IsServerGroupNameUniqueForUpdate(
	ctx context.Context,
	name valueobject.ServerGroupName,
	excludeID valueobject.ServerGroupID,
) (bool, error) {
	exists, err := s.serverGroupRepo.ExistsByNameExcludingID(ctx, name, excludeID)
	if err != nil {
		return false, err
	}
	
	return !exists, nil
}