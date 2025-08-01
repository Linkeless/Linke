package persistence

import (
	"context"
	"errors"

	"linke/internal/server/domain/model"
	"linke/internal/server/domain/repository"
	"linke/internal/server/domain/valueobject"

	"gorm.io/gorm"
)

// ShadowsocksServerGORMRepository implements ShadowsocksServerRepository using GORM
type ShadowsocksServerGORMRepository struct {
	db *gorm.DB
}

// NewShadowsocksServerGORMRepository creates a new GORM shadowsocks server repository
func NewShadowsocksServerGORMRepository(db *gorm.DB) repository.ShadowsocksServerRepository {
	return &ShadowsocksServerGORMRepository{
		db: db,
	}
}

// Save saves a shadowsocks server
func (r *ShadowsocksServerGORMRepository) Save(ctx context.Context, server *model.ShadowsocksServer) error {
	po := &ShadowsocksServerPO{}
	po.FromDomain(server)
	
	if server.ID().IsZero() {
		// Create new record
		if err := r.db.WithContext(ctx).Create(po).Error; err != nil {
			return err
		}
		
		// Update the domain model with the generated ID
		id, err := valueobject.NewServerID(po.ID)
		if err != nil {
			return err
		}
		server.MarkAsCreated(id)
	} else {
		// Update existing record
		if err := r.db.WithContext(ctx).Save(po).Error; err != nil {
			return err
		}
	}
	
	return nil
}

// FindByID finds a shadowsocks server by ID
func (r *ShadowsocksServerGORMRepository) FindByID(ctx context.Context, id valueobject.ServerID) (*model.ShadowsocksServer, error) {
	var po ShadowsocksServerPO
	
	if err := r.db.WithContext(ctx).First(&po, id.Value()).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repository.ErrShadowsocksServerNotFound
		}
		return nil, err
	}
	
	return po.ToDomain()
}

// FindByGroupID finds shadowsocks servers by group ID
func (r *ShadowsocksServerGORMRepository) FindByGroupID(ctx context.Context, groupID valueobject.ServerGroupID) ([]*model.ShadowsocksServer, error) {
	var pos []ShadowsocksServerPO
	
	if err := r.db.WithContext(ctx).
		Where("group_id = ?", groupID.Value()).
		Order("sort ASC, created_at DESC").
		Find(&pos).Error; err != nil {
		return nil, err
	}
	
	return r.convertPOsToDomain(pos)
}

// FindVisibleByGroupID finds visible shadowsocks servers by group ID
func (r *ShadowsocksServerGORMRepository) FindVisibleByGroupID(ctx context.Context, groupID valueobject.ServerGroupID) ([]*model.ShadowsocksServer, error) {
	var pos []ShadowsocksServerPO
	
	if err := r.db.WithContext(ctx).
		Where("group_id = ? AND show = ?", groupID.Value(), 1).
		Order("sort ASC, created_at DESC").
		Find(&pos).Error; err != nil {
		return nil, err
	}
	
	return r.convertPOsToDomain(pos)
}

// FindAll finds all shadowsocks servers with filters and pagination
func (r *ShadowsocksServerGORMRepository) FindAll(ctx context.Context, filters repository.ShadowsocksServerFilters, limit, offset int) ([]*model.ShadowsocksServer, int64, error) {
	var pos []ShadowsocksServerPO
	var total int64
	
	query := r.db.WithContext(ctx).Model(&ShadowsocksServerPO{})
	
	// Apply filters
	if filters.GroupID != nil {
		query = query.Where("group_id = ?", filters.GroupID.Value())
	}
	
	if filters.IsVisible != nil {
		show := 0
		if *filters.IsVisible {
			show = 1
		}
		query = query.Where("show = ?", show)
	}
	
	if filters.Tags != "" {
		query = query.Where("tags LIKE ?", "%"+filters.Tags+"%")
	}
	
	if filters.Cipher != nil {
		query = query.Where("cipher = ?", filters.Cipher.Value())
	}
	
	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	// Apply pagination and ordering
	query = query.Order("sort ASC, created_at DESC")
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	
	if offset > 0 {
		query = query.Offset(offset)
	}
	
	if err := query.Find(&pos).Error; err != nil {
		return nil, 0, err
	}
	
	servers, err := r.convertPOsToDomain(pos)
	if err != nil {
		return nil, 0, err
	}
	
	return servers, total, nil
}

// FindVisible finds all visible shadowsocks servers
func (r *ShadowsocksServerGORMRepository) FindVisible(ctx context.Context) ([]*model.ShadowsocksServer, error) {
	var pos []ShadowsocksServerPO
	
	if err := r.db.WithContext(ctx).
		Where("show = ?", 1).
		Order("sort ASC, created_at DESC").
		Find(&pos).Error; err != nil {
		return nil, err
	}
	
	return r.convertPOsToDomain(pos)
}

// Delete deletes a shadowsocks server
func (r *ShadowsocksServerGORMRepository) Delete(ctx context.Context, id valueobject.ServerID) error {
	result := r.db.WithContext(ctx).Delete(&ShadowsocksServerPO{}, id.Value())
	if result.Error != nil {
		return result.Error
	}
	
	if result.RowsAffected == 0 {
		return repository.ErrShadowsocksServerNotFound
	}
	
	return nil
}

// Count returns the total count of shadowsocks servers
func (r *ShadowsocksServerGORMRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	
	if err := r.db.WithContext(ctx).Model(&ShadowsocksServerPO{}).Count(&count).Error; err != nil {
		return 0, err
	}
	
	return count, nil
}

// CountByGroupID returns the count of shadowsocks servers in a group
func (r *ShadowsocksServerGORMRepository) CountByGroupID(ctx context.Context, groupID valueobject.ServerGroupID) (int64, error) {
	var count int64
	
	if err := r.db.WithContext(ctx).Model(&ShadowsocksServerPO{}).
		Where("group_id = ?", groupID.Value()).
		Count(&count).Error; err != nil {
		return 0, err
	}
	
	return count, nil
}

// convertPOsToDomain converts a slice of POs to domain models
func (r *ShadowsocksServerGORMRepository) convertPOsToDomain(pos []ShadowsocksServerPO) ([]*model.ShadowsocksServer, error) {
	servers := make([]*model.ShadowsocksServer, len(pos))
	for i, po := range pos {
		server, err := po.ToDomain()
		if err != nil {
			return nil, err
		}
		servers[i] = server
	}
	return servers, nil
}