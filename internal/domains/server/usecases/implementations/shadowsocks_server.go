package implementations

import (
	"context"
	"fmt"
	"strings"
	"time"

	"linke/internal/domains/server/dto"
	"linke/internal/domains/server/entities"
	"linke/internal/shared/database"
	"linke/internal/shared/logger"

	"gorm.io/gorm"
)

type ShadowsocksServerService struct {
	db *database.Database
}

func NewShadowsocksServerService(db *database.Database) *ShadowsocksServerService {
	return &ShadowsocksServerService{
		db: db,
	}
}

// CreateShadowsocksServer creates a new shadowsocks server
func (s *ShadowsocksServerService) CreateShadowsocksServer(ctx context.Context, req *dto.CreateShadowsocksServerRequest) (*entities.ShadowsocksServer, error) {
	// Set default values
	if req.Rate == 0 {
		req.Rate = 1.0
	}

	now := int(time.Now().Unix())
	server := &entities.ShadowsocksServer{
		GroupID:      req.GroupID,
		RouteID:      req.RouteID,
		ParentID:     req.ParentID,
		Name:         req.Name,
		Tags:         req.Tags,
		Host:         req.Host,
		Port:         req.Port,
		ServerPort:   req.ServerPort,
		Cipher:       req.Cipher,
		Obfs:         req.Obfs,
		ObfsSettings: req.ObfsSettings,
		Excludes:     req.Excludes,
		IPs:          req.IPs,
		Rate:         req.Rate,
		Show:         req.Show,
		Sort:         req.Sort,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.db.DB.WithContext(ctx).Create(server).Error; err != nil {
		logger.Error("Failed to create shadowsocks server",
			logger.String("name", req.Name),
			logger.String("host", req.Host),
			logger.Uint("group_id", req.GroupID),
			logger.ErrorField(err),
		)
		return nil, fmt.Errorf("failed to create shadowsocks server: %w", err)
	}

	logger.Info("Shadowsocks server created successfully",
		logger.Int("server_id", server.ID),
		logger.String("name", server.Name),
		logger.String("host", server.Host),
		logger.Uint("group_id", server.GroupID),
	)

	return server, nil
}

// GetShadowsocksServerByID retrieves a shadowsocks server by ID
func (s *ShadowsocksServerService) GetShadowsocksServerByID(ctx context.Context, id int) (*entities.ShadowsocksServer, error) {
	var server entities.ShadowsocksServer
	if err := s.db.DB.WithContext(ctx).First(&server, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("shadowsocks server not found")
		}
		return nil, fmt.Errorf("failed to get shadowsocks server: %w", err)
	}
	return &server, nil
}

// GetShadowsocksServers retrieves shadowsocks servers with optional filters
func (s *ShadowsocksServerService) GetShadowsocksServers(ctx context.Context, req *dto.GetShadowsocksServersRequest) ([]*entities.ShadowsocksServer, int64, error) {
	var servers []*entities.ShadowsocksServer
	var total int64

	query := s.db.DB.WithContext(ctx).Model(&entities.ShadowsocksServer{}).Select("id, route_id, parent_id, name, tags, excludes, ips, rate, host, port, server_port, cipher, obfs, obfs_settings, `show`, sort, created_at, updated_at")

	// Apply filters
	if req.GroupID != nil {
		query = query.Where("group_id = ?", *req.GroupID)
	}
	if req.Show != nil {
		query = query.Where("`show` = ?", *req.Show)
	}
	if req.Name != "" {
		query = query.Where("name LIKE ?", "%"+req.Name+"%")
	}

	// Determine sort field safely
	orderClause := "sort ASC, created_at DESC"
	if req.SortBy != "" {
		// allowlist
		field := strings.ToLower(req.SortBy)
		direction := "ASC"
		if strings.ToLower(req.SortOrder) == "desc" {
			direction = "DESC"
		}
		switch field {
		case "sort":
			orderClause = "sort " + direction + ", created_at DESC"
		case "created_at":
			orderClause = "created_at " + direction
		case "updated_at":
			orderClause = "updated_at " + direction
		case "name":
			orderClause = "name " + direction + ", sort ASC"
		case "rate":
			orderClause = "rate " + direction + ", sort ASC"
		default:
			orderClause = "sort ASC, created_at DESC"
		}
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count shadowsocks servers: %w", err)
	}

	// Apply pagination
	if req.Limit > 0 {
		query = query.Limit(req.Limit)
	}
	if req.Offset > 0 {
		query = query.Offset(req.Offset)
	}

	// Get servers
	if err := query.Order(orderClause).Find(&servers).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get shadowsocks servers: %w", err)
	}

	return servers, total, nil
}

