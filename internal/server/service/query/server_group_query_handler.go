package query

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

// ServerGroupQueryHandler handles server group queries
type ServerGroupQueryHandler struct {
	serverGroupRepo repository.ServerGroupRepository
}

// NewServerGroupQueryHandler creates a new server group query handler
func NewServerGroupQueryHandler(serverGroupRepo repository.ServerGroupRepository) *ServerGroupQueryHandler {
	return &ServerGroupQueryHandler{
		serverGroupRepo: serverGroupRepo,
	}
}

// HandleGetServerGroup handles the get server group query
func (h *ServerGroupQueryHandler) HandleGetServerGroup(ctx context.Context, query GetServerGroupQuery) (*model.ServerGroup, error) {
	id, err := valueobject.NewServerGroupID(query.ID)
	if err != nil {
		return nil, err
	}
	
	group, err := h.serverGroupRepo.FindByID(ctx, id)
	if err != nil {
		return nil, ErrServerGroupNotFound
	}
	
	return group, nil
}

// HandleGetServerGroups handles the get server groups query
func (h *ServerGroupQueryHandler) HandleGetServerGroups(ctx context.Context, query GetServerGroupsQuery) ([]*model.ServerGroup, int64, error) {
	limit := query.Limit
	if limit == 0 {
		limit = 10 // Default limit
	}
	
	offset := query.Offset
	
	groups, total, err := h.serverGroupRepo.FindAll(ctx, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	
	return groups, total, nil
}

// HandleGetAllServerGroups handles the get all server groups query
func (h *ServerGroupQueryHandler) HandleGetAllServerGroups(ctx context.Context, query GetAllServerGroupsQuery) ([]*model.ServerGroup, error) {
	groups, _, err := h.serverGroupRepo.FindAll(ctx, 0, 0) // No pagination
	if err != nil {
		return nil, err
	}
	
	return groups, nil
}