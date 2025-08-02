package application

import (
	"go.uber.org/fx"

	"linke/internal/application/handlers"
	"linke/internal/application/services"
	// "linke/internal/application/workflows" // 暂时禁用
)

// Module 应用层模块
// 提供跨领域业务协调、复杂业务工作流和应用级服务
var Module = fx.Module("application",
	// 应用级服务 (简化版本)
	fx.Provide(
		services.NewSimpleApplicationService,
	),

	// 业务工作流 (暂时禁用，等待领域模块修复)
	// fx.Provide(
	//	workflows.NewSubscriptionWorkflow,
	//	workflows.NewReferralWorkflow,
	// ),

	// 应用级处理器 (简化版本)
	fx.Provide(
		handlers.NewSimpleApplicationHandler,
		handlers.NewTaskHandler,
	),

	// 应用层初始化
	fx.Invoke(func() {
		// 应用层初始化逻辑
	}),
)
