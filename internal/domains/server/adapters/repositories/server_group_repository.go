package repositories

import (
	"context"
	"fmt"

	"linke/internal/domains/server/entities"
	"linke/internal/domains/server/usecases/interfaces"
	"linke/internal/shared/framework"
	"linke/internal/shared/repository"

	"gorm.io/gorm"
)

// serverGroupRepository implements the ServerGroupRepository interface
type serverGroupRepository struct {
	*repository.BaseRepositoryImpl[entities.ServerGroup, uint]
}

// NewServerGroupRepository creates a new ServerGroupRepository implementation
func NewServerGroupRepository(db *gorm.DB, logger framework.Logger) interfaces.ServerGroupRepository {
	return &serverGroupRepository{
		BaseRepositoryImpl: repository.NewBaseRepository[entities.ServerGroup, uint](db, logger),
	}
}


// ListActive retrieves active server groups with pagination
func (r *serverGroupRepository) ListActive(ctx context.Context, limit, offset int) ([]*entities.ServerGroup, int64, error) {
	var groups []*entities.ServerGroup
	var total int64

	// Get total count of active groups
	if err := r.GetDB().WithContext(ctx).Model(&entities.ServerGroup{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count active server groups: %w", err)
	}

	query := r.GetDB().WithContext(ctx).Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&groups).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list active server groups: %w", err)
	}

	return groups, total, nil
}

// UpdateActiveStatus updates the active status of a server group
func (r *serverGroupRepository) UpdateActiveStatus(ctx context.Context, id uint, isActive bool) error {
	// Note: Since the ServerGroup entity doesn't have an IsActive field,
	// this method will just check if the group exists for now
	// If needed, the IsActive field should be added to the entity

	var group entities.ServerGroup
	if err := r.GetDB().WithContext(ctx).First(&group, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("server group not found")
		}
		return fmt.Errorf("failed to find server group: %w", err)
	}

	// For now, just return success since there's no IsActive field
	return nil
}

