package repositories

import (
	"context"
	"fmt"

	"linke/internal/domains/server/entities"
	"linke/internal/domains/server/usecases/interfaces"
	"linke/internal/shared/logger"

	"gorm.io/gorm"
)

// shadowsocksServerRepository implements the ShadowsocksServerRepository interface
type shadowsocksServerRepository struct {
	db *gorm.DB
}

// NewShadowsocksServerRepository creates a new ShadowsocksServerRepository implementation
func NewShadowsocksServerRepository(db *gorm.DB) interfaces.ShadowsocksServerRepository {
	return &shadowsocksServerRepository{
		db: db,
	}
}

// Create creates a new shadowsocks server in the database
func (r *shadowsocksServerRepository) Create(ctx context.Context, server *entities.ShadowsocksServer) error {
	if err := r.db.WithContext(ctx).Create(server).Error; err != nil {
		logger.Error("Failed to create shadowsocks server in repository",
			logger.String("name", server.Name),
			logger.String("host", server.Host),
			logger.Error2("error", err),
		)
		return fmt.Errorf("failed to create shadowsocks server: %w", err)
	}

	logger.Debug("Shadowsocks server created successfully in repository",
		logger.Int("server_id", server.ID),
		logger.String("name", server.Name),
		logger.String("host", server.Host),
	)
	return nil
}

// GetByID retrieves a shadowsocks server by ID
func (r *shadowsocksServerRepository) GetByID(ctx context.Context, id uint) (*entities.ShadowsocksServer, error) {
	var server entities.ShadowsocksServer
	if err := r.db.WithContext(ctx).Preload("ServerGroup").First(&server, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("shadowsocks server not found")
		}
		logger.Error("Failed to get shadowsocks server by ID",
			logger.Uint("server_id", id),
			logger.Error2("error", err),
		)
		return nil, fmt.Errorf("failed to get shadowsocks server: %w", err)
	}
	return &server, nil
}

// Update updates an existing shadowsocks server
func (r *shadowsocksServerRepository) Update(ctx context.Context, server *entities.ShadowsocksServer) error {
	if err := r.db.WithContext(ctx).Save(server).Error; err != nil {
		logger.Error("Failed to update shadowsocks server in repository",
			logger.Int("server_id", server.ID),
			logger.String("name", server.Name),
			logger.Error2("error", err),
		)
		return fmt.Errorf("failed to update shadowsocks server: %w", err)
	}

	logger.Debug("Shadowsocks server updated successfully in repository",
		logger.Int("server_id", server.ID),
		logger.String("name", server.Name),
	)
	return nil
}

// Delete soft deletes a shadowsocks server
func (r *shadowsocksServerRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&entities.ShadowsocksServer{}, id)
	if result.Error != nil {
		logger.Error("Failed to delete shadowsocks server in repository",
			logger.Uint("server_id", id),
			logger.Error2("error", result.Error),
		)
		return fmt.Errorf("failed to delete shadowsocks server: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("shadowsocks server not found")
	}

	logger.Debug("Shadowsocks server deleted successfully in repository",
		logger.Uint("server_id", id),
	)
	return nil
}

// ListByGroup retrieves shadowsocks servers by group ID with pagination
func (r *shadowsocksServerRepository) ListByGroup(ctx context.Context, groupID uint, limit, offset int) ([]*entities.ShadowsocksServer, int64, error) {
	var servers []*entities.ShadowsocksServer
	var total int64

	// Get total count for the group
	if err := r.db.WithContext(ctx).Model(&entities.ShadowsocksServer{}).Where("group_id = ?", groupID).Count(&total).Error; err != nil {
		logger.Error("Failed to count shadowsocks servers by group",
			logger.Uint("group_id", groupID),
			logger.Error2("error", err),
		)
		return nil, 0, fmt.Errorf("failed to count shadowsocks servers by group: %w", err)
	}

	query := r.db.WithContext(ctx).Preload("ServerGroup").Where("group_id = ?", groupID).Order("sort ASC, created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&servers).Error; err != nil {
		logger.Error("Failed to list shadowsocks servers by group",
			logger.Uint("group_id", groupID),
			logger.Int("limit", limit),
			logger.Int("offset", offset),
			logger.Error2("error", err),
		)
		return nil, 0, fmt.Errorf("failed to list shadowsocks servers by group: %w", err)
	}

	return servers, total, nil
}

