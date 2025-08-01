package persistence

import (
	"context"
	"errors"

	"linke/internal/server/domain/model"
	"linke/internal/server/domain/repository"
	"linke/internal/server/domain/valueobject"

	"gorm.io/gorm"
)

// ServerGroupGORMRepository implements ServerGroupRepository using GORM
type ServerGroupGORMRepository struct {
	db *gorm.DB
}

// NewServerGroupGORMRepository creates a new GORM server group repository
func NewServerGroupGORMRepository(db *gorm.DB) repository.ServerGroupRepository {
	return &ServerGroupGORMRepository{
		db: db,
	}
}

// Save saves a server group
func (r *ServerGroupGORMRepository) Save(ctx context.Context, group *model.ServerGroup) error {
	po := &ServerGroupPO{}
	po.FromDomain(group)
	
	if group.ID().IsZero() {
		// Create new record
		if err := r.db.WithContext(ctx).Create(po).Error; err != nil {
			return err
		}
		
		// Update the domain model with the generated ID
		id, err := valueobject.NewServerGroupID(po.ID)
		if err != nil {
			return err
		}
		group.MarkAsCreated(id)
	} else {
		// Update existing record
		if err := r.db.WithContext(ctx).Save(po).Error; err != nil {
			return err
		}
	}
	
	return nil
}

// FindByID finds a server group by ID
func (r *ServerGroupGORMRepository) FindByID(ctx context.Context, id valueobject.ServerGroupID) (*model.ServerGroup, error) {
	var po ServerGroupPO
	
	if err := r.db.WithContext(ctx).First(&po, id.Value()).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repository.ErrServerGroupNotFound
		}
		return nil, err
	}
	
	return po.ToDomain()
}

// FindByName finds a server group by name
func (r *ServerGroupGORMRepository) FindByName(ctx context.Context, name valueobject.ServerGroupName) (*model.ServerGroup, error) {
	var po ServerGroupPO
	
	if err := r.db.WithContext(ctx).Where("name = ?", name.Value()).First(&po).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repository.ErrServerGroupNotFound
		}
		return nil, err
	}
	
	return po.ToDomain()
}

// FindAll finds all server groups with pagination
func (r *ServerGroupGORMRepository) FindAll(ctx context.Context, limit, offset int) ([]*model.ServerGroup, int64, error) {
	var pos []ServerGroupPO
	var total int64
	
	query := r.db.WithContext(ctx).Model(&ServerGroupPO{})
	
	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	// Apply pagination and ordering
	query = query.Order("created_at DESC")
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	
	if offset > 0 {
		query = query.Offset(offset)
	}
	
	if err := query.Find(&pos).Error; err != nil {
		return nil, 0, err
	}
	
	// Convert to domain models
	groups := make([]*model.ServerGroup, len(pos))
	for i, po := range pos {
		group, err := po.ToDomain()
		if err != nil {
			return nil, 0, err
		}
		groups[i] = group
	}
	
	return groups, total, nil
}

// ExistsByName checks if a server group with the given name exists
func (r *ServerGroupGORMRepository) ExistsByName(ctx context.Context, name valueobject.ServerGroupName) (bool, error) {
	var count int64
	
	if err := r.db.WithContext(ctx).Model(&ServerGroupPO{}).Where("name = ?", name.Value()).Count(&count).Error; err != nil {
		return false, err
	}
	
	return count > 0, nil
}

// ExistsByNameExcludingID checks if a server group with the given name exists, excluding a specific ID
func (r *ServerGroupGORMRepository) ExistsByNameExcludingID(ctx context.Context, name valueobject.ServerGroupName, excludeID valueobject.ServerGroupID) (bool, error) {
	var count int64
	
	if err := r.db.WithContext(ctx).Model(&ServerGroupPO{}).
		Where("name = ? AND id != ?", name.Value(), excludeID.Value()).
		Count(&count).Error; err != nil {
		return false, err
	}
	
	return count > 0, nil
}

// Delete deletes a server group
func (r *ServerGroupGORMRepository) Delete(ctx context.Context, id valueobject.ServerGroupID) error {
	result := r.db.WithContext(ctx).Delete(&ServerGroupPO{}, id.Value())
	if result.Error != nil {
		return result.Error
	}
	
	if result.RowsAffected == 0 {
		return repository.ErrServerGroupNotFound
	}
	
	return nil
}

// Count returns the total count of server groups
func (r *ServerGroupGORMRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	
	if err := r.db.WithContext(ctx).Model(&ServerGroupPO{}).Count(&count).Error; err != nil {
		return 0, err
	}
	
	return count, nil
}