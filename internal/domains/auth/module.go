package auth

import (
	"time"
	"go.uber.org/fx"
	"gorm.io/gorm"

	"linke/internal/domains/auth/adapters/repositories"
	"linke/internal/domains/auth/usecases/implementations"
	"linke/internal/domains/auth/usecases/interfaces"
	"linke/internal/domains/auth/handlers"
	"linke/internal/shared/config"
)

// Module Auth 领域模块
// 提供身份认证、授权管理、会话管理等功能
var Module = fx.Module("auth",
	// 提供 Repository 实现
	fx.Provide(
		fx.Annotate(
			repositories.NewJWTBlacklistRepository,
			fx.As(new(interfaces.JWTBlacklistRepository)),
		),
		fx.Annotate(
			repositories.NewLoginAttemptRepository,
			fx.As(new(interfaces.LoginAttemptRepository)),
		),
		fx.Annotate(
			repositories.NewAccountLockoutRepository,
			fx.As(new(interfaces.AccountLockoutRepository)),
		),
	),

	// 提供基础服务实现
	fx.Provide(
		fx.Annotate(
			implementations.NewJWTBlacklistService,
			fx.As(new(interfaces.JWTBlacklistService)),
		),
		fx.Annotate(
			implementations.NewLoginSecurityService,
			fx.As(new(interfaces.LoginSecurityService)),
		),
	),

	// 提供核心服务实现
	fx.Provide(
		fx.Annotate(
			implementations.NewJWTService,
			fx.As(new(interfaces.JWTService)),
		),
		fx.Annotate(
			implementations.NewOAuthService,
			fx.As(new(interfaces.OAuthService)),
		),
		fx.Annotate(
			implementations.NewAuthService,
			fx.As(new(interfaces.AuthService)),
		),
	),

	// 提供 Handler 实现
	fx.Provide(
		handlers.NewAuthHandler,
	),

	// 提供登录安全配置
	fx.Provide(
		func() *implementations.LoginSecurityConfig {
			return &implementations.LoginSecurityConfig{
				MaxFailures:        5,
				LockoutDuration:    1 * time.Hour,
				FailureWindow:      24 * time.Hour,
				ProgressiveLockout: true,
			}
		},
	),

	// 模块初始化钩子
	fx.Invoke(func(db *gorm.DB, cfg *config.Config) {
		// 验证 JWT 配置的安全性
		if cfg.JWT.Secret == "" || len(cfg.JWT.Secret) < 32 {
			panic("JWT_SECRET must be set and at least 32 characters long for security")
		}
		
		// 可以添加其他认证相关的初始化逻辑
	}),
)

// ServiceProvider 为外部模块提供认证服务接口
type ServiceProvider struct {
	AuthService interfaces.AuthService
	JWTService  interfaces.JWTService
	OAuthService interfaces.OAuthService
}

// NewServiceProvider 创建认证服务提供者
func NewServiceProvider(
	authService interfaces.AuthService,
	jwtService interfaces.JWTService,
	oauthService interfaces.OAuthService,
) *ServiceProvider {
	return &ServiceProvider{
		AuthService:  authService,
		JWTService:   jwtService,
		OAuthService: oauthService,
	}
}

// 对外暴露的服务提供者模块
var ServiceModule = fx.Module("auth-service",
	fx.Provide(NewServiceProvider),
)