package payment

import (
	"go.uber.org/fx"
	"gorm.io/gorm"

	"linke/internal/domains/payment/adapters/repositories"
	"linke/internal/domains/payment/adapters/handlers"
	"linke/internal/domains/payment/usecases/implementations"
	"linke/internal/domains/payment/usecases/interfaces"
	"linke/internal/shared/cache"
	"linke/internal/shared/queue"

	"github.com/go-redis/redis/v8"
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
		fx.Annotate(
			repositories.NewPaymentRetryRepository,
			fx.As(new(interfaces.PaymentRetryRepository)),
		),
		fx.Annotate(
			repositories.NewPaymentRetryHistoryRepository,
			fx.As(new(interfaces.PaymentRetryHistoryRepository)),
		),
		fx.Annotate(
			repositories.NewPaymentMethodRepository,
			fx.As(new(interfaces.PaymentMethodRepository)),
		),
	),

	// 提供 Service 实现
	fx.Provide(
		// Base service implementations
		implementations.NewPaymentService,
		implementations.NewPaymentConfigService,

		// Payment method service implementation
		fx.Annotate(
			implementations.NewPaymentMethodService,
			fx.As(new(interfaces.PaymentMethodService)),
		),

		// Task queue for retry service
		fx.Annotate(
			NewTaskQueueForPayment,
			fx.As(new(implementations.TaskQueueInterface)),
		),

		// Retry service implementation
		fx.Annotate(
			implementations.NewPaymentRetryService,
			fx.As(new(interfaces.PaymentRetryService)),
		),

		// Worker implementation
		implementations.NewPaymentRetryWorker,

		// Cached service implementations
		fx.Annotate(
			implementations.NewCachedPaymentService,
			fx.As(new(interfaces.PaymentService)),
		),
		fx.Annotate(
			implementations.NewCachedPaymentConfigService,
			fx.As(new(interfaces.PaymentConfigService)),
		),
	),

	// 提供 Handler 实现
	fx.Provide(
		handlers.NewPaymentHandler,
		handlers.NewPaymentMethodHandler,
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
	PaymentRetryService  interfaces.PaymentRetryService
	PaymentMethodService interfaces.PaymentMethodService
	CacheManager         cache.CacheManager
}

// NewServiceProvider 创建支付服务提供者
func NewServiceProvider(
	paymentService interfaces.PaymentService,
	paymentConfigService interfaces.PaymentConfigService,
	paymentRetryService interfaces.PaymentRetryService,
	paymentMethodService interfaces.PaymentMethodService,
	cacheManager cache.CacheManager,
) *ServiceProvider {
	return &ServiceProvider{
		PaymentService:       paymentService,
		PaymentConfigService: paymentConfigService,
		PaymentRetryService:  paymentRetryService,
		PaymentMethodService: paymentMethodService,
		CacheManager:         cacheManager,
	}
}

// NewTaskQueueForPayment 创建用于支付重试的任务队列
func NewTaskQueueForPayment(redisClient *redis.Client) implementations.TaskQueueInterface {
	return queue.NewDelayedTaskQueue(redisClient)
}

// 对外暴露的服务提供者模块
var ServiceModule = fx.Module("payment-service",
	fx.Provide(NewServiceProvider),
)