// CountByGroup returns the total number of shadowsocks servers in a group
func (r *shadowsocksServerRepository) CountByGroup(ctx context.Context, groupID uint) (int64, error) {
	var total int64
	if err := r.db.WithContext(ctx).Model(&entities.ShadowsocksServer{}).Where("group_id = ?", groupID).Count(&total).Error; err != nil {
		logger.Error("Failed to count shadowsocks servers by group",
			logger.Uint("group_id", groupID),
			logger.Error2("error", err),
		)
		return 0, fmt.Errorf("failed to count shadowsocks servers by group: %w", err)
	}
	return total, nil
}

// ListActive retrieves active shadowsocks servers with pagination
func (r *shadowsocksServerRepository) ListActive(ctx context.Context, limit, offset int) ([]*entities.ShadowsocksServer, int64, error) {
	var servers []*entities.ShadowsocksServer
	var total int64

	// Get total count of active servers (show = 1)
	if err := r.db.WithContext(ctx).Model(&entities.ShadowsocksServer{}).Where("show = ?", 1).Count(&total).Error; err != nil {
		logger.Error("Failed to count active shadowsocks servers",
			logger.Error2("error", err),
		)
		return nil, 0, fmt.Errorf("failed to count active shadowsocks servers: %w", err)
	}

	query := r.db.WithContext(ctx).Preload("ServerGroup").Where("show = ?", 1).Order("sort ASC, created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&servers).Error; err != nil {
		logger.Error("Failed to list active shadowsocks servers",
			logger.Int("limit", limit),
			logger.Int("offset", offset),
			logger.Error2("error", err),
		)
		return nil, 0, fmt.Errorf("failed to list active shadowsocks servers: %w", err)
	}

	return servers, total, nil
}

// ListByStatus retrieves shadowsocks servers by status with pagination
func (r *shadowsocksServerRepository) ListByStatus(ctx context.Context, status string, limit, offset int) ([]*entities.ShadowsocksServer, int64, error) {
	// Note: The current entity doesn't have a status field, using 'show' field as status indicator
	// If a separate status field is needed, it should be added to the entity

	var servers []*entities.ShadowsocksServer
	var total int64

	var query *gorm.DB

	// Map status to show field values
	switch status {
	case "active":
		query = r.db.WithContext(ctx).Model(&entities.ShadowsocksServer{}).Where("show = ?", 1)
	case "inactive":
		query = r.db.WithContext(ctx).Model(&entities.ShadowsocksServer{}).Where("show = ?", 0)
	default:
		query = r.db.WithContext(ctx).Model(&entities.ShadowsocksServer{})
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		logger.Error("Failed to count shadowsocks servers by status",
			logger.String("status", status),
			logger.Error2("error", err),
		)
		return nil, 0, fmt.Errorf("failed to count shadowsocks servers by status: %w", err)
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
		logger.Error("Failed to list shadowsocks servers by status",
			logger.String("status", status),
			logger.Int("limit", limit),
			logger.Int("offset", offset),
			logger.Error2("error", err),
		)
		return nil, 0, fmt.Errorf("failed to list shadowsocks servers by status: %w", err)
	}

	return servers, total, nil
}

