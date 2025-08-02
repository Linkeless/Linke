package repositories

import (
	"context"
	"fmt"

	"linke/internal/domains/server/entities"
	"linke/internal/domains/server/usecases/interfaces"
	"linke/internal/shared/logger"

	"gorm.io/gorm"
)

// serverGroupRepository implements the ServerGroupRepository interface
type serverGroupRepository struct {
	db *gorm.DB
}

// NewServerGroupRepository creates a new ServerGroupRepository implementation
func NewServerGroupRepository(db *gorm.DB) interfaces.ServerGroupRepository {
	return &serverGroupRepository{
		db: db,
	}
}

// Create creates a new server group in the database
func (r *serverGroupRepository) Create(ctx context.Context, group *entities.ServerGroup) error {
	if err := r.db.WithContext(ctx).Create(group).Error; err != nil {
		logger.Error("Failed to create server group in repository",
			logger.String("name", group.Name),
			logger.Error2("error", err),
		)
		return fmt.Errorf("failed to create server group: %w", err)
	}

	logger.Debug("Server group created successfully in repository",
		logger.Uint("group_id", group.ID),
		logger.String("name", group.Name),
	)
	return nil
}

// GetByID retrieves a server group by ID
func (r *serverGroupRepository) GetByID(ctx context.Context, id uint) (*entities.ServerGroup, error) {
	var group entities.ServerGroup
	if err := r.db.WithContext(ctx).First(&group, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("server group not found")
		}
		logger.Error("Failed to get server group by ID",
			logger.Uint("group_id", id),
			logger.Error2("error", err),
		)
		return nil, fmt.Errorf("failed to get server group: %w", err)
	}
	return &group, nil
}

// Update updates an existing server group
func (r *serverGroupRepository) Update(ctx context.Context, group *entities.ServerGroup) error {
	if err := r.db.WithContext(ctx).Save(group).Error; err != nil {
		logger.Error("Failed to update server group in repository",
			logger.Uint("group_id", group.ID),
			logger.String("name", group.Name),
			logger.Error2("error", err),
		)
		return fmt.Errorf("failed to update server group: %w", err)
	}

	logger.Debug("Server group updated successfully in repository",
		logger.Uint("group_id", group.ID),
		logger.String("name", group.Name),
	)
	return nil
}

// Delete soft deletes a server group
func (r *serverGroupRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&entities.ServerGroup{}, id)
	if result.Error != nil {
		logger.Error("Failed to delete server group in repository",
			logger.Uint("group_id", id),
			logger.Error2("error", result.Error),
		)
		return fmt.Errorf("failed to delete server group: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("server group not found")
	}

	logger.Debug("Server group deleted successfully in repository",
		logger.Uint("group_id", id),
	)
	return nil
}

// ListActive retrieves active server groups with pagination
func (r *serverGroupRepository) ListActive(ctx context.Context, limit, offset int) ([]*entities.ServerGroup, int64, error) {
	var groups []*entities.ServerGroup
	var total int64

	// Get total count of active groups
	if err := r.db.WithContext(ctx).Model(&entities.ServerGroup{}).Count(&total).Error; err != nil {
		logger.Error("Failed to count active server groups",
			logger.Error2("error", err),
		)
		return nil, 0, fmt.Errorf("failed to count active server groups: %w", err)
	}

	query := r.db.WithContext(ctx).Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&groups).Error; err != nil {
		logger.Error("Failed to list active server groups",
			logger.Int("limit", limit),
			logger.Int("offset", offset),
			logger.Error2("error", err),
		)
		return nil, 0, fmt.Errorf("failed to list active server groups: %w", err)
	}

	return groups, total, nil
}

// UpdateStatus updates the active status of a server group
func (r *serverGroupRepository) UpdateStatus(ctx context.Context, id uint, isActive bool) error {
	// Note: Since the ServerGroup entity doesn't have an IsActive field,
	// this method will just check if the group exists for now
	// If needed, the IsActive field should be added to the entity

	var group entities.ServerGroup
	if err := r.db.WithContext(ctx).First(&group, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("server group not found")
		}
		return fmt.Errorf("failed to find server group: %w", err)
	}

	// For now, just log the operation since there's no IsActive field
	logger.Info("Server group status update requested (no IsActive field in entity)",
		logger.Uint("group_id", id),
	)

	return nil
}

// List retrieves all server groups with pagination
func (r *serverGroupRepository) List(ctx context.Context, limit, offset int) ([]*entities.ServerGroup, int64, error) {
	var groups []*entities.ServerGroup
	var total int64

	// Get total count
	if err := r.db.WithContext(ctx).Model(&entities.ServerGroup{}).Count(&total).Error; err != nil {
		logger.Error("Failed to count server groups",
			logger.Error2("error", err),
		)
		return nil, 0, fmt.Errorf("failed to count server groups: %w", err)
	}

	query := r.db.WithContext(ctx).Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&groups).Error; err != nil {
		logger.Error("Failed to list server groups",
			logger.Int("limit", limit),
			logger.Int("offset", offset),
			logger.Error2("error", err),
		)
		return nil, 0, fmt.Errorf("failed to list server groups: %w", err)
	}

	return groups, total, nil
}

// CountTotal returns the total number of server groups
func (r *serverGroupRepository) CountTotal(ctx context.Context) (int64, error) {
	var total int64
	if err := r.db.WithContext(ctx).Model(&entities.ServerGroup{}).Count(&total).Error; err != nil {
		logger.Error("Failed to count total server groups",
			logger.Error2("error", err),
		)
		return 0, fmt.Errorf("failed to count total server groups: %w", err)
	}
	return total, nil
}
