package subscription

import (
	"go.uber.org/fx"
	"gorm.io/gorm"

	couponInterfaces "linke/internal/domains/coupon/usecases/interfaces"
	invoiceInterfaces "linke/internal/domains/invoice/usecases/interfaces"
	paymentInterfaces "linke/internal/domains/payment/usecases/interfaces"
	"linke/internal/domains/subscription/adapters/repositories"
	"linke/internal/domains/subscription/handlers"
	"linke/internal/domains/subscription/usecases/implementations"
	"linke/internal/domains/subscription/usecases/interfaces"
	"linke/internal/shared/cache"
)

// NewSubscriptionPlanServiceWithCache creates a subscription plan service with cache support
func NewSubscriptionPlanServiceWithCache(
	db *gorm.DB,
	cacheManager cache.CacheManager,
	cacheKeys *cache.AllCacheKeys,
) interfaces.SubscriptionPlanService {
	return implementations.NewSubscriptionPlanServiceWithCache(
		db,
		cacheManager,
		cacheKeys,
	)
}

// NewUserSubscriptionServiceWithCache creates a user subscription service with cache support
func NewUserSubscriptionServiceWithCache(
	db *gorm.DB,
	subscriptionPlanService interfaces.SubscriptionPlanService,
	cacheManager cache.CacheManager,
	cacheKeys *cache.AllCacheKeys,
) interfaces.UserSubscriptionService {
	return implementations.NewUserSubscriptionServiceWithCache(
		db,
		subscriptionPlanService,
		cacheManager,
		cacheKeys,
	)
}

// NewSubscriptionOrderServiceWithInvoice creates a subscription order service with invoice service dependency
func NewSubscriptionOrderServiceWithInvoice(
	db *gorm.DB,
	subscriptionPlanService interfaces.SubscriptionPlanService,
	userSubscriptionService interfaces.UserSubscriptionService,
	paymentService paymentInterfaces.PaymentService,
	paymentMethodService paymentInterfaces.PaymentMethodService,
	couponService couponInterfaces.CouponService,
	invoiceService invoiceInterfaces.InvoiceService,
	cacheManager cache.CacheManager,
	cacheKeys *cache.AllCacheKeys,
) interfaces.SubscriptionOrderService {
	return implementations.NewSubscriptionOrderServiceWithCache(
		db,
		subscriptionPlanService,
		userSubscriptionService,
		paymentService,
		paymentMethodService,
		couponService,
		invoiceService,
		cacheManager,
		cacheKeys,
	)
}

// Module Subscription 领域模块
// 提供订阅计划管理、用户订阅生命周期、计费周期管理等功能
var Module = fx.Module("subscription",
	// 提供 Repository 实现
	fx.Provide(
		fx.Annotate(
			repositories.NewSubscriptionPlanRepository,
			fx.As(new(interfaces.SubscriptionPlanRepository)),
		),
		fx.Annotate(
			repositories.NewSubscriptionOrderRepository,
			fx.As(new(interfaces.SubscriptionOrderRepository)),
		),
		fx.Annotate(
			repositories.NewUserSubscriptionRepository,
			fx.As(new(interfaces.UserSubscriptionRepository)),
		),
		// Usage tracking repositories
		fx.Annotate(
			repositories.NewUsageRepository,
			fx.As(new(interfaces.UsageRepository)),
		),
		fx.Annotate(
			repositories.NewAlertRepository,
			fx.As(new(interfaces.AlertRepository)),
		),
	),

	// 提供基础服务实现（带缓存支持）
	fx.Provide(
		fx.Annotate(
			NewSubscriptionPlanServiceWithCache,
			fx.As(new(interfaces.SubscriptionPlanService)),
		),
		fx.Annotate(
			NewUserSubscriptionServiceWithCache,
			fx.As(new(interfaces.UserSubscriptionService)),
		),
		// Usage tracking services
		fx.Annotate(
			implementations.NewUsageAlertService,
			fx.As(new(interfaces.UsageAlertService)),
		),
		fx.Annotate(
			implementations.NewUsageTrackingService,
			fx.As(new(interfaces.UsageTrackingService)),
		),
	),

	// 提供复合服务实现
	// 注意：这些服务依赖其他服务，需要在基础服务之后提供
	fx.Provide(
		fx.Annotate(
			NewSubscriptionOrderServiceWithInvoice,
			fx.As(new(interfaces.SubscriptionOrderService)),
		),
	),

	// 提供 Handler 实现
	fx.Provide(
		handlers.NewSubscriptionOrderHandler,
		handlers.NewUserSubscriptionHandler,
		handlers.NewQuickPurchaseHandler,
		handlers.NewUsageAlertHandler,
		handlers.NewUsageHandler,
		handlers.NewAdminSubscriptionHandler,
	),

	// 提供扩展服务（这些可能没有接口定义）
	// 注意：这些服务可能还没有实现，暂时注释掉
	// fx.Provide(
	//	implementations.NewUserSubscriptionExtendedService,
	//	implementations.NewUserSubscriptionServerGroupService,
	// ),

	// 模块初始化钩子
	fx.Invoke(func(db *gorm.DB) {
		// 确保订阅相关表存在并且结构正确
		// 可以添加默认订阅计划的初始化逻辑
	}),
)

// ServiceProvider 为外部模块提供订阅服务接口
type ServiceProvider struct {
	SubscriptionPlanService  interfaces.SubscriptionPlanService
	UserSubscriptionService  interfaces.UserSubscriptionService
	SubscriptionOrderService interfaces.SubscriptionOrderService
	UsageTrackingService     interfaces.UsageTrackingService
	UsageAlertService        interfaces.UsageAlertService
}

// NewServiceProvider 创建订阅服务提供者
func NewServiceProvider(
	planService interfaces.SubscriptionPlanService,
	userSubService interfaces.UserSubscriptionService,
	orderService interfaces.SubscriptionOrderService,
	usageTrackingService interfaces.UsageTrackingService,
	usageAlertService interfaces.UsageAlertService,
) *ServiceProvider {
	return &ServiceProvider{
		SubscriptionPlanService:  planService,
		UserSubscriptionService:  userSubService,
		SubscriptionOrderService: orderService,
		UsageTrackingService:     usageTrackingService,
		UsageAlertService:        usageAlertService,
	}
}

// 对外暴露的服务提供者模块
var ServiceModule = fx.Module("subscription-service",
	fx.Provide(NewServiceProvider),
)