// GetActiveShadowsocksServers retrieves all active shadowsocks servers
func (s *ShadowsocksServerService) GetActiveShadowsocksServers(ctx context.Context) ([]*entities.ShadowsocksServer, error) {
	var servers []*entities.ShadowsocksServer
	if err := s.db.DB.WithContext(ctx).
		Where("`show` = ?", 1).
		Order("sort ASC, created_at DESC").
		Find(&servers).Error; err != nil {
		return nil, fmt.Errorf("failed to get active shadowsocks servers: %w", err)
	}
	return servers, nil
}

// GetShadowsocksServersByGroupID retrieves shadowsocks servers by group ID
func (s *ShadowsocksServerService) GetShadowsocksServersByGroupID(ctx context.Context, groupID uint) ([]*entities.ShadowsocksServer, error) {
	var servers []*entities.ShadowsocksServer
	if err := s.db.DB.WithContext(ctx).
		Where("group_id = ? AND `show` = ?", groupID, 1).
		Order("sort ASC, created_at DESC").
		Preload("ServerGroup"). // Load the server group relationship
		Find(&servers).Error; err != nil {
		return nil, fmt.Errorf("failed to get shadowsocks servers by group: %w", err)
	}
	return servers, nil
}

// GetShadowsocksServer retrieves a shadowsocks server by uint ID
func (s *ShadowsocksServerService) GetShadowsocksServer(ctx context.Context, serverID uint) (*entities.ShadowsocksServer, error) {
	var server entities.ShadowsocksServer
	if err := s.db.DB.WithContext(ctx).First(&server, serverID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("shadowsocks server not found")
		}
		return nil, fmt.Errorf("failed to get shadowsocks server: %w", err)
	}
	return &server, nil
}

// UpdateShadowsocksServer updates a shadowsocks server
func (s *ShadowsocksServerService) UpdateShadowsocksServer(ctx context.Context, serverID uint, req *dto.UpdateShadowsocksServerRequest) (*entities.ShadowsocksServer, error) {
	var server entities.ShadowsocksServer
	if err := s.db.DB.WithContext(ctx).First(&server, serverID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("shadowsocks server not found")
		}
		return nil, fmt.Errorf("failed to get shadowsocks server: %w", err)
	}

	// Update fields
	updates := make(map[string]any)
	updates["updated_at"] = int(time.Now().Unix())

	if req.GroupID != nil {
		updates["group_id"] = *req.GroupID
	}
	if req.RouteID != nil {
		updates["route_id"] = *req.RouteID
	}
	if req.ParentID != nil {
		updates["parent_id"] = *req.ParentID
	}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Tags != nil {
		updates["tags"] = *req.Tags
	}
	if req.Host != nil {
		updates["host"] = *req.Host
	}
	if req.Port != nil {
		updates["port"] = *req.Port
	}
	if req.ServerPort != nil {
		updates["server_port"] = *req.ServerPort
	}
	if req.Cipher != nil {
		updates["cipher"] = *req.Cipher
	}
	if req.Obfs != nil {
		updates["obfs"] = *req.Obfs
	}
	if req.ObfsSettings != nil {
		updates["obfs_settings"] = *req.ObfsSettings
	}
	if req.Excludes != nil {
		updates["excludes"] = *req.Excludes
	}
	if req.IPs != nil {
		updates["ips"] = *req.IPs
	}
	if req.Rate != nil {
		updates["rate"] = *req.Rate
	}
	if req.Show != nil {
		updates["show"] = *req.Show
	}
	if req.Sort != nil {
		updates["sort"] = *req.Sort
	}

	if err := s.db.DB.WithContext(ctx).Model(&server).Updates(updates).Error; err != nil {
		logger.Error("Failed to update shadowsocks server",
			logger.Uint("server_id", serverID),
			logger.ErrorField(err),
		)
		return nil, fmt.Errorf("failed to update shadowsocks server: %w", err)
	}

	logger.Info("Shadowsocks server updated successfully",
		logger.Uint("server_id", serverID),
		logger.String("name", server.Name),
	)

	return &server, nil
}

