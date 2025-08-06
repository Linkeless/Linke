package server

import (
	"go.uber.org/fx"
	"gorm.io/gorm"

	"linke/internal/domains/server/adapters/repositories"
	"linke/internal/domains/server/handlers"
	"linke/internal/domains/server/usecases/implementations"
	"linke/internal/domains/server/usecases/interfaces"
	"linke/internal/shared/config"
)

// Module Server 领域模块
// 提供服务器管理、健康监控、负载均衡等功能
var Module = fx.Module("server",
	// 提供 Repository 实现
	fx.Provide(
		fx.Annotate(
			repositories.NewServerGroupRepository,
			fx.As(new(interfaces.ServerGroupRepository)),
		),
		fx.Annotate(
			repositories.NewShadowsocksServerRepository,
			fx.As(new(interfaces.ShadowsocksServerRepository)),
		),
	),

	// 提供 Service 实现
	fx.Provide(
		fx.Annotate(
			implementations.NewServerGroupService,
			fx.As(new(interfaces.ServerGroupService)),
		),
		fx.Annotate(
			implementations.NewShadowsocksServerService,
			fx.As(new(interfaces.ShadowsocksServerService)),
		),
	),

	// 提供 Handler 实现
	fx.Provide(
		handlers.NewServerAPIHandler,
		handlers.NewAdminServerHandler,
		handlers.NewAdminServerGroupHandler,
	),

	// 模块初始化钩子
	fx.Invoke(func(db *gorm.DB, cfg *config.Config) {
		// 确保服务器相关表存在并且结构正确
		// 可以添加默认服务器组的初始化逻辑
	}),
)

// ServiceProvider 为外部模块提供服务器服务接口
type ServiceProvider struct {
	ServerGroupService       interfaces.ServerGroupService
	ShadowsocksServerService interfaces.ShadowsocksServerService
}

// NewServiceProvider 创建服务器服务提供者
func NewServiceProvider(
	serverGroupService interfaces.ServerGroupService,
	shadowsocksServerService interfaces.ShadowsocksServerService,
) *ServiceProvider {
	return &ServiceProvider{
		ServerGroupService:       serverGroupService,
		ShadowsocksServerService: shadowsocksServerService,
	}
}

// 对外暴露的服务提供者模块
var ServiceModule = fx.Module("server-service",
	fx.Provide(NewServiceProvider),
)
