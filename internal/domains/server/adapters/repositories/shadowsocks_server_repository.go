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

// shadowsocksServerRepository implements the ShadowsocksServerRepository interface
type shadowsocksServerRepository struct {
	*repository.BaseRepositoryImpl[entities.ShadowsocksServer, int]
}

// NewShadowsocksServerRepository creates a new ShadowsocksServerRepository implementation
func NewShadowsocksServerRepository(db *gorm.DB, logger framework.Logger) interfaces.ShadowsocksServerRepository {
	return &shadowsocksServerRepository{
		BaseRepositoryImpl: repository.NewBaseRepository[entities.ShadowsocksServer, int](db, logger),
	}
}


// ListByGroup retrieves shadowsocks servers by group ID with pagination
func (r *shadowsocksServerRepository) ListByGroup(ctx context.Context, groupID uint, limit, offset int) ([]*entities.ShadowsocksServer, int64, error) {
	var servers []*entities.ShadowsocksServer
	var total int64

	// Get total count for the group
	if err := r.GetDB().WithContext(ctx).Model(&entities.ShadowsocksServer{}).Where("group_id = ?", groupID).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count shadowsocks servers by group: %w", err)
	}

	query := r.GetDB().WithContext(ctx).Preload("ServerGroup").Where("group_id = ?", groupID).Order("sort ASC, created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&servers).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list shadowsocks servers by group: %w", err)
	}

	return servers, total, nil
}

// CountByGroup returns the total number of shadowsocks servers in a group
func (r *shadowsocksServerRepository) CountByGroup(ctx context.Context, groupID uint) (int64, error) {
	var total int64
	if err := r.GetDB().WithContext(ctx).Model(&entities.ShadowsocksServer{}).Where("group_id = ?", groupID).Count(&total).Error; err != nil {
		return 0, fmt.Errorf("failed to count shadowsocks servers by group: %w", err)
	}
	return total, nil
}

// ListActive retrieves active shadowsocks servers with pagination
func (r *shadowsocksServerRepository) ListActive(ctx context.Context, limit, offset int) ([]*entities.ShadowsocksServer, int64, error) {
	var servers []*entities.ShadowsocksServer
	var total int64

	// Get total count of active servers (show = 1)
	if err := r.GetDB().WithContext(ctx).Model(&entities.ShadowsocksServer{}).Where("show = ?", 1).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count active shadowsocks servers: %w", err)
	}

	query := r.GetDB().WithContext(ctx).Preload("ServerGroup").Where("show = ?", 1).Order("sort ASC, created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&servers).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list active shadowsocks servers: %w", err)
	}

	return servers, total, nil
}


// ListByLocation retrieves shadowsocks servers by location with pagination
func (r *shadowsocksServerRepository) ListByLocation(ctx context.Context, country, region string, limit, offset int) ([]*entities.ShadowsocksServer, int64, error) {
	// Note: The current entity doesn't have country/region fields
	// This is a placeholder implementation. If location fields are needed, they should be added to the entity

	var servers []*entities.ShadowsocksServer
	var total int64

	query := r.GetDB().WithContext(ctx).Model(&entities.ShadowsocksServer{})

	// For now, we'll use name field to filter by location information if it contains the country/region
	if country != "" {
		query = query.Where("name LIKE ?", "%"+country+"%")
	}
	if region != "" {
		query = query.Where("name LIKE ?", "%"+region+"%")
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count shadowsocks servers by location: %w", err)
	}

	// Apply preload, ordering, and pagination
	query = query.Preload("ServerGroup").Order("sort ASC, created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&servers).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list shadowsocks servers by location: %w", err)
	}

	return servers, total, nil
}


// GetLocationStats returns statistics about servers grouped by location
func (r *shadowsocksServerRepository) GetLocationStats(ctx context.Context) (map[string]int64, error) {
	// Note: Since the entity doesn't have dedicated location fields,
	// this is a placeholder implementation that could be enhanced with proper location fields

	var results []struct {
		Location string
		Count    int64
	}

	// This is a simplified implementation that extracts location info from the name field
	// In a real implementation, you would have dedicated location fields
	if err := r.GetDB().WithContext(ctx).Raw(`
		SELECT 
			CASE 
				WHEN name LIKE '%US%' OR name LIKE '%America%' THEN 'US'
				WHEN name LIKE '%CN%' OR name LIKE '%China%' THEN 'CN'
				WHEN name LIKE '%HK%' OR name LIKE '%Hong Kong%' THEN 'HK'
				WHEN name LIKE '%JP%' OR name LIKE '%Japan%' THEN 'JP'
				WHEN name LIKE '%SG%' OR name LIKE '%Singapore%' THEN 'SG'
				ELSE 'Other'
			END as location,
			COUNT(*) as count
		FROM shadowsocks_servers 
		GROUP BY location
	`).Scan(&results).Error; err != nil {
		return nil, fmt.Errorf("failed to get location stats: %w", err)
	}

	stats := make(map[string]int64)
	for _, result := range results {
		stats[result.Location] = result.Count
	}

	return stats, nil
}

// GetStatusStats returns statistics about servers grouped by status
func (r *shadowsocksServerRepository) GetStatusStats(ctx context.Context) (map[string]int64, error) {
	var results []struct {
		Status string
		Count  int64
	}

	if err := r.GetDB().WithContext(ctx).Raw(`
		SELECT 
			CASE 
				WHEN show = 1 THEN 'active'
				WHEN show = 0 THEN 'inactive'
				ELSE 'unknown'
			END as status,
			COUNT(*) as count
		FROM shadowsocks_servers 
		GROUP BY status
	`).Scan(&results).Error; err != nil {
		return nil, fmt.Errorf("failed to get status stats: %w", err)
	}

	stats := make(map[string]int64)
	for _, result := range results {
		stats[result.Status] = result.Count
	}

	return stats, nil
}