// DeleteShadowsocksServer soft deletes a shadowsocks server
func (s *ShadowsocksServerService) DeleteShadowsocksServer(ctx context.Context, serverID uint) error {
	result := s.db.DB.WithContext(ctx).Delete(&entities.ShadowsocksServer{}, serverID)
	if result.Error != nil {
		logger.Error("Failed to delete shadowsocks server",
			logger.Uint("server_id", serverID),
			logger.ErrorField(result.Error),
		)
		return fmt.Errorf("failed to delete shadowsocks server: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("shadowsocks server not found")
	}

	logger.Info("Shadowsocks server deleted successfully",
		logger.Uint("server_id", serverID),
	)

	return nil
}

// GetServersByGroup retrieves shadowsocks servers by group ID
func (s *ShadowsocksServerService) GetServersByGroup(ctx context.Context, groupID uint) ([]*entities.ShadowsocksServer, error) {
	return s.GetShadowsocksServersByGroupID(ctx, groupID)
}

// GetVisibleServers retrieves visible shadowsocks servers with optional group filter
func (s *ShadowsocksServerService) GetVisibleServers(ctx context.Context, groupID *uint) ([]*entities.ShadowsocksServer, error) {
	query := s.db.DB.WithContext(ctx).Where("`show` = ?", 1)

	if groupID != nil {
		query = query.Where("group_id = ?", *groupID)
	}

	var servers []*entities.ShadowsocksServer
	if err := query.Order("sort ASC, created_at DESC").Find(&servers).Error; err != nil {
		return nil, fmt.Errorf("failed to get visible servers: %w", err)
	}
	return servers, nil
}

// UpdateServerStatus updates a shadowsocks server's status
func (s *ShadowsocksServerService) UpdateServerStatus(ctx context.Context, serverID uint, status string) error {
	result := s.db.DB.WithContext(ctx).Model(&entities.ShadowsocksServer{}).
		Where("id = ?", serverID).
		Update("show", status)

	if result.Error != nil {
		return fmt.Errorf("failed to update server status: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("shadowsocks server not found")
	}

	return nil
}

// BulkUpdateServers updates multiple shadowsocks servers
func (s *ShadowsocksServerService) BulkUpdateServers(ctx context.Context, serverIDs []uint, updates map[string]any) error {
	if len(serverIDs) == 0 {
		return nil
	}

	updates["updated_at"] = int(time.Now().Unix())

	result := s.db.DB.WithContext(ctx).Model(&entities.ShadowsocksServer{}).
		Where("id IN ?", serverIDs).
		Updates(updates)

	if result.Error != nil {
		return fmt.Errorf("failed to bulk update servers: %w", result.Error)
	}

	return nil
}

// CheckServerHealth checks a shadowsocks server's health
func (s *ShadowsocksServerService) CheckServerHealth(ctx context.Context, serverID uint) (map[string]any, error) {
	// TODO: Implement actual health check logic
	// For now, return a placeholder response
	server, err := s.GetShadowsocksServer(ctx, serverID)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"server_id":  serverID,
		"status":     "unknown", // TODO: Implement actual health check
		"host":       server.Host,
		"port":       server.Port,
		"checked_at": time.Now(),
	}, nil
}

// GetServerStatistics gets a shadowsocks server's statistics
func (s *ShadowsocksServerService) GetServerStatistics(ctx context.Context, serverID uint) (map[string]any, error) {
	// TODO: Implement actual statistics collection
	// For now, return a placeholder response
	server, err := s.GetShadowsocksServer(ctx, serverID)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"server_id":      serverID,
		"connections":    0,    // TODO: Implement actual stats
		"bandwidth_up":   0,    // TODO: Implement actual stats
		"bandwidth_down": 0,    // TODO: Implement actual stats
		"uptime":         "0s", // TODO: Implement actual stats
		"last_updated":   time.Now(),
		"name":           server.Name,
		"host":           server.Host,
	}, nil
}
