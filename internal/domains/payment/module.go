package payment

import (
	"go.uber.org/fx"
	"gorm.io/gorm"

	"linke/internal/domains/payment/adapters/handlers"
	"linke/internal/domains/payment/adapters/repositories"
	"linke/internal/domains/payment/gateways"
	"linke/internal/domains/payment/usecases/implementations"
	"linke/internal/domains/payment/usecases/interfaces"
	"linke/internal/shared/cache"
)

// Module Payment 领域模块
// 提供epay支付处理、支付记录管理等功能
var Module = fx.Module("payment",
	// 提供 Repository 实现
	fx.Provide(
		fx.Annotate(
			repositories.NewPaymentRecordRepository,
			fx.As(new(interfaces.PaymentRecordRepository)),
		),
		fx.Annotate(
			repositories.NewPaymentConfigRepository,
			fx.As(new(interfaces.PaymentConfigRepository)),
		),
		fx.Annotate(
			repositories.NewPaymentMethodRepository,
			fx.As(new(interfaces.PaymentMethodRepository)),
		),
	),

	// 提供 Service 实现
	fx.Provide(
		// Gateway factory
		gateways.NewGatewayFactory,

		// Base service implementations that need EventBus
		fx.Annotate(
			implementations.NewPaymentService,
			fx.As(new(interfaces.PaymentService)),
		),
		fx.Annotate(
			implementations.NewPaymentConfigService,
			fx.As(new(interfaces.PaymentConfigService)),
		),

		// Payment method service implementation
		fx.Annotate(
			implementations.NewPaymentMethodService,
			fx.As(new(interfaces.PaymentMethodService)),
		),

		// Cached service implementations (decorators)
		implementations.NewCachedPaymentService,
	),

	// 提供 Handler 实现
	fx.Provide(
		handlers.NewPaymentOrderHandler,
		handlers.NewPaymentConfigHandler,
		handlers.NewPaymentRecordHandler,
		handlers.NewPaymentMethodHandler,
	),

	// 模块初始化钩子
	fx.Invoke(func(db *gorm.DB) {
		// 确保支付相关表存在并且结构正确
		// 可以添加默认epay网关配置的初始化逻辑
	}),
)

// ServiceProvider 为外部模块提供支付服务接口
type ServiceProvider struct {
	PaymentService       interfaces.PaymentService
	PaymentConfigService interfaces.PaymentConfigService
	PaymentMethodService interfaces.PaymentMethodService
	GatewayFactory       *gateways.GatewayFactory
	CacheManager         cache.CacheManager
}

// NewServiceProvider 创建支付服务提供者
func NewServiceProvider(
	paymentService interfaces.PaymentService,
	paymentConfigService interfaces.PaymentConfigService,
	paymentMethodService interfaces.PaymentMethodService,
	gatewayFactory *gateways.GatewayFactory,
	cacheManager cache.CacheManager,
) *ServiceProvider {
	return &ServiceProvider{
		PaymentService:       paymentService,
		PaymentConfigService: paymentConfigService,
		PaymentMethodService: paymentMethodService,
		GatewayFactory:       gatewayFactory,
		CacheManager:         cacheManager,
	}
}

// 对外暴露的服务提供者模块
var ServiceModule = fx.Module("payment-service",
	fx.Provide(NewServiceProvider),
)
