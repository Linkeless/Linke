package application

import (
	"go.uber.org/fx"

	"linke/internal/application/handlers"
	"linke/internal/application/services"
	"linke/internal/application/workflows"
)

// Module 应用层模块
// 提供跨领域业务协调、复杂业务工作流和应用级服务
var Module = fx.Module("application",
	// 应用级服务 (完整版本)
	fx.Provide(
		services.NewApplicationService,
	),

	// 业务工作流 (完整实现)
	fx.Provide(
		// SubscriptionWorkflow 需要多个服务依赖，通过依赖注入自动连接
		workflows.NewSubscriptionWorkflow,
		workflows.NewReferralWorkflow,
	),

	// 应用级处理器 (完整版本)
	fx.Provide(
		handlers.NewApplicationHandler,
		handlers.NewTaskHandler,
	),

	// 应用层初始化
	fx.Invoke(func() {
		// 应用层初始化逻辑
		// SubscriptionWorkflow 的所有依赖都通过 fx 自动注入
		// 包括：logger、database、各种服务接口
	}),
)
