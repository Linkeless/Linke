package ticket

import (
	"go.uber.org/fx"
	"gorm.io/gorm"

	"linke/internal/domains/ticket/adapters/repositories"
	"linke/internal/domains/ticket/handlers"
	"linke/internal/domains/ticket/usecases/implementations"
	"linke/internal/domains/ticket/usecases/interfaces"
)

// Module Ticket 领域模块
// 提供客户支持、工单管理、SLA 跟踪等功能
var Module = fx.Module("ticket",
	// 提供 Repository 实现
	fx.Provide(
		fx.Annotate(
			repositories.NewTicketRepository,
			fx.As(new(interfaces.TicketRepository)),
		),
		fx.Annotate(
			repositories.NewTicketMessageRepository,
			fx.As(new(interfaces.TicketMessageRepository)),
		),
	),

	// 提供 Service 实现
	fx.Provide(
		fx.Annotate(
			implementations.NewTicketService,
			fx.As(new(interfaces.TicketService)),
		),
		fx.Annotate(
			implementations.NewTicketMessageService,
			fx.As(new(interfaces.TicketMessageService)),
		),
	),

	// 提供 Handler 实现
	fx.Provide(
		handlers.NewAdminTicketHandler,
	),

	// 模块初始化钩子
	fx.Invoke(func(db *gorm.DB) {
		// 确保工单相关表存在并且结构正确
		// 可以添加工单模板或 SLA 规则的初始化逻辑
	}),
)

// ServiceProvider 为外部模块提供工单服务接口
type ServiceProvider struct {
	TicketService        interfaces.TicketService
	TicketMessageService interfaces.TicketMessageService
}

// NewServiceProvider 创建工单服务提供者
func NewServiceProvider(
	ticketService interfaces.TicketService,
	ticketMessageService interfaces.TicketMessageService,
) *ServiceProvider {
	return &ServiceProvider{
		TicketService:        ticketService,
		TicketMessageService: ticketMessageService,
	}
}

// 对外暴露的服务提供者模块
var ServiceModule = fx.Module("ticket-service",
	fx.Provide(NewServiceProvider),
)
