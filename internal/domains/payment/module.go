package payment

import (
	"go.uber.org/fx"
	"gorm.io/gorm"

	"linke/internal/domains/payment/adapters/repositories"
	"linke/internal/domains/payment/handlers"
	"linke/internal/domains/payment/usecases/implementations"
	"linke/internal/domains/payment/usecases/interfaces"
)

// Module Payment 领域模块
// 提供支付处理、多网关集成、支付记录管理等功能
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
	),

	// 提供 Service 实现
	fx.Provide(
		fx.Annotate(
			implementations.NewPaymentService,
			fx.As(new(interfaces.PaymentService)),
		),
		fx.Annotate(
			implementations.NewPaymentConfigService,
			fx.As(new(interfaces.PaymentConfigService)),
		),
	),

	// 提供 Handler 实现
	fx.Provide(
		handlers.NewPaymentHandler,
	),

	// 模块初始化钩子
	fx.Invoke(func(db *gorm.DB) {
		// 确保支付相关表存在并且结构正确
		// 可以添加默认支付网关配置的初始化逻辑
	}),
)

// ServiceProvider 为外部模块提供支付服务接口
type ServiceProvider struct {
	PaymentService       interfaces.PaymentService
	PaymentConfigService interfaces.PaymentConfigService
}

// NewServiceProvider 创建支付服务提供者
func NewServiceProvider(
	paymentService interfaces.PaymentService,
	paymentConfigService interfaces.PaymentConfigService,
) *ServiceProvider {
	return &ServiceProvider{
		PaymentService:       paymentService,
		PaymentConfigService: paymentConfigService,
	}
}

// 对外暴露的服务提供者模块
var ServiceModule = fx.Module("payment-service",
	fx.Provide(NewServiceProvider),
)
