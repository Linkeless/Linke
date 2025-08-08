package auth

import (
	"go.uber.org/fx"
	"gorm.io/gorm"

	"linke/internal/domains/auth/adapters/repositories"
	"linke/internal/domains/auth/handlers"
	"linke/internal/domains/auth/usecases/implementations"
	"linke/internal/domains/auth/usecases/interfaces"
	referralInterfaces "linke/internal/domains/referral/usecases/interfaces"
	userInterfaces "linke/internal/domains/user/usecases/interfaces"
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
		// Provide JWTBlacklistService as concrete type first
		implementations.NewJWTBlacklistService,
		// Provide interface wrapper
		fx.Annotate(
			func(service *implementations.JWTBlacklistService) interfaces.JWTBlacklistService {
				return service
			},
			fx.As(new(interfaces.JWTBlacklistService)),
		),
		// 🔴 DISABLED: Login security service temporarily disabled
		// fx.Annotate(
		// 	implementations.NewLoginSecurityService,
		// 	fx.As(new(interfaces.LoginSecurityService)),
		// ),
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
			func(
				db *gorm.DB,
				userService userInterfaces.UserService,
				userRepository userInterfaces.UserRepository,
				userBindingService userInterfaces.UserAccountBindingService,
				jwtService interfaces.JWTService,
				inviteCodeService referralInterfaces.InviteCodeService,
				// loginSecurityService interfaces.LoginSecurityService, // 🔴 DISABLED: Removed dependency
			) interfaces.AuthService {
				// Pass nil for loginSecurityService to disable login security
				return implementations.NewAuthService(db, userService, userRepository, userBindingService, jwtService, inviteCodeService, nil)
			},
			fx.As(new(interfaces.AuthService)),
		),
	),

	// 提供 Handler 实现
	fx.Provide(
		handlers.NewAuthHandler,
		handlers.NewAdminAuthHandler,
	),

	// 🔴 DISABLED: Login security configuration not needed when service is disabled
	// fx.Provide(
	// 	func() *implementations.LoginSecurityConfig {
	// 		return &implementations.LoginSecurityConfig{
	// 			// Production values (for future reference):
	// 			MaxFailures:        5,               // 5 failed attempts trigger lockout
	// 			LockoutDuration:    15 * time.Minute, // 15 minutes lockout duration
	// 			FailureWindow:      1 * time.Hour,    // 1 hour failure counting window
	// 			ProgressiveLockout: true,            // Enable progressive lockout
	// 		}
	// 	},
	// ),

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
	AuthService  interfaces.AuthService
	JWTService   interfaces.JWTService
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
