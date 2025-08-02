package invoice

import (
	"go.uber.org/fx"
	"gorm.io/gorm"

	"linke/internal/domains/invoice/usecases/implementations"
	"linke/internal/domains/invoice/usecases/interfaces"
)

// Module Invoice 领域模块
// 提供发票生成、管理、PDF 生成、通知发送等功能
var Module = fx.Module("invoice",
	// 注意：目前 invoice 领域还没有 repository 实现
	// 当添加了 repository 时，需要在这里提供

	// 提供 Service 实现
	fx.Provide(
		fx.Annotate(
			implementations.NewInvoiceService,
			fx.As(new(interfaces.InvoiceService)),
		),
	),

	// 注意：目前 invoice 领域还没有 handler 实现
	// 当添加了 handler 时，需要在这里提供

	// 模块初始化钩子
	fx.Invoke(func(db *gorm.DB) {
		// 确保发票相关表存在并且结构正确
		// 可以添加发票模板的初始化逻辑
	}),
)

// ServiceProvider 为外部模块提供发票服务接口
type ServiceProvider struct {
	InvoiceService interfaces.InvoiceService
}

// NewServiceProvider 创建发票服务提供者
func NewServiceProvider(invoiceService interfaces.InvoiceService) *ServiceProvider {
	return &ServiceProvider{
		InvoiceService: invoiceService,
	}
}

// 对外暴露的服务提供者模块
var ServiceModule = fx.Module("invoice-service",
	fx.Provide(NewServiceProvider),
)
