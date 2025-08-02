package implementations

import (
	"context"
	"fmt"

	"linke/internal/shared/logger"
	"linke/internal/domains/server/entities"
	"linke/internal/shared/database"

	"gorm.io/gorm"
)

type ServerGroupService struct {
	db *database.Database
}

func NewServerGroupService(db *database.Database) *ServerGroupService {
	return &ServerGroupService{
		db: db,
	}
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

// CreateServerGroup creates a new server group
func (s *ServerGroupService) CreateServerGroup(ctx context.Context, req *CreateServerGroupRequest) (*entities.ServerGroup, error) {
	// Check if server group with the same name already exists
	var existingGroup entities.ServerGroup
	if err := s.db.DB.WithContext(ctx).Where("name = ?", req.Name).First(&existingGroup).Error; err == nil {
		return nil, fmt.Errorf("server group with name '%s' already exists", req.Name)
	} else if err != gorm.ErrRecordNotFound {
		logger.Error("Failed to check existing server group", logger.Error2("error", err))
		return nil, fmt.Errorf("failed to check existing server group: %w", err)
	}

	group := &entities.ServerGroup{
		Name: req.Name,
	}

	if err := s.db.DB.WithContext(ctx).Create(group).Error; err != nil {
		logger.Error("Failed to create server group", 
			logger.String("name", req.Name),
			logger.Error2("error", err))
		return nil, fmt.Errorf("failed to create server group: %w", err)
	}

	logger.Info("Server group created successfully", 
		logger.Uint("group_id", group.ID),
		logger.String("name", group.Name))

	return group, nil
}

// GetServerGroup gets a server group by ID
func (s *ServerGroupService) GetServerGroup(ctx context.Context, id uint) (*entities.ServerGroup, error) {
	var group entities.ServerGroup
	if err := s.db.DB.WithContext(ctx).First(&group, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("server group not found")
		}
		logger.Error("Failed to get server group", logger.Error2("error", err), logger.Uint("group_id", id))
		return nil, fmt.Errorf("failed to get server group: %w", err)
	}

	return &group, nil
}

// GetServerGroups gets server groups with filtering and pagination
func (s *ServerGroupService) GetServerGroups(ctx context.Context, req *GetServerGroupsRequest) ([]*entities.ServerGroup, int64, error) {
	var groups []*entities.ServerGroup
	var total int64

	query := s.db.DB.WithContext(ctx).Model(&entities.ServerGroup{})

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		logger.Error("Failed to count server groups", logger.Error2("error", err))
		return nil, 0, fmt.Errorf("failed to count server groups: %w", err)
	}

	// Apply pagination and ordering
	query = query.Order("created_at DESC")

	if req.Limit > 0 {
		query = query.Limit(req.Limit)
	}

	if req.Offset > 0 {
		query = query.Offset(req.Offset)
	}

	if err := query.Find(&groups).Error; err != nil {
		logger.Error("Failed to get server groups", logger.Error2("error", err))
		return nil, 0, fmt.Errorf("failed to get server groups: %w", err)
	}

	return groups, total, nil
}

// UpdateServerGroup updates a server group
func (s *ServerGroupService) UpdateServerGroup(ctx context.Context, id uint, req *UpdateServerGroupRequest) (*entities.ServerGroup, error) {
	// Get existing group
	group, err := s.GetServerGroup(ctx, id)
	if err != nil {
		return nil, err
	}

	// Prepare updates
	updates := make(map[string]interface{})

	if req.Name != nil {
		// Check if another group with the same name already exists
		var existingGroup entities.ServerGroup
		if err := s.db.DB.WithContext(ctx).Where("name = ? AND id != ?", *req.Name, id).First(&existingGroup).Error; err == nil {
			return nil, fmt.Errorf("server group with name '%s' already exists", *req.Name)
		} else if err != gorm.ErrRecordNotFound {
			logger.Error("Failed to check existing server group", logger.Error2("error", err))
			return nil, fmt.Errorf("failed to check existing server group: %w", err)
		}
		updates["name"] = *req.Name
	}

	if len(updates) == 0 {
		return group, nil
	}

	// Update the group
	if err := s.db.DB.WithContext(ctx).Model(group).Updates(updates).Error; err != nil {
		logger.Error("Failed to update server group", logger.Error2("error", err), logger.Uint("group_id", id))
		return nil, fmt.Errorf("failed to update server group: %w", err)
	}

	// Reload the group
	if err := s.db.DB.WithContext(ctx).First(group, id).Error; err != nil {
		logger.Error("Failed to reload updated server group", logger.Error2("error", err), logger.Uint("group_id", id))
		return nil, fmt.Errorf("failed to reload updated server group: %w", err)
	}

	logger.Info("Server group updated successfully", logger.Uint("group_id", group.ID))

	return group, nil
}

// DeleteServerGroup soft deletes a server group
func (s *ServerGroupService) DeleteServerGroup(ctx context.Context, id uint) error {
	group, err := s.GetServerGroup(ctx, id)
	if err != nil {
		return err
	}

	if err := s.db.DB.WithContext(ctx).Delete(group).Error; err != nil {
		logger.Error("Failed to delete server group", logger.Error2("error", err), logger.Uint("group_id", id))
		return fmt.Errorf("failed to delete server group: %w", err)
	}

	logger.Info("Server group deleted successfully", logger.Uint("group_id", id))

	return nil
}

// GetAllServerGroups gets all server groups for dropdown/selection purposes
func (s *ServerGroupService) GetAllServerGroups(ctx context.Context) ([]*entities.ServerGroup, error) {
	var groups []*entities.ServerGroup

	if err := s.db.DB.WithContext(ctx).Order("name ASC").Find(&groups).Error; err != nil {
		logger.Error("Failed to get all server groups", logger.Error2("error", err))
		return nil, fmt.Errorf("failed to get all server groups: %w", err)
	}

	return groups, nil
}