package query

import (
	"context"
	"errors"

	"linke/internal/server/domain/model"
	"linke/internal/server/domain/repository"
	"linke/internal/server/domain/valueobject"
)

var (
	ErrShadowsocksServerNotFound = errors.New("shadowsocks server not found")
)

// ShadowsocksServerQueryHandler handles shadowsocks server queries
type ShadowsocksServerQueryHandler struct {
	shadowsocksServerRepo repository.ShadowsocksServerRepository
}

// NewShadowsocksServerQueryHandler creates a new shadowsocks server query handler
func NewShadowsocksServerQueryHandler(shadowsocksServerRepo repository.ShadowsocksServerRepository) *ShadowsocksServerQueryHandler {
	return &ShadowsocksServerQueryHandler{
		shadowsocksServerRepo: shadowsocksServerRepo,
	}
}

// HandleGetShadowsocksServer handles the get shadowsocks server query
func (h *ShadowsocksServerQueryHandler) HandleGetShadowsocksServer(ctx context.Context, query GetShadowsocksServerQuery) (*model.ShadowsocksServer, error) {
	id, err := valueobject.NewServerID(query.ID)
	if err != nil {
		return nil, err
	}
	
	server, err := h.shadowsocksServerRepo.FindByID(ctx, id)
	if err != nil {
		return nil, ErrShadowsocksServerNotFound
	}
	
	return server, nil
}

// HandleGetShadowsocksServers handles the get shadowsocks servers query
func (h *ShadowsocksServerQueryHandler) HandleGetShadowsocksServers(ctx context.Context, query GetShadowsocksServersQuery) ([]*model.ShadowsocksServer, int64, error) {
	limit := query.Limit
	if limit == 0 {
		limit = 10 // Default limit
	}
	
	offset := query.Offset
	
	// Build filters
	filters := repository.ShadowsocksServerFilters{}
	
	if query.GroupID != nil {
		groupID, err := valueobject.NewServerGroupID(*query.GroupID)
		if err != nil {
			return nil, 0, err
		}
		filters.GroupID = &groupID
	}
	
	if query.Show != nil {
		isVisible := *query.Show == 1
		filters.IsVisible = &isVisible
	}
	
	if query.Tags != "" {
		filters.Tags = query.Tags
	}
	
	if query.Cipher != "" {
		cipher, err := valueobject.NewCipher(query.Cipher)
		if err != nil {
			return nil, 0, err
		}
		filters.Cipher = &cipher
	}
	
	servers, total, err := h.shadowsocksServerRepo.FindAll(ctx, filters, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	
	return servers, total, nil
}

// HandleGetShadowsocksServersByGroup handles the get shadowsocks servers by group query
func (h *ShadowsocksServerQueryHandler) HandleGetShadowsocksServersByGroup(ctx context.Context, query GetShadowsocksServersByGroupQuery) ([]*model.ShadowsocksServer, error) {
	groupID, err := valueobject.NewServerGroupID(query.GroupID)
	if err != nil {
		return nil, err
	}
	
	if query.VisibleOnly {
		return h.shadowsocksServerRepo.FindVisibleByGroupID(ctx, groupID)
	}
	
	return h.shadowsocksServerRepo.FindByGroupID(ctx, groupID)
}

// HandleGetVisibleShadowsocksServers handles the get visible shadowsocks servers query
func (h *ShadowsocksServerQueryHandler) HandleGetVisibleShadowsocksServers(ctx context.Context, query GetVisibleShadowsocksServersQuery) ([]*model.ShadowsocksServer, error) {
	return h.shadowsocksServerRepo.FindVisible(ctx)
}

// HandleGetShadowsocksServerCount handles the get shadowsocks server count query
func (h *ShadowsocksServerQueryHandler) HandleGetShadowsocksServerCount(ctx context.Context, query GetShadowsocksServerCountQuery) (int64, error) {
	if query.GroupID != nil {
		groupID, err := valueobject.NewServerGroupID(*query.GroupID)
		if err != nil {
			return 0, err
		}
		return h.shadowsocksServerRepo.CountByGroupID(ctx, groupID)
	}
	
	return h.shadowsocksServerRepo.Count(ctx)
}