// UpdateStatus updates the status of a shadowsocks server
func (r *shadowsocksServerRepository) UpdateStatus(ctx context.Context, id uint, status string) error {
	// Map status to show field value
	var showValue int
	switch status {
	case "active":
		showValue = 1
	case "inactive":
		showValue = 0
	default:
		return fmt.Errorf("invalid status: %s", status)
	}

	result := r.db.WithContext(ctx).Model(&entities.ShadowsocksServer{}).Where("id = ?", id).Update("show", showValue)
	if result.Error != nil {
		logger.Error("Failed to update shadowsocks server status",
			logger.Uint("server_id", id),
			logger.String("status", status),
			logger.Error2("error", result.Error),
		)
		return fmt.Errorf("failed to update shadowsocks server status: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("shadowsocks server not found")
	}

	logger.Debug("Shadowsocks server status updated successfully",
		logger.Uint("server_id", id),
		logger.String("status", status),
	)
	return nil
}

// ListByLocation retrieves shadowsocks servers by location with pagination
func (r *shadowsocksServerRepository) ListByLocation(ctx context.Context, country, region string, limit, offset int) ([]*entities.ShadowsocksServer, int64, error) {
	// Note: The current entity doesn't have country/region fields
	// This is a placeholder implementation. If location fields are needed, they should be added to the entity

	var servers []*entities.ShadowsocksServer
	var total int64

	query := r.db.WithContext(ctx).Model(&entities.ShadowsocksServer{})

	// For now, we'll use name field to filter by location information if it contains the country/region
	if country != "" {
		query = query.Where("name LIKE ?", "%"+country+"%")
	}
	if region != "" {
		query = query.Where("name LIKE ?", "%"+region+"%")
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		logger.Error("Failed to count shadowsocks servers by location",
			logger.String("country", country),
			logger.String("region", region),
			logger.Error2("error", err),
		)
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
		logger.Error("Failed to list shadowsocks servers by location",
			logger.String("country", country),
			logger.String("region", region),
			logger.Int("limit", limit),
			logger.Int("offset", offset),
			logger.Error2("error", err),
		)
		return nil, 0, fmt.Errorf("failed to list shadowsocks servers by location: %w", err)
	}

	return servers, total, nil
}

// List retrieves all shadowsocks servers with pagination
func (r *shadowsocksServerRepository) List(ctx context.Context, limit, offset int) ([]*entities.ShadowsocksServer, int64, error) {
	var servers []*entities.ShadowsocksServer
	var total int64

	// Get total count
	if err := r.db.WithContext(ctx).Model(&entities.ShadowsocksServer{}).Count(&total).Error; err != nil {
		logger.Error("Failed to count shadowsocks servers",
			logger.Error2("error", err),
		)
		return nil, 0, fmt.Errorf("failed to count shadowsocks servers: %w", err)
	}

	query := r.db.WithContext(ctx).Preload("ServerGroup").Order("sort ASC, created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&servers).Error; err != nil {
		logger.Error("Failed to list shadowsocks servers",
			logger.Int("limit", limit),
			logger.Int("offset", offset),
			logger.Error2("error", err),
		)
		return nil, 0, fmt.Errorf("failed to list shadowsocks servers: %w", err)
	}

	return servers, total, nil
}

// CountTotal returns the total number of shadowsocks servers
func (r *shadowsocksServerRepository) CountTotal(ctx context.Context) (int64, error) {
	var total int64
	if err := r.db.WithContext(ctx).Model(&entities.ShadowsocksServer{}).Count(&total).Error; err != nil {
		logger.Error("Failed to count total shadowsocks servers",
			logger.Error2("error", err),
		)
		return 0, fmt.Errorf("failed to count total shadowsocks servers: %w", err)
	}
	return total, nil
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
	if err := r.db.WithContext(ctx).Raw(`
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
		logger.Error("Failed to get location stats",
			logger.Error2("error", err),
		)
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

	if err := r.db.WithContext(ctx).Raw(`
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
		logger.Error("Failed to get status stats",
			logger.Error2("error", err),
		)
		return nil, fmt.Errorf("failed to get status stats: %w", err)
	}

	stats := make(map[string]int64)
	for _, result := range results {
		stats[result.Status] = result.Count
	}

	return stats, nil
}
