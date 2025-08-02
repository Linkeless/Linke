package services

import (
	"context"

	"linke/internal/shared/logger"
	"linke/internal/shared/database"
)

// SimpleApplicationService 简化的应用层服务
// 当领域模块暂时不可用时使用
type SimpleApplicationService struct {
	logger   logger.Logger
	database *database.Database
}

// NewSimpleApplicationService 创建简化的应用层服务
func NewSimpleApplicationService(
	logger logger.Logger,
	database *database.Database,
) *SimpleApplicationService {
	return &SimpleApplicationService{
		logger:   logger,
		database: database,
	}
}

// HealthCheck 系统健康检查
func (s *SimpleApplicationService) HealthCheck(ctx context.Context) map[string]interface{} {
	result := make(map[string]interface{})
	
	// 数据库健康检查
	dbHealth := s.database.HealthCheck(ctx)
	result["database"] = dbHealth
	
	// 添加应用层状态
	result["application"] = map[string]interface{}{
		"status": "healthy",
		"mode":   "simplified", // 表示使用简化模式
		"note":   "Domain modules temporarily disabled due to import issues",
	}
	
	return result
}