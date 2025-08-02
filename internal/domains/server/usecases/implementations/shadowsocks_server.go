package implementations

import (
	"context"
	"fmt"
	"time"

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

// CreateShadowsocksServerRequest represents the request to create a shadowsocks server
type CreateShadowsocksServerRequest struct {
	GroupID      uint    `json:"group_id" binding:"required"`
	RouteID      string  `json:"route_id,omitempty" binding:"max=255"`
	ParentID     *int    `json:"parent_id,omitempty"`
	Name         string  `json:"name" binding:"required,max=255"`
	Tags         string  `json:"tags,omitempty" binding:"max=255"`
	Host         string  `json:"host" binding:"required,max=255"`
	Port         int     `json:"port" binding:"required,min=1,max=65535"`
	ServerPort   int     `json:"server_port" binding:"required,min=1,max=65535"`
	Cipher       string  `json:"cipher" binding:"required,max=255"`
	Obfs         string  `json:"obfs,omitempty" binding:"max=11"`
	ObfsSettings string  `json:"obfs_settings,omitempty" binding:"max=255"`
	Excludes     string  `json:"excludes,omitempty"`
	IPs          string  `json:"ips,omitempty" binding:"max=255"`
	Rate         float64 `json:"rate" binding:"required,min=0.1"`
	Show         int     `json:"show" binding:"min=0,max=1"`
	Sort         *int    `json:"sort,omitempty"`
}

// UpdateShadowsocksServerRequest represents the request to update a shadowsocks server
type UpdateShadowsocksServerRequest struct {
	GroupID      *uint    `json:"group_id,omitempty"`
	RouteID      *string  `json:"route_id,omitempty" binding:"omitempty,max=255"`
	ParentID     *int     `json:"parent_id,omitempty"`
	Name         *string  `json:"name,omitempty" binding:"omitempty,max=255"`
	Tags         *string  `json:"tags,omitempty" binding:"omitempty,max=255"`
	Host         *string  `json:"host,omitempty" binding:"omitempty,max=255"`
	Port         *int     `json:"port,omitempty" binding:"omitempty,min=1,max=65535"`
	ServerPort   *int     `json:"server_port,omitempty" binding:"omitempty,min=1,max=65535"`
	Cipher       *string  `json:"cipher,omitempty" binding:"omitempty,max=255"`
	Obfs         *string  `json:"obfs,omitempty" binding:"omitempty,max=11"`
	ObfsSettings *string  `json:"obfs_settings,omitempty" binding:"omitempty,max=255"`
	Excludes     *string  `json:"excludes,omitempty"`
	IPs          *string  `json:"ips,omitempty" binding:"omitempty,max=255"`
	Rate         *float64 `json:"rate,omitempty" binding:"omitempty,min=0.1"`
	Show         *int     `json:"show,omitempty" binding:"omitempty,min=0,max=1"`
	Sort         *int     `json:"sort,omitempty"`
}

// GetShadowsocksServersRequest represents the request to get shadowsocks servers with filters
type GetShadowsocksServersRequest struct {
	GroupID *uint `json:"group_id,omitempty"`
	Show    *int  `json:"show,omitempty"`
	Limit   int   `json:"limit,omitempty"`
	Offset  int   `json:"offset,omitempty"`
}

// CreateShadowsocksServer creates a new shadowsocks server
func (s *ShadowsocksServerService) CreateShadowsocksServer(ctx context.Context, req *CreateShadowsocksServerRequest) (*entities.ShadowsocksServer, error) {
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
			logger.Error2("error", err),
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
func (s *ShadowsocksServerService) GetShadowsocksServers(ctx context.Context, req *GetShadowsocksServersRequest) ([]*entities.ShadowsocksServer, int64, error) {
	var servers []*entities.ShadowsocksServer
	var total int64

	query := s.db.DB.WithContext(ctx).Model(&entities.ShadowsocksServer{})

	// Apply filters
	if req.GroupID != nil {
		query = query.Where("group_id = ?", *req.GroupID)
	}
	if req.Show != nil {
		query = query.Where("show = ?", *req.Show)
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
	if err := query.Order("sort ASC, created_at DESC").Find(&servers).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get shadowsocks servers: %w", err)
	}

	return servers, total, nil
}

// GetActiveShadowsocksServers retrieves all active shadowsocks servers
func (s *ShadowsocksServerService) GetActiveShadowsocksServers(ctx context.Context) ([]*entities.ShadowsocksServer, error) {
	var servers []*entities.ShadowsocksServer
	if err := s.db.DB.WithContext(ctx).
		Where("show = ?", 1).
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
		Where("group_id = ? AND show = ?", groupID, 1).
		Order("sort ASC, created_at DESC").
		Preload("ServerGroup"). // Load the server group relationship
		Find(&servers).Error; err != nil {
		return nil, fmt.Errorf("failed to get shadowsocks servers by group: %w", err)
	}
	return servers, nil
}

// UpdateShadowsocksServer updates a shadowsocks server
func (s *ShadowsocksServerService) UpdateShadowsocksServer(ctx context.Context, id int, req *UpdateShadowsocksServerRequest) (*entities.ShadowsocksServer, error) {
	var server entities.ShadowsocksServer
	if err := s.db.DB.WithContext(ctx).First(&server, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("shadowsocks server not found")
		}
		return nil, fmt.Errorf("failed to get shadowsocks server: %w", err)
	}

	// Update fields
	updates := make(map[string]interface{})
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
			logger.Int("server_id", id),
			logger.Error2("error", err),
		)
		return nil, fmt.Errorf("failed to update shadowsocks server: %w", err)
	}

	logger.Info("Shadowsocks server updated successfully",
		logger.Int("server_id", id),
		logger.String("name", server.Name),
	)

	return &server, nil
}

// DeleteShadowsocksServer soft deletes a shadowsocks server
func (s *ShadowsocksServerService) DeleteShadowsocksServer(ctx context.Context, id int) error {
	result := s.db.DB.WithContext(ctx).Delete(&entities.ShadowsocksServer{}, id)
	if result.Error != nil {
		logger.Error("Failed to delete shadowsocks server",
			logger.Int("server_id", id),
			logger.Error2("error", result.Error),
		)
		return fmt.Errorf("failed to delete shadowsocks server: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("shadowsocks server not found")
	}

	logger.Info("Shadowsocks server deleted successfully",
		logger.Int("server_id", id),
	)

	return nil
}
