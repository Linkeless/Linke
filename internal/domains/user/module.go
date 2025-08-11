package user

import (
	"go.uber.org/fx"
	"gorm.io/gorm"

	"linke/internal/domains/user/adapters/repositories"
	"linke/internal/domains/user/adapters/handlers"
	"linke/internal/domains/user/usecases/implementations"
	"linke/internal/domains/user/usecases/interfaces"
	"linke/internal/shared/cache"
	"linke/internal/shared/events"
	"linke/internal/shared/framework"
)

// Module User 领域模块
// 提供用户生命周期管理、用户信息维护、用户状态管理、第三方账号绑定等功能
var Module = fx.Module("user",
	// 提供 Repository 实现
	fx.Provide(
		fx.Annotate(
			repositories.NewUserRepository,
			fx.As(new(interfaces.UserRepository)),
		),
		fx.Annotate(
			repositories.NewUserAccountBindingRepository,
			fx.As(new(interfaces.UserAccountBindingRepository)),
		),
	),

	// 提供 Service 实现
	fx.Provide(
		// 提供统一的 UserService 实现，支持可配置的缓存和事件功能
		fx.Annotate(
			func(db *gorm.DB, logger framework.Logger, cacheManager cache.CacheManager, eventBus events.EventBus) interfaces.UserService {
				// 创建带有缓存和事件功能的统一用户服务
				return implementations.NewUnifiedUserService(
					db,
					logger,
					implementations.WithCaching(cacheManager.GetCache()),
					implementations.WithEvents(eventBus),
				)
			},
			fx.As(new(interfaces.UserService)),
		),

		// 提供用户账号绑定服务
		fx.Annotate(
			implementations.NewUserAccountBindingService,
			fx.As(new(interfaces.UserAccountBindingService)),
		),
	),

	// 提供 Handler 实现
	fx.Provide(
		handlers.NewUserProfileHandler,
		handlers.NewAdminUserHandler,
		handlers.NewUserAccountBindingHandler,
	),

	// 模块初始化钩子
	fx.Invoke(func(db *gorm.DB, logger framework.Logger, cacheManager cache.CacheManager) {
		// 确保用户表存在并且结构正确
		// 这里可以添加用户领域特定的初始化逻辑
		logger.Info("User domain module initialized with caching support and account binding")
	}),
)

// ServiceProvider 为外部模块提供用户服务接口
// 这是一个辅助结构，用于其他领域模块获取用户服务
type ServiceProvider struct {
	UserService interfaces.UserService
}

// NewServiceProvider 创建用户服务提供者
func NewServiceProvider(userService interfaces.UserService) *ServiceProvider {
	return &ServiceProvider{
		UserService: userService,
	}
}

// 对外暴露的服务提供者模块
var ServiceModule = fx.Module("user-service",
	fx.Provide(NewServiceProvider),
)
