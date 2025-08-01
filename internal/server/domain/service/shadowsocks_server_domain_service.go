package service

import (
	"context"
	"errors"

	"linke/internal/server/domain/model"
	"linke/internal/server/domain/repository"
	"linke/internal/server/domain/valueobject"
)

var (
	ErrServerGroupNotFound = errors.New("server group not found")
)

// ShadowsocksServerDomainService provides domain services for shadowsocks servers
type ShadowsocksServerDomainService struct {
	serverGroupRepo       repository.ServerGroupRepository
	shadowsocksServerRepo repository.ShadowsocksServerRepository
}

// NewShadowsocksServerDomainService creates a new shadowsocks server domain service
func NewShadowsocksServerDomainService(
	serverGroupRepo repository.ServerGroupRepository,
	shadowsocksServerRepo repository.ShadowsocksServerRepository,
) *ShadowsocksServerDomainService {
	return &ShadowsocksServerDomainService{
		serverGroupRepo:       serverGroupRepo,
		shadowsocksServerRepo: shadowsocksServerRepo,
	}
}

// ValidateServerGroup validates that a server group exists
func (s *ShadowsocksServerDomainService) ValidateServerGroup(ctx context.Context, groupID valueobject.ServerGroupID) error {
	_, err := s.serverGroupRepo.FindByID(ctx, groupID)
	if err != nil {
		return ErrServerGroupNotFound
	}
	
	return nil
}

// CanMoveToGroup checks if a server can be moved to a different group
func (s *ShadowsocksServerDomainService) CanMoveToGroup(ctx context.Context, groupID valueobject.ServerGroupID) error {
	return s.ValidateServerGroup(ctx, groupID)
}

// GetServersByGroup gets servers by group with visibility filter
func (s *ShadowsocksServerDomainService) GetServersByGroup(
	ctx context.Context,
	groupID valueobject.ServerGroupID,
	visibleOnly bool,
) ([]*model.ShadowsocksServer, error) {
	if visibleOnly {
		return s.shadowsocksServerRepo.FindVisibleByGroupID(ctx, groupID)
	}
	
	return s.shadowsocksServerRepo.FindByGroupID(ctx, groupID)
}